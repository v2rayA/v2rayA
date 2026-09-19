package db_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db"
	"github.com/v2rayA/v2rayA/db/configure"
	"go.etcd.io/bbolt"
)

const (
	validConnections = `{"touches":[{"_type":"subscriptionServer","id":1,"sub":0,"outbound":"proxy"},{"_type":"subscriptionServer","id":2,"sub":0,"outbound":"proxy"}]}`
	outboundSetting  = `{"probeURL":"https://example.test/ping","probeInterval":"45s","type":"leastping"}`
	subscriptions    = `[{"address":"https://example.test/subscription","servers":[{},{}]}]`
)

func TestBoltMigrationPublishesVerifiedDatabaseAndRestarts(t *testing.T) {
	configDir := t.TempDir()
	conf.GetEnvironmentConfig().Config = configDir
	boltPath := filepath.Join(configDir, "bolt.db")
	sqlitePath := filepath.Join(configDir, "v2raya.db")

	writeBoltFixture(t, boltPath, []byte(`{"touches":`))
	before, err := os.ReadFile(boltPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := db.MigrateFromBoltDB(); err == nil {
		t.Fatal("migration with corrupt outbound connections succeeded")
	} else if !strings.Contains(err.Error(), "migration verification failed") {
		t.Fatalf("migration failed before verification: %v", err)
	}
	if _, err := os.Stat(sqlitePath); !os.IsNotExist(err) {
		t.Fatalf("failed migration published %s: %v", sqlitePath, err)
	}
	after, err := os.ReadFile(boltPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(after, before) {
		t.Fatal("failed migration changed bolt.db")
	}
	temps, err := filepath.Glob(filepath.Join(configDir, ".v2raya.db.migrate-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(temps) != 0 {
		t.Fatalf("failed migration left temporary files: %v", temps)
	}

	setBoltValue(t, boltPath, "outbound.proxy", "connectedServers", []byte(validConnections))
	// An earlier release built the SQLite file in place: an empty one next
	// to the intact bolt.db must not win.
	if err := os.WriteFile(sqlitePath, nil, 0600); err != nil {
		t.Fatal(err)
	}
	if err := db.MigrateFromBoltDB(); err != nil {
		t.Fatalf("migration after fixture repair: %v", err)
	}
	if err := db.Open(); err != nil {
		t.Fatalf("open migrated database: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	connected := configure.GetConnectedServersByOutbound("proxy")
	if connected == nil || connected.Len() != 2 {
		t.Fatalf("connected servers: %+v", connected)
	}
	for i, which := range connected.Get() {
		if which.TYPE != configure.SubscriptionServerType || which.Sub != 0 || which.ID != i+1 || which.Outbound != "proxy" {
			t.Fatalf("connected server %d: %+v", i, which)
		}
	}
	wantSetting := configure.OutboundSetting{
		ProbeURL:      "https://example.test/ping",
		ProbeInterval: "45s",
		Type:          configure.LeastPing,
	}
	if got := configure.GetOutboundSetting("proxy"); got != wantSetting {
		t.Fatalf("outbound setting: got %+v, want %+v", got, wantSetting)
	}

	backup, err := os.ReadFile(boltPath + ".bak")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(boltPath, backup, 0600); err != nil {
		t.Fatal(err)
	}
	if err := runStartupSequence(); err != nil {
		t.Fatalf("startup with both databases: %v", err)
	}
	if err := runStartupSequence(); err != nil {
		t.Fatalf("second startup: %v", err)
	}
	if _, err := os.Stat(boltPath); !os.IsNotExist(err) {
		t.Fatalf("startup left bolt.db in place: %v", err)
	}
}

func runStartupSequence() error {
	err := db.Open()
	if !errors.Is(err, db.ErrNeedMigration) {
		return err
	}
	if err := db.MigrateFromBoltDB(); err != nil {
		return err
	}
	return db.Open()
}

func writeBoltFixture(t *testing.T, path string, connections []byte) {
	t.Helper()
	database, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Update(func(tx *bbolt.Tx) error {
		touch, err := tx.CreateBucket([]byte("touch"))
		if err != nil {
			return err
		}
		if err := touch.Put([]byte("subscriptions"), []byte(subscriptions)); err != nil {
			return err
		}
		outbound, err := tx.CreateBucket([]byte("outbound.proxy"))
		if err != nil {
			return err
		}
		if err := outbound.Put([]byte("setting"), []byte(outboundSetting)); err != nil {
			return err
		}
		return outbound.Put([]byte("connectedServers"), connections)
	}); err != nil {
		_ = database.Close()
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}
}

func setBoltValue(t *testing.T, path, bucket, key string, value []byte) {
	t.Helper()
	database, err := bbolt.Open(path, 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte(bucket)).Put([]byte(key), value)
	}); err != nil {
		_ = database.Close()
		t.Fatal(err)
	}
	if err := database.Close(); err != nil {
		t.Fatal(err)
	}
}

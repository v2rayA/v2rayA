package db_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db"
	"github.com/v2rayA/v2rayA/db/configure"
)

// A database written by the first SQLite migration holds the groups in
// outbound_connections; opening it must move them where the runtime reads.
func TestLegacyOutboundTablesAreRecoveredOnOpen(t *testing.T) {
	configDir := t.TempDir()
	conf.GetEnvironmentConfig().Config = configDir
	sqlitePath := filepath.Join(configDir, "v2raya.db")
	if err := os.WriteFile(sqlitePath, nil, 0600); err != nil {
		t.Fatal(err)
	}
	seed := func(stmt string, args ...interface{}) {
		t.Helper()
		if _, err := db.GetDB().Exec(stmt, args...); err != nil {
			t.Fatal(err)
		}
	}
	// the legacy tables and the rows the old migration produced
	seed(`CREATE TABLE IF NOT EXISTS outbound_connections (id INTEGER PRIMARY KEY, outbound_name TEXT, server_id INTEGER, sort INTEGER)`)
	seed(`CREATE TABLE IF NOT EXISTS outbound_settings (outbound_name TEXT PRIMARY KEY, setting_json TEXT)`)
	seed(`INSERT INTO system_config (key, value) VALUES ('system:setting', '{}')`)
	seed(`INSERT INTO subscriptions (id, address, sort) VALUES (7, 'https://example.test/sub', 0)`)
	seed(`INSERT INTO servers (id, type, sub_id, config_json, sort) VALUES (11, 'subscription_server', 7, '{}', 0), (12, 'subscription_server', 7, '{}', 1), (13, 'server', NULL, '{}', 0)`)
	seed(`INSERT INTO outbound_connections (outbound_name, server_id, sort) VALUES ('proxy', 12, 0), ('proxy', 13, 1)`)
	seed(`INSERT INTO outbound_settings (outbound_name, setting_json) VALUES ('proxy', '{"probeURL":"https://example.test/ping","probeInterval":"45s","type":"leastping"}')`)
	seed(`INSERT INTO system_config (key, value) VALUES ('outbound.second:connectedServers', '{"touches":[]}')`)
	seed(`INSERT INTO outbound_connections (outbound_name, server_id, sort) VALUES ('second', 11, 0)`)

	if err := db.MigrateSchema(db.GetDB()); err != nil {
		t.Fatal(err)
	}
	got := configure.GetConnectedServersByOutbound("proxy")
	if got == nil || got.Len() != 2 {
		t.Fatalf("proxy members: %+v", got)
	}
	if w := got.Get()[0]; w.TYPE != configure.SubscriptionServerType || w.ID != 2 || w.Sub != 0 {
		t.Fatalf("first member: %+v", w)
	}
	if w := got.Get()[1]; w.TYPE != configure.ServerType || w.ID != 1 {
		t.Fatalf("second member: %+v", w)
	}
	if s := configure.GetOutboundSetting("proxy"); s.ProbeInterval != "45s" {
		t.Fatalf("setting: %+v", s)
	}
	// a key written since the first migration is not overwritten
	if got := configure.GetConnectedServersByOutbound("second"); got != nil && got.Len() != 0 {
		t.Fatalf("second was overwritten: %+v", got.Get())
	}
	_ = db.Close()
}

package db

import (
	"database/sql"
	"testing"
)

func TestInitSchemaLeavesLegacyOutboundTablesUnmanaged(t *testing.T) {
	database, err := sql.Open(sqliteDriverName, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	database.SetMaxOpenConns(1)

	if err := InitSchema(database); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"outbound_connections", "outbound_settings"} {
		if tableExists(t, database, name) {
			t.Fatalf("fresh schema created legacy table %s", name)
		}
	}

	if _, err := database.Exec(`
		CREATE TABLE outbound_connections (id INTEGER PRIMARY KEY);
		CREATE TABLE outbound_settings (outbound_name TEXT PRIMARY KEY);
	`); err != nil {
		t.Fatal(err)
	}
	if err := InitSchema(database); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"outbound_connections", "outbound_settings"} {
		if !tableExists(t, database, name) {
			t.Fatalf("schema initialization dropped legacy table %s", name)
		}
	}
}

func tableExists(t *testing.T, database *sql.DB, name string) bool {
	t.Helper()
	var count int
	if err := database.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?", name).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count != 0
}

func TestMigrateSubscriptionNodeDNSAndHWID(t *testing.T) {
	database, err := sql.Open(sqliteDriverName, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	database.SetMaxOpenConns(1)
	if _, err := database.Exec(`CREATE TABLE system_config (key TEXT PRIMARY KEY, value TEXT NOT NULL);
		CREATE TABLE subscriptions (address TEXT, remarks TEXT, auto_select INTEGER);
		INSERT INTO subscriptions VALUES ('https://example.test/sub', 'old', 0);`); err != nil {
		t.Fatal(err)
	}
	if err := MigrateSchema(database); err != nil {
		t.Fatal(err)
	}
	var id, persisted string
	var shared bool
	if err := database.QueryRow("SELECT hwid, send_hwid FROM subscriptions").Scan(&id, &shared); err != nil || len(id) != 32 || shared {
		t.Fatalf("existing subscription must get a random ID with sharing off: %v", err)
	}
	var value string
	if err := database.QueryRow("SELECT node_dns FROM subscriptions").Scan(&value); err != nil || value != "auto" {
		t.Fatalf("migrated node DNS = %q, %v", value, err)
	}
	if _, err := database.Exec("UPDATE subscriptions SET node_dns = 'tls://1.1.1.1:853'"); err != nil {
		t.Fatal(err)
	}
	if err := MigrateSchema(database); err != nil {
		t.Fatal(err)
	}
	if err := database.QueryRow("SELECT hwid FROM subscriptions").Scan(&persisted); err != nil || persisted != id {
		t.Fatalf("migration changed the subscription identifier: %v", err)
	}
	if err := database.QueryRow("SELECT node_dns FROM subscriptions").Scan(&value); err != nil || value != "tls://1.1.1.1:853" {
		t.Fatalf("second migration overwrote node DNS = %q, %v", value, err)
	}
}

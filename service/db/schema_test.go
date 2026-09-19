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

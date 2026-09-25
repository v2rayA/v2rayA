package db

import (
	"database/sql"
	"testing"
)

func TestSubscriptionPolicySchemaUpgrade(t *testing.T) {
	database, err := sql.Open(sqliteDriverName, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	database.SetMaxOpenConns(1)
	if _, err := database.Exec(`CREATE TABLE subscriptions (id INTEGER PRIMARY KEY, address TEXT, remarks TEXT, auto_select INTEGER); INSERT INTO subscriptions VALUES (7, 'https://example.test/sub', 'keep me', 1)`); err != nil {
		t.Fatal(err)
	}
	if err := MigrateSchema(database); err != nil {
		t.Fatal(err)
	}
	var enabled, regular, failure int
	var address, remarks string
	if err := database.QueryRow("SELECT address, remarks, auto_update, update_interval_minutes, failure_interval_minutes FROM subscriptions WHERE id=7").Scan(&address, &remarks, &enabled, &regular, &failure); err != nil {
		t.Fatal(err)
	}
	if enabled != 0 || regular != 0 || failure != 1 || address != "https://example.test/sub" || remarks != "keep me" {
		t.Fatalf("migration changed existing subscription: %s %s %d %d", address, remarks, regular, failure)
	}
	if _, err := database.Exec("UPDATE subscriptions SET auto_update=1, update_interval_minutes=120, failure_interval_minutes=2 WHERE id=7"); err != nil {
		t.Fatal(err)
	}
	if err := MigrateSchema(database); err != nil {
		t.Fatal(err)
	}
	if err := database.QueryRow("SELECT auto_update, update_interval_minutes, failure_interval_minutes FROM subscriptions WHERE id=7").Scan(&enabled, &regular, &failure); err != nil {
		t.Fatal(err)
	}
	if enabled != 1 || regular != 120 || failure != 2 {
		t.Fatal("repeated migration reset saved policies")
	}
}

func TestBoltSubscriptionPolicyMigration(t *testing.T) {
	database, err := sql.Open(sqliteDriverName, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	database.SetMaxOpenConns(1)
	if err := InitSchema(database); err != nil {
		t.Fatal(err)
	}
	tx, err := database.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	if err := migrateSubscriptions([]byte(`[{"address":"https://example.test/old-fork","monitor":true,"preferFirst":true,"autoSelect":true},{"address":"https://example.test/upstream"}]`), tx); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	rows, err := database.Query("SELECT auto_update, update_interval_minutes, failure_interval_minutes FROM subscriptions ORDER BY sort")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var monitor, first, auto int
		if err := rows.Scan(&monitor, &first, &auto); err != nil {
			t.Fatal(err)
		}
		if monitor != 0 || first != 0 || auto != 1 {
			t.Fatalf("legacy flags enabled new automation: %d %d %d", monitor, first, auto)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
}

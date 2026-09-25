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
	var monitor, first int
	var address, remarks string
	if err := database.QueryRow("SELECT address, remarks, monitor, prefer_first FROM subscriptions WHERE id=7").Scan(&address, &remarks, &monitor, &first); err != nil {
		t.Fatal(err)
	}
	if monitor != 0 || first != 0 || address != "https://example.test/sub" || remarks != "keep me" {
		t.Fatalf("migration changed existing subscription: %s %s %d %d", address, remarks, monitor, first)
	}
	if _, err := database.Exec("UPDATE subscriptions SET monitor=1, prefer_first=1 WHERE id=7"); err != nil {
		t.Fatal(err)
	}
	if err := MigrateSchema(database); err != nil {
		t.Fatal(err)
	}
	if err := database.QueryRow("SELECT monitor, prefer_first FROM subscriptions WHERE id=7").Scan(&monitor, &first); err != nil {
		t.Fatal(err)
	}
	if monitor != 1 || first != 1 {
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
	rows, err := database.Query("SELECT monitor, prefer_first, auto_select FROM subscriptions ORDER BY sort")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for i := 0; rows.Next(); i++ {
		var monitor, first, auto int
		if err := rows.Scan(&monitor, &first, &auto); err != nil {
			t.Fatal(err)
		}
		expected := 0
		if i == 0 {
			expected = 1
		}
		if monitor != expected || first != expected || auto != expected {
			t.Fatalf("subscription %d: policy lost or default enabled: %d %d %d", i, monitor, first, auto)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
}

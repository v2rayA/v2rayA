package db

import (
	"database/sql"
	"path/filepath"
	"testing"

	"go.etcd.io/bbolt"
)

func openSchemaTestDB(t *testing.T) *sql.DB {
	t.Helper()
	if sqliteDriverName == "" {
		t.Skip("SQLite driver is unavailable")
	}
	database, err := sql.Open(sqliteDriverName, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = database.Close() })
	return database
}

func TestMigrateSchemaCopiesLegacySubscriptionUpdatePolicy(t *testing.T) {
	database := openSchemaTestDB(t)
	statements := []string{
		`CREATE TABLE system_config (key TEXT PRIMARY KEY, value TEXT NOT NULL)`,
		`CREATE TABLE subscriptions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			address TEXT NOT NULL DEFAULT '',
			remarks TEXT NOT NULL DEFAULT '',
			status TEXT NOT NULL DEFAULT '',
			info TEXT DEFAULT '',
			auto_select INTEGER NOT NULL DEFAULT 0,
			sort INTEGER NOT NULL DEFAULT 0
		)`,
		`INSERT INTO system_config (key, value) VALUES (
			'system:setting',
			'{"subscriptionAutoUpdateMode":"auto_update_at_intervals","subscriptionAutoUpdateIntervalHour":12}'
		)`,
		`INSERT INTO subscriptions (address) VALUES ('https://example.invalid/subscription')`,
	}
	for _, statement := range statements {
		if _, err := database.Exec(statement); err != nil {
			t.Fatal(err)
		}
	}

	if err := MigrateSchema(database); err != nil {
		t.Fatal(err)
	}

	var mode string
	var intervalHour int
	var lastAttempt sql.NullString
	if err := database.QueryRow(
		`SELECT auto_update_mode, auto_update_interval_hour, last_update_attempt_at FROM subscriptions`,
	).Scan(&mode, &intervalHour, &lastAttempt); err != nil {
		t.Fatal(err)
	}
	if mode != "auto_update_at_intervals" || intervalHour != 12 {
		t.Fatalf("migrated policy = (%q, %d), want (%q, %d)", mode, intervalHour, "auto_update_at_intervals", 12)
	}
	if lastAttempt.Valid {
		t.Fatalf("last update attempt = %q, want NULL", lastAttempt.String)
	}

	if _, err := database.Exec(
		`UPDATE subscriptions SET auto_update_mode = 'auto_update', auto_update_interval_hour = 0`,
	); err != nil {
		t.Fatal(err)
	}
	if err := MigrateSchema(database); err != nil {
		t.Fatal(err)
	}
	if err := database.QueryRow(
		`SELECT auto_update_mode, auto_update_interval_hour FROM subscriptions`,
	).Scan(&mode, &intervalHour); err != nil {
		t.Fatal(err)
	}
	if mode != "auto_update" || intervalHour != 0 {
		t.Fatalf("policy after second migration = (%q, %d), want (%q, %d)", mode, intervalHour, "auto_update", 0)
	}
}

func TestSubscriptionUpdatePolicyDefaultsToDisabled(t *testing.T) {
	database := openSchemaTestDB(t)
	if err := InitSchema(database); err != nil {
		t.Fatal(err)
	}
	if err := MigrateSchema(database); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO subscriptions (address) VALUES ('https://example.invalid/subscription')`); err != nil {
		t.Fatal(err)
	}

	var mode string
	var intervalHour int
	if err := database.QueryRow(
		`SELECT auto_update_mode, auto_update_interval_hour FROM subscriptions`,
	).Scan(&mode, &intervalHour); err != nil {
		t.Fatal(err)
	}
	if mode != "none" || intervalHour != 0 {
		t.Fatalf("default policy = (%q, %d), want (%q, %d)", mode, intervalHour, "none", 0)
	}
}

func TestBoltMigrationCopiesLegacySubscriptionUpdatePolicy(t *testing.T) {
	database := openSchemaTestDB(t)
	if err := InitSchema(database); err != nil {
		t.Fatal(err)
	}

	boltDB, err := bbolt.Open(filepath.Join(t.TempDir(), "bolt.db"), 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = boltDB.Close() })
	if err := boltDB.Update(func(tx *bbolt.Tx) error {
		system, err := tx.CreateBucket([]byte("system"))
		if err != nil {
			return err
		}
		if err := system.Put(
			[]byte("setting"),
			[]byte(`{"subscriptionAutoUpdateMode":"auto_update_at_intervals","subscriptionAutoUpdateIntervalHour":6}`),
		); err != nil {
			return err
		}
		touch, err := tx.CreateBucket([]byte("touch"))
		if err != nil {
			return err
		}
		return touch.Put(
			[]byte("subscriptions"),
			[]byte(`[{"address":"https://example.invalid/subscription","servers":[]}]`),
		)
	}); err != nil {
		t.Fatal(err)
	}

	tx, err := database.Begin()
	if err != nil {
		t.Fatal(err)
	}
	if err := migrateSystemBucket(boltDB, tx); err != nil {
		_ = tx.Rollback()
		t.Fatal(err)
	}
	if _, err := migrateTouchBucket(boltDB, tx); err != nil {
		_ = tx.Rollback()
		t.Fatal(err)
	}
	if err := migrateLegacySubscriptionUpdatePolicy(tx); err != nil {
		_ = tx.Rollback()
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	var mode string
	var intervalHour int
	if err := database.QueryRow(
		`SELECT auto_update_mode, auto_update_interval_hour FROM subscriptions`,
	).Scan(&mode, &intervalHour); err != nil {
		t.Fatal(err)
	}
	if mode != "auto_update_at_intervals" || intervalHour != 6 {
		t.Fatalf("migrated policy = (%q, %d), want (%q, %d)", mode, intervalHour, "auto_update_at_intervals", 6)
	}
}

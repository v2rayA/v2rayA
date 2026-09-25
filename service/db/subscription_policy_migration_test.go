package db

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tidwall/gjson"
	"go.etcd.io/bbolt"
)

func TestSQLiteSubscriptionUpdatePolicyMigration(t *testing.T) {
	tests := []struct {
		legacy, want string
		minutes      int
	}{
		{"none", "disabled", 0},
		{"auto_update", "on_start", 0},
		{"auto_update_at_intervals", "at_interval", 420},
	}
	for _, test := range tests {
		t.Run(test.legacy, func(t *testing.T) {
			database, err := sql.Open(sqliteDriverName, ":memory:")
			if err != nil {
				t.Fatal(err)
			}
			defer database.Close()
			database.SetMaxOpenConns(1)
			if _, err := database.Exec(`
				CREATE TABLE system_config (key TEXT PRIMARY KEY, value TEXT NOT NULL);
				CREATE TABLE subscriptions (id INTEGER PRIMARY KEY, address TEXT, remarks TEXT, auto_select INTEGER);
				INSERT INTO subscriptions VALUES (7, 'https://example.test/sub', 'keep me', 1);
				INSERT INTO system_config VALUES ('system:setting', ?)
			`, `{"subscriptionAutoUpdateMode":"`+test.legacy+`","subscriptionAutoUpdateIntervalHour":7}`); err != nil {
				t.Fatal(err)
			}
			if err := MigrateSchema(database); err != nil {
				t.Fatal(err)
			}
			assertSubscriptionPolicy(t, database, test.want, test.minutes, 0)
			assertAutomaticProxySetting(t, database, true, "300s")

			if _, err := database.Exec(`
				UPDATE subscriptions SET update_mode='interval_failsafe', update_interval_minutes=11, failure_interval_minutes=2;
				UPDATE system_config SET value='{"probeURL":"https://example.test/check","probeInterval":"90s","type":"leastping","autoAdd":false}' WHERE key='outbound.proxy:setting'
			`); err != nil {
				t.Fatal(err)
			}
			if err := MigrateSchema(database); err != nil {
				t.Fatal(err)
			}
			assertSubscriptionPolicyValues(t, database, "interval_failsafe", 11, 2, 0)
			assertAutomaticProxySetting(t, database, false, "90s")
		})
	}
}

func TestBoltImportMigratesAllSubscriptionUpdateModes(t *testing.T) {
	for _, legacy := range []string{"none", "auto_update", "auto_update_at_intervals"} {
		t.Run(legacy, func(t *testing.T) {
			boltPath := filepath.Join(t.TempDir(), "bolt.db")
			bolt, err := bbolt.Open(boltPath, 0600, nil)
			if err != nil {
				t.Fatal(err)
			}
			if err := bolt.Update(func(tx *bbolt.Tx) error {
				system, err := tx.CreateBucket([]byte("system"))
				if err != nil {
					return err
				}
				if err := system.Put([]byte("setting"), []byte(`{"subscriptionAutoUpdateMode":"`+legacy+`","subscriptionAutoUpdateIntervalHour":7}`)); err != nil {
					return err
				}
				touch, err := tx.CreateBucket([]byte("touch"))
				if err != nil {
					return err
				}
				return touch.Put([]byte("subscriptions"), []byte(`[{"address":"https://example.test/sub","autoSelect":true,"servers":[]}]`))
			}); err != nil {
				t.Fatal(err)
			}
			defer bolt.Close()

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
			if err := migrateSystemBucket(bolt, tx); err != nil {
				t.Fatal(err)
			}
			if err := migrateTouchBucket(bolt, tx); err != nil {
				t.Fatal(err)
			}
			if err := migrateSubscriptionUpdatePolicy(tx); err != nil {
				t.Fatal(err)
			}
			if err := migrateAutoSelectToAutomaticGroup(tx); err != nil {
				t.Fatal(err)
			}
			if err := tx.Commit(); err != nil {
				t.Fatal(err)
			}

			want, minutes := "disabled", 0
			if legacy == "auto_update" {
				want = "on_start"
			} else if legacy == "auto_update_at_intervals" {
				want, minutes = "at_interval", 420
			}
			assertSubscriptionPolicy(t, database, want, minutes, 0)
			assertAutomaticProxySetting(t, database, true, "300s")
			var setting string
			if err := database.QueryRow("SELECT value FROM system_config WHERE key='system:setting'").Scan(&setting); err != nil {
				t.Fatal(err)
			}
			if got := gjson.Get(setting, "subscriptionAutoUpdateMode").String(); got != "none" {
				t.Fatalf("legacy mode was not reset: %q", got)
			}
		})
	}
}

func TestAutoSelectMigrationLeavesManualProxyAloneWhenUnused(t *testing.T) {
	database, err := sql.Open(sqliteDriverName, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	database.SetMaxOpenConns(1)
	if _, err := database.Exec(`
		CREATE TABLE system_config (key TEXT PRIMARY KEY, value TEXT NOT NULL);
		CREATE TABLE subscriptions (id INTEGER PRIMARY KEY, address TEXT, remarks TEXT, auto_select INTEGER);
		INSERT INTO subscriptions VALUES (1, 'https://example.test/sub', '', 0);
		INSERT INTO system_config VALUES ('outbound.proxy:setting', '{"probeURL":"https://example.test/check","probeInterval":"60s","type":"leastping"}')
	`); err != nil {
		t.Fatal(err)
	}
	if err := MigrateSchema(database); err != nil {
		t.Fatal(err)
	}
	assertAutomaticProxySetting(t, database, false, "60s")
}

func assertAutomaticProxySetting(t *testing.T, database *sql.DB, enabled bool, interval string) {
	t.Helper()
	var setting string
	if err := database.QueryRow("SELECT value FROM system_config WHERE key='outbound.proxy:setting'").Scan(&setting); err != nil {
		t.Fatal(err)
	}
	if got := gjson.Get(setting, "autoAdd").Bool(); got != enabled {
		t.Fatalf("autoAdd = %v; want %v (%s)", got, enabled, setting)
	}
	if got := gjson.Get(setting, "probeInterval").String(); got != interval {
		t.Fatalf("probeInterval = %q; want %q (%s)", got, interval, setting)
	}
}

func assertSubscriptionPolicy(t *testing.T, database *sql.DB, mode string, minutes, autoSelect int) {
	assertSubscriptionPolicyValues(t, database, mode, minutes, 1, autoSelect)
	var direct bool
	if err := database.QueryRow("SELECT allow_direct_recovery FROM subscriptions").Scan(&direct); err != nil {
		t.Fatal(err)
	}
	if direct {
		t.Fatal("legacy migration enabled direct downloads without consent")
	}
}

func TestDirectRecoveryMigrationDoesNotChangeExistingPolicy(t *testing.T) {
	database, err := sql.Open(sqliteDriverName, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	database.SetMaxOpenConns(1)
	oldSchema := strings.Replace(schemaSQL, "    allow_direct_recovery INTEGER NOT NULL DEFAULT 0,\n", "", 1)
	if _, err := database.Exec(oldSchema); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec("INSERT INTO subscriptions (address, update_mode, update_interval_minutes, failure_interval_minutes) VALUES ('https://example.test/sub', 'interval_failsafe', 10, 2)"); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec("INSERT INTO system_config (key, value) VALUES (?, '1')", subscriptionUpdateMigrationKey); err != nil {
		t.Fatal(err)
	}
	if err := MigrateSchema(database); err != nil {
		t.Fatal(err)
	}
	assertSubscriptionPolicyValues(t, database, "interval_failsafe", 10, 2, 0)
	var allowed bool
	if err := database.QueryRow("SELECT allow_direct_recovery FROM subscriptions").Scan(&allowed); err != nil || allowed {
		t.Fatalf("upgrade opted in to direct recovery: allowed=%v err=%v", allowed, err)
	}
	if _, err := database.Exec("UPDATE subscriptions SET allow_direct_recovery = 1"); err != nil {
		t.Fatal(err)
	}
	if err := MigrateSchema(database); err != nil {
		t.Fatal(err)
	}
	if err := database.QueryRow("SELECT allow_direct_recovery FROM subscriptions").Scan(&allowed); err != nil || !allowed {
		t.Fatalf("restart reset explicit consent: allowed=%v err=%v", allowed, err)
	}
}

func assertSubscriptionPolicyValues(t *testing.T, database *sql.DB, mode string, minutes, failure, autoSelect int) {
	t.Helper()
	var gotMode string
	var gotMinutes, gotFailure, gotAutoSelect int
	if err := database.QueryRow("SELECT update_mode, update_interval_minutes, failure_interval_minutes, auto_select FROM subscriptions").Scan(&gotMode, &gotMinutes, &gotFailure, &gotAutoSelect); err != nil {
		t.Fatal(err)
	}
	if gotMode != mode || gotMinutes != minutes || gotFailure != failure || gotAutoSelect != autoSelect {
		t.Fatalf("policy = %s/%d/%d, autoSelect=%d; want %s/%d/%d, autoSelect=%d", gotMode, gotMinutes, gotFailure, gotAutoSelect, mode, minutes, failure, autoSelect)
	}
}

package db

import (
	"database/sql"
	"testing"
)

func TestAutoSelectMigrationPreservesManualMembers(t *testing.T) {
	for _, tt := range []struct {
		name, refs string
		wantAuto   bool
	}{
		{"empty", `{"touches":[]}`, true},
		{"automatic-only", `{"touches":[{"_type":"subscriptionServer","id":1,"sub":1}]}`, true},
		{"standalone", `{"touches":[{"_type":"server","id":1}]}`, false},
		{"other-subscription", `{"touches":[{"_type":"subscriptionServer","id":1,"sub":0}]}`, false},
		{"mixed", `{"touches":[{"_type":"subscriptionServer","id":1,"sub":1},{"_type":"server","id":1}]}`, false},
		{"stale-subscription", `{"touches":[{"_type":"subscriptionServer","id":1,"sub":9}]}`, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			database, err := sql.Open(sqliteDriverName, ":memory:")
			if err != nil {
				t.Fatal(err)
			}
			defer database.Close()
			database.SetMaxOpenConns(1)
			if err := InitSchema(database); err != nil {
				t.Fatal(err)
			}
			// Database IDs and sort values deliberately differ from catalog ordinals.
			if _, err := database.Exec(`
				INSERT INTO subscriptions (id, auto_select, sort) VALUES (7, 1, 30), (9, 0, 10);
				INSERT INTO system_config VALUES ('outbound.proxy:setting', '{"probeURL":"https://example.test/check","probeInterval":"60s","type":"leastping","selected":"keep"}');
				INSERT INTO system_config VALUES ('outbound.proxy:connectedServers', ?)
			`, tt.refs); err != nil {
				t.Fatal(err)
			}
			if err := MigrateSchema(database); err != nil {
				t.Fatal(err)
			}
			interval := "60s"
			if tt.wantAuto {
				interval = "300s"
			}
			assertAutomaticProxySetting(t, database, tt.wantAuto, interval)
			var refs string
			if err := database.QueryRow("SELECT value FROM system_config WHERE key='outbound.proxy:connectedServers'").Scan(&refs); err != nil || refs != tt.refs {
				t.Fatalf("migration changed membership: %s, err=%v", refs, err)
			}
			var enabled int
			if err := database.QueryRow("SELECT COUNT(*) FROM subscriptions WHERE auto_select != 0").Scan(&enabled); err != nil || enabled != 0 {
				t.Fatalf("legacy auto-select remains enabled: %d, err=%v", enabled, err)
			}
			if _, err := database.Exec("UPDATE system_config SET value='{\"autoAdd\":false,\"probeInterval\":\"90s\"}' WHERE key='outbound.proxy:setting'"); err != nil {
				t.Fatal(err)
			}
			if err := MigrateSchema(database); err != nil {
				t.Fatal(err)
			}
			assertAutomaticProxySetting(t, database, false, "90s")
		})
	}
}

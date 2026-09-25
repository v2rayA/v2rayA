package db

import (
	"database/sql"
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

const subscriptionUpdateMigrationKey = "migration:subscription_update_policy_v1"

type subscriptionPolicyStore interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

func migrateSubscriptionUpdatePolicy(store subscriptionPolicyStore) error {
	var migrated int
	if err := store.QueryRow("SELECT COUNT(*) FROM system_config WHERE key = ?", subscriptionUpdateMigrationKey).Scan(&migrated); err != nil {
		return err
	}
	if migrated != 0 {
		return nil
	}

	var setting string
	err := store.QueryRow("SELECT value FROM system_config WHERE key = 'system:setting'").Scan(&setting)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	legacyMode := gjson.Get(setting, "subscriptionAutoUpdateMode").String()
	hours := int(gjson.Get(setting, "subscriptionAutoUpdateIntervalHour").Int())
	mode, minutes := "", 0
	switch legacyMode {
	case "auto_update":
		mode = "on_start"
	case "auto_update_at_intervals":
		mode = "at_interval"
		if hours <= 0 {
			hours = 1
		}
		minutes = hours * 60
	}
	if mode != "" {
		if _, err := store.Exec("UPDATE subscriptions SET update_mode = ?, update_interval_minutes = ? WHERE update_mode = 'disabled'", mode, minutes); err != nil {
			return err
		}
	}
	if setting != "" {
		setting, err = sjson.Set(setting, "subscriptionAutoUpdateMode", "none")
		if err != nil {
			return fmt.Errorf("reset legacy subscription update mode: %w", err)
		}
		if _, err := store.Exec("UPDATE system_config SET value = ? WHERE key = 'system:setting'", setting); err != nil {
			return err
		}
	}
	if _, err := store.Exec("INSERT INTO system_config (key, value) VALUES (?, '1')", subscriptionUpdateMigrationKey); err != nil {
		return err
	}
	log.Info("Migrated legacy subscription update mode %q to per-subscription policies", legacyMode)
	return nil
}

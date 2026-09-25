package db

import (
	"database/sql"
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

const autoSelectGroupMigrationKey = "migration:auto_select_group_v1"

func migrateAutoSelectToAutomaticGroup(store subscriptionPolicyStore) error {
	var migrated int
	if err := store.QueryRow("SELECT COUNT(*) FROM system_config WHERE key = ?", autoSelectGroupMigrationKey).Scan(&migrated); err != nil {
		return err
	}
	if migrated != 0 {
		return nil
	}

	rows, err := store.Query("SELECT id FROM subscriptions WHERE auto_select != 0 ORDER BY id")
	if err != nil {
		return err
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return err
		}
		count++
		log.Info("Migrating subscription %d from auto-select to automatic PROXY membership", id)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}

	if count > 0 {
		const settingKey = "outbound.proxy:setting"
		setting := `{"probeURL":"https://www.gstatic.com/generate_204","probeInterval":"300s","type":"leastping"}`
		if err := store.QueryRow("SELECT value FROM system_config WHERE key = ?", settingKey).Scan(&setting); err != nil && err != sql.ErrNoRows {
			return err
		}
		if gjson.Get(setting, "probeURL").String() == "" {
			setting, err = sjson.Set(setting, "probeURL", "https://www.gstatic.com/generate_204")
			if err != nil {
				return fmt.Errorf("set automatic PROXY probe URL: %w", err)
			}
		}
		if gjson.Get(setting, "probeInterval").String() == "" || gjson.Get(setting, "probeInterval").String() == "60s" {
			setting, err = sjson.Set(setting, "probeInterval", "300s")
			if err != nil {
				return fmt.Errorf("set automatic PROXY interval: %w", err)
			}
		}
		if gjson.Get(setting, "type").String() == "" {
			setting, err = sjson.Set(setting, "type", "leastping")
			if err != nil {
				return fmt.Errorf("set automatic PROXY type: %w", err)
			}
		}
		setting, err = sjson.Set(setting, "autoAdd", true)
		if err != nil {
			return fmt.Errorf("enable automatic PROXY membership: %w", err)
		}
		if _, err := store.Exec("INSERT OR REPLACE INTO system_config (key, value) VALUES (?, ?)", settingKey, setting); err != nil {
			return err
		}
		if _, err := store.Exec("UPDATE subscriptions SET auto_select = 0 WHERE auto_select != 0"); err != nil {
			return err
		}
	}
	if _, err := store.Exec("INSERT INTO system_config (key, value) VALUES (?, '1')", autoSelectGroupMigrationKey); err != nil {
		return err
	}
	return nil
}

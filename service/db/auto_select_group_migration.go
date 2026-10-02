package db

import (
	"database/sql"
	"encoding/json"
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

	rows, err := store.Query("SELECT auto_select FROM subscriptions ORDER BY sort")
	if err != nil {
		return err
	}
	defer rows.Close()
	count := 0
	autoSelect := []bool{}
	for rows.Next() {
		var enabled int
		if err := rows.Scan(&enabled); err != nil {
			return err
		}
		autoSelect = append(autoSelect, enabled != 0)
		if enabled != 0 {
			count++
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if err := rows.Close(); err != nil {
		return err
	}

	if count > 0 {
		manual, err := proxyHasManualMembers(store, autoSelect)
		if err != nil {
			return err
		}
		const settingKey = "outbound.proxy:setting"
		setting := `{"probeURL":"https://www.gstatic.com/generate_204","probeInterval":"300s","type":"leastping"}`
		if err := store.QueryRow("SELECT value FROM system_config WHERE key = ?", settingKey).Scan(&setting); err != nil && err != sql.ErrNoRows {
			return err
		}
		if !manual && gjson.Get(setting, "probeURL").String() == "" {
			setting, err = sjson.Set(setting, "probeURL", "https://www.gstatic.com/generate_204")
			if err != nil {
				return fmt.Errorf("set automatic PROXY probe URL: %w", err)
			}
		}
		if !manual && (gjson.Get(setting, "probeInterval").String() == "" || gjson.Get(setting, "probeInterval").String() == "60s") {
			setting, err = sjson.Set(setting, "probeInterval", "300s")
			if err != nil {
				return fmt.Errorf("set automatic PROXY interval: %w", err)
			}
		}
		if !manual && gjson.Get(setting, "type").String() == "" {
			setting, err = sjson.Set(setting, "type", "leastping")
			if err != nil {
				return fmt.Errorf("set automatic PROXY type: %w", err)
			}
		}
		setting, err = sjson.Set(setting, "autoAdd", !manual)
		if err != nil {
			return fmt.Errorf("migrate automatic PROXY membership: %w", err)
		}
		if _, err := store.Exec("INSERT OR REPLACE INTO system_config (key, value) VALUES (?, ?)", settingKey, setting); err != nil {
			return err
		}
		if _, err := store.Exec("UPDATE subscriptions SET auto_select = 0 WHERE auto_select != 0"); err != nil {
			return err
		}
		if manual {
			log.Info("Retired subscription auto-select; PROXY has manual members outside those subscriptions, so automatic membership remains off. Enable 'Automatically add available servers' in PROXY settings to use the whole catalog.")
		} else {
			log.Info("Migrated subscription auto-select to automatic PROXY membership using the whole catalog")
		}
	}
	if _, err := store.Exec("INSERT INTO system_config (key, value) VALUES (?, '1')", autoSelectGroupMigrationKey); err != nil {
		return err
	}
	return nil
}

// Sub is an ordinal in the catalog's sort order, not a subscription database ID.
func proxyHasManualMembers(store subscriptionPolicyStore, autoSelect []bool) (bool, error) {
	var raw string
	if err := store.QueryRow("SELECT value FROM system_config WHERE key = ?", "outbound.proxy:connectedServers").Scan(&raw); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	var connections struct {
		Touches []*struct {
			Type string `json:"_type"`
			Sub  int    `json:"sub"`
		} `json:"touches"`
	}
	if err := json.Unmarshal([]byte(raw), &connections); err != nil {
		return false, fmt.Errorf("read PROXY members for auto-select migration: %w", err)
	}
	for _, ref := range connections.Touches {
		if ref == nil || ref.Type != "subscriptionServer" || ref.Sub < 0 || ref.Sub >= len(autoSelect) || !autoSelect[ref.Sub] {
			return true, nil
		}
	}
	return false, nil
}

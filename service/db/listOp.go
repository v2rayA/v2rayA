package db

import (
	"database/sql"
	"encoding/json"
	"fmt"

	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
)

type subscriptionListValue struct {
	DatabaseID             int64             `json:"_databaseId"`
	Remarks                string            `json:"remarks,omitempty"`
	Address                string            `json:"address"`
	Status                 string            `json:"status"`
	Info                   string            `json:"info"`
	Servers                []json.RawMessage `json:"servers"`
	AutoSelect             bool              `json:"autoSelect"`
	AutoUpdateMode         string            `json:"autoUpdateMode"`
	AutoUpdateIntervalHour int               `json:"autoUpdateIntervalHour"`
	LastUpdateAttemptAt    string            `json:"lastUpdateAttemptAt,omitempty"`
}

func marshalSubscriptionListValue(id int64, remarks, address, status, info string, servers []string,
	autoSelect bool, autoUpdateMode string, autoUpdateIntervalHour int, lastUpdateAttemptAt string,
) ([]byte, error) {
	rawServers := make([]json.RawMessage, len(servers))
	for i, server := range servers {
		rawServers[i] = json.RawMessage(server)
	}
	return json.Marshal(subscriptionListValue{
		DatabaseID:             id,
		Remarks:                remarks,
		Address:                address,
		Status:                 status,
		Info:                   info,
		Servers:                rawServers,
		AutoSelect:             autoSelect,
		AutoUpdateMode:         autoUpdateMode,
		AutoUpdateIntervalHour: autoUpdateIntervalHour,
		LastUpdateAttemptAt:    lastUpdateAttemptAt,
	})
}

// ListSet sets an element at a specific index in a list.
func ListSet(bucket string, key string, index int, val interface{}) (err error) {
	db := GetDB()

	switch bucket + "/" + key {
	case "touch/servers":
		b, err := jsoniter.Marshal(val)
		if err != nil {
			return err
		}
		result, err := db.Exec(
			"UPDATE servers SET config_json = ?, updated_at = CURRENT_TIMESTAMP WHERE type = 'server' AND sort = ?",
			string(b), index,
		)
		if err != nil {
			return err
		}
		rows, _ := result.RowsAffected()
		if rows == 0 {
			return fmt.Errorf("ListSet: server at index %d not found", index)
		}
		return nil

	case "touch/subscriptions":
		b, err := jsoniter.Marshal(val)
		if err != nil {
			return err
		}
		parsed := gjson.ParseBytes(b)
		var subID int64
		if err := db.QueryRow("SELECT id FROM subscriptions WHERE sort = ?", index).Scan(&subID); err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("ListSet: subscription at index %d not found", index)
			}
			return err
		}
		address := parsed.Get("address").String()
		remarks := parsed.Get("remarks").String()
		status := parsed.Get("status").String()
		info := parsed.Get("info").String()
		autoSelect := 0
		if parsed.Get("autoSelect").Bool() {
			autoSelect = 1
		}
		autoUpdateMode := parsed.Get("autoUpdateMode").String()
		if autoUpdateMode == "" {
			autoUpdateMode = "none"
		}
		autoUpdateIntervalHour := int(parsed.Get("autoUpdateIntervalHour").Int())

		result, err := db.Exec(
			"UPDATE subscriptions SET address = ?, remarks = ?, status = ?, info = ?, auto_select = ?, auto_update_mode = ?, auto_update_interval_hour = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			address, remarks, status, info, autoSelect, autoUpdateMode, autoUpdateIntervalHour, subID,
		)
		if err != nil {
			return err
		}
		rows, _ := result.RowsAffected()
		if rows == 0 {
			return fmt.Errorf("ListSet: subscription at index %d not found", index)
		}

		// Update servers within this subscription.
		// Clean up outbound_connections first to satisfy foreign key constraint;
		// otherwise the delete fails and the insert below duplicates the list.
		if _, err := db.Exec(`
			DELETE FROM outbound_connections
			WHERE server_id IN (SELECT id FROM servers WHERE type = 'subscription_server' AND sub_id = ?)
		`, subID); err != nil {
			return fmt.Errorf("ListSet: failed to clear outbound connections of subscription %d: %w", index, err)
		}
		if _, err := db.Exec("DELETE FROM servers WHERE type = 'subscription_server' AND sub_id = ?", subID); err != nil {
			return fmt.Errorf("ListSet: failed to clear old servers of subscription %d: %w", index, err)
		}

		servers := parsed.Get("servers").Array()
		for j, s := range servers {
			_, err := db.Exec(
				"INSERT INTO servers (type, sub_id, config_json, sort) VALUES ('subscription_server', ?, ?, ?)",
				subID, s.Raw, j,
			)
			if err != nil {
				return fmt.Errorf("ListSet: failed to update subscription server %d/%d: %w", index, j, err)
			}
		}
		return nil

	default:
		return fmt.Errorf("ListSet: unsupported bucket/key: %s/%s", bucket, key)
	}
}

// MarkSubscriptionUpdateAttempt records an update attempt regardless of its outcome.
func MarkSubscriptionUpdateAttempt(databaseID int64) error {
	result, err := GetDB().Exec(
		"UPDATE subscriptions SET last_update_attempt_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now') WHERE id = ?",
		databaseID,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("MarkSubscriptionUpdateAttempt: subscription %d not found", databaseID)
	}
	return nil
}

func GetSubscriptionIndexByID(databaseID int64) (int, error) {
	var index int
	err := GetDB().QueryRow("SELECT sort FROM subscriptions WHERE id = ?", databaseID).Scan(&index)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("subscription %d not found", databaseID)
	}
	return index, err
}

// ListGet retrieves an element at a specific index from a list.
func ListGet(bucket string, key string, index int) (b []byte, err error) {
	db := GetDB()

	switch bucket + "/" + key {
	case "touch/servers":
		var configJSON string
		err = db.QueryRow(
			"SELECT config_json FROM servers WHERE type = 'server' AND sort = ?", index,
		).Scan(&configJSON)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("ListGet: can't get element from an empty list")
		}
		if err != nil {
			return nil, err
		}
		return []byte(configJSON), nil

	case "touch/subscriptions":
		return listGetSubscription(db, index)

	default:
		return nil, fmt.Errorf("ListGet: unsupported bucket/key: %s/%s", bucket, key)
	}
}

func listGetSubscription(db *sql.DB, index int) ([]byte, error) {
	var id int64
	var address, remarks, status, info, autoUpdateMode string
	var autoSelectInt, autoUpdateIntervalHour int
	var lastUpdateAttemptAt sql.NullString
	err := db.QueryRow(
		"SELECT id, address, remarks, status, info, auto_select, auto_update_mode, auto_update_interval_hour, last_update_attempt_at FROM subscriptions WHERE sort = ?", index,
	).Scan(&id, &address, &remarks, &status, &info, &autoSelectInt, &autoUpdateMode, &autoUpdateIntervalHour, &lastUpdateAttemptAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("ListGet: can't get element from an empty list")
	}
	if err != nil {
		return nil, err
	}

	// Reconstruct the subscription JSON with servers
	rows, err := db.Query(
		"SELECT config_json FROM servers WHERE type = 'subscription_server' AND sub_id = ? ORDER BY sort",
		id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var servers []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		servers = append(servers, s)
	}
	autoSelect := autoSelectInt != 0
	return marshalSubscriptionListValue(
		id, remarks, address, status, info, servers, autoSelect,
		autoUpdateMode, autoUpdateIntervalHour, lastUpdateAttemptAt.String,
	)
}

// ListAppend appends values to a list.
func ListAppend(bucket string, key string, val interface{}) (err error) {
	db := GetDB()

	switch bucket + "/" + key {
	case "touch/servers":
		// Marshal the value to JSON first, then parse as array
		b, err := jsoniter.Marshal(val)
		if err != nil {
			return err
		}
		parsed := gjson.ParseBytes(b)
		if parsed.IsArray() {
			for _, item := range parsed.Array() {
				var maxSort int
				db.QueryRow("SELECT COALESCE(MAX(sort), -1) FROM servers WHERE type = 'server'").Scan(&maxSort)

				_, err = db.Exec(
					"INSERT INTO servers (type, config_json, sort) VALUES ('server', ?, ?)",
					item.Raw, maxSort+1,
				)
				if err != nil {
					return err
				}
			}
		} else {
			var maxSort int
			db.QueryRow("SELECT COALESCE(MAX(sort), -1) FROM servers WHERE type = 'server'").Scan(&maxSort)
			_, err = db.Exec(
				"INSERT INTO servers (type, config_json, sort) VALUES ('server', ?, ?)",
				string(b), maxSort+1,
			)
			if err != nil {
				return err
			}
		}
		return nil

	case "touch/subscriptions":
		b, err := jsoniter.Marshal(val)
		if err != nil {
			return err
		}
		parsed := gjson.ParseBytes(b)
		if parsed.IsArray() {
			for _, item := range parsed.Array() {
				address := item.Get("address").String()
				remarks := item.Get("remarks").String()
				status := item.Get("status").String()
				info := item.Get("info").String()
				autoSelect := 0
				if item.Get("autoSelect").Bool() {
					autoSelect = 1
				}
				autoUpdateMode := item.Get("autoUpdateMode").String()
				if autoUpdateMode == "" {
					autoUpdateMode = "none"
				}
				autoUpdateIntervalHour := int(item.Get("autoUpdateIntervalHour").Int())

				var maxSort int
				db.QueryRow("SELECT COALESCE(MAX(sort), -1) FROM subscriptions").Scan(&maxSort)
				newSort := maxSort + 1

				res, err := db.Exec(
					"INSERT INTO subscriptions (address, remarks, status, info, auto_select, auto_update_mode, auto_update_interval_hour, last_update_attempt_at, sort) VALUES (?, ?, ?, ?, ?, ?, ?, strftime('%Y-%m-%dT%H:%M:%SZ', 'now'), ?)",
					address, remarks, status, info, autoSelect, autoUpdateMode, autoUpdateIntervalHour, newSort,
				)
				if err != nil {
					return err
				}

				subID, _ := res.LastInsertId()

				servers := item.Get("servers").Array()
				for j, s := range servers {
					_, err := db.Exec(
						"INSERT INTO servers (type, sub_id, config_json, sort) VALUES ('subscription_server', ?, ?, ?)",
						subID, s.Raw, j,
					)
					if err != nil {
						return fmt.Errorf("ListAppend: failed to insert subscription server: %w", err)
					}
				}
			}
		}
		return nil

	default:
		b, err := jsoniter.Marshal(val)
		if err != nil {
			return err
		}
		_, err = db.Exec("INSERT OR REPLACE INTO system_config (key, value) VALUES (?, ?)",
			bucket+"/"+key, string(b))
		return err
	}
}

// ListGetAll retrieves all elements from a list.
func ListGetAll(bucket string, key string) (list [][]byte, err error) {
	db := GetDB()

	switch bucket + "/" + key {
	case "touch/servers":
		rows, err := db.Query("SELECT config_json FROM servers WHERE type = 'server' ORDER BY sort")
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var configJSON string
			if err := rows.Scan(&configJSON); err != nil {
				return nil, err
			}
			list = append(list, []byte(configJSON))
		}
		return list, rows.Err()

	case "touch/subscriptions":
		rows, err := db.Query("SELECT id, address, remarks, status, info, auto_select, auto_update_mode, auto_update_interval_hour, last_update_attempt_at FROM subscriptions ORDER BY sort")
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var id int64
			var address, remarks, status, info, autoUpdateMode string
			var autoSelectInt, autoUpdateIntervalHour int
			var lastUpdateAttemptAt sql.NullString
			if err := rows.Scan(&id, &address, &remarks, &status, &info, &autoSelectInt, &autoUpdateMode, &autoUpdateIntervalHour, &lastUpdateAttemptAt); err != nil {
				return nil, err
			}

			serverRows, err := db.Query(
				"SELECT config_json FROM servers WHERE type = 'subscription_server' AND sub_id = ? ORDER BY sort",
				id,
			)
			if err != nil {
				return nil, err
			}

			var servers []string
			for serverRows.Next() {
				var s string
				if err := serverRows.Scan(&s); err != nil {
					serverRows.Close()
					return nil, err
				}
				servers = append(servers, s)
			}
			serverRows.Close()

			autoSelect := autoSelectInt != 0
			result, err := marshalSubscriptionListValue(
				id, remarks, address, status, info, servers, autoSelect,
				autoUpdateMode, autoUpdateIntervalHour, lastUpdateAttemptAt.String,
			)
			if err != nil {
				return nil, err
			}
			list = append(list, result)
		}
		return list, rows.Err()

	default:
		return nil, fmt.Errorf("ListGetAll: unsupported bucket/key: %s/%s", bucket, key)
	}
}

// ListRemove removes elements at specified indexes from a list.
func ListRemove(bucket, key string, indexes []int) error {
	if len(indexes) == 0 {
		return fmt.Errorf("ListRemove: nothing to remove")
	}

	db := GetDB()

	switch bucket + "/" + key {
	case "touch/servers":
		for _, idx := range indexes {
			// Clean up outbound_connections first to satisfy foreign key constraint
			_, _ = db.Exec(`
				DELETE FROM outbound_connections
				WHERE server_id IN (SELECT id FROM servers WHERE type = 'server' AND sort = ?)
			`, idx)
			_, err := db.Exec("DELETE FROM servers WHERE type = 'server' AND sort = ?", idx)
			if err != nil {
				return err
			}
		}
		// Reorder remaining servers
		_, err := db.Exec(`
			UPDATE servers SET sort = (
				SELECT COUNT(*) FROM servers s2 
				WHERE s2.type = 'server' AND s2.sort < servers.sort
			) WHERE type = 'server'
		`)
		return err

	case "touch/subscriptions":
		for _, idx := range indexes {
			var subID int64
			err := db.QueryRow("SELECT id FROM subscriptions WHERE sort = ?", idx).Scan(&subID)
			if err != nil {
				if err == sql.ErrNoRows {
					continue
				}
				return err
			}
			// Clean up outbound_connections first to satisfy foreign key constraint
			_, _ = db.Exec(`
				DELETE FROM outbound_connections
				WHERE server_id IN (SELECT id FROM servers WHERE type = 'subscription_server' AND sub_id = ?)
			`, subID)
			_, err = db.Exec("DELETE FROM servers WHERE type = 'subscription_server' AND sub_id = ?", subID)
			if err != nil {
				return err
			}
			_, err = db.Exec("DELETE FROM subscriptions WHERE id = ?", subID)
			if err != nil {
				return err
			}
		}
		// Reorder remaining subscriptions
		_, err := db.Exec(`
			UPDATE subscriptions SET sort = (
				SELECT COUNT(*) FROM subscriptions s2 
				WHERE s2.sort < subscriptions.sort
			)
		`)
		return err

	default:
		return fmt.Errorf("ListRemove: unsupported bucket/key: %s/%s", bucket, key)
	}
}

// ListLen returns the length of a list.
func ListLen(bucket string, key string) (length int, err error) {
	db := GetDB()

	switch bucket + "/" + key {
	case "touch/servers":
		err = db.QueryRow("SELECT COUNT(*) FROM servers WHERE type = 'server'").Scan(&length)
		return length, err

	case "touch/subscriptions":
		err = db.QueryRow("SELECT COUNT(*) FROM subscriptions").Scan(&length)
		return length, err

	default:
		return 0, fmt.Errorf("ListLen: unsupported bucket/key: %s/%s", bucket, key)
	}
}

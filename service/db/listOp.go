package db

import (
	"database/sql"
	"fmt"

	jsoniter "github.com/json-iterator/go"
	"github.com/tidwall/gjson"
)

// ListSet sets an element at a specific index in a list.
func ListSet(bucket string, key string, index int, val interface{}) (err error) {
	switch bucket + "/" + key {
	case "touch/servers":
		return ServersSet(index, val)
	case "touch/subscriptions":
		return SubscriptionsSet(index, val, nil)
	default:
		return fmt.Errorf("ListSet: unsupported bucket/key: %s/%s", bucket, key)
	}
}

func ServersSet(index int, val interface{}) error {
	b, err := jsoniter.Marshal(val)
	if err != nil {
		return err
	}
	result, err := GetDB().Exec(
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
}

// ListSetWithTransaction runs beforeSet and the list replacement in one
// transaction. The callback may write related state through transaction-aware
// database operations.
func ListSetWithTransaction(bucket string, key string, index int, val interface{}, beforeSet func(tx *sql.Tx) error) error {
	if bucket+"/"+key != "touch/subscriptions" {
		return fmt.Errorf("ListSetWithTransaction: unsupported bucket/key: %s/%s", bucket, key)
	}
	return SubscriptionsSet(index, val, beforeSet)
}

func SubscriptionsSet(index int, val interface{}, beforeSet func(tx *sql.Tx) error) error {
	b, err := jsoniter.Marshal(val)
	if err != nil {
		return err
	}
	return ReadModifyWrite(func(tx *sql.Tx) error {
		if beforeSet != nil {
			if err := beforeSet(tx); err != nil {
				return err
			}
		}
		return setSubscription(tx, index, gjson.ParseBytes(b))
	})
}

func setSubscription(tx *sql.Tx, index int, parsed gjson.Result) error {
	address := parsed.Get("address").String()
	remarks := parsed.Get("remarks").String()
	status := parsed.Get("status").String()
	info := parsed.Get("info").String()
	autoSelect := 0
	if parsed.Get("autoSelect").Bool() {
		autoSelect = 1
	}

	// The first statement writes, so the transaction takes the write lock
	// up front under busy_timeout instead of upgrading a read snapshot that
	// another connection's commit may have made stale (SQLITE_BUSY_SNAPSHOT).
	var subID int64
	err := tx.QueryRow(
		"UPDATE subscriptions SET address = ?, remarks = ?, status = ?, info = ?, auto_select = ?, updated_at = CURRENT_TIMESTAMP WHERE sort = ? RETURNING id",
		address, remarks, status, info, autoSelect, index,
	).Scan(&subID)
	if err == sql.ErrNoRows {
		return fmt.Errorf("ListSet: subscription at index %d not found", index)
	}
	if err != nil {
		return err
	}

	if _, err := tx.Exec("DELETE FROM servers WHERE type = 'subscription_server' AND sub_id = ?", subID); err != nil {
		return fmt.Errorf("ListSet: failed to clear old servers of subscription %d: %w", index, err)
	}

	servers := parsed.Get("servers").Array()
	for j, s := range servers {
		_, err := tx.Exec(
			"INSERT INTO servers (type, sub_id, config_json, sort) VALUES ('subscription_server', ?, ?, ?)",
			subID, s.Raw, j,
		)
		if err != nil {
			return fmt.Errorf("ListSet: failed to update subscription server %d/%d: %w", index, j, err)
		}
	}
	return nil
}

// ListGet retrieves an element at a specific index from a list.
func ListGet(bucket string, key string, index int) (b []byte, err error) {
	switch bucket + "/" + key {
	case "touch/servers":
		return serversGet(index)
	case "touch/subscriptions":
		return SubscriptionsGet(index)
	default:
		return nil, fmt.Errorf("ListGet: unsupported bucket/key: %s/%s", bucket, key)
	}
}

func serversGet(index int) ([]byte, error) {
	var configJSON string
	err := GetDB().QueryRow(
		"SELECT config_json FROM servers WHERE type = 'server' AND sort = ?", index,
	).Scan(&configJSON)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("ListGet: can't get element from an empty list")
	}
	if err != nil {
		return nil, err
	}
	return []byte(configJSON), nil
}

func SubscriptionsGet(index int) ([]byte, error) {
	db := GetDB()
	var subID int64
	var address, remarks, status, info string
	var autoSelectInt int
	err := db.QueryRow(
		"SELECT id, address, remarks, status, info, auto_select FROM subscriptions WHERE sort = ?", index,
	).Scan(&subID, &address, &remarks, &status, &info, &autoSelectInt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("ListGet: can't get element from an empty list")
	}
	if err != nil {
		return nil, err
	}

	// Reconstruct the subscription JSON with servers
	rows, err := db.Query(
		"SELECT config_json FROM servers WHERE type = 'subscription_server' AND sub_id = ? ORDER BY sort",
		subID,
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

	return subscriptionJSON(remarks, address, status, info, servers, autoSelectInt != 0)
}

// ListAppend appends values to a list.
func ListAppend(bucket string, key string, val interface{}) (err error) {
	switch bucket + "/" + key {
	case "touch/servers":
		return ServersAppend(val)
	case "touch/subscriptions":
		return SubscriptionsAppend(val)
	default:
		db := GetDB()
		b, err := jsoniter.Marshal(val)
		if err != nil {
			return err
		}
		_, err = db.Exec("INSERT OR REPLACE INTO system_config (key, value) VALUES (?, ?)",
			bucket+"/"+key, string(b))
		return err
	}
}

func ServersAppend(val interface{}) (err error) {
	db := GetDB()
	b, err := jsoniter.Marshal(val)
	if err != nil {
		return err
	}
	parsed := gjson.ParseBytes(b)
	if parsed.IsArray() {
		for _, item := range parsed.Array() {
			_, err = db.Exec(
				"INSERT INTO servers (type, config_json, sort) VALUES ('server', ?, (SELECT COALESCE(MAX(sort), -1) + 1 FROM servers WHERE type = 'server'))",
				item.Raw,
			)
			if err != nil {
				return err
			}
		}
	} else {
		_, err = db.Exec(
			"INSERT INTO servers (type, config_json, sort) VALUES ('server', ?, (SELECT COALESCE(MAX(sort), -1) + 1 FROM servers WHERE type = 'server'))",
			string(b),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func SubscriptionsAppend(val interface{}) (err error) {
	db := GetDB()
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

			res, err := db.Exec(
				"INSERT INTO subscriptions (address, remarks, status, info, auto_select, sort) VALUES (?, ?, ?, ?, ?, (SELECT COALESCE(MAX(sort), -1) + 1 FROM subscriptions))",
				address, remarks, status, info, autoSelect,
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
}

// ListGetAll retrieves all elements from a list.
func ListGetAll(bucket string, key string) (list [][]byte, err error) {
	switch bucket + "/" + key {
	case "touch/servers":
		return ServersGetAll()
	case "touch/subscriptions":
		return SubscriptionsGetAll()
	default:
		return nil, fmt.Errorf("ListGetAll: unsupported bucket/key: %s/%s", bucket, key)
	}
}

func ServersGetAll() (list [][]byte, err error) {
	rows, err := GetDB().Query("SELECT config_json FROM servers WHERE type = 'server' ORDER BY sort")
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
}

func SubscriptionsGetAll() (list [][]byte, err error) {
	db := GetDB()
	rows, err := db.Query("SELECT id, address, remarks, status, info, auto_select FROM subscriptions ORDER BY sort")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var address, remarks, status, info string
		var autoSelectInt int
		if err := rows.Scan(&id, &address, &remarks, &status, &info, &autoSelectInt); err != nil {
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

		result, err := subscriptionJSON(remarks, address, status, info, servers, autoSelectInt != 0)
		if err != nil {
			return nil, err
		}
		list = append(list, result)
	}
	return list, rows.Err()
}

// ListRemove removes elements at specified indexes from a list.
func ListRemove(bucket, key string, indexes []int) error {
	if len(indexes) == 0 {
		return fmt.Errorf("ListRemove: nothing to remove")
	}
	switch bucket + "/" + key {
	case "touch/servers":
		return ServersRemove(indexes)
	case "touch/subscriptions":
		return SubscriptionsRemove(indexes)
	default:
		return fmt.Errorf("ListRemove: unsupported bucket/key: %s/%s", bucket, key)
	}
}

func ServersRemove(indexes []int) error {
	return serversRemove(GetDB(), indexes)
}

// execer is what a removal needs from either the pool or one transaction.
type execer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

func serversRemove(db execer, indexes []int) error {
	if len(indexes) == 0 {
		return fmt.Errorf("ListRemove: nothing to remove")
	}
	for _, idx := range indexes {
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
}

func SubscriptionsRemove(indexes []int) error {
	return subscriptionsRemove(GetDB(), indexes)
}

func subscriptionsRemove(db execer, indexes []int) error {
	if len(indexes) == 0 {
		return fmt.Errorf("ListRemove: nothing to remove")
	}
	for _, idx := range indexes {
		var subID int64
		err := db.QueryRow("SELECT id FROM subscriptions WHERE sort = ?", idx).Scan(&subID)
		if err != nil {
			if err == sql.ErrNoRows {
				continue
			}
			return err
		}
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
}

// RemoveTx deletes subscriptions and servers by ordinal, and stores the
// connected lists that were renumbered for the deletion, in one
// transaction: a reader never sees ordinals that point at the wrong row.
func RemoveTx(subscriptions, servers []int, connected func(tx *sql.Tx) error) error {
	return ReadModifyWrite(func(tx *sql.Tx) error {
		if err := connected(tx); err != nil {
			return err
		}
		if len(subscriptions) > 0 {
			if err := subscriptionsRemove(tx, subscriptions); err != nil {
				return err
			}
		}
		if len(servers) > 0 {
			if err := serversRemove(tx, servers); err != nil {
				return err
			}
		}
		return nil
	})
}

// ListLen returns the length of a list.
func ListLen(bucket string, key string) (length int, err error) {
	switch bucket + "/" + key {
	case "touch/servers":
		return ServersLen()
	case "touch/subscriptions":
		return SubscriptionsLen()
	default:
		return 0, fmt.Errorf("ListLen: unsupported bucket/key: %s/%s", bucket, key)
	}
}

func ServersLen() (length int, err error) {
	err = GetDB().QueryRow("SELECT COUNT(*) FROM servers WHERE type = 'server'").Scan(&length)
	return length, err
}

func SubscriptionsLen() (length int, err error) {
	err = GetDB().QueryRow("SELECT COUNT(*) FROM subscriptions").Scan(&length)
	return length, err
}

// subscriptionJSON rebuilds a stored subscription from its columns. The
// server rows are already JSON; the scalar columns are user- or
// provider-supplied text and must be escaped, otherwise a remark with a
// quote made the whole subscription unreadable.
func subscriptionJSON(remarks, address, status, info string, servers []string, autoSelect bool) ([]byte, error) {
	raw := make([]jsoniter.RawMessage, 0, len(servers))
	for _, s := range servers {
		raw = append(raw, jsoniter.RawMessage(s))
	}
	return jsoniter.Marshal(struct {
		Remarks    string                `json:"remarks"`
		Address    string                `json:"address"`
		Status     string                `json:"status"`
		Info       string                `json:"info"`
		Servers    []jsoniter.RawMessage `json:"servers"`
		AutoSelect bool                  `json:"autoSelect"`
	}{remarks, address, status, info, raw, autoSelect})
}

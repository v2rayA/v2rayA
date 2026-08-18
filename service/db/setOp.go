package db

import (
	"fmt"

	"github.com/v2rayA/v2rayA/common"
)

// SetAdd adds members to a set identified by key.
// Currently supports "outbounds/names" -> outbound_names table.
func SetAdd(bucket string, key string, val interface{}) (err error) {
	db := GetDB()

	switch bucket + "/" + key {
	case "outbounds/names":
		name, ok := val.(string)
		if !ok {
			return fmt.Errorf("SetAdd: outbound name must be a string")
		}
		// Get the next sort value
		var maxSort int
		db.QueryRow("SELECT COALESCE(MAX(sort), -1) FROM outbound_names").Scan(&maxSort)
		_, err = db.Exec(
			"INSERT OR IGNORE INTO outbound_names (name, sort) VALUES (?, ?)",
			name, maxSort+1,
		)
		return err

	default:
		return fmt.Errorf("SetAdd: unsupported bucket/key: %s/%s", bucket, key)
	}
}

// SetRemove removes members from a set identified by key.
func SetRemove(bucket string, key string, val interface{}) (err error) {
	db := GetDB()

	switch bucket + "/" + key {
	case "outbounds/names":
		name, ok := val.(string)
		if !ok {
			return fmt.Errorf("SetRemove: outbound name must be a string")
		}
		// Delete related connections and settings first
		_, _ = db.Exec("DELETE FROM outbound_connections WHERE outbound_name = ?", name)
		_, _ = db.Exec("DELETE FROM outbound_settings WHERE outbound_name = ?", name)
		// Delete the outbound name
		_, err = db.Exec("DELETE FROM outbound_names WHERE name = ?", name)
		return err

	default:
		return fmt.Errorf("SetRemove: unsupported bucket/key: %s/%s", bucket, key)
	}
}

// SetIsMember checks if a member exists in a set.
func SetIsMember(bucket string, key string, member string) (bool, error) {
	db := GetDB()

	switch bucket + "/" + key {
	case "outbounds/names":
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM outbound_names WHERE name = ?", member).Scan(&count)
		if err != nil {
			return false, err
		}
		return count > 0, nil

	default:
		return false, fmt.Errorf("SetIsMember: unsupported bucket/key: %s/%s", bucket, key)
	}
}

// SetMembers returns all members of a set.
func SetMembers(bucket string, key string) ([]string, error) {
	db := GetDB()

	switch bucket + "/" + key {
	case "outbounds/names":
		rows, err := db.Query("SELECT name FROM outbound_names ORDER BY sort")
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var members []string
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				return nil, err
			}
			members = append(members, name)
		}
		return members, rows.Err()

	default:
		return nil, fmt.Errorf("SetMembers: unsupported bucket/key: %s/%s", bucket, key)
	}
}

// StringSetGetAll returns all members of a set as strings.
// This is the function used by configure.GetOutbounds().
func StringSetGetAll(bucket string, key string) (members []string, err error) {
	return SetMembers(bucket, key)
}

// Ensure common import is used
var _ = common.BytesCopy

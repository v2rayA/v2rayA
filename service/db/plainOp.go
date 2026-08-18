package db

import (
	"database/sql"
	"fmt"
	"strings"

	jsoniter "github.com/json-iterator/go"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

// makeKey constructs a prefixed key using bucket:key format for data isolation
func makeKey(bucket string, key string) string {
	return bucket + ":" + key
}

// Get retrieves a value from system_config table by key and unmarshals it into val.
func Get(bucket string, key string, val interface{}) (err error) {
	db := GetDB()
	fullKey := makeKey(bucket, key)
	var value string
	err = db.QueryRow("SELECT value FROM system_config WHERE key = ?", fullKey).Scan(&value)
	if err == sql.ErrNoRows && bucket == "accounts" {
		// Backward compatibility: legacy SQLite stored account keys without
		// bucket prefix, e.g. "admin" instead of "accounts:admin".
		err = db.QueryRow("SELECT value FROM system_config WHERE key = ?", key).Scan(&value)
	}
	if err == sql.ErrNoRows {
		return fmt.Errorf("Get: key is not found")
	}
	if err != nil {
		return err
	}
	return jsoniter.Unmarshal([]byte(value), val)
}

// GetRaw retrieves a raw byte value from system_config table by key.
func GetRaw(bucket string, key string) (b []byte, err error) {
	db := GetDB()
	fullKey := makeKey(bucket, key)
	var value string
	err = db.QueryRow("SELECT value FROM system_config WHERE key = ?", fullKey).Scan(&value)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("GetRaw: key is not found")
	}
	if err != nil {
		return nil, err
	}
	return common.BytesCopy([]byte(value)), nil
}

// Exists checks if a key exists in the system_config table.
func Exists(bucket string, key string) (exists bool) {
	db := GetDB()
	fullKey := makeKey(bucket, key)
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM system_config WHERE key = ?", fullKey).Scan(&count)
	if err == nil && count == 0 && bucket == "accounts" {
		err = db.QueryRow("SELECT COUNT(*) FROM system_config WHERE key = ?", key).Scan(&count)
	}
	if err != nil {
		log.Warn("%v", err)
		return false
	}
	return count > 0
}

// GetBucketLen returns the number of entries in the specified bucket.
func GetBucketLen(bucket string) (length int, err error) {
	db := GetDB()
	prefix := bucket + ":"
	err = db.QueryRow("SELECT COUNT(*) FROM system_config WHERE key LIKE ?", prefix+"%").Scan(&length)
	return length, err
}

// GetBucketKeys returns all keys in the specified bucket (without the bucket prefix).
func GetBucketKeys(bucket string) (keys []string, err error) {
	db := GetDB()
	prefix := bucket + ":"
	query := "SELECT key FROM system_config WHERE key LIKE ? ORDER BY key"
	args := []interface{}{prefix + "%"}
	if bucket == "accounts" {
		// Backward compatibility: also include legacy unprefixed keys.
		query = "SELECT key FROM system_config WHERE key LIKE ? OR key NOT LIKE '%:%' ORDER BY key"
	}
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		// Remove the bucket prefix from the key.
		key = strings.TrimPrefix(key, prefix)
		keys = append(keys, key)
	}
	return keys, rows.Err()
}

// Set inserts or replaces a value in the system_config table.
func Set(bucket string, key string, val interface{}) (err error) {
	b, err := jsoniter.Marshal(val)
	if err != nil {
		return err
	}
	db := GetDB()
	fullKey := makeKey(bucket, key)
	_, err = db.Exec("INSERT OR REPLACE INTO system_config (key, value) VALUES (?, ?)", fullKey, string(b))
	return err
}

// Delete removes a single key from system_config.
func Delete(bucket string, key string) error {
	db := GetDB()
	fullKey := makeKey(bucket, key)
	query := "DELETE FROM system_config WHERE key = ?"
	args := []interface{}{fullKey}
	if bucket == "accounts" {
		// Remove both new-format and legacy-format account keys.
		query = "DELETE FROM system_config WHERE key = ? OR key = ?"
		args = []interface{}{fullKey, key}
	}
	_, err := db.Exec(query, args...)
	return err
}

// BucketClear removes all entries from the specified bucket in system_config table.
func BucketClear(bucket string) error {
	db := GetDB()
	prefix := bucket + ":"
	_, err := db.Exec("DELETE FROM system_config WHERE key LIKE ?", prefix+"%")
	return err
}

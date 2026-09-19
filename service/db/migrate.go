package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"go.etcd.io/bbolt"
)

// MigrateFromBoltDB migrates data from BoltDB to SQLite.
// It opens bolt.db, reads all data, creates a fresh SQLite database,
// writes data to SQLite, verifies integrity, then renames bolt.db to a backup.
// This function does NOT use GetDB() — it creates its own SQLite connection
// so that migration can happen before the normal SQLite initialization.
func MigrateFromBoltDB() error {
	confPath := conf.GetEnvironmentConfig().Config
	boltPath := filepath.Join(confPath, "bolt.db")
	sqlitePath := filepath.Join(confPath, "v2raya.db")

	if _, err := os.Stat(boltPath); err != nil {
		if os.IsNotExist(err) {
			log.Info("No BoltDB file found at %s, skipping migration", boltPath)
			return nil
		}
		return fmt.Errorf("failed to inspect BoltDB: %w", err)
	}

	if _, err := os.Stat(sqlitePath); err == nil {
		// Both files: the SQLite one is authoritative when it holds data.
		// An earlier release built it in place, so a failed migration could
		// leave an empty or partial file next to the intact bolt.db; that one
		// is discarded and the migration runs again.
		if sqliteHoldsData(sqlitePath) {
			backupPath, err := moveBoltAside(boltPath)
			if err != nil {
				return fmt.Errorf("SQLite database already exists but bolt.db could not be moved aside: %w", err)
			}
			log.Warn("SQLite database already exists; moved stale bolt.db to %s", backupPath)
			return nil
		}
		log.Warn("SQLite database at %s holds no data; discarding it and migrating bolt.db again", sqlitePath)
		for _, suffix := range []string{"", "-wal", "-shm"} {
			if err := os.Remove(sqlitePath + suffix); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to discard the partial SQLite database: %w", err)
			}
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to inspect SQLite database: %w", err)
	}

	log.Warn("Migrating from BoltDB to SQLite...")

	// Open BoltDB (read-only)
	boltDB, err := bbolt.Open(boltPath, 0600, &bbolt.Options{ReadOnly: true})
	if err != nil {
		return fmt.Errorf("failed to open BoltDB: %w", err)
	}
	defer func() {
		if boltDB != nil {
			_ = boltDB.Close()
		}
	}()

	tempFile, err := os.CreateTemp(confPath, ".v2raya.db.migrate-*")
	if err != nil {
		return fmt.Errorf("failed to create temporary SQLite database: %w", err)
	}
	tempPath := tempFile.Name()
	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("failed to close temporary SQLite database: %w", err)
	}
	defer func() {
		_ = os.Remove(tempPath)
		_ = os.Remove(tempPath + "-shm")
		_ = os.Remove(tempPath + "-wal")
	}()

	// Create a fresh SQLite database independently (do NOT use GetDB).
	sqldb, err := createSQLiteDB(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create SQLite database for migration: %w", err)
	}
	defer func() {
		if sqldb != nil {
			_ = sqldb.Close()
		}
	}()

	// Perform migration in a single transaction
	tx, err := sqldb.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin SQLite transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	// Migrate system bucket
	if err := migrateSystemBucket(boltDB, tx); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to migrate system bucket: %w", err)
	}

	// Migrate touch bucket (servers and subscriptions)
	if err := migrateTouchBucket(boltDB, tx); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to migrate touch bucket: %w", err)
	}

	// NOTE: Accounts are NOT migrated. Users must re-register after migration.
	// This is intentional: the old MD5-based password hashing is deprecated,
	// and requiring re-registration ensures users set up fresh bcrypt-based credentials.

	// Migrate outbounds bucket
	if err := migrateOutboundsBucket(boltDB, tx); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to migrate outbounds bucket: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration transaction: %w", err)
	}

	// Verify data integrity
	if err := verifyMigration(boltDB, sqldb); err != nil {
		return fmt.Errorf("migration verification failed: %w", err)
	}

	if err := sqldb.Close(); err != nil {
		return fmt.Errorf("failed to close migrated SQLite database: %w", err)
	}
	sqldb = nil
	if err := boltDB.Close(); err != nil {
		return fmt.Errorf("failed to close BoltDB: %w", err)
	}
	boltDB = nil

	if err := os.Rename(tempPath, sqlitePath); err != nil {
		return fmt.Errorf("failed to publish migrated SQLite database: %w", err)
	}
	backupPath, err := moveBoltAside(boltPath)
	if err != nil {
		if rollbackErr := os.Rename(sqlitePath, tempPath); rollbackErr != nil {
			return fmt.Errorf("failed to back up bolt.db: %v; failed to remove published SQLite database: %w", err, rollbackErr)
		}
		return fmt.Errorf("failed to back up bolt.db: %w", err)
	}

	log.Warn("Migration completed successfully. Old BoltDB backed up to %s", backupPath)
	return nil
}

// sqliteHoldsData reports whether a SQLite file has the system settings a
// completed migration always writes; a missing or empty table means the
// file is the remains of a migration that never finished.
func sqliteHoldsData(path string) bool {
	db, err := sql.Open(sqliteDriverName, sqliteDSN(path))
	if err != nil {
		return false
	}
	defer db.Close()
	var n int
	if err := db.QueryRow("SELECT COUNT(*) FROM system_config WHERE key LIKE 'system:%'").Scan(&n); err != nil {
		return false
	}
	return n > 0
}

func moveBoltAside(boltPath string) (string, error) {
	backupPath := boltPath + ".bak"
	for suffix := 1; ; suffix++ {
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			break
		} else if err != nil {
			return "", err
		}
		backupPath = fmt.Sprintf("%s.bak.%d", boltPath, suffix)
	}
	if err := os.Rename(boltPath, backupPath); err != nil {
		return "", err
	}
	return backupPath, nil
}

// createSQLiteDB creates a new SQLite database file at the given path,
// opens a connection, configures PRAGMAs, and initializes the schema.
func createSQLiteDB(dbPath string) (*sql.DB, error) {
	// Ensure the database file exists
	f, err := os.Create(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQLite database file: %w", err)
	}
	f.Close()

	if err := validateSQLiteDriver(); err != nil {
		return nil, err
	}

	db, err := sql.Open(sqliteDriverName, dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Configure PRAGMAs for WAL mode and performance
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA foreign_keys=ON",
		"PRAGMA cache_size=-8000",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to set PRAGMA %s: %w", p, err)
		}
	}

	// Initialize schema
	if err := InitSchema(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return db, nil
}

// migrateSystemBucket migrates the system bucket to system_config table
func migrateSystemBucket(boltDB *bbolt.DB, tx *sql.Tx) error {
	return boltDB.View(func(btx *bbolt.Tx) error {
		bkt := btx.Bucket([]byte("system"))
		if bkt == nil {
			log.Info("No system bucket found, skipping")
			return nil
		}

		stmt, err := tx.Prepare("INSERT OR REPLACE INTO system_config (key, value) VALUES (?, ?)")
		if err != nil {
			return err
		}
		defer stmt.Close()

		return bkt.ForEach(func(k, v []byte) error {
			// Add "system:" prefix to key for proper bucket isolation
			prefixedKey := "system:" + string(k)
			_, err := stmt.Exec(prefixedKey, string(v))
			if err != nil {
				return fmt.Errorf("failed to insert system config %s: %w", string(k), err)
			}
			log.Info("Migrated system config: %s", string(k))
			return nil
		})
	})
}

// migrateTouchBucket migrates the touch bucket (servers and subscriptions).
func migrateTouchBucket(boltDB *bbolt.DB, tx *sql.Tx) error {
	return boltDB.View(func(btx *bbolt.Tx) error {
		bkt := btx.Bucket([]byte("touch"))
		if bkt == nil {
			log.Info("No touch bucket found, skipping")
			return nil
		}

		// Migrate servers
		serversJSON := bkt.Get([]byte("servers"))
		if serversJSON != nil {
			if err := migrateServers(serversJSON, tx); err != nil {
				return err
			}
		}

		// Migrate subscriptions
		subsJSON := bkt.Get([]byte("subscriptions"))
		if subsJSON != nil {
			if err := migrateSubscriptions(subsJSON, tx); err != nil {
				return err
			}
		}

		return nil
	})
}

// migrateServers migrates the servers JSON array to the servers table
func migrateServers(data []byte, tx *sql.Tx) error {
	if !gjson.ValidBytes(data) {
		return fmt.Errorf("invalid JSON for servers")
	}

	parsed := gjson.ParseBytes(data)
	if !parsed.IsArray() {
		return fmt.Errorf("servers data is not an array")
	}

	results := parsed.Array()
	if len(results) == 0 {
		return nil
	}

	stmt, err := tx.Prepare(`
		INSERT INTO servers (type, address, port, protocol, config_json, latency, link, url, sort)
		VALUES ('server', '', 0, '', ?, '', '', '', ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i, r := range results {
		configJSON := r.Raw
		if _, err := stmt.Exec(configJSON, i); err != nil {
			return fmt.Errorf("failed to insert server %d: %w", i, err)
		}
		log.Info("Migrated server %d", i)
	}

	return nil
}

// migrateSubscriptions migrates the subscriptions JSON array to the subscriptions table.
func migrateSubscriptions(data []byte, tx *sql.Tx) error {
	if !gjson.ValidBytes(data) {
		return fmt.Errorf("invalid JSON for subscriptions")
	}

	parsed := gjson.ParseBytes(data)
	if !parsed.IsArray() {
		return fmt.Errorf("subscriptions data is not an array")
	}

	results := parsed.Array()
	if len(results) == 0 {
		return nil
	}

	subStmt, err := tx.Prepare(`
		INSERT INTO subscriptions (address, remarks, status, info, auto_select, sort)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer subStmt.Close()

	serverStmt, err := tx.Prepare(`
		INSERT INTO servers (type, sub_id, address, port, protocol, config_json, latency, link, url, sort)
		VALUES ('subscription_server', ?, '', 0, '', ?, '', '', '', ?)
	`)
	if err != nil {
		return err
	}
	defer serverStmt.Close()

	for i, r := range results {
		address := r.Get("address").String()
		remarks := r.Get("remarks").String()
		status := r.Get("status").String()
		info := r.Get("info").String()
		autoSelect := 0
		if r.Get("autoSelect").Bool() {
			autoSelect = 1
		}

		res, err := subStmt.Exec(address, remarks, status, info, autoSelect, i)
		if err != nil {
			return fmt.Errorf("failed to insert subscription %d: %w", i, err)
		}

		subID, err := res.LastInsertId()
		if err != nil {
			return fmt.Errorf("failed to get last insert id for subscription %d: %w", i, err)
		}

		// Migrate servers within this subscription
		servers := r.Get("servers").Array()
		for j, s := range servers {
			configJSON := s.Raw
			if _, err := serverStmt.Exec(subID, configJSON, j); err != nil {
				return fmt.Errorf("failed to insert subscription server %d/%d: %w", i, j, err)
			}
		}

		log.Info("Migrated subscription %d with %d servers", i, len(servers))
	}

	return nil
}

// migrateOutboundsBucket migrates the outbounds bucket
func migrateOutboundsBucket(boltDB *bbolt.DB, tx *sql.Tx) error {
	return boltDB.View(func(btx *bbolt.Tx) error {
		// Migrate outbound names from outbounds/names set
		outboundsBkt := btx.Bucket([]byte("outbounds"))
		if outboundsBkt != nil {
			namesData := outboundsBkt.Get([]byte("names"))
			if namesData != nil {
				if err := migrateOutboundNames(namesData, tx); err != nil {
					return err
				}
			}
		}

		// Migrate outbound settings and connections from outbound.{name} buckets
		return btx.ForEach(func(name []byte, _ *bbolt.Bucket) error {
			bucketName := string(name)
			if len(bucketName) > 9 && bucketName[:9] == "outbound." {
				outboundName := bucketName[9:]
				outboundBkt := btx.Bucket(name)
				if outboundBkt == nil {
					return nil
				}

				// Migrate setting
				settingData := outboundBkt.Get([]byte("setting"))
				if settingData != nil {
					if err := migrateOutboundSetting(outboundName, settingData, tx); err != nil {
						return err
					}
				}

				// Migrate connectedServers
				connData := outboundBkt.Get([]byte("connectedServers"))
				if connData != nil {
					if err := migrateOutboundConnections(outboundName, connData, tx); err != nil {
						return err
					}
				}
			}
			return nil
		})
	})
}

// migrateOutboundNames migrates the outbound names set
func migrateOutboundNames(data []byte, tx *sql.Tx) error {
	// Insert the default "proxy" outbound
	if _, err := tx.Exec("INSERT OR IGNORE INTO outbound_names (name, sort) VALUES ('proxy', 0)"); err != nil {
		return fmt.Errorf("failed to insert default outbound name: %w", err)
	}

	// Try to extract names from the gob data
	names := extractStringsFromGob(data)
	for i, name := range names {
		if name == "" {
			continue
		}
		if _, err := tx.Exec("INSERT OR IGNORE INTO outbound_names (name, sort) VALUES (?, ?)", name, i+1); err != nil {
			return fmt.Errorf("failed to insert outbound name %s: %w", name, err)
		}
		log.Info("Migrated outbound name: %s", name)
	}

	return nil
}

// extractStringsFromGob attempts to extract string values from gob-encoded data
func extractStringsFromGob(data []byte) []string {
	var names []string
	current := make([]byte, 0)
	for i := 0; i < len(data); i++ {
		if data[i] >= 32 && data[i] < 127 {
			current = append(current, data[i])
		} else {
			if len(current) > 1 {
				s := string(current)
				if s != "map" && s != "uint8" && s != "string" && s != "interface" &&
					len(s) > 2 && s != "false" && s != "true" {
					names = append(names, s)
				}
			}
			current = make([]byte, 0)
		}
	}
	return names
}

// migrateOutboundSetting migrates an outbound setting
func migrateOutboundSetting(outboundName string, data []byte, tx *sql.Tx) error {
	// Ensure outbound name exists
	if _, err := tx.Exec("INSERT OR IGNORE INTO outbound_names (name, sort) VALUES (?, 0)", outboundName); err != nil {
		return err
	}

	key := makeKey("outbound."+outboundName, "setting")
	if _, err := tx.Exec("INSERT OR REPLACE INTO system_config (key, value) VALUES (?, ?)", key, string(data)); err != nil {
		return fmt.Errorf("failed to insert outbound setting for %s: %w", outboundName, err)
	}
	log.Info("Migrated outbound setting: %s", outboundName)
	return nil
}

// migrateOutboundConnections preserves Bolt ordinals because configure.Which
// resolves them against the ordered server and subscription lists at runtime.
func migrateOutboundConnections(outboundName string, data []byte, tx *sql.Tx) error {
	if _, err := tx.Exec("INSERT OR IGNORE INTO outbound_names (name, sort) VALUES (?, 0)", outboundName); err != nil {
		return fmt.Errorf("failed to ensure outbound name %s exists: %w", outboundName, err)
	}

	key := makeKey("outbound."+outboundName, "connectedServers")
	if _, err := tx.Exec("INSERT OR REPLACE INTO system_config (key, value) VALUES (?, ?)", key, string(data)); err != nil {
		return fmt.Errorf("failed to insert outbound connections for %s: %w", outboundName, err)
	}
	return nil
}

// verifyMigration compares data counts between BoltDB and SQLite
func verifyMigration(boltDB *bbolt.DB, sqldb *sql.DB) error {
	log.Info("Verifying migration integrity...")

	err := boltDB.View(func(btx *bbolt.Tx) error {
		// Verify system config
		systemBkt := btx.Bucket([]byte("system"))
		if systemBkt != nil {
			var boltCount int
			systemBkt.ForEach(func(_, _ []byte) error {
				boltCount++
				return nil
			})

			var sqlCount int
			if err := sqldb.QueryRow("SELECT COUNT(*) FROM system_config WHERE key LIKE 'system:%'").Scan(&sqlCount); err != nil {
				return fmt.Errorf("failed to count migrated system config: %w", err)
			}

			if boltCount != sqlCount {
				return fmt.Errorf("system config count mismatch: BoltDB=%d, SQLite=%d", boltCount, sqlCount)
			}
			log.Info("System config: %d entries verified", boltCount)
		}

		// Verify servers
		touchBkt := btx.Bucket([]byte("touch"))
		if touchBkt != nil {
			serversData := touchBkt.Get([]byte("servers"))
			if serversData != nil {
				boltServerCount := len(gjson.ParseBytes(serversData).Array())
				var sqlServerCount int
				sqldb.QueryRow("SELECT COUNT(*) FROM servers WHERE type = 'server'").Scan(&sqlServerCount)
				if boltServerCount != sqlServerCount {
					return fmt.Errorf("server count mismatch: BoltDB=%d, SQLite=%d", boltServerCount, sqlServerCount)
				}
				log.Info("Servers: %d entries verified", boltServerCount)
			}

			// Verify subscriptions
			subsData := touchBkt.Get([]byte("subscriptions"))
			if subsData != nil {
				boltSubCount := len(gjson.ParseBytes(subsData).Array())
				var sqlSubCount int
				sqldb.QueryRow("SELECT COUNT(*) FROM subscriptions").Scan(&sqlSubCount)
				if boltSubCount != sqlSubCount {
					return fmt.Errorf("subscription count mismatch: BoltDB=%d, SQLite=%d", boltSubCount, sqlSubCount)
				}
				log.Info("Subscriptions: %d entries verified", boltSubCount)
			}
		}

		return btx.ForEach(func(name []byte, bucket *bbolt.Bucket) error {
			if bucket == nil || !strings.HasPrefix(string(name), "outbound.") {
				return nil
			}
			for _, key := range []string{"setting", "connectedServers"} {
				value := bucket.Get([]byte(key))
				if value == nil {
					continue
				}
				fullKey := makeKey(string(name), key)
				var migrated string
				if err := sqldb.QueryRow("SELECT value FROM system_config WHERE key = ?", fullKey).Scan(&migrated); err != nil {
					return fmt.Errorf("failed to read migrated %s: %w", fullKey, err)
				}
				if migrated != string(value) {
					return fmt.Errorf("outbound config mismatch for %s", fullKey)
				}
				if err := verifyOutboundValue(key, []byte(migrated)); err != nil {
					return fmt.Errorf("invalid migrated %s: %w", fullKey, err)
				}
			}
			return nil
		})
	})

	if err != nil {
		return err
	}

	log.Warn("Migration verification passed!")
	return nil
}

func verifyOutboundValue(key string, value []byte) error {
	switch key {
	case "setting":
		var setting struct {
			ProbeURL      string `json:"probeURL"`
			ProbeInterval string `json:"probeInterval"`
			Type          string `json:"type"`
			Selected      string `json:"selected,omitempty"`
		}
		return json.Unmarshal(value, &setting)
	case "connectedServers":
		var whiches struct {
			Touches []struct {
				Type     string `json:"_type"`
				ID       int    `json:"id"`
				Sub      int    `json:"sub"`
				Outbound string `json:"outbound"`
			} `json:"touches"`
		}
		return json.Unmarshal(value, &whiches)
	default:
		return nil
	}
}

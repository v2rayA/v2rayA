package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tidwall/gjson"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"go.etcd.io/bbolt"
)

// MigrateFromBoltDB migrates data from BoltDB to SQLite.
// It opens bolt.db, reads all data, creates a fresh SQLite database,
// writes data to SQLite, verifies integrity, then renames bolt.db to bolt.db.bak.
// This function does NOT use GetDB() — it creates its own SQLite connection
// so that migration can happen before the normal SQLite initialization.
func MigrateFromBoltDB() error {
	confPath := conf.GetEnvironmentConfig().Config
	boltPath := filepath.Join(confPath, "bolt.db")
	sqlitePath := filepath.Join(confPath, "v2raya.db")

	// Check if BoltDB file exists
	if _, err := os.Stat(boltPath); os.IsNotExist(err) {
		log.Info("No BoltDB file found at %s, skipping migration", boltPath)
		return nil
	}

	// Check if SQLite already exists (migration already done or fresh start)
	if _, err := os.Stat(sqlitePath); err == nil {
		log.Info("SQLite database already exists at %s, skipping migration", sqlitePath)
		return nil
	}

	log.Warn("Migrating from BoltDB to SQLite...")

	// Open BoltDB (read-only)
	boltDB, err := bbolt.Open(boltPath, 0600, &bbolt.Options{ReadOnly: true})
	if err != nil {
		return fmt.Errorf("failed to open BoltDB: %w", err)
	}
	defer boltDB.Close()

	// Create a fresh SQLite database independently (do NOT use GetDB)
	sqldb, err := createSQLiteDB(sqlitePath)
	if err != nil {
		return fmt.Errorf("failed to create SQLite database for migration: %w", err)
	}
	defer sqldb.Close()

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
	// subIDMap maps BoltDB subscription index (0-based) -> SQLite subscription ID
	subIDMap, err := migrateTouchBucket(boltDB, tx)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to migrate touch bucket: %w", err)
	}

	// NOTE: Accounts are NOT migrated. Users must re-register after migration.
	// This is intentional: the old MD5-based password hashing is deprecated,
	// and requiring re-registration ensures users set up fresh bcrypt-based credentials.

	// Migrate outbounds bucket
	if err := migrateOutboundsBucket(boltDB, tx, subIDMap); err != nil {
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

	// Close BoltDB before renaming
	boltDB.Close()

	// Rename old BoltDB to .bak as backup
	backupPath := boltPath + ".bak"
	if err := os.Rename(boltPath, backupPath); err != nil {
		return fmt.Errorf("failed to rename bolt.db to bolt.db.bak: %w", err)
	}

	log.Warn("Migration completed successfully. Old BoltDB backed up to bolt.db.bak")
	return nil
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

// migrateTouchBucket migrates the touch bucket (servers and subscriptions)
// Returns a map of BoltDB subscription index (0-based) -> SQLite subscription ID
// for use in outbound connection migration.
func migrateTouchBucket(boltDB *bbolt.DB, tx *sql.Tx) (map[int64]int64, error) {
	var subIDMap map[int64]int64
	err := boltDB.View(func(btx *bbolt.Tx) error {
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
			var err error
			subIDMap, err = migrateSubscriptions(subsJSON, tx)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return subIDMap, nil
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
// Returns a map of BoltDB subscription index (0-based) -> SQLite subscription ID.
func migrateSubscriptions(data []byte, tx *sql.Tx) (map[int64]int64, error) {
	if !gjson.ValidBytes(data) {
		return nil, fmt.Errorf("invalid JSON for subscriptions")
	}

	parsed := gjson.ParseBytes(data)
	if !parsed.IsArray() {
		return nil, fmt.Errorf("subscriptions data is not an array")
	}

	results := parsed.Array()
	if len(results) == 0 {
		return nil, nil
	}

	subIDMap := make(map[int64]int64, len(results))

	subStmt, err := tx.Prepare(`
		INSERT INTO subscriptions (address, remarks, status, info, auto_select, sort)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return nil, err
	}
	defer subStmt.Close()

	serverStmt, err := tx.Prepare(`
		INSERT INTO servers (type, sub_id, address, port, protocol, config_json, latency, link, url, sort)
		VALUES ('subscription_server', ?, '', 0, '', ?, '', '', '', ?)
	`)
	if err != nil {
		return nil, err
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
			return nil, fmt.Errorf("failed to insert subscription %d: %w", i, err)
		}

		subID, err := res.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("failed to get last insert id for subscription %d: %w", i, err)
		}

		// Record the mapping: BoltDB index (0-based) -> SQLite subscription ID
		subIDMap[int64(i)] = subID

		// Migrate servers within this subscription
		servers := r.Get("servers").Array()
		for j, s := range servers {
			configJSON := s.Raw
			if _, err := serverStmt.Exec(subID, configJSON, j); err != nil {
				return nil, fmt.Errorf("failed to insert subscription server %d/%d: %w", i, j, err)
			}
		}

		log.Info("Migrated subscription %d with %d servers", i, len(servers))
	}

	return subIDMap, nil
}

// migrateOutboundsBucket migrates the outbounds bucket
func migrateOutboundsBucket(boltDB *bbolt.DB, tx *sql.Tx, subIDMap map[int64]int64) error {
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
					if err := migrateOutboundConnections(outboundName, connData, tx, subIDMap); err != nil {
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

	if _, err := tx.Exec("INSERT OR REPLACE INTO outbound_settings (outbound_name, setting_json) VALUES (?, ?)",
		outboundName, string(data)); err != nil {
		return fmt.Errorf("failed to insert outbound setting for %s: %w", outboundName, err)
	}
	log.Info("Migrated outbound setting: %s", outboundName)
	return nil
}

// migrateOutboundConnections migrates connected servers for an outbound.
// subIDMap maps BoltDB subscription index (0-based) -> SQLite subscription ID.
func migrateOutboundConnections(outboundName string, data []byte, tx *sql.Tx, subIDMap map[int64]int64) error {
	if !gjson.ValidBytes(data) {
		return nil
	}

	// Ensure the outbound name exists in outbound_names table to satisfy
	// the FOREIGN KEY constraint on outbound_connections.outbound_name.
	// This is critical because migrateOutboundSetting may not be called
	// if the outbound bucket has connectedServers but no setting data.
	if _, err := tx.Exec("INSERT OR IGNORE INTO outbound_names (name, sort) VALUES (?, 0)", outboundName); err != nil {
		return fmt.Errorf("failed to ensure outbound name %s exists: %w", outboundName, err)
	}

	parsed := gjson.ParseBytes(data)
	touches := parsed.Get("touches").Array()

	stmt, err := tx.Prepare(`
		INSERT INTO outbound_connections (outbound_name, server_id, sort)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i, t := range touches {
		touchType := t.Get("_type").String()
		id := t.Get("id").Int()

		var serverID int64
		switch touchType {
		case "server":
			// In BoltDB, server IDs are 1-based sequential numbers.
			// After migration to SQLite, servers get new auto-increment IDs.
			// We look up the server by type='server' and sort=(id-1) since
			// sort is set to the array index (0-based) during migration.
			if err := tx.QueryRow(
				"SELECT id FROM servers WHERE type = 'server' AND sort = ?",
				id-1,
			).Scan(&serverID); err != nil {
				log.Warn("Could not find server with sort=%d (original id=%d): %v — skipping connection entry", id-1, id, err)
				continue
			}
		case "subscriptionServer":
			// In BoltDB, subscription server IDs are 1-based within each subscription.
			// The "sub" field in the touch data is the BoltDB subscription index (0-based).
			// We need to map it to the SQLite subscription ID using subIDMap.
			boltSubIdx := t.Get("sub").Int()
			sqlSubID, ok := subIDMap[boltSubIdx]
			if !ok {
				log.Warn("Could not find SQLite subscription ID for BoltDB subscription index %d — skipping connection entry", boltSubIdx)
				continue
			}
			if err := tx.QueryRow(
				"SELECT id FROM servers WHERE type = 'subscription_server' AND sub_id = ? AND sort = ?",
				sqlSubID, id-1,
			).Scan(&serverID); err != nil {
				log.Warn("Could not find subscription server for sub_id=%d (BoltDB sub=%d), sort=%d: %v — skipping connection entry", sqlSubID, boltSubIdx, id-1, err)
				continue
			}
		default:
			continue
		}

		if _, err := stmt.Exec(outboundName, serverID, i); err != nil {
			// Log the error and skip this connection entry rather than failing the entire migration.
			// This provides resilience against edge cases where server references may be stale.
			log.Warn("Failed to insert outbound connection for %s (server_id=%d, sort=%d): %v — skipping", outboundName, serverID, i, err)
			continue
		}
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
			sqldb.QueryRow("SELECT COUNT(*) FROM system_config").Scan(&sqlCount)

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

		return nil
	})

	if err != nil {
		return err
	}

	log.Warn("Migration verification passed!")
	return nil
}

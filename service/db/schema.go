package db

import (
	"database/sql"
	"fmt"
	jsoniter "github.com/json-iterator/go"

	"github.com/v2rayA/v2rayA/pkg/util/log"
)

// SQL statements for creating all tables
const schemaSQL = `
CREATE TABLE IF NOT EXISTS system_config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS servers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL DEFAULT 'server',
    sub_id INTEGER DEFAULT NULL,
    address TEXT NOT NULL DEFAULT '',
    port INTEGER NOT NULL DEFAULT 0,
    protocol TEXT NOT NULL DEFAULT '',
    config_json TEXT NOT NULL DEFAULT '{}',
    intel TEXT DEFAULT '',
    latency TEXT DEFAULT '',
    link TEXT DEFAULT '',
    url TEXT DEFAULT '',
    sort INTEGER NOT NULL DEFAULT 0,
    group_id TEXT DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS subscriptions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    address TEXT NOT NULL DEFAULT '',
    remarks TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT '',
    info TEXT DEFAULT '',
    auto_select INTEGER NOT NULL DEFAULT 0,
    filter TEXT DEFAULT '',
    group_id TEXT DEFAULT '',
    sort INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS accounts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS outbound_names (
    name TEXT PRIMARY KEY,
    sort INTEGER NOT NULL DEFAULT 0
);
`

// InitSchema creates all tables if they don't exist
func InitSchema(db *sql.DB) error {
	log.Info("Initializing database schema")
	_, err := db.Exec(schemaSQL)
	if err != nil {
		log.Fatal("Failed to initialize database schema: %v", err)
		return err
	}
	log.Info("Database schema initialized successfully")
	return nil
}

// MigrateSchema applies incremental schema migrations for existing databases.
// Unlike InitSchema (which uses CREATE TABLE IF NOT EXISTS), this handles
// ALTER TABLE additions for columns added after the initial schema version.
func MigrateSchema(db *sql.DB) error {
	// Check if remarks column exists (added after initial schema)
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('subscriptions') WHERE name = 'remarks'").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check for remarks column: %w", err)
	}
	if count == 0 {
		log.Info("Adding remarks column to subscriptions table")
		if _, err := db.Exec("ALTER TABLE subscriptions ADD COLUMN remarks TEXT NOT NULL DEFAULT ''"); err != nil {
			return fmt.Errorf("failed to add remarks column: %w", err)
		}
	}

	// Check if auto_select column exists (added after initial schema)
	err = db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('subscriptions') WHERE name = 'auto_select'").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check for auto_select column: %w", err)
	}
	if count == 0 {
		log.Info("Adding auto_select column to subscriptions table")
		if _, err := db.Exec("ALTER TABLE subscriptions ADD COLUMN auto_select INTEGER NOT NULL DEFAULT 0"); err != nil {
			return fmt.Errorf("failed to add auto_select column: %w", err)
		}
	}

	return migrateLegacyOutboundTables(db)
}

// migrateLegacyOutboundTables copies what the first SQLite migration wrote
// into outbound_connections and outbound_settings — tables the runtime
// never read — into the system_config keys it does read, for outbounds
// that have no key yet. A database that never had the tables, or whose
// keys were written since, is left alone.
func migrateLegacyOutboundTables(db *sql.DB) error {
	var tables int
	if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name IN ('outbound_connections', 'outbound_settings')").Scan(&tables); err != nil {
		return err
	}
	if tables == 0 {
		return nil
	}
	hasKey := func(key string) (bool, error) {
		var n int
		err := db.QueryRow("SELECT COUNT(*) FROM system_config WHERE key = ?", key).Scan(&n)
		return n > 0, err
	}
	rows, err := db.Query(`SELECT c.outbound_name, s.type, s.sort, COALESCE(sub.sort, -1)
		FROM outbound_connections c
		JOIN servers s ON s.id = c.server_id
		LEFT JOIN subscriptions sub ON sub.id = s.sub_id
		ORDER BY c.outbound_name, c.sort`)
	if err != nil {
		// outbound_settings may exist without outbound_connections
		rows = nil
	}
	type which struct {
		TYPE     string `json:"_type"`
		ID       int    `json:"id"`
		Sub      int    `json:"sub"`
		Outbound string `json:"outbound"`
	}
	groups := map[string][]which{}
	var order []string
	if rows != nil {
		for rows.Next() {
			var name, typ string
			var sort, subSort int
			if err := rows.Scan(&name, &typ, &sort, &subSort); err != nil {
				rows.Close()
				return err
			}
			w := which{ID: sort + 1, Outbound: name}
			if typ == "subscription_server" {
				w.TYPE, w.Sub = "subscriptionServer", subSort
			} else {
				w.TYPE = "server"
			}
			if _, seen := groups[name]; !seen {
				order = append(order, name)
			}
			groups[name] = append(groups[name], w)
		}
		rows.Close()
	}
	for _, name := range order {
		key := makeKey("outbound."+name, "connectedServers")
		if ok, err := hasKey(key); err != nil || ok {
			if err != nil {
				return err
			}
			continue
		}
		b, err := jsoniter.Marshal(map[string]interface{}{"touches": groups[name]})
		if err != nil {
			return err
		}
		if _, err := db.Exec("INSERT INTO system_config (key, value) VALUES (?, ?)", key, string(b)); err != nil {
			return err
		}
		log.Warn("Recovered the members of outbound %q from the first SQLite migration (%d)", name, len(groups[name]))
	}
	settings, err := db.Query("SELECT outbound_name, setting_json FROM outbound_settings")
	if err != nil {
		return nil
	}
	defer settings.Close()
	for settings.Next() {
		var name, setting string
		if err := settings.Scan(&name, &setting); err != nil {
			return err
		}
		key := makeKey("outbound."+name, "setting")
		if ok, err := hasKey(key); err != nil || ok || setting == "" || setting == "{}" {
			if err != nil {
				return err
			}
			continue
		}
		if _, err := db.Exec("INSERT INTO system_config (key, value) VALUES (?, ?)", key, setting); err != nil {
			return err
		}
		log.Warn("Recovered the setting of outbound %q from the first SQLite migration", name)
	}
	return nil
}

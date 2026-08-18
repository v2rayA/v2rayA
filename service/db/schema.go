package db

import (
	"database/sql"
	"fmt"

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

CREATE TABLE IF NOT EXISTS outbound_connections (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    outbound_name TEXT NOT NULL,
    server_id INTEGER NOT NULL,
    sort INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (outbound_name) REFERENCES outbound_names(name),
    FOREIGN KEY (server_id) REFERENCES servers(id),
    UNIQUE(outbound_name, server_id)
);

CREATE TABLE IF NOT EXISTS outbound_settings (
    outbound_name TEXT PRIMARY KEY,
    setting_json TEXT NOT NULL DEFAULT '{}',
    FOREIGN KEY (outbound_name) REFERENCES outbound_names(name)
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

	return nil
}

package db

import (
	"database/sql"

	"api-proxy/internal/db/migrate"
)

// Wraps a sql.DB and provides all data-access methods. Connection setup
// is handled by the caller; this type owns schema initialization and
// per-table operations defined across the db package.
type DB struct {
	conn *sql.DB
}

// Constructs a DB wrapper around the given *sql.DB connection.
func New(conn *sql.DB) *DB {
	return &DB{conn: conn}
}

// Creates the required tables on first run and runs all registered
// migrations so dependent code can assume bootstrapped state.
func (d *DB) InitSchema() error {
	const schema = `
CREATE TABLE IF NOT EXISTS rules (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL DEFAULT '',
	src TEXT NOT NULL,
	dest TEXT NOT NULL,
	dest_api_key TEXT NOT NULL DEFAULT '',
	skip_cert_verify INTEGER NOT NULL DEFAULT 0,
	api_key TEXT NOT NULL DEFAULT '',
	force_api_key INTEGER NOT NULL DEFAULT 0,
	tags TEXT NOT NULL DEFAULT '[]',
	comment TEXT NOT NULL DEFAULT '',
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_rules_src ON rules(src);

CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT NOT NULL UNIQUE,
	password TEXT NOT NULL,
	role TEXT NOT NULL DEFAULT 'user',
	created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users(username);

CREATE TABLE IF NOT EXISTS server_config (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	auth_jwt TEXT NOT NULL DEFAULT '',
	salt TEXT NOT NULL DEFAULT ''
);
INSERT OR IGNORE INTO server_config (id, auth_jwt, salt) VALUES (1, '', '');
`
	if _, err := d.conn.Exec(schema); err != nil {
		return err
	}
	return migrate.Run(d.conn)
}

// Maps a Go bool to its SQLite integer representation (1 for true,
// 0 for false).
func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

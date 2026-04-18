// Package migrate contains ordered database migrations that run after
// InitSchema has created the base tables. Each migration is applied at
// most once; applied names are recorded in the schema_migrations table
// so subsequent runs skip them. Migrations are executed in the order
// they appear in the migrations slice.
package migrate

import (
	"database/sql"
)

// Identifies a single migration step: a unique, stable name paired
// with the function that applies it. The name is what gets stored in
// schema_migrations, so renaming an already-applied migration will
// cause it to run again.
type Migration struct {
	Name string
	Run  func(*sql.DB) error
}

// Ordered list of migrations applied by Run, in declaration order.
// New migrations should be appended to the end; never reordered or
// renamed, since schema_migrations stores the name verbatim.
var migrations = []Migration{
	{Name: "00000000_init_server_config", Run: initServerConfig},
	{Name: "00000001_init_first_user", Run: initFirstUser},
}

// Applies every registered migration that has not already been recorded
// in schema_migrations, in declaration order. Migrations are expected
// to be idempotent: if a migration partially completes and the process
// crashes before recording success, it will be retried on next run.
func Run(conn *sql.DB) error {
	if _, err := conn.Exec(`
CREATE TABLE IF NOT EXISTS schema_migrations (
	name TEXT PRIMARY KEY,
	applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
)`); err != nil {
		return err
	}

	applied, err := loadAppliedMigrations(conn)
	if err != nil {
		return err
	}

	for _, m := range migrations {
		if applied[m.Name] {
			continue
		}
		if err := m.Run(conn); err != nil {
			return err
		}
		if _, err := conn.Exec(`INSERT INTO schema_migrations (name) VALUES (?)`, m.Name); err != nil {
			return err
		}
	}
	return nil
}

// Returns the set of migration names already recorded as applied.
func loadAppliedMigrations(conn *sql.DB) (map[string]bool, error) {
	rows, err := conn.Query(`SELECT name FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	applied := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		applied[name] = true
	}
	return applied, rows.Err()
}

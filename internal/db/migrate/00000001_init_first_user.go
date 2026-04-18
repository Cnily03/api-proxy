package migrate

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"

	"api-proxy/internal/crypto"
	"api-proxy/internal/model"
)

// Default credentials used to bootstrap the first admin account when
// the users table is empty. The frontend hashes plaintext passwords as
// sha256(plaintext + VITE_PASSWORD_SALT) before sending them; this
// migration reproduces that pre-hash so login with the default
// credentials works out of the box.
const (
	defaultAdminUsername      = "admin"
	defaultAdminPlaintext     = "password"
	defaultFrontendPasswdSalt = ".openicu"
)

// Creates a default admin account when the users table has no rows.
// Idempotent: if any user already exists, this migration is a no-op.
func initFirstUser(conn *sql.DB) error {
	var count int
	if err := conn.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	var salt string
	if err := conn.QueryRow(`SELECT salt FROM server_config WHERE id = 1`).Scan(&salt); err != nil {
		return err
	}

	frontendHash := sha256Hex(defaultAdminPlaintext + defaultFrontendPasswdSalt)
	dbHash, err := crypto.HashPassword(frontendHash, salt)
	if err != nil {
		return err
	}

	_, err = conn.Exec(
		`INSERT INTO users (username, password, role) VALUES (?, ?, ?)`,
		defaultAdminUsername, dbHash, string(model.RoleAdmin),
	)
	return err
}

// Returns the hex-encoded SHA-256 digest of the given string.
func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

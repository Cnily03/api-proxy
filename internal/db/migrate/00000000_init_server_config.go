package migrate

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"

	"api-proxy/internal/model"
)

// Ensures the server_config singleton row (id = 1) is fully populated:
// generates and persists the JWT signing configuration when auth_jwt is
// empty, and generates a random password salt when salt is empty. This
// runs as the zeroth migration so subsequent code can assume both
// fields are present.
func initServerConfig(conn *sql.DB) error {
	var authJWT, salt string
	if err := conn.QueryRow(`SELECT auth_jwt, salt FROM server_config WHERE id = 1`).Scan(&authJWT, &salt); err != nil {
		return err
	}

	if authJWT == "" {
		secret, err := randomHex(32)
		if err != nil {
			return err
		}
		cfg := model.JWTConfig{
			Key:    secret,
			Alg:    "HS256",
			ExpT:   7 * 24 * 3600, // 7 days
			RenewT: 5 * 24 * 3600, // 5 days
		}
		jwtJSON, _ := json.Marshal(cfg)
		if _, err := conn.Exec(`UPDATE server_config SET auth_jwt = ? WHERE id = 1`, string(jwtJSON)); err != nil {
			return err
		}
	}

	if salt == "" {
		rnd, err := randomAlphanumeric(8)
		if err != nil {
			return err
		}
		if _, err := conn.Exec(`UPDATE server_config SET salt = ? WHERE id = 1`, "."+rnd); err != nil {
			return err
		}
	}

	return nil
}

// Generates a cryptographically random byte sequence of length n and
// returns its hex encoding (resulting string is 2*n characters).
func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// Generates a cryptographically random alphanumeric string of length n
// drawn from [a-zA-Z0-9].
func randomAlphanumeric(n int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b), nil
}

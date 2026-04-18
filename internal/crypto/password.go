// Package crypto provides password-hashing primitives shared across the
// application. Centralizing these here avoids import cycles between the
// db package and its migrations, both of which need to hash passwords.
package crypto

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters tuned per OWASP 2023 guidance: 64 MiB memory,
// 3 iterations, 2-way parallelism, 32-byte output.
const (
	argonTime    uint32 = 3
	argonMemory  uint32 = 64 * 1024 // KiB
	argonThread  uint8  = 2
	argonKeyLen  uint32 = 32
	argonSaltLen        = 16
)

// Produces a PHC-formatted argon2id hash of passwd+pepper (pepper acting
// as a server-wide secret). A fresh per-record 16-byte salt is generated
// and embedded in the resulting string:
//
//	$argon2id$v=19$m=65536,t=3,p=2$<salt-b64>$<hash-b64>
func HashPassword(passwd, pepper string) (string, error) {
	recSalt := make([]byte, argonSaltLen)
	if _, err := rand.Read(recSalt); err != nil {
		return "", err
	}
	key := argon2.IDKey([]byte(passwd+pepper), recSalt, argonTime, argonMemory, argonThread, argonKeyLen)
	enc := base64.RawStdEncoding
	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemory, argonTime, argonThread,
		enc.EncodeToString(recSalt), enc.EncodeToString(key),
	), nil
}

// Reports whether the stored argon2id PHC hash matches passwd+pepper.
// Uses constant-time comparison and re-derives with the stored parameters.
func VerifyPassword(dbHash, passwd, pepper string) bool {
	parts := strings.Split(dbHash, "$")
	// Expected: ["", "argon2id", "v=19", "m=...,t=...,p=...", "<salt>", "<hash>"]
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return false
	}
	var m, t uint32
	var p uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m, &t, &p); err != nil {
		return false
	}
	enc := base64.RawStdEncoding
	recSalt, err := enc.DecodeString(parts[4])
	if err != nil {
		return false
	}
	expected, err := enc.DecodeString(parts[5])
	if err != nil {
		return false
	}
	actual := argon2.IDKey([]byte(passwd+pepper), recSalt, t, m, p, uint32(len(expected)))
	return subtle.ConstantTimeCompare(actual, expected) == 1
}

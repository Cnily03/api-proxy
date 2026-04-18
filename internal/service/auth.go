package service

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"api-proxy/internal/crypto"
	"api-proxy/internal/db"
	"api-proxy/internal/model"

	"github.com/golang-jwt/jwt/v5"
)

// ── Login rate limiter ──

var ErrTooManyAttempts = errors.New("too many attempts, please try again later")

const (
	loginMaxAttempts = 5
	loginWindowDur   = 5 * time.Minute
)

type loginAttempts struct {
	mu    sync.Mutex
	times []time.Time
}

var loginLimiter sync.Map // username -> *loginAttempts

// Returns ErrTooManyAttempts when the user has exceeded the allowed failed
// login attempts within the rate-limit window; otherwise returns nil.
func checkLoginLimit(username string) error {
	now := time.Now()
	val, _ := loginLimiter.LoadOrStore(username, &loginAttempts{})
	a := val.(*loginAttempts)
	a.mu.Lock()
	defer a.mu.Unlock()

	// Prune old entries
	cutoff := now.Add(-loginWindowDur)
	valid := a.times[:0]
	for _, t := range a.times {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	a.times = valid

	if len(a.times) >= loginMaxAttempts {
		return ErrTooManyAttempts
	}
	return nil
}

// Records a failed login attempt at the current time for the given username,
// used by the rate limiter to decide future lockouts.
func recordLoginFailure(username string) {
	val, _ := loginLimiter.LoadOrStore(username, &loginAttempts{})
	a := val.(*loginAttempts)
	a.mu.Lock()
	a.times = append(a.times, time.Now())
	a.mu.Unlock()
}

type AuthService struct {
	db     *db.DB
	config *model.ServerConfig
}

// Constructs an AuthService backed by the given database handle.
func NewAuthService(d *db.DB) *AuthService {
	return &AuthService{db: d}
}

// Loads and caches the server configuration from the database. The
// configuration is expected to already be populated by the db package's
// migrations run during InitSchema.
func (s *AuthService) LoadConfig() error {
	cfg, err := s.db.GetServerConfig()
	if err != nil {
		return err
	}
	s.config = cfg
	return nil
}

// Returns the cached server configuration loaded by LoadConfig.
func (s *AuthService) Config() *model.ServerConfig {
	return s.config
}

// Authenticates a user using the frontend-hashed password (as passwd)
// (sha256(plaintext + VITE_PASSWORD_SALT)), enforcing the login rate limit,
// and returns a signed JWT on success.
func (s *AuthService) Login(username, passwd string) (string, error) {
	if err := checkLoginLimit(username); err != nil {
		return "", err
	}
	u, err := s.db.GetUserByUsername(username)
	if err != nil {
		recordLoginFailure(username)
		return "", errors.New("invalid credentials")
	}
	if !verifyDBHash(u.Password, passwd, s.config.Salt) {
		recordLoginFailure(username)
		return "", errors.New("invalid credentials")
	}
	token, err := s.generateJWT(u)
	if err != nil {
		return "", err
	}
	return token, nil
}

// Parses and validates a JWT string, returning the associated user when
// the token is well-formed, correctly signed, and unexpired.
func (s *AuthService) ValidateToken(tokenStr string) (*model.User, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(s.config.JWT.Key), nil
	})
	if err != nil {
		return nil, errors.New("invalid token")
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	sub, _ := claims.GetSubject()
	if sub == "" {
		return nil, errors.New("invalid token")
	}
	var userID int64
	if _, err := fmt.Sscanf(sub, "%d", &userID); err != nil {
		return nil, errors.New("invalid token")
	}
	return s.db.GetUserByID(userID)
}

// Returns all users currently stored in the database.
func (s *AuthService) ListUsers() ([]model.User, error) {
	return s.db.ListUsers()
}

// Creates a new user with the given role, hashing the frontend-hashed
// password (as passwd) for database storage; returns the new user's ID.
func (s *AuthService) CreateUser(username, passwd string, role model.UserRole) (int64, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return 0, errors.New("username is required")
	}
	if passwd == "" {
		return 0, errors.New("password is required")
	}
	dbHash, err := hashForDB(passwd, s.config.Salt)
	if err != nil {
		return 0, err
	}
	return s.db.CreateUser(model.User{Username: username, Password: dbHash, Role: role})
}

// Updates the specified user's username, password, and/or role; empty
// string/zero values for any field leave that field unchanged.
func (s *AuthService) UpdateUser(id int64, username, passwd string, role model.UserRole) error {
	u, err := s.db.GetUserByID(id)
	if err != nil {
		return err
	}
	if username != "" {
		u.Username = strings.TrimSpace(username)
	}
	if passwd != "" {
		dbHash, err := hashForDB(passwd, s.config.Salt)
		if err != nil {
			return err
		}
		u.Password = dbHash
	}
	if role != "" {
		u.Role = role
	}
	return s.db.UpdateUser(*u)
}

// Deletes the user with the given ID from the database.
func (s *AuthService) DeleteUser(id int64) error {
	return s.db.DeleteUser(id)
}

// Verifies the old frontend-hashed password and, on success, replaces the
// stored password with the hash of the new frontend-hashed password.
func (s *AuthService) ChangePassword(userID int64, oldFrontendHash, newFrontendHash string) error {
	u, err := s.db.GetUserByID(userID)
	if err != nil {
		return err
	}
	if !verifyDBHash(u.Password, oldFrontendHash, s.config.Salt) {
		return errors.New("old password is incorrect")
	}
	newHash, err := hashForDB(newFrontendHash, s.config.Salt)
	if err != nil {
		return err
	}
	u.Password = newHash
	return s.db.UpdateUser(*u)
}

// ── Helpers ──

// Produces a signed HS256 JWT for the given user, with issued-at and
// expiration claims derived from the server JWT config.
func (s *AuthService) generateJWT(u *model.User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": fmt.Sprintf("%d", u.ID),
		"iat": now.Unix(),
		"exp": now.Add(time.Duration(s.config.JWT.ExpT) * time.Second).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWT.Key))
}

// Issues a fresh JWT when the given token's issued-at time is older than
// the configured renew_t threshold; returns an empty string when renewal
// is not yet needed or the token is invalid.
func (s *AuthService) RenewTokenIfNeeded(tokenStr string) string {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		return []byte(s.config.JWT.Key), nil
	})
	if err != nil || !token.Valid {
		return ""
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return ""
	}
	iatFloat, ok := claims["iat"].(float64)
	if !ok {
		return ""
	}
	age := time.Since(time.Unix(int64(iatFloat), 0))
	if age < time.Duration(s.config.JWT.RenewT)*time.Second {
		return ""
	}
	sub, _ := claims.GetSubject()
	var userID int64
	if _, err := fmt.Sscanf(sub, "%d", &userID); err != nil {
		return ""
	}
	u, err := s.db.GetUserByID(userID)
	if err != nil {
		return ""
	}
	newToken, err := s.generateJWT(u)
	if err != nil {
		return ""
	}
	return newToken
}

// Produces the password hash stored in the DB; delegates to the crypto
// package so the storage format and parameters are owned by one place.
func hashForDB(passwd, salt string) (string, error) {
	return crypto.HashPassword(passwd, salt)
}

// Reports whether the stored hash matches the given password and salt.
func verifyDBHash(dbHash, passwd, salt string) bool {
	return crypto.VerifyPassword(dbHash, passwd, salt)
}

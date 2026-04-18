package model

// Named string type representing a user's privilege level.
type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

// Database-backed user account; the Password field holds the bcrypt
// hash and is omitted from JSON responses.
type User struct {
	ID       int64    `json:"id"`
	Username string   `json:"username"`
	Password string   `json:"-"`
	Role     UserRole `json:"role"`
}

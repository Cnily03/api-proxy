package db

import (
	"database/sql"

	"api-proxy/internal/model"
)

// Returns all users ordered by ID ascending.
func (d *DB) ListUsers() ([]model.User, error) {
	rows, err := d.conn.Query(`SELECT id, username, password, role FROM users ORDER BY id ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.User, 0)
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Password, &u.Role); err != nil {
			return nil, err
		}
		items = append(items, u)
	}
	return items, rows.Err()
}

// Looks up a single user by username; returns sql.ErrNoRows when the
// user does not exist.
func (d *DB) GetUserByUsername(username string) (*model.User, error) {
	var u model.User
	err := d.conn.QueryRow(`SELECT id, username, password, role FROM users WHERE username=?`, username).
		Scan(&u.ID, &u.Username, &u.Password, &u.Role)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Looks up a single user by ID; returns sql.ErrNoRows when the user
// does not exist.
func (d *DB) GetUserByID(id int64) (*model.User, error) {
	var u model.User
	err := d.conn.QueryRow(`SELECT id, username, password, role FROM users WHERE id=?`, id).
		Scan(&u.ID, &u.Username, &u.Password, &u.Role)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Inserts a new user row and returns the generated auto-increment ID.
func (d *DB) CreateUser(u model.User) (int64, error) {
	res, err := d.conn.Exec(`INSERT INTO users (username, password, role) VALUES (?, ?, ?)`,
		u.Username, u.Password, u.Role)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Updates an existing user's username, password, and role by ID;
// returns sql.ErrNoRows when no matching row exists.
func (d *DB) UpdateUser(u model.User) error {
	res, err := d.conn.Exec(`UPDATE users SET username=?, password=?, role=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`,
		u.Username, u.Password, u.Role, u.ID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// Deletes the user with the given ID; returns sql.ErrNoRows when no
// matching row exists.
func (d *DB) DeleteUser(id int64) error {
	res, err := d.conn.Exec(`DELETE FROM users WHERE id=?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// Returns the total number of users in the database.
func (d *DB) CountUsers() (int, error) {
	var count int
	err := d.conn.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}

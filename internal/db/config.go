package db

import (
	"encoding/json"

	"api-proxy/internal/model"
)

// Loads the singleton server configuration row from server_config,
// decoding the JWT JSON and salt columns into a ServerConfig value.
// Assumes InitSchema (which runs migrations) has already populated the
// row's auth_jwt and salt fields.
func (d *DB) GetServerConfig() (*model.ServerConfig, error) {
	cfg := &model.ServerConfig{}
	var authJWT, salt string
	err := d.conn.QueryRow(`SELECT auth_jwt, salt FROM server_config WHERE id = 1`).Scan(&authJWT, &salt)
	if err != nil {
		return nil, err
	}
	if authJWT != "" {
		_ = json.Unmarshal([]byte(authJWT), &cfg.JWT)
	}
	cfg.Salt = salt
	return cfg, nil
}

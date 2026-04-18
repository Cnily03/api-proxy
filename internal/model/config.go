package model

// Configuration for signing and validating admin-panel JWTs.
// ExpT is the token lifetime in seconds; RenewT is the age after
// which a token is eligible for silent renewal.
type JWTConfig struct {
	Key    string `json:"key"`
	Alg    string `json:"alg"`
	ExpT   int64  `json:"exp_t"`
	RenewT int64  `json:"renew_t"`
}

// Persistent server-wide configuration loaded at startup: the JWT
// settings and the backend-side password salt.
type ServerConfig struct {
	JWT  JWTConfig `json:"jwt_config"`
	Salt string    `json:"salt"`
}

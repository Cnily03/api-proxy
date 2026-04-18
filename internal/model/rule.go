package model

// Reverse-proxy rule mapping an incoming Src path prefix to a
// destination URL, with optional API-key forwarding and TLS options.
type Rule struct {
	ID             int64    `json:"id"`
	Name           string   `json:"name"`
	Src            string   `json:"src"`
	Dest           string   `json:"dest"`
	DestAPIKey     string   `json:"dest_api_key"`
	SkipCertVerify bool     `json:"skip_cert_verify"`
	APIKey         string   `json:"api_key"`
	ForceAPIKey    bool     `json:"force_api_key"`
	Comment        string   `json:"comment"`
	Tags           []string `json:"tags"`
}

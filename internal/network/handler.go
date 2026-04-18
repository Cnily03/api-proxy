package network

import (
	"net/http"

	"api-proxy/internal/network/api"
	"api-proxy/internal/service"
)

// Builds the admin panel's top-level HTTP handler: mounts the /api
// routes and, when staticDir is non-empty, serves static files from
// that directory without directory listings; otherwise non-API paths
// return 404. proxyEndpoint is the public base URL of the reverse
// proxy returned by GET /api/proxy/endpoint.
func Setup(rules *service.RuleService, auth *service.AuthService, staticDir, proxyEndpoint string) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/api/", api.Setup(rules, auth, proxyEndpoint))

	if staticDir != "" {
		mux.Handle("/", newStaticServer(http.Dir(staticDir)))
	} else {
		mux.HandleFunc("/", http.NotFound)
	}

	return mux
}

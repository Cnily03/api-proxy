package proxy

import (
	"crypto/tls"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"api-proxy/internal/model"
	"api-proxy/internal/service"
)

// Reverse-proxy HTTP handler that routes requests according to matching
// rules, holding both verifying and non-verifying TLS transports so
// individual rules can opt out of certificate verification.
type Handler struct {
	service            *service.RuleService
	insecureHTTPClient http.RoundTripper
	secureHTTPClient   http.RoundTripper
}

// Builds the reverse-proxy handler, preparing its two TLS transports
// and wrapping it with permissive CORS middleware. A built-in /ping
// route short-circuits the pipeline (ahead of CORS and rule lookup)
// so health checks work regardless of configuration or method.
func NewHandler(s *service.RuleService) http.Handler {
	baseTransport := cloneDefaultTransport()
	insecureTransport := cloneDefaultTransport()
	insecureTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	h := &Handler{
		service:            s,
		secureHTTPClient:   baseTransport,
		insecureHTTPClient: insecureTransport,
	}
	return withCORS(withPing(http.HandlerFunc(h.serveHTTP)))
}

// Middleware that answers "pong"
func withPing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ping" {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("pong"))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Middleware that attaches permissive CORS headers so the reverse
// proxy can be called from browser contexts across origins, and
// short-circuits OPTIONS preflight requests.
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,PATCH,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Core request entry point: matches the URL against a rule, enforces
// the optional per-rule API-key requirement, then forwards the request
// using httputil.ReverseProxy.
func (h *Handler) serveHTTP(w http.ResponseWriter, r *http.Request) {
	rule, err := h.service.FindMatchByPath(r.URL.Path)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, "route lookup failed", http.StatusInternalServerError)
		return
	}

	// Enforce API key check when force_api_key is enabled
	token := extractBearerToken(r)
	if rule.ForceAPIKey && token != rule.APIKey {
		http.Error(w, "unauthorized: invalid api key", http.StatusUnauthorized)
		return
	}

	target, err := url.Parse(rule.Dest)
	if err != nil {
		http.Error(w, "invalid target route", http.StatusInternalServerError)
		return
	}

	proxy := &httputil.ReverseProxy{
		Director:      director(rule, target),
		Transport:     pickTransport(h, rule.SkipCertVerify),
		FlushInterval: -1,
		ErrorHandler: func(rw http.ResponseWriter, _ *http.Request, e error) {
			http.Error(rw, "proxy error: "+e.Error(), http.StatusBadGateway)
		},
	}
	proxy.ServeHTTP(w, r)
}

// Builds the ReverseProxy Director that rewrites scheme, host, path,
// and the Origin/Referer headers for a matched rule and destination.
func director(rule *model.Rule, target *url.URL) func(req *http.Request) {
	return func(req *http.Request) {
		originalPath := req.URL.Path
		newPath := joinTargetPath(target.Path, originalPath, rule.Src)

		rewriteAuthorization(req, rule)

		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = newPath
		req.URL.RawPath = ""
		req.Host = target.Host

		rewriteOriginAndReferer(req, target, newPath)
	}
}

// Rewrites the outbound Authorization header based on the rule:
// when the caller's bearer token matches rule.APIKey, replace it with
// rule.DestAPIKey (or strip it when DestAPIKey is empty); otherwise
// leave the header untouched.
func rewriteAuthorization(req *http.Request, rule *model.Rule) {
	if rule.DestAPIKey == "" && rule.APIKey != "" {
		// No any operation
		return
	}

	token := extractBearerToken(req)

	if token != rule.APIKey {
		// No rewrite when not match
		return
	}

	// Matched, apply the replacement
	if rule.DestAPIKey == "" {
		req.Header.Del("Authorization")
	} else {
		req.Header.Set("Authorization", "Bearer "+rule.DestAPIKey)
	}
}

// Updates the Origin and Referer headers to reflect the destination
// origin and rewritten path so upstreams see a consistent caller URL.
func rewriteOriginAndReferer(req *http.Request, target *url.URL, newPath string) {
	if req.Header.Get("Origin") != "" {
		req.Header.Set("Origin", target.Scheme+"://"+target.Host)
	}

	referer := req.Header.Get("Referer")
	if referer == "" {
		return
	}
	parsed, err := url.Parse(referer)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return
	}
	parsed.Scheme = target.Scheme
	parsed.Host = target.Host
	parsed.Path = newPath
	req.Header.Set("Referer", parsed.String())
}

// Joins the destination's base path with the portion of the incoming
// path following the rule's src prefix, producing the forwarded URL
// path while preserving leading-slash conventions.
func joinTargetPath(basePath string, incomingPath string, src string) string {
	base := basePath
	if base == "" {
		base = "/"
	}
	if !strings.HasPrefix(base, "/") {
		base = "/" + base
	}
	if base != "/" {
		base = strings.TrimRight(base, "/")
	}

	suffix := strings.TrimPrefix(incomingPath, src)
	if suffix == "" {
		return base
	}
	if !strings.HasPrefix(suffix, "/") {
		suffix = "/" + suffix
	}
	if base == "/" {
		return suffix
	}
	return base + suffix
}

// Extracts the bearer token from the Authorization header, returning
// an empty string when the header is missing or not a bearer token.
func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth != "" && strings.HasPrefix(strings.ToLower(auth), "bearer ") {
		return strings.TrimSpace(auth[len("Bearer "):])
	}
	return ""
}

// Selects the verifying or non-verifying TLS transport depending on
// the rule's SkipCertVerify setting.
func pickTransport(h *Handler, skipCertVerify bool) http.RoundTripper {
	if skipCertVerify {
		return h.insecureHTTPClient
	}
	return h.secureHTTPClient
}

// Returns a deep copy of http.DefaultTransport so per-handler tweaks
// (e.g., disabling TLS verification) don't leak into the global one.
func cloneDefaultTransport() *http.Transport {
	if t, ok := http.DefaultTransport.(*http.Transport); ok {
		return t.Clone()
	}
	return &http.Transport{}
}

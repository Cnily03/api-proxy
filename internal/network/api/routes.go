package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"api-proxy/internal/model"
	"api-proxy/internal/service"
)

// Collection of service dependencies used by the REST endpoints under
// /api, grouping rule and auth services on one struct for method
// receivers.
type Handler struct {
	rules *service.RuleService
	auth  *service.AuthService
}

// Registers all /api routes (health, auth, rules, users) on a fresh
// ServeMux and returns it, wiring each handler to the appropriate
// auth middleware.
func Setup(rules *service.RuleService, auth *service.AuthService) http.Handler {
	h := &Handler{rules: rules, auth: auth}
	mux := http.NewServeMux()

	// Health check (public)
	mux.HandleFunc("GET /api/healthz", h.handleHealth)

	// Auth (public)
	mux.HandleFunc("POST /api/auth/login", h.handleLogin)

	// Protected routes
	mux.HandleFunc("GET /api/auth/me", h.requireAuth(h.handleGetMe))
	mux.HandleFunc("POST /api/auth/password", h.requireAuth(h.handleChangePassword))

	// Rules (read: any user; write: admin)
	mux.HandleFunc("GET /api/rules", h.requireAuth(h.handleListRules))
	mux.HandleFunc("POST /api/rules", h.requireAdmin(h.handleCreateRule))
	mux.HandleFunc("PUT /api/rules/{id}", h.requireAdmin(h.handleUpdateRule))
	mux.HandleFunc("DELETE /api/rules/{id}", h.requireAdmin(h.handleDeleteRule))

	// Users (admin only)
	mux.HandleFunc("GET /api/users", h.requireAdmin(h.handleListUsers))
	mux.HandleFunc("POST /api/users", h.requireAdmin(h.handleCreateUser))
	mux.HandleFunc("PUT /api/users/{id}", h.requireAdmin(h.handleUpdateUser))
	mux.HandleFunc("DELETE /api/users/{id}", h.requireAdmin(h.handleDeleteUser))

	return mux
}

// ── Health ──────────────────────────────────────────────────

// Returns a trivial JSON {"status":"ok"} payload, used by readiness
// probes to verify the admin API is reachable.
func (h *Handler) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ── Auth middleware ─────────────────────────────────────────

// Middleware that validates the bearer token on the incoming request,
// attaches the resolved user to the request context, and emits a
// renewed token via the X-Set-Token response header when appropriate.
func (h *Handler) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			writeError(w, http.StatusUnauthorized, "authorization required")
			return
		}
		user, err := h.auth.ValidateToken(token)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		// Auto-renew token if past renew_t
		if newToken := h.auth.RenewTokenIfNeeded(token); newToken != "" {
			w.Header().Set("X-Set-Token", newToken)
		}
		r = r.WithContext(setUserCtx(r.Context(), user))
		next(w, r)
	}
}

// Middleware that composes requireAuth with a role check, rejecting
// authenticated non-admin users with HTTP 403.
func (h *Handler) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return h.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		user := getUserCtx(r.Context())
		if user == nil || user.Role != model.RoleAdmin {
			writeError(w, http.StatusForbidden, "admin access required")
			return
		}
		next(w, r)
	})
}

// ── Helpers ─────────────────────────────────────────────────

// Parses a positive int64 ID from a URL path parameter, returning a
// generic "invalid id" error on empty, non-numeric, or non-positive
// input.
func parseID(raw string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid id")
	}
	return id, nil
}

// Extracts the bearer token from the request's Authorization header,
// returning an empty string when absent or malformed.
func extractToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}

// Encodes the body as JSON with the given status code and the
// appropriate Content-Type header.
func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// Sends a plain-text error response using http.Error with the given
// status code and message.
func writeError(w http.ResponseWriter, status int, message string) {
	http.Error(w, message, status)
}

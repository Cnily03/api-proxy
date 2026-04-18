package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"api-proxy/internal/service"
)

// POST /api/auth/login endpoint: verifies credentials and, on success,
// returns the issued JWT via the X-Set-Token response header with a
// 204 status.
func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	token, err := h.auth.Login(body.Username, body.Password)
	if err != nil {
		if errors.Is(err, service.ErrTooManyAttempts) {
			writeError(w, http.StatusTooManyRequests, err.Error())
		} else {
			writeError(w, http.StatusUnauthorized, err.Error())
		}
		return
	}
	w.Header().Set("X-Set-Token", token)
	w.WriteHeader(http.StatusNoContent)
}

// GET /api/auth/me endpoint: returns the authenticated user resolved
// from the request context by requireAuth.
func (h *Handler) handleGetMe(w http.ResponseWriter, r *http.Request) {
	user := getUserCtx(r.Context())
	writeJSON(w, http.StatusOK, user)
}

// POST /api/auth/password endpoint: verifies the old password, then
// replaces it with the new one for the authenticated user.
func (h *Handler) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	user := getUserCtx(r.Context())
	var body struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if err := h.auth.ChangePassword(user.ID, body.OldPassword, body.NewPassword); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

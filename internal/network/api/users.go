package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"api-proxy/internal/model"
)

// GET /api/users endpoint (admin only): returns all user accounts
// in the system.
func (h *Handler) handleListUsers(w http.ResponseWriter, _ *http.Request) {
	users, err := h.auth.ListUsers()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, users)
}

// POST /api/users endpoint (admin only): creates a new user account
// with the given role (defaulting to user when unspecified) and
// returns the generated ID with a 201 status.
func (h *Handler) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Username string         `json:"username"`
		Password string         `json:"password"`
		Role     model.UserRole `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if body.Role == "" {
		body.Role = model.RoleUser
	}
	id, err := h.auth.CreateUser(body.Username, body.Password, body.Role)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

// PUT /api/users/{id} endpoint (admin only): updates the specified
// account's username, password, and/or role, returning 404 when the
// user does not exist.
func (h *Handler) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var body struct {
		Username string         `json:"username"`
		Password string         `json:"password"`
		Role     model.UserRole `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	if err := h.auth.UpdateUser(id, body.Username, body.Password, body.Role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// DELETE /api/users/{id} endpoint (admin only): removes the account
// identified by the path parameter, returning 404 when the user does
// not exist.
func (h *Handler) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.auth.DeleteUser(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"api-proxy/internal/model"
)

// GET /api/rules endpoint: returns all rules, blanking the destination
// and destination API key fields for non-admin callers.
func (h *Handler) handleListRules(w http.ResponseWriter, r *http.Request) {
	items, err := h.rules.List()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	user := getUserCtx(r.Context())
	if user.Role != model.RoleAdmin {
		for i := range items {
			items[i].Dest = ""
			items[i].DestAPIKey = ""
		}
	}
	writeJSON(w, http.StatusOK, items)
}

// POST /api/rules endpoint: decodes a Rule JSON body, creates the
// rule, and returns the generated ID with a 201 status.
func (h *Handler) handleCreateRule(w http.ResponseWriter, r *http.Request) {
	var item model.Rule
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	id, err := h.rules.Create(item)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

// PUT /api/rules/{id} endpoint: updates the rule identified by the
// path parameter, returning 404 when the rule does not exist.
func (h *Handler) handleUpdateRule(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	var item model.Rule
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json body")
		return
	}
	item.ID = id
	if err := h.rules.Update(item); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "rule not found")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// DELETE /api/rules/{id} endpoint: deletes the rule identified by
// the path parameter, returning 404 when the rule does not exist.
func (h *Handler) handleDeleteRule(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.rules.Delete(id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "rule not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

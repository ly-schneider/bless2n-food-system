package handler

import (
	"backend/internal/repository"
	"backend/internal/response"
	"encoding/json"
	"net/http"
	"strconv"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type AdminSessionsHandler struct {
	refresh repository.RefreshTokenRepository
	users   repository.UserRepository
}

func NewAdminSessionsHandler(rts repository.RefreshTokenRepository, users repository.UserRepository) *AdminSessionsHandler {
	return &AdminSessionsHandler{refresh: rts, users: users}
}

// List godoc
// @Summary List active session families
// @Tags admin-sessions
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit" minimum(1) maximum(200) default(50)
// @Param offset query int false "Offset" minimum(0) default(0)
// @Success 200 {object} map[string]interface{}
// @Router /v1/admin/sessions [get]
func (h *AdminSessionsHandler) List(w http.ResponseWriter, r *http.Request) {
	limit := 50
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	rows, total, err := h.refresh.ListActiveFamiliesAll(r.Context(), limit, offset)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to list sessions")
		return
	}
	// enrich with user email
	type RowDTO struct {
		UserID     string `json:"userId"`
		Email      string `json:"email"`
		FamilyID   string `json:"familyId"`
		Device     string `json:"device"`
		CreatedAt  any    `json:"createdAt"`
		LastUsedAt any    `json:"lastUsedAt"`
	}
	out := make([]RowDTO, 0, len(rows))
	cache := map[bson.ObjectID]string{}
	for _, rrow := range rows {
		email := cache[rrow.UserID]
		if email == "" {
			if u, err := h.users.FindByID(r.Context(), rrow.UserID); err == nil && u != nil {
				email = u.Email
			} else {
				email = ""
			}
			cache[rrow.UserID] = email
		}
		out = append(out, RowDTO{UserID: rrow.UserID.Hex(), Email: email, FamilyID: rrow.FamilyID, Device: rrow.Device, CreatedAt: rrow.CreatedAt, LastUsedAt: rrow.LastUsedAt})
	}
	response.WriteJSON(w, http.StatusOK, map[string]any{"items": out, "count": total})
}

// RevokeFamily godoc
// @Summary Revoke a session family for user
// @Tags admin-sessions
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param payload body revokeBody true "Family payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /v1/admin/users/{id}/sessions/revoke [post]
type revokeBody struct {
	FamilyID string `json:"familyId"`
}

func (h *AdminSessionsHandler) RevokeFamily(w http.ResponseWriter, r *http.Request) {
	uid := chiURLParam(r, "id")
	_, err := bson.ObjectIDFromHex(uid)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	var body revokeBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.FamilyID == "" {
		response.WriteError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	if err := h.refresh.RevokeFamily(r.Context(), body.FamilyID, "admin_revoked"); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to revoke")
		return
	}
	response.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// RevokeAll godoc
// @Summary Revoke all sessions for user
// @Tags admin-sessions
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Router /v1/admin/users/{id}/sessions/revoke-all [post]
func (h *AdminSessionsHandler) RevokeAll(w http.ResponseWriter, r *http.Request) {
	uid := chiURLParam(r, "id")
	uoid, err := bson.ObjectIDFromHex(uid)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid user id")
		return
	}
	if err := h.refresh.RevokeAllByUser(r.Context(), uoid, "admin_revoked_all"); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "failed to revoke all")
		return
	}
	response.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

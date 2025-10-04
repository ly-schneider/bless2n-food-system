package handler

import (
    "backend/internal/repository"
    "backend/internal/response"
    "backend/internal/domain"
    "net/http"
    "strconv"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type AdminUserHandler struct {
    users repository.UserRepository
}

func NewAdminUserHandler(users repository.UserRepository) *AdminUserHandler { return &AdminUserHandler{ users: users } }

func (h *AdminUserHandler) List(w http.ResponseWriter, r *http.Request) {
    limit := 50; offset := 0
    if v := r.URL.Query().Get("limit"); v != "" { if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 { limit = n } }
    if v := r.URL.Query().Get("offset"); v != "" { if n, err := strconv.Atoi(v); err == nil && n >= 0 { offset = n } }
    items, total, err := h.users.List(r.Context(), limit, offset)
    if err != nil { response.WriteError(w, http.StatusInternalServerError, "failed to list users"); return }
    // redact
    type UserDTO struct { ID string `json:"id"`; Email string `json:"email"`; Role string `json:"role"` }
    out := make([]UserDTO, 0, len(items))
    for _, u := range items { out = append(out, UserDTO{ ID: u.ID.Hex(), Email: u.Email, Role: string(u.Role) }) }
    response.WriteJSON(w, http.StatusOK, map[string]any{"items": out, "count": total})
}

// POST /v1/admin/users/{id}/promote - change role from customer to admin only
func (h *AdminUserHandler) Promote(w http.ResponseWriter, r *http.Request) {
    id := chiURLParam(r, "id"); oid, err := primitive.ObjectIDFromHex(id); if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid id"); return }
    u, err := h.users.FindByID(r.Context(), oid)
    if err != nil || u == nil { response.WriteError(w, http.StatusNotFound, "not found"); return }
    if u.Role == domain.UserRoleAdmin { response.WriteJSON(w, http.StatusOK, map[string]any{"ok": true}); return }
    // Only allow customer -> admin
    if u.Role != domain.UserRoleCustomer { response.WriteError(w, http.StatusBadRequest, "unsupported role change"); return }
    if err := h.users.UpdateRole(r.Context(), oid, domain.UserRoleAdmin); err != nil { response.WriteError(w, http.StatusInternalServerError, "update failed"); return }
    response.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

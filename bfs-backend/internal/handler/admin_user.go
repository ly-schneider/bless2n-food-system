package handler

import (
    "backend/internal/repository"
    "backend/internal/response"
    "backend/internal/domain"
    "net/http"
    "strconv"
    "encoding/json"
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type AdminUserHandler struct {
    users repository.UserRepository
}

func NewAdminUserHandler(users repository.UserRepository) *AdminUserHandler { return &AdminUserHandler{ users: users } }

// List godoc
// @Summary List users
// @Tags admin-users
// @Security BearerAuth
// @Produce json
// @Param limit query int false "Limit" minimum(1) maximum(200) default(50)
// @Param offset query int false "Offset" minimum(0) default(0)
// @Success 200 {object} map[string]interface{}
// @Router /v1/admin/users [get]
func (h *AdminUserHandler) List(w http.ResponseWriter, r *http.Request) {
    limit := 50; offset := 0
    if v := r.URL.Query().Get("limit"); v != "" { if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 { limit = n } }
    if v := r.URL.Query().Get("offset"); v != "" { if n, err := strconv.Atoi(v); err == nil && n >= 0 { offset = n } }
    items, total, err := h.users.List(r.Context(), limit, offset)
    if err != nil { response.WriteError(w, http.StatusInternalServerError, "failed to list users"); return }
    // redact
    type UserDTO struct {
        ID         string    `json:"id"`
        Email      string    `json:"email"`
        FirstName  string    `json:"firstName"`
        LastName   string    `json:"lastName"`
        Role       string    `json:"role"`
        IsVerified bool      `json:"isVerified"`
        CreatedAt  time.Time `json:"createdAt"`
        UpdatedAt  time.Time `json:"updatedAt"`
    }
    out := make([]UserDTO, 0, len(items))
    for _, u := range items {
        out = append(out, UserDTO{
            ID:         u.ID.Hex(),
            Email:      u.Email,
            FirstName:  u.FirstName,
            LastName:   u.LastName,
            Role:       string(u.Role),
            IsVerified: u.IsVerified,
            CreatedAt:  u.CreatedAt,
            UpdatedAt:  u.UpdatedAt,
        })
    }
    response.WriteJSON(w, http.StatusOK, map[string]any{"items": out, "count": total})
}

// Promote godoc
// @Summary Promote user to admin
// @Tags admin-users
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /v1/admin/users/{id}/promote [post]
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

// GetByID godoc
// @Summary Get user by ID
// @Tags admin-users
// @Security BearerAuth
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /v1/admin/users/{id} [get]
func (h *AdminUserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    id := chiURLParam(r, "id"); oid, err := primitive.ObjectIDFromHex(id); if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid id"); return }
    u, err := h.users.FindByID(r.Context(), oid)
    if err != nil || u == nil { response.WriteError(w, http.StatusNotFound, "not found"); return }
    type UserDTO struct {
        ID         string    `json:"id"`
        Email      string    `json:"email"`
        FirstName  string    `json:"firstName"`
        LastName   string    `json:"lastName"`
        Role       string    `json:"role"`
        IsVerified bool      `json:"isVerified"`
        CreatedAt  time.Time `json:"createdAt"`
        UpdatedAt  time.Time `json:"updatedAt"`
    }
    response.WriteJSON(w, http.StatusOK, map[string]any{"user": UserDTO{
        ID:         u.ID.Hex(),
        Email:      u.Email,
        FirstName:  u.FirstName,
        LastName:   u.LastName,
        Role:       string(u.Role),
        IsVerified: u.IsVerified,
        CreatedAt:  u.CreatedAt,
        UpdatedAt:  u.UpdatedAt,
    }})
}

// PatchRole godoc
// @Summary Update user role
// @Tags admin-users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param payload body patchRoleBody true "Role payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /v1/admin/users/{id}/role [patch]
type patchRoleBody struct { Role string `json:"role"` }
func (h *AdminUserHandler) PatchRole(w http.ResponseWriter, r *http.Request) {
    id := chiURLParam(r, "id"); oid, err := primitive.ObjectIDFromHex(id); if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid id"); return }
    var body patchRoleBody
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil || (body.Role != string(domain.UserRoleAdmin) && body.Role != string(domain.UserRoleCustomer)) {
        response.WriteError(w, http.StatusBadRequest, "invalid role"); return
    }
    role := domain.UserRole(body.Role)
    if err := h.users.UpdateRole(r.Context(), oid, role); err != nil { response.WriteError(w, http.StatusInternalServerError, "update failed"); return }
    response.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// PatchProfile godoc
// @Summary Update user profile
// @Description Admin can update email, names, role, and verification
// @Tags admin-users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param payload body adminPatchUserBody true "User payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /v1/admin/users/{id} [patch]
type adminPatchUserBody struct {
    Email      *string `json:"email,omitempty"`
    FirstName  *string `json:"firstName,omitempty"`
    LastName   *string `json:"lastName,omitempty"`
    Role       *string `json:"role,omitempty"`
    IsVerified *bool   `json:"isVerified,omitempty"`
}

func (h *AdminUserHandler) PatchProfile(w http.ResponseWriter, r *http.Request) {
    id := chiURLParam(r, "id")
    oid, err := primitive.ObjectIDFromHex(id)
    if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid id"); return }

    u, err := h.users.FindByID(r.Context(), oid)
    if err != nil || u == nil { response.WriteError(w, http.StatusNotFound, "not found"); return }

    var body adminPatchUserBody
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
        response.WriteError(w, http.StatusBadRequest, "invalid json")
        return
    }

    // Names
    if body.FirstName != nil || body.LastName != nil {
        if err := h.users.UpdateNames(r.Context(), oid, body.FirstName, body.LastName); err != nil {
            response.WriteError(w, http.StatusInternalServerError, "update names failed"); return
        }
        // refresh
        if nu, err := h.users.FindByID(r.Context(), oid); err == nil && nu != nil { u = nu }
    }
    // Email / verification
    if body.Email != nil || body.IsVerified != nil {
        newEmail := u.Email
        if body.Email != nil { newEmail = *body.Email }
        newVerified := u.IsVerified
        if body.IsVerified != nil { newVerified = *body.IsVerified }
        if err := h.users.UpdateEmail(r.Context(), oid, newEmail, newVerified); err != nil {
            response.WriteError(w, http.StatusInternalServerError, "update email failed"); return
        }
        if nu, err := h.users.FindByID(r.Context(), oid); err == nil && nu != nil { u = nu }
    }
    // Role
    if body.Role != nil {
        if *body.Role != string(domain.UserRoleAdmin) && *body.Role != string(domain.UserRoleCustomer) {
            response.WriteError(w, http.StatusBadRequest, "invalid role")
            return
        }
        if err := h.users.UpdateRole(r.Context(), oid, domain.UserRole(*body.Role)); err != nil {
            response.WriteError(w, http.StatusInternalServerError, "update role failed"); return
        }
        if nu, err := h.users.FindByID(r.Context(), oid); err == nil && nu != nil { u = nu }
    }

    type UserDTO struct {
        ID         string    `json:"id"`
        Email      string    `json:"email"`
        FirstName  string    `json:"firstName"`
        LastName   string    `json:"lastName"`
        Role       string    `json:"role"`
        IsVerified bool      `json:"isVerified"`
        CreatedAt  time.Time `json:"createdAt"`
        UpdatedAt  time.Time `json:"updatedAt"`
    }
    response.WriteJSON(w, http.StatusOK, map[string]any{"user": UserDTO{
        ID: u.ID.Hex(), Email: u.Email, FirstName: u.FirstName, LastName: u.LastName,
        Role: string(u.Role), IsVerified: u.IsVerified, CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt,
    }})
}

// Delete godoc
// @Summary Delete user
// @Tags admin-users
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]interface{}
// @Router /v1/admin/users/{id} [delete]
func (h *AdminUserHandler) Delete(w http.ResponseWriter, r *http.Request) {
    id := chiURLParam(r, "id")
    oid, err := primitive.ObjectIDFromHex(id)
    if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid id"); return }
    if err := h.users.DeleteByID(r.Context(), oid); err != nil {
        response.WriteError(w, http.StatusInternalServerError, "delete failed"); return
    }
    w.WriteHeader(http.StatusNoContent)
}

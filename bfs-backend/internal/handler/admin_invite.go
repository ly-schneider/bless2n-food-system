package handler

import (
    "backend/internal/domain"
    "backend/internal/middleware"
    "backend/internal/repository"
    "backend/internal/response"
    "backend/internal/utils"
    "backend/internal/service"
    "encoding/json"
    "net/http"
    "strconv"
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

type AdminInviteHandler struct {
    invites repository.AdminInviteRepository
    users   repository.UserRepository
    audit   repository.AuditRepository
    email   service.EmailService
}

func NewAdminInviteHandler(invites repository.AdminInviteRepository, users repository.UserRepository, audit repository.AuditRepository, email service.EmailService) *AdminInviteHandler {
    return &AdminInviteHandler{ invites: invites, users: users, audit: audit, email: email }
}

// GET /v1/admin/invites
func (h *AdminInviteHandler) List(w http.ResponseWriter, r *http.Request) {
    var status *string
    if v := r.URL.Query().Get("status"); v != "" { status = &v }
    var email *string
    if v := r.URL.Query().Get("email"); v != "" { email = &v }
    limit := 50; offset := 0
    if v := r.URL.Query().Get("limit"); v != "" { if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 { limit = n } }
    if v := r.URL.Query().Get("offset"); v != "" { if n, err := strconv.Atoi(v); err == nil && n >= 0 { offset = n } }
    items, total, err := h.invites.List(r.Context(), status, email, limit, offset)
    if err != nil { response.WriteError(w, http.StatusInternalServerError, "failed to list invites"); return }
    type InviteDTO struct { ID string `json:"id"`; Email string `json:"email"`; Status string `json:"status"`; ExpiresAt time.Time `json:"expiresAt"`; CreatedAt time.Time `json:"createdAt"`; UsedAt *time.Time `json:"usedAt,omitempty"` }
    out := make([]InviteDTO, 0, len(items))
    for _, it := range items { out = append(out, InviteDTO{ ID: it.ID.Hex(), Email: it.InviteeEmail, Status: it.Status, ExpiresAt: it.ExpiresAt, CreatedAt: it.CreatedAt, UsedAt: it.UsedAt }) }
    response.WriteJSON(w, http.StatusOK, map[string]any{"items": out, "count": total})
}

// POST /v1/admin/invites: create new ADMIN invite (admin-only)
type createInviteBody struct { Email string `json:"email"`; ExpiresInSec *int `json:"expiresInSec,omitempty"` }

func (h *AdminInviteHandler) Create(w http.ResponseWriter, r *http.Request) {
    claims, ok := middleware.GetUserFromContext(r.Context()); if !ok { response.WriteError(w, http.StatusUnauthorized, "unauthorized"); return }
    var body createInviteBody
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Email == "" { response.WriteError(w, http.StatusBadRequest, "invalid payload"); return }
    ttl := 7*24*3600
    if body.ExpiresInSec != nil && *body.ExpiresInSec > 0 && *body.ExpiresInSec <= 30*24*3600 {
        ttl = *body.ExpiresInSec
    }
    token, err := utils.GenerateRandomURLSafe(32); if err != nil { response.WriteError(w, http.StatusInternalServerError, "token error"); return }
    hash := utils.HashTokenSHA256(token)
    inviter, _ := primitive.ObjectIDFromHex(claims.Subject)
    inv, err := h.invites.Create(r.Context(), inviter, body.Email, hash, time.Now().UTC().Add(time.Duration(ttl)*time.Second))
    if err != nil { response.WriteError(w, http.StatusInternalServerError, "create failed"); return }
    _ = h.audit.Insert(r.Context(), &domain.AuditLog{ Action: domain.AuditCreate, EntityType: "admin_invite", EntityID: inv.ID.Hex(), After: inv, ActorUserID: objIDPtr(claims.Subject), ActorRole: strPtr(string(claims.Role)), RequestID: getRequestIDPtr(r) })
    // Send invitation email (best-effort)
    _ = h.email.SendAdminInvite(r.Context(), body.Email, token, inv.ExpiresAt)
    // Response: include token only for internal calls
    resp := map[string]any{"id": inv.ID.Hex()}
    if r.Header.Get("X-Internal-Call") == "1" { resp["token"] = token }
    response.WriteJSON(w, http.StatusCreated, resp)
}

// POST /v1/admin/invites/{id}/revoke
func (h *AdminInviteHandler) Revoke(w http.ResponseWriter, r *http.Request) {
    id := chiURLParam(r, "id"); oid, err := primitive.ObjectIDFromHex(id); if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid id"); return }
    ok, err := h.invites.Revoke(r.Context(), oid); if err != nil { response.WriteError(w, http.StatusInternalServerError, "revoke failed"); return }
    if !ok { response.WriteError(w, http.StatusBadRequest, "not pending"); return }
    _ = h.audit.Insert(r.Context(), &domain.AuditLog{ Action: domain.AuditUpdate, EntityType: "admin_invite", EntityID: id, After: map[string]any{"status": "revoked"}, RequestID: getRequestIDPtr(r) })
    response.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// POST /v1/invites/accept (public)
type acceptInviteBody struct { Token string `json:"token"`; FirstName *string `json:"firstName,omitempty"`; LastName *string `json:"lastName,omitempty"` }

func (h *AdminInviteHandler) Accept(w http.ResponseWriter, r *http.Request) {
    var body acceptInviteBody
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Token == "" { response.WriteError(w, http.StatusBadRequest, "invalid payload"); return }
    hash := utils.HashTokenSHA256(body.Token)
    inv, err := h.invites.FindByTokenHash(r.Context(), hash)
    if err != nil || inv == nil { response.WriteError(w, http.StatusUnauthorized, "invalid token"); return }
    if inv.Status != "pending" || time.Now().UTC().After(inv.ExpiresAt) { response.WriteError(w, http.StatusUnauthorized, "expired or used"); return }
    // Upsert or upgrade user to admin role
    if _, err := h.users.UpsertByEmailWithRole(r.Context(), inv.InviteeEmail, domain.UserRoleAdmin, true, body.FirstName, body.LastName); err != nil {
        response.WriteError(w, http.StatusInternalServerError, "user create failed"); return
    }
    if err := h.invites.MarkAccepted(r.Context(), inv.ID); err != nil { response.WriteError(w, http.StatusInternalServerError, "mark accepted failed"); return }
    _ = h.audit.Insert(r.Context(), &domain.AuditLog{ Action: domain.AuditUpdate, EntityType: "admin_invite", EntityID: inv.ID.Hex(), After: map[string]any{"status": "accepted"} })
    response.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// POST /v1/admin/invites/{id}/resend - rotate token and resend email
func (h *AdminInviteHandler) Resend(w http.ResponseWriter, r *http.Request) {
    id := chiURLParam(r, "id"); oid, err := primitive.ObjectIDFromHex(id); if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid id"); return }
    inv, err := h.invites.FindByID(r.Context(), oid)
    if err != nil || inv == nil { response.WriteError(w, http.StatusNotFound, "not found"); return }
    if inv.Status != "pending" { response.WriteError(w, http.StatusBadRequest, "invite not pending"); return }
    token, err := utils.GenerateRandomURLSafe(32); if err != nil { response.WriteError(w, http.StatusInternalServerError, "token error"); return }
    hash := utils.HashTokenSHA256(token)
    exp := time.Now().UTC().Add(7*24*time.Hour)
    if err := h.invites.UpdateToken(r.Context(), inv.ID, hash, exp); err != nil { response.WriteError(w, http.StatusInternalServerError, "update failed"); return }
    _ = h.email.SendAdminInvite(r.Context(), inv.InviteeEmail, token, exp)
    response.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

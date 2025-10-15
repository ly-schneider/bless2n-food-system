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
    // For issuing sessions on acceptance
    jwt     service.JWTService
    refresh repository.RefreshTokenRepository
}

func NewAdminInviteHandler(invites repository.AdminInviteRepository, users repository.UserRepository, audit repository.AuditRepository, email service.EmailService, jwt service.JWTService, refresh repository.RefreshTokenRepository) *AdminInviteHandler {
    return &AdminInviteHandler{ invites: invites, users: users, audit: audit, email: email, jwt: jwt, refresh: refresh }
}

// List godoc
// @Summary List admin invites
// @Tags admin-invites
// @Security BearerAuth
// @Produce json
// @Param status query string false "Filter by status"
// @Param email query string false "Filter by email"
// @Param limit query int false "Limit" minimum(1) maximum(200) default(50)
// @Param offset query int false "Offset" minimum(0) default(0)
// @Success 200 {object} map[string]interface{}
// @Router /v1/admin/invites [get]
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
    type InviteDTO struct {
        ID         string     `json:"id"`
        Email      string     `json:"email"`
        Status     string     `json:"status"`
        InvitedBy  string     `json:"invitedBy"`
        ExpiresAt  time.Time  `json:"expiresAt"`
        CreatedAt  time.Time  `json:"createdAt"`
        UsedAt     *time.Time `json:"usedAt,omitempty"`
    }
    out := make([]InviteDTO, 0, len(items))
    for _, it := range items {
        out = append(out, InviteDTO{
            ID:        it.ID.Hex(),
            Email:     it.InviteeEmail,
            Status:    it.Status,
            InvitedBy: it.InvitedBy.Hex(),
            ExpiresAt: it.ExpiresAt,
            CreatedAt: it.CreatedAt,
            UsedAt:    it.UsedAt,
        })
    }
    response.WriteJSON(w, http.StatusOK, map[string]any{"items": out, "count": total})
}

// Create godoc
// @Summary Create admin invite
// @Tags admin-invites
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body createInviteBody true "Invite payload"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /v1/admin/invites [post]
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

// Revoke godoc
// @Summary Revoke admin invite
// @Tags admin-invites
// @Security BearerAuth
// @Param id path string true "Invite ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /v1/admin/invites/{id}/revoke [post]
func (h *AdminInviteHandler) Revoke(w http.ResponseWriter, r *http.Request) {
    id := chiURLParam(r, "id"); oid, err := primitive.ObjectIDFromHex(id); if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid id"); return }
    ok, err := h.invites.Revoke(r.Context(), oid); if err != nil { response.WriteError(w, http.StatusInternalServerError, "revoke failed"); return }
    if !ok { response.WriteError(w, http.StatusBadRequest, "not pending"); return }
    _ = h.audit.Insert(r.Context(), &domain.AuditLog{ Action: domain.AuditUpdate, EntityType: "admin_invite", EntityID: id, After: map[string]any{"status": "revoked"}, RequestID: getRequestIDPtr(r) })
    response.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// Accept godoc
// @Summary Accept admin invite
// @Tags invites
// @Accept json
// @Produce json
// @Param payload body acceptInviteBody true "Accept payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /v1/invites/accept [post]
type acceptInviteBody struct { Token string `json:"token"`; FirstName *string `json:"firstName,omitempty"`; LastName *string `json:"lastName,omitempty"` }

func (h *AdminInviteHandler) Accept(w http.ResponseWriter, r *http.Request) {
    var body acceptInviteBody
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Token == "" { response.WriteError(w, http.StatusBadRequest, "invalid payload"); return }
    hash := utils.HashTokenSHA256(body.Token)
    inv, err := h.invites.FindByTokenHash(r.Context(), hash)
    if err != nil || inv == nil { response.WriteError(w, http.StatusUnauthorized, "invalid token"); return }
    if inv.Status != "pending" || time.Now().UTC().After(inv.ExpiresAt) { response.WriteError(w, http.StatusUnauthorized, "expired or used"); return }
    // Require name to create admin user
    if body.FirstName == nil || *body.FirstName == "" {
        response.WriteError(w, http.StatusBadRequest, "firstName required"); return
    }
    // Upsert or upgrade user to admin role (with provided name)
    u, err := h.users.UpsertByEmailWithRole(r.Context(), inv.InviteeEmail, domain.UserRoleAdmin, true, body.FirstName, body.LastName)
    if err != nil { response.WriteError(w, http.StatusInternalServerError, "user create failed"); return }
    if err := h.invites.MarkAccepted(r.Context(), inv.ID); err != nil { response.WriteError(w, http.StatusInternalServerError, "mark accepted failed"); return }
    _ = h.audit.Insert(r.Context(), &domain.AuditLog{ Action: domain.AuditUpdate, EntityType: "admin_invite", EntityID: inv.ID.Hex(), After: map[string]any{"status": "accepted"} })
    // Issue session (access + refresh) and set cookies similar to OTP flow
    if h.jwt != nil && h.refresh != nil {
        // Generate tokens and persist refresh token family
        access, err := h.jwt.GenerateAccessToken(u)
        if err != nil { response.WriteError(w, http.StatusInternalServerError, "token error"); return }
        rt, err := h.jwt.GenerateRefreshToken()
        if err != nil { response.WriteError(w, http.StatusInternalServerError, "token error"); return }
        family, err := utils.GenerateFamilyID()
        if err != nil { response.WriteError(w, http.StatusInternalServerError, "token error"); return }
        now := time.Now().UTC()
        // derive a friendly client id from headers (best-effort)
        clientID := r.Header.Get("X-Forwarded-User-Agent")
        if clientID == "" { clientID = r.Header.Get("X-Original-User-Agent") }
        if clientID == "" { clientID = r.Header.Get("X-Client-UA") }
        if clientID == "" { clientID = r.Header.Get("User-Agent") }
        if clientID == "" { clientID = "Invite" }
        if len(clientID) > 64 { clientID = clientID[:64] }
        if _, err := h.refresh.Create(r.Context(), &domain.RefreshToken{
            UserID:     u.ID,
            ClientID:   clientID,
            TokenHash:  utils.HashTokenSHA256(rt),
            IssuedAt:   now,
            LastUsedAt: time.Time{},
            ExpiresAt:  now.Add(service.RefreshTokenDuration),
            IsRevoked:  false,
            FamilyID:   family,
        }); err != nil {
            response.WriteError(w, http.StatusInternalServerError, "session error"); return
        }
        // Cookies: refresh (HttpOnly) and CSRF (non-HttpOnly)
        middleware.SetAuthCookie(w, r, utils.RefreshCookieName, rt, int(7*24*3600))
        csrf, _ := utils.GenerateCSRFToken()
        csrfName := utils.CSRFCookieName
        csrfSecure := middleware.IsHTTPS(r)
        if csrfSecure { csrfName = "__Host-" + csrfName }
        middleware.SetSecureCookie(w, middleware.SecureCookieOptions{
            Name:     csrfName,
            Value:    csrf,
            Path:     "/",
            MaxAge:   7*24*3600,
            HttpOnly: false,
            Secure:   csrfSecure,
            SameSite: http.SameSiteLaxMode,
        })
        // Response like other auth flows
        w.Header().Set("Content-Type", "application/json")
        resp := map[string]any{
            "access_token": access,
            "expires_in":   int64(service.AccessTokenDuration.Seconds()),
            "token_type":   "Bearer",
            "user":         map[string]any{"id": u.ID.Hex(), "email": u.Email, "role": string(u.Role)},
        }
        if r.Header.Get("X-Internal-Call") == "1" {
            resp["refresh_token"] = rt
            resp["csrf_token"] = csrf
        }
        _ = json.NewEncoder(w).Encode(resp)
        return
    }
    // Fallback
    response.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// Verify godoc
// @Summary Verify admin invite token
// @Tags invites
// @Accept json
// @Produce json
// @Param payload body verifyInviteBody true "Verify payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /v1/invites/verify [post]
type verifyInviteBody struct { Token string `json:"token"` }
func (h *AdminInviteHandler) Verify(w http.ResponseWriter, r *http.Request) {
    var body verifyInviteBody
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Token == "" { response.WriteError(w, http.StatusBadRequest, "invalid payload"); return }
    hash := utils.HashTokenSHA256(body.Token)
    inv, err := h.invites.FindByTokenHash(r.Context(), hash)
    if err != nil || inv == nil { response.WriteError(w, http.StatusUnauthorized, "invalid token"); return }
    status := inv.Status
    if status == "pending" && time.Now().UTC().After(inv.ExpiresAt) { status = "expired" }
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(map[string]any{
        "email":     inv.InviteeEmail,
        "expiresAt": inv.ExpiresAt,
        "status":    status,
    })
}

// Resend godoc
// @Summary Resend admin invite
// @Tags admin-invites
// @Security BearerAuth
// @Param id path string true "Invite ID"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /v1/admin/invites/{id}/resend [post]
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

// Delete godoc
// @Summary Delete admin invite
// @Tags admin-invites
// @Security BearerAuth
// @Param id path string true "Invite ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]interface{}
// @Router /v1/admin/invites/{id} [delete]
func (h *AdminInviteHandler) Delete(w http.ResponseWriter, r *http.Request) {
    id := chiURLParam(r, "id")
    oid, err := primitive.ObjectIDFromHex(id)
    if err != nil { response.WriteError(w, http.StatusBadRequest, "invalid id"); return }
    if err := h.invites.Delete(r.Context(), oid); err != nil { response.WriteError(w, http.StatusInternalServerError, "delete failed"); return }
    w.WriteHeader(http.StatusNoContent)
}

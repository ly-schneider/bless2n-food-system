package handler

import (
    "backend/internal/config"
    "backend/internal/service"
    "backend/internal/utils"
    "fmt"
    "net/http"
    "strings"
    "time"
)

type DevHandler struct {
    email service.EmailService
    cfg   config.Config
}

func NewDevHandler(email service.EmailService, cfg config.Config) *DevHandler {
    return &DevHandler{email: email, cfg: cfg}
}

// PreviewLoginEmail renders the login email (HTML by default, or text when format=text) for preview in non-prod.
func (h *DevHandler) PreviewLoginEmail(w http.ResponseWriter, r *http.Request) {
    // Only allow when not prod
    if strings.EqualFold(h.cfg.App.AppEnv, "prod") {
        http.NotFound(w, r)
        return
    }
    q := r.URL.Query()
    to := q.Get("email")
    if to == "" { to = "dev@example.com" }
    code := q.Get("code")
    if code == "" { if c, err := utils.GenerateOTP(); err == nil { code = c } else { code = "123456" } }
    subj, text, html, err := h.email.PreviewLoginEmail(r.Context(), to, code, clientIP(r), r.UserAgent(), 10*time.Minute)
    if err != nil {
        http.Error(w, fmt.Sprintf("render failed: %v", err), http.StatusInternalServerError)
        return
    }
    _ = subj // not used in preview response headers
    if strings.EqualFold(q.Get("format"), "text") {
        w.Header().Set("Content-Type", "text/plain; charset=utf-8")
        _, _ = w.Write([]byte(text))
        return
    }
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    _, _ = w.Write([]byte(html))
}

// PreviewEmailChangeEmail renders the email-change verification email for preview in non-prod.
func (h *DevHandler) PreviewEmailChangeEmail(w http.ResponseWriter, r *http.Request) {
    if strings.EqualFold(h.cfg.App.AppEnv, "prod") {
        http.NotFound(w, r)
        return
    }
    q := r.URL.Query()
    to := q.Get("newEmail")
    if to == "" { to = "new-email@example.com" }
    code := q.Get("code")
    if code == "" { if c, err := utils.GenerateOTP(); err == nil { code = c } else { code = "123456" } }
    subj, text, html, err := h.email.PreviewEmailChangeVerification(r.Context(), to, code, clientIP(r), r.UserAgent(), 15*time.Minute)
    if err != nil {
        http.Error(w, fmt.Sprintf("render failed: %v", err), http.StatusInternalServerError)
        return
    }
    _ = subj
    if strings.EqualFold(q.Get("format"), "text") {
        w.Header().Set("Content-Type", "text/plain; charset=utf-8")
        _, _ = w.Write([]byte(text))
        return
    }
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    _, _ = w.Write([]byte(html))
}

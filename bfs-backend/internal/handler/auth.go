package handler

import (
    "backend/internal/domain"
    "backend/internal/middleware"
    "backend/internal/service"
    "backend/internal/response"
    "backend/internal/utils"
    "encoding/json"
    "context"
    "time"
    "net"
    "net/http"
    "strings"
    "fmt"
    
    "github.com/go-chi/chi/v5"

    "github.com/go-playground/validator/v10"
)

// Reference response types to satisfy imports for Swagger type resolution
var _ = response.ProblemDetails{}

type AuthHandler struct {
    authService service.AuthService
    federatedService service.FederatedAuthService
    validator   *validator.Validate
}

func NewAuthHandler(authService service.AuthService, federated service.FederatedAuthService) *AuthHandler {
    return &AuthHandler{
        authService: authService,
        federatedService: federated,
        validator:   validator.New(),
    }
}

type requestOTPBody struct{
    Email string `json:"email" validate:"required,email"`
}

// RequestOTP godoc
// @Summary Request OTP code
// @Description Starts email OTP login flow. Always returns 202 to prevent user enumeration.
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body requestOTPBody true "Request payload"
// @Success 202 {object} map[string]string
// @Router /v1/auth/otp/request [post]
func (h *AuthHandler) RequestOTP(w http.ResponseWriter, r *http.Request) {
    var body requestOTPBody
    _ = json.NewDecoder(r.Body).Decode(&body)
    // Always respond 202 to avoid enumeration
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusAccepted)
    // Perform async work detached from the request context so it
    // isn't canceled when the handler returns 202 Accepted.
    go func(email, ip, ua string){
        ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
        defer cancel()
        _ = h.authService.RequestOTP(ctx, email, ip, ua)
    }(body.Email, clientIP(r), clientIDFromRequest(r))
    _, _ = w.Write([]byte(`{"message":"If the email is registered, you'll receive a code shortly."}`))
}

type verifyOTPBody struct{
    Email string `json:"email"`
    OTP   string `json:"otp"`
}

// AuthTokenResponse represents the successful auth response.
type AuthTokenResponse struct {
    AccessToken  string       `json:"access_token"`
    ExpiresIn    int64        `json:"expires_in"`
    TokenType    string       `json:"token_type"`
    User         *serviceUser `json:"user"`
}

// VerifyOTP godoc
// @Summary Verify OTP code
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body verifyOTPBody true "Verification payload"
// @Success 200 {object} AuthTokenResponse
// @Failure 400 {object} response.ProblemDetails
// @Failure 401 {object} response.ProblemDetails
// @Router /v1/auth/otp/verify [post]
func (h *AuthHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
    var body verifyOTPBody
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }
    clientID := clientIDFromRequest(r)
    var pair *service.TokenPairResponse
    var user *serviceUser
    if body.Email != "" && body.OTP != "" {
        p, u, err := h.authService.VerifyWithCode(r.Context(), body.Email, body.OTP, clientID)
        if err != nil { http.Error(w, "Unauthorized", http.StatusUnauthorized); return }
        pair, user = p, toServiceUser(u)
    } else {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }
    // Set cookies: refresh (HttpOnly) and CSRF (non-HttpOnly)
    // Use HTTPS (__Host-) when available; fall back to non-secure names/flags on localhost HTTP.
    middleware.SetAuthCookie(w, r, utils.RefreshCookieName, pair.RefreshToken, int(7*24*3600))
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
    w.Header().Set("Content-Type", "application/json")
    resp := map[string]any{
        "access_token": pair.AccessToken,
        "expires_in":   pair.ExpiresIn,
        "token_type":   pair.TokenType,
        "user":         user,
    }
    if r.Header.Get("X-Internal-Call") == "1" {
        resp["refresh_token"] = pair.RefreshToken
    }
    _ = json.NewEncoder(w).Encode(resp)
}

// Federated: Google (Authorization Code + PKCE)
type verifyGoogleCodeBody struct {
    Code         string `json:"code"`
    CodeVerifier string `json:"code_verifier"`
    RedirectURI  string `json:"redirect_uri"`
    Nonce        string `json:"nonce,omitempty"`
}

// GoogleCode godoc
// @Summary Sign in with Google (Authorization Code + PKCE)
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body verifyGoogleCodeBody true "Google code payload"
// @Success 200 {object} AuthTokenResponse
// @Failure 400 {object} response.ProblemDetails
// @Failure 401 {object} response.ProblemDetails
// @Router /v1/auth/google/code [post]
func (h *AuthHandler) GoogleCode(w http.ResponseWriter, r *http.Request) {
    var body verifyGoogleCodeBody
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Code == "" || body.CodeVerifier == "" {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }
    clientID := clientIDFromRequest(r)
    pair, user, err := h.federatedService.SignInWithGoogleCode(r.Context(), body.Code, body.CodeVerifier, body.RedirectURI, body.Nonce, clientID)
    if err != nil { http.Error(w, "Unauthorized", http.StatusUnauthorized); return }
    // Cookies
    middleware.SetAuthCookie(w, r, utils.RefreshCookieName, pair.RefreshToken, int(7*24*3600))
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
    w.Header().Set("Content-Type", "application/json")
    resp := map[string]any{
        "access_token": pair.AccessToken,
        "expires_in":   pair.ExpiresIn,
        "token_type":   pair.TokenType,
        "user":         toServiceUser(user),
    }
    if r.Header.Get("X-Internal-Call") == "1" {
        resp["refresh_token"] = pair.RefreshToken
    }
    _ = json.NewEncoder(w).Encode(resp)
}

// (Apple implementation removed)

// Refresh godoc
// @Summary Refresh access token
// @Description Uses HttpOnly refresh cookie to mint new tokens and CSRF cookie
// @Tags auth
// @Produce json
// @Success 200 {object} AuthTokenResponse
// @Failure 401 {object} response.ProblemDetails
// @Router /v1/auth/refresh [post]
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
    // NOTE: No CSRF protection required for refresh endpoint since:
    // 1. It's used to establish new CSRF tokens (chicken-egg problem)
    // 2. Already protected by HttpOnly refresh token cookie
    // 3. Prevents deadlock when CSRF tokens expire
    
    // Refresh cookie required (supports both HTTPS and local HTTP cookie names)
    rtCookie, err := r.Cookie("__Host-" + utils.RefreshCookieName)
    if err != nil {
        rtCookie, err = r.Cookie(utils.RefreshCookieName)
    }
    if err != nil || rtCookie.Value == "" { http.Error(w, "Unauthorized", http.StatusUnauthorized); return }
    pair, user, err := h.authService.Refresh(r.Context(), rtCookie.Value, clientIDFromRequest(r))
    if err != nil {
        // clear cookie on reuse or invalid
        middleware.ClearSecureAuthCookie(w, utils.RefreshCookieName)
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    // Rotate cookies
    middleware.SetAuthCookie(w, r, utils.RefreshCookieName, pair.RefreshToken, int(7*24*3600))
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
    w.Header().Set("Content-Type", "application/json")
    resp := map[string]any{
        "access_token": pair.AccessToken,
        "expires_in":   pair.ExpiresIn,
        "token_type":   pair.TokenType,
        "user":         toServiceUser(user),
    }
    if r.Header.Get("X-Internal-Call") == "1" {
        resp["refresh_token"] = pair.RefreshToken
    }
    _ = json.NewEncoder(w).Encode(resp)
}

// Logout godoc
// @Summary Logout current session
// @Description Clears refresh and CSRF cookies and revokes refresh token family.
// @Tags auth
// @Success 204 "No Content"
// @Failure 403 {object} response.ProblemDetails
// @Router /v1/auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
    csrfHeader := r.Header.Get(utils.CSRFHeaderName)
    csrfCookie, _ := r.Cookie("__Host-" + utils.CSRFCookieName)
    if csrfCookie == nil {
        csrfCookie, _ = r.Cookie(utils.CSRFCookieName)
    }
    if csrfHeader == "" || csrfCookie == nil || csrfHeader != csrfCookie.Value {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }
    rtCookie, err := r.Cookie("__Host-" + utils.RefreshCookieName)
    if err != nil {
        rtCookie, err = r.Cookie(utils.RefreshCookieName)
    }
    if err == nil && rtCookie.Value != "" {
        _ = h.authService.Logout(r.Context(), rtCookie.Value)
    }
    middleware.ClearAuthCookie(w, r, utils.RefreshCookieName)
    // Also clear CSRF
    csrfName := utils.CSRFCookieName
    csrfSecure := middleware.IsHTTPS(r)
    if csrfSecure { csrfName = "__Host-" + csrfName }
    middleware.SetSecureCookie(w, middleware.SecureCookieOptions{
        Name:     csrfName,
        Value:    "",
        Path:     "/",
        MaxAge:   -1,
        HttpOnly: false,
        Secure:   csrfSecure,
        SameSite: http.SameSiteLaxMode,
    })
    w.WriteHeader(http.StatusNoContent)
}

// RevokeSession revokes a session family by id
// RevokeSession godoc
// @Summary Revoke a session family
// @Tags auth
// @Security BearerAuth
// @Param id path string true "Session family ID"
// @Success 204 "No Content"
// @Failure 401 {object} response.ProblemDetails
// @Router /v1/users/me/sessions/{id} [delete]
func (h *AuthHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
    claims, ok := middleware.GetUserFromContext(r.Context())
    if !ok { http.Error(w, "Unauthorized", http.StatusUnauthorized); return }
    familyID := chi.URLParam(r, "id")
    if familyID == "" {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }
    if err := h.authService.RevokeSessionFamily(r.Context(), claims.Subject, familyID); err != nil {
        http.Error(w, "Failed to revoke", http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}

// RevokeAllSessions revokes all refresh token families for the authenticated user
// RevokeAllSessions godoc
// @Summary Revoke all sessions for user
// @Tags auth
// @Security BearerAuth
// @Success 204 "No Content"
// @Failure 401 {object} response.ProblemDetails
// @Router /v1/users/me/sessions [delete]
func (h *AuthHandler) RevokeAllSessions(w http.ResponseWriter, r *http.Request) {
    claims, ok := middleware.GetUserFromContext(r.Context())
    if !ok { http.Error(w, "Unauthorized", http.StatusUnauthorized); return }
    if err := h.authService.RevokeAllSessions(r.Context(), claims.Subject); err != nil {
        http.Error(w, "Failed to revoke all", http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusNoContent)
}

// Sessions lists active session families (devices) for the authenticated user.
// It groups refresh tokens by family and exposes minimal metadata.
// Sessions godoc
// @Summary List active sessions
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} response.ProblemDetails
// @Router /v1/users/me/sessions [get]
func (h *AuthHandler) Sessions(w http.ResponseWriter, r *http.Request) {
    claims, ok := middleware.GetUserFromContext(r.Context())
    if !ok { http.Error(w, "Unauthorized", http.StatusUnauthorized); return }
    sessions, err := h.authService.ListUserActiveSessions(r.Context(), claims.Subject)
    if err != nil { http.Error(w, "Failed to load sessions", http.StatusInternalServerError); return }
    // Mark current session heuristically via UA
    ua := clientIDFromRequest(r)
    for i := range sessions {
        if dev, ok := sessions[i]["device"].(string); ok && dev == ua {
            sessions[i]["current"] = true
        }
    }
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(map[string]any{"sessions": sessions})
}

// Me returns basic profile; assumes JWT middleware filled context
// Me godoc
// @Summary Current auth claims
// @Tags auth
// @Security BearerAuth
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
    claims, ok := middleware.GetUserFromContext(r.Context())
    if !ok { http.Error(w, "Unauthorized", http.StatusUnauthorized); return }
    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(map[string]any{
        "sub":   claims.Subject,
        "role":  claims.Role,
        "aud":   claims.Audience,
        "iss":   claims.Issuer,
        "exp":   claims.ExpiresAt.Time,
    })
}

// Helpers
type serviceUser struct{
    ID    string `json:"id"`
    Email string `json:"email"`
    Role  string `json:"role"`
}
func toServiceUser(u *domain.User) *serviceUser { return &serviceUser{ID: u.ID.Hex(), Email: u.Email, Role: string(u.Role)} }

func clientIDFromRequest(r *http.Request) string {
    // Prefer forwarded browser UA when requests come via a server proxy
    ua := r.Header.Get("X-Forwarded-User-Agent")
    if ua == "" {
        ua = r.Header.Get("X-Original-User-Agent")
    }
    if ua == "" {
        ua = r.Header.Get("X-Client-UA")
    }
    if ua == "" {
        ua = r.Header.Get("User-Agent")
    }
    label := friendlyDeviceLabel(ua)
    if len(label) > 64 { label = label[:64] }
    return label
}
func clientIP(r *http.Request) string {
    if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
        // take the first IP in the list
        parts := strings.Split(xff, ",")
        ip := strings.TrimSpace(parts[0])
        if ip == "::1" { return "localhost" }
        return ip
    }
    if xr := r.Header.Get("X-Real-IP"); xr != "" { return xr }
    host, _, _ := net.SplitHostPort(r.RemoteAddr)
    if host == "::1" || host == "127.0.0.1" { return "localhost" }
    return host
}

// friendlyDeviceLabel maps a raw User-Agent to a human-friendly device label.
func friendlyDeviceLabel(ua string) string {
    s := strings.ToLower(strings.TrimSpace(ua))
    if s == "" {
        return "Unknown Device"
    }
    // Server SDKs
    if s == "node" || strings.Contains(s, "node ") || strings.Contains(s, "node/") {
        return "Server (Node)"
    }
    if strings.Contains(s, "undici") {
        return "Server (Undici)"
    }
    if strings.Contains(s, "axios") {
        return "Server (Axios)"
    }
    if strings.Contains(s, "python-requests") {
        return "Server (Python)"
    }

    // OS detection
    os := "Unknown OS"
    switch {
    case strings.Contains(s, "windows nt"):
        os = "Windows"
    case strings.Contains(s, "mac os x") || strings.Contains(s, "macintosh"):
        os = "macOS"
    case strings.Contains(s, "android"):
        os = "Android"
    case strings.Contains(s, "iphone") || strings.Contains(s, "ipad") || strings.Contains(s, "cpu iphone os"):
        os = "iOS"
    case strings.Contains(s, "linux"):
        os = "Linux"
    }

    // Browser detection (rough)
    browser := ""
    switch {
    case strings.Contains(s, "edg/") || strings.Contains(s, "edg "):
        browser = "Edge"
    case strings.Contains(s, "chrome/") && !strings.Contains(s, "chromium"):
        browser = "Chrome"
    case strings.Contains(s, "safari") && !strings.Contains(s, "chrome"):
        browser = "Safari"
    case strings.Contains(s, "firefox"):
        browser = "Firefox"
    case strings.Contains(s, "chromium"):
        browser = "Chromium"
    case strings.Contains(s, "opr/") || strings.Contains(s, "opera"):
        browser = "Opera"
    }

    if browser != "" {
        return fmt.Sprintf("%s on %s", browser, os)
    }
    // Fallback to raw UA truncated
    if len(ua) > 64 { ua = ua[:64] }
    return ua
}

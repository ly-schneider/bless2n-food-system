package auth

import (
	"backend/internal/generated/ent/devicebinding"
	"backend/internal/repository"
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	// Session refresh constants matching Better Auth config
	sessionUpdateAge = 24 * time.Hour      // 1 day
	sessionExpiresIn = 90 * 24 * time.Hour // 90 days

	// Cookie names for Better Auth session tokens
	cookieNameDev  = "better-auth.session_token"
	cookieNameProd = "__Secure-better-auth.session_token"
)

// SessionMiddleware validates Better Auth session tokens by looking up sessions
// in Postgres. Supports both cookie-based (browser) and Bearer token (API) auth.
// Also supports device token detection for combined user/device auth scenarios.
type SessionMiddleware struct {
	sessionRepo repository.SessionRepository
	bindingRepo repository.DeviceBindingRepository
	logger      *zap.Logger
	cookieName  string // optional override from config
}

func NewSessionMiddleware(sessionRepo repository.SessionRepository, bindingRepo repository.DeviceBindingRepository, logger *zap.Logger) *SessionMiddleware {
	return &SessionMiddleware{
		sessionRepo: sessionRepo,
		bindingRepo: bindingRepo,
		logger:      logger,
	}
}

// SetCookieName allows overriding the session cookie name.
func (m *SessionMiddleware) SetCookieName(name string) {
	m.cookieName = name
}

// RequireAuth returns middleware that requires a valid session.
func (m *SessionMiddleware) RequireAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, err := m.validateAndRefresh(r)
			if err != nil {
				m.handleAuthError(w, err)
				return
			}
			ctx = WithAuthType(ctx, AuthTypeUser)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth returns middleware that validates a session if present but doesn't require it.
// For Bearer tokens, it first checks if the token is a device token (POS/Station) before
// treating it as a user session.
func (m *SessionMiddleware) OptionalAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if this is a Bearer token that might be a device token
			if bearerToken, err := ExtractBearerToken(r); err == nil && m.bindingRepo != nil {
				tokenHash := repository.HashToken(bearerToken)
				if binding, bindErr := m.bindingRepo.GetByTokenHash(r.Context(), tokenHash); bindErr == nil {
					// This is a device token - validate the underlying session and set device context
					session, sessErr := m.sessionRepo.GetByToken(r.Context(), bearerToken)
					if sessErr != nil {
						m.handleAuthError(w, ErrInvalidToken)
						return
					}

					ctx := r.Context()
					ctx = WithAuthType(ctx, AuthTypeDevice)
					if binding.DeviceID != nil {
						ctx = WithDeviceID(ctx, *binding.DeviceID)
					} else {
						ctx = WithDeviceID(ctx, binding.ID)
					}
					switch binding.DeviceType {
					case devicebinding.DeviceTypePOS:
						ctx = WithDeviceType(ctx, DeviceTypePOS)
					case devicebinding.DeviceTypeSTATION:
						ctx = WithDeviceType(ctx, DeviceTypeStation)
					}
					ctx = WithUserID(ctx, session.UserID)
					if session.Role != "" {
						ctx = WithUserRole(ctx, string(session.Role))
					}

					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}

			// Not a device token or no binding repo - try regular session auth
			ctx, err := m.validateAndRefresh(r)
			if err != nil {
				if errors.Is(err, ErrMissingToken) {
					// No session - continue without auth
					next.ServeHTTP(w, r)
					return
				}
				m.handleAuthError(w, err)
				return
			}
			ctx = WithAuthType(ctx, AuthTypeUser)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole returns middleware that requires a specific role.
func (m *SessionMiddleware) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := GetUserRole(r.Context())
			if !ok || userRole != role {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequirePermission returns middleware that checks the user has a specific permission.
// Uses the RBAC role->permission mapping from rbac.go.
func (m *SessionMiddleware) RequirePermission(perm Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := GetUserRole(r.Context())
			if !ok {
				http.Error(w, "Forbidden: no role", http.StatusForbidden)
				return
			}
			if !HasPermission(Role(userRole), perm) {
				m.logger.Debug("permission denied",
					zap.String("role", userRole),
					zap.String("permission", string(perm)),
				)
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// extractSessionToken extracts the session token from the request.
// Checks cookies first (browser), then falls back to Authorization: Bearer header.
func (m *SessionMiddleware) extractSessionToken(r *http.Request) (string, error) {
	// Try custom cookie name first (if configured)
	if m.cookieName != "" {
		if c, err := r.Cookie(m.cookieName); err == nil && c.Value != "" {
			return decodeCookieValue(c.Value), nil
		}
	}

	// Try production secure cookie
	if c, err := r.Cookie(cookieNameProd); err == nil && c.Value != "" {
		return decodeCookieValue(c.Value), nil
	}

	// Try development cookie
	if c, err := r.Cookie(cookieNameDev); err == nil && c.Value != "" {
		return decodeCookieValue(c.Value), nil
	}

	// Fall back to Authorization: Bearer header
	token, err := extractToken(r)
	if err != nil {
		return "", ErrMissingToken
	}
	return token, nil
}

// decodeCookieValue URL-decodes a cookie value and strips the HMAC signature.
// Better Auth session cookies are formatted as "token.signature" where:
// - "token" is the session token stored in the database
// - "signature" is an HMAC that Better Auth uses for cookie integrity
// The Go backend only needs the token portion for the DB lookup.
// The base64 characters (+, /, =) may also be URL-encoded by the browser.
func decodeCookieValue(v string) string {
	decoded, err := url.QueryUnescape(v)
	if err != nil {
		decoded = v
	}
	// Strip the HMAC signature suffix (everything after the last dot)
	if idx := strings.LastIndex(decoded, "."); idx > 0 {
		return decoded[:idx]
	}
	return decoded
}

// validateAndRefresh validates the session token and performs sliding refresh if needed.
func (m *SessionMiddleware) validateAndRefresh(r *http.Request) (context.Context, error) {
	tokenString, err := m.extractSessionToken(r)
	if err != nil {
		return nil, err
	}

	session, err := m.sessionRepo.GetByToken(r.Context(), tokenString)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidToken
		}
		m.logger.Error("session lookup failed", zap.Error(err))
		return nil, ErrInvalidToken
	}

	// Perform sliding session refresh if session.UpdatedAt is older than updateAge
	if session.UpdatedAt.Add(sessionUpdateAge).Before(time.Now().UTC()) {
		if refreshErr := m.sessionRepo.RefreshSession(r.Context(), tokenString, sessionExpiresIn); refreshErr != nil {
			// Log but don't fail the request - the session is still valid
			m.logger.Warn("failed to refresh session", zap.Error(refreshErr))
		}
	}

	ctx := r.Context()
	ctx = WithUserID(ctx, session.UserID)
	if session.Role != "" {
		ctx = WithUserRole(ctx, string(session.Role))
	}
	if session.Email != "" {
		ctx = WithUserEmail(ctx, session.Email)
	}
	if session.Name != "" {
		ctx = WithUserName(ctx, session.Name)
	}

	return ctx, nil
}

func (m *SessionMiddleware) handleAuthError(w http.ResponseWriter, err error) {
	w.Header().Set("WWW-Authenticate", `Bearer realm="api"`)

	switch {
	case errors.Is(err, ErrMissingToken):
		http.Error(w, "Unauthorized: missing session", http.StatusUnauthorized)
	case errors.Is(err, ErrTokenExpired):
		http.Error(w, "Unauthorized: session expired", http.StatusUnauthorized)
	default:
		m.logger.Debug("auth error", zap.Error(err))
		http.Error(w, "Unauthorized: invalid session", http.StatusUnauthorized)
	}
}

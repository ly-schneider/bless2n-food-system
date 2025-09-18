package middleware

import (
	"backend/internal/service"
	"context"
	"net/http"
	"slices"
	"strings"

	"go.uber.org/zap"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
)

type JWTMiddleware struct {
	jwtService service.JWTService
}

func NewJWTMiddleware(jwtService service.JWTService) *JWTMiddleware {
	return &JWTMiddleware{
		jwtService: jwtService,
	}
}

func (m *JWTMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := m.extractToken(r)
		if token == "" {
			m.writeErrorResponse(w, http.StatusUnauthorized, "Missing authorization token")
			return
		}

		claims, err := m.jwtService.ValidateAccessToken(token)
		if err != nil {
			zap.L().Debug("token validation failed", zap.Error(err))
			m.writeErrorResponse(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		// Add user info to context
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *JWTMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := m.extractToken(r)
		if token != "" {
			if claims, err := m.jwtService.ValidateAccessToken(token); err == nil {
				ctx := context.WithValue(r.Context(), UserContextKey, claims)
				r = r.WithContext(ctx)
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (m *JWTMiddleware) RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(UserContextKey).(*service.TokenClaims)
			if !ok {
				m.writeErrorResponse(w, http.StatusUnauthorized, "Authentication required")
				return
			}

			userRole := string(claims.Role)
			hasRole := slices.Contains(roles, userRole)

			if !hasRole {
				zap.L().Warn("insufficient permissions",
					zap.String("user_id", claims.Subject),
					zap.String("user_role", userRole),
					zap.Strings("required_roles", roles))
				m.writeErrorResponse(w, http.StatusForbidden, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func (m *JWTMiddleware) extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}

	return ""
}

func (m *JWTMiddleware) writeErrorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	response := `{"error": true, "message": "` + message + `", "status": ` +
		strings.TrimSpace(strings.Fields(http.StatusText(status))[0]) + `}`
	_, _ = w.Write([]byte(response))
}

func GetUserFromContext(ctx context.Context) (*service.TokenClaims, bool) {
	user, ok := ctx.Value(UserContextKey).(*service.TokenClaims)
	return user, ok
}

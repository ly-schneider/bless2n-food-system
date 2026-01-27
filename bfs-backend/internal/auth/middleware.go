package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

var (
	ErrMissingToken   = errors.New("missing authorization token")
	ErrInvalidToken   = errors.New("invalid token")
	ErrTokenExpired   = errors.New("token expired")
	ErrInvalidIssuer  = errors.New("invalid token issuer")
	ErrInvalidSubject = errors.New("invalid token subject")
)

// NeonAuthMiddleware validates JWTs from Neon Auth.
type NeonAuthMiddleware struct {
	jwksClient *JWKSClient
	issuer     string // Expected issuer (origin of NEON_AUTH_URL)
	audience   string // Optional audience to validate
	logger     *zap.Logger
}

// NewNeonAuthMiddleware creates a new Neon Auth middleware.
func NewNeonAuthMiddleware(jwksClient *JWKSClient, neonAuthURL string, audience string, logger *zap.Logger) *NeonAuthMiddleware {
	// Extract issuer from URL (the origin)
	parsedURL, err := url.Parse(neonAuthURL)
	if err != nil {
		logger.Fatal("invalid NEON_AUTH_URL", zap.Error(err))
	}
	issuer := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)

	return &NeonAuthMiddleware{
		jwksClient: jwksClient,
		issuer:     issuer,
		audience:   audience,
		logger:     logger,
	}
}

// RequireAuth returns middleware that requires a valid JWT.
func (m *NeonAuthMiddleware) RequireAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, err := m.validateRequest(r)
			if err != nil {
				m.handleAuthError(w, err)
				return
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth returns middleware that validates JWT if present but doesn't require it.
func (m *NeonAuthMiddleware) OptionalAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, err := m.validateRequest(r)
			if err != nil {
				// On optional auth, just continue without auth context if token is missing
				if errors.Is(err, ErrMissingToken) {
					next.ServeHTTP(w, r)
					return
				}
				// But still fail on invalid tokens
				m.handleAuthError(w, err)
				return
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole returns middleware that requires a specific role.
func (m *NeonAuthMiddleware) RequireRole(role string) func(http.Handler) http.Handler {
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

func (m *NeonAuthMiddleware) validateRequest(r *http.Request) (context.Context, error) {
	tokenString, err := extractToken(r)
	if err != nil {
		return nil, err
	}

	claims, err := m.validateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Add user info to context
	ctx := r.Context()
	ctx = WithUserID(ctx, claims.Subject)

	// Extract role from claims
	if role := extractRoleFromClaims(claims); role != "" {
		ctx = WithUserRole(ctx, role)
	}

	return ctx, nil
}

func extractToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", ErrMissingToken
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", ErrInvalidToken
	}

	return parts[1], nil
}

func (m *NeonAuthMiddleware) validateToken(tokenString string) (*customClaims, error) {
	// Parse without validation first to get the key ID
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	unverified, _, err := parser.ParseUnverified(tokenString, &jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("%w: parse error", ErrInvalidToken)
	}

	kid, ok := unverified.Header["kid"].(string)
	if !ok {
		return nil, fmt.Errorf("%w: missing kid header", ErrInvalidToken)
	}

	// Get the signing key
	jwk, err := m.jwksClient.GetKey(kid)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	// Get the public key based on algorithm
	var key any
	switch jwk.Alg {
	case "RS256", "RS384", "RS512":
		key, err = jwk.GetRSAPublicKey()
	case "ES256", "ES384", "ES512":
		key, err = jwk.GetECDSAPublicKey()
	default:
		return nil, fmt.Errorf("%w: unsupported algorithm %s", ErrInvalidToken, jwk.Alg)
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	// Parse and validate the token
	claims := &customClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(_ *jwt.Token) (any, error) {
		return key, nil
	}, jwt.WithValidMethods([]string{jwk.Alg}))

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	// Validate issuer
	if claims.Issuer != m.issuer {
		m.logger.Debug("issuer mismatch",
			zap.String("expected", m.issuer),
			zap.String("actual", claims.Issuer),
		)
		return nil, ErrInvalidIssuer
	}

	// Validate audience if configured
	if m.audience != "" {
		audValid := false
		for _, aud := range claims.Audience {
			if aud == m.audience {
				audValid = true
				break
			}
		}
		if !audValid {
			return nil, fmt.Errorf("%w: audience mismatch", ErrInvalidToken)
		}
	}

	// Validate subject is not empty
	if claims.Subject == "" {
		return nil, ErrInvalidSubject
	}

	return claims, nil
}

// customClaims extends standard claims with role.
type customClaims struct {
	jwt.RegisteredClaims
	Role  string   `json:"role,omitempty"`
	Roles []string `json:"roles,omitempty"`
}

// extractRoleFromClaims extracts role from custom claims.
func extractRoleFromClaims(claims *customClaims) string {
	// Check single role field first
	if claims.Role != "" {
		return claims.Role
	}
	// Fall back to roles array (take first)
	if len(claims.Roles) > 0 {
		return claims.Roles[0]
	}
	return ""
}

func (m *NeonAuthMiddleware) handleAuthError(w http.ResponseWriter, err error) {
	w.Header().Set("WWW-Authenticate", `Bearer realm="api"`)

	switch {
	case errors.Is(err, ErrMissingToken):
		http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
	case errors.Is(err, ErrTokenExpired):
		http.Error(w, "Unauthorized: token expired", http.StatusUnauthorized)
	case errors.Is(err, ErrInvalidIssuer):
		http.Error(w, "Unauthorized: invalid issuer", http.StatusUnauthorized)
	case errors.Is(err, ErrInvalidSubject):
		http.Error(w, "Unauthorized: invalid subject", http.StatusUnauthorized)
	default:
		m.logger.Debug("auth error", zap.Error(err))
		http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
	}
}

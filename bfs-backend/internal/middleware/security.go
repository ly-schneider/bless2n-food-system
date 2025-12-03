package middleware

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const NonceContextKey contextKey = "csp_nonce"

func WithNonce(ctx context.Context, nonce string) context.Context {
	return context.WithValue(ctx, NonceContextKey, nonce)
}
func GetNonce(ctx context.Context) (string, bool) {
	nonce, ok := ctx.Value(NonceContextKey).(string)
	return nonce, ok
}

type SecurityMiddleware struct {
	enableHSTS     bool
	enableCSP      bool
	trustedOrigins []string
	appEnv         string
}

type SecurityConfig struct {
	EnableHSTS     bool     `json:"enableHSTS"`
	EnableCSP      bool     `json:"enableCSP"`
	TrustedOrigins []string `json:"trustedOrigins"`
	AppEnv         string   `json:"appEnv"`
}

func NewSecurityMiddleware(cfg SecurityConfig) *SecurityMiddleware {
	return &SecurityMiddleware{
		enableHSTS:     cfg.EnableHSTS,
		enableCSP:      cfg.EnableCSP,
		trustedOrigins: cfg.TrustedOrigins,
		appEnv:         strings.ToLower(cfg.AppEnv),
	}
}

func generateNonce() (string, error) {
	b := make([]byte, 16) // 128-bit
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(b), nil
}

func isHTTPS(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	return strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

// IsHTTPS is an exported helper for downstream packages to decide cookie policy.
func IsHTTPS(r *http.Request) bool { return isHTTPS(r) }

func (s *SecurityMiddleware) isOriginAllowed(origin string) bool {
	if origin == "" {
		return false
	}
	for _, o := range s.trustedOrigins {
		if o == origin {
			return true
		}
	}
	return false
}

func (s *SecurityMiddleware) SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		w.Header().Set("X-Frame-Options", "DENY")

		// HSTS only when enabled AND over HTTPS (avoid on dev/http)
		if s.enableHSTS && s.appEnv != "dev" && isHTTPS(r) {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		// CSP: Report-Only in dev, Enforcing in staging/production
		if s.enableCSP {
			if nonce, err := generateNonce(); err == nil {
				headerName := "Content-Security-Policy"
				if s.appEnv == "dev" {
					headerName = "Content-Security-Policy-Report-Only"
				}

				var csp strings.Builder

				// Relaxed CSP for Swagger UI (requires inline styles/scripts)
				if strings.HasPrefix(r.URL.Path, "/swagger/") {
					fmt.Fprintf(&csp,
						"default-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:;",
					)
				} else {
					// Base policy (nonce-based JS)
					fmt.Fprintf(&csp,
						"default-src 'self'; object-src 'none'; base-uri 'self'; frame-ancestors 'none'; script-src 'self' 'nonce-%s';",
						nonce,
					)

					// Dev: allow HMR / API calls to your dev servers via connect-src
					if s.appEnv == "dev" {
						connect := make([]string, 0, 1+len(s.trustedOrigins)*2)
						connect = append(connect, "'self'")
						for _, o := range s.trustedOrigins {
							connect = append(connect, o)
							// Derive ws(s) URL from http(s) origin for dev tooling (HMR)
							if u, err := url.Parse(o); err == nil && u.Host != "" {
								switch u.Scheme {
								case "http":
									connect = append(connect, "ws://"+u.Host)
								case "https":
									connect = append(connect, "wss://"+u.Host)
								}
							}
						}
						fmt.Fprintf(&csp, " connect-src %s;", strings.Join(connect, " "))
					}
				}

				w.Header().Set(headerName, csp.String())
				// stash nonce for templates
				r = r.WithContext(WithNonce(r.Context(), nonce))
			}
		}

		next.ServeHTTP(w, r)
	})
}

// No-store on sensitive endpoints
func (s *SecurityMiddleware) CacheControlForSensitive(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.isSensitiveEndpoint(r.URL.Path) {
			w.Header().Set("Cache-Control", "no-store, max-age=0")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		}
		next.ServeHTTP(w, r)
	})
}
func (s *SecurityMiddleware) isSensitiveEndpoint(path string) bool {
	for _, p := range []string{"/auth/", "/profile", "/admin/"} {
		if strings.Contains(path, p) {
			return true
		}
	}
	return false
}

// CORS: exact-origin allowlist; credentials-safe; dynamic preflight
func (s *SecurityMiddleware) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowed := s.isOriginAllowed(origin)

		// Methods/Headers
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		// Echo requested headers for preflight if provided; else default
		if reqHdr := r.Header.Get("Access-Control-Request-Headers"); reqHdr != "" {
			w.Header().Set("Access-Control-Allow-Headers", reqHdr)
		} else {
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type, X-CSRF-Token, X-Requested-With")
		}
		// Cache preflights (browsers cap this; 7200s is a common ceiling)
		w.Header().Set("Access-Control-Max-Age", "7200")

		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin) // reflect
			w.Header().Set("Vary", "Origin")                      // cache safety for dynamic ACAO
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Preflight short-circuit
		if r.Method == http.MethodOptions {
			if !allowed {
				http.Error(w, "CORS origin not allowed", http.StatusForbidden)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type SecureCookieOptions struct {
	Name     string
	Value    string
	Path     string
	Domain   string
	MaxAge   int
	HttpOnly bool
	Secure   bool
	SameSite http.SameSite
}

func SetSecureCookie(w http.ResponseWriter, opts SecureCookieOptions) {
	http.SetCookie(w, &http.Cookie{
		Name:     opts.Name,
		Value:    opts.Value,
		Path:     opts.Path,
		Domain:   opts.Domain,
		MaxAge:   opts.MaxAge,
		HttpOnly: opts.HttpOnly,
		Secure:   opts.Secure,
		SameSite: opts.SameSite,
	})
}

// __Host- cookies must be Secure, Path=/, and have NO Domain attribute (HTTPS only)
func SetSecureAuthCookie(w http.ResponseWriter, name, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     "__Host-" + name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode, // Lax is ergonomic & CSRF-aware for same-site
	})
}
func ClearSecureAuthCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "__Host-" + name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

// SetAuthCookie sets a refresh/auth cookie that works in both HTTPS and local HTTP dev.
// - On HTTPS: uses __Host- prefix and Secure=true
// - On HTTP (dev): no prefix and Secure=false (still HttpOnly + SameSite=Lax)
func SetAuthCookie(w http.ResponseWriter, r *http.Request, name, value string, maxAge int) {
	if isHTTPS(r) {
		SetSecureAuthCookie(w, name, value, maxAge)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearAuthCookie clears the refresh/auth cookie for both HTTPS and local HTTP dev.
func ClearAuthCookie(w http.ResponseWriter, r *http.Request, name string) {
	if isHTTPS(r) {
		ClearSecureAuthCookie(w, name)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}

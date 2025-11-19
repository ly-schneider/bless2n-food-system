package middleware

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"backend/internal/origin"
)

type OriginConfig struct {
	DefaultFrontendOrigin string
	DefaultBackendOrigin  string
	AllowedFrontendHosts  []string
	AllowedBackendHosts   []string
}

type OriginMiddleware struct {
	defaultFrontend     string
	defaultBackend      string
	allowedFrontendHosts map[string]struct{}
	allowedBackendHosts  map[string]struct{}
}

func NewOriginMiddleware(cfg OriginConfig) *OriginMiddleware {
	allowedFrontend := make(map[string]struct{})
	for _, host := range cfg.AllowedFrontendHosts {
		if h := canonicalHost(host); h != "" {
			allowedFrontend[h] = struct{}{}
		}
	}
	allowedBackend := make(map[string]struct{})
	for _, host := range cfg.AllowedBackendHosts {
		if h := canonicalHost(host); h != "" {
			allowedBackend[h] = struct{}{}
		}
	}
	return &OriginMiddleware{
		defaultFrontend:     strings.TrimRight(cfg.DefaultFrontendOrigin, "/"),
		defaultBackend:      strings.TrimRight(cfg.DefaultBackendOrigin, "/"),
		allowedFrontendHosts: allowedFrontend,
		allowedBackendHosts:  allowedBackend,
	}
}

func (m *OriginMiddleware) Inject(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info := origin.Info{
			Frontend: m.resolveFrontend(r),
			Backend:  m.resolveBackend(r),
		}
		next.ServeHTTP(w, r.WithContext(origin.WithContext(r.Context(), info)))
	})
}

func (m *OriginMiddleware) resolveFrontend(r *http.Request) string {
	host := canonicalHost(r.Header.Get("X-Frontend-Host"))
	proto := strings.ToLower(strings.TrimSpace(r.Header.Get("X-Frontend-Proto")))
	if proto == "" {
		proto = "https"
	}

	if host != "" {
		if len(m.allowedFrontendHosts) == 0 {
			return m.defaultFrontend
		}
		if _, ok := m.allowedFrontendHosts[host]; ok {
			return fmt.Sprintf("%s://%s", proto, host)
		}
	}
	return m.defaultFrontend
}

func (m *OriginMiddleware) resolveBackend(r *http.Request) string {
	proto := strings.ToLower(strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")))
	if proto == "" {
		if r.TLS != nil {
			proto = "https"
		} else {
			proto = "http"
		}
	}
	
	// Validate X-Forwarded-Host to prevent header forgery attacks
	forwardedHost := canonicalHost(r.Header.Get("X-Forwarded-Host"))
	var host string
	
	if forwardedHost != "" {
		// If AllowedBackendHosts is configured, validate the header
		if len(m.allowedBackendHosts) > 0 {
			if _, ok := m.allowedBackendHosts[forwardedHost]; ok {
				host = forwardedHost
			}
			// If not in allowlist, fall through to use default
		} else {
			// No allowlist configured - trust the header (backward compatible)
			host = forwardedHost
		}
	}
	
	// Fall back to r.Host if X-Forwarded-Host is not trusted or not present
	if host == "" {
		host = canonicalHost(r.Host)
	}
	
	if host == "" {
		return m.defaultBackend
	}
	return fmt.Sprintf("%s://%s", proto, host)
}

func canonicalHost(raw string) string {
	if raw == "" {
		return ""
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if strings.HasPrefix(raw, "http://") || strings.HasPrefix(raw, "https://") {
		if u, err := url.Parse(raw); err == nil {
			raw = u.Host
		}
	}
	parts := strings.Split(raw, "/")
	raw = parts[0]
	raw = strings.TrimSpace(raw)
	return strings.ToLower(raw)
}

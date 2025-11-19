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
}

type OriginMiddleware struct {
	defaultFrontend string
	defaultBackend  string
	allowedHosts    map[string]struct{}
}

func NewOriginMiddleware(cfg OriginConfig) *OriginMiddleware {
	allowed := make(map[string]struct{})
	for _, host := range cfg.AllowedFrontendHosts {
		if h := canonicalHost(host); h != "" {
			allowed[h] = struct{}{}
		}
	}
	return &OriginMiddleware{
		defaultFrontend: strings.TrimRight(cfg.DefaultFrontendOrigin, "/"),
		defaultBackend:  strings.TrimRight(cfg.DefaultBackendOrigin, "/"),
		allowedHosts:    allowed,
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
		if len(m.allowedHosts) == 0 {
			return m.defaultFrontend
		}
		if _, ok := m.allowedHosts[host]; ok {
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
	host := canonicalHost(r.Header.Get("X-Forwarded-Host"))
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

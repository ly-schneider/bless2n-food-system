package middleware

import (
	"backend/internal/utils"
	"net/http"
)

// CSRFMiddleware enforces double-submit cookie for state-changing requests.
// It allows GET/HEAD/OPTIONS, and requires X-CSRF header to match the csrf cookie for others.
type CSRFMiddleware struct{}

func NewCSRFMiddleware() *CSRFMiddleware { return &CSRFMiddleware{} }

func (m *CSRFMiddleware) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			next.ServeHTTP(w, r)
			return
		}
		header := r.Header.Get(utils.CSRFHeaderName)
		if header == "" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		// Accept either secure or dev cookie name
		c, _ := r.Cookie("__Host-" + utils.CSRFCookieName)
		if c == nil {
			c, _ = r.Cookie(utils.CSRFCookieName)
		}
		if c == nil || c.Value == "" || c.Value != header {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

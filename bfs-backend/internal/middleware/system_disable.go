package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"backend/internal/response"
	"backend/internal/service"
)

type AdminChecker interface {
	IsAdmin(r *http.Request) bool
}

type SystemDisableMiddleware struct {
	settings     service.SettingsService
	adminChecker AdminChecker

	mu        sync.RWMutex
	cached    bool
	fetchedAt time.Time
}

const systemCacheTTL = 5 * time.Second

func NewSystemDisableMiddleware(settings service.SettingsService, adminChecker AdminChecker) *SystemDisableMiddleware {
	return &SystemDisableMiddleware{
		settings:     settings,
		adminChecker: adminChecker,
		cached:       true,
	}
}

func (m *SystemDisableMiddleware) isEnabled(ctx context.Context) bool {
	m.mu.RLock()
	if !m.fetchedAt.IsZero() && time.Since(m.fetchedAt) < systemCacheTTL {
		v := m.cached
		m.mu.RUnlock()
		return v
	}
	m.mu.RUnlock()

	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.fetchedAt.IsZero() && time.Since(m.fetchedAt) < systemCacheTTL {
		return m.cached
	}

	enabled, err := m.settings.IsSystemEnabled(ctx)
	if err != nil {
		return true
	}
	m.cached = enabled
	m.fetchedAt = time.Now()
	return m.cached
}

func (m *SystemDisableMiddleware) RequireEnabled(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.isEnabled(r.Context()) {
			if m.adminChecker.IsAdmin(r) {
				next.ServeHTTP(w, r)
				return
			}
			response.WriteProblem(w, response.NewProblem(
				http.StatusServiceUnavailable,
				"Service Unavailable",
				"The food system is currently closed.",
			))
			return
		}
		next.ServeHTTP(w, r)
	})
}

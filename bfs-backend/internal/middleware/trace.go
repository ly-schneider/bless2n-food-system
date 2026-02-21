package middleware

import (
	"backend/internal/trace"
	"fmt"
	"net/http"

	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	chiMw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (w *statusRecorder) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func TraceRoute(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		if span := sentry.SpanFromContext(r.Context()); span != nil {
			if rctx := chi.RouteContext(r.Context()); rctx != nil {
				if pattern := rctx.RoutePattern(); pattern != "" {
					span.Name = r.Method + " " + pattern
					span.SetTag("http.route", pattern)
				}
			}
		}
	})
}

func LogServerErrors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		if rec.status >= 500 {
			trace.Tag(r.Context(), "error", "true")
			trace.Tag(r.Context(), "http.status_code", fmt.Sprintf("%d", rec.status))

			reqID := chiMw.GetReqID(r.Context())
			logger := zap.L()
			logger.Error("http 5xx",
				zap.Int("status", rec.status),
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_ip", r.RemoteAddr),
				zap.String("request_id", reqID),
			)

			if hub := sentry.GetHubFromContext(r.Context()); hub != nil {
				hub.WithScope(func(scope *sentry.Scope) {
					scope.SetTag("status", fmt.Sprintf("%d", rec.status))
					scope.SetTag("method", r.Method)
					scope.SetTag("path", r.URL.Path)
					hub.CaptureMessage(fmt.Sprintf("HTTP %d: %s %s", rec.status, r.Method, r.URL.Path))
				})
			}

			SentryLogger(r.Context()).Error().
				Int("status", rec.status).
				String("method", r.Method).
				String("path", r.URL.Path).
				String("request_id", reqID).
				Emitf("HTTP %d: %s %s", rec.status, r.Method, r.URL.Path)
		}
	})
}

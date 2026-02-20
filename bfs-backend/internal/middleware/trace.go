package middleware

import (
	"fmt"
	"net/http"

	"github.com/getsentry/sentry-go"
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

func LogServerErrors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		if rec.status >= 500 {
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
		}
	})
}

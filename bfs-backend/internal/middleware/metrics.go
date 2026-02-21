package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/getsentry/sentry-go/attribute"
	"github.com/go-chi/chi/v5"
)

func HTTPMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rec, r)

		route := chi.RouteContext(r.Context()).RoutePattern()
		if route == "" {
			route = "unknown"
		}
		attrs := sentry.WithAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.route", route),
			attribute.String("http.status_code", fmt.Sprintf("%d", rec.status)),
		)

		meter := sentry.NewMeter(r.Context())
		meter.Count("http.request.count", 1, attrs)
		meter.Distribution("http.request.duration", float64(time.Since(start).Milliseconds()), sentry.WithUnit(sentry.UnitMillisecond), attrs)
	})
}

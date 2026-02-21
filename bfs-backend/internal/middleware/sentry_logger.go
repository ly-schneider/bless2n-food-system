package middleware

import (
	"context"
	"net/http"

	"github.com/getsentry/sentry-go"
)

type sentryLoggerKey struct{}

func SentryLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := sentry.NewLogger(r.Context())
		ctx := context.WithValue(r.Context(), sentryLoggerKey{}, logger)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func SentryLogger(ctx context.Context) sentry.Logger {
	if l, ok := ctx.Value(sentryLoggerKey{}).(sentry.Logger); ok {
		return l
	}
	return sentry.NewLogger(context.Background())
}

package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type ctxKey struct{ name string }

var requestIDKey = &ctxKey{"request-id"}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = uuid.NewString()
		}

		// Make the ID available to handlers & middleware further down the chain.
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		r = r.WithContext(ctx)

		// Echo the header so clients / proxies see the same value.
		w.Header().Set("X-Request-ID", id)

		next.ServeHTTP(w, r)
	})
}

// RequestIDFromContext returns the ID previously stored by the middleware.
// It returns an empty string when absent.
func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDKey).(string); ok {
		return v
	}
	return ""
}

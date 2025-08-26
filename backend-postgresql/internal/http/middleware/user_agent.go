package middleware

import (
	"context"
	"net/http"
)

type uaCtxKey struct{ name string }

var UserAgentKey = &uaCtxKey{"user-agent"}

func UserAgent(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), UserAgentKey, r.UserAgent())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ExtractUAFromContext(ctx context.Context) *string {
	v := ctx.Value(UserAgentKey)
	if ua, ok := v.(string); ok {
		return &ua
	}
	return nil
}

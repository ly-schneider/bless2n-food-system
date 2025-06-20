package middleware

import (
	"context"
	"net"
	"net/http"
)

type ipCtxKey struct{ name string }

var PublicIPKey = &ipCtxKey{"public-ip"}

func PublicIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ipStr, _, _ := net.SplitHostPort(r.RemoteAddr)
		ip := net.ParseIP(ipStr)
		ctx := context.WithValue(r.Context(), PublicIPKey, ip)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ExtractIPFromContext(ctx context.Context) *net.IP {
	v := ctx.Value(PublicIPKey)
	if ip, ok := v.(net.IP); ok {
		return &ip
	}
	return nil
}

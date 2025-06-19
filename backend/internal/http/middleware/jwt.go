package middleware

import (
	"net/http"
	"strings"

	"backend/internal/config"
	"backend/internal/db"
	"backend/internal/logger"

	"github.com/golang-jwt/jwt/v5"
)

// JWT returns a chi-compatible middleware that injects the user-ID from a
// verified JWT into the request context so that pgx's BeforeAcquire hook can
// push it into the session variable app.current_user_id.
func JWT(next http.Handler) http.Handler {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Load().App.JWTSecretKey), nil
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")

		if !strings.HasPrefix(auth, "Bearer ") {
			// anonymous request – let RLS filter everything
			next.ServeHTTP(w, r)
			return
		}

		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		claims := jwt.MapClaims{}

		_, err := jwt.ParseWithClaims(tokenStr, claims, keyFunc,
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
		if err != nil {
			logger.Warn("JWT parse / verify failed", "err", err)
			// continue as anonymous; don't leak details
			next.ServeHTTP(w, r)
			return
		}

		uid, _ := claims["sub"].(string) // empty string if claim missing
		ctx := db.InjectUser(r.Context(), uid)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

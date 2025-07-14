package middleware

import (
	"context"
	"net/http"
	"strings"

	"backend/internal/apperrors"
	"backend/internal/config"
	"backend/internal/db"
	"backend/internal/http/respond"
	"backend/internal/logger"

	"github.com/golang-jwt/jwt/v5"
)

type uidCtxKey struct{ name string }

var UserIDKey = &uidCtxKey{"user-id"}

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
			if strings.Contains(err.Error(), "token is expired") {
				logger.Warn("JWT expired", "err", err)
				respond.NewWriter(w).WriteError(apperrors.Unauthorized("JWT token is expired"))
				return
			}

			logger.Warn("JWT parse / verify failed", "err", err)
			respond.NewWriter(w).WriteError(apperrors.Unauthorized("JWT token is invalid or malformed"))
			return
		}

		uid := claims["sub"]
		if uid == nil {
			logger.Warn("JWT missing user ID", "claims", claims)
			respond.NewWriter(w).WriteError(apperrors.Unauthorized("JWT token is missing user ID"))
			return
		}

		uidStr, ok := uid.(string)
		if !ok || uidStr == "" {
			logger.Warn("JWT user ID is not a string or is empty", "claims", claims)
			respond.NewWriter(w).WriteError(apperrors.Unauthorized("JWT token user ID is invalid"))
			return
		}

		ctx := db.InjectUser(r.Context(), uidStr)

		// Make uid available in the context for other service uses
		ctx = context.WithValue(ctx, UserIDKey, uidStr)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func ExtractUserIDFromContext(ctx context.Context) *string {
	v := ctx.Value(UserIDKey)
	if uid, ok := v.(string); ok {
		return &uid
	}
	return nil
}

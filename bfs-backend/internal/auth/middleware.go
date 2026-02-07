package auth

import (
	"errors"
	"net/http"
	"strings"
)

var (
	ErrMissingToken   = errors.New("missing authorization token")
	ErrInvalidToken   = errors.New("invalid token")
	ErrTokenExpired   = errors.New("token expired")
	ErrInvalidIssuer  = errors.New("invalid token issuer")
	ErrInvalidSubject = errors.New("invalid token subject")
)

func extractToken(r *http.Request) (string, error) {
	return ExtractBearerToken(r)
}

func ExtractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", ErrMissingToken
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", ErrInvalidToken
	}

	return parts[1], nil
}

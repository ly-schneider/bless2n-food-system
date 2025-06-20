package domain

import "errors"

var (
	// User‑related
	ErrUserNotFound  = errors.New("user not found")
	ErrUserDisabled  = errors.New("user is disabled")
	ErrInvalidUserID = errors.New("invalid user id")

	// Refresh‑token‑related
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrRefreshTokenRevoked  = errors.New("refresh token is revoked")
	ErrRefreshTokenExpired  = errors.New("refresh token has expired")

	// Generic helpers
	ErrParseBody                = errors.New("unable to parse request body")
	ErrNotFound                 = errors.New("resource not found")
	ErrAlreadyExist             = errors.New("resource already exists")
	ErrInvalidBody              = errors.New("invalid request body")
	ErrInvalidBodyMissingFields = errors.New("invalid request body: missing required fields")
)

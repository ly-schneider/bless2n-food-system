package domain

import "errors"

var (
	// User errors
	ErrUserNotFound    = errors.New("user not found")
	ErrEmailTaken      = errors.New("email already taken")
	ErrInvalidPassword = errors.New("invalid password")

	// Token errors
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrRefreshTokenExpired  = errors.New("refresh token expired")
	ErrRefreshTokenRevoked  = errors.New("refresh token revoked")

	// OTP errors
	ErrOTPTokenNotFound = errors.New("OTP token not found")
	ErrOTPTokenExpired  = errors.New("OTP token expired")
	ErrOTPTokenUsed     = errors.New("OTP token already used")
	ErrInvalidOTP       = errors.New("invalid OTP")

	// Validation errors
	ErrInvalidBodyMissingFields = errors.New("missing required fields")
	ErrValidationFailed         = errors.New("validation failed")

	// Verification errors  
	ErrVerificationTokenNotFound = errors.New("verification token not found")
	ErrVerificationTokenExpired  = errors.New("verification token expired")
)

// Role mapping for backward compatibility
var Roles = map[string]struct {
	ID   int
	Name string
}{
	"admin":    {ID: 1, Name: "admin"},
	"user":     {ID: 2, Name: "user"},
	"customer": {ID: 2, Name: "customer"},
}
package service

import (
	"context"

	"backend/internal/domain"
)

// AuthService defines the authentication service interface
type AuthService interface {
	RegisterCustomer(ctx context.Context, req RegisterCustomerRequest) (*RegisterCustomerResponse, error)
	VerifyOTP(ctx context.Context, req VerifyOTPRequest) (*VerifyOTPResponse, error)
	ResendOTP(ctx context.Context, req ResendOTPRequest) (*ResendOTPResponse, error)
	RequestLoginOTP(ctx context.Context, req RequestLoginOTPRequest) (*RequestLoginOTPResponse, error)
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
	RefreshToken(ctx context.Context, req RefreshTokenRequest) (*RefreshTokenResponse, error)
	Logout(ctx context.Context, req LogoutRequest) (*LogoutResponse, error)
}

// Request/Response types for Customer Registration
type RegisterCustomerRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type RegisterCustomerResponse struct {
	Message string `json:"message"`
	UserID  string `json:"user_id"`
}

// Request/Response types for OTP Verification
type VerifyOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp" validate:"required,len=6"`
}

type VerifyOTPResponse struct {
	Message string       `json:"message"`
	User    *domain.User `json:"user,omitempty"`
	Token   string       `json:"token,omitempty"`
}

// Request/Response types for OTP Resend
type ResendOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResendOTPResponse struct {
	Message string `json:"message"`
}

// Request/Response types for Login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	OTP      string `json:"otp" validate:"required,len=6"`
	ClientID string `json:"client_id" validate:"required"`
}

type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int64        `json:"expires_in"`
	User         *domain.User `json:"user"`
}

// Request/Response types for Token Refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
	ClientID     string `json:"client_id" validate:"required"`
}

type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// Request/Response types for Logout
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type LogoutResponse struct {
	Message string `json:"message"`
}

// Request/Response types for Login OTP Request
type RequestLoginOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type RequestLoginOTPResponse struct {
	Message string `json:"message"`
}

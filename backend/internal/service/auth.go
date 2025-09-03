package service

import (
	"context"

	"backend/internal/domain"
)

// AuthService defines the authentication service interface
type AuthService interface {
	RegisterCustomer(ctx context.Context, req RegisterCustomerRequest) (*RegisterCustomerResponse, error)
	RequestOTP(ctx context.Context, req RequestOTPRequest) (*RequestOTPResponse, error)
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
	Email    string `json:"email" validate:"required,email"`
	OTP      string `json:"otp" validate:"required,len=6"`
	ClientID string `json:"client_id" validate:"required"`
}

type VerifyOTPResponse struct {
	Message      string       `json:"message"`
	User         *domain.User `json:"user,omitempty"`
	AccessToken  string       `json:"access_token,omitempty"`
	RefreshToken string       `json:"refresh_token,omitempty"`
	TokenType    string       `json:"token_type,omitempty"`
	ExpiresIn    int64        `json:"expires_in,omitempty"`
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
type RequestOTPRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type RequestOTPResponse struct {
	Message string `json:"message"`
}

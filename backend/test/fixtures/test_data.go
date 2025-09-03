package fixtures

import (
	"time"

	"backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestUsers contains predefined test users for e2e testing
var TestUsers = struct {
	CustomerUser  *domain.User
	AdminUser     *domain.User
}{
	CustomerUser: &domain.User{
		ID:         primitive.NewObjectID(),
		Email:      "customer@test.com",
		Role:       domain.UserRoleCustomer,
		IsVerified: true,
		IsDisabled: false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	},
	AdminUser: &domain.User{
		ID:         primitive.NewObjectID(),
		Email:      "admin@test.com",
		FirstName:  "Test",
		LastName:   "Admin",
		Role:       domain.UserRoleAdmin,
		IsVerified: true,
		IsDisabled: false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	},
}

// ValidRegistrationRequest contains test data for registration
var ValidRegistrationRequest = map[string]any{
	"email": "newcustomer@test.com",
}

// ValidOTPRequest contains test data for OTP verification
var ValidOTPRequest = map[string]any{
	"email":     "customer@test.com",
	"otp":       "123456",
	"client_id": "test-client-123",
}

// ValidLoginRequest contains test data for login
var ValidLoginRequest = map[string]any{
	"email":     "customer@test.com", 
	"otp":       "123456",
	"client_id": "test-client-123",
}

// ValidRefreshTokenRequest contains test data for token refresh
var ValidRefreshTokenRequest = map[string]any{
	"refresh_token": "test-refresh-token",
	"client_id":     "test-client-123",
}

// InvalidRequests contains various invalid request payloads for testing
var InvalidRequests = struct {
	InvalidEmail     map[string]any
	MissingEmail     map[string]any
	InvalidOTP       map[string]any
	MissingOTP       map[string]any
	MissingClientID  map[string]any
}{
	InvalidEmail: map[string]any{
		"email": "invalid-email",
	},
	MissingEmail: map[string]any{},
	InvalidOTP: map[string]any{
		"email":     "customer@test.com",
		"otp":       "12345", // Too short
		"client_id": "test-client-123",
	},
	MissingOTP: map[string]any{
		"email":     "customer@test.com",
		"client_id": "test-client-123",
	},
	MissingClientID: map[string]any{
		"email": "customer@test.com",
		"otp":   "123456",
	},
}
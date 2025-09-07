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

// TestStations contains predefined test stations for e2e testing
var TestStations = struct {
	PendingStation  *domain.Station
	ApprovedStation *domain.Station
	RejectedStation *domain.Station
}{
	PendingStation: &domain.Station{
		ID:        primitive.NewObjectID(),
		Name:      "Test Pending Station",
		Status:    domain.StationStatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	ApprovedStation: &domain.Station{
		ID:         primitive.NewObjectID(),
		Name:       "Test Approved Station",
		Status:     domain.StationStatusApproved,
		ApprovedBy: &TestUsers.AdminUser.ID,
		ApprovedAt: func() *time.Time { t := time.Now(); return &t }(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	},
	RejectedStation: &domain.Station{
		ID:              primitive.NewObjectID(),
		Name:            "Test Rejected Station",
		Status:          domain.StationStatusRejected,
		RejectedBy:      &TestUsers.AdminUser.ID,
		RejectedAt:      func() *time.Time { t := time.Now(); return &t }(),
		RejectionReason: func() *string { r := "Not meeting requirements"; return &r }(),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	},
}

// ValidStationRequest contains test data for station creation
var ValidStationRequest = map[string]any{
	"name": "New Test Station",
}

// ValidStationStatusApprovalRequest contains test data for station approval
var ValidStationStatusApprovalRequest = map[string]any{
	"approve": true,
}

// ValidStationStatusRejectionRequest contains test data for station rejection
var ValidStationStatusRejectionRequest = map[string]any{
	"approve": false,
	"reason":  "Does not meet requirements",
}

// InvalidRequests contains various invalid request payloads for testing
var InvalidRequests = struct {
	InvalidEmail        map[string]any
	MissingEmail        map[string]any
	InvalidOTP          map[string]any
	MissingOTP          map[string]any
	MissingClientID     map[string]any
	InvalidStationName  map[string]any
	MissingStationName  map[string]any
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
	InvalidStationName: map[string]any{
		"name": "", // Empty name
	},
	MissingStationName: map[string]any{},
}
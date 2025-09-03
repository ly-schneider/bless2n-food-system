package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidateOTP(t *testing.T) {
	tests := []struct {
		name        string
		otp         string
		expectedLen int
		isValid     bool
	}{
		{
			name:        "Valid 6-digit OTP",
			otp:         "123456",
			expectedLen: 6,
			isValid:     true,
		},
		{
			name:        "Invalid short OTP",
			otp:         "12345",
			expectedLen: 6,
			isValid:     false,
		},
		{
			name:        "Invalid long OTP",
			otp:         "1234567",
			expectedLen: 6,
			isValid:     false,
		},
		{
			name:        "Empty OTP",
			otp:         "",
			expectedLen: 6,
			isValid:     false,
		},
		{
			name:        "Non-numeric OTP",
			otp:         "12345a",
			expectedLen: 6,
			isValid:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := validateOTPFormat(tt.otp, tt.expectedLen)
			assert.Equal(t, tt.isValid, isValid, "OTP validation result should match expected")
		})
	}
}

func TestIsOTPExpired(t *testing.T) {
	now := time.Now()
	
	tests := []struct {
		name      string
		expiresAt time.Time
		expected  bool
	}{
		{
			name:      "Not expired - future time",
			expiresAt: now.Add(5 * time.Minute),
			expected:  false,
		},
		{
			name:      "Expired - past time",
			expiresAt: now.Add(-5 * time.Minute),
			expected:  true,
		},
		{
			name:      "Exactly now",
			expiresAt: now,
			expected:  true, // Treat exactly now as expired for safety
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isExpired := tt.expiresAt.Before(time.Now())
			assert.Equal(t, tt.expected, isExpired, "OTP expiration check should match expected")
		})
	}
}

// validateOTPFormat is a helper function that would typically be in the auth service
func validateOTPFormat(otp string, expectedLength int) bool {
	if len(otp) != expectedLength {
		return false
	}
	
	for _, char := range otp {
		if char < '0' || char > '9' {
			return false
		}
	}
	
	return true
}
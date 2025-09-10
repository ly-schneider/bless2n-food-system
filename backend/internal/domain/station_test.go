package domain

import (
    "testing"

    v10 "github.com/go-playground/validator/v10"
    "github.com/stretchr/testify/assert"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

func TestStation_Validation(t *testing.T) {
    validate := v10.New()
	
	tests := []struct {
		name        string
		station     Station
		expectError bool
		errorField  string
	}{
		{
			name: "valid pending station",
			station: Station{
				ID:     primitive.NewObjectID(),
				Name:   "Test Station",
				Status: StationStatusPending,
			},
			expectError: false,
		},
		{
			name: "valid approved station",
			station: Station{
				ID:     primitive.NewObjectID(),
				Name:   "Test Station",
				Status: StationStatusApproved,
			},
			expectError: false,
		},
		{
			name: "valid rejected station",
			station: Station{
				ID:     primitive.NewObjectID(),
				Name:   "Test Station",
				Status: StationStatusRejected,
			},
			expectError: false,
		},
		{
			name: "missing name",
			station: Station{
				ID:     primitive.NewObjectID(),
				Status: StationStatusPending,
			},
			expectError: true,
			errorField:  "Name",
		},
		{
			name: "missing status",
			station: Station{
				ID:   primitive.NewObjectID(),
				Name: "Test Station",
			},
			expectError: true,
			errorField:  "Status",
		},
		{
			name: "invalid status",
			station: Station{
				ID:     primitive.NewObjectID(),
				Name:   "Test Station",
				Status: StationStatus("invalid"),
			},
			expectError: true,
			errorField:  "Status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
        err := validate.Struct(tt.station)
			
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorField != "" {
                validationErrors, ok := err.(v10.ValidationErrors)
					assert.True(t, ok)
					found := false
					for _, validationError := range validationErrors {
						if validationError.Field() == tt.errorField {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected validation error for field %s", tt.errorField)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStationRequest_Validation(t *testing.T) {
    validate := v10.New()
	
	tests := []struct {
		name        string
		request     StationRequest
		expectError bool
		errorField  string
	}{
		{
			name: "valid station request",
			request: StationRequest{
				Name: "New Station",
			},
			expectError: false,
		},
		{
			name: "missing name",
			request: StationRequest{},
			expectError: true,
			errorField:  "Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
        err := validate.Struct(tt.request)
			
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorField != "" {
                validationErrors, ok := err.(v10.ValidationErrors)
					assert.True(t, ok)
					found := false
					for _, validationError := range validationErrors {
						if validationError.Field() == tt.errorField {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected validation error for field %s", tt.errorField)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStationStatusRequest_Validation(t *testing.T) {
    validate := v10.New()
	
	tests := []struct {
		name        string
		request     StationStatusRequest
		expectError bool
		errorField  string
	}{
		{
			name: "valid approve request",
			request: StationStatusRequest{
				Approve: boolPtr(true),
			},
			expectError: false,
		},
		{
			name: "valid reject request with reason",
			request: StationStatusRequest{
				Approve: boolPtr(false),
				Reason:  stringPtr("Not suitable"),
			},
			expectError: false,
		},
		{
			name: "missing approve field",
			request: StationStatusRequest{
				Reason: stringPtr("Some reason"),
			},
			expectError: true,
			errorField:  "Approve",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
        err := validate.Struct(tt.request)
			
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorField != "" {
                validationErrors, ok := err.(v10.ValidationErrors)
					assert.True(t, ok)
					found := false
					for _, validationError := range validationErrors {
						if validationError.Field() == tt.errorField {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected validation error for field %s", tt.errorField)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStationStatus_Constants(t *testing.T) {
	assert.Equal(t, StationStatus("pending"), StationStatusPending)
	assert.Equal(t, StationStatus("approved"), StationStatusApproved)
	assert.Equal(t, StationStatus("rejected"), StationStatusRejected)
}

func boolPtr(b bool) *bool {
	return &b
}

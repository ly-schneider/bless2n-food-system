package domain

import (
    "testing"

    v10 "github.com/go-playground/validator/v10"
    "github.com/stretchr/testify/assert"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

func TestUser_Validation(t *testing.T) {
    validate := v10.New()
	
	tests := []struct {
		name        string
		user        User
		expectError bool
		errorField  string
	}{
		{
			name: "valid customer user",
			user: User{
				ID:         primitive.NewObjectID(),
				Email:      "customer@example.com",
				Role:       UserRoleCustomer,
				IsVerified: false,
				IsDisabled: false,
			},
			expectError: false,
		},
		{
			name: "valid admin user",
			user: User{
				ID:        primitive.NewObjectID(),
				Email:     "admin@example.com",
				FirstName: "John",
				LastName:  "Doe",
				Role:      UserRoleAdmin,
				IsVerified: true,
				IsDisabled: false,
			},
			expectError: false,
		},
		{
			name: "missing email",
			user: User{
				ID:   primitive.NewObjectID(),
				Role: UserRoleCustomer,
			},
			expectError: true,
			errorField:  "Email",
		},
		{
			name: "invalid email format",
			user: User{
				ID:    primitive.NewObjectID(),
				Email: "invalid-email",
				Role:  UserRoleCustomer,
			},
			expectError: true,
			errorField:  "Email",
		},
		{
			name: "missing role",
			user: User{
				ID:    primitive.NewObjectID(),
				Email: "user@example.com",
			},
			expectError: true,
			errorField:  "Role",
		},
		{
			name: "invalid role",
			user: User{
				ID:    primitive.NewObjectID(),
				Email: "user@example.com",
				Role:  UserRole("invalid"),
			},
			expectError: true,
			errorField:  "Role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
        err := validate.Struct(tt.user)
			
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

func TestUserRole_Constants(t *testing.T) {
	assert.Equal(t, UserRole("admin"), UserRoleAdmin)
	assert.Equal(t, UserRole("customer"), UserRoleCustomer)
}

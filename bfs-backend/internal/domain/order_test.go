package domain

import (
    "testing"

    v10 "github.com/go-playground/validator/v10"
    "github.com/stretchr/testify/assert"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

func TestOrder_Validation(t *testing.T) {
    validate := v10.New()
	customerID := primitive.NewObjectID()
	
	tests := []struct {
		name        string
		order       Order
		expectError bool
		errorField  string
	}{
		{
			name: "valid order with customer",
			order: Order{
				ID:         primitive.NewObjectID(),
				CustomerID: &customerID,
				Total:      29.99,
				Status:     OrderStatusPending,
			},
			expectError: false,
		},
		{
			name: "valid order with contact email only",
			order: Order{
				ID:           primitive.NewObjectID(),
				ContactEmail: stringPtr("contact@example.com"),
				Total:        19.99,
				Status:       OrderStatusPaid,
			},
			expectError: false,
		},
        {
            name: "zero total is invalid due to 'required'",
            order: Order{
                ID:         primitive.NewObjectID(),
                CustomerID: &customerID,
                Total:      0,
                Status:     OrderStatusPending,
            },
            expectError: true,
            errorField:  "Total",
        },
		{
			name: "missing total",
			order: Order{
				ID:         primitive.NewObjectID(),
				CustomerID: &customerID,
				Status:     OrderStatusPending,
			},
			expectError: true,
			errorField:  "Total",
		},
		{
			name: "negative total",
			order: Order{
				ID:         primitive.NewObjectID(),
				CustomerID: &customerID,
				Total:      -10.99,
				Status:     OrderStatusPending,
			},
			expectError: true,
			errorField:  "Total",
		},
		{
			name: "missing status",
			order: Order{
				ID:         primitive.NewObjectID(),
				CustomerID: &customerID,
				Total:      29.99,
			},
			expectError: true,
			errorField:  "Status",
		},
		{
			name: "invalid status",
			order: Order{
				ID:         primitive.NewObjectID(),
				CustomerID: &customerID,
				Total:      29.99,
				Status:     OrderStatus("invalid"),
			},
			expectError: true,
			errorField:  "Status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
        err := validate.Struct(tt.order)
			
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

func TestOrderStatus_Constants(t *testing.T) {
	assert.Equal(t, OrderStatus("pending"), OrderStatusPending)
	assert.Equal(t, OrderStatus("paid"), OrderStatusPaid)
	assert.Equal(t, OrderStatus("cancelled"), OrderStatusCancelled)
	assert.Equal(t, OrderStatus("refunded"), OrderStatusRefunded)
}

func stringPtr(s string) *string {
	return &s
}

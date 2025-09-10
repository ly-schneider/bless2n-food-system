package domain

import (
    "testing"

    v10 "github.com/go-playground/validator/v10"
    "github.com/stretchr/testify/assert"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

func TestProduct_Validation(t *testing.T) {
    validate := v10.New()
	categoryID := primitive.NewObjectID()
	
	tests := []struct {
		name        string
		product     Product
		expectError bool
		errorField  string
	}{
		{
			name: "valid simple product",
			product: Product{
				ID:         primitive.NewObjectID(),
				CategoryID: categoryID,
				Type:       ProductTypeSimple,
				Name:       "Test Product",
				Price:      9.99,
				IsActive:   true,
			},
			expectError: false,
		},
		{
			name: "valid bundle product",
			product: Product{
				ID:         primitive.NewObjectID(),
				CategoryID: categoryID,
				Type:       ProductTypeBundle,
				Name:       "Test Bundle",
				Price:      19.99,
				IsActive:   true,
			},
			expectError: false,
		},
		{
			name: "missing category ID",
			product: Product{
				Type:     ProductTypeSimple,
				Name:     "Test Product",
				Price:    9.99,
				IsActive: true,
			},
			expectError: true,
			errorField:  "CategoryID",
		},
		{
			name: "missing type",
			product: Product{
				ID:         primitive.NewObjectID(),
				CategoryID: categoryID,
				Name:       "Test Product",
				Price:      9.99,
				IsActive:   true,
			},
			expectError: true,
			errorField:  "Type",
		},
		{
			name: "invalid type",
			product: Product{
				ID:         primitive.NewObjectID(),
				CategoryID: categoryID,
				Type:       ProductType("invalid"),
				Name:       "Test Product",
				Price:      9.99,
				IsActive:   true,
			},
			expectError: true,
			errorField:  "Type",
		},
		{
			name: "missing name",
			product: Product{
				ID:         primitive.NewObjectID(),
				CategoryID: categoryID,
				Type:       ProductTypeSimple,
				Price:      9.99,
				IsActive:   true,
			},
			expectError: true,
			errorField:  "Name",
		},
		{
			name: "missing price",
			product: Product{
				ID:         primitive.NewObjectID(),
				CategoryID: categoryID,
				Type:       ProductTypeSimple,
				Name:       "Test Product",
				IsActive:   true,
			},
			expectError: true,
			errorField:  "Price",
		},
		{
			name: "negative price",
			product: Product{
				ID:         primitive.NewObjectID(),
				CategoryID: categoryID,
				Type:       ProductTypeSimple,
				Name:       "Test Product",
				Price:      -5.00,
				IsActive:   true,
			},
			expectError: true,
			errorField:  "Price",
		},
        {
            name: "zero price is invalid due to 'required'",
            product: Product{
                ID:         primitive.NewObjectID(),
                CategoryID: categoryID,
                Type:       ProductTypeSimple,
                Name:       "Free Product",
                Price:      0,
                IsActive:   true,
            },
            expectError: true,
            errorField:  "Price",
        },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
        err := validate.Struct(tt.product)
			
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

func TestProductType_Constants(t *testing.T) {
	assert.Equal(t, ProductType("simple"), ProductTypeSimple)
	assert.Equal(t, ProductType("bundle"), ProductTypeBundle)
}

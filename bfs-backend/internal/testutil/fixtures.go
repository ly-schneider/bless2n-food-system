package testutil

import (
	"time"

	"backend/internal/domain"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CreateTestUser creates a test user with default values
func CreateTestUser(email string, role domain.UserRole) *domain.User {
	id := primitive.NewObjectID()
	now := time.Now()
	
	user := &domain.User{
		ID:         id,
		Email:      email,
		Role:       role,
		IsVerified: true,
		IsDisabled: false,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	
	if role == domain.UserRoleAdmin {
		user.FirstName = "Test"
		user.LastName = "Admin"
	}
	
	return user
}

// CreateTestCustomer creates a test customer user
func CreateTestCustomer(email string) *domain.User {
	return CreateTestUser(email, domain.UserRoleCustomer)
}

// CreateTestAdmin creates a test admin user
func CreateTestAdmin(email string) *domain.User {
	return CreateTestUser(email, domain.UserRoleAdmin)
}

// CreateTestCategory creates a test category
func CreateTestCategory(name string) *domain.Category {
    id := primitive.NewObjectID()
    now := time.Now()

    return &domain.Category{
        ID:        id,
        Name:      name,
        IsActive:  true,
        CreatedAt: now,
        UpdatedAt: now,
    }
}

// CreateTestProduct creates a test product
func CreateTestProduct(categoryID primitive.ObjectID, name string, price float64) *domain.Product {
	id := primitive.NewObjectID()
	now := time.Now()
	
	return &domain.Product{
		ID:         id,
		CategoryID: categoryID,
		Type:       domain.ProductTypeSimple,
		Name:       name,
		Price:      price,
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// CreateTestOrder creates a test order
func CreateTestOrder(customerID *primitive.ObjectID, total float64) *domain.Order {
	id := primitive.NewObjectID()
	now := time.Now()
	
	return &domain.Order{
		ID:         id,
		CustomerID: customerID,
		Total:      total,
		Status:     domain.OrderStatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// CreateTestStation creates a test station (pending by default)
func CreateTestStation(name string) *domain.Station {
    id := primitive.NewObjectID()
    now := time.Now()

    return &domain.Station{
        ID:        id,
        Name:      name,
        Status:    domain.StationStatusPending,
        CreatedAt: now,
        UpdatedAt: now,
    }
}

// CreateTestAdminInvite creates a test admin invite
func CreateTestAdminInvite(invitedBy primitive.ObjectID, inviteeEmail string) *domain.AdminInvite {
    id := primitive.NewObjectID()
    now := time.Now()
    expiresAt := now.Add(24 * time.Hour)

    return &domain.AdminInvite{
        ID:           id,
        InvitedBy:    invitedBy,
        InviteeEmail: inviteeEmail,
        ExpiresAt:    expiresAt,
    }
}

// CreateTestOTPToken creates a test OTP token
func CreateTestOTPToken(userID primitive.ObjectID, tokenType domain.TokenType) *domain.OTPToken {
    id := primitive.NewObjectID()
    now := time.Now()
    expiresAt := now.Add(10 * time.Minute)

    return &domain.OTPToken{
        ID:        id,
        UserID:    userID,
        TokenHash: "hash:123456",
        Type:      tokenType,
        CreatedAt: now,
        ExpiresAt: expiresAt,
        Attempts:  0,
    }
}

// CreateTestRefreshToken creates a test refresh token
func CreateTestRefreshToken(userID primitive.ObjectID, tokenHash string, clientID string) *domain.RefreshToken {
    id := primitive.NewObjectID()
    now := time.Now()
    expiresAt := now.Add(7 * 24 * time.Hour)

    return &domain.RefreshToken{
        ID:        id,
        UserID:    userID,
        ClientID:  clientID,
        TokenHash: tokenHash,
        IssuedAt:  now,
        LastUsedAt: now,
        ExpiresAt: expiresAt,
        IsRevoked: false,
        FamilyID:  primitive.NewObjectID().Hex(),
    }
}

// CreateTestProductBundleComponent creates a test product bundle component
func CreateTestProductBundleComponent(bundleID primitive.ObjectID, componentID primitive.ObjectID, quantity int) *domain.ProductBundleComponent {
	return &domain.ProductBundleComponent{
		BundleID:           bundleID,
		ComponentProductID: componentID,
		Quantity:           quantity,
	}
}

// CreateTestStationProduct creates a test station product
func CreateTestStationProduct(stationID primitive.ObjectID, productID primitive.ObjectID) *domain.StationProduct {
	return &domain.StationProduct{
		StationID: stationID,
		ProductID: productID,
	}
}

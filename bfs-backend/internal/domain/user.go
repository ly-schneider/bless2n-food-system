package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserRole string

const (
	UserRoleAdmin    UserRole = "admin"
	UserRoleCustomer UserRole = "customer"
)

type User struct {
	ID         primitive.ObjectID `bson:"_id"`
	Email      string             `bson:"email" validate:"required,email"`
	FirstName  string             `bson:"first_name,omitempty"` // Only for admins
	LastName   string             `bson:"last_name,omitempty"`  // Only for admins
	Role       UserRole           `bson:"role" validate:"required,oneof=admin customer"`
	IsVerified bool               `bson:"is_verified"`
	// StripeCustomerID stores the associated Stripe Customer ID for this user,
	// used to prefill and suppress email collection on Checkout when logged in.
	StripeCustomerID *string   `bson:"stripe_customer_id,omitempty"`
	CreatedAt        time.Time `bson:"created_at"`
	UpdatedAt        time.Time `bson:"updated_at"`
}

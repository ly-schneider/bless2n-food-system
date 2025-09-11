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
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email          string             `bson:"email" json:"email" validate:"required,email"`
	FirstName      string             `bson:"first_name,omitempty" json:"first_name,omitempty"` // Only for admins
	LastName       string             `bson:"last_name,omitempty" json:"last_name,omitempty"`   // Only for admins
	Role           UserRole           `bson:"role" json:"role" validate:"required,oneof=admin customer"`
	RoleID         int                `bson:"role_id,omitempty" json:"role_id,omitempty"` // For backward compatibility
	IsVerified     bool               `bson:"is_verified" json:"is_verified"`
	IsDisabled     bool               `bson:"is_disabled" json:"is_disabled"`
	DisabledReason *string            `bson:"disabled_reason,omitempty" json:"disabled_reason,omitempty"`
	CreatedAt      time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
}

package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID             uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	FirstName      string    `gorm:"not null" json:"first_name" validate:"required"`
	LastName       string    `gorm:"not null" json:"last_name" validate:"required"`
	Email          string    `gorm:"unique;not null" json:"email" validate:"required,email"`
	PasswordHash   string    `gorm:"not null" json:"-"`
	IsVerified     bool      `gorm:"default:false" json:"is_verified"`
	MfaEnabled     bool      `gorm:"default:false" json:"mfa_enabled"`
	IsDisabled     bool      `gorm:"default:false" json:"is_disabled"`
	DisabledReason *string   `json:"disabled_reason,omitempty"`
	RoleID         uint      `gorm:"not null" json:"role_id"`
	Role           *Role     `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

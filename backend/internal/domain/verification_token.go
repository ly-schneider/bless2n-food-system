package domain

import (
	"time"

	"github.com/google/uuid"
)

type VerificationToken struct {
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	User      User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Token     string    `gorm:"primary_key" json:"token"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
}

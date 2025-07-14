package domain

import (
	"backend/internal/model"
	"time"
)

type VerificationToken struct {
	UserID    model.NanoID14 `gorm:"type:nano_id;primaryKey" json:"user_id"`
	TokenHash string         `gorm:"type:argon2_hash;primaryKey" json:"-" validate:"required"`
	ExpiresAt time.Time      `gorm:"not null" json:"expires_at" validate:"required"`
	User      *User          `gorm:"constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

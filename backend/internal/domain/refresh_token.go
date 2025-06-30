package domain

import (
	"backend/internal/model"
	"time"
)

type RefreshToken struct {
	UserID    model.NanoID14 `gorm:"type:nano_id;primaryKey"        json:"user_id"`
	TokenHash string         `gorm:"type:argon2_hash;primaryKey"    json:"-"        validate:"required"`
	IssuedAt  time.Time      `gorm:"not null;autoCreateTime"        json:"issued_at"`
	ExpiresAt time.Time      `gorm:"not null"                       json:"expires_at" validate:"required"`
	IsRevoked bool           `gorm:"not null;default:false"         json:"is_revoked"`
	User      *User          `gorm:"constraint:OnDelete:CASCADE"    json:"user,omitempty"`
}

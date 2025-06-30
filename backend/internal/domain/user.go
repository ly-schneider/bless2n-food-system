package domain

import (
	"backend/internal/model"
	"backend/internal/utils"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID             model.NanoID14 `gorm:"type:nano_id;primaryKey" json:"id"`
	FirstName      string         `gorm:"not null"         json:"first_name" validate:"required"`
	LastName       string         `gorm:"not null"         json:"last_name"  validate:"required"`
	Email          string         `gorm:"unique;not null"  json:"email"      validate:"required,email"`
	PasswordHash   string         `gorm:"type:argon2_hash;not null" json:"-"`
	IsVerified     bool           `gorm:"default:false" json:"is_verified"`
	IsDisabled     bool           `gorm:"default:false" json:"is_disabled"`
	DisabledReason *string        `json:"disabled_reason,omitempty"`
	RoleID         uint           `gorm:"not null;index"                         json:"role_id"`
	Role           *Role          `gorm:"constraint:OnDelete:RESTRICT"           json:"role,omitempty"`
	CreatedAt      time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      *time.Time     `gorm:"autoUpdateTime" json:"updated_at,omitempty"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = model.NanoID14(utils.Must())
	}
	return nil
}

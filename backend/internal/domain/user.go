package domain

import (
	"backend/internal/model"
	"backend/internal/utils"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID             model.NanoID14 `gorm:"type:char(14);primaryKey;collate:C" json:"id"`
	FirstName      string         `gorm:"not null" json:"first_name" validate:"required"`
	LastName       string         `gorm:"not null" json:"last_name" validate:"required"`
	Email          string         `gorm:"unique;not null" json:"email" validate:"required,email"`
	PasswordHash   string         `gorm:"not null" json:"-"`
	IsVerified     bool           `gorm:"default:false" json:"is_verified"`
	IsDisabled     bool           `gorm:"default:false" json:"is_disabled"`
	DisabledReason *string        `json:"disabled_reason,omitempty"`
	RoleID         uint           `gorm:"not null" json:"role_id"`
	Role           *Role          `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

/* ---------- auto-generate IDs ---------- */

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = model.NanoID14(utils.Must())
	}
	return nil
}

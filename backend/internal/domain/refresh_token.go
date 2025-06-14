package domain

import (
	"backend/internal/model"
	"backend/internal/utils"
	"time"

	"gorm.io/gorm"
)

type RefreshToken struct {
	ID        model.NanoID14 `gorm:"type:char(14);primaryKey;collate:C" json:"id"`
	UserID    model.NanoID14 `gorm:"type:char(14);not null" json:"user_id"`
	User      User           `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	TokenHash string         `gorm:"not null" json:"-"`
	IssuedAt  time.Time      `gorm:"not null;default:now()" json:"issued_at"`
	ExpiresAt time.Time      `gorm:"not null" json:"expires_at"`
	Revoked   bool           `gorm:"default:false;not null" json:"revoked"`
}

/* ---------- auto-generate IDs ---------- */

func (r *RefreshToken) BeforeCreate(tx *gorm.DB) error {
	if r.ID == "" {
		r.ID = model.NanoID14(utils.Must())
	}
	return nil
}

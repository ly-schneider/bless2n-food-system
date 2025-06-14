package domain

import (
	"backend/internal/model"
	"backend/internal/utils"
	"time"
)

type VerificationToken struct {
	ID        model.NanoID14 `gorm:"type:char(14);primaryKey;collate:C" json:"id"`
	UserID    model.NanoID14 `gorm:"type:char(14);not null" json:"user_id"`
	User      User           `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Token     uint           `gorm:"not null" json:"token"`
	ExpiresAt time.Time      `gorm:"not null" json:"expires_at"`
}

/* ---------- auto-generate IDs ---------- */

func (v *VerificationToken) BeforeCreate() error {
	if v.ID == "" {
		v.ID = model.NanoID14(utils.Must())
	}
	return nil
}

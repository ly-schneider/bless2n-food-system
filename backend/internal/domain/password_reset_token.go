package domain

import (
	"backend/internal/model"
	"backend/internal/utils"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PasswordResetToken struct {
	ID        model.NanoID14  `gorm:"type:char(14);primaryKey;collate:C" json:"id"`
	UserID    *model.NanoID14 `gorm:"type:char(14);not null" json:"user_id"`
	User      User            `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
	Token     uuid.UUID       `gorm:"type:uuid;not null;default:gen_random_uuid()" json:"token"`
	ExpiresAt time.Time       `gorm:"not null" json:"expires_at"`
}

/* ---------- auto-generate IDs ---------- */

func (p *PasswordResetToken) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = model.NanoID14(utils.Must())
	}
	return nil
}

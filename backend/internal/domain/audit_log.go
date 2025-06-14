package domain

import (
	"backend/internal/model"
	"backend/internal/utils"
	"net"
	"time"

	"gorm.io/gorm"
)

// Why collate:C?
// A binary collation avoids locale-dependent ordering quirks and makes B-tree comparisons marginally faster, while taking 15 B per row (14 B data + 1 B header)
type AuditLog struct {
	ID        model.NanoID14  `gorm:"type:char(14);primaryKey;collate:C" json:"id"`
	UserID    *model.NanoID14 `gorm:"type:char(14)" json:"user_id,omitempty"`
	User      *User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	IP        net.IP          `gorm:"type:inet" json:"ip,omitempty"`
	Event     string          `json:"event"`
	CreatedAt time.Time       `gorm:"default:now()" json:"created_at"`
}

/* ---------- auto-generate IDs ---------- */

func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = model.NanoID14(utils.Must())
	}
	return nil
}

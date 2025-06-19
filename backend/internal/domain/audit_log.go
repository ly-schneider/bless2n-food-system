package domain

import (
	"backend/internal/model"
	"net"
	"time"
)

type AuditLog struct {
	ID        uint           `gorm:"primaryKey;autoIncrement"   json:"id"`
	UserID    model.NanoID14 `gorm:"type:nano_id;not null"      json:"user_id"`
	PublicIP  net.IP         `gorm:"type:inet;not null"         json:"public_ip"  validate:"required,ip"`
	Event     string         `gorm:"not null"                   json:"event"      validate:"required"`
	CreatedAt time.Time      `gorm:"autoCreateTime"             json:"created_at"`
	User      *User          `gorm:"constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

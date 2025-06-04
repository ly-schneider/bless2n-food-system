package domain

import (
	"net"
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID    *uuid.UUID `gorm:"type:uuid" json:"user_id,omitempty"`
	User      *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	IP        net.IP     `gorm:"type:inet" json:"ip,omitempty"`
	Event     string     `json:"event"`
	CreatedAt time.Time  `gorm:"default:now()" json:"created_at"`
}

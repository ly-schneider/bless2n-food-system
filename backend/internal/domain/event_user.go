package domain

import (
	"backend/internal/model"
	"time"
)

type EventUser struct {
	EventID     model.NanoID14 `gorm:"type:nano_id;primaryKey"                                         json:"event_id"`
	UserID      model.NanoID14 `gorm:"type:nano_id;primaryKey"                                         json:"user_id"`
	EventRoleID uint           `gorm:"column:event_role;not null"                                      json:"event_role"`
	InvitedAt   time.Time      `gorm:"not null;autoCreateTime"                                         json:"invited_at"`
	JoinedAt    *time.Time     `json:"joined_at,omitempty"`
	Event       *Event         `gorm:"constraint:OnDelete:CASCADE"                                     json:"event,omitempty"`
	User        *User          `gorm:"constraint:OnDelete:CASCADE"                                     json:"user,omitempty"`
	EventRole   *EventRole     `gorm:"foreignKey:EventRoleID;references:ID;constraint:OnDelete:RESTRICT" json:"event_role_detail,omitempty"`
}

package domain

import (
	"backend/internal/model"
	"backend/internal/utils"
	"time"

	"gorm.io/gorm"
)

type EventInvite struct {
	ID           model.NanoID14 `gorm:"type:nano_id;primaryKey"      json:"id"`
	EventID      model.NanoID14 `gorm:"type:nano_id;not null"        json:"event_id"`
	EventRoleID  uint           `gorm:"column:event_role;not null"   json:"event_role"`
	InviteeEmail string         `gorm:"not null"                     json:"invitee_email" validate:"required,email"`
	ExpiresAt    time.Time      `gorm:"not null"                     json:"expires_at"`
	Event        *Event         `gorm:"constraint:OnDelete:CASCADE"  json:"event,omitempty"`
	EventRole    *EventRole     `gorm:"foreignKey:EventRoleID;references:ID;constraint:OnDelete:RESTRICT" json:"event_role_detail,omitempty"`
}

func (e *EventInvite) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = model.NanoID14(utils.Must())
	}
	return nil
}

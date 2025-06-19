package domain

import (
	"backend/internal/model"
	"backend/internal/utils"
	"time"

	"gorm.io/gorm"
)

type EventCategory struct {
	ID        model.NanoID14 `gorm:"type:nano_id;primaryKey"                                                        json:"id"`
	EventID   model.NanoID14 `gorm:"type:nano_id;not null"                                                           json:"event_id"`
	Event     *Event         `gorm:"constraint:OnDelete:CASCADE"                                                     json:"event,omitempty"`
	Name      string         `gorm:"type:varchar(20);not null"                                                       json:"name"          validate:"required,max=20"`
	Emoji     string         `gorm:"type:text;check:char_length(emoji) = 1 AND octet_length(emoji) <= 35"            json:"emoji"`
	IsActive  bool           `gorm:"not null;default:true"                                                           json:"is_active"`
	CreatedAt time.Time      `gorm:"autoCreateTime"                                                                  json:"created_at"`
	UpdatedAt *time.Time     `gorm:"autoUpdateTime"                                                                  json:"updated_at,omitempty"`
}

func (ec *EventCategory) BeforeCreate(tx *gorm.DB) error {
	if ec.ID == "" {
		ec.ID = model.NanoID14(utils.Must())
	}
	return nil
}

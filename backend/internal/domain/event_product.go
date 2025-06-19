package domain

import (
	"backend/internal/model"
	"backend/internal/utils"
	"time"

	"gorm.io/gorm"
)

type EventProduct struct {
	ID              model.NanoID14 `gorm:"type:nano_id;primaryKey"                                                       json:"id"`
	EventID         model.NanoID14 `gorm:"type:nano_id;not null"                                                          json:"event_id"`
	Event           *Event         `gorm:"constraint:OnDelete:CASCADE"                                                    json:"event,omitempty"`
	EventCategoryID model.NanoID14 `gorm:"type:nano_id;not null"                                                          json:"event_category_id"`
	EventCategory   *EventCategory `gorm:"constraint:OnDelete:CASCADE"                                                    json:"event_category,omitempty"`
	Name            string         `gorm:"type:varchar(30);not null"                                                      json:"name"          validate:"required,max=30"`
	Emoji           string         `gorm:"type:text;check:char_length(emoji) = 1 AND octet_length(emoji) <= 35"           json:"emoji"`
	Price           float64        `gorm:"type:numeric(6,2);not null;check:price >= 0"                                    json:"price"         validate:"required"`
	IsActive        bool           `gorm:"not null;default:true"                                                          json:"is_active"`
	CreatedAt       time.Time      `gorm:"autoCreateTime"                                                                 json:"created_at"`
	UpdatedAt       *time.Time     `gorm:"autoUpdateTime"                                                                 json:"updated_at,omitempty"`
}

func (p *EventProduct) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = model.NanoID14(utils.Must())
	}
	return nil
}

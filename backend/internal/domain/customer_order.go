package domain

import (
	"backend/internal/model"
	"backend/internal/utils"
	"time"

	"gorm.io/gorm"
)

type CustomerOrder struct {
	ID        model.NanoID14 `gorm:"type:nano_id;primaryKey"                                                                          json:"id"`
	EventID   model.NanoID14 `gorm:"type:nano_id;not null"                                                                             json:"event_id"`
	Event     *Event         `gorm:"constraint:OnDelete:CASCADE"                                                                       json:"event,omitempty"`
	DeviceID  model.NanoID14 `gorm:"type:nano_id;not null"                                                                             json:"device_id"`
	Device    *EventDevice   `gorm:"foreignKey:DeviceID;references:ID;constraint:OnDelete:CASCADE"                                     json:"device,omitempty"`
	Total     float64        `gorm:"type:numeric(6,2);not null;check:total >= 0"                                                       json:"total"         validate:"required"`
	Status    string         `gorm:"type:order_status_enum;not null"                     json:"status"        validate:"required"`
	CreatedAt time.Time      `gorm:"autoCreateTime"                                                                                     json:"created_at"`
	UpdatedAt *time.Time     `gorm:"autoUpdateTime"                                                                                     json:"updated_at,omitempty"`
}

func (o *CustomerOrder) BeforeCreate(tx *gorm.DB) error {
	if o.ID == "" {
		o.ID = model.NanoID14(utils.Must())
	}
	return nil
}

package domain

import (
	"backend/internal/model"
	"backend/internal/utils"
	"time"
)

type DevicePairing struct {
	ID            model.NanoID14 `gorm:"type:nano_id;primaryKey"                    json:"id"`
	EventDeviceID model.NanoID14 `gorm:"type:nano_id;not null"                      json:"event_device_id"`
	Role          string         `gorm:"type:device_pairing_role_enum;not null" json:"role" validate:"required"`
	CreatedAt     time.Time      `gorm:"autoCreateTime;not null"                      json:"created_at"`
	UpdatedAt     *time.Time     `gorm:"autoUpdateTime"                               json:"updated_at,omitempty"`
	EventDevice   *EventDevice   `gorm:"constraint:OnDelete:CASCADE"                  json:"event_device,omitempty"`
}

func (dp *DevicePairing) BeforeCreate() error {
	if dp.ID == "" {
		dp.ID = model.NanoID14(utils.Must())
	}
	return nil
}

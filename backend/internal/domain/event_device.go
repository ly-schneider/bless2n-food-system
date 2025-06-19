package domain

import (
	"backend/internal/model"
	"time"
)

type EventDevice struct {
	ID         uint           `gorm:"primaryKey;autoIncrement"    json:"id"`
	EventID    model.NanoID14 `gorm:"type:nano_id;not null"       json:"event_id"`
	DeviceID   model.NanoID14 `gorm:"type:nano_id;not null"       json:"device_id"`
	AssignedAt time.Time      `gorm:"autoCreateTime;not null"     json:"assigned_at"`
	Event      *Event         `gorm:"constraint:OnDelete:CASCADE" json:"event,omitempty"`
	Device     *Device        `gorm:"constraint:OnDelete:CASCADE" json:"device,omitempty"`
}

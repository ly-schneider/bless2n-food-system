package model

import (
	"time"

	"github.com/google/uuid"
)

// Device represents a POS or station device.
type Device struct {
	ID        uuid.UUID    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string       `gorm:"type:varchar(20);not null"`
	Model     *string      `gorm:"type:varchar(50)"`
	OS        *string      `gorm:"type:varchar(20)"`
	DeviceKey string       `gorm:"type:varchar(100);not null;uniqueIndex:idx_device_device_key"`
	Type      DeviceType   `gorm:"type:device_type;not null"`
	Status    CommonStatus `gorm:"type:common_status;not null;default:'pending'"`
	DecidedBy *uuid.UUID   `gorm:"type:uuid"`
	DecidedAt *time.Time   `gorm:"type:timestamptz"`
	ExpiresAt *time.Time   `gorm:"type:timestamptz"`
	CreatedAt time.Time    `gorm:"not null;autoCreateTime"`
	UpdatedAt time.Time    `gorm:"not null;autoUpdateTime"`

	// Relations
	Products []DeviceProduct `gorm:"foreignKey:DeviceID"`
}

func (Device) TableName() string {
	return "app.device"
}

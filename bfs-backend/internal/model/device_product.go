package model

import (
	"github.com/google/uuid"
)

// DeviceProduct represents the products assigned to a device.
type DeviceProduct struct {
	DeviceID  uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProductID uuid.UUID `gorm:"type:uuid;primaryKey"`

	// Relations
	Device  Device  `gorm:"foreignKey:DeviceID"`
	Product Product `gorm:"foreignKey:ProductID"`
}

func (DeviceProduct) TableName() string {
	return "app.device_product"
}

package model

import (
	"time"

	"github.com/google/uuid"
)

// InventoryLedger represents an inventory adjustment entry.
type InventoryLedger struct {
	ID          uuid.UUID       `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ProductID   uuid.UUID       `gorm:"type:uuid;not null"`
	Delta       int             `gorm:"not null"`
	Reason      InventoryReason `gorm:"type:inventory_reason;not null"`
	CreatedAt   time.Time       `gorm:"not null;autoCreateTime"`
	OrderID     *uuid.UUID      `gorm:"type:uuid"`
	OrderLineID *uuid.UUID      `gorm:"type:uuid"`
	DeviceID    *uuid.UUID      `gorm:"type:uuid"`
	CreatedBy   *uuid.UUID      `gorm:"type:uuid"`

	// Relations
	Product   Product    `gorm:"foreignKey:ProductID"`
	Order     *Order     `gorm:"foreignKey:OrderID"`
	OrderLine *OrderLine `gorm:"foreignKey:OrderLineID"`
	Device    *Device    `gorm:"foreignKey:DeviceID"`
}

func (InventoryLedger) TableName() string {
	return "app.inventory_ledger"
}

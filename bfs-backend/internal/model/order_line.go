package model

import (
	"github.com/google/uuid"
)

// OrderLine represents a line item in an order.
type OrderLine struct {
	ID             uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OrderID        uuid.UUID     `gorm:"type:uuid;not null"`
	LineType       OrderItemType `gorm:"type:order_item_type;not null"`
	ProductID      uuid.UUID     `gorm:"type:uuid;not null"`
	Title          string        `gorm:"type:varchar(20);not null"`
	Quantity       int           `gorm:"not null;default:1"`
	UnitPriceCents int64         `gorm:"not null;default:0"`
	ParentLineID   *uuid.UUID    `gorm:"type:uuid"`
	MenuSlotID     *uuid.UUID    `gorm:"type:uuid"`
	MenuSlotName   *string       `gorm:"type:varchar(20)"`

	// Relations
	Order      Order        `gorm:"foreignKey:OrderID"`
	Product    Product      `gorm:"foreignKey:ProductID"`
	ParentLine *OrderLine   `gorm:"foreignKey:ParentLineID"`
	ChildLines []OrderLine  `gorm:"foreignKey:ParentLineID"`
	MenuSlot   *MenuSlot    `gorm:"foreignKey:MenuSlotID"`
	Redemption *OrderLineRedemption `gorm:"foreignKey:OrderLineID"`
}

func (OrderLine) TableName() string {
	return "app.order_line"
}

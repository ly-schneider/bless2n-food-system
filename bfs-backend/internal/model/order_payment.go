package model

import (
	"time"

	"github.com/google/uuid"
)

// OrderPayment represents a payment for an order.
type OrderPayment struct {
	ID            uuid.UUID     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OrderID       uuid.UUID     `gorm:"type:uuid;not null"`
	Method        PaymentMethod `gorm:"type:payment_method;not null"`
	AmountCents   int64         `gorm:"not null"`
	ReceivedCents int64         `gorm:"not null;default:0"`
	DeviceID      *uuid.UUID    `gorm:"type:uuid"`
	PaidAt        time.Time     `gorm:"not null;default:now()"`

	// Relations
	Order  Order   `gorm:"foreignKey:OrderID"`
	Device *Device `gorm:"foreignKey:DeviceID"`
}

func (OrderPayment) TableName() string {
	return "app.order_payment"
}

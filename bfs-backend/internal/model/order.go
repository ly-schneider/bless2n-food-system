package model

import (
	"time"

	"github.com/google/uuid"
)

// Order represents a customer order.
type Order struct {
	ID           uuid.UUID   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	CustomerID   *uuid.UUID  `gorm:"type:uuid"`
	ContactEmail *string     `gorm:"type:varchar(50)"`
	TotalCents   int64       `gorm:"not null"`
	Status       OrderStatus `gorm:"type:order_status;not null"`
	Origin       OrderOrigin `gorm:"type:order_origin;not null"`
	CreatedAt    time.Time   `gorm:"not null;autoCreateTime"`
	UpdatedAt    time.Time   `gorm:"not null;autoUpdateTime"`

	// Relations
	Payments []OrderPayment `gorm:"foreignKey:OrderID"`
	Lines    []OrderLine    `gorm:"foreignKey:OrderID"`
}

func (Order) TableName() string {
	return "app.order"
}

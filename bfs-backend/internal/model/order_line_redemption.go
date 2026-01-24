package model

import (
	"time"

	"github.com/google/uuid"
)

// OrderLineRedemption represents a redemption record for an order line.
type OrderLineRedemption struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	OrderLineID uuid.UUID `gorm:"type:uuid;not null"`
	RedeemedAt  time.Time `gorm:"not null;default:now()"`

	// Relations
	OrderLine OrderLine `gorm:"foreignKey:OrderLineID"`
}

func (OrderLineRedemption) TableName() string {
	return "app.order_line_redemption"
}

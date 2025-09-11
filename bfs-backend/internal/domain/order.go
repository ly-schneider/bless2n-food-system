package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusCancelled OrderStatus = "cancelled"
	OrderStatusRefunded  OrderStatus = "refunded"
)

type Order struct {
	ID           primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	CustomerID   *primitive.ObjectID `bson:"customer_id,omitempty" json:"customer_id,omitempty"`
	ContactEmail *string             `bson:"contact_email,omitempty" json:"contact_email,omitempty"`
	Total        float64             `bson:"total" json:"total" validate:"required,gte=0"`
	Status       OrderStatus         `bson:"status" json:"status" validate:"required,oneof=pending paid cancelled refunded"`
	CreatedAt    time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time           `bson:"updated_at" json:"updated_at"`
}

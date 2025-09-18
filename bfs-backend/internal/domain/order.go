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
	ID           primitive.ObjectID  `bson:"_id"`
	CustomerID   *primitive.ObjectID `bson:"customer_id,omitempty"`
	ContactEmail *string             `bson:"contact_email,omitempty"`
	TotalCents   Cents               `bson:"total_cents" validate:"required,gte=0"`
	Status       OrderStatus         `bson:"status" validate:"required,oneof=pending paid cancelled refunded"`
	CreatedAt    time.Time           `bson:"created_at"`
	UpdatedAt    time.Time           `bson:"updated_at"`
}

type CreateOrderDTO struct {
	ContactEmail *string              `json:"contactEmail,omitempty"`
	OrderItems   []CreateOrderItemDTO `json:"orderItems" validate:"required,dive,required"`
}

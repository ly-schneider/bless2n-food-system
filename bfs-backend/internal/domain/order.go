package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusCancelled OrderStatus = "cancelled"
	OrderStatusRefunded  OrderStatus = "refunded"
)

type Order struct {
	ID           bson.ObjectID  `bson:"_id"`
	CustomerID   *bson.ObjectID `bson:"customer_id,omitempty"`
	ContactEmail *string        `bson:"contact_email,omitempty"`
	TotalCents   Cents          `bson:"total_cents" validate:"required,gte=0"`
	Status       OrderStatus    `bson:"status" validate:"required,oneof=pending paid cancelled refunded"`
	// Origin of the order: default "shop" (web shop) or "pos" (in-person terminal)
	Origin *OrderOrigin `bson:"origin,omitempty"`
	// Deprecated: previous Stripe Checkout session id
	StripeSessionID *string `bson:"stripe_session_id,omitempty"`
	// New: Payment Intents references
	StripePaymentIntentID *string `bson:"stripe_payment_intent_id,omitempty"`
	StripeChargeID        *string `bson:"stripe_charge_id,omitempty"`
	StripeCustomerID      *string `bson:"stripe_customer_id,omitempty"`
	// Idempotency key from client to avoid duplicate orders/PI
	PaymentAttemptID *string `bson:"payment_attempt_id,omitempty"`
	// POS payments summary when origin == pos
	PosPayment *PosPayment `bson:"pos_payment,omitempty"`
	CreatedAt  time.Time   `bson:"created_at"`
	UpdatedAt  time.Time   `bson:"updated_at"`
}

type CreateOrderDTO struct {
	ContactEmail *string              `json:"contactEmail,omitempty"`
	OrderItems   []CreateOrderItemDTO `json:"orderItems" validate:"required,dive,required"`
}

// OrderOrigin enumerates order creation channels.
type OrderOrigin string

const (
	OrderOriginShop OrderOrigin = "shop"
	OrderOriginPOS  OrderOrigin = "pos"
)

// PosPayment stores POS tender details attached to an order.
type PosPayment struct {
	Method              string  `bson:"method" json:"method"` // "cash" or "card"
	AmountReceivedCents *Cents  `bson:"amount_received_cents,omitempty" json:"amountReceivedCents,omitempty"`
	ChangeCents         *Cents  `bson:"change_cents,omitempty" json:"changeCents,omitempty"`
	Processor           *string `bson:"processor,omitempty" json:"processor,omitempty"` // e.g., "sumup"
	TransactionID       *string `bson:"transaction_id,omitempty" json:"transactionId,omitempty"`
	Status              *string `bson:"status,omitempty" json:"status,omitempty"` // succeeded / canceled / failed
}

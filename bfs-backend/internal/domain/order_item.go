package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrderItemType string

const (
	OrderItemTypeSimple    OrderItemType = "simple"
	OrderItemTypeBundle    OrderItemType = "bundle"
	OrderItemTypeComponent OrderItemType = "component"
)

type OrderItem struct {
	ID                primitive.ObjectID  `bson:"_id"`
	OrderID           primitive.ObjectID  `bson:"order_id" validate:"required"`
	ProductID         primitive.ObjectID  `bson:"product_id" validate:"required"`
	Title             string              `bson:"title" validate:"required"`
	Quantity          int                 `bson:"quantity" validate:"required,gt=0"`
	PricePerUnitCents Cents               `bson:"price_per_unit_cents" validate:"required,gte=0"`
	ParentItemID      *primitive.ObjectID `bson:"parent_item_id,omitempty"`
	MenuSlotID        *primitive.ObjectID `bson:"menu_slot_id,omitempty"`
	MenuSlotName      *string             `bson:"menu_slot_name,omitempty"`
	IsRedeemed        bool                `bson:"is_redeemed"`
	RedeemedAt        *time.Time          `bson:"redeemed_at,omitempty"`
}

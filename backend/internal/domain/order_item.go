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
	ID                primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	OrderID           primitive.ObjectID  `bson:"order_id" json:"order_id" validate:"required"`
	ProductID         primitive.ObjectID  `bson:"product_id" json:"product_id" validate:"required"`
	ParentItemID      *primitive.ObjectID `bson:"parent_item_id,omitempty" json:"parent_item_id,omitempty"`
	Type              OrderItemType       `bson:"type" json:"type" validate:"required,oneof=simple bundle component"`
	Title             string              `bson:"title" json:"title" validate:"required"`
	Quantity          int                 `bson:"quantity" json:"quantity" validate:"required,gt=0"`
	PricePerUnit      float64             `bson:"price_per_unit" json:"price_per_unit" validate:"required,gte=0"`
	IsRedeemed        bool                `bson:"is_redeemed" json:"is_redeemed"`
	RedeemedAt        *time.Time          `bson:"redeemed_at,omitempty" json:"redeemed_at,omitempty"`
	RedeemedStationID *primitive.ObjectID `bson:"redeemed_station_id,omitempty" json:"redeemed_station_id,omitempty"`
	RedeemedDeviceID  *primitive.ObjectID `bson:"redeemed_device_id,omitempty" json:"redeemed_device_id,omitempty"`
}

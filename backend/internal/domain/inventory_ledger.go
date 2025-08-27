package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type InventoryReason string

const (
	InventoryReasonOpeningBalance InventoryReason = "opening_balance"
	InventoryReasonSale           InventoryReason = "sale"
	InventoryReasonRefund         InventoryReason = "refund"
	InventoryReasonManualAdjust   InventoryReason = "manual_adjust"
	InventoryReasonCorrection     InventoryReason = "correction"
)

type InventoryLedger struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ProductID primitive.ObjectID `bson:"product_id" json:"product_id" validate:"required"`
	Delta     int                `bson:"delta" json:"delta" validate:"required"`
	Reason    InventoryReason    `bson:"reason" json:"reason" validate:"required,oneof=opening_balance sale refund manual_adjust correction"`
	Timestamp time.Time          `bson:"ts" json:"ts"`
}
package domain

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MenuSlotItem struct {
	ID         primitive.ObjectID `bson:"_id"`
	MenuSlotID primitive.ObjectID `bson:"menu_slot_id" validate:"required"`
	ProductID  primitive.ObjectID `bson:"product_id" validate:"required"`
}

type MenuSlotItemDTO = ProductSummaryDTO

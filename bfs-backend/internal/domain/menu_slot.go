package domain

import (
	"go.mongodb.org/mongo-driver/v2/bson/primitive"
)

type MenuSlot struct {
	ID        primitive.ObjectID `bson:"_id"`
	ProductID primitive.ObjectID `bson:"product_id" validate:"required"`
	Name      string             `bson:"name" validate:"required"`
	Sequence  int                `bson:"sequence" validate:"required"`
}

type MenuSlotDTO struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Sequence     int               `json:"sequence"`
	MenuSlotItem []MenuSlotItemDTO `json:"menuSlotItems,omitempty"`
}

package domain

import "go.mongodb.org/mongo-driver/v2/bson"

type MenuSlot struct {
	ID        bson.ObjectID `bson:"_id"`
	ProductID bson.ObjectID `bson:"product_id" validate:"required"`
	Name      string        `bson:"name" validate:"required"`
	Sequence  int           `bson:"sequence" validate:"required"`
}

type MenuSlotDTO struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	Sequence     int               `json:"sequence"`
	MenuSlotItem []MenuSlotItemDTO `json:"menuSlotItems,omitempty"`
}

package domain

import "go.mongodb.org/mongo-driver/v2/bson"

type MenuSlotItem struct {
	ID         bson.ObjectID `bson:"_id"`
	MenuSlotID bson.ObjectID `bson:"menu_slot_id" validate:"required"`
	ProductID  bson.ObjectID `bson:"product_id" validate:"required"`
}

type MenuSlotItemDTO = ProductSummaryDTO

type CreateMenuSlotItemDTO struct {
	MenuSlotID bson.ObjectID `json:"menuSlotId" validate:"required"`
	ProductID  bson.ObjectID `json:"productId" validate:"required"`
}

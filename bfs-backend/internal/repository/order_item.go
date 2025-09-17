package repository

import (
	"backend/internal/database"

	"go.mongodb.org/mongo-driver/mongo"
)

type OrderItemRepository interface {
}

type orderItemRepository struct {
	collection *mongo.Collection
}

func NewOrderItemRepository(db *database.MongoDB) OrderItemRepository {
	return &orderItemRepository{
		collection: db.Database.Collection(database.OrderItemsCollection),
	}
}

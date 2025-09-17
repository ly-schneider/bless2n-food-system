package repository

import (
	"backend/internal/database"

	"go.mongodb.org/mongo-driver/mongo"
)

type OrderRepository interface {
}

type orderRepository struct {
	collection *mongo.Collection
}

func NewOrderRepository(db *database.MongoDB) OrderRepository {
	return &orderRepository{
		collection: db.Database.Collection(database.OrdersCollection),
	}
}

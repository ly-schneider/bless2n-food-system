package repository

import (
	"backend/internal/database"

	"go.mongodb.org/mongo-driver/mongo"
)

type StationProductRepository interface {
}

type stationProductRepository struct {
	collection *mongo.Collection
}

func NewStationProductRepository(db *database.MongoDB) StationProductRepository {
	return &stationProductRepository{
		collection: db.Database.Collection(database.StationProductsCollection),
	}
}

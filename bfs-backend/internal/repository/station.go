package repository

import (
	"backend/internal/database"

	"go.mongodb.org/mongo-driver/mongo"
)

type StationRepository interface {
}

type stationRepository struct {
	collection *mongo.Collection
}

func NewStationRepository(db *database.MongoDB) StationRepository {
	return &stationRepository{
		collection: db.Database.Collection(database.StationsCollection),
	}
}

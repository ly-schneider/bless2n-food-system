package repository

import (
	"backend/internal/database"

	"go.mongodb.org/mongo-driver/mongo"
)

type StationRequestRepository interface {
}

type stationRequestRepository struct {
	collection *mongo.Collection
}

func NewStationRequestRepository(db *database.MongoDB) StationRequestRepository {
	return &stationRequestRepository{
		collection: db.Database.Collection(database.StationRequestsCollection),
	}
}

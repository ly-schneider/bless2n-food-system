package app

import (
	"backend/internal/config"
	"backend/internal/database"

	"go.mongodb.org/mongo-driver/mongo"
)

func NewDB(cfg config.Config) (*database.MongoDB, error) {
	return database.NewMongoDB(cfg)
}

func ProvideDatabase(mongoDB *database.MongoDB) *mongo.Database {
	return mongoDB.Database
}

package app

import (
	"backend/internal/config"
	"backend/internal/database"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"gorm.io/gorm"
)

// MongoDB providers (deprecated - will be removed after migration)

func NewDB(cfg config.Config) (*database.MongoDB, error) {
	return database.NewMongoDB(cfg)
}

func ProvideDatabase(mongoDB *database.MongoDB) *mongo.Database {
	return mongoDB.Database
}

// GORM/PostgreSQL providers

func NewGormDB(cfg config.Config) (*database.GormDB, error) {
	return database.NewGormDB(cfg)
}

func ProvideGormDB(gormDB *database.GormDB) *gorm.DB {
	return gormDB.DB
}

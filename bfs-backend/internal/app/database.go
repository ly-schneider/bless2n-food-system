package app

import (
	"backend/internal/config"
	"backend/internal/database"
)

func NewDB(cfg config.Config) (*database.MongoDB, error) {
	return database.NewMongoDB(cfg)
}

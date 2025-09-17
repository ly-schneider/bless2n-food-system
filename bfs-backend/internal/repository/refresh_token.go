package repository

import (
	"backend/internal/database"

	"go.mongodb.org/mongo-driver/mongo"
)

type RefreshTokenRepository interface {
}

type refreshTokenRepository struct {
	collection *mongo.Collection
}

func NewRefreshTokenRepository(db *database.MongoDB) RefreshTokenRepository {
	return &refreshTokenRepository{
		collection: db.Database.Collection(database.RefreshTokensCollection),
	}
}

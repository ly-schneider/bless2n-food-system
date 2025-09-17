package repository

import (
	"backend/internal/database"

	"go.mongodb.org/mongo-driver/mongo"
)

type OTPTokenRepository interface {
}

type otpTokenRepository struct {
	collection *mongo.Collection
}

func NewOTPTokenRepository(db *database.MongoDB) OTPTokenRepository {
	return &otpTokenRepository{
		collection: db.Database.Collection(database.OTPTokensCollection),
	}
}

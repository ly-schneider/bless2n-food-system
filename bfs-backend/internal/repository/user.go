package repository

import (
	"backend/internal/database"

	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository interface {
}

type userRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *database.MongoDB) UserRepository {
	return &userRepository{
		collection: db.Database.Collection(database.UsersCollection),
	}
}

package repository

import (
    "backend/internal/database"
    "backend/internal/domain"
    "context"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
)

type UserRepository interface {
    FindByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error)
}

type userRepository struct {
    collection *mongo.Collection
}

func NewUserRepository(db *database.MongoDB) UserRepository {
    return &userRepository{
        collection: db.Database.Collection(database.UsersCollection),
    }
}

func (r *userRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
    var u domain.User
    if err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&u); err != nil {
        return nil, err
    }
    return &u, nil
}

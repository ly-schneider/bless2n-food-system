package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Category struct {
    ID        primitive.ObjectID `bson:"_id"`
    Name      string             `bson:"name" validate:"required"`
    IsActive  bool               `bson:"is_active"`
    Position  int                `bson:"position"`
    CreatedAt time.Time          `bson:"created_at"`
    UpdatedAt time.Time          `bson:"updated_at"`
}

type CategoryDTO struct {
    ID       string `json:"id"`
    Name     string `json:"name"`
    IsActive bool   `json:"isActive"`
    Position int    `json:"position"`
}

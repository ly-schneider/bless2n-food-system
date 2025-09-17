package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Station struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name" validate:"required"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

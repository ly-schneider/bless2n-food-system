package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type Station struct {
	ID   primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name string             `bson:"name" json:"name" validate:"required"`
}
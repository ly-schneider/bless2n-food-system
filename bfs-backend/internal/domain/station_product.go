package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type StationProduct struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	StationID primitive.ObjectID `bson:"station_id" validate:"required"`
	ProductID primitive.ObjectID `bson:"product_id" validate:"required"`
}

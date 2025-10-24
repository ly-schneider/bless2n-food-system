package domain

import "go.mongodb.org/mongo-driver/v2/bson"

type StationProduct struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	StationID bson.ObjectID `bson:"station_id" validate:"required"`
	ProductID bson.ObjectID `bson:"product_id" validate:"required"`
}

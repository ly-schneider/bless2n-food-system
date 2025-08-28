package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type StationProduct struct {
	StationID primitive.ObjectID `bson:"station_id" json:"station_id" validate:"required"`
	ProductID primitive.ObjectID `bson:"product_id" json:"product_id" validate:"required"`
}

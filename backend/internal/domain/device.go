package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type Device struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	StationID primitive.ObjectID `bson:"station_id" json:"station_id" validate:"required"`
	Name      string             `bson:"name" json:"name" validate:"required"`
	Model     string             `bson:"model" json:"model" validate:"required"`
	OS        string             `bson:"os" json:"os" validate:"required"`
	Version   string             `bson:"version" json:"version" validate:"required"`
	IsActive  bool               `bson:"is_active" json:"is_active"`
}
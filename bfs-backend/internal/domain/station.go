package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Station struct {
	ID   bson.ObjectID `bson:"_id,omitempty"`
	Name string        `bson:"name" validate:"required"`
	// DeviceKey uniquely identifies a browser/device claiming to be this station
	DeviceKey  string     `bson:"device_key" validate:"required"`
	Approved   bool       `bson:"approved"`
	ApprovedAt *time.Time `bson:"approved_at,omitempty"`
	CreatedAt  time.Time  `bson:"created_at"`
	UpdatedAt  time.Time  `bson:"updated_at"`
}

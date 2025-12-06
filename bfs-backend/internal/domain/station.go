package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type StationRequestStatus string

const (
	StationRequestStatusPending  StationRequestStatus = "pending"
	StationRequestStatusApproved StationRequestStatus = "approved"
	StationRequestStatusRejected StationRequestStatus = "rejected"
	StationRequestStatusRevoked  StationRequestStatus = "revoked"
)

type Station struct {
	ID    bson.ObjectID `bson:"_id,omitempty"`
	Name  string        `bson:"name" validate:"required"`
	Model string        `bson:"model,omitempty"`
	OS    string        `bson:"os,omitempty"`
	// DeviceKey uniquely identifies a browser/device claiming to be this station
	DeviceKey  string               `bson:"device_key" validate:"required"`
	Status     StationRequestStatus `bson:"status,omitempty"`
	Approved   bool                 `bson:"approved"`
	ApprovedAt *time.Time           `bson:"approved_at,omitempty"`
	DecidedBy  *bson.ObjectID       `bson:"decided_by,omitempty"`
	DecidedAt  *time.Time           `bson:"decided_at,omitempty"`
	ExpiresAt  *time.Time           `bson:"expires_at,omitempty"`
	CreatedAt  time.Time            `bson:"created_at"`
	UpdatedAt  time.Time            `bson:"updated_at"`
}

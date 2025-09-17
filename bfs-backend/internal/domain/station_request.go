package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StationRequestStatus string

const (
	StationRequestStatusPending  StationRequestStatus = "pending"
	StationRequestStatusApproved StationRequestStatus = "approved"
	StationRequestStatusRejected StationRequestStatus = "rejected"
)

type StationRequest struct {
	ID        primitive.ObjectID   `bson:"_id"`
	Name      string               `bson:"name" validate:"required"`
	Model     string               `bson:"model" validate:"required"`
	OS        string               `bson:"os" validate:"required"`
	Status    StationRequestStatus `bson:"status" validate:"required,oneof=pending approved rejected"`
	DecidedBy *primitive.ObjectID  `bson:"decided_by,omitempty"`
	DecidedAt *time.Time           `bson:"decided_at,omitempty"`
	CreatedAt time.Time            `bson:"created_at"`
	ExpiresAt time.Time            `bson:"expires_at"`
}

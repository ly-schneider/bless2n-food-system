package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type PosRequestStatus string

const (
	PosRequestStatusPending  PosRequestStatus = "pending"
	PosRequestStatusApproved PosRequestStatus = "approved"
	PosRequestStatusRejected PosRequestStatus = "rejected"
)

// PosRequest is created by a device to request POS access and awaits admin approval.
type PosRequest struct {
	ID          bson.ObjectID    `bson:"_id"`
	Name        string           `bson:"name" validate:"required"`
	Model       string           `bson:"model" validate:"required"`
	OS          string           `bson:"os" validate:"required"`
	DeviceToken string           `bson:"device_token" validate:"required"`
	Status      PosRequestStatus `bson:"status" validate:"required,oneof=pending approved rejected"`
	DecidedBy   *bson.ObjectID   `bson:"decided_by,omitempty"`
	DecidedAt   *time.Time       `bson:"decided_at,omitempty"`
	CreatedAt   time.Time        `bson:"created_at"`
	ExpiresAt   time.Time        `bson:"expires_at"`
}

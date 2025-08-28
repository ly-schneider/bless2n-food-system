package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type DeviceRequestStatus string

const (
	DeviceRequestStatusPending  DeviceRequestStatus = "pending"
	DeviceRequestStatusApproved DeviceRequestStatus = "approved"
	DeviceRequestStatusRejected DeviceRequestStatus = "rejected"
)

type DeviceRequest struct {
	ID         primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	DeviceID   primitive.ObjectID  `bson:"device_id" json:"device_id" validate:"required"`
	ApprovedBy *primitive.ObjectID `bson:"approved_by,omitempty" json:"approved_by,omitempty"`
	Status     DeviceRequestStatus `bson:"status" json:"status" validate:"required,oneof=pending approved rejected"`
	CreatedAt  time.Time           `bson:"created_at" json:"created_at"`
	DecidedAt  *time.Time          `bson:"decided_at,omitempty" json:"decided_at,omitempty"`
	ExpiresAt  time.Time           `bson:"expires_at" json:"expires_at" validate:"required"`
}

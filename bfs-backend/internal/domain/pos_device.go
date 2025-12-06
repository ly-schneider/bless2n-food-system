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
	PosRequestStatusRevoked  PosRequestStatus = "revoked"
)

// PosDevice represents a tablet or terminal approved to use the POS surface.
type PosDevice struct {
	ID    bson.ObjectID `bson:"_id,omitempty"`
	Name  string        `bson:"name" validate:"required"`
	Model string        `bson:"model,omitempty"`
	OS    string        `bson:"os,omitempty"`
	// DeviceToken uniquely identifies the browser/device claiming to be this POS device
	DeviceToken string           `bson:"device_token" validate:"required"`
	Status      PosRequestStatus `bson:"status,omitempty"`
	Approved    bool             `bson:"approved"`
	ApprovedAt  *time.Time       `bson:"approved_at,omitempty"`
	DecidedBy   *bson.ObjectID   `bson:"decided_by,omitempty"`
	DecidedAt   *time.Time       `bson:"decided_at,omitempty"`
	// Optional device capabilities/config
	CardCapable *bool     `bson:"card_capable,omitempty"`
	PrinterMAC  *string   `bson:"printer_mac,omitempty"`
	PrinterUUID *string   `bson:"printer_uuid,omitempty"`
	CreatedAt   time.Time `bson:"created_at"`
	UpdatedAt   time.Time `bson:"updated_at"`
}

package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type StationStatus string

const (
	StationStatusPending  StationStatus = "pending"
	StationStatusApproved StationStatus = "approved"
	StationStatusRejected StationStatus = "rejected"
)

type Station struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name" validate:"required"`
	Status      StationStatus      `bson:"status" json:"status" validate:"required,oneof=pending approved rejected"`
	ApprovedBy  *primitive.ObjectID `bson:"approved_by,omitempty" json:"approved_by,omitempty"`
	ApprovedAt  *time.Time         `bson:"approved_at,omitempty" json:"approved_at,omitempty"`
	RejectedBy  *primitive.ObjectID `bson:"rejected_by,omitempty" json:"rejected_by,omitempty"`
	RejectedAt  *time.Time         `bson:"rejected_at,omitempty" json:"rejected_at,omitempty"`
	RejectionReason *string        `bson:"rejection_reason,omitempty" json:"rejection_reason,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

type StationRequest struct {
	Name string `json:"name" validate:"required"`
}

type StationStatusRequest struct {
	Approve *bool   `json:"approve" validate:"required"`
	Reason  *string `json:"reason,omitempty"`
}

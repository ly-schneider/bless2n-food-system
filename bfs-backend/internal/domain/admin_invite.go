package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AdminInvite struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	InvitedBy    primitive.ObjectID `bson:"invited_by" json:"invited_by" validate:"required"`
	InviteeEmail string             `bson:"invitee_email" json:"invitee_email" validate:"required,email"`
	ExpiresAt    time.Time          `bson:"expires_at" json:"expires_at" validate:"required"`
}

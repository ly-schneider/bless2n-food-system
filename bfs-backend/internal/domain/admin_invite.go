package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson/primitive"
)

type AdminInvite struct {
	ID           primitive.ObjectID `bson:"_id" validate:"required"`
	InvitedBy    primitive.ObjectID `bson:"invited_by" validate:"required"`
	InviteeEmail string             `bson:"invitee_email" validate:"required,email"`
	TokenHash    string             `bson:"token_hash" validate:"required"`
	ExpiresAt    time.Time          `bson:"expires_at" validate:"required"`
	Status       string             `bson:"status" validate:"required,oneof=pending accepted expired revoked"`
	UsedAt       *time.Time         `bson:"used_at,omitempty"`
	CreatedAt    time.Time          `bson:"created_at"`
}

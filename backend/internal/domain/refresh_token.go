package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RefreshToken struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID         primitive.ObjectID `bson:"user_id" json:"user_id" validate:"required"`
	ClientID       string             `bson:"client_id" json:"client_id" validate:"required"`
	TokenHash      string             `bson:"token_hash" json:"-" validate:"required"`
	IssuedAt       time.Time          `bson:"issued_at" json:"issued_at"`
	LastUsedAt     time.Time          `bson:"last_used_at" json:"last_used_at"`
	ExpiresAt      time.Time          `bson:"expires_at" json:"expires_at" validate:"required"`
	IsRevoked      bool               `bson:"is_revoked" json:"is_revoked"`
	RevokedReason  *string            `bson:"revoked_reason,omitempty" json:"revoked_reason,omitempty"`
	FamilyID       string             `bson:"family_id" json:"family_id" validate:"required"`
}
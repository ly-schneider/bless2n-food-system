package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RefreshToken struct {
	ID            primitive.ObjectID `bson:"_id"`
	UserID        primitive.ObjectID `bson:"user_id" validate:"required"`
	ClientID      string             `bson:"client_id" validate:"required"`
	TokenHash     string             `bson:"token_hash" validate:"required"`
	IssuedAt      time.Time          `bson:"issued_at"`
	LastUsedAt    time.Time          `bson:"last_used_at"`
	ExpiresAt     time.Time          `bson:"expires_at" validate:"required"`
	IsRevoked     bool               `bson:"is_revoked"`
	RevokedReason *string            `bson:"revoked_reason,omitempty"`
	FamilyID      string             `bson:"family_id" validate:"required"`
}

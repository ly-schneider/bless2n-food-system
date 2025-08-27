package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OTPToken struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id" validate:"required"`
	TokenHash string             `bson:"token_hash" json:"-" validate:"required"`
	Type      TokenType          `bson:"type" json:"type" validate:"required,oneof=login password_reset"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UsedAt    *time.Time         `bson:"used_at,omitempty" json:"used_at,omitempty"`
	Attempts  int                `bson:"attempts" json:"attempts"`
	ExpiresAt time.Time          `bson:"expires_at" json:"expires_at" validate:"required"`
}
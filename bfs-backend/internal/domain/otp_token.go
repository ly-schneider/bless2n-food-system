package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type OTPToken struct {
	ID        bson.ObjectID `bson:"_id"`
	UserID    bson.ObjectID `bson:"user_id" validate:"required"`
	TokenHash string        `bson:"token_hash" validate:"required"`
	CreatedAt time.Time     `bson:"created_at"`
	UsedAt    *time.Time    `bson:"used_at,omitempty"`
	Attempts  int           `bson:"attempts"`
	ExpiresAt time.Time     `bson:"expires_at" validate:"required"`
}

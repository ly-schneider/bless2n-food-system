package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// EmailChangeToken stores a pending email change verification code for a user
type EmailChangeToken struct {
	ID        bson.ObjectID `bson:"_id"`
	UserID    bson.ObjectID `bson:"user_id" validate:"required"`
	NewEmail  string        `bson:"new_email" validate:"required,email"`
	TokenHash string        `bson:"token_hash" validate:"required"`
	CreatedAt time.Time     `bson:"created_at"`
	UsedAt    *time.Time    `bson:"used_at,omitempty"`
	Attempts  int           `bson:"attempts"`
	ExpiresAt time.Time     `bson:"expires_at" validate:"required"`
}

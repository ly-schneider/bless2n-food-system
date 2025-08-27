package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuditEvent string

const (
	EventUserCreated       AuditEvent = "user_created"
	EventUserLoggedIn      AuditEvent = "user_logged_in"
	EventUserLoggedOut     AuditEvent = "user_logged_out"
	EventUserRefreshedToken AuditEvent = "user_refreshed_token"
	EventUserVerified      AuditEvent = "user_verified"
	EventPasswordReset     AuditEvent = "password_reset"
)

type AuditLog struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID    primitive.ObjectID `bson:"user_id" json:"user_id" validate:"required"`
	Event     AuditEvent         `bson:"event" json:"event" validate:"required"`
	PublicIP  string             `bson:"public_ip,omitempty" json:"public_ip,omitempty"`
	UserAgent string             `bson:"user_agent,omitempty" json:"user_agent,omitempty"`
	Details   map[string]interface{} `bson:"details,omitempty" json:"details,omitempty"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
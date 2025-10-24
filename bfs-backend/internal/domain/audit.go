package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson/primitive"
)

type AuditAction string

const (
	AuditCreate AuditAction = "create"
	AuditUpdate AuditAction = "update"
	AuditDelete AuditAction = "delete"
)

type AuditLog struct {
	ID          primitive.ObjectID  `bson:"_id"`
	ActorUserID *primitive.ObjectID `bson:"actor_user_id,omitempty"`
	ActorRole   *string             `bson:"actor_role,omitempty"`
	Action      AuditAction         `bson:"action"`
	EntityType  string              `bson:"entity_type"`
	EntityID    string              `bson:"entity_id"`
	Before      interface{}         `bson:"before,omitempty"`
	After       interface{}         `bson:"after,omitempty"`
	RequestID   *string             `bson:"request_id,omitempty"`
	CreatedAt   time.Time           `bson:"created_at"`
}

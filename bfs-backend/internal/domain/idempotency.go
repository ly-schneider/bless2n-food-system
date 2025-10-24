package domain

import (
	"time"
)

// IdempotencyRecord stores results for an idempotent operation
type IdempotencyRecord struct {
	Key       string         `bson:"key"`
	Scope     string         `bson:"scope"` // e.g., station:<stationId>:order:<orderId>
	Response  map[string]any `bson:"response"`
	CreatedAt time.Time      `bson:"created_at"`
	ExpiresAt *time.Time     `bson:"expires_at,omitempty"`
}

package domain

import (
    "time"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type IdentityProvider string

const (
    ProviderGoogle IdentityProvider = "google"
)

// IdentityLink represents a federated identity mapped to a local user.
// Uniqueness is enforced on (provider, provider_user_id).
type IdentityLink struct {
    ID              primitive.ObjectID `bson:"_id"`
    UserID          primitive.ObjectID `bson:"user_id" validate:"required"`
    Provider        IdentityProvider   `bson:"provider" validate:"required,oneof=google"`
    ProviderUserID  string             `bson:"provider_user_id" validate:"required"` // OIDC subject (sub)
    EmailSnapshot   *string            `bson:"email_snapshot,omitempty"`
    DisplayName     *string            `bson:"display_name,omitempty"`
    AvatarURL       *string            `bson:"avatar_url,omitempty"`
    CreatedAt       time.Time          `bson:"created_at"`
    UpdatedAt       time.Time          `bson:"updated_at"`
}

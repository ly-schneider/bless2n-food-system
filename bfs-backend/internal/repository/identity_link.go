package repository

import (
	"backend/internal/database"
	"backend/internal/domain"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type IdentityLinkRepository interface {
	FindByProviderAndSub(ctx context.Context, provider domain.IdentityProvider, sub string) (*domain.IdentityLink, error)
	Create(ctx context.Context, link *domain.IdentityLink) (*domain.IdentityLink, error)
	UpsertLink(ctx context.Context, provider domain.IdentityProvider, sub string, userID primitive.ObjectID, email *string, name *string, avatar *string) (*domain.IdentityLink, error)
}

type identityLinkRepository struct {
	collection *mongo.Collection
}

func NewIdentityLinkRepository(db *database.MongoDB) IdentityLinkRepository {
	return &identityLinkRepository{collection: db.Database.Collection(database.IdentityLinksCollection)}
}

func (r *identityLinkRepository) FindByProviderAndSub(ctx context.Context, provider domain.IdentityProvider, sub string) (*domain.IdentityLink, error) {
	var link domain.IdentityLink
	if err := r.collection.FindOne(ctx, bson.M{"provider": provider, "provider_user_id": sub}).Decode(&link); err != nil {
		return nil, err
	}
	return &link, nil
}

func (r *identityLinkRepository) Create(ctx context.Context, link *domain.IdentityLink) (*domain.IdentityLink, error) {
	now := time.Now().UTC()
	if link.ID.IsZero() {
		link.ID = primitive.NewObjectID()
	}
	link.CreatedAt = now
	link.UpdatedAt = now
	if _, err := r.collection.InsertOne(ctx, link); err != nil {
		return nil, err
	}
	return link, nil
}

func (r *identityLinkRepository) UpsertLink(ctx context.Context, provider domain.IdentityProvider, sub string, userID primitive.ObjectID, email *string, name *string, avatar *string) (*domain.IdentityLink, error) {
	now := time.Now().UTC()
	filter := bson.M{"provider": provider, "provider_user_id": sub}
	update := bson.M{
		"$setOnInsert": bson.M{"_id": primitive.NewObjectID(), "created_at": now},
		"$set": bson.M{
			"user_id":        userID,
			"email_snapshot": email,
			"display_name":   name,
			"avatar_url":     avatar,
			"updated_at":     now,
		},
	}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
	var out domain.IdentityLink
	if err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&out); err != nil {
		return nil, err
	}
	return &out, nil
}

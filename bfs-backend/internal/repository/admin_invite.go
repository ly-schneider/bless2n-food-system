package repository

import (
	"context"
	"time"

	"backend/internal/database"
	"backend/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AdminInviteRepository interface {
	Create(ctx context.Context, invite *domain.AdminInvite) error
	GetByEmail(ctx context.Context, email string) (*domain.AdminInvite, error)
	Delete(ctx context.Context, id primitive.ObjectID) error
	DeleteExpired(ctx context.Context) error
}

type adminInviteRepository struct {
	collection *mongo.Collection
}

func NewAdminInviteRepository(db *database.MongoDB) AdminInviteRepository {
	return &adminInviteRepository{
		collection: db.Database.Collection(database.AdminInvitesCollection),
	}
}

func (r *adminInviteRepository) Create(ctx context.Context, invite *domain.AdminInvite) error {
	invite.ID = primitive.NewObjectID()
	_, err := r.collection.InsertOne(ctx, invite)
	return err
}

func (r *adminInviteRepository) GetByEmail(ctx context.Context, email string) (*domain.AdminInvite, error) {
	var invite domain.AdminInvite
	filter := bson.M{
		"invitee_email": email,
		"expires_at":    bson.M{"$gt": time.Now()},
	}

	err := r.collection.FindOne(ctx, filter).Decode(&invite)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &invite, nil
}

func (r *adminInviteRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *adminInviteRepository) DeleteExpired(ctx context.Context) error {
	filter := bson.M{"expires_at": bson.M{"$lte": time.Now()}}
	_, err := r.collection.DeleteMany(ctx, filter)
	return err
}

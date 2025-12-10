package repository

import (
	"backend/internal/database"
	"backend/internal/domain"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const defaultPosSettingsID = "default"

type PosSettingsRepository interface {
	Get(ctx context.Context) (*domain.PosSettings, error)
	UpsertMode(ctx context.Context, mode domain.PosFulfillmentMode) error
}

type posSettingsRepository struct {
	collection *mongo.Collection
}

func NewPosSettingsRepository(db *database.MongoDB) PosSettingsRepository {
	return &posSettingsRepository{collection: db.Database.Collection(database.PosSettingsCollection)}
}

func (r *posSettingsRepository) Get(ctx context.Context) (*domain.PosSettings, error) {
	var s domain.PosSettings
	err := r.collection.FindOne(ctx, bson.M{"_id": defaultPosSettingsID}).Decode(&s)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &domain.PosSettings{ID: defaultPosSettingsID, Mode: domain.PosModeQRCode}, nil
		}
		return nil, err
	}
	if s.ID == "" {
		s.ID = defaultPosSettingsID
	}
	if s.Mode == "" {
		s.Mode = domain.PosModeQRCode
	}
	return &s, nil
}

func (r *posSettingsRepository) UpsertMode(ctx context.Context, mode domain.PosFulfillmentMode) error {
	_, err := r.collection.UpdateByID(
		ctx,
		defaultPosSettingsID,
		bson.M{"$set": bson.M{"mode": mode, "updated_at": time.Now().UTC()}},
		options.UpdateOne().SetUpsert(true),
	)
	return err
}

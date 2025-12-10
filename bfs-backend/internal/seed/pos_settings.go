package seed

import (
	"backend/internal/database"
	"backend/internal/domain"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

type PosSettingsSeeder struct {
	Logger *zap.Logger
}

type PosSettingsDocument struct {
	ID        string                    `bson:"_id"`
	Mode      domain.PosFulfillmentMode `bson:"mode"`
	UpdatedAt time.Time                 `bson:"updated_at"`
}

func NewPosSettingsSeeder(logger *zap.Logger) PosSettingsSeeder {
	return PosSettingsSeeder{Logger: logger}
}

func (s PosSettingsSeeder) Name() string {
	return "pos_settings"
}

func (s PosSettingsSeeder) Seed(ctx context.Context, db *mongo.Database) error {
	logger := loggerOrNop(s.Logger)
	coll := db.Collection(database.PosSettingsCollection)

	now := time.Now().UTC()
	doc := PosSettingsDocument{
		ID:        "default",
		Mode:      domain.PosModeQRCode,
		UpdatedAt: now,
	}

	filter := bson.M{"_id": doc.ID}
	update := bson.M{
		"$setOnInsert": bson.M{"_id": doc.ID},
		"$set": bson.M{
			"mode":       doc.Mode,
			"updated_at": doc.UpdatedAt,
		},
	}
	opts := options.UpdateOne().SetUpsert(true)
	if _, err := coll.UpdateOne(ctx, filter, update, opts); err != nil {
		return fmt.Errorf("upsert pos settings: %w", err)
	}

	logger.Info("POS settings seeded", zap.String("mode", string(doc.Mode)))
	return nil
}

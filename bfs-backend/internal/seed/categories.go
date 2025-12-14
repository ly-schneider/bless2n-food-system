package seed

import (
	"backend/internal/database"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

type CategorySeeder struct {
	Logger *zap.Logger
}

type CategoryDocument struct {
	ID        bson.ObjectID `bson:"_id"`
	Name      string        `bson:"name"`
	IsActive  bool          `bson:"is_active"`
	Position  int           `bson:"position"`
	CreatedAt time.Time     `bson:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at"`
}

type categorySeed struct {
	Name     string
	Position int
}

var categorySeeds = []categorySeed{
	{Name: "Menus", Position: 0},
	{Name: "Burgers", Position: 1},
	{Name: "Beilagen", Position: 2},
	{Name: "Getränke", Position: 3},
}

func NewCategorySeeder(logger *zap.Logger) CategorySeeder {
	return CategorySeeder{Logger: logger}
}

func (s CategorySeeder) Name() string {
	return "categories"
}

func (s CategorySeeder) Seed(ctx context.Context, db *mongo.Database) error {
	logger := loggerOrNop(s.Logger)
	coll := db.Collection(database.CategoriesCollection)

	for _, seed := range categorySeeds {
		now := time.Now().UTC()
		doc := CategoryDocument{
			ID:        bson.NewObjectID(),
			Name:      seed.Name,
			IsActive:  true,
			Position:  seed.Position,
			CreatedAt: seededAt,
			UpdatedAt: now,
		}

		filter := bson.M{"name": seed.Name}
		update := bson.M{
			"$setOnInsert": bson.M{
				"_id":        doc.ID,
				"created_at": doc.CreatedAt,
			},
			"$set": bson.M{
				"name":       doc.Name,
				"is_active":  doc.IsActive,
				"position":   doc.Position,
				"updated_at": doc.UpdatedAt,
			},
		}

		opts := options.UpdateOne().SetUpsert(true)
		if _, err := coll.UpdateOne(ctx, filter, update, opts); err != nil {
			return fmt.Errorf("upsert category %s: %w", seed.Name, err)
		}
	}

	count, err := coll.CountDocuments(ctx, bson.D{})
	if err == nil {
		logger.Info("Categories seeded", zap.Int64("count", count))
	}

	return nil
}

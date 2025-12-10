package seed

import (
	"backend/internal/database"
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

type MenuSlotSeeder struct {
	Logger *zap.Logger
}

type MenuSlotDocument struct {
	ID        bson.ObjectID `bson:"_id"`
	ProductID bson.ObjectID `bson:"product_id"`
	Name      string        `bson:"name"`
	Sequence  int           `bson:"sequence"`
}

type menuSlotSeed struct {
	MenuName string
	Name     string
	Sequence int
}

var menuSlotSeeds = []menuSlotSeed{
	{MenuName: "Menu Gross", Name: "Burger", Sequence: 1},
	{MenuName: "Menu Gross", Name: "Beilage", Sequence: 2},
	{MenuName: "Menu Gross", Name: "Getränk", Sequence: 3},
	{MenuName: "Menu Klein", Name: "Burger", Sequence: 1},
	{MenuName: "Menu Klein", Name: "Getränk", Sequence: 2},
}

func NewMenuSlotSeeder(logger *zap.Logger) MenuSlotSeeder {
	return MenuSlotSeeder{Logger: logger}
}

func (s MenuSlotSeeder) Name() string {
	return "menu_slots"
}

func (s MenuSlotSeeder) Seed(ctx context.Context, db *mongo.Database) error {
	logger := loggerOrNop(s.Logger)
	coll := db.Collection(database.MenuSlotsCollection)

	menuNames := collectMenuNames()
	menuIDs, err := productIDsByName(ctx, db, menuNames)
	if err != nil {
		return err
	}
	for _, name := range menuNames {
		if _, ok := menuIDs[name]; !ok {
			return fmt.Errorf("menu product %s missing - seed products first", name)
		}
	}

	for _, seed := range menuSlotSeeds {
		productID := menuIDs[seed.MenuName]
		doc := MenuSlotDocument{
			ID:        bson.NewObjectID(),
			ProductID: productID,
			Name:      seed.Name,
			Sequence:  seed.Sequence,
		}

		filter := bson.M{
			"product_id": doc.ProductID,
			"name":       doc.Name,
		}
		update := bson.M{
			"$setOnInsert": bson.M{
				"_id": doc.ID,
			},
			"$set": bson.M{
				"product_id": doc.ProductID,
				"name":       doc.Name,
				"sequence":   doc.Sequence,
			},
		}

		opts := options.UpdateOne().SetUpsert(true)
		if _, err := coll.UpdateOne(ctx, filter, update, opts); err != nil {
			return fmt.Errorf("upsert menu slot %s/%s: %w", seed.MenuName, seed.Name, err)
		}
	}

	count, err := coll.CountDocuments(ctx, bson.D{})
	if err == nil {
		logger.Info("Menu slots seeded", zap.Int64("count", count))
	}
	return nil
}

func collectMenuNames() []string {
	seen := make(map[string]struct{})
	var names []string
	for _, seed := range menuSlotSeeds {
		if _, ok := seen[seed.MenuName]; ok {
			continue
		}
		seen[seed.MenuName] = struct{}{}
		names = append(names, seed.MenuName)
	}
	return names
}

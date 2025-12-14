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

type JetonSeeder struct {
	Logger *zap.Logger
}

type JetonDocument struct {
	ID           bson.ObjectID `bson:"_id"`
	Name         string        `bson:"name"`
	PaletteColor string        `bson:"palette_color"`
	HexColor     *string       `bson:"hex_color,omitempty"`
	CreatedAt    time.Time     `bson:"created_at"`
	UpdatedAt    time.Time     `bson:"updated_at"`
}

type jetonSeed struct {
	Name         string
	PaletteColor string
	HexColor     *string
}

var jetonSeeds = []jetonSeed{
	{Name: "Burger", PaletteColor: "blue"},
	{Name: "Getränk", PaletteColor: "red"},
	{Name: "Beilage", PaletteColor: "yellow"},
	{Name: "Menu", PaletteColor: "green"},
}

func NewJetonSeeder(logger *zap.Logger) JetonSeeder {
	return JetonSeeder{Logger: logger}
}

func (s JetonSeeder) Name() string {
	return "jetons"
}

func (s JetonSeeder) Seed(ctx context.Context, db *mongo.Database) error {
	logger := loggerOrNop(s.Logger)
	coll := db.Collection(database.JetonsCollection)

	for _, seed := range jetonSeeds {
		now := time.Now().UTC()
		doc := JetonDocument{
			ID:           bson.NewObjectID(),
			Name:         seed.Name,
			PaletteColor: seed.PaletteColor,
			HexColor:     seed.HexColor,
			CreatedAt:    seededAt,
			UpdatedAt:    now,
		}

		filter := bson.M{"name": seed.Name}
		update := bson.M{
			"$setOnInsert": bson.M{
				"_id":        doc.ID,
				"created_at": doc.CreatedAt,
			},
			"$set": bson.M{
				"name":          doc.Name,
				"palette_color": doc.PaletteColor,
				"hex_color":     doc.HexColor,
				"updated_at":    doc.UpdatedAt,
			},
		}
		opts := options.UpdateOne().SetUpsert(true)
		if _, err := coll.UpdateOne(ctx, filter, update, opts); err != nil {
			return fmt.Errorf("upsert jeton %s: %w", seed.Name, err)
		}
	}

	if err := s.assignJetonsToProducts(ctx, db, logger); err != nil {
		return err
	}

	count, err := coll.CountDocuments(ctx, bson.D{})
	if err == nil {
		logger.Info("Jetons seeded", zap.Int64("count", count))
	}
	return nil
}

func (s JetonSeeder) assignJetonsToProducts(ctx context.Context, db *mongo.Database, logger *zap.Logger) error {
	jetonNames := make([]string, 0, len(jetonSeeds))
	for _, seed := range jetonSeeds {
		jetonNames = append(jetonNames, seed.Name)
	}
	jetonIDs, err := jetonIDsByName(ctx, db, jetonNames)
	if err != nil {
		return err
	}

	categories, err := categoryNamesByID(ctx, db)
	if err != nil {
		return err
	}

	products := db.Collection(database.ProductsCollection)
	cur, err := products.Find(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("load products for jeton assignment: %w", err)
	}
	defer func() { _ = cur.Close(ctx) }()

	now := time.Now().UTC()
	var updated int
	for cur.Next(ctx) {
		var p struct {
			ID         bson.ObjectID      `bson:"_id"`
			CategoryID bson.ObjectID      `bson:"category_id"`
			Type       domain.ProductType `bson:"type"`
		}
		if err := cur.Decode(&p); err != nil {
			return fmt.Errorf("decode product for jeton assignment: %w", err)
		}

		categoryName := categories[p.CategoryID]
		jetonName := ""
		switch categoryName {
		case "Burgers":
			jetonName = "Burger"
		case "Beilagen":
			jetonName = "Beilage"
		case "Getränke":
			jetonName = "Getränk"
		case "Menus":
			jetonName = "Menu"
		}
		if p.Type == domain.ProductTypeMenu {
			jetonName = "Menu"
		}
		jetonID, ok := jetonIDs[jetonName]
		if !ok || jetonName == "" {
			continue
		}

		update := bson.M{
			"$set": bson.M{
				"jeton_id":   jetonID,
				"updated_at": now,
			},
		}
		if _, err := products.UpdateByID(ctx, p.ID, update); err != nil {
			return fmt.Errorf("assign jeton to product %s: %w", p.ID.Hex(), err)
		}
		updated++
	}
	if err := cur.Err(); err != nil {
		return err
	}

	logger.Info("Jetons assigned to products", zap.Int("productsUpdated", updated))
	return nil
}

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

type ProductSeeder struct {
	Logger *zap.Logger
}

type ProductDocument struct {
	ID         bson.ObjectID      `bson:"_id"`
	CategoryID bson.ObjectID      `bson:"category_id"`
	Type       domain.ProductType `bson:"type"`
	Name       string             `bson:"name"`
	Image      *string            `bson:"image,omitempty"`
	PriceCents domain.Cents       `bson:"price_cents"`
	JetonID    *bson.ObjectID     `bson:"jeton_id,omitempty"`
	IsActive   bool               `bson:"is_active"`
	CreatedAt  time.Time          `bson:"created_at"`
	UpdatedAt  time.Time          `bson:"updated_at"`
}

type productSeed struct {
	Name         string
	CategoryName string
	Type         domain.ProductType
	Image        string
	PriceCents   domain.Cents
	IsActive     bool
}

var productSeeds = []productSeed{
	{Name: "Smash Burger", CategoryName: "Burgers", Type: domain.ProductTypeSimple, Image: "/assets/images/products/bless2n-takeaway-smash-burger-16x9.png", PriceCents: 850, IsActive: true},
	{Name: "Veggie Burger", CategoryName: "Burgers", Type: domain.ProductTypeSimple, Image: "/assets/images/products/bless2n-takeaway-veggie-burger-16x9.png", PriceCents: 850, IsActive: true},
	{Name: "Pommes", CategoryName: "Beilagen", Type: domain.ProductTypeSimple, Image: "/assets/images/products/bless2n-takeaway-pommes-16x9.png", PriceCents: 400, IsActive: true},
	{Name: "Coca Cola", CategoryName: "Getränke", Type: domain.ProductTypeSimple, Image: "/assets/images/products/bless2n-takeaway-coca-cola-16x9.png", PriceCents: 250, IsActive: true},
	{Name: "Ice Tea Lemon", CategoryName: "Getränke", Type: domain.ProductTypeSimple, Image: "/assets/images/products/bless2n-takeaway-ice-tea-lemon-16x9.png", PriceCents: 250, IsActive: true},
	{Name: "Red Bull", CategoryName: "Getränke", Type: domain.ProductTypeSimple, Image: "/assets/images/products/bless2n-takeaway-red-bull-16x9.png", PriceCents: 250, IsActive: true},
	{Name: "El Tony Mate", CategoryName: "Getränke", Type: domain.ProductTypeSimple, Image: "/assets/images/products/bless2n-takeaway-el-tony-mate-16x9.png", PriceCents: 250, IsActive: true},
	{Name: "Wasser Prickelnd", CategoryName: "Getränke", Type: domain.ProductTypeSimple, Image: "/assets/images/products/bless2n-takeaway-wasser-prickelnd-16x9.png", PriceCents: 250, IsActive: true},
	{Name: "Menu Gross", CategoryName: "Menus", Type: domain.ProductTypeMenu, Image: "/assets/images/products/bless2n-takeaway-menu-2-gross-16x9.png", PriceCents: 1400, IsActive: true},
	{Name: "Menu Klein", CategoryName: "Menus", Type: domain.ProductTypeMenu, Image: "/assets/images/products/bless2n-takeaway-menu-1-klein-16x9.png", PriceCents: 1000, IsActive: true},
}

func NewProductSeeder(logger *zap.Logger) ProductSeeder {
	return ProductSeeder{Logger: logger}
}

func (s ProductSeeder) Name() string {
	return "products"
}

func (s ProductSeeder) Seed(ctx context.Context, db *mongo.Database) error {
	logger := loggerOrNop(s.Logger)
	coll := db.Collection(database.ProductsCollection)

	categoryNames := uniqueCategoryNames()
	categoryIDs, err := categoryIDsByName(ctx, db, categoryNames)
	if err != nil {
		return err
	}
	for _, name := range categoryNames {
		if _, ok := categoryIDs[name]; !ok {
			return fmt.Errorf("category %s missing - seed categories first", name)
		}
	}

	for _, seed := range productSeeds {
		now := time.Now().UTC()
		categoryID := categoryIDs[seed.CategoryName]

		doc := ProductDocument{
			ID:         bson.NewObjectID(),
			CategoryID: categoryID,
			Type:       seed.Type,
			Name:       seed.Name,
			Image:      ptr(seed.Image),
			PriceCents: seed.PriceCents,
			IsActive:   seed.IsActive,
			CreatedAt:  seededAt,
			UpdatedAt:  now,
		}

		filter := bson.M{"name": seed.Name}
		update := bson.M{
			"$setOnInsert": bson.M{
				"_id":        doc.ID,
				"created_at": doc.CreatedAt,
			},
			"$set": bson.M{
				"category_id": doc.CategoryID,
				"type":        doc.Type,
				"name":        doc.Name,
				"image":       doc.Image,
				"price_cents": doc.PriceCents,
				"is_active":   doc.IsActive,
				"updated_at":  doc.UpdatedAt,
			},
		}

		opts := options.UpdateOne().SetUpsert(true)
		if _, err := coll.UpdateOne(ctx, filter, update, opts); err != nil {
			return fmt.Errorf("upsert product %s: %w", seed.Name, err)
		}
	}

	count, err := coll.CountDocuments(ctx, bson.D{})
	if err == nil {
		logger.Info("Products seeded", zap.Int64("count", count))
	}
	return nil
}

func uniqueCategoryNames() []string {
	seen := make(map[string]struct{})
	var names []string
	for _, seed := range productSeeds {
		if _, ok := seen[seed.CategoryName]; ok {
			continue
		}
		seen[seed.CategoryName] = struct{}{}
		names = append(names, seed.CategoryName)
	}
	return names
}

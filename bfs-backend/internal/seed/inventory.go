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

type InventorySeeder struct {
	Logger     *zap.Logger
	OpeningQty int
}

type InventoryLedgerDocument struct {
	ID        bson.ObjectID          `bson:"_id"`
	ProductID bson.ObjectID          `bson:"product_id"`
	Delta     int                    `bson:"delta"`
	Reason    domain.InventoryReason `bson:"reason"`
	CreatedAt time.Time              `bson:"created_at"`
}

func NewInventorySeeder(logger *zap.Logger, openingQty int) InventorySeeder {
	return InventorySeeder{Logger: logger, OpeningQty: openingQty}
}

func (s InventorySeeder) Name() string {
	return "inventory_ledger"
}

func (s InventorySeeder) Seed(ctx context.Context, db *mongo.Database) error {
	if s.OpeningQty <= 0 {
		return nil
	}
	logger := loggerOrNop(s.Logger)

	products := db.Collection(database.ProductsCollection)
	cur, err := products.Find(ctx, bson.M{"type": domain.ProductTypeSimple})
	if err != nil {
		return fmt.Errorf("find simple products: %w", err)
	}
	defer func() { _ = cur.Close(ctx) }()

	ledger := db.Collection(database.InventoryLedgerCollection)
	var seeded int
	for cur.Next(ctx) {
		var p struct {
			ID bson.ObjectID `bson:"_id"`
		}
		if err := cur.Decode(&p); err != nil {
			return fmt.Errorf("decode product for inventory seeding: %w", err)
		}

		doc := InventoryLedgerDocument{
			ID:        bson.NewObjectID(),
			ProductID: p.ID,
			Delta:     s.OpeningQty,
			Reason:    domain.InventoryReasonOpeningBalance,
			CreatedAt: seededAt,
		}

		filter := bson.M{
			"product_id": doc.ProductID,
			"reason":     doc.Reason,
		}
		update := bson.M{
			"$setOnInsert": bson.M{
				"_id":        doc.ID,
				"created_at": doc.CreatedAt,
			},
			"$set": bson.M{
				"product_id": doc.ProductID,
				"delta":      doc.Delta,
				"reason":     doc.Reason,
			},
		}
		opts := options.UpdateOne().SetUpsert(true)
		if _, err := ledger.UpdateOne(ctx, filter, update, opts); err != nil {
			return fmt.Errorf("upsert inventory ledger for product %s: %w", p.ID.Hex(), err)
		}
		seeded++
	}
	if err := cur.Err(); err != nil {
		return err
	}

	logger.Info("Inventory ledger seeded", zap.Int("entries", seeded), zap.Int("openingQty", s.OpeningQty))
	return nil
}

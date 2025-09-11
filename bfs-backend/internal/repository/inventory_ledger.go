package repository

import (
	"context"
	"time"

	"backend/internal/database"
	"backend/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type InventoryLedgerRepository interface {
	Create(ctx context.Context, entry *domain.InventoryLedger) error
	GetByProductID(ctx context.Context, productID primitive.ObjectID, limit, offset int) ([]*domain.InventoryLedger, error)
	GetCurrentStock(ctx context.Context, productID primitive.ObjectID) (int, error)
	GetStockMovements(ctx context.Context, productID primitive.ObjectID, reason domain.InventoryReason, limit, offset int) ([]*domain.InventoryLedger, error)
}

type inventoryLedgerRepository struct {
	collection *mongo.Collection
}

func NewInventoryLedgerRepository(db *database.MongoDB) InventoryLedgerRepository {
	return &inventoryLedgerRepository{
		collection: db.Database.Collection(database.InventoryLedgerCollection),
	}
}

func (r *inventoryLedgerRepository) Create(ctx context.Context, entry *domain.InventoryLedger) error {
	entry.ID = primitive.NewObjectID()
	entry.Timestamp = time.Now()

	_, err := r.collection.InsertOne(ctx, entry)
	return err
}

func (r *inventoryLedgerRepository) GetByProductID(ctx context.Context, productID primitive.ObjectID, limit, offset int) ([]*domain.InventoryLedger, error) {
	filter := bson.M{"product_id": productID}
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "ts", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var entries []*domain.InventoryLedger
	for cursor.Next(ctx) {
		var entry domain.InventoryLedger
		if err := cursor.Decode(&entry); err != nil {
			return nil, err
		}
		entries = append(entries, &entry)
	}

	return entries, cursor.Err()
}

func (r *inventoryLedgerRepository) GetCurrentStock(ctx context.Context, productID primitive.ObjectID) (int, error) {
	pipeline := []bson.M{
		{"$match": bson.M{"product_id": productID}},
		{"$group": bson.M{
			"_id":   nil,
			"stock": bson.M{"$sum": "$delta"},
		}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	type result struct {
		Stock int `bson:"stock"`
	}

	var results []result
	if err := cursor.All(ctx, &results); err != nil {
		return 0, err
	}

	if len(results) == 0 {
		return 0, nil
	}

	return results[0].Stock, nil
}

func (r *inventoryLedgerRepository) GetStockMovements(ctx context.Context, productID primitive.ObjectID, reason domain.InventoryReason, limit, offset int) ([]*domain.InventoryLedger, error) {
	filter := bson.M{
		"product_id": productID,
		"reason":     reason,
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "ts", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var entries []*domain.InventoryLedger
	for cursor.Next(ctx) {
		var entry domain.InventoryLedger
		if err := cursor.Decode(&entry); err != nil {
			return nil, err
		}
		entries = append(entries, &entry)
	}

	return entries, cursor.Err()
}

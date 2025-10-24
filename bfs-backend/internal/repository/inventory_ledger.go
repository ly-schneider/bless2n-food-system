package repository

import (
	"backend/internal/database"
	"backend/internal/domain"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type InventoryLedgerRepository interface {
	Append(ctx context.Context, entry *domain.InventoryLedger) error
	AppendMany(ctx context.Context, entries []*domain.InventoryLedger) error
	// SumByProductIDs returns the total stock per product by summing deltas
	SumByProductIDs(ctx context.Context, ids []primitive.ObjectID) (map[primitive.ObjectID]int64, error)
}

type inventoryLedgerRepository struct {
	collection *mongo.Collection
}

func NewInventoryLedgerRepository(db *database.MongoDB) InventoryLedgerRepository {
	return &inventoryLedgerRepository{
		collection: db.Database.Collection(database.InventoryLedgerCollection),
	}
}

func (r *inventoryLedgerRepository) Append(ctx context.Context, entry *domain.InventoryLedger) error {
	if entry == nil {
		return nil
	}
	if entry.ID.IsZero() {
		entry.ID = primitive.NewObjectID()
	}
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now().UTC()
	}
	_, err := r.collection.InsertOne(ctx, entry)
	return err
}

func (r *inventoryLedgerRepository) AppendMany(ctx context.Context, entries []*domain.InventoryLedger) error {
	if len(entries) == 0 {
		return nil
	}
	docs := make([]interface{}, 0, len(entries))
	now := time.Now().UTC()
	for _, e := range entries {
		if e == nil {
			continue
		}
		if e.ID.IsZero() {
			e.ID = primitive.NewObjectID()
		}
		if e.CreatedAt.IsZero() {
			e.CreatedAt = now
		}
		docs = append(docs, e)
	}
	if len(docs) == 0 {
		return nil
	}
	_, err := r.collection.InsertMany(ctx, docs)
	return err
}

func (r *inventoryLedgerRepository) SumByProductIDs(ctx context.Context, ids []primitive.ObjectID) (map[primitive.ObjectID]int64, error) {
	result := map[primitive.ObjectID]int64{}
	if len(ids) == 0 {
		return result, nil
	}
	// Aggregate sum of deltas grouped by product_id
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{"product_id": bson.M{"$in": ids}}}},
		{{Key: "$group", Value: bson.M{"_id": "$product_id", "total": bson.M{"$sum": "$delta"}}}},
	}
	cur, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cur.Close(ctx) }()
	for cur.Next(ctx) {
		var row struct {
			ID    primitive.ObjectID `bson:"_id"`
			Total int64              `bson:"total"`
		}
		if err := cur.Decode(&row); err != nil {
			return nil, err
		}
		result[row.ID] = row.Total
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

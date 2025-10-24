package repository

import (
	"backend/internal/database"
	"backend/internal/domain"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type OrderItemRepository interface {
	InsertMany(ctx context.Context, items []*domain.OrderItem) error
	DeleteByOrderID(ctx context.Context, id bson.ObjectID) error
	// FindByOrderID returns all items for a given order id
	FindByOrderID(ctx context.Context, id bson.ObjectID) ([]*domain.OrderItem, error)
	// FindByFilter returns items by arbitrary filter (internal use)
	FindByFilter(ctx context.Context, filter any) ([]*domain.OrderItem, error)
	// UpdateRedeemForOrderByProductIDs marks items redeemed in a single conditional write
	UpdateRedeemForOrderByProductIDs(ctx context.Context, orderID bson.ObjectID, productIDs []bson.ObjectID, redeemedAt time.Time) (matched int64, modified int64, err error)
}

type orderItemRepository struct {
	collection *mongo.Collection
}

func NewOrderItemRepository(db *database.MongoDB) OrderItemRepository {
	return &orderItemRepository{
		collection: db.Database.Collection(database.OrderItemsCollection),
	}
}

func (r *orderItemRepository) InsertMany(ctx context.Context, items []*domain.OrderItem) error {
	docs := make([]interface{}, 0, len(items))
	for _, it := range items {
		docs = append(docs, it)
	}
	if len(docs) == 0 {
		return nil
	}
	_, err := r.collection.InsertMany(ctx, docs)
	return err
}

func (r *orderItemRepository) DeleteByOrderID(ctx context.Context, id bson.ObjectID) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"order_id": id})
	return err
}

func (r *orderItemRepository) FindByOrderID(ctx context.Context, id bson.ObjectID) ([]*domain.OrderItem, error) {
	cur, err := r.collection.Find(ctx, bson.M{"order_id": id})
	if err != nil {
		return nil, err
	}
	defer func() { _ = cur.Close(ctx) }()

	var items []*domain.OrderItem
	for cur.Next(ctx) {
		var it domain.OrderItem
		if err := cur.Decode(&it); err != nil {
			return nil, err
		}
		items = append(items, &it)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *orderItemRepository) FindByFilter(ctx context.Context, filter any) ([]*domain.OrderItem, error) {
	cur, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cur.Close(ctx) }()
	var items []*domain.OrderItem
	for cur.Next(ctx) {
		var it domain.OrderItem
		if err := cur.Decode(&it); err != nil {
			return nil, err
		}
		items = append(items, &it)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *orderItemRepository) UpdateRedeemForOrderByProductIDs(ctx context.Context, orderID bson.ObjectID, productIDs []bson.ObjectID, redeemedAt time.Time) (int64, int64, error) {
	if len(productIDs) == 0 {
		return 0, 0, nil
	}
	filter := bson.M{"order_id": orderID, "product_id": bson.M{"$in": productIDs}, "is_redeemed": false}
	update := bson.M{"$set": bson.M{"is_redeemed": true, "redeemed_at": redeemedAt}}
	res, err := r.collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, 0, err
	}
	return res.MatchedCount, res.ModifiedCount, nil
}

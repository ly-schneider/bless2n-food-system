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

type OrderItemRepository interface {
	Create(ctx context.Context, item *domain.OrderItem) error
	CreateBatch(ctx context.Context, items []*domain.OrderItem) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*domain.OrderItem, error)
	GetByOrderID(ctx context.Context, orderID primitive.ObjectID) ([]*domain.OrderItem, error)
	GetUnredeemedByOrderID(ctx context.Context, orderID primitive.ObjectID) ([]*domain.OrderItem, error)
	GetByProductID(ctx context.Context, productID primitive.ObjectID, unredeemedOnly bool) ([]*domain.OrderItem, error)
	GetByParentItemID(ctx context.Context, parentItemID primitive.ObjectID) ([]*domain.OrderItem, error)
	Update(ctx context.Context, item *domain.OrderItem) error
	MarkAsRedeemed(ctx context.Context, id primitive.ObjectID, stationID, deviceID primitive.ObjectID) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	DeleteByOrderID(ctx context.Context, orderID primitive.ObjectID) error
	GetRedeemedByStation(ctx context.Context, stationID primitive.ObjectID, limit, offset int) ([]*domain.OrderItem, error)
	GetRedeemedByDevice(ctx context.Context, deviceID primitive.ObjectID, limit, offset int) ([]*domain.OrderItem, error)
}

type orderItemRepository struct {
	collection *mongo.Collection
}

func NewOrderItemRepository(db *database.MongoDB) OrderItemRepository {
	return &orderItemRepository{
		collection: db.Database.Collection(database.OrderItemsCollection),
	}
}

func (r *orderItemRepository) Create(ctx context.Context, item *domain.OrderItem) error {
	item.ID = primitive.NewObjectID()
	item.IsRedeemed = false

	_, err := r.collection.InsertOne(ctx, item)
	return err
}

func (r *orderItemRepository) CreateBatch(ctx context.Context, items []*domain.OrderItem) error {
	if len(items) == 0 {
		return nil
	}

	docs := make([]interface{}, len(items))
	for i, item := range items {
		item.ID = primitive.NewObjectID()
		item.IsRedeemed = false
		docs[i] = item
	}

	_, err := r.collection.InsertMany(ctx, docs)
	return err
}

func (r *orderItemRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.OrderItem, error) {
	var item domain.OrderItem
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&item)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *orderItemRepository) GetByOrderID(ctx context.Context, orderID primitive.ObjectID) ([]*domain.OrderItem, error) {
	filter := bson.M{"order_id": orderID}
	opts := options.Find().SetSort(bson.D{{Key: "parent_item_id", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []*domain.OrderItem
	for cursor.Next(ctx) {
		var item domain.OrderItem
		if err := cursor.Decode(&item); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}

	return items, cursor.Err()
}

func (r *orderItemRepository) GetUnredeemedByOrderID(ctx context.Context, orderID primitive.ObjectID) ([]*domain.OrderItem, error) {
	filter := bson.M{
		"order_id":    orderID,
		"is_redeemed": false,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []*domain.OrderItem
	for cursor.Next(ctx) {
		var item domain.OrderItem
		if err := cursor.Decode(&item); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}

	return items, cursor.Err()
}

func (r *orderItemRepository) GetByProductID(ctx context.Context, productID primitive.ObjectID, unredeemedOnly bool) ([]*domain.OrderItem, error) {
	filter := bson.M{"product_id": productID}
	if unredeemedOnly {
		filter["is_redeemed"] = false
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []*domain.OrderItem
	for cursor.Next(ctx) {
		var item domain.OrderItem
		if err := cursor.Decode(&item); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}

	return items, cursor.Err()
}

func (r *orderItemRepository) GetByParentItemID(ctx context.Context, parentItemID primitive.ObjectID) ([]*domain.OrderItem, error) {
	filter := bson.M{"parent_item_id": parentItemID}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []*domain.OrderItem
	for cursor.Next(ctx) {
		var item domain.OrderItem
		if err := cursor.Decode(&item); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}

	return items, cursor.Err()
}

func (r *orderItemRepository) Update(ctx context.Context, item *domain.OrderItem) error {
	filter := bson.M{"_id": item.ID}
	update := bson.M{"$set": item}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *orderItemRepository) MarkAsRedeemed(ctx context.Context, id primitive.ObjectID, stationID, deviceID primitive.ObjectID) error {
	now := time.Now()
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"is_redeemed":         true,
			"redeemed_at":         now,
			"redeemed_station_id": stationID,
			"redeemed_device_id":  deviceID,
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *orderItemRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *orderItemRepository) DeleteByOrderID(ctx context.Context, orderID primitive.ObjectID) error {
	filter := bson.M{"order_id": orderID}
	_, err := r.collection.DeleteMany(ctx, filter)
	return err
}

func (r *orderItemRepository) GetRedeemedByStation(ctx context.Context, stationID primitive.ObjectID, limit, offset int) ([]*domain.OrderItem, error) {
	filter := bson.M{
		"redeemed_station_id": stationID,
		"is_redeemed":         true,
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "redeemed_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []*domain.OrderItem
	for cursor.Next(ctx) {
		var item domain.OrderItem
		if err := cursor.Decode(&item); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}

	return items, cursor.Err()
}

func (r *orderItemRepository) GetRedeemedByDevice(ctx context.Context, deviceID primitive.ObjectID, limit, offset int) ([]*domain.OrderItem, error) {
	filter := bson.M{
		"redeemed_device_id": deviceID,
		"is_redeemed":        true,
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "redeemed_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var items []*domain.OrderItem
	for cursor.Next(ctx) {
		var item domain.OrderItem
		if err := cursor.Decode(&item); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}

	return items, cursor.Err()
}

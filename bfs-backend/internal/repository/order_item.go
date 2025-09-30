package repository

import (
    "backend/internal/database"
    "backend/internal/domain"
    "context"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
)

type OrderItemRepository interface {
    InsertMany(ctx context.Context, items []*domain.OrderItem) error
    DeleteByOrderID(ctx context.Context, id primitive.ObjectID) error
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

func (r *orderItemRepository) DeleteByOrderID(ctx context.Context, id primitive.ObjectID) error {
    _, err := r.collection.DeleteMany(ctx, bson.M{"order_id": id})
    return err
}

package repository

import (
    "backend/internal/database"
    "backend/internal/domain"
    "context"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type IdempotencyRepository interface {
    Get(ctx context.Context, scope, key string) (*domain.IdempotencyRecord, error)
    SaveIfAbsent(ctx context.Context, rec *domain.IdempotencyRecord) (bool, error)
}

type idempotencyRepository struct {
    collection *mongo.Collection
}

func NewIdempotencyRepository(db *database.MongoDB) IdempotencyRepository {
    coll := db.Database.Collection("idempotency_keys")
    // Unique index on (scope, key)
    _, _ = coll.Indexes().CreateOne(context.Background(), mongo.IndexModel{
        Keys: bson.D{{Key: "scope", Value: 1}, {Key: "key", Value: 1}},
        Options: options.Index().SetUnique(true),
    })
    return &idempotencyRepository{collection: coll}
}

func (r *idempotencyRepository) Get(ctx context.Context, scope, key string) (*domain.IdempotencyRecord, error) {
    var rec domain.IdempotencyRecord
    err := r.collection.FindOne(ctx, bson.M{"scope": scope, "key": key}).Decode(&rec)
    if err != nil { return nil, err }
    return &rec, nil
}

func (r *idempotencyRepository) SaveIfAbsent(ctx context.Context, rec *domain.IdempotencyRecord) (bool, error) {
    if rec == nil { return false, nil }
    if rec.CreatedAt.IsZero() { rec.CreatedAt = time.Now().UTC() }
    _, err := r.collection.InsertOne(ctx, rec)
    if err != nil {
        if we, ok := err.(mongo.WriteException); ok {
            for _, e := range we.WriteErrors { if e.Code == 11000 { return false, nil } }
        }
        return false, err
    }
    return true, nil
}


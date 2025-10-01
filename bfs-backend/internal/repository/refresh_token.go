package repository

import (
    "backend/internal/database"
    "backend/internal/domain"
    "context"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
)

type RefreshTokenRepository interface {
    Create(ctx context.Context, t *domain.RefreshToken) (*domain.RefreshToken, error)
    FindByHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
    MarkUsed(ctx context.Context, id any, usedAt time.Time) error
    RevokeFamily(ctx context.Context, familyID string, reason string) error
    ListActiveByUser(ctx context.Context, userID primitive.ObjectID) ([]domain.RefreshToken, error)
    RevokeAllByUser(ctx context.Context, userID primitive.ObjectID, reason string) error
}

type refreshTokenRepository struct {
    collection *mongo.Collection
}

func NewRefreshTokenRepository(db *database.MongoDB) RefreshTokenRepository {
    return &refreshTokenRepository{
        collection: db.Database.Collection(database.RefreshTokensCollection),
    }
}

func (r *refreshTokenRepository) Create(ctx context.Context, t *domain.RefreshToken) (*domain.RefreshToken, error) {
    if t.ID.IsZero() {
        t.ID = primitive.NewObjectID()
    }
    if _, err := r.collection.InsertOne(ctx, t); err != nil {
        return nil, err
    }
    return t, nil
}

func (r *refreshTokenRepository) FindByHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
    var t domain.RefreshToken
    if err := r.collection.FindOne(ctx, bson.M{"token_hash": tokenHash}).Decode(&t); err != nil {
        return nil, err
    }
    return &t, nil
}

func (r *refreshTokenRepository) MarkUsed(ctx context.Context, id any, usedAt time.Time) error {
    _, err := r.collection.UpdateByID(ctx, id, bson.M{"$set": bson.M{"last_used_at": usedAt}})
    return err
}

func (r *refreshTokenRepository) RevokeFamily(ctx context.Context, familyID string, reason string) error {
    _, err := r.collection.UpdateMany(ctx, bson.M{"family_id": familyID, "is_revoked": false}, bson.M{"$set": bson.M{"is_revoked": true, "revoked_reason": reason}})
    return err
}

func (r *refreshTokenRepository) ListActiveByUser(ctx context.Context, userID primitive.ObjectID) ([]domain.RefreshToken, error) {
    now := time.Now().UTC()
    cur, err := r.collection.Find(ctx, bson.M{
        "user_id":    userID,
        "is_revoked": false,
        "expires_at": bson.M{"$gt": now},
    })
    if err != nil {
        return nil, err
    }
    defer func() { _ = cur.Close(ctx) }()
    var out []domain.RefreshToken
    for cur.Next(ctx) {
        var t domain.RefreshToken
        if err := cur.Decode(&t); err != nil { return nil, err }
        out = append(out, t)
    }
    return out, nil
}

func (r *refreshTokenRepository) RevokeAllByUser(ctx context.Context, userID primitive.ObjectID, reason string) error {
    _, err := r.collection.UpdateMany(ctx, bson.M{"user_id": userID, "is_revoked": false}, bson.M{"$set": bson.M{"is_revoked": true, "revoked_reason": reason}})
    return err
}

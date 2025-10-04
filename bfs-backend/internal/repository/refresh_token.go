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
    // ListActiveFamiliesAll returns active session families grouped by user and family id
    ListActiveFamiliesAll(ctx context.Context, limit, offset int) ([]SessionFamily, int64, error)
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

type SessionFamily struct {
    UserID     primitive.ObjectID `bson:"user_id" json:"userId"`
    FamilyID   string             `bson:"family_id" json:"familyId"`
    Device     string             `bson:"device" json:"device"`
    CreatedAt  time.Time          `bson:"created_at" json:"createdAt"`
    LastUsedAt time.Time          `bson:"last_used_at" json:"lastUsedAt"`
}

func (r *refreshTokenRepository) ListActiveFamiliesAll(ctx context.Context, limit, offset int) ([]SessionFamily, int64, error) {
    now := time.Now().UTC()
    match := bson.D{{Key: "$match", Value: bson.M{"is_revoked": false, "expires_at": bson.M{"$gt": now}}}}
    group := bson.D{{Key: "$group", Value: bson.M{
        "_id": bson.M{"user_id": "$user_id", "family_id": "$family_id"},
        "device": bson.M{"$first": "$client_id"},
        "created_at": bson.M{"$min": "$issued_at"},
        "last_used_at": bson.M{"$max": "$last_used_at"},
        "user_id": bson.M{"$first": "$user_id"},
        "family_id": bson.M{"$first": "$family_id"},
    }}}
    sort := bson.D{{Key: "$sort", Value: bson.M{"last_used_at": -1}}}
    skip := bson.D{}
    if offset > 0 { skip = bson.D{{Key: "$skip", Value: offset}} }
    limitStage := bson.D{}
    if limit > 0 { limitStage = bson.D{{Key: "$limit", Value: limit}} }

    pipeline := mongo.Pipeline{match, group, sort}
    if len(skip) > 0 { pipeline = append(pipeline, skip) }
    if len(limitStage) > 0 { pipeline = append(pipeline, limitStage) }

    cur, err := r.collection.Aggregate(ctx, pipeline)
    if err != nil { return nil, 0, err }
    defer func() { _ = cur.Close(ctx) }()
    var rows []SessionFamily
    for cur.Next(ctx) {
        var row SessionFamily
        if err := cur.Decode(&row); err != nil { return nil, 0, err }
        rows = append(rows, row)
    }
    if err := cur.Err(); err != nil { return nil, 0, err }

    // Count total groups
    countPipe := mongo.Pipeline{match, group, bson.D{{Key: "$count", Value: "total"}}}
    ccur, err := r.collection.Aggregate(ctx, countPipe)
    if err != nil { return rows, 0, nil }
    defer func() { _ = ccur.Close(ctx) }()
    var total int64
    if ccur.Next(ctx) {
        var doc struct{ Total int64 `bson:"total"` }
        if derr := ccur.Decode(&doc); derr == nil { total = doc.Total }
    }
    return rows, total, nil
}

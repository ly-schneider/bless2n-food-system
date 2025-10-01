package repository

import (
    "backend/internal/database"
    "backend/internal/domain"
    "context"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type EmailChangeTokenRepository interface {
    Create(ctx context.Context, t *domain.EmailChangeToken) (*domain.EmailChangeToken, error)
    CreateWithCode(ctx context.Context, userID primitive.ObjectID, newEmail, codeHash string, expiresAt time.Time) (*domain.EmailChangeToken, error)
    FindActiveByUser(ctx context.Context, userID primitive.ObjectID) ([]domain.EmailChangeToken, error)
    MarkUsed(ctx context.Context, id primitive.ObjectID, usedAt time.Time) error
    IncrementAttempts(ctx context.Context, id primitive.ObjectID) (int, error)
    DeleteByUser(ctx context.Context, userID primitive.ObjectID) error
}

type emailChangeTokenRepository struct {
    collection *mongo.Collection
}

func NewEmailChangeTokenRepository(db *database.MongoDB) EmailChangeTokenRepository {
    return &emailChangeTokenRepository{collection: db.Database.Collection(database.EmailChangeTokensCollection)}
}

func (r *emailChangeTokenRepository) Create(ctx context.Context, t *domain.EmailChangeToken) (*domain.EmailChangeToken, error) {
    if t.ID.IsZero() { t.ID = primitive.NewObjectID() }
    t.CreatedAt = time.Now().UTC()
    if _, err := r.collection.InsertOne(ctx, t); err != nil { return nil, err }
    return t, nil
}

func (r *emailChangeTokenRepository) CreateWithCode(ctx context.Context, userID primitive.ObjectID, newEmail, codeHash string, expiresAt time.Time) (*domain.EmailChangeToken, error) {
    t := &domain.EmailChangeToken{
        ID:        primitive.NewObjectID(),
        UserID:    userID,
        NewEmail:  newEmail,
        TokenHash: codeHash,
        CreatedAt: time.Now().UTC(),
        Attempts:  0,
        ExpiresAt: expiresAt,
    }
    if _, err := r.collection.InsertOne(ctx, t); err != nil { return nil, err }
    return t, nil
}

func (r *emailChangeTokenRepository) FindActiveByUser(ctx context.Context, userID primitive.ObjectID) ([]domain.EmailChangeToken, error) {
    now := time.Now().UTC()
    cur, err := r.collection.Find(ctx, bson.M{
        "user_id":    userID,
        "used_at":    bson.M{"$exists": false},
        "expires_at": bson.M{"$gt": now},
    }, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
    if err != nil { return nil, err }
    defer func(){ _ = cur.Close(ctx) }()
    var out []domain.EmailChangeToken
    for cur.Next(ctx) {
        var t domain.EmailChangeToken
        if err := cur.Decode(&t); err != nil { return nil, err }
        out = append(out, t)
    }
    return out, nil
}

func (r *emailChangeTokenRepository) MarkUsed(ctx context.Context, id primitive.ObjectID, usedAt time.Time) error {
    _, err := r.collection.UpdateByID(ctx, id, bson.M{"$set": bson.M{"used_at": usedAt}})
    return err
}

func (r *emailChangeTokenRepository) IncrementAttempts(ctx context.Context, id primitive.ObjectID) (int, error) {
    res := r.collection.FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$inc": bson.M{"attempts": 1}}, options.FindOneAndUpdate().SetReturnDocument(options.After))
    var t domain.EmailChangeToken
    if err := res.Decode(&t); err != nil { return 0, err }
    return t.Attempts, nil
}

func (r *emailChangeTokenRepository) DeleteByUser(ctx context.Context, userID primitive.ObjectID) error {
    _, err := r.collection.DeleteMany(ctx, bson.M{"user_id": userID})
    return err
}

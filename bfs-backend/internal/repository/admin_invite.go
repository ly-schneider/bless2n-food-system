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

type AdminInviteRepository interface {
    Create(ctx context.Context, invitedBy primitive.ObjectID, email string, tokenHash string, expiresAt time.Time) (*domain.AdminInvite, error)
    FindByID(ctx context.Context, id primitive.ObjectID) (*domain.AdminInvite, error)
    FindByTokenHash(ctx context.Context, tokenHash string) (*domain.AdminInvite, error)
    List(ctx context.Context, status *string, email *string, limit, offset int) ([]*domain.AdminInvite, int64, error)
    Revoke(ctx context.Context, id primitive.ObjectID) (bool, error)
    MarkAccepted(ctx context.Context, id primitive.ObjectID) error
    UpdateToken(ctx context.Context, id primitive.ObjectID, tokenHash string, expiresAt time.Time) error
    Delete(ctx context.Context, id primitive.ObjectID) error
}

type adminInviteRepository struct {
    collection *mongo.Collection
}

func NewAdminInviteRepository(db *database.MongoDB) AdminInviteRepository {
    coll := db.Database.Collection(database.AdminInvitesCollection)
    // Helpful indexes: token_hash unique; status; invitee_email
    _, _ = coll.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
        { Keys: bson.D{{Key: "token_hash", Value: 1}}, Options: options.Index().SetUnique(true) },
        { Keys: bson.D{{Key: "status", Value: 1}} },
        { Keys: bson.D{{Key: "invitee_email", Value: 1}} },
        { Keys: bson.D{{Key: "expires_at", Value: 1}} },
    })
    return &adminInviteRepository{ collection: coll }
}

func (r *adminInviteRepository) Create(ctx context.Context, invitedBy primitive.ObjectID, email string, tokenHash string, expiresAt time.Time) (*domain.AdminInvite, error) {
    now := time.Now().UTC()
    inv := &domain.AdminInvite{
        ID:           primitive.NewObjectID(),
        InvitedBy:    invitedBy,
        InviteeEmail: email,
        TokenHash:    tokenHash,
        ExpiresAt:    expiresAt,
        Status:       "pending",
        CreatedAt:    now,
    }
    if _, err := r.collection.InsertOne(ctx, inv); err != nil { return nil, err }
    return inv, nil
}

func (r *adminInviteRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.AdminInvite, error) {
    var out domain.AdminInvite
    if err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&out); err != nil { return nil, err }
    return &out, nil
}

func (r *adminInviteRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*domain.AdminInvite, error) {
    var out domain.AdminInvite
    if err := r.collection.FindOne(ctx, bson.M{"token_hash": tokenHash}).Decode(&out); err != nil { return nil, err }
    return &out, nil
}

func (r *adminInviteRepository) List(ctx context.Context, status *string, email *string, limit, offset int) ([]*domain.AdminInvite, int64, error) {
    filter := bson.M{}
    if status != nil && *status != "" { filter["status"] = *status }
    if email != nil && *email != "" { filter["invitee_email"] = *email }
    total, err := r.collection.CountDocuments(ctx, filter)
    if err != nil { return nil, 0, err }
    opts := options.Find().SetSort(bson.M{"created_at": -1})
    if limit > 0 { opts.SetLimit(int64(limit)) }
    if offset > 0 { opts.SetSkip(int64(offset)) }
    cur, err := r.collection.Find(ctx, filter, opts)
    if err != nil { return nil, 0, err }
    defer func() { _ = cur.Close(ctx) }()
    var out []*domain.AdminInvite
    for cur.Next(ctx) {
        var it domain.AdminInvite
        if err := cur.Decode(&it); err != nil { return nil, 0, err }
        out = append(out, &it)
    }
    if err := cur.Err(); err != nil { return nil, 0, err }
    return out, total, nil
}

func (r *adminInviteRepository) Revoke(ctx context.Context, id primitive.ObjectID) (bool, error) {
    res, err := r.collection.UpdateOne(ctx, bson.M{"_id": id, "status": "pending"}, bson.M{"$set": bson.M{"status": "revoked"}})
    if err != nil { return false, err }
    return res.ModifiedCount > 0, nil
}

func (r *adminInviteRepository) MarkAccepted(ctx context.Context, id primitive.ObjectID) error {
    now := time.Now().UTC()
    _, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"status": "accepted", "used_at": now}})
    return err
}

func (r *adminInviteRepository) UpdateToken(ctx context.Context, id primitive.ObjectID, tokenHash string, expiresAt time.Time) error {
    _, err := r.collection.UpdateByID(ctx, id, bson.M{"$set": bson.M{"token_hash": tokenHash, "expires_at": expiresAt, "status": "pending"}})
    return err
}

func (r *adminInviteRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
    _, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
    return err
}

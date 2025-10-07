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

type StationRequestRepository interface {
    Create(ctx context.Context, req *domain.StationRequest) error
    FindPendingByDeviceKey(ctx context.Context, deviceKey string) (*domain.StationRequest, error)
    FindByID(ctx context.Context, id primitive.ObjectID) (*domain.StationRequest, error)
    List(ctx context.Context, status *domain.StationRequestStatus) ([]*domain.StationRequest, error)
    UpdateStatus(ctx context.Context, id primitive.ObjectID, status domain.StationRequestStatus, decidedBy *primitive.ObjectID, decidedAt time.Time) error
}

type stationRequestRepository struct {
	collection *mongo.Collection
}

func NewStationRequestRepository(db *database.MongoDB) StationRequestRepository {
    return &stationRequestRepository{
        collection: db.Database.Collection(database.StationRequestsCollection),
    }
}

func (r *stationRequestRepository) Create(ctx context.Context, req *domain.StationRequest) error {
    if req == nil { return nil }
    if req.ID.IsZero() { req.ID = primitive.NewObjectID() }
    now := time.Now().UTC()
    if req.CreatedAt.IsZero() { req.CreatedAt = now }
    if req.ExpiresAt.IsZero() { req.ExpiresAt = now.Add(30 * 24 * time.Hour) }
    _, err := r.collection.InsertOne(ctx, req)
    return err
}

func (r *stationRequestRepository) FindPendingByDeviceKey(ctx context.Context, deviceKey string) (*domain.StationRequest, error) {
    var req domain.StationRequest
    if err := r.collection.FindOne(ctx, bson.M{"device_key": deviceKey, "status": domain.StationRequestStatusPending}).Decode(&req); err != nil { return nil, err }
    return &req, nil
}

func (r *stationRequestRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*domain.StationRequest, error) {
    var req domain.StationRequest
    if err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&req); err != nil { return nil, err }
    return &req, nil
}

func (r *stationRequestRepository) List(ctx context.Context, status *domain.StationRequestStatus) ([]*domain.StationRequest, error) {
    filter := bson.M{}
    if status != nil && *status != "" { filter["status"] = *status }
    cur, err := r.collection.Find(ctx, filter)
    if err != nil { return nil, err }
    defer func(){ _ = cur.Close(ctx) }()
    var out []*domain.StationRequest
    for cur.Next(ctx) {
        var it domain.StationRequest
        if err := cur.Decode(&it); err != nil { return nil, err }
        out = append(out, &it)
    }
    if err := cur.Err(); err != nil { return nil, err }
    return out, nil
}

func (r *stationRequestRepository) UpdateStatus(ctx context.Context, id primitive.ObjectID, status domain.StationRequestStatus, decidedBy *primitive.ObjectID, decidedAt time.Time) error {
    set := bson.M{ "status": status, "updated_at": time.Now().UTC(), "decided_at": decidedAt }
    if decidedBy != nil { set["decided_by"] = *decidedBy }
    _, err := r.collection.UpdateByID(ctx, id, bson.M{"$set": set})
    return err
}

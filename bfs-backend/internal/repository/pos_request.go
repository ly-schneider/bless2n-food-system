package repository

import (
	"backend/internal/database"
	"backend/internal/domain"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type PosRequestRepository interface {
	Create(ctx context.Context, req *domain.PosRequest) error
	FindPendingByToken(ctx context.Context, token string) (*domain.PosRequest, error)
	FindByID(ctx context.Context, id bson.ObjectID) (*domain.PosRequest, error)
	List(ctx context.Context, status *domain.PosRequestStatus) ([]*domain.PosRequest, error)
	UpdateStatus(ctx context.Context, id bson.ObjectID, status domain.PosRequestStatus, decidedBy *bson.ObjectID, decidedAt time.Time) error
}

type posRequestRepository struct{ collection *mongo.Collection }

func NewPosRequestRepository(db *database.MongoDB) PosRequestRepository {
	return &posRequestRepository{collection: db.Database.Collection(database.PosRequestsCollection)}
}

func (r *posRequestRepository) Create(ctx context.Context, req *domain.PosRequest) error {
	if req == nil {
		return nil
	}
	if req.ID.IsZero() {
		req.ID = bson.NewObjectID()
	}
	now := time.Now().UTC()
	if req.CreatedAt.IsZero() {
		req.CreatedAt = now
	}
	if req.ExpiresAt.IsZero() {
		req.ExpiresAt = now.Add(30 * 24 * time.Hour)
	}
	_, err := r.collection.InsertOne(ctx, req)
	return err
}

func (r *posRequestRepository) FindPendingByToken(ctx context.Context, token string) (*domain.PosRequest, error) {
	var req domain.PosRequest
	if err := r.collection.FindOne(ctx, bson.M{"device_token": token, "status": domain.PosRequestStatusPending}).Decode(&req); err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *posRequestRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.PosRequest, error) {
	var req domain.PosRequest
	if err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&req); err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *posRequestRepository) List(ctx context.Context, status *domain.PosRequestStatus) ([]*domain.PosRequest, error) {
	filter := bson.M{}
	if status != nil && *status != "" {
		filter["status"] = *status
	}
	cur, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cur.Close(ctx) }()
	out := make([]*domain.PosRequest, 0)
	for cur.Next(ctx) {
		var it domain.PosRequest
		if err := cur.Decode(&it); err != nil {
			return nil, err
		}
		out = append(out, &it)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *posRequestRepository) UpdateStatus(ctx context.Context, id bson.ObjectID, status domain.PosRequestStatus, decidedBy *bson.ObjectID, decidedAt time.Time) error {
	set := bson.M{"status": status, "updated_at": time.Now().UTC(), "decided_at": decidedAt}
	if decidedBy != nil {
		set["decided_by"] = *decidedBy
	}
	_, err := r.collection.UpdateByID(ctx, id, bson.M{"$set": set})
	return err
}

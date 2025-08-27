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

type DeviceRequestRepository interface {
	Create(ctx context.Context, request *domain.DeviceRequest) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*domain.DeviceRequest, error)
	GetByDeviceID(ctx context.Context, deviceID primitive.ObjectID) ([]*domain.DeviceRequest, error)
	GetPending(ctx context.Context, limit, offset int) ([]*domain.DeviceRequest, error)
	Approve(ctx context.Context, id primitive.ObjectID, approvedBy primitive.ObjectID) error
	Reject(ctx context.Context, id primitive.ObjectID, approvedBy primitive.ObjectID) error
	DeleteExpired(ctx context.Context) error
}

type deviceRequestRepository struct {
	collection *mongo.Collection
}

func NewDeviceRequestRepository(db *database.MongoDB) DeviceRequestRepository {
	return &deviceRequestRepository{
		collection: db.Database.Collection(database.DeviceRequestsCollection),
	}
}

func (r *deviceRequestRepository) Create(ctx context.Context, request *domain.DeviceRequest) error {
	request.ID = primitive.NewObjectID()
	request.Status = domain.DeviceRequestStatusPending
	request.CreatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, request)
	return err
}

func (r *deviceRequestRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.DeviceRequest, error) {
	var request domain.DeviceRequest
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&request)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &request, nil
}

func (r *deviceRequestRepository) GetByDeviceID(ctx context.Context, deviceID primitive.ObjectID) ([]*domain.DeviceRequest, error) {
	filter := bson.M{"device_id": deviceID}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var requests []*domain.DeviceRequest
	for cursor.Next(ctx) {
		var request domain.DeviceRequest
		if err := cursor.Decode(&request); err != nil {
			return nil, err
		}
		requests = append(requests, &request)
	}

	return requests, cursor.Err()
}

func (r *deviceRequestRepository) GetPending(ctx context.Context, limit, offset int) ([]*domain.DeviceRequest, error) {
	filter := bson.M{
		"status":     domain.DeviceRequestStatusPending,
		"expires_at": bson.M{"$gt": time.Now()},
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "created_at", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var requests []*domain.DeviceRequest
	for cursor.Next(ctx) {
		var request domain.DeviceRequest
		if err := cursor.Decode(&request); err != nil {
			return nil, err
		}
		requests = append(requests, &request)
	}

	return requests, cursor.Err()
}

func (r *deviceRequestRepository) Approve(ctx context.Context, id primitive.ObjectID, approvedBy primitive.ObjectID) error {
	now := time.Now()
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"status":      domain.DeviceRequestStatusApproved,
			"approved_by": approvedBy,
			"decided_at":  now,
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *deviceRequestRepository) Reject(ctx context.Context, id primitive.ObjectID, approvedBy primitive.ObjectID) error {
	now := time.Now()
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"status":      domain.DeviceRequestStatusRejected,
			"approved_by": approvedBy,
			"decided_at":  now,
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *deviceRequestRepository) DeleteExpired(ctx context.Context) error {
	filter := bson.M{"expires_at": bson.M{"$lte": time.Now()}}
	_, err := r.collection.DeleteMany(ctx, filter)
	return err
}
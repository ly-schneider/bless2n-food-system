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

type StationRepository interface {
	Create(ctx context.Context, station *domain.Station) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Station, error)
	GetByName(ctx context.Context, name string) (*domain.Station, error)
	Update(ctx context.Context, station *domain.Station) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	List(ctx context.Context, limit, offset int) ([]*domain.Station, error)
	ListByStatus(ctx context.Context, status domain.StationStatus, limit, offset int) ([]*domain.Station, error)
	ApproveStation(ctx context.Context, stationID, adminID primitive.ObjectID) error
	RejectStation(ctx context.Context, stationID, adminID primitive.ObjectID, reason string) error
}

type stationRepository struct {
	collection *mongo.Collection
}

func NewStationRepository(db *database.MongoDB) StationRepository {
	return &stationRepository{
		collection: db.Database.Collection(database.StationsCollection),
	}
}

func (r *stationRepository) Create(ctx context.Context, station *domain.Station) error {
	station.ID = primitive.NewObjectID()
	station.CreatedAt = time.Now()
	station.UpdatedAt = time.Now()
	if station.Status == "" {
		station.Status = domain.StationStatusPending
	}
	_, err := r.collection.InsertOne(ctx, station)
	return err
}

func (r *stationRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Station, error) {
	var station domain.Station
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&station)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &station, nil
}

func (r *stationRepository) GetByName(ctx context.Context, name string) (*domain.Station, error) {
	var station domain.Station
	err := r.collection.FindOne(ctx, bson.M{"name": name}).Decode(&station)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &station, nil
}

func (r *stationRepository) Update(ctx context.Context, station *domain.Station) error {
	filter := bson.M{"_id": station.ID}
	station.UpdatedAt = time.Now()
	update := bson.M{"$set": station}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *stationRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *stationRepository) List(ctx context.Context, limit, offset int) ([]*domain.Station, error) {
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "name", Value: 1}})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var stations []*domain.Station
	for cursor.Next(ctx) {
		var station domain.Station
		if err := cursor.Decode(&station); err != nil {
			return nil, err
		}
		stations = append(stations, &station)
	}

	return stations, cursor.Err()
}

func (r *stationRepository) ListByStatus(ctx context.Context, status domain.StationStatus, limit, offset int) ([]*domain.Station, error) {
	filter := bson.M{"status": status}
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var stations []*domain.Station
	for cursor.Next(ctx) {
		var station domain.Station
		if err := cursor.Decode(&station); err != nil {
			return nil, err
		}
		stations = append(stations, &station)
	}

	return stations, cursor.Err()
}


func (r *stationRepository) ApproveStation(ctx context.Context, stationID, adminID primitive.ObjectID) error {
	filter := bson.M{
		"_id": stationID,
		"status": domain.StationStatusPending,
	}
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":      domain.StationStatusApproved,
			"approved_by": adminID,
			"approved_at": now,
			"updated_at":  now,
		},
		"$unset": bson.M{
			"rejected_by":      "",
			"rejected_at":      "",
			"rejection_reason": "",
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *stationRepository) RejectStation(ctx context.Context, stationID, adminID primitive.ObjectID, reason string) error {
	filter := bson.M{
		"_id": stationID,
		"status": domain.StationStatusPending,
	}
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"status":           domain.StationStatusRejected,
			"rejected_by":      adminID,
			"rejected_at":      now,
			"rejection_reason": reason,
			"updated_at":       now,
		},
		"$unset": bson.M{
			"approved_by": "",
			"approved_at": "",
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

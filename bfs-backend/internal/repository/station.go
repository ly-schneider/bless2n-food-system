package repository

import (
	"backend/internal/database"
	"backend/internal/domain"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type StationRepository interface {
	UpsertPendingByDeviceKey(ctx context.Context, name, deviceKey string) (*domain.Station, error)
	FindByDeviceKey(ctx context.Context, deviceKey string) (*domain.Station, error)
	ApproveByDeviceKey(ctx context.Context, deviceKey string) error
	FindByID(ctx context.Context, id bson.ObjectID) (*domain.Station, error)
	List(ctx context.Context) ([]*domain.Station, error)
}

type stationRepository struct {
	collection *mongo.Collection
}

func NewStationRepository(db *database.MongoDB) StationRepository {
	return &stationRepository{
		collection: db.Database.Collection(database.StationsCollection),
	}
}

func (r *stationRepository) UpsertPendingByDeviceKey(ctx context.Context, name, deviceKey string) (*domain.Station, error) {
	now := time.Now().UTC()
	// Ensure a record exists for this deviceKey (approved=false by default)
	filter := bson.M{"device_key": deviceKey}
	update := bson.M{"$setOnInsert": bson.M{
		"_id":        bson.NewObjectID(),
		"name":       name,
		"device_key": deviceKey,
		"approved":   false,
		"created_at": now,
	}, "$set": bson.M{"updated_at": now}}
	opts := optionsFindOneAndUpdateUpsert()
	var st domain.Station
	if err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&st); err != nil {
		// If no doc existed, after upsert Decode may fail; fetch it
		if err == mongo.ErrNoDocuments {
			return r.FindByDeviceKey(ctx, deviceKey)
		}
		return nil, err
	}
	return &st, nil
}

func (r *stationRepository) FindByDeviceKey(ctx context.Context, deviceKey string) (*domain.Station, error) {
	var st domain.Station
	if err := r.collection.FindOne(ctx, bson.M{"device_key": deviceKey}).Decode(&st); err != nil {
		return nil, err
	}
	return &st, nil
}

func (r *stationRepository) ApproveByDeviceKey(ctx context.Context, deviceKey string) error {
	now := time.Now().UTC()
	_, err := r.collection.UpdateOne(ctx, bson.M{"device_key": deviceKey}, bson.M{"$set": bson.M{"approved": true, "approved_at": now, "updated_at": now}})
	return err
}

func (r *stationRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.Station, error) {
	var st domain.Station
	if err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&st); err != nil {
		return nil, err
	}
	return &st, nil
}

func (r *stationRepository) List(ctx context.Context) ([]*domain.Station, error) {
	cur, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = cur.Close(ctx) }()
	var out []*domain.Station
	for cur.Next(ctx) {
		var st domain.Station
		if err := cur.Decode(&st); err != nil {
			return nil, err
		}
		out = append(out, &st)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// optionsFindOneAndUpdateUpsert returns upsert=true and return after update
func optionsFindOneAndUpdateUpsert() options.Lister[options.FindOneAndUpdateOptions] {
    return options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)
}

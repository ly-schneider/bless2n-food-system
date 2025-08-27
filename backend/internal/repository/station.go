package repository

import (
	"context"

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
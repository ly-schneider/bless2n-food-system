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

type DeviceRepository interface {
	Create(ctx context.Context, device *domain.Device) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Device, error)
	GetByStationID(ctx context.Context, stationID primitive.ObjectID) ([]*domain.Device, error)
	Update(ctx context.Context, device *domain.Device) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	SetActive(ctx context.Context, id primitive.ObjectID, isActive bool) error
}

type deviceRepository struct {
	collection *mongo.Collection
}

func NewDeviceRepository(db *database.MongoDB) DeviceRepository {
	return &deviceRepository{
		collection: db.Database.Collection(database.DevicesCollection),
	}
}

func (r *deviceRepository) Create(ctx context.Context, device *domain.Device) error {
	device.ID = primitive.NewObjectID()
	device.IsActive = true
	_, err := r.collection.InsertOne(ctx, device)
	return err
}

func (r *deviceRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Device, error) {
	var device domain.Device
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&device)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &device, nil
}

func (r *deviceRepository) GetByStationID(ctx context.Context, stationID primitive.ObjectID) ([]*domain.Device, error) {
	filter := bson.M{"station_id": stationID}
	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var devices []*domain.Device
	for cursor.Next(ctx) {
		var device domain.Device
		if err := cursor.Decode(&device); err != nil {
			return nil, err
		}
		devices = append(devices, &device)
	}

	return devices, cursor.Err()
}

func (r *deviceRepository) Update(ctx context.Context, device *domain.Device) error {
	filter := bson.M{"_id": device.ID}
	update := bson.M{"$set": device}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *deviceRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *deviceRepository) SetActive(ctx context.Context, id primitive.ObjectID, isActive bool) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"is_active": isActive}}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}
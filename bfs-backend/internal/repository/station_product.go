package repository

import (
	"context"

	"backend/internal/database"
	"backend/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type StationProductRepository interface {
	Create(ctx context.Context, sp *domain.StationProduct) error
	GetByStationID(ctx context.Context, stationID primitive.ObjectID) ([]*domain.StationProduct, error)
	GetByProductID(ctx context.Context, productID primitive.ObjectID) ([]*domain.StationProduct, error)
	Delete(ctx context.Context, stationID, productID primitive.ObjectID) error
	DeleteByStationID(ctx context.Context, stationID primitive.ObjectID) error
	DeleteByProductID(ctx context.Context, productID primitive.ObjectID) error
}

type stationProductRepository struct {
	collection *mongo.Collection
}

func NewStationProductRepository(db *database.MongoDB) StationProductRepository {
	return &stationProductRepository{
		collection: db.Database.Collection(database.StationProductsCollection),
	}
}

func (r *stationProductRepository) Create(ctx context.Context, sp *domain.StationProduct) error {
	_, err := r.collection.InsertOne(ctx, sp)
	return err
}

func (r *stationProductRepository) GetByStationID(ctx context.Context, stationID primitive.ObjectID) ([]*domain.StationProduct, error) {
	filter := bson.M{"station_id": stationID}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var stationProducts []*domain.StationProduct
	for cursor.Next(ctx) {
		var sp domain.StationProduct
		if err := cursor.Decode(&sp); err != nil {
			return nil, err
		}
		stationProducts = append(stationProducts, &sp)
	}

	return stationProducts, cursor.Err()
}

func (r *stationProductRepository) GetByProductID(ctx context.Context, productID primitive.ObjectID) ([]*domain.StationProduct, error) {
	filter := bson.M{"product_id": productID}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var stationProducts []*domain.StationProduct
	for cursor.Next(ctx) {
		var sp domain.StationProduct
		if err := cursor.Decode(&sp); err != nil {
			return nil, err
		}
		stationProducts = append(stationProducts, &sp)
	}

	return stationProducts, cursor.Err()
}

func (r *stationProductRepository) Delete(ctx context.Context, stationID, productID primitive.ObjectID) error {
	filter := bson.M{
		"station_id": stationID,
		"product_id": productID,
	}
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

func (r *stationProductRepository) DeleteByStationID(ctx context.Context, stationID primitive.ObjectID) error {
	filter := bson.M{"station_id": stationID}
	_, err := r.collection.DeleteMany(ctx, filter)
	return err
}

func (r *stationProductRepository) DeleteByProductID(ctx context.Context, productID primitive.ObjectID) error {
	filter := bson.M{"product_id": productID}
	_, err := r.collection.DeleteMany(ctx, filter)
	return err
}

package repository

import (
	"backend/internal/database"
	"backend/internal/domain"
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type StationProductRepository interface {
	ListProductIDsByStation(ctx context.Context, stationID bson.ObjectID) ([]bson.ObjectID, error)
	IsProductAssigned(ctx context.Context, stationID, productID bson.ObjectID) (bool, error)
	AddProducts(ctx context.Context, stationID bson.ObjectID, productIDs []bson.ObjectID) (int64, error)
	RemoveProduct(ctx context.Context, stationID, productID bson.ObjectID) (bool, error)
}

type stationProductRepository struct {
	collection *mongo.Collection
}

func NewStationProductRepository(db *database.MongoDB) StationProductRepository {
	coll := db.Database.Collection(database.StationProductsCollection)
	// Ensure unique index on (station_id, product_id)
	_, _ = coll.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "station_id", Value: 1}, {Key: "product_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	return &stationProductRepository{collection: coll}
}

func (r *stationProductRepository) ListProductIDsByStation(ctx context.Context, stationID bson.ObjectID) ([]bson.ObjectID, error) {
	cur, err := r.collection.Find(ctx, bson.M{"station_id": stationID})
	if err != nil {
		return nil, err
	}
	defer func() { _ = cur.Close(ctx) }()
	out := make([]bson.ObjectID, 0)
	for cur.Next(ctx) {
		var sp domain.StationProduct
		if err := cur.Decode(&sp); err != nil {
			return nil, err
		}
		out = append(out, sp.ProductID)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *stationProductRepository) IsProductAssigned(ctx context.Context, stationID, productID bson.ObjectID) (bool, error) {
	err := r.collection.FindOne(ctx, bson.M{"station_id": stationID, "product_id": productID}).Err()
	if err == mongo.ErrNoDocuments {
		return false, nil
	}
	return err == nil, err
}

func (r *stationProductRepository) AddProducts(ctx context.Context, stationID bson.ObjectID, productIDs []bson.ObjectID) (int64, error) {
	if len(productIDs) == 0 {
		return 0, nil
	}
	models := make([]mongo.WriteModel, 0, len(productIDs))
	for _, pid := range productIDs {
		models = append(models, mongo.NewUpdateOneModel().
			SetFilter(bson.M{"station_id": stationID, "product_id": pid}).
			SetUpdate(bson.M{"$setOnInsert": bson.M{"_id": bson.NewObjectID(), "station_id": stationID, "product_id": pid}}).
			SetUpsert(true),
		)
	}
	if len(models) == 0 {
		return 0, nil
	}
	res, err := r.collection.BulkWrite(ctx, models, options.BulkWrite().SetOrdered(false))
	if err != nil {
		return 0, err
	}
	// Upserts count reflects newly created pairs
	return res.UpsertedCount, nil
}

func (r *stationProductRepository) RemoveProduct(ctx context.Context, stationID, productID bson.ObjectID) (bool, error) {
	res, err := r.collection.DeleteOne(ctx, bson.M{"station_id": stationID, "product_id": productID})
	if err != nil {
		return false, err
	}
	return res.DeletedCount > 0, nil
}

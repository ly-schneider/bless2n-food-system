package repository

import (
	"context"

	"backend/internal/database"
	"backend/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProductBundleComponentRepository interface {
	Create(ctx context.Context, component *domain.ProductBundleComponent) error
	GetByBundleID(ctx context.Context, bundleID primitive.ObjectID) ([]*domain.ProductBundleComponent, error)
	Update(ctx context.Context, component *domain.ProductBundleComponent) error
	Delete(ctx context.Context, bundleID, componentID primitive.ObjectID) error
	DeleteByBundleID(ctx context.Context, bundleID primitive.ObjectID) error
}

type productBundleComponentRepository struct {
	collection *mongo.Collection
}

func NewProductBundleComponentRepository(db *database.MongoDB) ProductBundleComponentRepository {
	return &productBundleComponentRepository{
		collection: db.Database.Collection(database.ProductBundleComponentsCollection),
	}
}

func (r *productBundleComponentRepository) Create(ctx context.Context, component *domain.ProductBundleComponent) error {
	_, err := r.collection.InsertOne(ctx, component)
	return err
}

func (r *productBundleComponentRepository) GetByBundleID(ctx context.Context, bundleID primitive.ObjectID) ([]*domain.ProductBundleComponent, error) {
	filter := bson.M{"bundle_id": bundleID}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var components []*domain.ProductBundleComponent
	for cursor.Next(ctx) {
		var component domain.ProductBundleComponent
		if err := cursor.Decode(&component); err != nil {
			return nil, err
		}
		components = append(components, &component)
	}

	return components, cursor.Err()
}

func (r *productBundleComponentRepository) Update(ctx context.Context, component *domain.ProductBundleComponent) error {
	filter := bson.M{
		"bundle_id":            component.BundleID,
		"component_product_id": component.ComponentProductID,
	}
	update := bson.M{"$set": component}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *productBundleComponentRepository) Delete(ctx context.Context, bundleID, componentID primitive.ObjectID) error {
	filter := bson.M{
		"bundle_id":            bundleID,
		"component_product_id": componentID,
	}
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

func (r *productBundleComponentRepository) DeleteByBundleID(ctx context.Context, bundleID primitive.ObjectID) error {
	filter := bson.M{"bundle_id": bundleID}
	_, err := r.collection.DeleteMany(ctx, filter)
	return err
}

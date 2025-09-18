package repository

import (
	"backend/internal/database"
	"backend/internal/domain"
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ProductRepository interface {
	GetAll(ctx context.Context, limit int, offset int) ([]*domain.Product, error)
	GetByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*domain.Product, error)
	GetByCategoryID(ctx context.Context, categoryID primitive.ObjectID, limit int, offset int) ([]*domain.Product, error)
}

type productRepository struct {
	collection *mongo.Collection
}

func NewProductRepository(db *database.MongoDB) ProductRepository {
	return &productRepository{
		collection: db.Database.Collection(database.ProductsCollection),
	}
}

func (r *productRepository) GetAll(ctx context.Context, limit int, offset int) (products []*domain.Product, err error) {

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(primitive.M{"name": 1})

    cursor, err := r.collection.Find(ctx, primitive.M{}, opts)
    if err != nil {
        return nil, err
    }
    defer func() {
        if cerr := cursor.Close(ctx); err == nil && cerr != nil {
            err = cerr
        }
    }()

	for cursor.Next(ctx) {
		var product domain.Product
        if derr := cursor.Decode(&product); derr != nil {
            err = derr
            return nil, err
        }
        products = append(products, &product)
    }

    if derr := cursor.Err(); derr != nil {
        err = derr
        return nil, err
    }

	return products, nil
}

func (r *productRepository) GetByIDs(ctx context.Context, ids []primitive.ObjectID) (products []*domain.Product, err error) {

    cursor, err := r.collection.Find(ctx, primitive.M{"_id": primitive.M{"$in": ids}})
    if err != nil {
        return nil, err
    }
    defer func() {
        if cerr := cursor.Close(ctx); err == nil && cerr != nil {
            err = cerr
        }
    }()

	for cursor.Next(ctx) {
		var product domain.Product
        if derr := cursor.Decode(&product); derr != nil {
            err = derr
            return nil, err
        }
        products = append(products, &product)
    }

    if derr := cursor.Err(); derr != nil {
        err = derr
        return nil, err
    }

	return products, nil
}

func (r *productRepository) GetByCategoryID(ctx context.Context, categoryID primitive.ObjectID, limit int, offset int) (products []*domain.Product, err error) {

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(primitive.M{"name": 1})

    cursor, err := r.collection.Find(ctx, primitive.M{"category_id": categoryID}, opts)
    if err != nil {
        return nil, err
    }
    defer func() {
        if cerr := cursor.Close(ctx); err == nil && cerr != nil {
            err = cerr
        }
    }()

	for cursor.Next(ctx) {
		var product domain.Product
        if derr := cursor.Decode(&product); derr != nil {
            err = derr
            return nil, err
        }
        products = append(products, &product)
    }

    if derr := cursor.Err(); derr != nil {
        err = derr
        return nil, err
    }

	return products, nil
}

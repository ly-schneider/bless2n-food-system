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

func (r *productRepository) GetAll(ctx context.Context, limit int, offset int) ([]*domain.Product, error) {
	var products []*domain.Product

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(primitive.M{"name": 1})

	cursor, err := r.collection.Find(ctx, primitive.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var product domain.Product
		if err := cursor.Decode(&product); err != nil {
			return nil, err
		}
		products = append(products, &product)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func (r *productRepository) GetByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*domain.Product, error) {
	var products []*domain.Product

	cursor, err := r.collection.Find(ctx, primitive.M{"_id": primitive.M{"$in": ids}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var product domain.Product
		if err := cursor.Decode(&product); err != nil {
			return nil, err
		}
		products = append(products, &product)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

func (r *productRepository) GetByCategoryID(ctx context.Context, categoryID primitive.ObjectID, limit int, offset int) ([]*domain.Product, error) {
	var products []*domain.Product

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(primitive.M{"name": 1})

	cursor, err := r.collection.Find(ctx, primitive.M{"category_id": categoryID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var product domain.Product
		if err := cursor.Decode(&product); err != nil {
			return nil, err
		}
		products = append(products, &product)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return products, nil
}

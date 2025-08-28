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

type ProductRepository interface {
	Create(ctx context.Context, product *domain.Product) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Product, error)
	GetByCategoryID(ctx context.Context, categoryID primitive.ObjectID, activeOnly bool, limit, offset int) ([]*domain.Product, error)
	Update(ctx context.Context, product *domain.Product) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	Search(ctx context.Context, query string, limit, offset int) ([]*domain.Product, error)
	SetActive(ctx context.Context, id primitive.ObjectID, isActive bool) error
	GetByType(ctx context.Context, productType domain.ProductType, limit, offset int) ([]*domain.Product, error)
}

type productRepository struct {
	collection *mongo.Collection
}

func NewProductRepository(db *database.MongoDB) ProductRepository {
	return &productRepository{
		collection: db.Database.Collection(database.ProductsCollection),
	}
}

func (r *productRepository) Create(ctx context.Context, product *domain.Product) error {
	product.ID = primitive.NewObjectID()
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()
	product.IsActive = true

	_, err := r.collection.InsertOne(ctx, product)
	return err
}

func (r *productRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Product, error) {
	var product domain.Product
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) GetByCategoryID(ctx context.Context, categoryID primitive.ObjectID, activeOnly bool, limit, offset int) ([]*domain.Product, error) {
	filter := bson.M{"category_id": categoryID}
	if activeOnly {
		filter["is_active"] = true
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "name", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []*domain.Product
	for cursor.Next(ctx) {
		var product domain.Product
		if err := cursor.Decode(&product); err != nil {
			return nil, err
		}
		products = append(products, &product)
	}

	return products, cursor.Err()
}

func (r *productRepository) Update(ctx context.Context, product *domain.Product) error {
	product.UpdatedAt = time.Now()
	filter := bson.M{"_id": product.ID}
	update := bson.M{"$set": product}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *productRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *productRepository) Search(ctx context.Context, query string, limit, offset int) ([]*domain.Product, error) {
	filter := bson.M{
		"name":      bson.M{"$regex": query, "$options": "i"},
		"is_active": true,
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "name", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []*domain.Product
	for cursor.Next(ctx) {
		var product domain.Product
		if err := cursor.Decode(&product); err != nil {
			return nil, err
		}
		products = append(products, &product)
	}

	return products, cursor.Err()
}

func (r *productRepository) SetActive(ctx context.Context, id primitive.ObjectID, isActive bool) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"is_active":  isActive,
			"updated_at": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *productRepository) GetByType(ctx context.Context, productType domain.ProductType, limit, offset int) ([]*domain.Product, error) {
	filter := bson.M{
		"type":      productType,
		"is_active": true,
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "name", Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []*domain.Product
	for cursor.Next(ctx) {
		var product domain.Product
		if err := cursor.Decode(&product); err != nil {
			return nil, err
		}
		products = append(products, &product)
	}

	return products, cursor.Err()
}

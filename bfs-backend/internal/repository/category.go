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

type CategoryRepository interface {
	Create(ctx context.Context, category *domain.Category) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Category, error)
	GetByName(ctx context.Context, name string) (*domain.Category, error)
	Update(ctx context.Context, category *domain.Category) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	List(ctx context.Context, activeOnly bool, limit, offset int) ([]*domain.Category, error)
	SetActive(ctx context.Context, id primitive.ObjectID, isActive bool) error
}

type categoryRepository struct {
	collection *mongo.Collection
}

func NewCategoryRepository(db *database.MongoDB) CategoryRepository {
	return &categoryRepository{
		collection: db.Database.Collection(database.CategoriesCollection),
	}
}

func (r *categoryRepository) Create(ctx context.Context, category *domain.Category) error {
	category.ID = primitive.NewObjectID()
	category.CreatedAt = time.Now()
	category.UpdatedAt = time.Now()
	category.IsActive = true

	_, err := r.collection.InsertOne(ctx, category)
	return err
}

func (r *categoryRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Category, error) {
	var category domain.Category
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&category)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) GetByName(ctx context.Context, name string) (*domain.Category, error) {
	var category domain.Category
	err := r.collection.FindOne(ctx, bson.M{"name": name}).Decode(&category)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) Update(ctx context.Context, category *domain.Category) error {
	category.UpdatedAt = time.Now()
	filter := bson.M{"_id": category.ID}
	update := bson.M{"$set": category}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *categoryRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *categoryRepository) List(ctx context.Context, activeOnly bool, limit, offset int) ([]*domain.Category, error) {
	filter := bson.M{}
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

	var categories []*domain.Category
	for cursor.Next(ctx) {
		var category domain.Category
		if err := cursor.Decode(&category); err != nil {
			return nil, err
		}
		categories = append(categories, &category)
	}

	return categories, cursor.Err()
}

func (r *categoryRepository) SetActive(ctx context.Context, id primitive.ObjectID, isActive bool) error {
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

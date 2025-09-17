package repository

import (
	"backend/internal/database"
	"backend/internal/domain"
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CategoryRepository interface {
	GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Category, error)
	GetByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*domain.Category, error)
}

type categoryRepository struct {
	collection *mongo.Collection
}

func NewCategoryRepository(db *database.MongoDB) CategoryRepository {
	return &categoryRepository{
		collection: db.Database.Collection(database.CategoriesCollection),
	}
}

func (r *categoryRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Category, error) {
	var category domain.Category
	err := r.collection.FindOne(ctx, primitive.M{"_id": id}).Decode(&category)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) GetByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*domain.Category, error) {
	var categories []*domain.Category

	cursor, err := r.collection.Find(ctx, primitive.M{"_id": primitive.M{"$in": ids}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var category domain.Category
		if err := cursor.Decode(&category); err != nil {
			return nil, err
		}
		categories = append(categories, &category)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

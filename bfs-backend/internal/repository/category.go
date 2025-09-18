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

func (r *categoryRepository) GetByIDs(ctx context.Context, ids []primitive.ObjectID) (categories []*domain.Category, err error) {
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
        var category domain.Category
        if derr := cursor.Decode(&category); derr != nil {
            err = derr
            return nil, err
        }
        categories = append(categories, &category)
    }

    if derr := cursor.Err(); derr != nil {
        err = derr
        return nil, err
    }

    return categories, nil
}

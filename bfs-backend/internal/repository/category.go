package repository

import (
    "backend/internal/database"
    "backend/internal/domain"
    "context"

    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type CategoryRepository interface {
    GetByID(ctx context.Context, id primitive.ObjectID) (*domain.Category, error)
    GetByIDs(ctx context.Context, ids []primitive.ObjectID) ([]*domain.Category, error)
    List(ctx context.Context, active *bool, q *string, limit, offset int) ([]*domain.Category, int64, error)
    Insert(ctx context.Context, c *domain.Category) (primitive.ObjectID, error)
    UpdateFields(ctx context.Context, id primitive.ObjectID, set primitive.M) error
    DeleteByID(ctx context.Context, id primitive.ObjectID) error
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

func (r *categoryRepository) List(ctx context.Context, active *bool, q *string, limit, offset int) ([]*domain.Category, int64, error) {
    filter := primitive.M{}
    if active != nil { filter["is_active"] = *active }
    if q != nil && *q != "" { filter["name"] = primitive.M{"$regex": *q, "$options": "i"} }
    total, err := r.collection.CountDocuments(ctx, filter)
    if err != nil { return nil, 0, err }
    opts := options.Find().SetSort(primitive.M{"name": 1})
    if limit > 0 { opts.SetLimit(int64(limit)) }
    if offset > 0 { opts.SetSkip(int64(offset)) }
    cur, err := r.collection.Find(ctx, filter, opts)
    if err != nil { return nil, 0, err }
    defer func() { _ = cur.Close(ctx) }()
    var out []*domain.Category
    for cur.Next(ctx) {
        var c domain.Category
        if err := cur.Decode(&c); err != nil { return nil, 0, err }
        out = append(out, &c)
    }
    if err := cur.Err(); err != nil { return nil, 0, err }
    return out, total, nil
}

func (r *categoryRepository) Insert(ctx context.Context, c *domain.Category) (primitive.ObjectID, error) {
    if c.ID.IsZero() { c.ID = primitive.NewObjectID() }
    now := time.Now().UTC()
    if c.CreatedAt.IsZero() { c.CreatedAt = now }
    c.UpdatedAt = now
    _, err := r.collection.InsertOne(ctx, c)
    if err != nil { return primitive.NilObjectID, err }
    return c.ID, nil
}

func (r *categoryRepository) UpdateFields(ctx context.Context, id primitive.ObjectID, set primitive.M) error {
    if set == nil { return nil }
    if set["updated_at"] == nil { set["updated_at"] = time.Now().UTC() }
    _, err := r.collection.UpdateByID(ctx, id, primitive.M{"$set": set})
    return err
}

func (r *categoryRepository) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
    _, err := r.collection.DeleteOne(ctx, primitive.M{"_id": id})
    return err
}

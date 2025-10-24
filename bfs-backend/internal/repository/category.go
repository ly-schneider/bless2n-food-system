package repository

import (
	"backend/internal/database"
	"backend/internal/domain"
	"context"

	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

type CategoryRepository interface {
	GetByID(ctx context.Context, id bson.ObjectID) (*domain.Category, error)
	GetByIDs(ctx context.Context, ids []bson.ObjectID) ([]*domain.Category, error)
	List(ctx context.Context, active *bool, q *string, limit, offset int) ([]*domain.Category, int64, error)
	Insert(ctx context.Context, c *domain.Category) (bson.ObjectID, error)
	UpdateFields(ctx context.Context, id bson.ObjectID, set bson.M) error
	DeleteByID(ctx context.Context, id bson.ObjectID) error
}

type categoryRepository struct {
	collection *mongo.Collection
}

func NewCategoryRepository(db *database.MongoDB) CategoryRepository {
	return &categoryRepository{
		collection: db.Database.Collection(database.CategoriesCollection),
	}
}

func (r *categoryRepository) GetByID(ctx context.Context, id bson.ObjectID) (*domain.Category, error) {
	var raw bson.M
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&raw)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return mapRawToCategory(raw), nil
}

func (r *categoryRepository) GetByIDs(ctx context.Context, ids []bson.ObjectID) (categories []*domain.Category, err error) {
	cursor, err := r.collection.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := cursor.Close(ctx); err == nil && cerr != nil {
			err = cerr
		}
	}()

	for cursor.Next(ctx) {
		var raw bson.M
		if derr := cursor.Decode(&raw); derr != nil {
			err = derr
			return nil, err
		}
		categories = append(categories, mapRawToCategory(raw))
	}

	if derr := cursor.Err(); derr != nil {
		err = derr
		return nil, err
	}

	return categories, nil
}

func (r *categoryRepository) List(ctx context.Context, active *bool, q *string, limit, offset int) ([]*domain.Category, int64, error) {
	filter := bson.M{}
	if active != nil {
		filter["is_active"] = *active
	}
	if q != nil && *q != "" {
		filter["name"] = bson.M{"$regex": *q, "$options": "i"}
	}
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		// Graceful fallback for environments where count may fail (e.g., certain compat layers)
		zap.L().Warn("categories count failed; falling back to page count", zap.Error(err))
		total = -1 // sentinel to recompute after fetch
	}
	// Sort by explicit position first, fallback by name for stability (ordered)
	opts := options.Find().SetSort(bson.D{{Key: "position", Value: 1}, {Key: "name", Value: 1}})
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}
	if offset > 0 {
		opts.SetSkip(int64(offset))
	}
	cur, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = cur.Close(ctx) }()
	var out []*domain.Category
	for cur.Next(ctx) {
		var raw bson.M
		if err := cur.Decode(&raw); err != nil {
			return nil, 0, err
		}
		out = append(out, mapRawToCategory(raw))
	}
	if err := cur.Err(); err != nil {
		return nil, 0, err
	}
	if total < 0 {
		total = int64(len(out))
	}
	return out, total, nil
}

func (r *categoryRepository) Insert(ctx context.Context, c *domain.Category) (bson.ObjectID, error) {
	if c.ID.IsZero() {
		c.ID = bson.NewObjectID()
	}
	now := time.Now().UTC()
	if c.CreatedAt.IsZero() {
		c.CreatedAt = now
	}
	c.UpdatedAt = now
	_, err := r.collection.InsertOne(ctx, c)
	if err != nil {
		return bson.NilObjectID, err
	}
	return c.ID, nil
}

func (r *categoryRepository) UpdateFields(ctx context.Context, id bson.ObjectID, set bson.M) error {
	if set == nil {
		return nil
	}
	if set["updated_at"] == nil {
		set["updated_at"] = time.Now().UTC()
	}
	_, err := r.collection.UpdateByID(ctx, id, bson.M{"$set": set})
	return err
}

func (r *categoryRepository) DeleteByID(ctx context.Context, id bson.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// Tolerant mapper to avoid decode errors on null/typed position
func mapRawToCategory(m bson.M) *domain.Category {
	var c domain.Category
	if v, ok := m["_id"].(bson.ObjectID); ok {
		c.ID = v
	}
	if v, ok := m["name"].(string); ok {
		c.Name = v
	}
	if v, ok := m["is_active"].(bool); ok {
		c.IsActive = v
	}
	// position may be int32/int64/float64/string/nil
	switch v := m["position"].(type) {
	case int32:
		c.Position = int(v)
	case int64:
		c.Position = int(v)
	case float64:
		c.Position = int(v)
	case string:
		// best-effort parse
		// ignore error => default 0
		// no strconv import to keep minimal; numbers from seeds are ints
		// leave as 0 if not numeric
	default:
		// missing or nil => default 0
	}
	if v, ok := m["created_at"].(time.Time); ok {
		c.CreatedAt = v
	}
	if v, ok := m["updated_at"].(time.Time); ok {
		c.UpdatedAt = v
	}
	return &c
}

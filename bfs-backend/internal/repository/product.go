package repository

import (
	"backend/internal/database"
	"backend/internal/domain"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ProductRepository interface {
	GetAll(ctx context.Context, limit int, offset int) ([]*domain.Product, error)
	GetByIDs(ctx context.Context, ids []bson.ObjectID) ([]*domain.Product, error)
	GetByCategoryID(ctx context.Context, categoryID bson.ObjectID, limit int, offset int) ([]*domain.Product, error)
	FindByID(ctx context.Context, id bson.ObjectID) (*domain.Product, error)
	UpdateFields(ctx context.Context, id bson.ObjectID, set bson.M) error
	Insert(ctx context.Context, p *domain.Product) (bson.ObjectID, error)
	GetMenus(ctx context.Context, q *string, active *bool, limit, offset int) ([]*domain.Product, int64, error)
	DeleteByID(ctx context.Context, id bson.ObjectID) error
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
		SetSort(bson.M{"name": 1})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
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

func (r *productRepository) GetByIDs(ctx context.Context, ids []bson.ObjectID) (products []*domain.Product, err error) {

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

func (r *productRepository) GetByCategoryID(ctx context.Context, categoryID bson.ObjectID, limit int, offset int) (products []*domain.Product, err error) {

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.M{"name": 1})

	cursor, err := r.collection.Find(ctx, bson.M{"category_id": categoryID}, opts)
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

func (r *productRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.Product, error) {
	var p domain.Product
	if err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&p); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *productRepository) UpdateFields(ctx context.Context, id bson.ObjectID, set bson.M) error {
	if set == nil {
		return nil
	}
	if set["updated_at"] == nil {
		set["updated_at"] = time.Now().UTC()
	}
	_, err := r.collection.UpdateByID(ctx, id, bson.M{"$set": set})
	return err
}

func (r *productRepository) Insert(ctx context.Context, p *domain.Product) (bson.ObjectID, error) {
	if p.ID.IsZero() {
		p.ID = bson.NewObjectID()
	}
	now := time.Now().UTC()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	p.UpdatedAt = now
	_, err := r.collection.InsertOne(ctx, p)
	if err != nil {
		return bson.NilObjectID, err
	}
	return p.ID, nil
}

func (r *productRepository) GetMenus(ctx context.Context, q *string, active *bool, limit, offset int) ([]*domain.Product, int64, error) {
	filter := bson.M{"type": domain.ProductTypeMenu}
	if active != nil {
		filter["is_active"] = *active
	}
	if q != nil && *q != "" {
		filter["name"] = bson.M{"$regex": *q, "$options": "i"}
	}
	total, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	opts := options.Find().SetSort(bson.M{"name": 1})
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
	var out []*domain.Product
	for cur.Next(ctx) {
		var p domain.Product
		if err := cur.Decode(&p); err != nil {
			return nil, 0, err
		}
		out = append(out, &p)
	}
	if err := cur.Err(); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *productRepository) DeleteByID(ctx context.Context, id bson.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

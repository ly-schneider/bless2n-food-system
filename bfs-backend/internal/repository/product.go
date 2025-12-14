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
	UpdateJeton(ctx context.Context, id bson.ObjectID, jetonID *bson.ObjectID) error
	Insert(ctx context.Context, p *domain.Product) (bson.ObjectID, error)
	GetMenus(ctx context.Context, q *string, active *bool, limit, offset int) ([]*domain.Product, int64, error)
	DeleteByID(ctx context.Context, id bson.ObjectID) error
	CountActiveWithoutJeton(ctx context.Context) (int64, error)
	CountByJetonIDs(ctx context.Context, ids []bson.ObjectID) (map[bson.ObjectID]int64, error)
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

func (r *productRepository) UpdateJeton(ctx context.Context, id bson.ObjectID, jetonID *bson.ObjectID) error {
	set := bson.M{"updated_at": time.Now().UTC()}
	update := bson.M{"$set": set}
	if jetonID != nil {
		set["jeton_id"] = *jetonID
	} else {
		update["$unset"] = bson.M{"jeton_id": ""}
	}
	_, err := r.collection.UpdateByID(ctx, id, update)
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

func (r *productRepository) CountActiveWithoutJeton(ctx context.Context) (int64, error) {
	filter := bson.M{
		"is_active": true,
		"type":      domain.ProductTypeSimple,
		"$or": []bson.M{
			{"jeton_id": bson.M{"$exists": false}},
			{"jeton_id": nil},
		},
	}
	return r.collection.CountDocuments(ctx, filter)
}

func (r *productRepository) CountByJetonIDs(ctx context.Context, ids []bson.ObjectID) (map[bson.ObjectID]int64, error) {
	out := make(map[bson.ObjectID]int64)
	if len(ids) == 0 {
		return out, nil
	}
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.M{
			"jeton_id": bson.M{"$in": ids},
			"type":     domain.ProductTypeSimple,
		}}},
		{{Key: "$group", Value: bson.M{"_id": "$jeton_id", "count": bson.M{"$sum": 1}}}},
	}
	cur, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cur.Close(ctx) }()
	for cur.Next(ctx) {
		var row struct {
			ID    bson.ObjectID `bson:"_id"`
			Count int64         `bson:"count"`
		}
		if err := cur.Decode(&row); err != nil {
			return nil, err
		}
		out[row.ID] = row.Count
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

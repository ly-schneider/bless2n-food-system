package seed

import (
	"backend/internal/database"
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func categoryIDsByName(ctx context.Context, db *mongo.Database, names []string) (map[string]bson.ObjectID, error) {
	result := make(map[string]bson.ObjectID)
	if len(names) == 0 {
		return result, nil
	}
	coll := db.Collection(database.CategoriesCollection)
	cur, err := coll.Find(ctx, bson.M{"name": bson.M{"$in": names}})
	if err != nil {
		return nil, fmt.Errorf("find categories: %w", err)
	}
	defer func() { _ = cur.Close(ctx) }()

	for cur.Next(ctx) {
		var c struct {
			ID   bson.ObjectID `bson:"_id"`
			Name string        `bson:"name"`
		}
		if err := cur.Decode(&c); err != nil {
			return nil, fmt.Errorf("decode category: %w", err)
		}
		result[c.Name] = c.ID
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func productIDsByName(ctx context.Context, db *mongo.Database, names []string) (map[string]bson.ObjectID, error) {
	result := make(map[string]bson.ObjectID)
	if len(names) == 0 {
		return result, nil
	}
	coll := db.Collection(database.ProductsCollection)
	cur, err := coll.Find(ctx, bson.M{"name": bson.M{"$in": names}})
	if err != nil {
		return nil, fmt.Errorf("find products: %w", err)
	}
	defer func() { _ = cur.Close(ctx) }()

	for cur.Next(ctx) {
		var p struct {
			ID   bson.ObjectID `bson:"_id"`
			Name string        `bson:"name"`
		}
		if err := cur.Decode(&p); err != nil {
			return nil, fmt.Errorf("decode product: %w", err)
		}
		result[p.Name] = p.ID
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func jetonIDsByName(ctx context.Context, db *mongo.Database, names []string) (map[string]bson.ObjectID, error) {
	result := make(map[string]bson.ObjectID)
	if len(names) == 0 {
		return result, nil
	}
	coll := db.Collection(database.JetonsCollection)
	cur, err := coll.Find(ctx, bson.M{"name": bson.M{"$in": names}})
	if err != nil {
		return nil, fmt.Errorf("find jetons: %w", err)
	}
	defer func() { _ = cur.Close(ctx) }()

	for cur.Next(ctx) {
		var j struct {
			ID   bson.ObjectID `bson:"_id"`
			Name string        `bson:"name"`
		}
		if err := cur.Decode(&j); err != nil {
			return nil, fmt.Errorf("decode jeton: %w", err)
		}
		result[j.Name] = j.ID
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func categoryNamesByID(ctx context.Context, db *mongo.Database) (map[bson.ObjectID]string, error) {
	result := make(map[bson.ObjectID]string)
	coll := db.Collection(database.CategoriesCollection)
	cur, err := coll.Find(ctx, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("find categories: %w", err)
	}
	defer func() { _ = cur.Close(ctx) }()

	for cur.Next(ctx) {
		var c struct {
			ID   bson.ObjectID `bson:"_id"`
			Name string        `bson:"name"`
		}
		if err := cur.Decode(&c); err != nil {
			return nil, fmt.Errorf("decode category: %w", err)
		}
		result[c.ID] = c.Name
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

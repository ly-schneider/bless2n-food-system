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

type JetonRepository interface {
	List(ctx context.Context) ([]*domain.Jeton, error)
	FindByID(ctx context.Context, id bson.ObjectID) (*domain.Jeton, error)
	FindByIDs(ctx context.Context, ids []bson.ObjectID) ([]*domain.Jeton, error)
	Insert(ctx context.Context, j *domain.Jeton) (bson.ObjectID, error)
	Update(ctx context.Context, id bson.ObjectID, set bson.M) error
	Delete(ctx context.Context, id bson.ObjectID) error
}

type jetonRepository struct {
	collection *mongo.Collection
}

func NewJetonRepository(db *database.MongoDB) JetonRepository {
	return &jetonRepository{collection: db.Database.Collection(database.JetonsCollection)}
}

func (r *jetonRepository) List(ctx context.Context) ([]*domain.Jeton, error) {
	cur, err := r.collection.Find(ctx, bson.M{}, options.Find().SetSort(bson.M{"name": 1}))
	if err != nil {
		return nil, err
	}
	defer func() { _ = cur.Close(ctx) }()
	out := make([]*domain.Jeton, 0)
	for cur.Next(ctx) {
		var j domain.Jeton
		if err := cur.Decode(&j); err != nil {
			return nil, err
		}
		out = append(out, &j)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *jetonRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.Jeton, error) {
	var j domain.Jeton
	if err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&j); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &j, nil
}

func (r *jetonRepository) FindByIDs(ctx context.Context, ids []bson.ObjectID) ([]*domain.Jeton, error) {
	if len(ids) == 0 {
		return []*domain.Jeton{}, nil
	}
	cur, err := r.collection.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return nil, err
	}
	defer func() { _ = cur.Close(ctx) }()
	out := make([]*domain.Jeton, 0)
	for cur.Next(ctx) {
		var j domain.Jeton
		if err := cur.Decode(&j); err != nil {
			return nil, err
		}
		out = append(out, &j)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *jetonRepository) Insert(ctx context.Context, j *domain.Jeton) (bson.ObjectID, error) {
	now := time.Now().UTC()
	if j.ID.IsZero() {
		j.ID = bson.NewObjectID()
	}
	if j.CreatedAt.IsZero() {
		j.CreatedAt = now
	}
	j.UpdatedAt = now
	if _, err := r.collection.InsertOne(ctx, j); err != nil {
		return bson.NilObjectID, err
	}
	return j.ID, nil
}

func (r *jetonRepository) Update(ctx context.Context, id bson.ObjectID, set bson.M) error {
	if set == nil {
		return nil
	}
	if set["updated_at"] == nil {
		set["updated_at"] = time.Now().UTC()
	}
	_, err := r.collection.UpdateByID(ctx, id, bson.M{"$set": set})
	return err
}

func (r *jetonRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

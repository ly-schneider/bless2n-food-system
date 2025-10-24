package repository

import (
	"backend/internal/database"
	"backend/internal/domain"
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type MenuSlotRepository interface {
	FindByProductIDs(ctx context.Context, productIDs []bson.ObjectID) ([]*domain.MenuSlot, error)
	Insert(ctx context.Context, slot *domain.MenuSlot) (bson.ObjectID, error)
	UpdateName(ctx context.Context, id bson.ObjectID, name string) error
	UpdateSequence(ctx context.Context, id bson.ObjectID, seq int) error
	UpdateSequences(ctx context.Context, seqs map[bson.ObjectID]int) error
	DeleteByID(ctx context.Context, id bson.ObjectID) error
}

type menuSlotRepository struct {
	collection *mongo.Collection
}

func NewMenuSlotRepository(db *database.MongoDB) MenuSlotRepository {
	return &menuSlotRepository{
		collection: db.Database.Collection(database.MenuSlotsCollection),
	}
}

func (r *menuSlotRepository) FindByProductIDs(ctx context.Context, productIDs []bson.ObjectID) (menuSlots []*domain.MenuSlot, err error) {
	cursor, err := r.collection.Find(ctx, bson.M{"product_id": bson.M{"$in": productIDs}})
	if err != nil {
		return nil, err
	}
	defer func() {
		if cerr := cursor.Close(ctx); err == nil && cerr != nil {
			err = cerr
		}
	}()

	for cursor.Next(ctx) {
		var menuSlot domain.MenuSlot
		if derr := cursor.Decode(&menuSlot); derr != nil {
			err = derr
			return nil, err
		}
		menuSlots = append(menuSlots, &menuSlot)
	}

	if derr := cursor.Err(); derr != nil {
		err = derr
		return nil, err
	}

	return menuSlots, nil
}

func (r *menuSlotRepository) Insert(ctx context.Context, slot *domain.MenuSlot) (bson.ObjectID, error) {
	if slot.ID.IsZero() {
		slot.ID = bson.NewObjectID()
	}
	_, err := r.collection.InsertOne(ctx, slot)
	if err != nil {
		return bson.NilObjectID, err
	}
	return slot.ID, nil
}

func (r *menuSlotRepository) UpdateName(ctx context.Context, id bson.ObjectID, name string) error {
	_, err := r.collection.UpdateByID(ctx, id, bson.M{"$set": bson.M{"name": name}})
	return err
}

func (r *menuSlotRepository) UpdateSequence(ctx context.Context, id bson.ObjectID, seq int) error {
	_, err := r.collection.UpdateByID(ctx, id, bson.M{"$set": bson.M{"sequence": seq}})
	return err
}

func (r *menuSlotRepository) UpdateSequences(ctx context.Context, seqs map[bson.ObjectID]int) error {
	if len(seqs) == 0 {
		return nil
	}
	for id, seq := range seqs {
		if _, err := r.collection.UpdateByID(ctx, id, bson.M{"$set": bson.M{"sequence": seq}}); err != nil {
			return err
		}
	}
	return nil
}

func (r *menuSlotRepository) DeleteByID(ctx context.Context, id bson.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

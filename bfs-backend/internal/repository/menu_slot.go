package repository

import (
	"backend/internal/database"
	"backend/internal/domain"
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MenuSlotRepository interface {
    FindByProductIDs(ctx context.Context, productIDs []primitive.ObjectID) ([]*domain.MenuSlot, error)
    Insert(ctx context.Context, slot *domain.MenuSlot) (primitive.ObjectID, error)
    UpdateName(ctx context.Context, id primitive.ObjectID, name string) error
    UpdateSequence(ctx context.Context, id primitive.ObjectID, seq int) error
    UpdateSequences(ctx context.Context, seqs map[primitive.ObjectID]int) error
    DeleteByID(ctx context.Context, id primitive.ObjectID) error
}

type menuSlotRepository struct {
	collection *mongo.Collection
}

func NewMenuSlotRepository(db *database.MongoDB) MenuSlotRepository {
	return &menuSlotRepository{
		collection: db.Database.Collection(database.MenuSlotsCollection),
	}
}

func (r *menuSlotRepository) FindByProductIDs(ctx context.Context, productIDs []primitive.ObjectID) (menuSlots []*domain.MenuSlot, err error) {
    cursor, err := r.collection.Find(ctx, primitive.M{"product_id": primitive.M{"$in": productIDs}})
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

func (r *menuSlotRepository) Insert(ctx context.Context, slot *domain.MenuSlot) (primitive.ObjectID, error) {
    if slot.ID.IsZero() { slot.ID = primitive.NewObjectID() }
    _, err := r.collection.InsertOne(ctx, slot)
    if err != nil { return primitive.NilObjectID, err }
    return slot.ID, nil
}

func (r *menuSlotRepository) UpdateName(ctx context.Context, id primitive.ObjectID, name string) error {
    _, err := r.collection.UpdateByID(ctx, id, primitive.M{"$set": primitive.M{"name": name}})
    return err
}

func (r *menuSlotRepository) UpdateSequence(ctx context.Context, id primitive.ObjectID, seq int) error {
    _, err := r.collection.UpdateByID(ctx, id, primitive.M{"$set": primitive.M{"sequence": seq}})
    return err
}

func (r *menuSlotRepository) UpdateSequences(ctx context.Context, seqs map[primitive.ObjectID]int) error {
    if len(seqs) == 0 { return nil }
    for id, seq := range seqs {
        if _, err := r.collection.UpdateByID(ctx, id, primitive.M{"$set": primitive.M{"sequence": seq}}); err != nil { return err }
    }
    return nil
}

func (r *menuSlotRepository) DeleteByID(ctx context.Context, id primitive.ObjectID) error {
    _, err := r.collection.DeleteOne(ctx, primitive.M{"_id": id})
    return err
}

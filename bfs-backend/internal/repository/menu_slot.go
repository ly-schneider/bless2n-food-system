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

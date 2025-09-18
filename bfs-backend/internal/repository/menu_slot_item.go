package repository

import (
	"backend/internal/database"
	"backend/internal/domain"
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type MenuSlotItemRepository interface {
	FindByMenuSlotIDs(ctx context.Context, menuSlotIDs []primitive.ObjectID) ([]*domain.MenuSlotItem, error)
}

type menuSlotItemRepository struct {
	collection *mongo.Collection
}

func NewMenuSlotItemRepository(db *database.MongoDB) MenuSlotItemRepository {
	return &menuSlotItemRepository{
		collection: db.Database.Collection(database.MenuSlotItemsCollection),
	}
}

func (r *menuSlotItemRepository) FindByMenuSlotIDs(ctx context.Context, menuSlotIDs []primitive.ObjectID) (menuSlotItems []*domain.MenuSlotItem, err error) {
    cursor, err := r.collection.Find(ctx, primitive.M{"menu_slot_id": primitive.M{"$in": menuSlotIDs}})
    if err != nil {
        return nil, err
    }
    defer func() {
        if cerr := cursor.Close(ctx); err == nil && cerr != nil {
            err = cerr
        }
    }()

    for cursor.Next(ctx) {
        var menuSlotItem domain.MenuSlotItem
        if derr := cursor.Decode(&menuSlotItem); derr != nil {
            err = derr
            return nil, err
        }
        menuSlotItems = append(menuSlotItems, &menuSlotItem)
    }

    if derr := cursor.Err(); derr != nil {
        err = derr
        return nil, err
    }

    return menuSlotItems, nil
}

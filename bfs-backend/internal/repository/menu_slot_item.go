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

func (r *menuSlotItemRepository) FindByMenuSlotIDs(ctx context.Context, menuSlotIDs []primitive.ObjectID) ([]*domain.MenuSlotItem, error) {
	var menuSlotItems []*domain.MenuSlotItem

	cursor, err := r.collection.Find(ctx, primitive.M{"menu_slot_id": primitive.M{"$in": menuSlotIDs}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var menuSlotItem domain.MenuSlotItem
		if err := cursor.Decode(&menuSlotItem); err != nil {
			return nil, err
		}
		menuSlotItems = append(menuSlotItems, &menuSlotItem)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return menuSlotItems, nil
}

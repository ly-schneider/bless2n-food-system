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
    Insert(ctx context.Context, item *domain.MenuSlotItem) error
    ExistsBySlotAndProduct(ctx context.Context, slotID primitive.ObjectID, productID primitive.ObjectID) (bool, error)
    DeleteBySlotAndProduct(ctx context.Context, slotID primitive.ObjectID, productID primitive.ObjectID) (bool, error)
    DeleteByMenuSlotID(ctx context.Context, slotID primitive.ObjectID) error
    CountByProductID(ctx context.Context, productID primitive.ObjectID) (int64, error)
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

func (r *menuSlotItemRepository) Insert(ctx context.Context, item *domain.MenuSlotItem) error {
    if item.ID.IsZero() { item.ID = primitive.NewObjectID() }
    _, err := r.collection.InsertOne(ctx, item)
    return err
}

func (r *menuSlotItemRepository) ExistsBySlotAndProduct(ctx context.Context, slotID primitive.ObjectID, productID primitive.ObjectID) (bool, error) {
    n, err := r.collection.CountDocuments(ctx, primitive.M{"menu_slot_id": slotID, "product_id": productID})
    return n > 0, err
}

func (r *menuSlotItemRepository) DeleteBySlotAndProduct(ctx context.Context, slotID primitive.ObjectID, productID primitive.ObjectID) (bool, error) {
    res, err := r.collection.DeleteOne(ctx, primitive.M{"menu_slot_id": slotID, "product_id": productID})
    if err != nil { return false, err }
    return res.DeletedCount > 0, nil
}

func (r *menuSlotItemRepository) DeleteByMenuSlotID(ctx context.Context, slotID primitive.ObjectID) error {
    _, err := r.collection.DeleteMany(ctx, primitive.M{"menu_slot_id": slotID})
    return err
}

func (r *menuSlotItemRepository) CountByProductID(ctx context.Context, productID primitive.ObjectID) (int64, error) {
    return r.collection.CountDocuments(ctx, primitive.M{"product_id": productID})
}

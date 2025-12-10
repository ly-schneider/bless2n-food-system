package seed

import (
	"backend/internal/database"
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

type MenuSlotItemSeeder struct {
	Logger *zap.Logger
}

type MenuSlotItemDocument struct {
	ID         bson.ObjectID `bson:"_id"`
	MenuSlotID bson.ObjectID `bson:"menu_slot_id"`
	ProductID  bson.ObjectID `bson:"product_id"`
}

type menuSlotItemSeed struct {
	MenuName     string
	SlotName     string
	ProductNames []string
}

var menuSlotItemSeeds = []menuSlotItemSeed{
	{MenuName: "Menu Gross", SlotName: "Burger", ProductNames: []string{"Smash Burger", "Veggie Burger"}},
	{MenuName: "Menu Gross", SlotName: "Beilage", ProductNames: []string{"Pommes"}},
	{MenuName: "Menu Gross", SlotName: "Getränk", ProductNames: []string{"Coca Cola", "Ice Tea Lemon", "Red Bull", "El Tony Mate", "Wasser Prickelnd"}},
	{MenuName: "Menu Klein", SlotName: "Burger", ProductNames: []string{"Smash Burger", "Veggie Burger"}},
	{MenuName: "Menu Klein", SlotName: "Getränk", ProductNames: []string{"Coca Cola", "Ice Tea Lemon", "Red Bull", "El Tony Mate", "Wasser Prickelnd"}},
}

func NewMenuSlotItemSeeder(logger *zap.Logger) MenuSlotItemSeeder {
	return MenuSlotItemSeeder{Logger: logger}
}

func (s MenuSlotItemSeeder) Name() string {
	return "menu_slot_items"
}

func (s MenuSlotItemSeeder) Seed(ctx context.Context, db *mongo.Database) error {
	logger := loggerOrNop(s.Logger)
	coll := db.Collection(database.MenuSlotItemsCollection)

	productNames := collectMenuSlotProductNames()
	productIDs, err := productIDsByName(ctx, db, productNames)
	if err != nil {
		return err
	}
	for _, name := range productNames {
		if _, ok := productIDs[name]; !ok {
			return fmt.Errorf("product %s missing - seed products first", name)
		}
	}

	menuNames := collectMenuNamesForSlotItems()
	menuIDs, err := productIDsByName(ctx, db, menuNames)
	if err != nil {
		return err
	}
	for _, name := range menuNames {
		if _, ok := menuIDs[name]; !ok {
			return fmt.Errorf("menu product %s missing - seed menu slots first", name)
		}
	}

	slotIDs, err := slotIDsByMenu(ctx, db, menuIDs)
	if err != nil {
		return err
	}

	for _, seed := range menuSlotItemSeeds {
		slotID, ok := slotIDs[slotKey(seed.MenuName, seed.SlotName)]
		if !ok {
			return fmt.Errorf("menu slot %s/%s missing - seed menu slots first", seed.MenuName, seed.SlotName)
		}
		for _, productName := range seed.ProductNames {
			productID := productIDs[productName]
			doc := MenuSlotItemDocument{
				ID:         bson.NewObjectID(),
				MenuSlotID: slotID,
				ProductID:  productID,
			}

			filter := bson.M{
				"menu_slot_id": doc.MenuSlotID,
				"product_id":   doc.ProductID,
			}
			update := bson.M{
				"$setOnInsert": bson.M{"_id": doc.ID},
				"$set": bson.M{
					"menu_slot_id": doc.MenuSlotID,
					"product_id":   doc.ProductID,
				},
			}
			opts := options.UpdateOne().SetUpsert(true)
			if _, err := coll.UpdateOne(ctx, filter, update, opts); err != nil {
				return fmt.Errorf("upsert menu_slot_item %s/%s -> %s: %w", seed.MenuName, seed.SlotName, productName, err)
			}
		}
	}

	count, err := coll.CountDocuments(ctx, bson.D{})
	if err == nil {
		logger.Info("Menu slot items seeded", zap.Int64("count", count))
	}
	return nil
}

func collectMenuSlotProductNames() []string {
	seen := make(map[string]struct{})
	var names []string
	for _, seed := range menuSlotItemSeeds {
		for _, name := range seed.ProductNames {
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			names = append(names, name)
		}
	}
	return names
}

func collectMenuNamesForSlotItems() []string {
	seen := make(map[string]struct{})
	var names []string
	for _, seed := range menuSlotItemSeeds {
		if _, ok := seen[seed.MenuName]; ok {
			continue
		}
		seen[seed.MenuName] = struct{}{}
		names = append(names, seed.MenuName)
	}
	return names
}

func slotIDsByMenu(ctx context.Context, db *mongo.Database, menuIDs map[string]bson.ObjectID) (map[string]bson.ObjectID, error) {
	result := make(map[string]bson.ObjectID)
	if len(menuIDs) == 0 {
		return result, nil
	}

	var ids []bson.ObjectID
	reverseLookup := make(map[bson.ObjectID]string)
	for name, id := range menuIDs {
		ids = append(ids, id)
		reverseLookup[id] = name
	}

	coll := db.Collection(database.MenuSlotsCollection)
	cur, err := coll.Find(ctx, bson.M{"product_id": bson.M{"$in": ids}})
	if err != nil {
		return nil, fmt.Errorf("find menu slots: %w", err)
	}
	defer func() { _ = cur.Close(ctx) }()

	for cur.Next(ctx) {
		var slot struct {
			ID        bson.ObjectID `bson:"_id"`
			ProductID bson.ObjectID `bson:"product_id"`
			Name      string        `bson:"name"`
		}
		if err := cur.Decode(&slot); err != nil {
			return nil, fmt.Errorf("decode menu slot: %w", err)
		}
		menuName := reverseLookup[slot.ProductID]
		result[slotKey(menuName, slot.Name)] = slot.ID
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func slotKey(menuName, slotName string) string {
	return menuName + "|" + slotName
}

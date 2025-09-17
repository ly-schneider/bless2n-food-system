package dev

import (
	"context"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type MongoConfig struct {
	DatabaseName       string
	ResetBeforeSeeding bool
	BaselineDir        string // default: ./db/seeds/dev
}

// SeedMongo seeds the MongoDB database with baseline and generated data
func SeedMongo(ctx context.Context, client *mongo.Client, cfg MongoConfig, logger *zap.Logger, force bool) error {
	// Environment gating
	appEnv := os.Getenv("APP_ENV")
	if appEnv != "dev" && !force {
		return fmt.Errorf("seeding refused: APP_ENV=%s (use --force to override)", appEnv)
	}

	// Set default baseline directory if not provided
	if cfg.BaselineDir == "" {
		cfg.BaselineDir = "./db/seeds/dev"
	}

	logger.Info("Starting MongoDB seeding",
		zap.String("database", cfg.DatabaseName),
		zap.Bool("resetBeforeSeeding", cfg.ResetBeforeSeeding),
		zap.String("baselineDir", cfg.BaselineDir),
	)

	db := client.Database(cfg.DatabaseName)

	// Setup faker seed if specified
	if fakerSeed := os.Getenv("FAKER_SEED"); fakerSeed != "" {
		if seed, err := strconv.ParseInt(fakerSeed, 10, 64); err == nil {
			_ = gofakeit.Seed(seed)
			logger.Info("Using deterministic faker seed", zap.Int64("seed", seed))
		}
	}

	// Reset collections if requested
	if cfg.ResetBeforeSeeding {
		logger.Info("Resetting collections before seeding")
		if err := resetCollections(ctx, db, logger); err != nil {
			return fmt.Errorf("failed to reset collections: %w", err)
		}
	}

	// Ensure indexes
	if err := ensureIndexes(ctx, db, logger); err != nil {
		return fmt.Errorf("failed to ensure indexes: %w", err)
	}

	// Load baseline data
	if err := loadBaselineData(ctx, db, cfg.BaselineDir, logger); err != nil {
		return fmt.Errorf("failed to load baseline data: %w", err)
	}

	// Generate bulk data
	if err := generateBulkData(ctx, db, logger); err != nil {
		return fmt.Errorf("failed to generate bulk data: %w", err)
	}

	logger.Info("MongoDB seeding completed successfully")
	return nil
}

func resetCollections(ctx context.Context, db *mongo.Database, logger *zap.Logger) error {
	collections := []string{
		"users", "admin_invites", "otp_tokens", "refresh_tokens",
		"stations", "device_requests", "categories", "products",
		"menu_slots", "menu_slot_items", "station_products",
		"orders", "order_items", "inventory_ledger",
	}

	for _, coll := range collections {
		collection := db.Collection(coll)
		result, err := collection.DeleteMany(ctx, bson.D{})
		if err != nil {
			return fmt.Errorf("failed to reset collection %s: %w", coll, err)
		}
		logger.Info("Reset collection", zap.String("collection", coll), zap.Int64("deleted", result.DeletedCount))
	}

	return nil
}

func ensureIndexes(ctx context.Context, db *mongo.Database, logger *zap.Logger) error {
	// Users collection indexes
	usersCollection := db.Collection("users")
	userIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "role", Value: 1}}},
		{Keys: bson.D{{Key: "is_verified", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: 1}}},
	}

	if _, err := usersCollection.Indexes().CreateMany(ctx, userIndexes); err != nil {
		return fmt.Errorf("failed to create users indexes: %w", err)
	}

	// Orders collection indexes
	ordersCollection := db.Collection("orders")
	orderIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "customer_id", Value: 1}}},
		{Keys: bson.D{{Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: 1}}},
	}

	if _, err := ordersCollection.Indexes().CreateMany(ctx, orderIndexes); err != nil {
		return fmt.Errorf("failed to create orders indexes: %w", err)
	}

	// Categories collection indexes
	categoriesCollection := db.Collection("categories")
	categoryIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "name", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "is_active", Value: 1}}},
	}

	if _, err := categoriesCollection.Indexes().CreateMany(ctx, categoryIndexes); err != nil {
		return fmt.Errorf("failed to create categories indexes: %w", err)
	}

	// Products collection indexes
	productsCollection := db.Collection("products")
	productIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "name", Value: "text"}}},
		{Keys: bson.D{{Key: "category_id", Value: 1}}},
		{Keys: bson.D{{Key: "is_active", Value: 1}}},
		{Keys: bson.D{{Key: "type", Value: 1}}},
	}

	if _, err := productsCollection.Indexes().CreateMany(ctx, productIndexes); err != nil {
		return fmt.Errorf("failed to create products indexes: %w", err)
	}

	// Order items collection indexes
	orderItemsCollection := db.Collection("order_items")
	orderItemIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "order_id", Value: 1}}},
		{Keys: bson.D{{Key: "product_id", Value: 1}}},
		{Keys: bson.D{{Key: "parent_item_id", Value: 1}}},
		{Keys: bson.D{{Key: "is_redeemed", Value: 1}}},
	}

	if _, err := orderItemsCollection.Indexes().CreateMany(ctx, orderItemIndexes); err != nil {
		return fmt.Errorf("failed to create order_items indexes: %w", err)
	}

	// Stations collection indexes
	stationsCollection := db.Collection("stations")
	stationIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "name", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: 1}}},
	}

	if _, err := stationsCollection.Indexes().CreateMany(ctx, stationIndexes); err != nil {
		return fmt.Errorf("failed to create stations indexes: %w", err)
	}

	// Menu slots collection indexes
	menuSlotsCollection := db.Collection("menu_slots")
	menuSlotIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "product_id", Value: 1}}},
		{Keys: bson.D{{Key: "sequence", Value: 1}}},
	}

	if _, err := menuSlotsCollection.Indexes().CreateMany(ctx, menuSlotIndexes); err != nil {
		return fmt.Errorf("failed to create menu_slots indexes: %w", err)
	}

	// Menu slot items collection indexes
	menuSlotItemsCollection := db.Collection("menu_slot_items")
	menuSlotItemIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "menu_slot_id", Value: 1}}},
		{Keys: bson.D{{Key: "product_id", Value: 1}}},
	}

	if _, err := menuSlotItemsCollection.Indexes().CreateMany(ctx, menuSlotItemIndexes); err != nil {
		return fmt.Errorf("failed to create menu_slot_items indexes: %w", err)
	}

	// Station products collection indexes
	stationProductsCollection := db.Collection("station_products")
	stationProductIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "station_id", Value: 1}}},
		{Keys: bson.D{{Key: "product_id", Value: 1}}},
	}

	if _, err := stationProductsCollection.Indexes().CreateMany(ctx, stationProductIndexes); err != nil {
		return fmt.Errorf("failed to create station_products indexes: %w", err)
	}

	logger.Info("Indexes ensured successfully")
	return nil
}

func loadBaselineData(ctx context.Context, db *mongo.Database, baselineDir string, logger *zap.Logger) error {
	// Skip baseline data loading - we generate users programmatically
	logger.Info("Skipping baseline data loading - users are generated programmatically")
	return nil
}

func generateBulkData(ctx context.Context, db *mongo.Database, logger *zap.Logger) error {
	// Generate healthy amounts of data:
	// 20 customers, 3 admins, 10 simple products, 2 menu products with slots, 3 stations, 0 orders

	// Generate customers
	if err := generateCustomers(ctx, db, logger); err != nil {
		return fmt.Errorf("failed to generate customers: %w", err)
	}

	// Generate admins
	if err := generateAdmins(ctx, db, logger); err != nil {
		return fmt.Errorf("failed to generate admins: %w", err)
	}

	// Generate categories first
	if err := generateCategories(ctx, db, logger); err != nil {
		return fmt.Errorf("failed to generate categories: %w", err)
	}

	// Generate simple products
	if err := generateSimpleProducts(ctx, db, logger); err != nil {
		return fmt.Errorf("failed to generate simple products: %w", err)
	}

	// Generate menu products
	if err := generateMenuProducts(ctx, db, logger); err != nil {
		return fmt.Errorf("failed to generate menu products: %w", err)
	}

	// Generate menu slots
	if err := generateMenuSlots(ctx, db, logger); err != nil {
		return fmt.Errorf("failed to generate menu slots: %w", err)
	}

	// Generate menu slot items
	if err := generateMenuSlotItems(ctx, db, logger); err != nil {
		return fmt.Errorf("failed to generate menu slot items: %w", err)
	}

	// Generate stations
	if err := generateStations(ctx, db, logger); err != nil {
		return fmt.Errorf("failed to generate stations: %w", err)
	}

	// Generate station products
	if err := generateStationProducts(ctx, db, logger); err != nil {
		return fmt.Errorf("failed to generate station products: %w", err)
	}

	// Skip orders generation (0 orders as specified)
	logger.Info("Skipping orders generation as requested (0 orders)")

	return nil
}

func generateCustomers(ctx context.Context, db *mongo.Database, logger *zap.Logger) error {
	collection := db.Collection("users")
	customerCount := 20

	var operations []mongo.WriteModel

	for i := 0; i < customerCount; i++ {
		now := time.Now()
		customer := map[string]interface{}{
			"_id":         primitive.NewObjectID(),
			"email":       gofakeit.Email(),
			"firstname":   gofakeit.FirstName(),
			"lastname":    gofakeit.LastName(),
			"role":        "customer",
			"is_verified": gofakeit.Bool(),
			"is_disabled": false,
			"created_at":  now.Add(-time.Duration(gofakeit.Number(0, 365*24)) * time.Hour),
			"updated_at":  now,
		}

		// Use upsert for idempotency
		filter := bson.D{{Key: "_id", Value: customer["_id"]}}
		update := bson.D{{Key: "$setOnInsert", Value: customer}}
		operations = append(operations, mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true))
	}

	if len(operations) > 0 {
		result, err := collection.BulkWrite(ctx, operations)
		if err != nil {
			return fmt.Errorf("failed to insert customers: %w", err)
		}

		logger.Info("Generated customers",
			zap.Int("count", customerCount),
			zap.Int64("inserted", result.InsertedCount),
			zap.Int64("upserted", result.UpsertedCount),
		)
	}

	return nil
}

func generateAdmins(ctx context.Context, db *mongo.Database, logger *zap.Logger) error {
	collection := db.Collection("users")
	adminCount := 3

	var operations []mongo.WriteModel

	for i := 0; i < adminCount; i++ {
		now := time.Now()
		admin := map[string]interface{}{
			"_id":         primitive.NewObjectID(),
			"email":       fmt.Sprintf("admin%d@dev.local", i+1),
			"firstname":   gofakeit.FirstName(),
			"lastname":    gofakeit.LastName(),
			"role":        "admin",
			"is_verified": true,
			"is_disabled": false,
			"created_at":  now.Add(-time.Duration(gofakeit.Number(0, 180*24)) * time.Hour),
			"updated_at":  now,
		}

		// Use upsert for idempotency
		filter := bson.D{{Key: "_id", Value: admin["_id"]}}
		update := bson.D{{Key: "$setOnInsert", Value: admin}}
		operations = append(operations, mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true))
	}

	if len(operations) > 0 {
		result, err := collection.BulkWrite(ctx, operations)
		if err != nil {
			return fmt.Errorf("failed to insert admins: %w", err)
		}

		logger.Info("Generated admins",
			zap.Int("count", adminCount),
			zap.Int64("inserted", result.InsertedCount),
			zap.Int64("upserted", result.UpsertedCount),
		)
	}

	return nil
}

// roundToHalf rounds a price to the nearest 0.5
func roundToHalf(price float64) float64 {
	return math.Round(price*2) / 2
}

func generateCategories(ctx context.Context, db *mongo.Database, logger *zap.Logger) error {
	collection := db.Collection("categories")

	categories := []string{"Beverages", "Snacks", "Meals", "Desserts", "Sides"}
	var operations []mongo.WriteModel

	for _, categoryName := range categories {
		now := time.Now()
		category := map[string]interface{}{
			"_id":        primitive.NewObjectID(),
			"name":       categoryName,
			"is_active":  true,
			"created_at": now.Add(-time.Duration(gofakeit.Number(0, 180*24)) * time.Hour),
			"updated_at": now,
		}

		// Use upsert for idempotency - match by name
		filter := bson.D{{Key: "name", Value: categoryName}}
		update := bson.D{{Key: "$setOnInsert", Value: category}}
		operations = append(operations, mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true))
	}

	if len(operations) > 0 {
		result, err := collection.BulkWrite(ctx, operations)
		if err != nil {
			return fmt.Errorf("failed to insert categories: %w", err)
		}

		logger.Info("Generated categories",
			zap.Int("count", len(categories)),
			zap.Int64("inserted", result.InsertedCount),
			zap.Int64("upserted", result.UpsertedCount),
		)
	}

	return nil
}

func generateSimpleProducts(ctx context.Context, db *mongo.Database, logger *zap.Logger) error {
	productsCollection := db.Collection("products")
	categoriesCollection := db.Collection("categories")

	// Get all categories to reference
	cursor, err := categoriesCollection.Find(ctx, bson.D{{Key: "is_active", Value: true}})
	if err != nil {
		return fmt.Errorf("failed to find categories: %w", err)
	}
	defer func() { _ = cursor.Close(ctx) }()

	var categories []struct {
		ID   primitive.ObjectID `bson:"_id"`
		Name string             `bson:"name"`
	}
	if err := cursor.All(ctx, &categories); err != nil {
		return fmt.Errorf("failed to decode categories: %w", err)
	}

	if len(categories) == 0 {
		return fmt.Errorf("no categories found - run generateCategories first")
	}

	productCount := 10
	var operations []mongo.WriteModel

	// Product data with short descriptions
	productData := []struct {
		name, description  string
		priceMin, priceMax float64
	}{
		{"Classic Burger", "Beef patty with lettuce and tomato", 8.0, 12.0},
		{"Chicken Wings", "Crispy buffalo wings", 6.0, 10.0},
		{"Caesar Salad", "Fresh romaine with caesar dressing", 7.0, 11.0},
		{"Margherita Pizza", "Tomato, mozzarella, fresh basil", 12.0, 18.0},
		{"Fish Tacos", "Grilled fish with cabbage slaw", 9.0, 13.0},
		{"Chocolate Cake", "Rich chocolate layer cake", 5.0, 8.0},
		{"Iced Coffee", "Cold brew with ice", 3.0, 5.0},
		{"French Fries", "Crispy golden fries", 4.0, 6.0},
		{"Nachos", "Tortilla chips with cheese", 6.0, 9.0},
		{"Smoothie Bowl", "Mixed berries with granola", 8.0, 12.0},
	}

	for i := 0; i < productCount; i++ {
		now := time.Now()
		data := productData[i%len(productData)]
		categoryIndex := i % len(categories)

		rawPrice := gofakeit.Float64Range(data.priceMin, data.priceMax)
		roundedPrice := roundToHalf(rawPrice)
		priceCents := int(roundedPrice * 100)

		product := map[string]interface{}{
			"_id":          primitive.NewObjectID(),
			"category_id":  categories[categoryIndex].ID,
			"type":         "simple",
			"name":         data.name,
			"price_cents":  priceCents,
			"is_active":    gofakeit.Bool(),
			"created_at":   now.Add(-time.Duration(gofakeit.Number(0, 180*24)) * time.Hour),
			"updated_at":   now,
		}

		// Use upsert for idempotency
		filter := bson.D{{Key: "_id", Value: product["_id"]}}
		update := bson.D{{Key: "$setOnInsert", Value: product}}
		operations = append(operations, mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true))
	}

	if len(operations) > 0 {
		result, err := productsCollection.BulkWrite(ctx, operations)
		if err != nil {
			return fmt.Errorf("failed to insert products: %w", err)
		}

		logger.Info("Generated simple products",
			zap.Int("count", productCount),
			zap.Int64("inserted", result.InsertedCount),
			zap.Int64("upserted", result.UpsertedCount),
		)
	}

	return nil
}

func generateMenuProducts(ctx context.Context, db *mongo.Database, logger *zap.Logger) error {
	productsCollection := db.Collection("products")
	categoriesCollection := db.Collection("categories")

	// Get Meals category for bundles
	var mealsCategory struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	err := categoriesCollection.FindOne(ctx, bson.D{{Key: "name", Value: "Meals"}}).Decode(&mealsCategory)
	if err != nil {
		return fmt.Errorf("failed to find Meals category: %w", err)
	}

	menuCount := 2
	var operations []mongo.WriteModel

	// Menu data with short descriptions
	menuData := []struct {
		name  string
		price float64
	}{
		{"Menu Big", 15.50},
		{"Menu Small", 12.00},
	}

	for i := 0; i < menuCount; i++ {
		now := time.Now()
		data := menuData[i]
		priceCents := int(data.price * 100)

		menuProduct := map[string]interface{}{
			"_id":          primitive.NewObjectID(),
			"category_id":  mealsCategory.ID,
			"type":         "menu",
			"name":         data.name,
			"price_cents":  priceCents,
			"is_active":    true,
			"created_at":   now.Add(-time.Duration(gofakeit.Number(0, 90*24)) * time.Hour),
			"updated_at":   now,
		}

		// Use upsert for idempotency
		filter := bson.D{{Key: "_id", Value: menuProduct["_id"]}}
		update := bson.D{{Key: "$setOnInsert", Value: menuProduct}}
		operations = append(operations, mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true))
	}

	if len(operations) > 0 {
		result, err := productsCollection.BulkWrite(ctx, operations)
		if err != nil {
			return fmt.Errorf("failed to insert bundle products: %w", err)
		}

		logger.Info("Generated menu products",
			zap.Int("count", menuCount),
			zap.Int64("inserted", result.InsertedCount),
			zap.Int64("upserted", result.UpsertedCount),
		)
	}

	return nil
}

func generateStations(ctx context.Context, db *mongo.Database, logger *zap.Logger) error {
	collection := db.Collection("stations")
	stationCount := 3

	var operations []mongo.WriteModel
	stationNames := []string{"North Station", "South Station", "Central Station"}

	for i := 0; i < stationCount; i++ {
		now := time.Now()
		station := map[string]interface{}{
			"_id":        primitive.NewObjectID(),
			"name":       stationNames[i],
			"created_at": now.Add(-time.Duration(gofakeit.Number(0, 60*24)) * time.Hour),
			"updated_at": now,
		}

		// Use upsert for idempotency
		filter := bson.D{{Key: "_id", Value: station["_id"]}}
		update := bson.D{{Key: "$setOnInsert", Value: station}}
		operations = append(operations, mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true))
	}

	if len(operations) > 0 {
		result, err := collection.BulkWrite(ctx, operations)
		if err != nil {
			return fmt.Errorf("failed to insert stations: %w", err)
		}

		logger.Info("Generated stations",
			zap.Int("count", stationCount),
			zap.Int64("inserted", result.InsertedCount),
			zap.Int64("upserted", result.UpsertedCount),
		)
	}

	return nil
}

func generateMenuSlots(ctx context.Context, db *mongo.Database, logger *zap.Logger) error {
	productsCollection := db.Collection("products")
	menuSlotsCollection := db.Collection("menu_slots")

	// Get all menu products
	menuCursor, err := productsCollection.Find(ctx, bson.D{{Key: "type", Value: "menu"}})
	if err != nil {
		return fmt.Errorf("failed to find menu products: %w", err)
	}
	defer func() { _ = menuCursor.Close(ctx) }()

	var menuProducts []struct {
		ID   primitive.ObjectID `bson:"_id"`
		Name string             `bson:"name"`
	}
	if err := menuCursor.All(ctx, &menuProducts); err != nil {
		return fmt.Errorf("failed to decode menu products: %w", err)
	}

	var operations []mongo.WriteModel
	slotNames := []string{"Burger", "Fries", "Drink"}

	// Create slots for each menu
	for _, menu := range menuProducts {
		for i, slotName := range slotNames {
			menuSlot := map[string]interface{}{
				"_id":        primitive.NewObjectID(),
				"product_id": menu.ID,
				"name":       slotName,
				"sequence":   i + 1,
			}

			// Use upsert for idempotency
			filter := bson.D{{Key: "_id", Value: menuSlot["_id"]}}
			update := bson.D{{Key: "$setOnInsert", Value: menuSlot}}
			operations = append(operations, mongo.NewUpdateOneModel().
				SetFilter(filter).
				SetUpdate(update).
				SetUpsert(true))
		}
	}

	if len(operations) > 0 {
		result, err := menuSlotsCollection.BulkWrite(ctx, operations)
		if err != nil {
			return fmt.Errorf("failed to insert menu slots: %w", err)
		}

		logger.Info("Generated menu slots",
			zap.Int("menus", len(menuProducts)),
			zap.Int("slots_per_menu", len(slotNames)),
			zap.Int64("slots_inserted", result.InsertedCount),
			zap.Int64("slots_upserted", result.UpsertedCount),
		)
	}

	return nil
}

func generateMenuSlotItems(ctx context.Context, db *mongo.Database, logger *zap.Logger) error {
	menuSlotsCollection := db.Collection("menu_slots")
	productsCollection := db.Collection("products")
	menuSlotItemsCollection := db.Collection("menu_slot_items")

	// Get all menu slots
	slotsCursor, err := menuSlotsCollection.Find(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("failed to find menu slots: %w", err)
	}
	defer func() { _ = slotsCursor.Close(ctx) }()

	var menuSlots []struct {
		ID   primitive.ObjectID `bson:"_id"`
		Name string             `bson:"name"`
	}
	if err := slotsCursor.All(ctx, &menuSlots); err != nil {
		return fmt.Errorf("failed to decode menu slots: %w", err)
	}

	// Get all simple products to use as options
	simpleCursor, err := productsCollection.Find(ctx, bson.D{{Key: "type", Value: "simple"}})
	if err != nil {
		return fmt.Errorf("failed to find simple products: %w", err)
	}
	defer func() { _ = simpleCursor.Close(ctx) }()

	var simpleProducts []struct {
		ID   primitive.ObjectID `bson:"_id"`
		Name string             `bson:"name"`
	}
	if err := simpleCursor.All(ctx, &simpleProducts); err != nil {
		return fmt.Errorf("failed to decode simple products: %w", err)
	}

	if len(simpleProducts) == 0 {
		return fmt.Errorf("no simple products found to create menu slot items")
	}

	var operations []mongo.WriteModel

	// Add 2-3 simple products as options for each slot
	for _, slot := range menuSlots {
		optionsCount := 2 + (len(slot.ID.Hex()) % 2) // 2 or 3 options per slot
		for i := 0; i < optionsCount && i < len(simpleProducts); i++ {
			slotItem := map[string]interface{}{
				"_id":          primitive.NewObjectID(),
				"menu_slot_id": slot.ID,
				"product_id":   simpleProducts[i].ID,
			}

			// Use upsert for idempotency
			filter := bson.D{{Key: "_id", Value: slotItem["_id"]}}
			update := bson.D{{Key: "$setOnInsert", Value: slotItem}}
			operations = append(operations, mongo.NewUpdateOneModel().
				SetFilter(filter).
				SetUpdate(update).
				SetUpsert(true))
		}
	}

	if len(operations) > 0 {
		result, err := menuSlotItemsCollection.BulkWrite(ctx, operations)
		if err != nil {
			return fmt.Errorf("failed to insert menu slot items: %w", err)
		}

		logger.Info("Generated menu slot items",
			zap.Int("slots", len(menuSlots)),
			zap.Int64("items_inserted", result.InsertedCount),
			zap.Int64("items_upserted", result.UpsertedCount),
		)
	}

	return nil
}

func generateStationProducts(ctx context.Context, db *mongo.Database, logger *zap.Logger) error {
	stationsCollection := db.Collection("stations")
	productsCollection := db.Collection("products")
	stationProductsCollection := db.Collection("station_products")

	// Get all stations
	stationsCursor, err := stationsCollection.Find(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("failed to find stations: %w", err)
	}
	defer func() { _ = stationsCursor.Close(ctx) }()

	var stations []struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	if err := stationsCursor.All(ctx, &stations); err != nil {
		return fmt.Errorf("failed to decode stations: %w", err)
	}

	// Get all simple products (only simple products can be redeemed at stations)
	simpleCursor, err := productsCollection.Find(ctx, bson.D{{Key: "type", Value: "simple"}})
	if err != nil {
		return fmt.Errorf("failed to find simple products: %w", err)
	}
	defer func() { _ = simpleCursor.Close(ctx) }()

	var simpleProducts []struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	if err := simpleCursor.All(ctx, &simpleProducts); err != nil {
		return fmt.Errorf("failed to decode simple products: %w", err)
	}

	var operations []mongo.WriteModel

	// Each station can redeem all simple products
	for _, station := range stations {
		for _, product := range simpleProducts {
			stationProduct := map[string]interface{}{
				"_id":        primitive.NewObjectID(),
				"station_id": station.ID,
				"product_id": product.ID,
			}

			// Use upsert for idempotency
			filter := bson.D{{Key: "_id", Value: stationProduct["_id"]}}
			update := bson.D{{Key: "$setOnInsert", Value: stationProduct}}
			operations = append(operations, mongo.NewUpdateOneModel().
				SetFilter(filter).
				SetUpdate(update).
				SetUpsert(true))
		}
	}

	if len(operations) > 0 {
		result, err := stationProductsCollection.BulkWrite(ctx, operations)
		if err != nil {
			return fmt.Errorf("failed to insert station products: %w", err)
		}

		logger.Info("Generated station products",
			zap.Int("stations", len(stations)),
			zap.Int("products_per_station", len(simpleProducts)),
			zap.Int64("relations_inserted", result.InsertedCount),
			zap.Int64("relations_upserted", result.UpsertedCount),
		)
	}

	return nil
}

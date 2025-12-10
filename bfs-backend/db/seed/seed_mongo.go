package seed

import (
	"backend/internal/database"
	"backend/internal/domain"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

type MongoConfig struct {
	DatabaseName       string
	ResetBeforeSeeding bool
	BaselineDir        string // default: ./db/seed
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
		cfg.BaselineDir = "./db/seed"
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

	// Reset database if requested (drop entire DB)
	if cfg.ResetBeforeSeeding {
		logger.Info("Dropping database before seeding", zap.String("database", cfg.DatabaseName))
		if err := dropDatabase(ctx, db, logger); err != nil {
			return fmt.Errorf("failed to drop database: %w", err)
		}
		// Recreate handle after drop
		db = client.Database(cfg.DatabaseName)
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

func dropDatabase(ctx context.Context, db *mongo.Database, logger *zap.Logger) error {
	if err := db.Drop(ctx); err != nil {
		return err
	}
	logger.Info("Dropped database successfully")
	return nil
}

func ensureIndexes(ctx context.Context, db *mongo.Database, logger *zap.Logger) error {
	// Ensure all collections exist even without data
	collections := []string{
		database.UsersCollection,
		database.IdentityLinksCollection,
		database.OrdersCollection,
		database.OrderItemsCollection,
		database.CategoriesCollection,
		database.ProductsCollection,
		database.MenuSlotsCollection,
		database.MenuSlotItemsCollection,
		database.InventoryLedgerCollection,
		database.AdminInvitesCollection,
		database.OTPTokensCollection,
		database.EmailChangeTokensCollection,
		database.RefreshTokensCollection,
		database.StationsCollection,
		database.StationProductsCollection,
		database.AuditLogsCollection,
		database.PosDevicesCollection,
		database.JetonsCollection,
		database.PosSettingsCollection,
	}
	for _, name := range collections {
		if err := ensureCollectionExists(ctx, db, name, logger); err != nil {
			return fmt.Errorf("ensure collection %s: %w", name, err)
		}
	}
	// Users collection indexes
	usersCollection := db.Collection(database.UsersCollection)
	userIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "role", Value: 1}}},
		{Keys: bson.D{{Key: "is_verified", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: 1}}},
	}

	if _, err := usersCollection.Indexes().CreateMany(ctx, userIndexes); err != nil {
		return fmt.Errorf("failed to create users indexes: %w", err)
	}

	// Identity links collection indexes
	identityLinks := db.Collection(database.IdentityLinksCollection)
	identityIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "provider", Value: 1}, {Key: "provider_user_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "user_id", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: 1}}},
	}
	if _, err := identityLinks.Indexes().CreateMany(ctx, identityIndexes); err != nil {
		return fmt.Errorf("failed to create identity_links indexes: %w", err)
	}

	// Orders collection indexes
	ordersCollection := db.Collection(database.OrdersCollection)
	orderIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "customer_id", Value: 1}}},
		{Keys: bson.D{{Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: 1}}},
	}

	if _, err := ordersCollection.Indexes().CreateMany(ctx, orderIndexes); err != nil {
		return fmt.Errorf("failed to create orders indexes: %w", err)
	}

	// Categories collection indexes
	categoriesCollection := db.Collection(database.CategoriesCollection)
	categoryIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "name", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "is_active", Value: 1}}},
		{Keys: bson.D{{Key: "position", Value: 1}}},
		{Keys: bson.D{{Key: "position", Value: 1}, {Key: "name", Value: 1}}},
	}

	if _, err := categoriesCollection.Indexes().CreateMany(ctx, categoryIndexes); err != nil {
		return fmt.Errorf("failed to create categories indexes: %w", err)
	}

	// Products collection indexes
	productsCollection := db.Collection(database.ProductsCollection)
	// Cosmos Mongo API does not support text indexes; use ascending index on name for compatibility
	productIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "name", Value: 1}}},
		{Keys: bson.D{{Key: "category_id", Value: 1}}},
		{Keys: bson.D{{Key: "is_active", Value: 1}}},
		{Keys: bson.D{{Key: "type", Value: 1}}},
	}

	if _, err := productsCollection.Indexes().CreateMany(ctx, productIndexes); err != nil {
		return fmt.Errorf("failed to create products indexes: %w", err)
	}

	// Order items collection indexes
	orderItemsCollection := db.Collection(database.OrderItemsCollection)
	orderItemIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "order_id", Value: 1}}},
		{Keys: bson.D{{Key: "product_id", Value: 1}}},
		{Keys: bson.D{{Key: "parent_item_id", Value: 1}}},
		{Keys: bson.D{{Key: "is_redeemed", Value: 1}}},
	}

	if _, err := orderItemsCollection.Indexes().CreateMany(ctx, orderItemIndexes); err != nil {
		return fmt.Errorf("failed to create order_items indexes: %w", err)
	}

	// Menu slots collection indexes
	menuSlotsCollection := db.Collection(database.MenuSlotsCollection)
	menuSlotIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "product_id", Value: 1}}},
		{Keys: bson.D{{Key: "sequence", Value: 1}}},
	}

	if _, err := menuSlotsCollection.Indexes().CreateMany(ctx, menuSlotIndexes); err != nil {
		return fmt.Errorf("failed to create menu_slots indexes: %w", err)
	}

	// Menu slot items collection indexes
	menuSlotItemsCollection := db.Collection(database.MenuSlotItemsCollection)
	menuSlotItemIndexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "menu_slot_id", Value: 1}}},
		{Keys: bson.D{{Key: "product_id", Value: 1}}},
	}

	if _, err := menuSlotItemsCollection.Indexes().CreateMany(ctx, menuSlotItemIndexes); err != nil {
		return fmt.Errorf("failed to create menu_slot_items indexes: %w", err)
	}

	// Inventory ledger indexes
	if err := ensureInventoryIndexes(ctx, db, logger); err != nil {
		return err
	}

	// Auth-related collections
	// OTP tokens: user lookup, expiry, and optional created_at for diagnostics; TTL on expires_at for cleanup
	otpTokens := db.Collection(database.OTPTokensCollection)
	otpIdx := []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}}},
		{Keys: bson.D{{Key: "expires_at", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(0)},
		{Keys: bson.D{{Key: "created_at", Value: 1}}},
	}
	if _, err := otpTokens.Indexes().CreateMany(ctx, otpIdx); err != nil {
		return fmt.Errorf("failed to create otp_tokens indexes: %w", err)
	}

	// Email change tokens: by user, new_email, expiry; TTL on expires_at
	emailChange := db.Collection(database.EmailChangeTokensCollection)
	ecIdx := []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}}},
		{Keys: bson.D{{Key: "new_email", Value: 1}}},
		{Keys: bson.D{{Key: "expires_at", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(0)},
		{Keys: bson.D{{Key: "created_at", Value: 1}}},
	}
	if _, err := emailChange.Indexes().CreateMany(ctx, ecIdx); err != nil {
		return fmt.Errorf("failed to create email_change_tokens indexes: %w", err)
	}

	// Refresh tokens: unique token_hash; user sessions listing and family revocation
	refreshTokens := db.Collection(database.RefreshTokensCollection)
	rtIdx := []mongo.IndexModel{
		{Keys: bson.D{{Key: "token_hash", Value: 1}}},
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "is_revoked", Value: 1}, {Key: "expires_at", Value: 1}}},
		{Keys: bson.D{{Key: "family_id", Value: 1}, {Key: "is_revoked", Value: 1}}},
		{Keys: bson.D{{Key: "last_used_at", Value: 1}}},
	}
	// token_hash should be unique; create separately to ensure provider compat
	uniq := options.Index().SetUnique(true)
	rtIdx[0].Options = uniq
	if _, err := refreshTokens.Indexes().CreateMany(ctx, rtIdx); err != nil {
		return fmt.Errorf("failed to create refresh_tokens indexes: %w", err)
	}

	// Admin invites: token_hash unique, invitee_email lookup, status, expiry; TTL on expires_at for cleanup
	adminInvites := db.Collection(database.AdminInvitesCollection)
	aiIdx := []mongo.IndexModel{
		{Keys: bson.D{{Key: "token_hash", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "invitee_email", Value: 1}}},
		{Keys: bson.D{{Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "expires_at", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(0)},
		{Keys: bson.D{{Key: "created_at", Value: 1}}},
	}
	if _, err := adminInvites.Indexes().CreateMany(ctx, aiIdx); err != nil {
		return fmt.Errorf("failed to create %s indexes: %w", database.AdminInvitesCollection, err)
	}

	// Stations: device_key unique, approved, created_at
	stations := db.Collection(database.StationsCollection)
	stIdx := []mongo.IndexModel{
		{Keys: bson.D{{Key: "device_key", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "approved", Value: 1}}},
		{Keys: bson.D{{Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: 1}}},
	}
	if _, err := stations.Indexes().CreateMany(ctx, stIdx); err != nil {
		return fmt.Errorf("failed to create %s indexes: %w", database.StationsCollection, err)
	}

	// Station products: unique (station_id, product_id)
	stationProducts := db.Collection(database.StationProductsCollection)
	spIdx := []mongo.IndexModel{
		{Keys: bson.D{{Key: "station_id", Value: 1}, {Key: "product_id", Value: 1}}, Options: options.Index().SetUnique(true)},
	}
	if _, err := stationProducts.Indexes().CreateMany(ctx, spIdx); err != nil {
		return fmt.Errorf("failed to create %s indexes: %w", database.StationProductsCollection, err)
	}

	// POS devices: device_token unique, approved, created_at
	posDevices := db.Collection(database.PosDevicesCollection)
	pdIdx := []mongo.IndexModel{
		{Keys: bson.D{{Key: "device_token", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "approved", Value: 1}}},
		{Keys: bson.D{{Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: 1}}},
	}
	if _, err := posDevices.Indexes().CreateMany(ctx, pdIdx); err != nil {
		return fmt.Errorf("failed to create %s indexes: %w", database.PosDevicesCollection, err)
	}

	// Jetons: unique name for clarity
	jetons := db.Collection(database.JetonsCollection)
	jetonIdx := []mongo.IndexModel{
		{Keys: bson.D{{Key: "name", Value: 1}}, Options: options.Index().SetUnique(true)},
	}
	if _, err := jetons.Indexes().CreateMany(ctx, jetonIdx); err != nil {
		return fmt.Errorf("failed to create %s indexes: %w", database.JetonsCollection, err)
	}

	// Audit logs: entity, actor, created_at
	audit := db.Collection(database.AuditLogsCollection)
	auIdx := []mongo.IndexModel{
		{Keys: bson.D{{Key: "entity_type", Value: 1}, {Key: "entity_id", Value: 1}}},
		{Keys: bson.D{{Key: "actor_user_id", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: 1}}},
	}
	if _, err := audit.Indexes().CreateMany(ctx, auIdx); err != nil {
		return fmt.Errorf("failed to create %s indexes: %w", database.AuditLogsCollection, err)
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
	// 2 admins, 10 simple products, 2 menu products with slots, 0 orders

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

	// Seed jetons and assign to products (POS stays in QR_CODE mode by default)
	if err := seedJetonsAndAssignProducts(ctx, db, logger); err != nil {
		return fmt.Errorf("failed to seed jetons: %w", err)
	}

	if err := seedPosSettings(ctx, db, logger); err != nil {
		return fmt.Errorf("failed to seed POS settings: %w", err)
	}

	// Seed inventory opening balance: +50 for each simple product
	if err := seedInventoryOpeningBalance(ctx, db, logger, 50); err != nil {
		return fmt.Errorf("failed to seed inventory opening balance: %w", err)
	}

	// Skip orders generation (0 orders as specified)
	logger.Info("Skipping orders generation as requested (0 orders)")

	return nil
}

// Additional indexes for inventory ledger
func ensureInventoryIndexes(ctx context.Context, db *mongo.Database, logger *zap.Logger) error {
	coll := db.Collection(database.InventoryLedgerCollection)
	idx := []mongo.IndexModel{
		{Keys: bson.D{{Key: "product_id", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: 1}}},
		{Keys: bson.D{{Key: "product_id", Value: 1}, {Key: "created_at", Value: -1}}},
	}
	if _, err := coll.Indexes().CreateMany(ctx, idx); err != nil {
		return fmt.Errorf("failed to create inventory_ledger indexes: %w", err)
	}
	return nil
}

// ensureCollectionExists attempts to create a collection; if it exists, ignore the error
func ensureCollectionExists(ctx context.Context, db *mongo.Database, name string, logger *zap.Logger) error {
	if err := db.CreateCollection(ctx, name); err != nil {
		if ce, ok := err.(mongo.CommandError); ok && ce.Code == 48 { // NamespaceExists
			return nil
		}
		msg := err.Error()
		if strings.Contains(msg, "NamespaceExists") || strings.Contains(msg, "already exists") {
			return nil
		}
		return err
	}
	logger.Info("Created collection", zap.String("name", name))
	return nil
}

func seedInventoryOpeningBalance(ctx context.Context, db *mongo.Database, logger *zap.Logger, qty int) error {
	if qty <= 0 {
		return nil
	}
	products := db.Collection("products")
	ledger := db.Collection("inventory_ledger")
	cur, err := products.Find(ctx, bson.M{"type": "simple"})
	if err != nil {
		return err
	}
	defer func() { _ = cur.Close(ctx) }()
	type production struct {
		ID bson.ObjectID `bson:"_id"`
	}
	entries := make([]interface{}, 0)
	now := time.Now().UTC()
	for cur.Next(ctx) {
		var p production
		if err := cur.Decode(&p); err != nil {
			return err
		}
		entries = append(entries, bson.M{
			"_id":        bson.NewObjectID(),
			"product_id": p.ID,
			"delta":      qty,
			"reason":     "opening_balance",
			"created_at": now,
		})
	}
	if err := cur.Err(); err != nil {
		return err
	}
	if len(entries) == 0 {
		return nil
	}
	if _, err := ledger.InsertMany(ctx, entries); err != nil {
		return err
	}
	logger.Info("Seeded inventory opening balance", zap.Int("entries", len(entries)), zap.Int("qty", qty))
	return nil
}

func seedJetonsAndAssignProducts(ctx context.Context, db *mongo.Database, logger *zap.Logger) error {
	jetons := db.Collection(database.JetonsCollection)
	categories := db.Collection(database.CategoriesCollection)
	products := db.Collection(database.ProductsCollection)

	now := time.Now().UTC()

	jetonSeeds := []struct {
		Name    string
		Palette string
	}{
		{"Burger", "blue"},
		{"Getränk", "red"},
		{"Pommes", "yellow"},
		{"Menu", "green"},
	}

	var names []string
	var jetonOps []mongo.WriteModel
	for _, seed := range jetonSeeds {
		names = append(names, seed.Name)
		update := bson.D{
			{Key: "$setOnInsert", Value: bson.D{
				{Key: "_id", Value: bson.NewObjectID()},
				{Key: "created_at", Value: now},
			}},
			{Key: "$set", Value: bson.D{
				{Key: "name", Value: seed.Name},
				{Key: "palette_color", Value: seed.Palette},
				{Key: "updated_at", Value: now},
			}},
		}
		jetonOps = append(jetonOps, mongo.NewUpdateOneModel().
			SetFilter(bson.D{{Key: "name", Value: seed.Name}}).
			SetUpdate(update).
			SetUpsert(true))
	}

	if len(jetonOps) > 0 {
		if _, err := jetons.BulkWrite(ctx, jetonOps); err != nil {
			return fmt.Errorf("failed to upsert jetons: %w", err)
		}
	}

	// Load jeton IDs by name for assignment
	cur, err := jetons.Find(ctx, bson.M{"name": bson.M{"$in": names}})
	if err != nil {
		return fmt.Errorf("failed to load jetons: %w", err)
	}
	defer func() { _ = cur.Close(ctx) }()

	jetonByName := make(map[string]bson.ObjectID)
	for cur.Next(ctx) {
		var j struct {
			ID   bson.ObjectID `bson:"_id"`
			Name string        `bson:"name"`
		}
		if err := cur.Decode(&j); err != nil {
			return fmt.Errorf("failed to decode jeton: %w", err)
		}
		jetonByName[j.Name] = j.ID
	}
	if err := cur.Err(); err != nil {
		return fmt.Errorf("cursor error while loading jetons: %w", err)
	}

	// Map category IDs to names for deterministic jeton selection
	catCur, err := categories.Find(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("failed to fetch categories: %w", err)
	}
	defer func() { _ = catCur.Close(ctx) }()

	categoryNames := make(map[bson.ObjectID]string)
	for catCur.Next(ctx) {
		var c struct {
			ID   bson.ObjectID `bson:"_id"`
			Name string        `bson:"name"`
		}
		if err := catCur.Decode(&c); err != nil {
			return fmt.Errorf("failed to decode category: %w", err)
		}
		categoryNames[c.ID] = c.Name
	}
	if err := catCur.Err(); err != nil {
		return fmt.Errorf("cursor error while loading categories: %w", err)
	}

	// Assign jetons to products based on category/type
	prodCur, err := products.Find(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("failed to fetch products: %w", err)
	}
	defer func() { _ = prodCur.Close(ctx) }()

	var productOps []mongo.WriteModel
	for prodCur.Next(ctx) {
		var p struct {
			ID         bson.ObjectID `bson:"_id"`
			Name       string        `bson:"name"`
			CategoryID bson.ObjectID `bson:"category_id"`
			Type       string        `bson:"type"`
		}
		if err := prodCur.Decode(&p); err != nil {
			return fmt.Errorf("failed to decode product: %w", err)
		}

		jetonName := ""
		if catName, ok := categoryNames[p.CategoryID]; ok {
			switch catName {
			case "Burgers":
				jetonName = "Burger"
			case "Beilagen":
				jetonName = "Pommes"
			case "Getränke":
				jetonName = "Getränk"
			case "Menus":
				jetonName = "Menu"
			}
		}

		if p.Type == "menu" && jetonName == "" {
			jetonName = "Menu"
		}

		jetonID, ok := jetonByName[jetonName]
		if !ok {
			// Fallback to any available jeton to ensure assignment
			for _, id := range jetonByName {
				jetonID = id
				ok = true
				break
			}
		}
		if !ok {
			continue
		}

		productOps = append(productOps, mongo.NewUpdateOneModel().
			SetFilter(bson.D{{Key: "_id", Value: p.ID}}).
			SetUpdate(bson.D{{Key: "$set", Value: bson.D{
				{Key: "jeton_id", Value: jetonID},
				{Key: "updated_at", Value: now},
			}}}))
	}
	if err := prodCur.Err(); err != nil {
		return fmt.Errorf("cursor error while loading products: %w", err)
	}

	if len(productOps) > 0 {
		if _, err := products.BulkWrite(ctx, productOps); err != nil {
			return fmt.Errorf("failed to assign jetons to products: %w", err)
		}
	}

	logger.Info("Seeded jetons and assigned to products",
		zap.Int("jetons", len(jetonByName)),
		zap.Int("productsUpdated", len(productOps)),
	)

	return nil
}

func seedPosSettings(ctx context.Context, db *mongo.Database, logger *zap.Logger) error {
	coll := db.Collection(database.PosSettingsCollection)
	now := time.Now().UTC()
	_, err := coll.UpdateByID(ctx, "default", bson.M{
		"$setOnInsert": bson.M{"_id": "default"},
		"$set":         bson.M{"mode": domain.PosModeQRCode, "updated_at": now},
	}, options.UpdateOne().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("failed to upsert POS settings: %w", err)
	}
	logger.Info("Seeded POS settings", zap.String("mode", string(domain.PosModeQRCode)))
	return nil
}

func generateAdmins(ctx context.Context, db *mongo.Database, logger *zap.Logger) error {
	collection := db.Collection("users")
	adminCount := 2

	var operations []mongo.WriteModel

	for i := 0; i < adminCount; i++ {
		now := time.Now()
		admin := map[string]interface{}{
			"_id":         bson.NewObjectID(),
			"email":       fmt.Sprintf("admin%d@dev.local", i+1),
			"first_name":  gofakeit.FirstName(),
			"last_name":   gofakeit.LastName(),
			"role":        "admin",
			"is_verified": true,
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

func generateCategories(ctx context.Context, db *mongo.Database, logger *zap.Logger) error {
	collection := db.Collection("categories")

	// Seed with explicit positions
	type seedCat struct {
		Name     string
		Position int
	}
	categories := []seedCat{{"Menus", 0}, {"Burgers", 1}, {"Beilagen", 2}, {"Getränke", 3}}
	var operations []mongo.WriteModel

	for _, cat := range categories {
		now := time.Now()
		category := map[string]interface{}{
			"_id":        bson.NewObjectID(),
			"name":       cat.Name,
			"is_active":  true,
			"position":   cat.Position,
			"created_at": now.Add(-time.Duration(gofakeit.Number(0, 180*24)) * time.Hour),
			"updated_at": now,
		}

		// Use upsert for idempotency - match by name and ensure is_active is true
		filter := bson.D{{Key: "name", Value: cat.Name}}
		update := bson.D{
			{Key: "$setOnInsert", Value: bson.D{
				{Key: "_id", Value: category["_id"]},
				{Key: "created_at", Value: category["created_at"]},
			}},
			{Key: "$set", Value: bson.D{
				{Key: "name", Value: cat.Name},
				{Key: "is_active", Value: true},
				{Key: "position", Value: cat.Position},
				{Key: "updated_at", Value: now},
			}},
		}
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
		ID   bson.ObjectID `bson:"_id"`
		Name string        `bson:"name"`
	}
	if err := cursor.All(ctx, &categories); err != nil {
		return fmt.Errorf("failed to decode categories: %w", err)
	}

	if len(categories) == 0 {
		return fmt.Errorf("no categories found - run generateCategories first")
	}

	productCount := 8
	var operations []mongo.WriteModel

	// Product data with actual available image URLs
	productData := []struct {
		name       string
		image      string
		priceCents int64
	}{
		{"Smash Burger", "/assets/images/products/bless2n-takeaway-smash-burger-16x9.png", 850},
		{"Veggie Burger", "/assets/images/products/bless2n-takeaway-veggie-burger-16x9.png", 850},
		{"Pommes", "/assets/images/products/bless2n-takeaway-pommes-16x9.png", 400},
		{"Coca Cola", "/assets/images/products/bless2n-takeaway-coca-cola-16x9.png", 250},
		{"Ice Tea Lemon", "/assets/images/products/bless2n-takeaway-ice-tea-lemon-16x9.png", 250},
		{"Red Bull", "/assets/images/products/bless2n-takeaway-red-bull-16x9.png", 250},
		{"El Tony Mate", "/assets/images/products/bless2n-takeaway-el-tony-mate-16x9.png", 250},
		{"Wasser Prickelnd", "/assets/images/products/bless2n-takeaway-wasser-prickelnd-16x9.png", 250},
	}

	for i := 0; i < productCount; i++ {
		now := time.Now()
		data := productData[i%len(productData)]

		// Assign category based on product type
		var categoryID bson.ObjectID
		for _, category := range categories {
			switch data.name {
			case "Smash Burger", "Veggie Burger":
				if category.Name == "Burgers" {
					categoryID = category.ID
				}
			case "Pommes":
				if category.Name == "Beilagen" {
					categoryID = category.ID
				}
			case "Coca Cola", "Ice Tea Lemon", "Red Bull", "El Tony Mate", "Wasser Prickelnd":
				if category.Name == "Getränke" {
					categoryID = category.ID
				}
			}
		}

		product := map[string]interface{}{
			"_id":         bson.NewObjectID(),
			"category_id": categoryID,
			"type":        "simple",
			"name":        data.name,
			"image":       data.image,
			"price_cents": data.priceCents,
			"is_active":   true,
			"created_at":  now.Add(-time.Duration(gofakeit.Number(0, 180*24)) * time.Hour),
			"updated_at":  now,
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

	// Get Menus category for bundles
	var menusCategory struct {
		ID bson.ObjectID `bson:"_id"`
	}
	err := categoriesCollection.FindOne(ctx, bson.D{{Key: "name", Value: "Menus"}}).Decode(&menusCategory)
	if err != nil {
		return fmt.Errorf("failed to find Menus category: %w", err)
	}

	menuCount := 2
	var operations []mongo.WriteModel

	// Menu data with actual available image URLs
	menuData := []struct {
		name       string
		image      string
		priceCents int64
	}{
		{"Menu Gross", "/assets/images/products/bless2n-takeaway-menu-2-gross-16x9.png", 1400},
		{"Menu Klein", "/assets/images/products/bless2n-takeaway-menu-1-klein-16x9.png", 1000},
	}

	for i := 0; i < menuCount; i++ {
		now := time.Now()
		data := menuData[i]

		menuProduct := map[string]interface{}{
			"_id":         bson.NewObjectID(),
			"category_id": menusCategory.ID,
			"type":        "menu",
			"name":        data.name,
			"image":       data.image,
			"price_cents": data.priceCents,
			"is_active":   true,
			"created_at":  now.Add(-time.Duration(gofakeit.Number(0, 90*24)) * time.Hour),
			"updated_at":  now,
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
		ID   bson.ObjectID `bson:"_id"`
		Name string        `bson:"name"`
	}
	if err := menuCursor.All(ctx, &menuProducts); err != nil {
		return fmt.Errorf("failed to decode menu products: %w", err)
	}

	var operations []mongo.WriteModel

	// Create slots for each menu with different configurations
	for _, menu := range menuProducts {
		var slotNames []string

		// Menu Gross has Burger, Beilage, Getränk
		// Menu Klein has only Burger, Getränk
		switch menu.Name {
		case "Menu Gross":
			slotNames = []string{"Burger", "Beilage", "Getränk"}
		case "Menu Klein":
			slotNames = []string{"Burger", "Getränk"}
		}

		for i, slotName := range slotNames {
			menuSlot := map[string]interface{}{
				"_id":        bson.NewObjectID(),
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
		ID   bson.ObjectID `bson:"_id"`
		Name string        `bson:"name"`
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
		ID   bson.ObjectID `bson:"_id"`
		Name string        `bson:"name"`
	}
	if err := simpleCursor.All(ctx, &simpleProducts); err != nil {
		return fmt.Errorf("failed to decode simple products: %w", err)
	}

	if len(simpleProducts) == 0 {
		return fmt.Errorf("no simple products found to create menu slot items")
	}

	// Categorize products by type for easier assignment
	var burgerProducts, friesProducts, drinkProducts []struct {
		ID   bson.ObjectID `bson:"_id"`
		Name string        `bson:"name"`
	}

	for _, product := range simpleProducts {
		switch product.Name {
		case "Smash Burger", "Veggie Burger":
			burgerProducts = append(burgerProducts, product)
		case "Pommes":
			friesProducts = append(friesProducts, product)
		case "Coca Cola", "Ice Tea Lemon", "Red Bull", "El Tony Mate", "Wasser Prickelnd":
			drinkProducts = append(drinkProducts, product)
		}
	}

	var operations []mongo.WriteModel

	// Add appropriate products to each slot based on slot name
	for _, slot := range menuSlots {
		var relevantProducts []struct {
			ID   bson.ObjectID `bson:"_id"`
			Name string        `bson:"name"`
		}

		switch slot.Name {
		case "Burger":
			relevantProducts = burgerProducts
		case "Beilage":
			relevantProducts = friesProducts
		case "Getränk":
			relevantProducts = drinkProducts
		}

		// Add all relevant products for this slot
		for _, product := range relevantProducts {
			slotItem := map[string]interface{}{
				"_id":          bson.NewObjectID(),
				"menu_slot_id": slot.ID,
				"product_id":   product.ID,
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

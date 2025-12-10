package seed

import (
	"backend/internal/database"
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.uber.org/zap"
)

type IndexSeeder struct {
	Logger *zap.Logger
}

func NewIndexSeeder(logger *zap.Logger) IndexSeeder {
	return IndexSeeder{Logger: logger}
}

func (s IndexSeeder) Name() string {
	return "indexes"
}

func (s IndexSeeder) Seed(ctx context.Context, db *mongo.Database) error {
	logger := loggerOrNop(s.Logger)
	logger.Info("Ensuring MongoDB collections and indexes")

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
			return err
		}
	}

	indexers := []func(context.Context, *mongo.Database) error{
		ensureUserIndexes,
		ensureIdentityLinkIndexes,
		ensureOrderIndexes,
		ensureCategoryIndexes,
		ensureProductIndexes,
		ensureOrderItemIndexes,
		ensureMenuSlotIndexes,
		ensureMenuSlotItemIndexes,
		ensureInventoryIndexes,
		ensureAuthIndexes,
		ensureAdminInviteIndexes,
		ensureStationIndexes,
		ensureStationProductIndexes,
		ensureAuditIndexes,
		ensurePosDeviceIndexes,
		ensureJetonIndexes,
	}

	for _, indexer := range indexers {
		if err := indexer(ctx, db); err != nil {
			return err
		}
	}

	logger.Info("Collections and indexes ensured")
	return nil
}

func ensureCollectionExists(ctx context.Context, db *mongo.Database, name string, logger *zap.Logger) error {
	if err := db.CreateCollection(ctx, name); err != nil {
		if ce, ok := err.(mongo.CommandError); ok && ce.Code == 48 { // NamespaceExists
			return nil
		}
		msg := err.Error()
		if strings.Contains(msg, "NamespaceExists") || strings.Contains(msg, "already exists") {
			return nil
		}
		return fmt.Errorf("create collection %s: %w", name, err)
	}
	logger.Info("Created collection", zap.String("name", name))
	return nil
}

func ensureUserIndexes(ctx context.Context, db *mongo.Database) error {
	coll := db.Collection(database.UsersCollection)
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "email", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "role", Value: 1}}},
		{Keys: bson.D{{Key: "is_verified", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: 1}}},
	}
	_, err := coll.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("users indexes: %w", err)
	}
	return nil
}

func ensureIdentityLinkIndexes(ctx context.Context, db *mongo.Database) error {
	coll := db.Collection(database.IdentityLinksCollection)
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "provider", Value: 1}, {Key: "provider_user_id", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "user_id", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: 1}}},
	}
	_, err := coll.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("identity_links indexes: %w", err)
	}
	return nil
}

func ensureOrderIndexes(ctx context.Context, db *mongo.Database) error {
	coll := db.Collection(database.OrdersCollection)
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "customer_id", Value: 1}}},
		{Keys: bson.D{{Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: 1}}},
	}
	_, err := coll.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("orders indexes: %w", err)
	}
	return nil
}

func ensureCategoryIndexes(ctx context.Context, db *mongo.Database) error {
	coll := db.Collection(database.CategoriesCollection)
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "name", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "is_active", Value: 1}}},
		{Keys: bson.D{{Key: "position", Value: 1}}},
		{Keys: bson.D{{Key: "position", Value: 1}, {Key: "name", Value: 1}}},
	}
	_, err := coll.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("categories indexes: %w", err)
	}
	return nil
}

func ensureProductIndexes(ctx context.Context, db *mongo.Database) error {
	coll := db.Collection(database.ProductsCollection)
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "name", Value: 1}}},
		{Keys: bson.D{{Key: "category_id", Value: 1}}},
		{Keys: bson.D{{Key: "is_active", Value: 1}}},
		{Keys: bson.D{{Key: "type", Value: 1}}},
	}
	_, err := coll.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("products indexes: %w", err)
	}
	return nil
}

func ensureOrderItemIndexes(ctx context.Context, db *mongo.Database) error {
	coll := db.Collection(database.OrderItemsCollection)
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "order_id", Value: 1}}},
		{Keys: bson.D{{Key: "product_id", Value: 1}}},
		{Keys: bson.D{{Key: "parent_item_id", Value: 1}}},
		{Keys: bson.D{{Key: "is_redeemed", Value: 1}}},
	}
	_, err := coll.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("order_items indexes: %w", err)
	}
	return nil
}

func ensureMenuSlotIndexes(ctx context.Context, db *mongo.Database) error {
	coll := db.Collection(database.MenuSlotsCollection)
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "product_id", Value: 1}}},
		{Keys: bson.D{{Key: "sequence", Value: 1}}},
	}
	_, err := coll.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("menu_slots indexes: %w", err)
	}
	return nil
}

func ensureMenuSlotItemIndexes(ctx context.Context, db *mongo.Database) error {
	coll := db.Collection(database.MenuSlotItemsCollection)
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "menu_slot_id", Value: 1}}},
		{Keys: bson.D{{Key: "product_id", Value: 1}}},
	}
	_, err := coll.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("menu_slot_items indexes: %w", err)
	}
	return nil
}

func ensureInventoryIndexes(ctx context.Context, db *mongo.Database) error {
	coll := db.Collection(database.InventoryLedgerCollection)
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "product_id", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: 1}}},
		{Keys: bson.D{{Key: "product_id", Value: 1}, {Key: "created_at", Value: -1}}},
	}
	_, err := coll.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("inventory_ledger indexes: %w", err)
	}
	return nil
}

func ensureAuthIndexes(ctx context.Context, db *mongo.Database) error {
	otp := db.Collection(database.OTPTokensCollection)
	otpIdx := []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}}},
		{Keys: bson.D{{Key: "expires_at", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(0)},
		{Keys: bson.D{{Key: "created_at", Value: 1}}},
	}
	if _, err := otp.Indexes().CreateMany(ctx, otpIdx); err != nil {
		return fmt.Errorf("otp_tokens indexes: %w", err)
	}

	emailChange := db.Collection(database.EmailChangeTokensCollection)
	ecIdx := []mongo.IndexModel{
		{Keys: bson.D{{Key: "user_id", Value: 1}}},
		{Keys: bson.D{{Key: "new_email", Value: 1}}},
		{Keys: bson.D{{Key: "expires_at", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(0)},
		{Keys: bson.D{{Key: "created_at", Value: 1}}},
	}
	if _, err := emailChange.Indexes().CreateMany(ctx, ecIdx); err != nil {
		return fmt.Errorf("email_change_tokens indexes: %w", err)
	}

	refresh := db.Collection(database.RefreshTokensCollection)
	rtIdx := []mongo.IndexModel{
		{Keys: bson.D{{Key: "token_hash", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "is_revoked", Value: 1}, {Key: "expires_at", Value: 1}}},
		{Keys: bson.D{{Key: "family_id", Value: 1}, {Key: "is_revoked", Value: 1}}},
		{Keys: bson.D{{Key: "last_used_at", Value: 1}}},
	}
	if _, err := refresh.Indexes().CreateMany(ctx, rtIdx); err != nil {
		return fmt.Errorf("refresh_tokens indexes: %w", err)
	}

	return nil
}

func ensureAdminInviteIndexes(ctx context.Context, db *mongo.Database) error {
	coll := db.Collection(database.AdminInvitesCollection)
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "token_hash", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "invitee_email", Value: 1}}},
		{Keys: bson.D{{Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "expires_at", Value: 1}}, Options: options.Index().SetExpireAfterSeconds(0)},
		{Keys: bson.D{{Key: "created_at", Value: 1}}},
	}
	_, err := coll.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("admin_invites indexes: %w", err)
	}
	return nil
}

func ensureStationIndexes(ctx context.Context, db *mongo.Database) error {
	coll := db.Collection(database.StationsCollection)
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "device_key", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "approved", Value: 1}}},
		{Keys: bson.D{{Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: 1}}},
	}
	_, err := coll.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("stations indexes: %w", err)
	}
	return nil
}

func ensureStationProductIndexes(ctx context.Context, db *mongo.Database) error {
	coll := db.Collection(database.StationProductsCollection)
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "station_id", Value: 1}, {Key: "product_id", Value: 1}}, Options: options.Index().SetUnique(true)},
	}
	_, err := coll.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("station_products indexes: %w", err)
	}
	return nil
}

func ensureAuditIndexes(ctx context.Context, db *mongo.Database) error {
	coll := db.Collection(database.AuditLogsCollection)
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "entity_type", Value: 1}, {Key: "entity_id", Value: 1}}},
		{Keys: bson.D{{Key: "actor_user_id", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: 1}}},
	}
	_, err := coll.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("audit_logs indexes: %w", err)
	}
	return nil
}

func ensurePosDeviceIndexes(ctx context.Context, db *mongo.Database) error {
	coll := db.Collection(database.PosDevicesCollection)
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "device_token", Value: 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{Key: "approved", Value: 1}}},
		{Keys: bson.D{{Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: 1}}},
	}
	_, err := coll.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("pos_devices indexes: %w", err)
	}
	return nil
}

func ensureJetonIndexes(ctx context.Context, db *mongo.Database) error {
	coll := db.Collection(database.JetonsCollection)
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "name", Value: 1}}, Options: options.Index().SetUnique(true)},
	}
	_, err := coll.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		return fmt.Errorf("jetons indexes: %w", err)
	}
	return nil
}

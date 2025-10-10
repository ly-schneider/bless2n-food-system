package database

import (
	"context"
	"time"

	"backend/internal/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func NewMongoDB(cfg config.Config) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.Mongo.URI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		zap.L().Error("failed to connect to MongoDB", zap.Error(err))
		return nil, err
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		zap.L().Error("failed to ping MongoDB", zap.Error(err))
		return nil, err
	}

	// Use the configured database name
	dbName := cfg.Mongo.Database
	database := client.Database(dbName)
	zap.L().Info("successfully connected to MongoDB", zap.String("database", dbName))

	return &MongoDB{
		Client:   client,
		Database: database,
	}, nil
}

func (m *MongoDB) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return m.Client.Disconnect(ctx)
}

// Collection names constants
const (
	UsersCollection           = "users"
	AdminInvitesCollection    = "admin_invites"
	OTPTokensCollection       = "otp_tokens"
	EmailChangeTokensCollection = "email_change_tokens"
	RefreshTokensCollection   = "refresh_tokens"
	IdentityLinksCollection   = "identity_links"
	StationsCollection        = "stations"
	StationRequestsCollection = "station_requests"
	CategoriesCollection      = "categories"
	ProductsCollection        = "products"
	MenuSlotsCollection       = "menu_slots"
	MenuSlotItemsCollection   = "menu_slot_items"
	StationProductsCollection = "station_products"
	InventoryLedgerCollection = "inventory_ledger"
    OrdersCollection          = "orders"
    OrderItemsCollection      = "order_items"
    AuditLogsCollection       = "audit_logs"
    PosDevicesCollection      = "pos_devices"
    PosRequestsCollection     = "pos_requests"
)

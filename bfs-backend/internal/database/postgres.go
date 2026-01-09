package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"backend/internal/config"
	"backend/migrations"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

// PostgresDB wraps a pgx connection pool
type PostgresDB struct {
	Pool *pgxpool.Pool
	DB   *sql.DB // Standard sql.DB for goose migrations
}

// NewPostgresDB creates a new PostgreSQL connection pool
func NewPostgresDB(cfg config.Config) (*PostgresDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create pgx pool
	poolConfig, err := pgxpool.ParseConfig(cfg.Postgres.DSN)
	if err != nil {
		zap.L().Error("failed to parse PostgreSQL DSN", zap.Error(err))
		return nil, fmt.Errorf("failed to parse PostgreSQL DSN: %w", err)
	}

	// Configure pool settings
	poolConfig.MaxConns = int32(cfg.Postgres.MaxConns)
	poolConfig.MinConns = int32(cfg.Postgres.MinConns)
	poolConfig.MaxConnLifetime = cfg.Postgres.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.Postgres.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		zap.L().Error("failed to connect to PostgreSQL", zap.Error(err))
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Ping to verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		zap.L().Error("failed to ping PostgreSQL", zap.Error(err))
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	// Create standard sql.DB from pgx pool for goose migrations
	sqlDB := stdlib.OpenDBFromPool(pool)

	zap.L().Info("successfully connected to PostgreSQL",
		zap.String("host", poolConfig.ConnConfig.Host),
		zap.String("database", poolConfig.ConnConfig.Database),
	)

	return &PostgresDB{
		Pool: pool,
		DB:   sqlDB,
	}, nil
}

// Close closes the PostgreSQL connection pool
func (p *PostgresDB) Close() error {
	if p.DB != nil {
		if err := p.DB.Close(); err != nil {
			zap.L().Error("failed to close sql.DB", zap.Error(err))
		}
	}
	if p.Pool != nil {
		p.Pool.Close()
	}
	return nil
}

// RunMigrations runs all pending goose migrations
func (p *PostgresDB) RunMigrations() error {
	goose.SetBaseFS(migrations.EmbeddedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(p.DB, "."); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	zap.L().Info("database migrations completed successfully")
	return nil
}

// MigrateDown rolls back the last migration
func (p *PostgresDB) MigrateDown() error {
	goose.SetBaseFS(migrations.EmbeddedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Down(p.DB, "."); err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	zap.L().Info("database migration rollback completed successfully")
	return nil
}

// MigrationStatus returns the current migration status
func (p *PostgresDB) MigrationStatus() error {
	goose.SetBaseFS(migrations.EmbeddedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Status(p.DB, "."); err != nil {
		return fmt.Errorf("failed to get migration status: %w", err)
	}

	return nil
}

// PostgreSQL table names (matching the migration file table names)
const (
	TableUsers              = "users"
	TableIdentityLinks      = "identity_links"
	TableRefreshTokens      = "refresh_tokens"
	TableOTPTokens          = "otp_tokens"
	TableEmailChangeTokens  = "email_change_tokens"
	TableCategories         = "categories"
	TableJetons             = "jetons"
	TableProducts           = "products"
	TableMenuSlots          = "menu_slots"
	TableMenuSlotItems      = "menu_slot_items"
	TableOrders             = "orders"
	TableOrderItems         = "order_items"
	TableStations           = "stations"
	TableStationProducts    = "station_products"
	TableInventoryLedger    = "inventory_ledger"
	TableIdempotencyRecords = "idempotency_records"
	TableAuditLogs          = "audit_logs"
	TableAdminInvites       = "admin_invites"
	TablePosDevices         = "pos_devices"
	TablePosSettings        = "pos_settings"
)

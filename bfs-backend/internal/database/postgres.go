package database

import (
	"context"
	"fmt"
	"time"

	"backend/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PostgresDB wraps a pgx connection pool
type PostgresDB struct {
	Pool *pgxpool.Pool
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

	zap.L().Info("successfully connected to PostgreSQL",
		zap.String("host", poolConfig.ConnConfig.Host),
		zap.String("database", poolConfig.ConnConfig.Database),
	)

	return &PostgresDB{
		Pool: pool,
	}, nil
}

// Close closes the PostgreSQL connection pool
func (p *PostgresDB) Close() error {
	if p.Pool != nil {
		p.Pool.Close()
	}
	return nil
}

// GormDB wraps a GORM database connection.
type GormDB struct {
	DB *gorm.DB
}

// NewGormDB creates a new GORM database connection using the existing DSN.
func NewGormDB(cfg config.Config) (*GormDB, error) {
	logLevel := logger.Silent
	if cfg.Logger.Development {
		logLevel = logger.Info
	}

	gormConfig := &gorm.Config{
		Logger:                 logger.Default.LogMode(logLevel),
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	}

	db, err := gorm.Open(postgres.Open(cfg.Postgres.DSN), gormConfig)
	if err != nil {
		zap.L().Error("failed to connect to PostgreSQL with GORM", zap.Error(err))
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Get underlying sql.DB to configure pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.Postgres.MaxConns)
	sqlDB.SetMaxIdleConns(cfg.Postgres.MinConns)
	sqlDB.SetConnMaxLifetime(cfg.Postgres.MaxConnLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.Postgres.MaxConnIdleTime)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	zap.L().Info("successfully connected to PostgreSQL with GORM")

	return &GormDB{DB: db}, nil
}

// Close closes the GORM database connection.
func (g *GormDB) Close() error {
	sqlDB, err := g.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

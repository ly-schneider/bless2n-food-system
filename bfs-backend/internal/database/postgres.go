package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"backend/internal/config"
	"backend/internal/generated/ent"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

type PostgresDB struct {
	Pool *pgxpool.Pool
}

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

func (p *PostgresDB) Close() error {
	if p.Pool != nil {
		p.Pool.Close()
	}
	return nil
}

func NewEntClient(cfg config.Config) (*ent.Client, *sql.DB, error) {
	db, err := sql.Open("pgx", cfg.Postgres.DSN)
	if err != nil {
		zap.L().Error("failed to open database for Ent", zap.Error(err))
		return nil, nil, fmt.Errorf("failed to open database for Ent: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.Postgres.MaxConns)
	db.SetMaxIdleConns(cfg.Postgres.MinConns)
	db.SetConnMaxLifetime(cfg.Postgres.MaxConnLifetime)
	db.SetConnMaxIdleTime(cfg.Postgres.MaxConnIdleTime)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, nil, fmt.Errorf("failed to ping PostgreSQL for Ent: %w", err)
	}

	drv := entsql.OpenDB(dialect.Postgres, db)
	client := ent.NewClient(ent.Driver(drv))

	zap.L().Info("successfully connected to PostgreSQL with Ent")

	return client, db, nil
}

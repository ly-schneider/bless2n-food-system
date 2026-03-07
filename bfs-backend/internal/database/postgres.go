package database

import (
	"database/sql"
	"fmt"

	"backend/internal/config"
	"backend/internal/generated/ent"
	"backend/internal/trace"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

func NewEntClient(cfg config.Config) (*ent.Client, *sql.DB, error) {
	db, err := sql.Open("pgx", cfg.Postgres.DSN)
	if err != nil {
		zap.L().Error("failed to open database for Ent", zap.Error(err))
		return nil, nil, fmt.Errorf("failed to open database for Ent: %w", err)
	}

	db.SetMaxOpenConns(cfg.Postgres.MaxConns)
	db.SetMaxIdleConns(cfg.Postgres.MinConns)
	db.SetConnMaxLifetime(cfg.Postgres.MaxConnLifetime)
	db.SetConnMaxIdleTime(cfg.Postgres.MaxConnIdleTime)

	drv := entsql.OpenDB(dialect.Postgres, db)
	client := ent.NewClient(ent.Driver(drv))
	client.Intercept(trace.QueryInterceptor())
	client.Use(trace.MutationHook())

	zap.L().Info("Ent client initialized (connection will be established on first query)")

	return client, db, nil
}

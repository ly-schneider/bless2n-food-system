package db

import (
	"context"
	"fmt"
	"time"

	"backend/internal/config"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type contextKey string

const userKey contextKey = "userID"

func InjectUser(ctx context.Context, uid string) context.Context {
	return context.WithValue(ctx, userKey, uid)
}

func New(cfg config.DBConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
		cfg.User, cfg.Pass, cfg.Host, cfg.Port, cfg.Name)

	pc, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("pgx parse DSN: %w", err)
	}

	pc.BeforeAcquire = func(ctx context.Context, conn *pgx.Conn) bool {
		uid, _ := ctx.Value(userKey).(string)

		if uid == "" {
			_, _ = conn.Exec(ctx, "RESET app.current_user_id")
			return true
		}

		if _, err := conn.Exec(ctx, "SET app.current_user_id = $1", uid); err != nil {
			return false
		}
		return true
	}

	pc.AfterRelease = func(conn *pgx.Conn) bool {
		_, _ = conn.Exec(context.Background(), "RESET app.current_user_id")
		return true
	}

	pc.MaxConns = 25
	pc.MinConns = 5
	pc.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), pc)
	if err != nil {
		return nil, fmt.Errorf("pgx new pool: %w", err)
	}

	std := stdlib.OpenDBFromPool(pool)
	std.SetMaxIdleConns(int(pc.MinConns))
	std.SetMaxOpenConns(int(pc.MaxConns))
	std.SetConnMaxIdleTime(pc.MaxConnIdleTime)

	return gorm.Open(postgres.New(postgres.Config{Conn: std}), &gorm.Config{})
}

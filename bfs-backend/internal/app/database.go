package app

import (
	"database/sql"

	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/generated/ent"
)

func NewEntClient(cfg config.Config) (*ent.Client, *sql.DB, error) {
	return database.NewEntClient(cfg)
}

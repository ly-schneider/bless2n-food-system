package app

import (
	"backend/internal/config"
	"backend/internal/db"

	"gorm.io/gorm"
)

func NewDB(cfg config.Config) (*gorm.DB, error) {
	return db.New(cfg.DB)
}

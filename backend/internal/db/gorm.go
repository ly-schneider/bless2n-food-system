package db

import (
	"backend/internal/config"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func New(cfg config.DBConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.Host, cfg.User, cfg.Pass, cfg.Name, cfg.Port,
	)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

package db

import (
	"backend/internal/config"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func New(cfg config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		cfg.DB.Host, cfg.DB.User, cfg.DB.Pass, cfg.DB.Name, cfg.DB.Port,
	)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

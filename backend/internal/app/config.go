package app

import (
	"backend/internal/config"
)

func NewConfig() config.Config {
	return config.Load()
}

func ProvideAppConfig(cfg config.Config) config.AppConfig {
	return cfg.App
}
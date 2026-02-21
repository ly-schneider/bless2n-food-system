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

func ProvideAndroidConfig(cfg config.Config) config.AndroidConfig {
	return cfg.Android
}

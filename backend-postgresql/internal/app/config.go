package app

import "backend/internal/config"

func NewConfig() config.Config {
	return config.Load()
}

// ProvideAppConfig extracts the AppConfig from the Config structure
func ProvideAppConfig(cfg config.Config) config.AppConfig {
	return cfg.App
}

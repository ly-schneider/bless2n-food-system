package app

import (
	"backend/internal/config"
)

func NewConfig() (config.Config, error) {
	return config.Load(), nil
}

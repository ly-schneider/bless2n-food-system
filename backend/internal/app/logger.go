package app

import (
	"backend/internal/config"

	"go.uber.org/zap"
)

func NewLogger(cfg config.Config) (*zap.Logger, error) {
	var l *zap.Logger
	var err error

	if cfg.Logger.Development {
		l, err = zap.NewDevelopment()
	} else {
		l, err = zap.NewProduction()
	}
	if err != nil {
		return nil, err
	}

	zap.ReplaceGlobals(l)
	return l, nil
}

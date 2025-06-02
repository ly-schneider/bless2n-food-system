package logger

import (
	"backend/internal/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Global logger instance
	L *zap.SugaredLogger
)

// Init initializes the global logger
func Init(cfg config.LoggerConfig) error {
	var config zap.Config

	if cfg.Development {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.Encoding = "console"
	} else {
		config = zap.NewProductionConfig()
		config.Encoding = "json"
	}

	// Parse log level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return err
	}
	config.Level = zap.NewAtomicLevelAt(level)

	logger, err := config.Build()
	if err != nil {
		return err
	}

	L = logger.Sugar()
	return nil
}

// InitDefault initializes the global logger with default settings
func InitDefault() error {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.Encoding = "console"
	config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)

	logger, err := config.Build()
	if err != nil {
		return err
	}

	L = logger.Sugar()
	return nil
}

// Sync flushes any buffered log entries
func Sync() {
	if L != nil {
		L.Sync()
	}
}

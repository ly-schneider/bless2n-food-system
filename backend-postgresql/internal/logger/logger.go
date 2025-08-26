package logger

import (
	"backend/internal/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	L      *zap.SugaredLogger
	logger *zap.Logger
)

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

	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return err
	}
	config.Level = zap.NewAtomicLevelAt(level)

	logger, err = config.Build()
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

// GetLogger returns the underlying zap.Logger
func GetLogger() *zap.Logger {
	return logger
}

// Sync flushes any buffered log entries
func Sync() error {
	if L != nil {
		return L.Sync()
	}
	return nil
}

// Safe logging functions that handle nil logger
func Info(msg string, keysAndValues ...interface{}) {
	if L != nil {
		L.Infow(msg, keysAndValues...)
	}
}

func Warn(msg string, keysAndValues ...interface{}) {
	if L != nil {
		L.Warnw(msg, keysAndValues...)
	}
}

func Error(msg string, keysAndValues ...interface{}) {
	if L != nil {
		L.Errorw(msg, keysAndValues...)
	}
}

func Debug(msg string, keysAndValues ...interface{}) {
	if L != nil {
		L.Debugw(msg, keysAndValues...)
	}
}

func Fatal(msg string, keysAndValues ...interface{}) {
	if L != nil {
		L.Fatalw(msg, keysAndValues...)
	}
}

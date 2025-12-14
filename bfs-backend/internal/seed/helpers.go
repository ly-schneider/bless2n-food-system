package seed

import (
	"time"

	"go.uber.org/zap"
)

// seededAt provides a deterministic creation timestamp for inserted documents.
var seededAt = time.Date(2024, time.January, 1, 12, 0, 0, 0, time.UTC)

func loggerOrNop(l *zap.Logger) *zap.Logger {
	if l != nil {
		return l
	}
	return zap.NewNop()
}

func ptr[T any](v T) *T {
	return &v
}

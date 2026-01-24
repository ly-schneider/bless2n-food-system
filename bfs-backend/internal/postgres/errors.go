package postgres

import (
	"errors"

	"gorm.io/gorm"
)

var (
	// ErrNotFound is returned when a record is not found.
	ErrNotFound = errors.New("record not found")
	// ErrConflict is returned when a unique constraint is violated.
	ErrConflict = errors.New("conflict: record already exists")
)

// translateError converts GORM errors to repository errors.
func translateError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	// Check for PostgreSQL unique violation (code 23505)
	if isUniqueViolation(err) {
		return ErrConflict
	}
	return err
}

// isUniqueViolation checks if the error is a PostgreSQL unique constraint violation.
func isUniqueViolation(err error) bool {
	// PostgreSQL error code for unique violation is 23505
	// We check by string matching since the error type varies
	return err != nil && (errors.Is(err, gorm.ErrDuplicatedKey) ||
		containsString(err.Error(), "23505") ||
		containsString(err.Error(), "duplicate key"))
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[:len(substr)] == substr || containsString(s[1:], substr)))
}

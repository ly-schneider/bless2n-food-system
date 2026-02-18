package repository

import (
	"errors"
	"strings"

	"backend/internal/generated/ent"
)

var (
	ErrNotFound = errors.New("record not found")
	ErrConflict = errors.New("conflict: record already exists")
)

func translateError(err error) error {
	if err == nil {
		return nil
	}
	if ent.IsNotFound(err) {
		return ErrNotFound
	}
	if ent.IsConstraintError(err) {
		return ErrConflict
	}
	if isUniqueViolation(err) || isForeignKeyViolation(err) {
		return ErrConflict
	}
	return err
}

func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "23505") || strings.Contains(msg, "duplicate key")
}

func isForeignKeyViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "23503") || strings.Contains(msg, "foreign key")
}

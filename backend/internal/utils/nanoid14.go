package utils

import (
	"errors"
	"strconv"
	"strings"

	nanoid "github.com/matoous/go-nanoid/v2"
)

// Fixed nanoid parameters used in the Rails application.
const (
	alphabet = "0123456789abcdefghijklmnopqrstuvwxyz"
	length   = 14
)

// New generates a unique public ID.
func New() (string, error) { return nanoid.Generate(alphabet, length) }

// Must is the same as New, but panics on error.
func Must() string { return nanoid.MustGenerate(alphabet, length) }

// Validate checks if a given field name’s public ID value is valid according to
// the constraints defined by package publicid.
func Validate(fieldName, id string) error {
	if id == "" {
		return errors.New(fieldName + " cannot be blank")
	}

	if len(id) != length {
		return errors.New(fieldName + " should be " + strconv.Itoa(length) + " characters long")
	}

	if strings.Trim(id, alphabet) != "" {
		return errors.New("invalid public ID: " + id + " for field: " + fieldName)
	}

	return nil
}

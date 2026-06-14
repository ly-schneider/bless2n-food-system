// Package id generates and validates Bless2n entity IDs: 12-char nanoids over
// a 55-char alphabet that excludes the confusables 0/O/o, I/i, J/j, L/l.
package id

import (
	"github.com/google/uuid"
	gonanoid "github.com/matoous/go-nanoid/v2"
)

const (
	Alphabet = "123456789ABCDEFGHKMNPQRSTUVWXYZ_abcdefghkmnpqrstuvwxyz-"
	Length   = 12
)

var inAlphabet [256]bool

func init() {
	for i := 0; i < len(Alphabet); i++ {
		inAlphabet[Alphabet[i]] = true
	}
}

func New() string {
	return gonanoid.MustGenerate(Alphabet, Length)
}

func IsNanoID(s string) bool {
	if len(s) != Length {
		return false
	}
	for i := 0; i < len(s); i++ {
		if !inAlphabet[s[i]] {
			return false
		}
	}
	return true
}

// Valid accepts either a nanoid (new rows) or a legacy uuid (rows that
// pre-date the migration), so it can guard references to existing data.
func Valid(s string) bool {
	if IsNanoID(s) {
		return true
	}
	_, err := uuid.Parse(s)
	return err == nil
}

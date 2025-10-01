package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// GenerateRandomURLSafe returns a URL-safe random string with the given number of bytes of entropy.
// The output length is ceil(n*4/3) due to base64url encoding.
func GenerateRandomURLSafe(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("rand: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// HashTokenSHA256 hashes a token with SHA-256 and returns base64url string.
// For high-entropy opaque tokens this is sufficient; comparison must be constant-time.
func HashTokenSHA256(token string) string {
	sum := sha256.Sum256([]byte(token))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// ConstantTimeEquals compares two base64url-encoded hashes in constant time.
func ConstantTimeEquals(a, b string) bool {
	// Decode to fixed-size arrays is not necessary; compare the raw strings length-safe.
	// Use subtle.ConstantTimeCompare on raw bytes.
	ab := []byte(a)
	bb := []byte(b)
	if len(ab) != len(bb) {
		// Still run comparison to avoid timing oracle on length.
		// Pad the shorter slice to the length of the longer with zeros.
		if len(ab) < len(bb) {
			pad := make([]byte, len(bb)-len(ab))
			ab = append(ab, pad...)
		} else {
			pad := make([]byte, len(ab)-len(bb))
			bb = append(bb, pad...)
		}
	}
	// Manual constant time compare
	var v byte
	for i := range ab {
		v |= ab[i] ^ bb[i]
	}
	return v == 0
}

// CSRF helpers
const (
	CSRFCookieName    = "csrf"
	RefreshCookieName = "rt"
	CSRFHeaderName    = "X-CSRF"
)

// GenerateCSRFToken returns a URL-safe random token (128-bit entropy)
func GenerateCSRFToken() (string, error) {
	return GenerateRandomURLSafe(16)
}

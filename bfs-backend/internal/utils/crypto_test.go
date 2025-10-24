package utils

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestGenerateRandomURLSafe(t *testing.T) {
	s, err := GenerateRandomURLSafe(16)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == "" {
		t.Fatalf("expected non-empty string")
	}
	if strings.ContainsAny(s, "+/= ") {
		t.Fatalf("string is not URL-safe: %q", s)
	}
	// ensure it decodes as base64url without padding
	if _, err := base64.RawURLEncoding.DecodeString(s); err != nil {
		t.Fatalf("not base64url: %v", err)
	}
}

func TestHashTokenSHA256_Deterministic(t *testing.T) {
	in := "token-value"
	a := HashTokenSHA256(in)
	b := HashTokenSHA256(in)
	if a != b {
		t.Fatalf("expected deterministic hash, got %q vs %q", a, b)
	}
}

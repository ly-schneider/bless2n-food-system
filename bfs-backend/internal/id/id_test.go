package id

import (
	"strings"
	"testing"
)

func TestNewShape(t *testing.T) {
	for range 1000 {
		got := New()
		if len(got) != Length {
			t.Fatalf("New() = %q, want length %d", got, Length)
		}
		if !IsNanoID(got) {
			t.Fatalf("New() = %q, not a nanoid", got)
		}
		if !Valid(got) {
			t.Fatalf("New() = %q, not Valid", got)
		}
		for i := 0; i < len(got); i++ {
			if !strings.ContainsRune(Alphabet, rune(got[i])) {
				t.Fatalf("New() = %q contains %q outside alphabet", got, got[i])
			}
		}
	}
}

func TestNewUniqueness(t *testing.T) {
	seen := make(map[string]struct{}, 100_000)
	for range 100_000 {
		got := New()
		if _, dup := seen[got]; dup {
			t.Fatalf("duplicate id %q after %d draws", got, len(seen))
		}
		seen[got] = struct{}{}
	}
}

func TestIsNanoID(t *testing.T) {
	valid := []string{"HS79U1yH7Zd3", "------------", "____________", "123456789ABC"}
	for _, s := range valid {
		if !IsNanoID(s) {
			t.Errorf("IsNanoID(%q) = false, want true", s)
		}
	}

	invalid := []string{
		"",
		"short",
		"HS79U1yH7Zd34", // 13 chars
		"tenant.seed1",
		"tenant seed1",
		"OIJLoijl0000",
		"00000000-0000-7000-8000-000000000001",
	}
	for _, s := range invalid {
		if IsNanoID(s) {
			t.Errorf("IsNanoID(%q) = true, want false", s)
		}
	}

	for _, c := range "0OoIiJjLl" {
		s := strings.Repeat(string(c), Length)
		if IsNanoID(s) {
			t.Errorf("IsNanoID(%q) = true, want false (confusable %q)", s, c)
		}
	}
}

// Valid must accept both new nanoids and the legacy uuids we keep in place.
func TestValidAcceptsBothFormats(t *testing.T) {
	valid := []string{
		"HS79U1yH7Zd3",                         // nanoid
		"00000000-0000-7000-8000-000000000001", // legacy uuid
		"019412ab-0000-7000-8000-000000000099",
	}
	for _, s := range valid {
		if !Valid(s) {
			t.Errorf("Valid(%q) = false, want true", s)
		}
	}

	invalid := []string{"", "short", "not-a-uuid-or-nanoid", "tenant seed1"}
	for _, s := range invalid {
		if Valid(s) {
			t.Errorf("Valid(%q) = true, want false", s)
		}
	}
}

func TestAlphabetInvariants(t *testing.T) {
	if len(Alphabet) != 55 {
		t.Fatalf("alphabet length = %d, want 55", len(Alphabet))
	}
	seen := map[rune]bool{}
	for _, c := range Alphabet {
		if seen[c] {
			t.Fatalf("duplicate %q in alphabet", c)
		}
		seen[c] = true
	}
	for _, c := range "0OoIiJjLl" {
		if seen[c] {
			t.Fatalf("confusable %q must not be in alphabet", c)
		}
	}
}

package qrsign

import (
	"crypto/ed25519"
	"encoding/base64"
	"testing"
	"time"
)

// samplePayload expires far in the future so default cases aren't seen as expired.
var verifyAt = time.Unix(1_700_000_100, 0)

func samplePayload() Payload {
	return Payload{
		Version:   Version,
		OrderID:   "order_xyz789",
		IssuedAt:  1_700_000_000,
		ExpiresAt: 4_070_908_800, // ~2099
		Lines: []Line{
			{ProductID: "prod_aaa1111", Quantity: 2},
			{ProductID: "prod_bbb2222", Quantity: 1},
		},
	}
}

func mustKey(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	return pub, priv
}

func TestSignVerifyRoundTrip(t *testing.T) {
	pub, priv := mustKey(t)
	p := samplePayload()

	token, err := Sign(priv, p)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	got, err := Verify(pub, token, verifyAt)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if got.Version != p.Version || got.OrderID != p.OrderID || got.IssuedAt != p.IssuedAt || got.ExpiresAt != p.ExpiresAt {
		t.Fatalf("payload mismatch: got %+v want %+v", got, p)
	}
	if len(got.Lines) != len(p.Lines) {
		t.Fatalf("lines len mismatch: got %d want %d", len(got.Lines), len(p.Lines))
	}
	for i := range p.Lines {
		if got.Lines[i] != p.Lines[i] {
			t.Fatalf("line %d mismatch: got %+v want %+v", i, got.Lines[i], p.Lines[i])
		}
	}
}

func TestSignatureIs64Bytes(t *testing.T) {
	_, priv := mustKey(t)
	token, err := Sign(priv, samplePayload())
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(raw) <= ed25519.SignatureSize {
		t.Fatalf("token too short: %d", len(raw))
	}
	body := raw[:len(raw)-ed25519.SignatureSize]
	if len(body) == 0 {
		t.Fatal("empty body prefix")
	}
}

func TestDeterministicEncoding(t *testing.T) {
	_, priv := mustKey(t)
	p := samplePayload()
	a, err := Sign(priv, p)
	if err != nil {
		t.Fatalf("sign a: %v", err)
	}
	b, err := Sign(priv, p)
	if err != nil {
		t.Fatalf("sign b: %v", err)
	}
	if a != b {
		t.Fatalf("encoding not deterministic:\n a=%s\n b=%s", a, b)
	}
}

func TestTamperedPayloadFails(t *testing.T) {
	pub, priv := mustKey(t)
	token, err := Sign(priv, samplePayload())
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	// Flip a body byte, not a signature byte.
	raw[0] ^= 0x01
	tampered := base64.RawURLEncoding.EncodeToString(raw)

	if _, err := Verify(pub, tampered, verifyAt); err != ErrBadSignature {
		t.Fatalf("expected ErrBadSignature, got %v", err)
	}
}

func TestWrongPublicKeyFails(t *testing.T) {
	_, priv := mustKey(t)
	otherPub, _ := mustKey(t)
	token, err := Sign(priv, samplePayload())
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if _, err := Verify(otherPub, token, verifyAt); err != ErrBadSignature {
		t.Fatalf("expected ErrBadSignature, got %v", err)
	}
}

func TestGarbageToken(t *testing.T) {
	pub, _ := mustKey(t)
	if _, err := Verify(pub, "!!!not-base64!!!", verifyAt); err != ErrMalformed {
		t.Fatalf("expected ErrMalformed, got %v", err)
	}
}

func TestTruncatedToken(t *testing.T) {
	pub, _ := mustKey(t)
	short := base64.RawURLEncoding.EncodeToString([]byte("tooshort"))
	if _, err := Verify(pub, short, verifyAt); err != ErrMalformed {
		t.Fatalf("expected ErrMalformed, got %v", err)
	}
}

func TestExactlySigSizeToken(t *testing.T) {
	pub, _ := mustKey(t)
	// 64 bytes = a signature with an empty body; rejected as malformed.
	exact := base64.RawURLEncoding.EncodeToString(make([]byte, ed25519.SignatureSize))
	if _, err := Verify(pub, exact, verifyAt); err != ErrMalformed {
		t.Fatalf("expected ErrMalformed, got %v", err)
	}
}

func TestVerifyRejectsUnknownVersion(t *testing.T) {
	pub, priv := mustKey(t)
	p := samplePayload()
	p.Version = Version + 1
	token, err := Sign(priv, p)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	// Valid signature but unknown schema version → rejected.
	if _, err := Verify(pub, token, verifyAt); err != ErrUnsupportedVersion {
		t.Fatalf("expected ErrUnsupportedVersion, got %v", err)
	}
}

func TestWrongLengthPublicKey(t *testing.T) {
	_, priv := mustKey(t)
	token, err := Sign(priv, samplePayload())
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if _, err := Verify(ed25519.PublicKey([]byte("short")), token, verifyAt); err != ErrBadSignature {
		t.Fatalf("expected ErrBadSignature, got %v", err)
	}
}

func TestVerifyRejectsExpired(t *testing.T) {
	pub, priv := mustKey(t)
	p := samplePayload()
	p.IssuedAt = 1_700_000_000
	p.ExpiresAt = 1_700_086_399 // end of the issue day
	token, err := Sign(priv, p)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}

	if _, err := Verify(pub, token, time.Unix(p.ExpiresAt, 0)); err != nil {
		t.Fatalf("expected valid at expiry boundary, got %v", err)
	}
	if _, err := Verify(pub, token, time.Unix(p.ExpiresAt+1, 0)); err != ErrExpired {
		t.Fatalf("expected ErrExpired, got %v", err)
	}
}

func TestVerifyRejectsZeroExpiry(t *testing.T) {
	pub, priv := mustKey(t)
	p := samplePayload()
	p.ExpiresAt = 0 // fail closed: a token with no expiry is treated as expired
	token, err := Sign(priv, p)
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	if _, err := Verify(pub, token, verifyAt); err != ErrExpired {
		t.Fatalf("expected ErrExpired, got %v", err)
	}
}

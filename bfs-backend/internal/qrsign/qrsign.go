package qrsign

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
)

// Version is the payload schema version; bumping it requires updating the web
// and Android verifiers in lockstep.
const Version = 1

var (
	ErrMalformed          = errors.New("qrsign: malformed token")
	ErrBadSignature       = errors.New("qrsign: signature verification failed")
	ErrBadPayload         = errors.New("qrsign: invalid payload json")
	ErrUnsupportedVersion = errors.New("qrsign: unsupported payload version")
	ErrExpired            = errors.New("qrsign: token expired")
)

type Line struct {
	ProductID string `json:"p"`
	Quantity  int    `json:"q"`
}

// Field order is part of the signed bytes — never reorder (canonical order v,o,i,e,l).
type Payload struct {
	Version   int    `json:"v"`
	OrderID   string `json:"o"`
	IssuedAt  int64  `json:"i"`
	ExpiresAt int64  `json:"e"`
	Lines     []Line `json:"l"`
}

func Sign(priv ed25519.PrivateKey, p Payload) (string, error) {
	body, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	sig := ed25519.Sign(priv, body)
	token := make([]byte, 0, len(body)+len(sig))
	token = append(token, body...)
	token = append(token, sig...)
	return base64.RawURLEncoding.EncodeToString(token), nil
}

// Verify checks the signature over the literal decoded body bytes, never the
// re-serialized payload (which could drift and break the signature).
func Verify(pub ed25519.PublicKey, token string, now time.Time) (Payload, error) {
	raw, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return Payload{}, ErrMalformed
	}
	// <= not <: a 64-byte token is an empty body and must be rejected.
	if len(raw) <= ed25519.SignatureSize {
		return Payload{}, ErrMalformed
	}
	split := len(raw) - ed25519.SignatureSize
	body, sig := raw[:split], raw[split:]

	if len(pub) != ed25519.PublicKeySize {
		return Payload{}, ErrBadSignature
	}
	if !ed25519.Verify(pub, body, sig) {
		return Payload{}, ErrBadSignature
	}

	var p Payload
	if err := json.Unmarshal(body, &p); err != nil {
		return Payload{}, ErrBadPayload
	}
	// A valid signature does not imply a known schema.
	if p.Version != Version {
		return Payload{}, ErrUnsupportedVersion
	}
	// Fail closed: a token with no/zero expiry is treated as already expired.
	if now.Unix() > p.ExpiresAt {
		return Payload{}, ErrExpired
	}
	return p, nil
}

package service

import (
	"crypto/ed25519"
	"encoding/base64"
	"fmt"

	"backend/internal/config"

	"go.uber.org/zap"
)

// QRKeyService exposes the deployment's Ed25519 keypair. The 32-byte seed comes
// from QR_ED25519_PRIVATE_SEED (base64); the public key is served at GET /v1/qr-config.
type QRKeyService interface {
	SigningKey() (ed25519.PrivateKey, ed25519.PublicKey)
	PublicKey() ed25519.PublicKey
}

type qrKeyService struct {
	priv ed25519.PrivateKey
	pub  ed25519.PublicKey
}

// NewQRKeyService derives the keypair from QR_ED25519_PRIVATE_SEED. The seed is
// mandatory — a missing or invalid one fails startup rather than letting the pod
// create unredeemable orders or sign with a bad key.
func NewQRKeyService(cfg config.Config, logger *zap.Logger) (QRKeyService, error) {
	seedB64 := cfg.QRSigning.Ed25519PrivateSeed
	if seedB64 == "" {
		return nil, fmt.Errorf("QR_ED25519_PRIVATE_SEED is required: it signs the offline-verifiable pickup QR token on every order")
	}

	seed, err := base64.StdEncoding.DecodeString(seedB64)
	if err != nil {
		return nil, fmt.Errorf("qr signing key: decode QR_ED25519_PRIVATE_SEED: %w", err)
	}
	if len(seed) != ed25519.SeedSize {
		return nil, fmt.Errorf("qr signing key: QR_ED25519_PRIVATE_SEED must decode to %d bytes, got %d", ed25519.SeedSize, len(seed))
	}

	priv := ed25519.NewKeyFromSeed(seed)
	pub, ok := priv.Public().(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("qr signing key: unexpected public key type")
	}

	logger.Info("QR signing enabled", zap.String("publicKey", base64.RawURLEncoding.EncodeToString(pub)))
	return &qrKeyService{priv: priv, pub: pub}, nil
}

func (s *qrKeyService) SigningKey() (ed25519.PrivateKey, ed25519.PublicKey) {
	return s.priv, s.pub
}

func (s *qrKeyService) PublicKey() ed25519.PublicKey {
	return s.pub
}

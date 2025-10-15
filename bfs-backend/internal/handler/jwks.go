package handler

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"

	"backend/internal/domain"
	"backend/internal/service"
)

type JWKSHandler struct {
	jwtService service.JWTService
	logger     *zap.Logger
}

func NewJWKSHandler(jwtService service.JWTService, logger *zap.Logger) *JWKSHandler {
	return &JWKSHandler{
		jwtService: jwtService,
		logger:     logger,
	}
}

// GetJWKS godoc
// @Summary JSON Web Key Set
// @Tags auth
// @Produce json
// @Success 200 {object} domain.JWKS
// @Router /.well-known/jwks.json [get]
func (h *JWKSHandler) GetJWKS(w http.ResponseWriter, r *http.Request) {
	publicKey, keyID, err := h.getPublicKeyInfo()
	if err != nil {
		h.logger.Error("Failed to get public key info", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	x := base64.RawURLEncoding.EncodeToString(publicKey)

	jwk := domain.JWK{
		KeyType:   "OKP",
		Curve:     "Ed25519",
		KeyID:     keyID,
		Algorithm: "EdDSA",
		Use:       "sig",
		X:         x,
	}

	jwks := domain.JWKS{
		Keys: []domain.JWK{jwk},
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if err := json.NewEncoder(w).Encode(jwks); err != nil {
		h.logger.Error("Failed to write JWKS response", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (h *JWKSHandler) getPublicKeyInfo() (ed25519.PublicKey, string, error) {
	publicKey := h.jwtService.GetPublicKey()

	keyID := base64.RawURLEncoding.EncodeToString(publicKey[:8])

	return publicKey, keyID, nil
}

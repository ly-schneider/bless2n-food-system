package api

import (
	"encoding/base64"
	"net/http"

	"backend/internal/generated/api/generated"
	"backend/internal/qrsign"
	"backend/internal/response"
)

// GetQRConfig serves the public Ed25519 verification key for offline pickup-token
// verification. No auth.
func (h *Handlers) GetQRConfig(w http.ResponseWriter, r *http.Request) {
	pub := h.qrKeys.PublicKey()

	response.WriteJSON(w, http.StatusOK, generated.QRConfig{
		Alg:            "ed25519",
		PayloadVersion: qrsign.Version,
		PublicKey:      base64.RawURLEncoding.EncodeToString(pub),
	})
}

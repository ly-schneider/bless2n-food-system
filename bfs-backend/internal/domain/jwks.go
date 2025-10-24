package domain

// JWKS represents a JSON Web Key Set response
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key for Ed25519
type JWK struct {
	KeyType   string `json:"kty"` // Key Type - "OKP" for Ed25519
	Curve     string `json:"crv"` // Curve - "Ed25519"
	KeyID     string `json:"kid"` // Key ID
	Algorithm string `json:"alg"` // Algorithm - "EdDSA"
	Use       string `json:"use"` // Public Key Use - "sig"
	X         string `json:"x"`   // Base64url-encoded public key
}

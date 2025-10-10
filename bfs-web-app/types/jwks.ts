export interface JWK {
  kty: string // Key Type - "OKP" for Ed25519
  crv: string // Curve - "Ed25519"
  kid: string // Key ID
  alg: string // Algorithm - "EdDSA"
  use: string // Public Key Use - "sig"
  x: string // Base64url-encoded public key
}

export interface JWKS {
  keys: JWK[]
}

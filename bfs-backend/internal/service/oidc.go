package service

import (
    "context"
    "crypto/ecdsa"
    "crypto/elliptic"
    "crypto/rsa"
    "crypto/sha256"
    "encoding/base64"
    "encoding/binary"
    "encoding/json"
    "errors"
    "fmt"
    "math/big"
    "net/http"
    "strings"
    "sync"
    "time"

    jwt "github.com/golang-jwt/jwt/v5"
)

// jwksCache caches JWKS per URL with a TTL.
type jwksCache struct{
    mu sync.Mutex
    entries map[string]*jwksCached
}
type jwksCached struct{
    keys map[string]any // kid -> *rsa.PublicKey or *ecdsa.PublicKey
    expiresAt time.Time
}

func newJWKSCache() *jwksCache { return &jwksCache{entries: make(map[string]*jwksCached)} }

func (c *jwksCache) getKey(ctx context.Context, jwksURL, kid string) (any, error) {
    now := time.Now()
    c.mu.Lock()
    entry := c.entries[jwksURL]
    if entry != nil && now.Before(entry.expiresAt) {
        if k, ok := entry.keys[kid]; ok { c.mu.Unlock(); return k, nil }
    }
    c.mu.Unlock()
    // refresh
    req, _ := http.NewRequestWithContext(ctx, "GET", jwksURL, nil)
    resp, err := http.DefaultClient.Do(req)
    if err != nil { return nil, err }
    defer func() { _ = resp.Body.Close() }()
    var body struct{ Keys []json.RawMessage `json:"keys"` }
    if err := json.NewDecoder(resp.Body).Decode(&body); err != nil { return nil, err }
    keys := make(map[string]any)
    for _, raw := range body.Keys {
        var hdr struct{ Kty, Kid, Crv, Alg, N, E, X, Y string }
        _ = json.Unmarshal(raw, &hdr)
        if hdr.Kid == "" { continue }
        switch strings.ToUpper(hdr.Kty) {
        case "RSA":
            if hdr.N == "" || hdr.E == "" { continue }
            nBytes, errN := base64.RawURLEncoding.DecodeString(hdr.N)
            eBytes, errE := base64.RawURLEncoding.DecodeString(hdr.E)
            if errN != nil || errE != nil { continue }
            n := new(big.Int).SetBytes(nBytes)
            var e int
            // Big-endian to int
            if len(eBytes) == 3 { e = int(binary.BigEndian.Uint32(append([]byte{0}, eBytes...))) } else if len(eBytes) == 4 { e = int(binary.BigEndian.Uint32(eBytes)) } else if len(eBytes) == 1 { e = int(eBytes[0]) } else { continue }
            keys[hdr.Kid] = &rsa.PublicKey{N: n, E: e}
        case "EC":
            // EC (e.g., ES256) support
            if hdr.X == "" || hdr.Y == "" { continue }
            xBytes, errX := base64.RawURLEncoding.DecodeString(hdr.X)
            yBytes, errY := base64.RawURLEncoding.DecodeString(hdr.Y)
            if errX != nil || errY != nil { continue }
            x := new(big.Int).SetBytes(xBytes)
            y := new(big.Int).SetBytes(yBytes)
            keys[hdr.Kid] = &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}
        }
    }
    // determine TTL from Cache-Control max-age if provided; default 1 hour
    ttl := time.Hour
    if cc := resp.Header.Get("Cache-Control"); cc != "" {
        for _, part := range strings.Split(cc, ",") {
            p := strings.TrimSpace(part)
            if strings.HasPrefix(strings.ToLower(p), "max-age=") {
                v := strings.TrimPrefix(p, "max-age=")
                if sec, err := time.ParseDuration(v+"s"); err == nil && sec > 0 { ttl = sec }
            }
        }
    }
    c.mu.Lock()
    c.entries[jwksURL] = &jwksCached{keys: keys, expiresAt: now.Add(ttl)}
    k := keys[kid]
    c.mu.Unlock()
    if k == nil { return nil, fmt.Errorf("kid not found") }
    return k, nil
}

// OIDC standard claims we care about
type oidcClaims struct {
    Iss            string        `json:"iss"`
    Aud            any           `json:"aud"` // string or array
    Exp            int64         `json:"exp"`
    Iat            int64         `json:"iat"`
    Sub            string        `json:"sub"`
    Email          string        `json:"email,omitempty"`
    EmailVerified  any           `json:"email_verified,omitempty"`
    Name           string        `json:"name,omitempty"`
    Picture        string        `json:"picture,omitempty"`
    Nonce          string        `json:"nonce,omitempty"`
}

func audContains(aud any, clientID string) bool {
    switch v := aud.(type) {
    case string:
        return v == clientID
    case []any:
        for _, e := range v {
            if s, ok := e.(string); ok && s == clientID { return true }
        }
    case []string:
        for _, s := range v { if s == clientID { return true } }
    }
    return false
}

// VerifyIDToken verifies a JWT against JWKS and basic OIDC claims.
func VerifyIDToken(ctx context.Context, tokenString string, jwksURL string, expectedIss []string, expectedAud string, expectedNonce string, cache *jwksCache) (*oidcClaims, error) {
    var mc jwt.MapClaims
    parsed, err := jwt.ParseWithClaims(tokenString, &mc, func(t *jwt.Token) (any, error) {
        // enforce allowed algs
        alg := t.Method.Alg()
        if alg != jwt.SigningMethodRS256.Alg() && alg != jwt.SigningMethodES256.Alg() {
            return nil, errors.New("unsupported_alg")
        }
        kid, _ := t.Header["kid"].(string)
        if kid == "" { return nil, errors.New("missing_kid") }
        key, err := cache.getKey(ctx, jwksURL, kid)
        if err != nil { return nil, fmt.Errorf("jwks_fetch: %w", err) }
        return key, nil
    }, jwt.WithIssuedAt(), jwt.WithAudience(expectedAud))
    if err != nil || !parsed.Valid { return nil, errors.New("invalid_signature") }

    // Map to oidcClaims
    oc := &oidcClaims{}
    if iss, _ := mc["iss"].(string); iss != "" { oc.Iss = iss }
    oc.Aud = mc["aud"]
    if expf, ok := mc["exp"].(float64); ok { oc.Exp = int64(expf) }
    if iatf, ok := mc["iat"].(float64); ok { oc.Iat = int64(iatf) }
    if sub, _ := mc["sub"].(string); sub != "" { oc.Sub = sub }
    if email, _ := mc["email"].(string); email != "" { oc.Email = email }
    if ev, ok := mc["email_verified"]; ok { oc.EmailVerified = ev }
    if name, _ := mc["name"].(string); name != "" { oc.Name = name }
    if picture, _ := mc["picture"].(string); picture != "" { oc.Picture = picture }
    if nonce, _ := mc["nonce"].(string); nonce != "" { oc.Nonce = nonce }

    // standard checks
    now := time.Now().Unix()
    if oc.Exp <= now { return nil, errors.New("expired") }
    if oc.Iat > now+60 { return nil, errors.New("invalid_iat") }
    okIss := false
    for _, iss := range expectedIss { if oc.Iss == iss { okIss = true; break } }
    if !okIss { return nil, errors.New("invalid_iss") }
    if !audContains(oc.Aud, expectedAud) { return nil, errors.New("invalid_aud") }
    if expectedNonce != "" && oc.Nonce != expectedNonce { return nil, errors.New("invalid_nonce") }
    return oc, nil
}

// Utility: derive code challenge from verifier
func CodeChallengeS256(verifier string) string {
    sum := sha256.Sum256([]byte(verifier))
    return base64.RawURLEncoding.EncodeToString(sum[:])
}

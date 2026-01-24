package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

var (
	ErrJWKSFetch      = errors.New("failed to fetch JWKS")
	ErrKeyNotFound    = errors.New("key not found in JWKS")
	ErrUnsupportedAlg = errors.New("unsupported key algorithm")
)

// JWKS represents a JSON Web Key Set.
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key.
type JWK struct {
	Kty string `json:"kty"` // Key Type (RSA, EC)
	Use string `json:"use"` // Public Key Use (sig)
	Kid string `json:"kid"` // Key ID
	Alg string `json:"alg"` // Algorithm (RS256, ES256)
	N   string `json:"n"`   // RSA modulus
	E   string `json:"e"`   // RSA exponent
	Crv string `json:"crv"` // EC curve (P-256)
	X   string `json:"x"`   // EC x coordinate
	Y   string `json:"y"`   // EC y coordinate
}

// JWKSClient fetches and caches JWKS from a remote endpoint.
type JWKSClient struct {
	url        string
	httpClient *http.Client
	logger     *zap.Logger

	mu          sync.RWMutex
	cache       *JWKS
	lastFetch   time.Time
	cacheTTL    time.Duration
	refreshDone chan struct{}
}

// NewJWKSClient creates a new JWKS client.
func NewJWKSClient(baseURL string, logger *zap.Logger) *JWKSClient {
	return &JWKSClient{
		url: baseURL + "/.well-known/jwks.json",
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger:   logger,
		cacheTTL: 5 * time.Minute,
	}
}

// Start begins the background JWKS refresh routine.
func (c *JWKSClient) Start(ctx context.Context) error {
	// Initial fetch
	if err := c.refresh(ctx); err != nil {
		return fmt.Errorf("initial JWKS fetch failed: %w", err)
	}

	c.refreshDone = make(chan struct{})

	// Start background refresh
	go c.backgroundRefresh(ctx)

	return nil
}

// Stop stops the background JWKS refresh routine.
func (c *JWKSClient) Stop() {
	if c.refreshDone != nil {
		close(c.refreshDone)
	}
}

func (c *JWKSClient) backgroundRefresh(ctx context.Context) {
	ticker := time.NewTicker(c.cacheTTL)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.refreshDone:
			return
		case <-ticker.C:
			if err := c.refresh(ctx); err != nil {
				c.logger.Warn("failed to refresh JWKS, using cached version",
					zap.Error(err),
					zap.Time("last_fetch", c.lastFetch),
				)
			}
		}
	}
}

func (c *JWKSClient) refresh(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: status %d", ErrJWKSFetch, resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("decode JWKS: %w", err)
	}

	c.mu.Lock()
	c.cache = &jwks
	c.lastFetch = time.Now()
	c.mu.Unlock()

	c.logger.Debug("refreshed JWKS cache",
		zap.Int("key_count", len(jwks.Keys)),
	)

	return nil
}

// GetKey retrieves a key by its ID from the cached JWKS.
func (c *JWKSClient) GetKey(kid string) (*JWK, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.cache == nil {
		return nil, ErrJWKSFetch
	}

	for i := range c.cache.Keys {
		if c.cache.Keys[i].Kid == kid {
			return &c.cache.Keys[i], nil
		}
	}

	return nil, ErrKeyNotFound
}

// GetRSAPublicKey converts a JWK to an RSA public key.
func (jwk *JWK) GetRSAPublicKey() (*rsa.PublicKey, error) {
	if jwk.Kty != "RSA" {
		return nil, fmt.Errorf("%w: expected RSA, got %s", ErrUnsupportedAlg, jwk.Kty)
	}

	// Decode n (modulus)
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("decode modulus: %w", err)
	}
	n := new(big.Int).SetBytes(nBytes)

	// Decode e (exponent)
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("decode exponent: %w", err)
	}
	e := int(new(big.Int).SetBytes(eBytes).Int64())

	return &rsa.PublicKey{
		N: n,
		E: e,
	}, nil
}

// GetECDSAPublicKey converts a JWK to an ECDSA public key.
func (jwk *JWK) GetECDSAPublicKey() (*ecdsa.PublicKey, error) {
	if jwk.Kty != "EC" {
		return nil, fmt.Errorf("%w: expected EC, got %s", ErrUnsupportedAlg, jwk.Kty)
	}

	// Decode x coordinate
	xBytes, err := base64.RawURLEncoding.DecodeString(jwk.X)
	if err != nil {
		return nil, fmt.Errorf("decode x coordinate: %w", err)
	}
	x := new(big.Int).SetBytes(xBytes)

	// Decode y coordinate
	yBytes, err := base64.RawURLEncoding.DecodeString(jwk.Y)
	if err != nil {
		return nil, fmt.Errorf("decode y coordinate: %w", err)
	}
	y := new(big.Int).SetBytes(yBytes)

	// Get curve based on crv field
	curve, err := getCurve(jwk.Crv)
	if err != nil {
		return nil, err
	}

	return &ecdsa.PublicKey{
		Curve: curve,
		X:     x,
		Y:     y,
	}, nil
}

func getCurve(crv string) (elliptic.Curve, error) {
	switch crv {
	case "P-256":
		return elliptic.P256(), nil
	case "P-384":
		return elliptic.P384(), nil
	case "P-521":
		return elliptic.P521(), nil
	default:
		return nil, fmt.Errorf("%w: unsupported curve %s", ErrUnsupportedAlg, crv)
	}
}

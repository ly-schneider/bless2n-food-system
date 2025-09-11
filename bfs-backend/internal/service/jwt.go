package service

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"backend/internal/domain"
)

const (
	AccessTokenDuration  = 10 * time.Minute
	RefreshTokenDuration = 7 * 24 * time.Hour
	RefreshTokenLength   = 32

	defaultAudience = "blessthun-food-api"
)

type TokenClaims struct {
	Role domain.UserRole `json:"role"`
	jwt.RegisteredClaims
}

type JWTService interface {
	GenerateAccessToken(user *domain.User) (string, error)
	GenerateRefreshToken() (string, error)
	ValidateAccessToken(tokenString string) (*TokenClaims, error)
	GenerateTokenPair(user *domain.User, clientID string) (*TokenPairResponse, error)
}

type TokenPairResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type jwtService struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
	issuer     string
	audience   string
}

func NewJWTService(jwtPrivPemPath string, jwtPubPemPath string, issuer string) JWTService {
	priv, pub, err := loadEd25519Keys(jwtPrivPemPath, jwtPubPemPath)
	if err != nil {
		panic(fmt.Sprintf("failed to load Ed25519 keys: %v", err))
	}
	return &jwtService{
		privateKey: priv,
		publicKey:  pub,
		issuer:     issuer,
		audience:   defaultAudience,
	}
}

func loadEd25519Keys(privPath, pubPath string) (ed25519.PrivateKey, ed25519.PublicKey, error) {
	privPEM, err := os.ReadFile(privPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read private key: %w", err)
	}
	block, _ := pem.Decode(privPEM)
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, nil, fmt.Errorf("invalid private key PEM: got %v", block)
	}
	anyKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("parse PKCS#8 private key: %w", err)
	}
	priv, ok := anyKey.(ed25519.PrivateKey)
	if !ok {
		return nil, nil, fmt.Errorf("not an Ed25519 private key")
	}
	if l := len(priv); l != ed25519.PrivateKeySize {
		return nil, nil, fmt.Errorf("bad ed25519 key length: %d", l)
	}

	pubPEM, err := os.ReadFile(pubPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read public key: %w", err)
	}
	pb, _ := pem.Decode(pubPEM)
	if pb == nil || pb.Type != "PUBLIC KEY" {
		return nil, nil, fmt.Errorf("invalid public key PEM: got %v", pb)
	}
	anyPub, err := x509.ParsePKIXPublicKey(pb.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("parse PKIX public key: %w", err)
	}
	pub, ok2 := anyPub.(ed25519.PublicKey)
	if !ok2 {
		return nil, nil, fmt.Errorf("not an Ed25519 public key")
	}
	if l := len(pub); l != ed25519.PublicKeySize {
		return nil, nil, fmt.Errorf("bad ed25519 key length: %d", l)
	}

	return priv, pub, nil
}

func (j *jwtService) GenerateAccessToken(user *domain.User) (string, error) {
	now := time.Now().UTC()

	jti, err := generateJTI()
	if err != nil {
		return "", fmt.Errorf("failed to generate JTI: %w", err)
	}

	claims := TokenClaims{
		Role: user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Subject:   user.ID.Hex(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(AccessTokenDuration)),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    j.issuer,
			Audience:  []string{j.audience},
		},
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	return tok.SignedString(j.privateKey)
}

func (j *jwtService) GenerateRefreshToken() (string, error) {
	// 32 chars from a 64-symbol alphabet → 192 bits of entropy
	bytes := make([]byte, RefreshTokenLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	out := make([]byte, RefreshTokenLength)
	for i, b := range bytes {
		out[i] = chars[int(b)%len(chars)]
	}
	return string(out), nil
}

func (j *jwtService) ValidateAccessToken(tokenString string) (*TokenClaims, error) {
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodEdDSA.Alg()}),
		jwt.WithIssuer(j.issuer),
		jwt.WithAudience(j.audience),
	)

	var claims TokenClaims
	token, err := parser.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodEdDSA {
			return nil, fmt.Errorf("unexpected signing alg: %v", t.Header["alg"])
		}
		return j.publicKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse/validate token: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return &claims, nil
}

func (j *jwtService) GenerateTokenPair(user *domain.User, clientID string) (*TokenPairResponse, error) {
	accessToken, err := j.GenerateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := j.GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPairResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(AccessTokenDuration.Seconds()),
	}, nil
}

func generateJTI() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("at_%x", b), nil
}

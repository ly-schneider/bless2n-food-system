package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"backend/internal/config"
	"backend/internal/domain"
	"backend/internal/logger"
	"backend/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error)
	Login(ctx context.Context, req *LoginRequest, ua string) (*BrowserLoginResp, *MobileLoginResp, error)
	Refresh(ctx context.Context, rt string, ua string) (*BrowserLoginResp, *MobileLoginResp, error)
}

type RegisterRequest struct {
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type RegisterResponse struct {
	User    domain.User `json:"user"`
	Message string      `json:"message"`
}

type BrowserLoginResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"-"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type MobileLoginResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type authService struct {
	userSvc         UserService
	refreshTokenSvc RefreshTokenService
}

func NewAuthService(u UserService, r RefreshTokenService) AuthService {
	return &authService{userSvc: u, refreshTokenSvc: r}
}

func (s *authService) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {

	// TODO: Implement registration logic:
	// - Hash password
	// - Check if user exists
	// - Create user
	// - Send verification email
	// - Generate tokens

	logger.L.Infow("processing user registration", "email", req.Email)

	// Validate request
	if req.FirstName == "" || req.LastName == "" || req.Email == "" || req.Password == "" {
		err := errors.New("invalid registration request: missing required fields")
		logger.L.Error(err.Error())
		return nil, err
	}

	// Check if user already exists
	existingUser, err := s.userSvc.GetByEmail(ctx, req.Email)
	if err != nil {
		logger.L.Errorw("error checking existing user", "email", req.Email, "error", err)
		return nil, err
	}
	if existingUser.ID != uuid.Nil {
		err := errors.New("user already exists with this email")
		logger.L.Errorw(err.Error(), "email", req.Email)
		return nil, err
	}

	logger.L.Infow("No existing user found, proceeding with registration", "email", req.Email)

	user := domain.User{
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Email:      req.Email,
		IsVerified: false,
		RoleID:     domain.Roles["user"].ID, // Default to user role
	}

	// Hash the password using bcrypt
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.L.Errorw("error hashing password", "email", req.Email, "error", err)
		return nil, errors.New("failed to hash password")
	}
	user.PasswordHash = string(passwordHash)

	// Create the user
	if err := s.userSvc.Create(ctx, &user); err != nil {
		logger.L.Errorw("error creating user", "email", req.Email, "error", err)
		return nil, err
	}

	// TODO: Send verification email (not implemented here)

	logger.L.Infow("User registration processed", "email", req.Email)

	return &RegisterResponse{
		User:    user,
		Message: "Registration successful. Please verify your email.",
	}, nil
}

func (s *authService) Login(ctx context.Context, req *LoginRequest, ua string) (*BrowserLoginResp, *MobileLoginResp, error) {

	// Validate request
	if req.Email == "" || req.Password == "" {
		err := errors.New("invalid login request: missing required fields")
		logger.L.Error(err.Error())
		return nil, nil, err
	}

	// Check if user exists
	user, err := s.userSvc.GetByEmail(ctx, req.Email)
	if err != nil {
		logger.L.Errorw("error fetching user", "email", req.Email, "error", err)
		return nil, nil, err
	}
	if user.ID == uuid.Nil {
		err := errors.New("user not found")
		logger.L.Errorw(err.Error(), "email", req.Email)
		return nil, nil, err
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		logger.L.Errorw("password verification failed", "email", req.Email, "error", err)
		return nil, nil, errors.New("invalid credentials")
	}

	// Issue access-token (15min) + refresh-token (30d)
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role.Name,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(15 * time.Minute).Unix(),
	})

	accessTokenString, err := accessToken.SignedString([]byte(config.Load().App.JWTSecretKey))
	if err != nil {
		logger.L.Errorw("error signing access token", "email", req.Email, "error", err)
		return nil, nil, errors.New("failed to generate access token")
	}

	// Generate refresh token and store the hash
	plainRefreshToken := randomString(64)
	hash := sha256.Sum256([]byte(plainRefreshToken))

	refreshToken := &domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: hex.EncodeToString(hash[:]),
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour), // 30 days
		Revoked:   false,
	}
	if err := s.refreshTokenSvc.Create(ctx, refreshToken); err != nil {
		logger.L.Errorw("error creating refresh token", "email", req.Email, "error", err)
		return nil, nil, errors.New("failed to create refresh token")
	}

	logger.L.Infow("User logged in successfully", "email", req.Email, "user_id", user.ID)

	if utils.IsMobile(ua) {
		return nil, &MobileLoginResp{
			AccessToken:  accessTokenString,
			RefreshToken: plainRefreshToken,
			ExpiresIn:    30 * 24 * 60 * 60, // 30 days in seconds
		}, nil
	}

	return &BrowserLoginResp{
		AccessToken:  accessTokenString,
		RefreshToken: plainRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    15 * 60, // 15 minutes in seconds
	}, nil, nil
}

func (s *authService) Refresh(ctx context.Context, rt string, ua string) (*BrowserLoginResp, *MobileLoginResp, error) {
	// Validate request
	if rt == "" {
		err := errors.New("invalid refresh token request: missing refresh token")
		logger.L.Error(err.Error())
		return nil, nil, err
	}

	// Hash the refresh token
	hash := sha256.Sum256([]byte(rt))
	hashedRT := hex.EncodeToString(hash[:])

	// Check if refresh token exists
	refreshToken, err := s.refreshTokenSvc.GetByTokenHash(ctx, hashedRT)
	if err != nil {
		logger.L.Errorw("error fetching refresh token", "refresh_token", rt, "error", err)
		return nil, nil, err
	}
	if refreshToken.ID == uuid.Nil {
		err := errors.New("refresh token not found")
		logger.L.Errorw(err.Error(), "refresh_token", rt)
		return nil, nil, err
	}

	// Check if refresh token is revoked or expired
	if refreshToken.Revoked || refreshToken.ExpiresAt.Before(time.Now()) {
		err := errors.New("refresh token is revoked or expired")
		logger.L.Errorw(err.Error(), "refresh_token", rt)
		return nil, nil, err
	}

	user, err := s.userSvc.Get(ctx, refreshToken.UserID)
	if err != nil {
		logger.L.Errorw("error fetching user", "user_id", refreshToken.UserID, "error", err)
		return nil, nil, err
	}
	if user.ID == uuid.Nil {
		err := errors.New("user not found")
		logger.L.Errorw(err.Error(), "user_id", refreshToken.UserID)
		return nil, nil, err
	}

	// Rotate
	// Issue access-token (15min) + refresh-token (30d)
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role.Name,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(15 * time.Minute).Unix(),
	})

	accessTokenString, err := accessToken.SignedString([]byte(config.Load().App.JWTSecretKey))
	if err != nil {
		logger.L.Errorw("error signing access token", "user_id", user.ID, "error", err)
		return nil, nil, errors.New("failed to generate access token")
	}

	// Generate refresh token and store the hash
	plainRefreshToken := randomString(64)
	newHashedRT := sha256.Sum256([]byte(plainRefreshToken))

	newRefreshToken := &domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: hex.EncodeToString(newHashedRT[:]),
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour), // 30 days
		Revoked:   false,
	}
	if err := s.refreshTokenSvc.Create(ctx, newRefreshToken); err != nil {
		logger.L.Errorw("error creating new refresh token", "user_id", user.ID, "error", err)
		return nil, nil, errors.New("failed to create refresh token")
	}

	// Revoke the old refresh token
	if err := s.refreshTokenSvc.Revoke(ctx, refreshToken.ID); err != nil {
		logger.L.Errorw("error revoking old refresh token", "refresh_token_id", refreshToken.ID, "error", err)
		return nil, nil, errors.New("failed to revoke old refresh token")
	}

	logger.L.Infow("User refreshed tokens successfully", "user_id", user.ID)

	if utils.IsMobile(ua) {
		return nil, &MobileLoginResp{
			AccessToken:  accessTokenString,
			RefreshToken: plainRefreshToken,
			ExpiresIn:    30 * 24 * 60 * 60, // 30 days in seconds
		}, nil
	}

	return &BrowserLoginResp{
		AccessToken:  accessTokenString,
		RefreshToken: plainRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    15 * 60, // 15 minutes in seconds
	}, nil, nil
}

func randomString(n int) string { // hex string of n bytes = 2n chars
	b := make([]byte, n)
	_, _ = rand.Read(b) // crypto-secure RNG
	return hex.EncodeToString(b)
}

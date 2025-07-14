package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"time"

	"backend/internal/apperrors"
	"backend/internal/config"
	"backend/internal/domain"
	"backend/internal/http/middleware"
	"backend/internal/jobs"
	"backend/internal/logger"
	"backend/internal/repository"
	"backend/internal/utils"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService interface {
	Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, *apperrors.APIError)
	Login(ctx context.Context, req *LoginRequest) (*BrowserLoginResp, *MobileLoginResp, *apperrors.APIError)
	Refresh(ctx context.Context, refreshToken string) (*BrowserLoginResp, *MobileLoginResp, *apperrors.APIError)
	Logout(ctx context.Context, refreshToken string) *apperrors.APIError
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
	userRepo         repository.UserRepository
	auditLogRepo     repository.AuditLogRepository
	refreshTokenRepo repository.RefreshTokenRepository
	jobService       *jobs.JobService
}

func NewAuthService(ur repository.UserRepository, alr repository.AuditLogRepository, rtr repository.RefreshTokenRepository, js *jobs.JobService) AuthService {
	return &authService{
		userRepo:         ur,
		auditLogRepo:     alr,
		refreshTokenRepo: rtr,
		jobService:       js,
	}
}

var (
	errEmailTaken       = errors.New("email already registered")
	errHashFailed       = errors.New("failed to hash password")
	errCreateUserFailed = errors.New("failed to create user")
)

func (s *authService) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, *apperrors.APIError) {
	if req.FirstName == "" || req.LastName == "" || req.Email == "" || req.Password == "" {
		logger.L.Error("missing required fields in registration request", "request", req)
		return nil, apperrors.BadRequest("invalid_body", "missing required fields", domain.ErrInvalidBodyMissingFields)
	}

	if _, err := s.userRepo.GetByEmail(ctx, req.Email); err == nil {
		return nil, apperrors.FromStatus(http.StatusConflict, "email_already_registered", errEmailTaken)
	} else if !errors.Is(err, domain.ErrUserNotFound) {
		return nil, apperrors.FromStatus(http.StatusInternalServerError, "failed to check user existence", err)
	}

	logger.L.Infow("No existing user found, proceeding with registration", "email", req.Email)

	user := domain.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		RoleID:    domain.Roles["user"].ID,
	}

	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		logger.L.Errorw("error hashing password", "email", req.Email, "error", err)
		return nil, apperrors.FromStatus(http.StatusInternalServerError, "there was an internal error", errHashFailed)
	}
	user.PasswordHash = passwordHash


	if err := s.userRepo.Create(ctx, &user); err != nil {
		logger.L.Errorw("error creating user", "email", req.Email, "error", err)
		return nil, apperrors.FromStatus(http.StatusInternalServerError, "there was an internal error", errCreateUserFailed)
	}

	auditLog := domain.AuditLog{
		UserID:   user.ID,
		PublicIP: middleware.ExtractIPFromContext(ctx),
		Event:    domain.EventUserCreated,
	}

	if err := s.auditLogRepo.Create(ctx, &auditLog); err != nil {
		logger.L.Errorw("error creating audit log for user registration", "user_id", user.ID, "email", req.Email, "error", err)
		return nil, apperrors.FromStatus(http.StatusInternalServerError, "failed to log user registration", err)
	}

	emailPayload := &jobs.EmailVerificationPayload{
		UserID:      string(user.ID),
		Email:       user.Email,
		FirstName:   user.FirstName,
		Token:       randomString(32),
		RequestedAt: time.Now().Unix(),
	}
	if err := s.jobService.EnqueueEmailVerification(ctx, emailPayload, 0); err != nil {
		logger.L.Errorw("failed to enqueue email verification job", "user_id", user.ID, "email", req.Email, "error", err)
	}

	logger.L.Infow("User registration processed", "email", req.Email)

	return &RegisterResponse{
		User:    user,
		Message: "Registration successful. Please verify your email.",
	}, nil
}

func (s *authService) Login(ctx context.Context, req *LoginRequest) (*BrowserLoginResp, *MobileLoginResp, *apperrors.APIError) {
	if req.Email == "" || req.Password == "" {
		logger.L.Error("missing required fields in login request", "request", req)
		return nil, nil, apperrors.BadRequest("invalid_body", "missing required fields", domain.ErrInvalidBodyMissingFields)
	}

	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, nil, apperrors.Unauthorized("email or password is incorrect")
		}
		return nil, nil, apperrors.FromStatus(http.StatusInternalServerError, "failed to retrieve user", err)
	}

	logger.L.Infow("User found for login", "user", user)

	passwordValid := utils.VerifyPassword(req.Password, user.PasswordHash)

	if !passwordValid {
		return nil, nil, apperrors.Unauthorized("email or password is incorrect")
	}

	// Issue access-token (15min) + refresh-token (30d)
	var roleName string
	if user.Role != nil {
		roleName = user.Role.Name
	} else {
		// Fallback: load role if not preloaded
		logger.L.Warnw("Role not preloaded, loading role by ID", "user_id", user.ID, "role_id", user.RoleID)
		// Map role ID to role name
		switch user.RoleID {
		case 1:
			roleName = "admin"
		case 2:
			roleName = "user"
		default:
			logger.L.Errorw("Unknown role ID", "user_id", user.ID, "role_id", user.RoleID)
			return nil, nil, apperrors.FromStatus(http.StatusInternalServerError, "invalid user role", errors.New("unknown role"))
		}
	}
	
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID,
		"role": roleName,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(15 * time.Minute).Unix(),
	})

	accessTokenString, err := accessToken.SignedString([]byte(config.Load().App.JWTSecretKey))
	if err != nil {
		logger.L.Errorw("error signing access token", "email", req.Email, "error", err)
		return nil, nil, apperrors.FromStatus(http.StatusInternalServerError, "failed to sign access token", err)
	}

	// Generate refresh token and store the hash
	plainRefreshToken := randomString(64)
	tokenHash, err := utils.HashToken(plainRefreshToken)
	if err != nil {
		logger.L.Errorw("error hashing refresh token", "email", req.Email, "error", err)
		return nil, nil, apperrors.FromStatus(http.StatusInternalServerError, "failed to create refresh token", err)
	}

	refreshToken := &domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
	if err := s.refreshTokenRepo.Create(ctx, refreshToken); err != nil {
		logger.L.Errorw("error creating refresh token", "email", req.Email, "error", err)
		return nil, nil, apperrors.FromStatus(http.StatusInternalServerError, "failed to create refresh token", err)
	}

	logger.L.Infow("User logged in successfully", "email", req.Email, "user_id", user.ID)

	auditLog := domain.AuditLog{
		UserID:   user.ID,
		PublicIP: middleware.ExtractIPFromContext(ctx),
		Event:    domain.EventUserLoggedIn,
	}
	if err := s.auditLogRepo.Create(ctx, &auditLog); err != nil {
		logger.L.Errorw("error creating audit log for user login", "user_id", user.ID, "email", req.Email, "error", err)
		return nil, nil, apperrors.FromStatus(http.StatusInternalServerError, "failed to log user login", err)
	}

	// Extract user agent and check if it's mobile
	userAgent := middleware.ExtractUAFromContext(ctx)
	isMobile := false
	if userAgent != nil {
		isMobile = utils.IsMobile(userAgent)
	}

	if isMobile {
		return nil, &MobileLoginResp{
			AccessToken:  accessTokenString,
			RefreshToken: plainRefreshToken,
			ExpiresIn:    30 * 24 * 60 * 60,
		}, nil
	}

	return &BrowserLoginResp{
		AccessToken:  accessTokenString,
		RefreshToken: plainRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    15 * 60,
	}, nil, nil
}

// Refresh generates a new access token and refresh token using the provided refresh token. Revokes the old refresh token.
func (s *authService) Refresh(ctx context.Context, refreshToken string) (*BrowserLoginResp, *MobileLoginResp, *apperrors.APIError) {
	if refreshToken == "" {
		logger.L.Error("missing refresh token in request")
		return nil, nil, apperrors.BadRequest("invalid_body", "missing refresh token", domain.ErrInvalidBodyMissingFields)
	}

	// Use new method to verify argon2 hashed tokens
	rt, err := s.refreshTokenRepo.GetValidTokenForUser(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrRefreshTokenNotFound) {
			return nil, nil, apperrors.Unauthorized("invalid refresh token")
		} else if errors.Is(err, domain.ErrRefreshTokenRevoked) {
			return nil, nil, apperrors.Unauthorized("refresh token has been revoked")
		} else if errors.Is(err, domain.ErrRefreshTokenExpired) {
			return nil, nil, apperrors.Unauthorized("refresh token has expired")
		}
		return nil, nil, apperrors.FromStatus(http.StatusInternalServerError, "failed to retrieve refresh token", err)
	}

	user, err := s.userRepo.GetByID(ctx, rt.UserID)
	if err != nil {
		return nil, nil, apperrors.FromStatus(http.StatusInternalServerError, "failed to retrieve user", err)
	}

	newAccessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role.Name,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(15 * time.Minute).Unix(),
	})

	newAccessTokenString, err := newAccessToken.SignedString([]byte(config.Load().App.JWTSecretKey))
	if err != nil {
		logger.L.Errorw("error signing new access token", "user_id", user.ID, "error", err)
		return nil, nil, apperrors.FromStatus(http.StatusInternalServerError, "failed to sign new access token", err)
	}

	newPlainRefreshToken := randomString(64)
	newTokenHash, err := utils.HashToken(newPlainRefreshToken)
	if err != nil {
		logger.L.Errorw("error hashing new refresh token", "user_id", user.ID, "error", err)
		return nil, nil, apperrors.FromStatus(http.StatusInternalServerError, "failed to create new refresh token", err)
	}

	newRefreshToken := &domain.RefreshToken{
		UserID:    user.ID,
		TokenHash: newTokenHash,
		IssuedAt:  time.Now(),
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	}
	if err := s.refreshTokenRepo.Create(ctx, newRefreshToken); err != nil {
		logger.L.Errorw("error creating new refresh token", "user_id", user.ID, "error", err)
		return nil, nil, apperrors.FromStatus(http.StatusInternalServerError, "failed to create new refresh token", err)
	}

	err = s.refreshTokenRepo.RevokeByHash(ctx, rt.TokenHash)
	if err != nil {
		logger.L.Errorw("error revoking old refresh token", "user_id", user.ID, "error", err)
		return nil, nil, apperrors.FromStatus(http.StatusInternalServerError, "failed to revoke old refresh token", err)
	}

	logger.L.Infow("User refreshed tokens successfully", "user_id", user.ID)

	auditLog := domain.AuditLog{
		UserID:   user.ID,
		PublicIP: middleware.ExtractIPFromContext(ctx),
		Event:    domain.EventUserRefreshedToken,
	}
	if err := s.auditLogRepo.Create(ctx, &auditLog); err != nil {
		logger.L.Errorw("error creating audit log for token refresh", "user_id", user.ID, "error", err)
		return nil, nil, apperrors.FromStatus(http.StatusInternalServerError, "failed to log token refresh", err)
	}

	isMobile := false
	userAgent := middleware.ExtractUAFromContext(ctx)
	if userAgent != nil {
		isMobile = utils.IsMobile(userAgent)
	}

	if isMobile {
		return nil, &MobileLoginResp{
			AccessToken:  newAccessTokenString,
			RefreshToken: newPlainRefreshToken,
			ExpiresIn:    30 * 24 * 60 * 60,
		}, nil
	}
	return &BrowserLoginResp{
		AccessToken:  newAccessTokenString,
		RefreshToken: newPlainRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    15 * 60,
	}, nil, nil
}

// Logout revokes the refresh token
func (s *authService) Logout(ctx context.Context, refreshToken string) *apperrors.APIError {
	if refreshToken == "" {
		logger.L.Error("missing refresh token in logout request")
		return apperrors.BadRequest("invalid_body", "missing refresh token", domain.ErrInvalidBodyMissingFields)
	}

	// Use new method to verify argon2 hashed tokens
	rt, err := s.refreshTokenRepo.GetValidTokenForUser(ctx, refreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrRefreshTokenNotFound) {
			return apperrors.Unauthorized("invalid refresh token")
		} else if errors.Is(err, domain.ErrRefreshTokenRevoked) {
			return apperrors.Unauthorized("refresh token has been revoked")
		} else if errors.Is(err, domain.ErrRefreshTokenExpired) {
			return apperrors.Unauthorized("refresh token has expired")
		}
		return apperrors.FromStatus(http.StatusInternalServerError, "failed to retrieve refresh token", err)
	}

	err = s.refreshTokenRepo.RevokeByHash(ctx, rt.TokenHash)
	if err != nil {
		if errors.Is(err, domain.ErrRefreshTokenNotFound) {
			return apperrors.Unauthorized("invalid refresh token")
		}
		return apperrors.FromStatus(http.StatusInternalServerError, "failed to revoke refresh token", err)
	}

	logger.L.Infow("User logged out successfully", "user_id", rt.UserID)

	auditLog := domain.AuditLog{
		UserID:   rt.UserID,
		PublicIP: middleware.ExtractIPFromContext(ctx),
		Event:    domain.EventUserLoggedOut,
	}
	if err := s.auditLogRepo.Create(ctx, &auditLog); err != nil {
		logger.L.Errorw("error creating audit log for user logout", "error", err)
		return apperrors.FromStatus(http.StatusInternalServerError, "failed to log user logout", err)
	}

	return nil
}

func randomString(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

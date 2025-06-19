package service

import (
	"context"
	"errors"
	"net/http"

	"backend/internal/apperrors"
	"backend/internal/domain"
	"backend/internal/logger"
	"backend/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, *apperrors.APIError)
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
	userRepo repository.UserRepository
}

func NewAuthService(ur repository.UserRepository) AuthService {
	return &authService{userRepo: ur}
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

	// Hash the password using bcrypt
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.L.Errorw("error hashing password", "email", req.Email, "error", err)
		return nil, apperrors.FromStatus(http.StatusInternalServerError, "there was an internal error", errHashFailed)
	}
	user.PasswordHash = string(passwordHash)

	// Create the user
	if err := s.userRepo.Create(ctx, &user); err != nil {
		logger.L.Errorw("error creating user", "email", req.Email, "error", err)
		return nil, apperrors.FromStatus(http.StatusInternalServerError, "there was an internal error", errCreateUserFailed)
	}

	logger.L.Infow("User registration processed", "email", req.Email)

	return &RegisterResponse{
		User:    user,
		Message: "Registration successful. Please verify your email.",
	}, nil
}

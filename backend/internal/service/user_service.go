package service

import (
	"context"
	"errors"

	"backend/internal/apperrors"
	"backend/internal/domain"
	"backend/internal/logger"
	"backend/internal/repository"
)

type UserService interface {
	GetByEmail(ctx context.Context, email string) (*domain.User, *apperrors.APIError)
	Create(ctx context.Context, u *domain.User) *apperrors.APIError
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(r repository.UserRepository) UserService {
	return &userService{repo: r}
}

func (s *userService) GetByEmail(ctx context.Context, email string) (*domain.User, *apperrors.APIError) {
	logger.L.Infow("Getting user by email", "email", email)

	if email == "" {
		err := errors.New("email cannot be empty")
		logger.L.Error(err.Error())
		return nil, apperrors.BadRequest("invalid_email", "email cannot be empty", err)
	}

	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		logger.L.Errorw("Failed to get user by email", "email", email, "error", err)
		return nil, apperrors.FromStatus(500, "failed to get user by email", err)
	}

	if user.ID == "" {
		logger.L.Infow("No user found with email", "email", email)
	} else {
		logger.L.Infow("Successfully retrieved user by email", "id", user.ID, "email", email)
	}
	return user, nil
}

func (s *userService) Create(ctx context.Context, u *domain.User) *apperrors.APIError {
	logger.L.Infow("Creating user", "email", u.Email, "first_name", u.FirstName, "last_name", u.LastName)

	if err := s.repo.Create(ctx, u); err != nil {
		logger.L.Errorw("Failed to create user", "email", u.Email, "error", err)
		return apperrors.FromStatus(500, "failed to create user", err)
	}

	logger.L.Infow("User created successfully", "id", u.ID, "email", u.Email)
	return nil
}

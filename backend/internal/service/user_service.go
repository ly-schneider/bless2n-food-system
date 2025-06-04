package service

import (
	"context"

	"backend/internal/domain"
	"backend/internal/logger"
	"backend/internal/repository"

	"github.com/google/uuid"
)

type UserService interface {
	List(ctx context.Context) ([]domain.User, error)
	Get(ctx context.Context, id uuid.UUID) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	Create(ctx context.Context, u *domain.User) error
	Update(ctx context.Context, id uuid.UUID, in *domain.User) (domain.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(r repository.UserRepository) UserService {
	return &userService{repo: r}
}

func (s *userService) List(ctx context.Context) ([]domain.User, error) {
	logger.L.Info("Listing users")
	users, err := s.repo.List(ctx)
	if err != nil {
		logger.L.Errorw("Failed to list users", "error", err)
		return nil, err
	}
	logger.L.Infow("Successfully listed users", "count", len(users))
	return users, nil
}

func (s *userService) Get(ctx context.Context, id uuid.UUID) (domain.User, error) {
	logger.L.Infow("Getting user", "id", id)
	user, err := s.repo.Get(ctx, id)
	if err != nil {
		logger.L.Errorw("Failed to get user", "id", id, "error", err)
		return user, err
	}
	logger.L.Infow("Successfully retrieved user", "id", id, "email", user.Email)
	return user, nil
}

func (s *userService) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	logger.L.Infow("Getting user by email", "email", email)
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		logger.L.Errorw("Failed to get user by email", "email", email, "error", err)
		return user, err
	}

	if user.ID == uuid.Nil {
		logger.L.Infow("No user found with email", "email", email)
	} else {
		logger.L.Infow("Successfully retrieved user by email", "id", user.ID, "email", email)
	}
	return user, nil
}

func (s *userService) Create(ctx context.Context, u *domain.User) error {
	logger.L.Infow("Creating user", "email", u.Email, "first_name", u.FirstName, "last_name", u.LastName)

	if err := s.repo.Create(ctx, u); err != nil {
		logger.L.Errorw("Failed to create user", "email", u.Email, "error", err)
		return err
	}

	logger.L.Infow("User created successfully", "id", u.ID, "email", u.Email)
	return nil
}

func (s *userService) Update(ctx context.Context, id uuid.UUID, in *domain.User) (domain.User, error) {
	logger.L.Infow("Updating user", "id", id)

	u, err := s.repo.Get(ctx, id)
	if err != nil {
		logger.L.Errorw("User not found for update", "id", id, "error", err)
		return u, err
	}

	u.FirstName = in.FirstName
	u.LastName = in.LastName
	u.Email = in.Email
	u.IsVerified = in.IsVerified
	u.MfaEnabled = in.MfaEnabled
	u.IsDisabled = in.IsDisabled
	u.DisabledReason = in.DisabledReason
	u.RoleID = in.RoleID

	if err := s.repo.Update(ctx, &u); err != nil {
		logger.L.Errorw("Failed to update user", "id", id, "error", err)
		return u, err
	}

	logger.L.Infow("User updated successfully", "id", id, "email", u.Email)
	return u, nil
}

func (s *userService) Delete(ctx context.Context, id uuid.UUID) error {
	logger.L.Infow("Deleting user", "id", id)

	if err := s.repo.Delete(ctx, id); err != nil {
		logger.L.Errorw("Failed to delete user", "id", id, "error", err)
		return err
	}

	logger.L.Infow("User deleted successfully", "id", id)
	return nil
}

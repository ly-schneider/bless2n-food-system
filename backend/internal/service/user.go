package service

import (
	"context"
	"errors"
	"fmt"

	"backend/internal/domain"
	"backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserService defines the user management service interface
type UserService interface {
	UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) (*UpdateProfileResponse, error)
	DeleteProfile(ctx context.Context, userID string) (*DeleteProfileResponse, error)
	GetProfile(ctx context.Context, userID string) (*domain.User, error)
}

// Request/Response types for Profile Update
type UpdateProfileRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type UpdateProfileResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// Request/Response types for Profile Delete
type DeleteProfileResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (s *userService) UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) (*UpdateProfileResponse, error) {
	// Convert userID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// Get current user
	user, err := s.userRepo.GetByID(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Check if user is disabled
	if user.IsDisabled {
		return nil, errors.New("account is disabled")
	}

	// Check if the new email is already in use by another user
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email availability: %w", err)
	}
	if existingUser != nil && existingUser.ID != user.ID {
		return nil, errors.New("email address is already in use")
	}

	// Update user email
	user.Email = req.Email

	// Save updated user
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return &UpdateProfileResponse{
		Message: "Profile updated successfully",
		Success: true,
	}, nil
}

func (s *userService) DeleteProfile(ctx context.Context, userID string) (*DeleteProfileResponse, error) {
	// Convert userID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// Get current user to verify it exists
	user, err := s.userRepo.GetByID(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Ensure only customers can delete their own profiles
	if user.Role != domain.UserRoleCustomer {
		return nil, errors.New("only customer accounts can be self-deleted")
	}

	// Delete user
	if err := s.userRepo.Delete(ctx, objectID); err != nil {
		return nil, fmt.Errorf("failed to delete profile: %w", err)
	}

	return &DeleteProfileResponse{
		Message: "Profile deleted successfully",
		Success: true,
	}, nil
}

func (s *userService) GetProfile(ctx context.Context, userID string) (*domain.User, error) {
	// Convert userID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// Get user
	user, err := s.userRepo.GetByID(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return user, nil
}
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backend/internal/domain"
	"backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AdminService defines the admin management service interface
type AdminService interface {
	ListCustomers(ctx context.Context, limit, offset int) (*ListCustomersResponse, error)
	BanCustomer(ctx context.Context, customerID string, req BanCustomerRequest) (*BanCustomerResponse, error)
	DeleteCustomer(ctx context.Context, customerID string) (*DeleteCustomerResponse, error)
	InviteAdmin(ctx context.Context, adminID string, req InviteAdminRequest) (*InviteAdminResponse, error)
}

// Request/Response types
type Customer struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	IsVerified bool   `json:"is_verified"`
	IsDisabled bool   `json:"is_disabled"`
	CreatedAt  string `json:"created_at"`
}

type ListCustomersResponse struct {
	Customers []Customer `json:"customers"`
	Total     int        `json:"total"`
}

type BanCustomerRequest struct {
	Reason string `json:"reason" validate:"required"`
}

type BanCustomerResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type DeleteCustomerResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type InviteAdminRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type InviteAdminResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type adminService struct {
	userRepo        repository.UserRepository
	adminInviteRepo repository.AdminInviteRepository
	emailService    EmailService
}

func NewAdminService(
	userRepo repository.UserRepository,
	adminInviteRepo repository.AdminInviteRepository,
	emailService EmailService,
) AdminService {
	return &adminService{
		userRepo:        userRepo,
		adminInviteRepo: adminInviteRepo,
		emailService:    emailService,
	}
}

func (s *adminService) ListCustomers(ctx context.Context, limit, offset int) (*ListCustomersResponse, error) {
	// Set default limit if not provided
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100 // Cap at 100 to prevent abuse
	}

	users, err := s.userRepo.ListCustomers(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list customers: %w", err)
	}

	customers := make([]Customer, len(users))
	for i, user := range users {
		customers[i] = Customer{
			ID:         user.ID.Hex(),
			Email:      user.Email,
			IsVerified: user.IsVerified,
			IsDisabled: user.IsDisabled,
			CreatedAt:  user.CreatedAt.Format(time.RFC3339),
		}
	}

	// Get total count for pagination
	totalCount, err := s.userRepo.CountCustomers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count customers: %w", err)
	}

	return &ListCustomersResponse{
		Customers: customers,
		Total:     totalCount,
	}, nil
}

func (s *adminService) BanCustomer(ctx context.Context, customerID string, req BanCustomerRequest) (*BanCustomerResponse, error) {
	// Convert customerID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(customerID)
	if err != nil {
		return nil, errors.New("invalid customer ID format")
	}

	// Get customer to verify it exists and is a customer
	user, err := s.userRepo.GetByID(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}
	if user == nil {
		return nil, errors.New("customer not found")
	}

	// Ensure user is a customer (not admin)
	if user.Role != domain.UserRoleCustomer {
		return nil, errors.New("can only ban customer accounts")
	}

	// Check if already banned
	if user.IsDisabled {
		return nil, errors.New("customer is already banned")
	}

	// Ban the customer
	if err := s.userRepo.Disable(ctx, objectID, req.Reason); err != nil {
		return nil, fmt.Errorf("failed to ban customer: %w", err)
	}

	return &BanCustomerResponse{
		Message: "Customer banned successfully",
		Success: true,
	}, nil
}

func (s *adminService) DeleteCustomer(ctx context.Context, customerID string) (*DeleteCustomerResponse, error) {
	// Convert customerID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(customerID)
	if err != nil {
		return nil, errors.New("invalid customer ID format")
	}

	// Get customer to verify it exists and is a customer
	user, err := s.userRepo.GetByID(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}
	if user == nil {
		return nil, errors.New("customer not found")
	}

	// Ensure user is a customer (not admin)
	if user.Role != domain.UserRoleCustomer {
		return nil, errors.New("can only delete customer accounts")
	}

	// Delete the customer
	if err := s.userRepo.Delete(ctx, objectID); err != nil {
		return nil, fmt.Errorf("failed to delete customer: %w", err)
	}

	return &DeleteCustomerResponse{
		Message: "Customer deleted successfully",
		Success: true,
	}, nil
}

func (s *adminService) InviteAdmin(ctx context.Context, adminID string, req InviteAdminRequest) (*InviteAdminResponse, error) {
	// Convert adminID to ObjectID
	adminObjectID, err := primitive.ObjectIDFromHex(adminID)
	if err != nil {
		return nil, errors.New("invalid admin ID format")
	}

	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, errors.New("user with this email already exists")
	}

	// Check if there's already a pending invite
	existingInvite, err := s.adminInviteRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing invite: %w", err)
	}
	if existingInvite != nil {
		return nil, errors.New("admin invitation already pending for this email")
	}

	// Create admin invite
	invite := &domain.AdminInvite{
		InvitedBy:    adminObjectID,
		InviteeEmail: req.Email,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour), // 7 days expiry
	}

	if err := s.adminInviteRepo.Create(ctx, invite); err != nil {
		return nil, fmt.Errorf("failed to create admin invite: %w", err)
	}

	// Send invitation email (optional, depends on email service implementation)
	emailReq := SendEmailRequest{
		To:      []string{req.Email},
		Subject: "Admin Invitation - Bless2n Food System",
		Body: `
			You have been invited to become an admin for the Bless2n Food System.
			This invitation expires in 7 days.
			
			Please contact your administrator to complete the setup process.
		`,
	}

	if err := s.emailService.SendEmail(ctx, emailReq); err != nil {
		// Log error but don't fail the invitation
		// The invite is created successfully, email failure is not critical
		fmt.Printf("Failed to send invitation email: %v", err)
	}

	return &InviteAdminResponse{
		Message: "Admin invitation sent successfully",
		Success: true,
	}, nil
}

package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"backend/internal/apperrors"
	"backend/internal/domain"
	"backend/internal/http/middleware"
	"backend/internal/http/respond"
	"backend/internal/model"
)

type UserHandler struct {
	vldt *validator.Validate
}

func NewUserHandler() UserHandler {
	return UserHandler{
		vldt: validator.New(),
	}
}

func (h UserHandler) Routes() chi.Router {
	r := chi.NewRouter()

	// Customer endpoints
	r.Put("/profile", h.UpdateProfile)
	r.Delete("/profile", h.DeleteProfile)

	// Admin endpoints
	r.Get("/customers", h.ListCustomers)
	r.Post("/customers/{customerID}/ban", h.BanCustomer)
	r.Delete("/customers/{customerID}", h.DeleteCustomer)
	r.Post("/admin/invite", h.InviteAdmin)

	return r
}

// UpdateProfileRequest represents the request to update customer profile
// @Description Customer profile update request
type UpdateProfileRequest struct {
	Email string `json:"email" validate:"required,email" example:"user@example.com"`
}

// ProfileResponse represents a customer profile
// @Description Customer profile information
type ProfileResponse struct {
	ID        string `json:"id" example:"abcd1234567890"`
	FirstName string `json:"first_name" example:"John"`
	LastName  string `json:"last_name" example:"Doe"`
	Email     string `json:"email" example:"user@example.com"`
}

// CustomerResponse represents customer information for admin views
// @Description Customer information for admin management
type CustomerResponse struct {
	ID             string  `json:"id" example:"abcd1234567890"`
	FirstName      string  `json:"first_name" example:"John"`
	LastName       string  `json:"last_name" example:"Doe"`
	Email          string  `json:"email" example:"user@example.com"`
	IsVerified     bool    `json:"is_verified" example:"true"`
	IsDisabled     bool    `json:"is_disabled" example:"false"`
	DisabledReason *string `json:"disabled_reason,omitempty" example:"Violated terms of service"`
	CreatedAt      string  `json:"created_at" example:"2024-01-15T10:30:00Z"`
}

// ListCustomersResponse represents the response for listing customers
// @Description List of customers with pagination
type ListCustomersResponse struct {
	Customers []CustomerResponse `json:"customers"`
	Total     int               `json:"total" example:"150"`
	Page      int               `json:"page" example:"1"`
	Limit     int               `json:"limit" example:"20"`
}

// BanCustomerRequest represents the request to ban a customer
// @Description Request to ban a customer with reason
type BanCustomerRequest struct {
	Reason string `json:"reason" validate:"required,min=1,max=255" example:"Violated terms of service"`
}

// InviteAdminRequest represents the request to invite a new admin
// @Description Request to invite an email to be an admin
type InviteAdminRequest struct {
	Email string `json:"email" validate:"required,email" example:"admin@example.com"`
}

// UpdateProfile godoc
// @Summary Update customer profile
// @Description Allows customers to update their email address
// @Tags customers
// @Accept json
// @Produce json
// @Param request body UpdateProfileRequest true "Profile update data"
// @Success 200 {object} ProfileResponse "Updated profile"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /v1/users/profile [put]
// @Security BearerAuth
func (h UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.NewWriter(w).WriteError(apperrors.BadRequest("bad_json", domain.ErrParseBody.Error(), err))
		return
	}

	if err := h.vldt.Struct(req); err != nil {
		respond.NewWriter(w).WriteError(apperrors.BadRequest("validation_error", domain.ErrInvalidBody.Error(), err))
		return
	}

	userID := middleware.ExtractUserIDFromContext(r.Context())
	if userID == nil || *userID == "" {
		respond.NewWriter(w).WriteError(apperrors.Unauthorized("User ID is required"))
		return
	}

	// TODO: Implement profile update logic
	response := ProfileResponse{
		ID:        *userID,
		FirstName: "John",
		LastName:  "Doe",
		Email:     req.Email,
	}

	respond.JSON(w, http.StatusOK, response)
}

// DeleteProfile godoc
// @Summary Delete customer profile
// @Description Allows customers to delete their own profile
// @Tags customers
// @Produce json
// @Success 204 "Profile deleted successfully"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /v1/users/profile [delete]
// @Security BearerAuth
func (h UserHandler) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	userID := middleware.ExtractUserIDFromContext(r.Context())
	if userID == nil || *userID == "" {
		respond.NewWriter(w).WriteError(apperrors.Unauthorized("User ID is required"))
		return
	}

	// TODO: Implement profile deletion logic

	respond.JSON(w, http.StatusNoContent, nil)
}

// ListCustomers godoc
// @Summary List all customers
// @Description Allows admins to list all customers with pagination
// @Tags admin
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Success 200 {object} ListCustomersResponse "List of customers"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - Admin access required"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /v1/users/customers [get]
// @Security BearerAuth
func (h UserHandler) ListCustomers(w http.ResponseWriter, r *http.Request) {
	// TODO: Check if user is admin

	// TODO: Implement customer listing with pagination
	customers := []CustomerResponse{
		{
			ID:         "abcd1234567890",
			FirstName:  "John",
			LastName:   "Doe",
			Email:      "john@example.com",
			IsVerified: true,
			IsDisabled: false,
			CreatedAt:  "2024-01-15T10:30:00Z",
		},
		{
			ID:             "efgh0987654321",
			FirstName:      "Jane",
			LastName:       "Smith",
			Email:          "jane@example.com",
			IsVerified:     true,
			IsDisabled:     true,
			DisabledReason: stringPtr("Violated terms of service"),
			CreatedAt:      "2024-01-10T09:15:00Z",
		},
	}

	response := ListCustomersResponse{
		Customers: customers,
		Total:     150,
		Page:      1,
		Limit:     20,
	}

	respond.JSON(w, http.StatusOK, response)
}

// BanCustomer godoc
// @Summary Ban a customer
// @Description Allows admins to ban a customer with a reason
// @Tags admin
// @Accept json
// @Produce json
// @Param customerID path string true "Customer ID"
// @Param request body BanCustomerRequest true "Ban reason"
// @Success 200 {object} map[string]string "Customer banned successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - Admin access required"
// @Failure 404 {object} map[string]string "Customer not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /v1/users/customers/{customerID}/ban [post]
// @Security BearerAuth
func (h UserHandler) BanCustomer(w http.ResponseWriter, r *http.Request) {
	customerID := chi.URLParam(r, "customerID")
	if customerID == "" {
		respond.NewWriter(w).WriteError(apperrors.BadRequest("missing_customer_id", "Customer ID is required", nil))
		return
	}

	var req BanCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.NewWriter(w).WriteError(apperrors.BadRequest("bad_json", domain.ErrParseBody.Error(), err))
		return
	}

	if err := h.vldt.Struct(req); err != nil {
		respond.NewWriter(w).WriteError(apperrors.BadRequest("validation_error", domain.ErrInvalidBody.Error(), err))
		return
	}

	// TODO: Check if user is admin
	// TODO: Implement customer banning logic

	respond.JSON(w, http.StatusOK, map[string]string{
		"message": "Customer banned successfully",
	})
}

// DeleteCustomer godoc
// @Summary Delete a customer
// @Description Allows admins to delete a customer account
// @Tags admin
// @Produce json
// @Param customerID path string true "Customer ID"
// @Success 204 "Customer deleted successfully"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - Admin access required"
// @Failure 404 {object} map[string]string "Customer not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /v1/users/customers/{customerID} [delete]
// @Security BearerAuth
func (h UserHandler) DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	customerID := chi.URLParam(r, "customerID")
	if customerID == "" {
		respond.NewWriter(w).WriteError(apperrors.BadRequest("missing_customer_id", "Customer ID is required", nil))
		return
	}

	// TODO: Check if user is admin
	// TODO: Implement customer deletion logic

	respond.JSON(w, http.StatusNoContent, nil)
}

// InviteAdmin godoc
// @Summary Invite new admin
// @Description Allows admins to invite an email address to become an admin
// @Tags admin
// @Accept json
// @Produce json
// @Param request body InviteAdminRequest true "Admin invitation data"
// @Success 201 {object} map[string]string "Admin invitation sent successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - Admin access required"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /v1/users/admin/invite [post]
// @Security BearerAuth
func (h UserHandler) InviteAdmin(w http.ResponseWriter, r *http.Request) {
	var req InviteAdminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.NewWriter(w).WriteError(apperrors.BadRequest("bad_json", domain.ErrParseBody.Error(), err))
		return
	}

	if err := h.vldt.Struct(req); err != nil {
		respond.NewWriter(w).WriteError(apperrors.BadRequest("validation_error", domain.ErrInvalidBody.Error(), err))
		return
	}

	// TODO: Check if user is admin
	// TODO: Implement admin invitation logic

	respond.JSON(w, http.StatusCreated, map[string]string{
		"message": "Admin invitation sent successfully",
	})
}

func stringPtr(s string) *string {
	return &s
}
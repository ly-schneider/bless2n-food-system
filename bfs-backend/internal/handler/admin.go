package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"backend/internal/middleware"
	"backend/internal/response"
	"backend/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type AdminHandler struct {
	adminService service.AdminService
	validator    *validator.Validate
}

func NewAdminHandler(adminService service.AdminService) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
		validator:    validator.New(),
	}
}

// ListCustomers godoc
// @Summary List all customers
// @Description Allow admins to list all customers
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit size" minimum(1) maximum(100) default(50)
// @Param offset query int false "Offset" minimum(0) default(0)
// @Success 200 {object} service.ListCustomersResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Router /v1/admin/customers [get]
func (h *AdminHandler) ListCustomers(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // default
	offset := 0 // default

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	svcResp, err := h.adminService.ListCustomers(r.Context(), limit, offset)
	if err != nil {
		zap.L().Error("failed to list customers", zap.Error(err))
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.WriteJSON(w, http.StatusOK, svcResp)
}

// BanCustomer godoc
// @Summary Ban a customer
// @Description Allow admins to ban a customer with a reason
// @Tags admin
// @Accept json
// @Produce json
// @Security Bearer
// @Param customer_id path string true "Customer ID"
// @Param request body service.BanCustomerRequest true "Ban reason payload"
// @Success 200 {object} service.BanCustomerResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /v1/admin/customers/{customer_id}/ban [post]
func (h *AdminHandler) BanCustomer(w http.ResponseWriter, r *http.Request) {
	// Get admin from JWT token
	_, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	customerID := chi.URLParam(r, "customer_id")
	if customerID == "" {
		response.WriteError(w, http.StatusBadRequest, "Customer ID is required")
		return
	}

	var req service.BanCustomerRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		zap.L().Error("failed to decode request body", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		zap.L().Error("validation failed", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	svcResp, err := h.adminService.BanCustomer(r.Context(), customerID, req)
	if err != nil {
		zap.L().Error("failed to ban customer", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.WriteJSON(w, http.StatusOK, svcResp)
}

// DeleteCustomer godoc
// @Summary Delete a customer
// @Description Allow admins to delete a customer
// @Tags admin
// @Produce json
// @Security Bearer
// @Param customer_id path string true "Customer ID"
// @Success 200 {object} service.DeleteCustomerResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /v1/admin/customers/{customer_id} [delete]
func (h *AdminHandler) DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	// Get admin from JWT token
	_, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	customerID := chi.URLParam(r, "customer_id")
	if customerID == "" {
		response.WriteError(w, http.StatusBadRequest, "Customer ID is required")
		return
	}

	svcResp, err := h.adminService.DeleteCustomer(r.Context(), customerID)
	if err != nil {
		zap.L().Error("failed to delete customer", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.WriteJSON(w, http.StatusOK, svcResp)
}

// InviteAdmin godoc
// @Summary Invite admin
// @Description Allow admins to invite emails to become admins
// @Tags admin
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body service.InviteAdminRequest true "Admin invite payload"
// @Success 201 {object} service.InviteAdminResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Router /v1/admin/invites [post]
func (h *AdminHandler) InviteAdmin(w http.ResponseWriter, r *http.Request) {
	// Get admin from JWT token
	admin, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req service.InviteAdminRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		zap.L().Error("failed to decode request body", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		zap.L().Error("validation failed", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	svcResp, err := h.adminService.InviteAdmin(r.Context(), admin.Subject, req)
	if err != nil {
		zap.L().Error("failed to invite admin", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.WriteJSON(w, http.StatusCreated, svcResp)
}

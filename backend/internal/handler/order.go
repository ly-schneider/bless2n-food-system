package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"backend/internal/domain"
	"backend/internal/middleware"
	"backend/internal/response"
	"backend/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type OrderHandler struct {
	orderService service.OrderService
	validator    *validator.Validate
	logger       *zap.Logger
}

func NewOrderHandler(orderService service.OrderService, logger *zap.Logger) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
		validator:    validator.New(),
		logger:       logger,
	}
}

// CreateOrder godoc
// @Summary Create a new order
// @Description Create a new order with items, either for authenticated users or with contact email for POS orders
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.CreateOrderRequest true "Order creation payload"
// @Success 201 {object} service.CreateOrderResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /v1/orders [post]
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req service.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		h.logger.Error("Validation failed", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	// If customer_id is not provided in request, try to get it from context (authenticated user)
	if req.CustomerID == nil {
		if userClaim, ok := middleware.GetUserFromContext(r.Context()); ok {
			req.CustomerID = &userClaim.Subject
		}
	}

	svcResp, err := h.orderService.CreateOrder(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to create order", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.WriteJSON(w, http.StatusCreated, svcResp)
}

// GetOrder godoc
// @Summary Get order by ID
// @Description Retrieve detailed information about a specific order including items
// @Tags orders
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} service.GetOrderResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /v1/orders/{id} [get]
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")
	if orderID == "" {
		response.WriteError(w, http.StatusBadRequest, "Order ID is required")
		return
	}

	svcResp, err := h.orderService.GetOrder(r.Context(), orderID)
	if err != nil {
		h.logger.Error("Failed to get order", zap.Error(err))
		if err.Error() == "order not found" {
			response.WriteError(w, http.StatusNotFound, err.Error())
		} else {
			response.WriteError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	response.WriteJSON(w, http.StatusOK, svcResp)
}

// UpdateOrder godoc
// @Summary Update order
// @Description Update order details (limited to pending orders)
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Param request body service.UpdateOrderRequest true "Order update payload"
// @Success 200 {object} service.UpdateOrderResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /v1/orders/{id} [put]
func (h *OrderHandler) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")
	if orderID == "" {
		response.WriteError(w, http.StatusBadRequest, "Order ID is required")
		return
	}

	var req service.UpdateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		h.logger.Error("Validation failed", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	svcResp, err := h.orderService.UpdateOrder(r.Context(), orderID, req)
	if err != nil {
		h.logger.Error("Failed to update order", zap.Error(err))
		if err.Error() == "order not found" {
			response.WriteError(w, http.StatusNotFound, err.Error())
		} else {
			response.WriteError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	response.WriteJSON(w, http.StatusOK, svcResp)
}

// DeleteOrder godoc
// @Summary Delete order
// @Description Delete an order (only pending or cancelled orders)
// @Tags orders
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Success 200 {object} service.DeleteOrderResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /v1/orders/{id} [delete]
func (h *OrderHandler) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	orderID := chi.URLParam(r, "id")
	if orderID == "" {
		response.WriteError(w, http.StatusBadRequest, "Order ID is required")
		return
	}

	svcResp, err := h.orderService.DeleteOrder(r.Context(), orderID)
	if err != nil {
		h.logger.Error("Failed to delete order", zap.Error(err))
		if err.Error() == "order not found" {
			response.WriteError(w, http.StatusNotFound, err.Error())
		} else {
			response.WriteError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	response.WriteJSON(w, http.StatusOK, svcResp)
}

// ListOrders godoc
// @Summary List orders
// @Description List orders with optional filtering by status
// @Tags orders
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status" Enums(pending, paid, cancelled, refunded)
// @Param limit query int false "Limit size" minimum(1) maximum(100) default(50)
// @Param offset query int false "Offset" minimum(0) default(0)
// @Success 200 {object} service.ListOrdersResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /v1/orders [get]
func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	statusStr := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	req := service.ListOrdersRequest{}

	if statusStr != "" {
		status := domain.OrderStatus(statusStr)
		// Validate status
		validStatuses := []domain.OrderStatus{
			domain.OrderStatusPending,
			domain.OrderStatusPaid,
			domain.OrderStatusCancelled,
			domain.OrderStatusRefunded,
		}
		isValid := false
		for _, validStatus := range validStatuses {
			if status == validStatus {
				isValid = true
				break
			}
		}
		if !isValid {
			response.WriteError(w, http.StatusBadRequest, "Invalid status value")
			return
		}
		req.Status = &status
	}

	if limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			req.Limit = limit
		}
	}

	if offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			req.Offset = offset
		}
	}

	svcResp, err := h.orderService.ListOrders(r.Context(), req)
	if err != nil {
		h.logger.Error("Failed to list orders", zap.Error(err))
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.WriteJSON(w, http.StatusOK, svcResp)
}

// UpdateOrderStatus godoc
// @Summary Update order status
// @Description Update the status of an order (admin only)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Order ID"
// @Param status query string true "New status" Enums(pending, paid, cancelled, refunded)
// @Success 200 {object} service.UpdateOrderStatusResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /v1/admin/orders/{id}/status [put]
func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	userClaim, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.logger.Error("User claims not found in context")
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if string(userClaim.Role) != "admin" {
		response.WriteError(w, http.StatusForbidden, "Access denied")
		return
	}

	orderID := chi.URLParam(r, "id")
	if orderID == "" {
		response.WriteError(w, http.StatusBadRequest, "Order ID is required")
		return
	}

	statusStr := r.URL.Query().Get("status")
	if statusStr == "" {
		response.WriteError(w, http.StatusBadRequest, "Status is required")
		return
	}

	status := domain.OrderStatus(statusStr)
	// Validate status
	validStatuses := []domain.OrderStatus{
		domain.OrderStatusPending,
		domain.OrderStatusPaid,
		domain.OrderStatusCancelled,
		domain.OrderStatusRefunded,
	}
	isValid := false
	for _, validStatus := range validStatuses {
		if status == validStatus {
			isValid = true
			break
		}
	}
	if !isValid {
		response.WriteError(w, http.StatusBadRequest, "Invalid status value")
		return
	}

	svcResp, err := h.orderService.UpdateOrderStatus(r.Context(), orderID, status)
	if err != nil {
		h.logger.Error("Failed to update order status", zap.Error(err))
		if err.Error() == "order not found" {
			response.WriteError(w, http.StatusNotFound, err.Error())
		} else {
			response.WriteError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	response.WriteJSON(w, http.StatusOK, svcResp)
}

// GetMyOrders godoc
// @Summary Get current user's orders
// @Description Retrieve orders for the authenticated user
// @Tags orders
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit size" minimum(1) maximum(100) default(50)
// @Param offset query int false "Offset" minimum(0) default(0)
// @Success 200 {object} service.ListOrdersResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /v1/orders/my [get]
func (h *OrderHandler) GetMyOrders(w http.ResponseWriter, r *http.Request) {
	userClaim, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		h.logger.Error("User claims not found in context")
		response.WriteError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	offset := 0

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

	customerID := userClaim.Subject
	svcResp, err := h.orderService.GetOrdersByCustomer(r.Context(), customerID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get user orders", zap.Error(err))
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.WriteJSON(w, http.StatusOK, svcResp)
}


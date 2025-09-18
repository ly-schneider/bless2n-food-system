package handler

import (
    "backend/internal/domain"
    "backend/internal/service"
    _ "backend/internal/response" // referenced in swagger annotations
    "encoding/json"
    "net/http"

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
// @Description Create a new order with the provided details
// @Tags orders
// @Accept json
// @Produce json
// @Success 200 {object} response.Ack
// @Failure 400 {object} response.ProblemDetails
// @Failure 401 {object} response.ProblemDetails
// @Failure 403 {object} response.ProblemDetails
// @Router /v1/orders [post]
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var createOrderDTO domain.CreateOrderDTO
	if err := json.NewDecoder(r.Body).Decode(&createOrderDTO); err != nil {
		h.logger.Error("failed to decode create order request", zap.Error(err))
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := h.validator.Struct(&createOrderDTO); err != nil {
		h.logger.Error("validation failed for create order request", zap.Error(err))
		http.Error(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	ack, err := h.orderService.CreateOrder(r.Context(), &createOrderDTO)
	if err != nil {
		h.logger.Error("failed to create order", zap.Error(err))
		http.Error(w, "Failed to create order: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(ack); err != nil {
		// Log the encoding error and return a 500
		h.logger.Error("failed to encode response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

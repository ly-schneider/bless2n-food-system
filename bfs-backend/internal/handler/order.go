package handler

import (
	"backend/internal/service"

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

package handler

import (
	"backend/internal/service"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type RedemptionHandler struct {
	orderService service.OrderService
	validator    *validator.Validate
	logger       *zap.Logger
}

func NewRedemptionHandler(orderService service.OrderService, logger *zap.Logger) *RedemptionHandler {
	return &RedemptionHandler{
		orderService: orderService,
		validator:    validator.New(),
		logger:       logger,
	}
}

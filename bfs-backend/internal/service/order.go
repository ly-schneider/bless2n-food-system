package service

import (
	"backend/internal/domain"
	"backend/internal/response"
	"context"
)

type OrderService interface {
	CreateOrder(ctx context.Context, createOrderDTO *domain.CreateOrderDTO) (*response.Ack, error)
}

type orderService struct{}

func NewOrderService() OrderService {
	return &orderService{}
}

func (s *orderService) CreateOrder(ctx context.Context, createOrderDTO *domain.CreateOrderDTO) (*response.Ack, error) {
	// TODO: implement creation logic via repository and validations
	return &response.Ack{Message: "Order created!"}, nil
}

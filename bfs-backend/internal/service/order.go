package service

type OrderService interface {
}

type orderService struct {
}

func NewOrderService() OrderService {
	return &orderService{}
}

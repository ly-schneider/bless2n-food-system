package service

import (
	"backend/internal/config"
	"backend/internal/domain"
	"backend/internal/repository"
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type POSService interface {
	// Device + access
	RequestAccess(ctx context.Context, name, model, os, token string) error
	GetDeviceByToken(ctx context.Context, token string) (*domain.PosDevice, error)
	// Orders
	CreateOrder(ctx context.Context, items []CreateCheckoutInputItem, customerEmail *string) (bson.ObjectID, error)
	PayCash(ctx context.Context, orderID bson.ObjectID, amountReceived domain.Cents) (change domain.Cents, err error)
	PayCard(ctx context.Context, orderID bson.ObjectID, processor string, transactionID *string, status string) error
	PayTwint(ctx context.Context, orderID bson.ObjectID, transactionID *string, status string) error
}

// Reuse CreateCheckoutInput item subtype
type CreateCheckoutInputItem = struct {
	ProductID     string            `json:"productId"`
	Quantity      int64             `json:"quantity"`
	Configuration map[string]string `json:"configuration,omitempty"`
}

type posService struct {
	cfg      config.Config
	devices  repository.PosDeviceRepository
	orders   repository.OrderRepository
	payments PaymentService
}

func NewPOSService(
	cfg config.Config,
	devices repository.PosDeviceRepository,
	orders repository.OrderRepository,
	payments PaymentService,
) POSService {
	return &posService{cfg: cfg, devices: devices, orders: orders, payments: payments}
}

func (s *posService) RequestAccess(ctx context.Context, name, model, os, token string) error {
	if name == "" || token == "" {
		return errors.New("invalid_payload")
	}
	_, err := s.devices.UpsertPendingByToken(ctx, name, model, os, token)
	return err
}

func (s *posService) GetDeviceByToken(ctx context.Context, token string) (*domain.PosDevice, error) {
	return s.devices.FindByToken(ctx, token)
}

func (s *posService) CreateOrder(ctx context.Context, items []CreateCheckoutInputItem, customerEmail *string) (bson.ObjectID, error) {
	if len(items) == 0 {
		return bson.NilObjectID, fmt.Errorf("no items")
	}
	in := CreateCheckoutInput{Items: items, CustomerEmail: customerEmail}
	prep, err := s.payments.PrepareAndCreateOrder(ctx, in, nil, nil)
	if err != nil {
		return bson.NilObjectID, err
	}
	// Mark origin as POS
	_ = s.orders.SetOrigin(ctx, prep.OrderID, domain.OrderOriginPOS)
	return prep.OrderID, nil
}

func (s *posService) PayCash(ctx context.Context, orderID bson.ObjectID, amountReceived domain.Cents) (domain.Cents, error) {
	if orderID.IsZero() {
		return 0, fmt.Errorf("invalid order id")
	}
	ord, err := s.orders.FindByID(ctx, orderID)
	if err != nil {
		return 0, err
	}
	if ord.Status != domain.OrderStatusPending {
		return 0, fmt.Errorf("not_pending")
	}
	if amountReceived < ord.TotalCents {
		return 0, fmt.Errorf("insufficient_cash")
	}
	change := amountReceived - ord.TotalCents
	if err := s.orders.SetPosPaymentCash(ctx, orderID, amountReceived, change); err != nil {
		return 0, err
	}
	return change, nil
}

func (s *posService) PayCard(ctx context.Context, orderID bson.ObjectID, processor string, transactionID *string, status string) error {
	if orderID.IsZero() {
		return fmt.Errorf("invalid order id")
	}
	markPaid := status == "succeeded"
	return s.orders.SetPosPaymentCard(ctx, orderID, processor, transactionID, status, markPaid)
}

func (s *posService) PayTwint(ctx context.Context, orderID bson.ObjectID, transactionID *string, status string) error {
	if orderID.IsZero() {
		return fmt.Errorf("invalid order id")
	}
	markPaid := status == "succeeded"
	return s.orders.SetPosPaymentTwint(ctx, orderID, transactionID, status, markPaid)
}

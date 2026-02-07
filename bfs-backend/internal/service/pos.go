package service

import (
	"context"
	"errors"
	"fmt"

	"backend/internal/config"
	"backend/internal/generated/ent"
	"backend/internal/generated/ent/device"
	"backend/internal/generated/ent/order"
	"backend/internal/repository"

	"github.com/google/uuid"
)

type POSService interface {
	// Device
	GetDeviceByToken(ctx context.Context, token string) (*ent.Device, error)
	GetDeviceByID(ctx context.Context, id uuid.UUID) (*ent.Device, error)
	// Orders
	CreateOrder(ctx context.Context, items []POSCheckoutItem, customerEmail *string) (uuid.UUID, error)
	PayCash(ctx context.Context, orderID uuid.UUID, deviceID *uuid.UUID) error
	PayCard(ctx context.Context, orderID uuid.UUID, deviceID *uuid.UUID) error
	PayTwint(ctx context.Context, orderID uuid.UUID, deviceID *uuid.UUID) error
	PayGratisGuest(ctx context.Context, orderID uuid.UUID, deviceID *uuid.UUID) error
	PayGratisVIP(ctx context.Context, orderID uuid.UUID, deviceID *uuid.UUID) error
	PayGratisStaff(ctx context.Context, orderID uuid.UUID, deviceID *uuid.UUID) error
	PayGratis100Club(ctx context.Context, orderID uuid.UUID, deviceID *uuid.UUID, elvantoPersonID, elvantoPersonName string, freeQty int) error
}

type POSCheckoutItem struct {
	ProductID     string            `json:"productId"`
	Quantity      int               `json:"quantity"`
	Configuration map[string]string `json:"configuration,omitempty"`
}

type posService struct {
	cfg      config.Config
	devices  repository.DeviceRepository
	orders   repository.OrderRepository
	payments PaymentService
	club100  Club100Service
}

func NewPOSService(
	cfg config.Config,
	devices repository.DeviceRepository,
	orders repository.OrderRepository,
	payments PaymentService,
	club100 Club100Service,
) POSService {
	return &posService{
		cfg:      cfg,
		devices:  devices,
		orders:   orders,
		payments: payments,
		club100:  club100,
	}
}

func (s *posService) GetDeviceByToken(ctx context.Context, token string) (*ent.Device, error) {
	d, err := s.devices.GetByDeviceKey(ctx, token)
	if err != nil {
		return nil, err
	}
	if d.Type != device.TypePOS {
		return nil, errors.New("device_not_pos")
	}
	return d, nil
}

func (s *posService) GetDeviceByID(ctx context.Context, id uuid.UUID) (*ent.Device, error) {
	d, err := s.devices.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if d.Type != device.TypePOS {
		return nil, errors.New("device_not_pos")
	}
	return d, nil
}

func (s *posService) CreateOrder(ctx context.Context, items []POSCheckoutItem, customerEmail *string) (uuid.UUID, error) {
	if len(items) == 0 {
		return uuid.Nil, fmt.Errorf("no items")
	}

	// Convert to payment service input
	checkoutItems := make([]CheckoutItemInput, len(items))
	for i, item := range items {
		checkoutItems[i] = CheckoutItemInput{
			ProductID:     item.ProductID,
			Quantity:      item.Quantity,
			Configuration: item.Configuration,
		}
	}

	in := CreateCheckoutInput{
		Items:         checkoutItems,
		CustomerEmail: customerEmail,
		Origin:        order.OriginPos,
	}

	prep, err := s.payments.PrepareAndCreateOrder(ctx, in, nil, nil)
	if err != nil {
		return uuid.Nil, err
	}

	return prep.OrderID, nil
}

func (s *posService) PayCash(ctx context.Context, orderID uuid.UUID, deviceID *uuid.UUID) error {
	if orderID == uuid.Nil {
		return fmt.Errorf("invalid order id")
	}

	ord, err := s.orders.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if ord.Status != order.StatusPending {
		return fmt.Errorf("not_pending")
	}

	return s.orders.SetPosPaymentCash(ctx, orderID, deviceID, ord.TotalCents)
}

func (s *posService) PayCard(ctx context.Context, orderID uuid.UUID, deviceID *uuid.UUID) error {
	if orderID == uuid.Nil {
		return fmt.Errorf("invalid order id")
	}

	ord, err := s.orders.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if ord.Status != order.StatusPending {
		return fmt.Errorf("not_pending")
	}

	return s.orders.SetPosPaymentCard(ctx, orderID, deviceID, ord.TotalCents)
}

func (s *posService) PayTwint(ctx context.Context, orderID uuid.UUID, deviceID *uuid.UUID) error {
	if orderID == uuid.Nil {
		return fmt.Errorf("invalid order id")
	}

	ord, err := s.orders.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if ord.Status != order.StatusPending {
		return fmt.Errorf("not_pending")
	}

	return s.orders.SetPosPaymentTwint(ctx, orderID, deviceID, ord.TotalCents)
}

func (s *posService) PayGratisGuest(ctx context.Context, orderID uuid.UUID, deviceID *uuid.UUID) error {
	if orderID == uuid.Nil {
		return fmt.Errorf("invalid order id")
	}

	ord, err := s.orders.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if ord.Status != order.StatusPending {
		return fmt.Errorf("not_pending")
	}

	return s.orders.SetPosPaymentGratisGuest(ctx, orderID, deviceID, ord.TotalCents)
}

func (s *posService) PayGratisVIP(ctx context.Context, orderID uuid.UUID, deviceID *uuid.UUID) error {
	if orderID == uuid.Nil {
		return fmt.Errorf("invalid order id")
	}

	ord, err := s.orders.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if ord.Status != order.StatusPending {
		return fmt.Errorf("not_pending")
	}

	return s.orders.SetPosPaymentGratisVIP(ctx, orderID, deviceID, ord.TotalCents)
}

func (s *posService) PayGratisStaff(ctx context.Context, orderID uuid.UUID, deviceID *uuid.UUID) error {
	if orderID == uuid.Nil {
		return fmt.Errorf("invalid order id")
	}

	ord, err := s.orders.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if ord.Status != order.StatusPending {
		return fmt.Errorf("not_pending")
	}

	return s.orders.SetPosPaymentGratisStaff(ctx, orderID, deviceID, ord.TotalCents)
}

func (s *posService) PayGratis100Club(ctx context.Context, orderID uuid.UUID, deviceID *uuid.UUID, elvantoPersonID, elvantoPersonName string, freeQty int) error {
	if orderID == uuid.Nil {
		return fmt.Errorf("invalid order id")
	}

	ord, err := s.orders.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if ord.Status != order.StatusPending {
		return fmt.Errorf("not_pending")
	}

	if err := s.club100.ValidateOrderForRedemption(ctx, orderID); err != nil {
		return err
	}

	if err := s.club100.RecordRedemption(ctx, elvantoPersonID, elvantoPersonName, orderID, freeQty); err != nil {
		return fmt.Errorf("record redemption: %w", err)
	}

	return s.orders.SetPosPaymentGratis100Club(ctx, orderID, deviceID, ord.TotalCents)
}

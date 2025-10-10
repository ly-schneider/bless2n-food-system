package service

import (
    "backend/internal/config"
    "backend/internal/domain"
    "backend/internal/repository"
    "context"
    "errors"
    "fmt"
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

type POSService interface {
    // Device + access
    RequestAccess(ctx context.Context, name, model, os, token string) error
    GetDeviceByToken(ctx context.Context, token string) (*domain.PosDevice, error)
    // Orders
    CreateOrder(ctx context.Context, items []CreateCheckoutInputItem, customerEmail *string) (primitive.ObjectID, error)
    PayCash(ctx context.Context, orderID primitive.ObjectID, amountReceived domain.Cents) (change domain.Cents, err error)
    PayCard(ctx context.Context, orderID primitive.ObjectID, processor string, transactionID *string, status string) error
}

// Reuse CreateCheckoutInput item subtype
type CreateCheckoutInputItem = struct {
    ProductID string             `json:"productId"`
    Quantity  int64              `json:"quantity"`
    Configuration map[string]string `json:"configuration,omitempty"`
}

type posService struct {
    cfg          config.Config
    devices      repository.PosDeviceRepository
    requests     repository.PosRequestRepository
    orders       repository.OrderRepository
    payments     PaymentService
}

func NewPOSService(
    cfg config.Config,
    devices repository.PosDeviceRepository,
    requests repository.PosRequestRepository,
    orders repository.OrderRepository,
    payments PaymentService,
) POSService {
    return &posService{ cfg: cfg, devices: devices, requests: requests, orders: orders, payments: payments }
}

func (s *posService) RequestAccess(ctx context.Context, name, model, os, token string) error {
    if name == "" || token == "" { return errors.New("invalid_payload") }
    // If there is already a pending request for this token, do nothing (idempotent)
    if _, err := s.requests.FindPendingByToken(ctx, token); err == nil {
        return nil
    }
    return s.requests.Create(ctx, &domain.PosRequest{
        Name:        name,
        Model:       model,
        OS:          os,
        DeviceToken: token,
        Status:      domain.PosRequestStatusPending,
        ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
    })
}

func (s *posService) GetDeviceByToken(ctx context.Context, token string) (*domain.PosDevice, error) {
    return s.devices.FindByToken(ctx, token)
}

func (s *posService) CreateOrder(ctx context.Context, items []CreateCheckoutInputItem, customerEmail *string) (primitive.ObjectID, error) {
    if len(items) == 0 { return primitive.NilObjectID, fmt.Errorf("no items") }
    in := CreateCheckoutInput{ Items: items, CustomerEmail: customerEmail }
    prep, err := s.payments.PrepareAndCreateOrder(ctx, in, nil, nil)
    if err != nil { return primitive.NilObjectID, err }
    // Mark origin as POS
    _ = s.orders.SetOrigin(ctx, prep.OrderID, domain.OrderOriginPOS)
    return prep.OrderID, nil
}

func (s *posService) PayCash(ctx context.Context, orderID primitive.ObjectID, amountReceived domain.Cents) (domain.Cents, error) {
    if orderID.IsZero() { return 0, fmt.Errorf("invalid order id") }
    ord, err := s.orders.FindByID(ctx, orderID)
    if err != nil { return 0, err }
    if ord.Status != domain.OrderStatusPending { return 0, fmt.Errorf("not_pending") }
    if amountReceived < ord.TotalCents { return 0, fmt.Errorf("insufficient_cash") }
    change := amountReceived - ord.TotalCents
    if err := s.orders.SetPosPaymentCash(ctx, orderID, amountReceived, change); err != nil { return 0, err }
    return change, nil
}

func (s *posService) PayCard(ctx context.Context, orderID primitive.ObjectID, processor string, transactionID *string, status string) error {
    if orderID.IsZero() { return fmt.Errorf("invalid order id") }
    markPaid := status == "succeeded"
    return s.orders.SetPosPaymentCard(ctx, orderID, processor, transactionID, status, markPaid)
}

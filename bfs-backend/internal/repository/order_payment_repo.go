package repository

import (
	"context"
	"time"

	"backend/internal/generated/ent"
	"backend/internal/generated/ent/orderpayment"

	"github.com/google/uuid"
)

type OrderPaymentRepository interface {
	Create(ctx context.Context, orderID uuid.UUID, method orderpayment.Method, amountCents int64, paidAt time.Time, deviceID *uuid.UUID) (*ent.OrderPayment, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.OrderPayment, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*ent.OrderPayment, error)
	Update(ctx context.Context, id, orderID uuid.UUID, method orderpayment.Method, amountCents int64, paidAt time.Time, deviceID *uuid.UUID) (*ent.OrderPayment, error)
}

type orderPaymentRepo struct {
	client *ent.Client
}

func NewOrderPaymentRepository(client *ent.Client) OrderPaymentRepository {
	return &orderPaymentRepo{client: client}
}

func (r *orderPaymentRepo) ec(ctx context.Context) *ent.Client {
	return ClientFromContext(ctx, r.client)
}

func (r *orderPaymentRepo) Create(ctx context.Context, orderID uuid.UUID, method orderpayment.Method, amountCents int64, paidAt time.Time, deviceID *uuid.UUID) (*ent.OrderPayment, error) {
	builder := r.ec(ctx).OrderPayment.Create().
		SetOrderID(orderID).
		SetMethod(method).
		SetAmountCents(amountCents).
		SetPaidAt(paidAt)
	if deviceID != nil {
		builder.SetDeviceID(*deviceID)
	}
	created, err := builder.Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return created, nil
}

func (r *orderPaymentRepo) GetByID(ctx context.Context, id uuid.UUID) (*ent.OrderPayment, error) {
	e, err := r.ec(ctx).OrderPayment.Get(ctx, id)
	if err != nil {
		return nil, translateError(err)
	}
	return e, nil
}

func (r *orderPaymentRepo) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*ent.OrderPayment, error) {
	rows, err := r.ec(ctx).OrderPayment.Query().
		Where(orderpayment.OrderIDEQ(orderID)).
		Order(orderpayment.ByPaidAt()).
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

func (r *orderPaymentRepo) Update(ctx context.Context, id, orderID uuid.UUID, method orderpayment.Method, amountCents int64, paidAt time.Time, deviceID *uuid.UUID) (*ent.OrderPayment, error) {
	builder := r.ec(ctx).OrderPayment.UpdateOneID(id).
		SetOrderID(orderID).
		SetMethod(method).
		SetAmountCents(amountCents).
		SetPaidAt(paidAt)
	if deviceID != nil {
		builder.SetDeviceID(*deviceID)
	} else {
		builder.ClearDeviceID()
	}
	updated, err := builder.Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return updated, nil
}

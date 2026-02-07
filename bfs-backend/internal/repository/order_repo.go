package repository

import (
	"context"
	"time"

	"backend/internal/generated/ent"
	"backend/internal/generated/ent/order"
	"backend/internal/generated/ent/orderpayment"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
)

type OrderRepository interface {
	Create(ctx context.Context, totalCents int64, status order.Status, origin order.Origin, customerID, contactEmail, paymentAttemptID *string, payrexxGatewayID, payrexxTransactionID *int) (*ent.Order, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.Order, error)
	GetByIDWithRelations(ctx context.Context, id uuid.UUID) (*ent.Order, error)
	GetByCustomerID(ctx context.Context, customerID string) ([]*ent.Order, error)
	GetByStatus(ctx context.Context, status order.Status) ([]*ent.Order, error)
	GetByDateRange(ctx context.Context, start, end time.Time) ([]*ent.Order, error)
	GetRecent(ctx context.Context, limit int) ([]*ent.Order, error)
	Update(ctx context.Context, id uuid.UUID, totalCents int64, status order.Status, origin order.Origin, customerID, contactEmail, paymentAttemptID *string, payrexxGatewayID, payrexxTransactionID *int) (*ent.Order, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status order.Status) error

	// Admin and customer listing with pagination
	ListAdmin(ctx context.Context, status *order.Status, from, to *time.Time, q *string, limit, offset int) ([]*ent.Order, int64, error)
	ListByCustomerIDPaginated(ctx context.Context, customerID string, limit, offset int) ([]*ent.Order, int64, error)

	// POS payment methods - creates payment record and updates order status
	SetPosPaymentCash(ctx context.Context, orderID uuid.UUID, deviceID *uuid.UUID, amountCents int64) error
	SetPosPaymentCard(ctx context.Context, orderID uuid.UUID, deviceID *uuid.UUID, amountCents int64) error
	SetPosPaymentTwint(ctx context.Context, orderID uuid.UUID, deviceID *uuid.UUID, amountCents int64) error
	SetPosPaymentGratisGuest(ctx context.Context, orderID uuid.UUID, deviceID *uuid.UUID, amountCents int64) error
	SetPosPaymentGratisVIP(ctx context.Context, orderID uuid.UUID, deviceID *uuid.UUID, amountCents int64) error
	SetPosPaymentGratisStaff(ctx context.Context, orderID uuid.UUID, deviceID *uuid.UUID, amountCents int64) error
	SetPosPaymentGratis100Club(ctx context.Context, orderID uuid.UUID, deviceID *uuid.UUID, amountCents int64) error

	// Additional methods
	DeleteIfPending(ctx context.Context, id uuid.UUID) (bool, error)
	SetPaymentAttemptID(ctx context.Context, id uuid.UUID, attemptID string) error
	FindPendingByAttemptID(ctx context.Context, attemptID string) (*ent.Order, error)
	DeletePendingByAttemptIDExcept(ctx context.Context, attemptID string, except uuid.UUID) (int64, error)

	// Aggregation
	GetEventMonths(ctx context.Context) ([]EventMonth, error)
}

type EventMonth struct {
	Year       int `json:"year" sql:"year"`
	Month      int `json:"month" sql:"month"`
	OrderCount int `json:"order_count" sql:"order_count"`
}

type orderRepo struct {
	client *ent.Client
}

func NewOrderRepository(client *ent.Client) OrderRepository {
	return &orderRepo{client: client}
}

func (r *orderRepo) ec(ctx context.Context) *ent.Client {
	return ClientFromContext(ctx, r.client)
}

func (r *orderRepo) Create(ctx context.Context, totalCents int64, status order.Status, origin order.Origin, customerID, contactEmail, paymentAttemptID *string, payrexxGatewayID, payrexxTransactionID *int) (*ent.Order, error) {
	builder := r.ec(ctx).Order.Create().
		SetTotalCents(totalCents).
		SetStatus(status).
		SetOrigin(origin)
	if customerID != nil {
		builder.SetCustomerID(*customerID)
	}
	if contactEmail != nil {
		builder.SetContactEmail(*contactEmail)
	}
	if paymentAttemptID != nil {
		builder.SetPaymentAttemptID(*paymentAttemptID)
	}
	if payrexxGatewayID != nil {
		builder.SetPayrexxGatewayID(*payrexxGatewayID)
	}
	if payrexxTransactionID != nil {
		builder.SetPayrexxTransactionID(*payrexxTransactionID)
	}
	created, err := builder.Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return created, nil
}

func (r *orderRepo) GetByID(ctx context.Context, id uuid.UUID) (*ent.Order, error) {
	e, err := r.ec(ctx).Order.Get(ctx, id)
	if err != nil {
		return nil, translateError(err)
	}
	return e, nil
}

func (r *orderRepo) GetByIDWithRelations(ctx context.Context, id uuid.UUID) (*ent.Order, error) {
	e, err := r.ec(ctx).Order.Query().
		Where(order.ID(id)).
		WithPayments().
		WithLines(func(q *ent.OrderLineQuery) {
			q.WithProduct().
				WithChildLines().
				WithRedemption()
		}).
		Only(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return e, nil
}

func (r *orderRepo) GetByCustomerID(ctx context.Context, customerID string) ([]*ent.Order, error) {
	rows, err := r.ec(ctx).Order.Query().
		Where(order.CustomerIDEQ(customerID)).
		WithLines().
		Order(order.ByCreatedAt(entDescOpt())).
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

func (r *orderRepo) GetByStatus(ctx context.Context, status order.Status) ([]*ent.Order, error) {
	rows, err := r.ec(ctx).Order.Query().
		Where(order.StatusEQ(status)).
		Order(order.ByCreatedAt(entDescOpt())).
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

func (r *orderRepo) GetByDateRange(ctx context.Context, start, end time.Time) ([]*ent.Order, error) {
	rows, err := r.ec(ctx).Order.Query().
		Where(
			order.CreatedAtGTE(start),
			order.CreatedAtLTE(end),
		).
		Order(order.ByCreatedAt(entDescOpt())).
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

func (r *orderRepo) GetRecent(ctx context.Context, limit int) ([]*ent.Order, error) {
	rows, err := r.ec(ctx).Order.Query().
		Order(order.ByCreatedAt(entDescOpt())).
		Limit(limit).
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

func (r *orderRepo) Update(ctx context.Context, id uuid.UUID, totalCents int64, status order.Status, origin order.Origin, customerID, contactEmail, paymentAttemptID *string, payrexxGatewayID, payrexxTransactionID *int) (*ent.Order, error) {
	builder := r.ec(ctx).Order.UpdateOneID(id).
		SetTotalCents(totalCents).
		SetStatus(status).
		SetOrigin(origin)
	if customerID != nil {
		builder.SetCustomerID(*customerID)
	} else {
		builder.ClearCustomerID()
	}
	if contactEmail != nil {
		builder.SetContactEmail(*contactEmail)
	} else {
		builder.ClearContactEmail()
	}
	if paymentAttemptID != nil {
		builder.SetPaymentAttemptID(*paymentAttemptID)
	} else {
		builder.ClearPaymentAttemptID()
	}
	if payrexxGatewayID != nil {
		builder.SetPayrexxGatewayID(*payrexxGatewayID)
	} else {
		builder.ClearPayrexxGatewayID()
	}
	if payrexxTransactionID != nil {
		builder.SetPayrexxTransactionID(*payrexxTransactionID)
	} else {
		builder.ClearPayrexxTransactionID()
	}
	updated, err := builder.Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return updated, nil
}

func (r *orderRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status order.Status) error {
	_, err := r.ec(ctx).Order.UpdateOneID(id).
		SetStatus(status).
		Save(ctx)
	return translateError(err)
}

func (r *orderRepo) ListAdmin(ctx context.Context, status *order.Status, from, to *time.Time, q *string, limit, offset int) ([]*ent.Order, int64, error) {
	// Build count query with filters
	countQ := r.ec(ctx).Order.Query()
	if status != nil {
		countQ = countQ.Where(order.StatusEQ(*status))
	}
	if from != nil {
		countQ = countQ.Where(order.CreatedAtGTE(*from))
	}
	if to != nil {
		countQ = countQ.Where(order.CreatedAtLT(*to))
	}
	if q != nil && *q != "" {
		searchPattern := *q
		countQ = countQ.Where(
			order.Or(
				order.ContactEmailContainsFold(searchPattern),
				// For UUID prefix search, use a raw predicate
				func(s *sql.Selector) {
					s.Where(sql.Like(s.C(order.FieldID)+"::text", searchPattern+"%"))
				},
			),
		)
	}

	total, err := countQ.Count(ctx)
	if err != nil {
		return nil, 0, translateError(err)
	}

	// Build data query with same filters
	dataQ := r.ec(ctx).Order.Query().
		WithLines().
		WithPayments()
	if status != nil {
		dataQ = dataQ.Where(order.StatusEQ(*status))
	}
	if from != nil {
		dataQ = dataQ.Where(order.CreatedAtGTE(*from))
	}
	if to != nil {
		dataQ = dataQ.Where(order.CreatedAtLT(*to))
	}
	if q != nil && *q != "" {
		searchPattern := *q
		dataQ = dataQ.Where(
			order.Or(
				order.ContactEmailContainsFold(searchPattern),
				func(s *sql.Selector) {
					s.Where(sql.Like(s.C(order.FieldID)+"::text", searchPattern+"%"))
				},
			),
		)
	}

	rows, err := dataQ.
		Order(order.ByCreatedAt(entDescOpt())).
		Limit(limit).
		Offset(offset).
		All(ctx)
	if err != nil {
		return nil, 0, translateError(err)
	}

	return rows, int64(total), nil
}

func (r *orderRepo) ListByCustomerIDPaginated(ctx context.Context, customerID string, limit, offset int) ([]*ent.Order, int64, error) {
	total, err := r.ec(ctx).Order.Query().
		Where(order.CustomerIDEQ(customerID)).
		Count(ctx)
	if err != nil {
		return nil, 0, translateError(err)
	}

	rows, err := r.ec(ctx).Order.Query().
		Where(order.CustomerIDEQ(customerID)).
		WithLines().
		Order(order.ByCreatedAt(entDescOpt())).
		Limit(limit).
		Offset(offset).
		All(ctx)
	if err != nil {
		return nil, 0, translateError(err)
	}

	return rows, int64(total), nil
}

func (r *orderRepo) setPosPayment(ctx context.Context, orderID uuid.UUID, deviceID *uuid.UUID, method orderpayment.Method, amountCents int64) error {
	tx, err := r.ec(ctx).Tx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	payBuilder := tx.OrderPayment.Create().
		SetOrderID(orderID).
		SetMethod(method).
		SetAmountCents(amountCents).
		SetPaidAt(time.Now())
	if deviceID != nil {
		payBuilder.SetDeviceID(*deviceID)
	}
	if _, err := payBuilder.Save(ctx); err != nil {
		return translateError(err)
	}

	if _, err := tx.Order.UpdateOneID(orderID).
		SetStatus(order.StatusPaid).
		Save(ctx); err != nil {
		return translateError(err)
	}

	return tx.Commit()
}

func (r *orderRepo) SetPosPaymentCash(ctx context.Context, orderID uuid.UUID, deviceID *uuid.UUID, amountCents int64) error {
	return r.setPosPayment(ctx, orderID, deviceID, orderpayment.MethodCASH, amountCents)
}

func (r *orderRepo) SetPosPaymentCard(ctx context.Context, orderID uuid.UUID, deviceID *uuid.UUID, amountCents int64) error {
	return r.setPosPayment(ctx, orderID, deviceID, orderpayment.MethodCARD, amountCents)
}

func (r *orderRepo) SetPosPaymentTwint(ctx context.Context, orderID uuid.UUID, deviceID *uuid.UUID, amountCents int64) error {
	return r.setPosPayment(ctx, orderID, deviceID, orderpayment.MethodTWINT, amountCents)
}

func (r *orderRepo) SetPosPaymentGratisGuest(ctx context.Context, orderID uuid.UUID, deviceID *uuid.UUID, amountCents int64) error {
	return r.setPosPayment(ctx, orderID, deviceID, orderpayment.MethodGRATIS_GUEST, amountCents)
}

func (r *orderRepo) SetPosPaymentGratisVIP(ctx context.Context, orderID uuid.UUID, deviceID *uuid.UUID, amountCents int64) error {
	return r.setPosPayment(ctx, orderID, deviceID, orderpayment.MethodGRATIS_VIP, amountCents)
}

func (r *orderRepo) SetPosPaymentGratisStaff(ctx context.Context, orderID uuid.UUID, deviceID *uuid.UUID, amountCents int64) error {
	return r.setPosPayment(ctx, orderID, deviceID, orderpayment.MethodGRATIS_STAFF, amountCents)
}

func (r *orderRepo) SetPosPaymentGratis100Club(ctx context.Context, orderID uuid.UUID, deviceID *uuid.UUID, amountCents int64) error {
	return r.setPosPayment(ctx, orderID, deviceID, orderpayment.MethodGRATIS_100CLUB, amountCents)
}

func (r *orderRepo) DeleteIfPending(ctx context.Context, id uuid.UUID) (bool, error) {
	n, err := r.ec(ctx).Order.Delete().
		Where(
			order.ID(id),
			order.StatusEQ(order.StatusPending),
		).
		Exec(ctx)
	if err != nil {
		return false, translateError(err)
	}
	return n > 0, nil
}

func (r *orderRepo) SetPaymentAttemptID(ctx context.Context, id uuid.UUID, attemptID string) error {
	if attemptID == "" {
		return nil
	}
	_, err := r.ec(ctx).Order.UpdateOneID(id).
		SetPaymentAttemptID(attemptID).
		Save(ctx)
	return translateError(err)
}

func (r *orderRepo) FindPendingByAttemptID(ctx context.Context, attemptID string) (*ent.Order, error) {
	e, err := r.ec(ctx).Order.Query().
		Where(
			order.PaymentAttemptIDEQ(attemptID),
			order.StatusEQ(order.StatusPending),
		).
		Only(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return e, nil
}

func (r *orderRepo) DeletePendingByAttemptIDExcept(ctx context.Context, attemptID string, except uuid.UUID) (int64, error) {
	if attemptID == "" {
		return 0, nil
	}
	n, err := r.ec(ctx).Order.Delete().
		Where(
			order.PaymentAttemptIDEQ(attemptID),
			order.StatusEQ(order.StatusPending),
			order.IDNEQ(except),
		).
		Exec(ctx)
	if err != nil {
		return 0, translateError(err)
	}
	return int64(n), nil
}

func (r *orderRepo) GetEventMonths(ctx context.Context) ([]EventMonth, error) {
	var result []EventMonth
	err := r.ec(ctx).Order.Query().
		Where(order.StatusEQ(order.StatusPaid)).
		Modify(func(s *sql.Selector) {
			yearExpr := "EXTRACT(YEAR FROM created_at)::int"
			monthExpr := "EXTRACT(MONTH FROM created_at)::int"
			s.Select(
				sql.As(yearExpr, "year"),
				sql.As(monthExpr, "month"),
				sql.As("COUNT(*)", "order_count"),
			)
			s.GroupBy(yearExpr, monthExpr)
			s.OrderBy(sql.Desc("year"), sql.Desc("month"))
		}).
		Scan(ctx, &result)
	if err != nil {
		return nil, translateError(err)
	}
	return result, nil
}

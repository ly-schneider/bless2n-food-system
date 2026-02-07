package repository

import (
	"context"
	"time"

	"backend/internal/generated/ent"
	"backend/internal/generated/ent/orderline"
	"backend/internal/generated/ent/orderlineredemption"

	"github.com/google/uuid"
)

type OrderLineRedemptionRepository interface {
	Create(ctx context.Context, orderLineID uuid.UUID, redeemedAt time.Time) (*ent.OrderLineRedemption, error)
	CreateBatch(ctx context.Context, orderLineIDs []uuid.UUID) ([]*ent.OrderLineRedemption, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.OrderLineRedemption, error)
	GetByOrderLineID(ctx context.Context, orderLineID uuid.UUID) (*ent.OrderLineRedemption, error)
	ExistsByOrderLineID(ctx context.Context, orderLineID uuid.UUID) (bool, error)
	RedeemUnredeemedByOrderLineIDs(ctx context.Context, orderLineIDs []uuid.UUID) (int64, error)
}

type orderLineRedemptionRepo struct {
	client *ent.Client
}

func NewOrderLineRedemptionRepository(client *ent.Client) OrderLineRedemptionRepository {
	return &orderLineRedemptionRepo{client: client}
}

func (r *orderLineRedemptionRepo) ec(ctx context.Context) *ent.Client {
	return ClientFromContext(ctx, r.client)
}

func (r *orderLineRedemptionRepo) Create(ctx context.Context, orderLineID uuid.UUID, redeemedAt time.Time) (*ent.OrderLineRedemption, error) {
	created, err := r.ec(ctx).OrderLineRedemption.Create().
		SetOrderLineID(orderLineID).
		SetRedeemedAt(redeemedAt).
		Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return created, nil
}

func (r *orderLineRedemptionRepo) CreateBatch(ctx context.Context, orderLineIDs []uuid.UUID) ([]*ent.OrderLineRedemption, error) {
	if len(orderLineIDs) == 0 {
		return nil, nil
	}
	builders := make([]*ent.OrderLineRedemptionCreate, len(orderLineIDs))
	for i, olID := range orderLineIDs {
		builders[i] = r.ec(ctx).OrderLineRedemption.Create().
			SetOrderLineID(olID)
	}
	created, err := r.ec(ctx).OrderLineRedemption.CreateBulk(builders...).Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return created, nil
}

func (r *orderLineRedemptionRepo) GetByID(ctx context.Context, id uuid.UUID) (*ent.OrderLineRedemption, error) {
	e, err := r.ec(ctx).OrderLineRedemption.Get(ctx, id)
	if err != nil {
		return nil, translateError(err)
	}
	return e, nil
}

func (r *orderLineRedemptionRepo) GetByOrderLineID(ctx context.Context, orderLineID uuid.UUID) (*ent.OrderLineRedemption, error) {
	e, err := r.ec(ctx).OrderLineRedemption.Query().
		Where(orderlineredemption.OrderLineIDEQ(orderLineID)).
		Only(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return e, nil
}

func (r *orderLineRedemptionRepo) ExistsByOrderLineID(ctx context.Context, orderLineID uuid.UUID) (bool, error) {
	exists, err := r.ec(ctx).OrderLineRedemption.Query().
		Where(orderlineredemption.OrderLineIDEQ(orderLineID)).
		Exist(ctx)
	if err != nil {
		return false, translateError(err)
	}
	return exists, nil
}

func (r *orderLineRedemptionRepo) RedeemUnredeemedByOrderLineIDs(ctx context.Context, orderLineIDs []uuid.UUID) (int64, error) {
	if len(orderLineIDs) == 0 {
		return 0, nil
	}

	// Find unredeemed order line IDs (those without a redemption record)
	unredeemedLines, err := r.ec(ctx).OrderLine.Query().
		Where(
			orderline.IDIn(orderLineIDs...),
			orderline.Not(orderline.HasRedemption()),
		).
		All(ctx)
	if err != nil {
		return 0, translateError(err)
	}

	if len(unredeemedLines) == 0 {
		return 0, nil
	}

	// Create redemptions for unredeemed lines
	builders := make([]*ent.OrderLineRedemptionCreate, len(unredeemedLines))
	for i, line := range unredeemedLines {
		builders[i] = r.ec(ctx).OrderLineRedemption.Create().
			SetOrderLineID(line.ID)
	}
	_, err = r.ec(ctx).OrderLineRedemption.CreateBulk(builders...).Save(ctx)
	if err != nil {
		return 0, translateError(err)
	}

	return int64(len(unredeemedLines)), nil
}

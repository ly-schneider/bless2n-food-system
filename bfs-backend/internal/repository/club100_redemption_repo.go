package repository

import (
	"context"

	"backend/internal/generated/ent"
	"backend/internal/generated/ent/club100redemption"

	"github.com/google/uuid"
)

type Club100RedemptionRepository interface {
	Create(ctx context.Context, elvantoPersonID, elvantoPersonName string, orderID uuid.UUID, quantity int) (*ent.Club100Redemption, error)
	GetTotalRedemptions(ctx context.Context, elvantoPersonID string) (int, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*ent.Club100Redemption, error)
}

type club100RedemptionRepo struct {
	client *ent.Client
}

func NewClub100RedemptionRepository(client *ent.Client) Club100RedemptionRepository {
	return &club100RedemptionRepo{client: client}
}

func (r *club100RedemptionRepo) ec(ctx context.Context) *ent.Client {
	return ClientFromContext(ctx, r.client)
}

func (r *club100RedemptionRepo) Create(ctx context.Context, elvantoPersonID, elvantoPersonName string, orderID uuid.UUID, quantity int) (*ent.Club100Redemption, error) {
	e, err := r.ec(ctx).Club100Redemption.Create().
		SetElvantoPersonID(elvantoPersonID).
		SetElvantoPersonName(elvantoPersonName).
		SetOrderID(orderID).
		SetFreeProductQuantity(quantity).
		Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return e, nil
}

func (r *club100RedemptionRepo) GetTotalRedemptions(ctx context.Context, elvantoPersonID string) (int, error) {
	rows, err := r.ec(ctx).Club100Redemption.Query().
		Where(club100redemption.ElvantoPersonIDEQ(elvantoPersonID)).
		All(ctx)
	if err != nil {
		return 0, translateError(err)
	}
	total := 0
	for _, row := range rows {
		total += row.FreeProductQuantity
	}
	return total, nil
}

func (r *club100RedemptionRepo) GetByOrderID(ctx context.Context, orderID uuid.UUID) ([]*ent.Club100Redemption, error) {
	rows, err := r.ec(ctx).Club100Redemption.Query().
		Where(club100redemption.OrderIDEQ(orderID)).
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

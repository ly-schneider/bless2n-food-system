package repository

import (
	"context"

	"backend/internal/generated/ent"

	"github.com/google/uuid"
)

type MenuSlotOptionRepository interface {
	Create(ctx context.Context, menuSlotID, optionProductID uuid.UUID) (*ent.MenuSlotOption, error)
	CreateBatch(ctx context.Context, menuSlotID uuid.UUID, optionProductIDs []uuid.UUID) ([]*ent.MenuSlotOption, error)
	GetByMenuSlotID(ctx context.Context, menuSlotID uuid.UUID) ([]*ent.MenuSlotOption, error)
	Delete(ctx context.Context, menuSlotID, optionProductID uuid.UUID) error
	DeleteByMenuSlotID(ctx context.Context, menuSlotID uuid.UUID) error
	CountByOptionProductID(ctx context.Context, optionProductID uuid.UUID) (int64, error)
}

type menuSlotOptionRepo struct {
	client *ent.Client
}

func NewMenuSlotOptionRepository(client *ent.Client) MenuSlotOptionRepository {
	return &menuSlotOptionRepo{client: client}
}

func (r *menuSlotOptionRepo) ec(ctx context.Context) *ent.Client {
	return ClientFromContext(ctx, r.client)
}

func (r *menuSlotOptionRepo) Create(ctx context.Context, menuSlotID, optionProductID uuid.UUID) (*ent.MenuSlotOption, error) {
	created, err := r.ec(ctx).MenuSlotOption.Create().
		SetMenuSlotID(menuSlotID).
		SetOptionProductID(optionProductID).
		Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return created, nil
}

func (r *menuSlotOptionRepo) CreateBatch(ctx context.Context, menuSlotID uuid.UUID, optionProductIDs []uuid.UUID) ([]*ent.MenuSlotOption, error) {
	if len(optionProductIDs) == 0 {
		return nil, nil
	}
	builders := make([]*ent.MenuSlotOptionCreate, len(optionProductIDs))
	for i, pid := range optionProductIDs {
		builders[i] = r.ec(ctx).MenuSlotOption.Create().
			SetMenuSlotID(menuSlotID).
			SetOptionProductID(pid)
	}
	created, err := r.ec(ctx).MenuSlotOption.CreateBulk(builders...).Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return created, nil
}

func (r *menuSlotOptionRepo) GetByMenuSlotID(ctx context.Context, menuSlotID uuid.UUID) ([]*ent.MenuSlotOption, error) {
	rows, err := r.ec(ctx).MenuSlotOption.Query().
		Where(entMenuSlotOptionMenuSlotID(menuSlotID)).
		WithOptionProduct().
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

func (r *menuSlotOptionRepo) Delete(ctx context.Context, menuSlotID, optionProductID uuid.UUID) error {
	_, err := r.ec(ctx).MenuSlotOption.Delete().
		Where(
			entMenuSlotOptionMenuSlotID(menuSlotID),
			entMenuSlotOptionProductID(optionProductID),
		).
		Exec(ctx)
	return translateError(err)
}

func (r *menuSlotOptionRepo) DeleteByMenuSlotID(ctx context.Context, menuSlotID uuid.UUID) error {
	_, err := r.ec(ctx).MenuSlotOption.Delete().
		Where(entMenuSlotOptionMenuSlotID(menuSlotID)).
		Exec(ctx)
	return translateError(err)
}

func (r *menuSlotOptionRepo) CountByOptionProductID(ctx context.Context, optionProductID uuid.UUID) (int64, error) {
	count, err := r.ec(ctx).MenuSlotOption.Query().
		Where(entMenuSlotOptionProductID(optionProductID)).
		Count(ctx)
	if err != nil {
		return 0, translateError(err)
	}
	return int64(count), nil
}

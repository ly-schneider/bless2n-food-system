package repository

import (
	"context"

	"backend/internal/generated/ent"
	"backend/internal/generated/ent/menuslot"

	"github.com/google/uuid"
)

type MenuSlotRepository interface {
	Create(ctx context.Context, menuProductID uuid.UUID, name string, sequence int) (*ent.MenuSlot, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.MenuSlot, error)
	GetByMenuProductID(ctx context.Context, menuProductID uuid.UUID) ([]*ent.MenuSlot, error)
	Update(ctx context.Context, id, menuProductID uuid.UUID, name string, sequence int) (*ent.MenuSlot, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByMenuProductID(ctx context.Context, menuProductID uuid.UUID) error
}

type menuSlotRepo struct {
	client *ent.Client
}

func NewMenuSlotRepository(client *ent.Client) MenuSlotRepository {
	return &menuSlotRepo{client: client}
}

func (r *menuSlotRepo) ec(ctx context.Context) *ent.Client {
	return ClientFromContext(ctx, r.client)
}

func (r *menuSlotRepo) Create(ctx context.Context, menuProductID uuid.UUID, name string, sequence int) (*ent.MenuSlot, error) {
	created, err := r.ec(ctx).MenuSlot.Create().
		SetMenuProductID(menuProductID).
		SetName(name).
		SetSequence(sequence).
		Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return created, nil
}

func (r *menuSlotRepo) GetByID(ctx context.Context, id uuid.UUID) (*ent.MenuSlot, error) {
	e, err := r.ec(ctx).MenuSlot.Query().
		Where(menuslot.ID(id)).
		WithOptions(func(oq *ent.MenuSlotOptionQuery) {
			oq.WithOptionProduct()
		}).
		Only(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return e, nil
}

func (r *menuSlotRepo) GetByMenuProductID(ctx context.Context, menuProductID uuid.UUID) ([]*ent.MenuSlot, error) {
	rows, err := r.ec(ctx).MenuSlot.Query().
		Where(menuslot.MenuProductIDEQ(menuProductID)).
		WithOptions(func(oq *ent.MenuSlotOptionQuery) {
			oq.WithOptionProduct()
		}).
		Order(menuslot.BySequence()).
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

func (r *menuSlotRepo) Update(ctx context.Context, id, menuProductID uuid.UUID, name string, sequence int) (*ent.MenuSlot, error) {
	updated, err := r.ec(ctx).MenuSlot.UpdateOneID(id).
		SetName(name).
		SetSequence(sequence).
		SetMenuProductID(menuProductID).
		Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return updated, nil
}

func (r *menuSlotRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return translateError(r.ec(ctx).MenuSlot.DeleteOneID(id).Exec(ctx))
}

func (r *menuSlotRepo) DeleteByMenuProductID(ctx context.Context, menuProductID uuid.UUID) error {
	_, err := r.ec(ctx).MenuSlot.Delete().
		Where(menuslot.MenuProductIDEQ(menuProductID)).
		Exec(ctx)
	return translateError(err)
}

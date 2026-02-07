package repository

import (
	"context"

	"backend/internal/generated/ent"
	"backend/internal/generated/ent/jeton"

	"github.com/google/uuid"
)

type JetonRepository interface {
	Create(ctx context.Context, name, color string) (*ent.Jeton, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.Jeton, error)
	GetAll(ctx context.Context) ([]*ent.Jeton, error)
	Update(ctx context.Context, id uuid.UUID, name, color string) (*ent.Jeton, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type jetonRepo struct {
	client *ent.Client
}

func NewJetonRepository(client *ent.Client) JetonRepository {
	return &jetonRepo{client: client}
}

func (r *jetonRepo) ec(ctx context.Context) *ent.Client {
	return ClientFromContext(ctx, r.client)
}

func (r *jetonRepo) Create(ctx context.Context, name, color string) (*ent.Jeton, error) {
	created, err := r.ec(ctx).Jeton.Create().
		SetName(name).
		SetColor(color).
		Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return created, nil
}

func (r *jetonRepo) GetByID(ctx context.Context, id uuid.UUID) (*ent.Jeton, error) {
	e, err := r.ec(ctx).Jeton.Get(ctx, id)
	if err != nil {
		return nil, translateError(err)
	}
	return e, nil
}

func (r *jetonRepo) GetAll(ctx context.Context) ([]*ent.Jeton, error) {
	rows, err := r.ec(ctx).Jeton.Query().
		Order(jeton.ByName()).
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

func (r *jetonRepo) Update(ctx context.Context, id uuid.UUID, name, color string) (*ent.Jeton, error) {
	updated, err := r.ec(ctx).Jeton.UpdateOneID(id).
		SetName(name).
		SetColor(color).
		Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return updated, nil
}

func (r *jetonRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return translateError(r.ec(ctx).Jeton.DeleteOneID(id).Exec(ctx))
}

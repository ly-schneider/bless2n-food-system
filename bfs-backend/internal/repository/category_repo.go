package repository

import (
	"context"

	"backend/internal/generated/ent"
	"backend/internal/generated/ent/category"

	"github.com/google/uuid"
)

type CategoryRepository interface {
	Create(ctx context.Context, name string, position int, isActive bool) (*ent.Category, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.Category, error)
	GetAll(ctx context.Context) ([]*ent.Category, error)
	GetAllActive(ctx context.Context) ([]*ent.Category, error)
	List(ctx context.Context, limit, offset int) ([]*ent.Category, int64, error)
	Update(ctx context.Context, id uuid.UUID, name string, position int, isActive bool) (*ent.Category, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type categoryRepo struct {
	client *ent.Client
}

func NewCategoryRepository(client *ent.Client) CategoryRepository {
	return &categoryRepo{client: client}
}

func (r *categoryRepo) ec(ctx context.Context) *ent.Client {
	return ClientFromContext(ctx, r.client)
}

func (r *categoryRepo) Create(ctx context.Context, name string, position int, isActive bool) (*ent.Category, error) {
	created, err := r.ec(ctx).Category.Create().
		SetName(name).
		SetIsActive(isActive).
		SetPosition(position).
		Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return created, nil
}

func (r *categoryRepo) GetByID(ctx context.Context, id uuid.UUID) (*ent.Category, error) {
	e, err := r.ec(ctx).Category.Get(ctx, id)
	if err != nil {
		return nil, translateError(err)
	}
	return e, nil
}

func (r *categoryRepo) GetAll(ctx context.Context) ([]*ent.Category, error) {
	rows, err := r.ec(ctx).Category.Query().
		Order(category.ByPosition()).
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

func (r *categoryRepo) GetAllActive(ctx context.Context) ([]*ent.Category, error) {
	rows, err := r.ec(ctx).Category.Query().
		Where(category.IsActive(true)).
		Order(category.ByPosition()).
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

func (r *categoryRepo) List(ctx context.Context, limit, offset int) ([]*ent.Category, int64, error) {
	q := r.ec(ctx).Category.Query()

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, translateError(err)
	}

	rows, err := r.ec(ctx).Category.Query().
		Order(category.ByPosition()).
		Limit(limit).
		Offset(offset).
		All(ctx)
	if err != nil {
		return nil, 0, translateError(err)
	}

	return rows, int64(total), nil
}

func (r *categoryRepo) Update(ctx context.Context, id uuid.UUID, name string, position int, isActive bool) (*ent.Category, error) {
	updated, err := r.ec(ctx).Category.UpdateOneID(id).
		SetName(name).
		SetIsActive(isActive).
		SetPosition(position).
		Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return updated, nil
}

func (r *categoryRepo) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.ec(ctx).Category.DeleteOneID(id).Exec(ctx)
	return translateError(err)
}

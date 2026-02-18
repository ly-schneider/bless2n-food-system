package repository

import (
	"context"

	"backend/internal/generated/ent"
	"backend/internal/generated/ent/product"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
)

type ProductRepository struct {
	client *ent.Client
}

func NewProductRepository(client *ent.Client) *ProductRepository {
	return &ProductRepository{client: client}
}

func (r *ProductRepository) Create(ctx context.Context, categoryID uuid.UUID, productType product.Type, name string, priceCents int64, isActive bool, image *string, jetonID *uuid.UUID) (*ent.Product, error) {
	builder := r.client.Product.Create().
		SetCategoryID(categoryID).
		SetType(productType).
		SetName(name).
		SetPriceCents(priceCents).
		SetIsActive(isActive)
	if image != nil {
		builder.SetImage(*image)
	}
	if jetonID != nil {
		builder.SetJetonID(*jetonID)
	}
	created, err := builder.Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return created, nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id uuid.UUID) (*ent.Product, error) {
	e, err := r.client.Product.Get(ctx, id)
	if err != nil {
		return nil, translateError(err)
	}
	return e, nil
}

func (r *ProductRepository) GetByIDWithRelations(ctx context.Context, id uuid.UUID) (*ent.Product, error) {
	e, err := r.client.Product.Query().
		Where(product.ID(id)).
		WithCategory().
		WithJeton().
		WithMenuSlots(func(q *ent.MenuSlotQuery) {
			q.WithOptions(func(oq *ent.MenuSlotOptionQuery) {
				oq.WithOptionProduct(func(pq *ent.ProductQuery) {
					pq.WithJeton()
				})
			})
		}).
		Only(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return e, nil
}

func (r *ProductRepository) GetAll(ctx context.Context) ([]*ent.Product, error) {
	rows, err := r.client.Product.Query().
		WithCategory().
		WithJeton().
		WithMenuSlots(func(q *ent.MenuSlotQuery) {
			q.WithOptions(func(oq *ent.MenuSlotOptionQuery) {
				oq.WithOptionProduct(func(pq *ent.ProductQuery) {
					pq.WithJeton()
				})
			})
		}).
		Order(product.ByName()).
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

func (r *ProductRepository) GetAllActive(ctx context.Context) ([]*ent.Product, error) {
	rows, err := r.client.Product.Query().
		Where(product.IsActive(true)).
		WithCategory().
		WithJeton().
		WithMenuSlots(func(q *ent.MenuSlotQuery) {
			q.WithOptions(func(oq *ent.MenuSlotOptionQuery) {
				oq.WithOptionProduct(func(pq *ent.ProductQuery) {
					pq.WithJeton()
				})
			})
		}).
		Order(product.ByName()).
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

func (r *ProductRepository) GetByCategory(ctx context.Context, categoryID uuid.UUID) ([]*ent.Product, error) {
	rows, err := r.client.Product.Query().
		Where(product.CategoryIDEQ(categoryID)).
		WithCategory().
		WithJeton().
		WithMenuSlots(func(q *ent.MenuSlotQuery) {
			q.WithOptions(func(oq *ent.MenuSlotOptionQuery) {
				oq.WithOptionProduct(func(pq *ent.ProductQuery) {
					pq.WithJeton()
				})
			})
		}).
		Order(product.ByName()).
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

func (r *ProductRepository) GetByCategoryActive(ctx context.Context, categoryID uuid.UUID) ([]*ent.Product, error) {
	rows, err := r.client.Product.Query().
		Where(
			product.CategoryIDEQ(categoryID),
			product.IsActive(true),
		).
		WithCategory().
		WithJeton().
		WithMenuSlots(func(q *ent.MenuSlotQuery) {
			q.WithOptions(func(oq *ent.MenuSlotOptionQuery) {
				oq.WithOptionProduct(func(pq *ent.ProductQuery) {
					pq.WithJeton()
				})
			})
		}).
		Order(product.ByName()).
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

func (r *ProductRepository) GetByIDs(ctx context.Context, ids []uuid.UUID) ([]*ent.Product, error) {
	rows, err := r.client.Product.Query().
		Where(product.IDIn(ids...)).
		WithCategory().
		WithJeton().
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

func (r *ProductRepository) Update(ctx context.Context, id, categoryID uuid.UUID, productType product.Type, name string, priceCents int64, isActive bool, image *string, jetonID *uuid.UUID) (*ent.Product, error) {
	builder := r.client.Product.UpdateOneID(id).
		SetCategoryID(categoryID).
		SetType(productType).
		SetName(name).
		SetPriceCents(priceCents).
		SetIsActive(isActive)
	if image != nil {
		builder.SetImage(*image)
	} else {
		builder.ClearImage()
	}
	if jetonID != nil {
		builder.SetJetonID(*jetonID)
	} else {
		builder.ClearJetonID()
	}
	updated, err := builder.Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return updated, nil
}

func (r *ProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return translateError(r.client.Product.DeleteOneID(id).Exec(ctx))
}

func (r *ProductRepository) GetMenus(ctx context.Context, q *string, active *bool, limit, offset int) ([]*ent.Product, int64, error) {
	// Build count query
	countQ := r.client.Product.Query().
		Where(product.TypeEQ(product.TypeMenu))
	if active != nil {
		countQ = countQ.Where(product.IsActive(*active))
	}
	if q != nil && *q != "" {
		countQ = countQ.Where(product.NameContainsFold(*q))
	}
	total, err := countQ.Count(ctx)
	if err != nil {
		return nil, 0, translateError(err)
	}

	// Build data query
	dataQ := r.client.Product.Query().
		Where(product.TypeEQ(product.TypeMenu)).
		WithCategory().
		WithJeton().
		WithMenuSlots(func(msq *ent.MenuSlotQuery) {
			msq.WithOptions(func(oq *ent.MenuSlotOptionQuery) {
				oq.WithOptionProduct(func(pq *ent.ProductQuery) {
					pq.WithJeton()
				})
			})
		})
	if active != nil {
		dataQ = dataQ.Where(product.IsActive(*active))
	}
	if q != nil && *q != "" {
		dataQ = dataQ.Where(product.NameContainsFold(*q))
	}
	rows, err := dataQ.
		Order(product.ByName()).
		Limit(limit).
		Offset(offset).
		All(ctx)
	if err != nil {
		return nil, 0, translateError(err)
	}

	return rows, int64(total), nil
}

func (r *ProductRepository) CountActiveWithoutJeton(ctx context.Context) (int64, error) {
	count, err := r.client.Product.Query().
		Where(
			product.IsActive(true),
			product.TypeEQ(product.TypeSimple),
			product.JetonIDIsNil(),
		).
		Count(ctx)
	if err != nil {
		return 0, translateError(err)
	}
	return int64(count), nil
}

func (r *ProductRepository) CountByJetonIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]int64, error) {
	result := make(map[uuid.UUID]int64)
	if len(ids) == 0 {
		return result, nil
	}

	// Use raw GroupBy query for aggregation
	var rows []struct {
		JetonID uuid.UUID `json:"jeton_id"`
		Count   int       `json:"count"`
	}
	err := r.client.Product.Query().
		Where(
			product.JetonIDIn(ids...),
			product.TypeEQ(product.TypeSimple),
		).
		GroupBy(product.FieldJetonID).
		Aggregate(ent.Count()).
		Scan(ctx, &rows)
	if err != nil {
		return nil, translateError(err)
	}

	for _, row := range rows {
		result[row.JetonID] = int64(row.Count)
	}
	return result, nil
}

func (r *ProductRepository) UpdateJeton(ctx context.Context, id uuid.UUID, jetonID *uuid.UUID) error {
	builder := r.client.Product.UpdateOneID(id)
	if jetonID != nil {
		builder.SetJetonID(*jetonID)
	} else {
		builder.ClearJetonID()
	}
	_, err := builder.Save(ctx)
	return translateError(err)
}

// productNameContainsILIKE provides ILIKE search via sql modifier.
// This is used when the generated NameContainsFold is not sufficient.
var _ = func() sql.Querier { return nil } // import anchor

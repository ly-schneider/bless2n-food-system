package repository

import (
	"context"
	"time"

	"backend/internal/generated/ent"
	"backend/internal/generated/ent/inventoryledger"

	"entgo.io/ent/dialect/sql"
)

type InventoryLedgerRepository interface {
	Create(ctx context.Context, productID string, delta int, reason inventoryledger.Reason, orderID, orderLineID, deviceID *string, createdBy *string) (*ent.InventoryLedger, error)
	CreateMany(ctx context.Context, entries []InventoryLedgerCreateParams) ([]*ent.InventoryLedger, error)
	GetByID(ctx context.Context, id string) (*ent.InventoryLedger, error)
	GetByProductID(ctx context.Context, productID string) ([]*ent.InventoryLedger, error)
	GetByProductIDWithPagination(ctx context.Context, productID string, limit, offset int) ([]*ent.InventoryLedger, error)
	GetByDateRange(ctx context.Context, start, end time.Time) ([]*ent.InventoryLedger, error)
	GetCurrentStock(ctx context.Context, productID string) (int, error)
	GetCurrentStockBatch(ctx context.Context, productIDs []string) (map[string]int, error)
	SumByProductIDs(ctx context.Context, ids []string) (map[string]int64, error)
}

// InventoryLedgerCreateParams holds the parameters for creating an inventory ledger entry in a batch.
type InventoryLedgerCreateParams struct {
	ProductID   string
	Delta       int
	Reason      inventoryledger.Reason
	OrderID     *string
	OrderLineID *string
	DeviceID    *string
	CreatedBy   *string
}

type inventoryLedgerRepo struct {
	client *ent.Client
}

func NewInventoryLedgerRepository(client *ent.Client) InventoryLedgerRepository {
	return &inventoryLedgerRepo{client: client}
}

func (r *inventoryLedgerRepo) ec(ctx context.Context) *ent.Client {
	return ClientFromContext(ctx, r.client)
}

func (r *inventoryLedgerRepo) Create(ctx context.Context, productID string, delta int, reason inventoryledger.Reason, orderID, orderLineID, deviceID *string, createdBy *string) (*ent.InventoryLedger, error) {
	builder := r.ec(ctx).InventoryLedger.Create().
		SetProductID(productID).
		SetDelta(delta).
		SetReason(reason)
	if orderID != nil {
		builder.SetOrderID(*orderID)
	}
	if orderLineID != nil {
		builder.SetOrderLineID(*orderLineID)
	}
	if deviceID != nil {
		builder.SetDeviceID(*deviceID)
	}
	if createdBy != nil {
		builder.SetCreatedBy(*createdBy)
	}
	created, err := builder.Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return created, nil
}

func (r *inventoryLedgerRepo) CreateMany(ctx context.Context, entries []InventoryLedgerCreateParams) ([]*ent.InventoryLedger, error) {
	if len(entries) == 0 {
		return nil, nil
	}
	builders := make([]*ent.InventoryLedgerCreate, len(entries))
	for i, entry := range entries {
		b := r.ec(ctx).InventoryLedger.Create().
			SetProductID(entry.ProductID).
			SetDelta(entry.Delta).
			SetReason(entry.Reason)
		if entry.OrderID != nil {
			b.SetOrderID(*entry.OrderID)
		}
		if entry.OrderLineID != nil {
			b.SetOrderLineID(*entry.OrderLineID)
		}
		if entry.DeviceID != nil {
			b.SetDeviceID(*entry.DeviceID)
		}
		if entry.CreatedBy != nil {
			b.SetCreatedBy(*entry.CreatedBy)
		}
		builders[i] = b
	}
	created, err := r.ec(ctx).InventoryLedger.CreateBulk(builders...).Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return created, nil
}

func (r *inventoryLedgerRepo) GetByID(ctx context.Context, id string) (*ent.InventoryLedger, error) {
	e, err := r.ec(ctx).InventoryLedger.Get(ctx, id)
	if err != nil {
		return nil, translateError(err)
	}
	return e, nil
}

func (r *inventoryLedgerRepo) GetByProductID(ctx context.Context, productID string) ([]*ent.InventoryLedger, error) {
	rows, err := r.ec(ctx).InventoryLedger.Query().
		Where(inventoryledger.ProductIDEQ(productID)).
		Order(inventoryledger.ByCreatedAt(entDescOpt())).
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

func (r *inventoryLedgerRepo) GetByProductIDWithPagination(ctx context.Context, productID string, limit, offset int) ([]*ent.InventoryLedger, error) {
	rows, err := r.ec(ctx).InventoryLedger.Query().
		Where(inventoryledger.ProductIDEQ(productID)).
		Order(inventoryledger.ByCreatedAt(entDescOpt())).
		Limit(limit).
		Offset(offset).
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

func (r *inventoryLedgerRepo) GetByDateRange(ctx context.Context, start, end time.Time) ([]*ent.InventoryLedger, error) {
	rows, err := r.ec(ctx).InventoryLedger.Query().
		Where(
			inventoryledger.CreatedAtGTE(start),
			inventoryledger.CreatedAtLTE(end),
		).
		Order(inventoryledger.ByCreatedAt(entDescOpt())).
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

func (r *inventoryLedgerRepo) GetCurrentStock(ctx context.Context, productID string) (int, error) {
	// Use raw SQL aggregation via Modify
	var result []struct {
		Sum int `json:"sum"`
	}
	err := r.ec(ctx).InventoryLedger.Query().
		Where(inventoryledger.ProductIDEQ(productID)).
		Modify(func(s *sql.Selector) {
			s.Select(sql.As(sql.Sum(s.C(inventoryledger.FieldDelta)), "sum"))
		}).
		Scan(ctx, &result)
	if err != nil {
		return 0, translateError(err)
	}
	if len(result) == 0 {
		return 0, nil
	}
	return result[0].Sum, nil
}

func (r *inventoryLedgerRepo) GetCurrentStockBatch(ctx context.Context, productIDs []string) (map[string]int, error) {
	if len(productIDs) == 0 {
		return make(map[string]int), nil
	}

	var results []struct {
		ProductID string `json:"product_id"`
		Stock     int    `json:"stock"`
	}
	err := r.ec(ctx).InventoryLedger.Query().
		Where(inventoryledger.ProductIDIn(productIDs...)).
		Modify(func(s *sql.Selector) {
			s.Select(
				s.C(inventoryledger.FieldProductID),
				sql.As(sql.Sum(s.C(inventoryledger.FieldDelta)), "stock"),
			).GroupBy(s.C(inventoryledger.FieldProductID))
		}).
		Scan(ctx, &results)
	if err != nil {
		return nil, translateError(err)
	}

	stocks := make(map[string]int, len(results))
	for _, r := range results {
		stocks[r.ProductID] = r.Stock
	}
	// Ensure all requested product IDs are in the map (default 0)
	for _, id := range productIDs {
		if _, ok := stocks[id]; !ok {
			stocks[id] = 0
		}
	}
	return stocks, nil
}

func (r *inventoryLedgerRepo) SumByProductIDs(ctx context.Context, ids []string) (map[string]int64, error) {
	result := make(map[string]int64)
	if len(ids) == 0 {
		return result, nil
	}

	var results []struct {
		ProductID string `json:"product_id"`
		Total     int64  `json:"total"`
	}
	err := r.ec(ctx).InventoryLedger.Query().
		Where(inventoryledger.ProductIDIn(ids...)).
		Modify(func(s *sql.Selector) {
			s.Select(
				s.C(inventoryledger.FieldProductID),
				sql.As(sql.Sum(s.C(inventoryledger.FieldDelta)), "total"),
			).GroupBy(s.C(inventoryledger.FieldProductID))
		}).
		Scan(ctx, &results)
	if err != nil {
		return nil, translateError(err)
	}

	for _, r := range results {
		result[r.ProductID] = r.Total
	}
	return result, nil
}

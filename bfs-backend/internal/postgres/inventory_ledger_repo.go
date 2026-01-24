package postgres

import (
	"context"
	"time"

	"backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// InventoryLedgerRepository defines the interface for inventory ledger data access.
type InventoryLedgerRepository interface {
	Create(ctx context.Context, entry *model.InventoryLedger) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.InventoryLedger, error)
	GetByProductID(ctx context.Context, productID uuid.UUID) ([]model.InventoryLedger, error)
	GetByProductIDWithPagination(ctx context.Context, productID uuid.UUID, limit, offset int) ([]model.InventoryLedger, error)
	GetByDateRange(ctx context.Context, start, end time.Time) ([]model.InventoryLedger, error)
	GetCurrentStock(ctx context.Context, productID uuid.UUID) (int, error)
	GetCurrentStockBatch(ctx context.Context, productIDs []uuid.UUID) (map[uuid.UUID]int, error)
}

type inventoryLedgerRepo struct {
	db *gorm.DB
}

// NewInventoryLedgerRepository creates a new InventoryLedgerRepository.
func NewInventoryLedgerRepository(db *gorm.DB) InventoryLedgerRepository {
	return &inventoryLedgerRepo{db: db}
}

func (r *inventoryLedgerRepo) Create(ctx context.Context, entry *model.InventoryLedger) error {
	return translateError(r.db.WithContext(ctx).Create(entry).Error)
}

func (r *inventoryLedgerRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.InventoryLedger, error) {
	var entry model.InventoryLedger
	err := r.db.WithContext(ctx).First(&entry, "id = ?", id).Error
	return &entry, translateError(err)
}

func (r *inventoryLedgerRepo) GetByProductID(ctx context.Context, productID uuid.UUID) ([]model.InventoryLedger, error) {
	var entries []model.InventoryLedger
	err := r.db.WithContext(ctx).
		Where("product_id = ?", productID).
		Order("created_at DESC").
		Find(&entries).Error
	return entries, translateError(err)
}

func (r *inventoryLedgerRepo) GetByProductIDWithPagination(ctx context.Context, productID uuid.UUID, limit, offset int) ([]model.InventoryLedger, error) {
	var entries []model.InventoryLedger
	err := r.db.WithContext(ctx).
		Where("product_id = ?", productID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&entries).Error
	return entries, translateError(err)
}

func (r *inventoryLedgerRepo) GetByDateRange(ctx context.Context, start, end time.Time) ([]model.InventoryLedger, error) {
	var entries []model.InventoryLedger
	err := r.db.WithContext(ctx).
		Where("created_at BETWEEN ? AND ?", start, end).
		Order("created_at DESC").
		Find(&entries).Error
	return entries, translateError(err)
}

func (r *inventoryLedgerRepo) GetCurrentStock(ctx context.Context, productID uuid.UUID) (int, error) {
	var sum *int
	err := r.db.WithContext(ctx).
		Model(&model.InventoryLedger{}).
		Select("COALESCE(SUM(delta), 0)").
		Where("product_id = ?", productID).
		Scan(&sum).Error
	if err != nil {
		return 0, translateError(err)
	}
	if sum == nil {
		return 0, nil
	}
	return *sum, nil
}

func (r *inventoryLedgerRepo) GetCurrentStockBatch(ctx context.Context, productIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	if len(productIDs) == 0 {
		return make(map[uuid.UUID]int), nil
	}

	type result struct {
		ProductID uuid.UUID
		Stock     int
	}
	var results []result

	err := r.db.WithContext(ctx).
		Model(&model.InventoryLedger{}).
		Select("product_id, COALESCE(SUM(delta), 0) as stock").
		Where("product_id IN ?", productIDs).
		Group("product_id").
		Scan(&results).Error
	if err != nil {
		return nil, translateError(err)
	}

	stocks := make(map[uuid.UUID]int, len(results))
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

package postgres

import (
	"context"

	"backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DeviceProductRepository defines the interface for device product assignment data access.
type DeviceProductRepository interface {
	Create(ctx context.Context, dp *model.DeviceProduct) error
	CreateBatch(ctx context.Context, dps []model.DeviceProduct) error
	GetByDeviceID(ctx context.Context, deviceID uuid.UUID) ([]model.DeviceProduct, error)
	GetByProductID(ctx context.Context, productID uuid.UUID) ([]model.DeviceProduct, error)
	Delete(ctx context.Context, deviceID, productID uuid.UUID) error
	DeleteByDeviceID(ctx context.Context, deviceID uuid.UUID) error
	ReplaceForDevice(ctx context.Context, deviceID uuid.UUID, productIDs []uuid.UUID) error
}

type deviceProductRepo struct {
	db *gorm.DB
}

// NewDeviceProductRepository creates a new DeviceProductRepository.
func NewDeviceProductRepository(db *gorm.DB) DeviceProductRepository {
	return &deviceProductRepo{db: db}
}

func (r *deviceProductRepo) Create(ctx context.Context, dp *model.DeviceProduct) error {
	return translateError(r.db.WithContext(ctx).Create(dp).Error)
}

func (r *deviceProductRepo) CreateBatch(ctx context.Context, dps []model.DeviceProduct) error {
	if len(dps) == 0 {
		return nil
	}
	return translateError(r.db.WithContext(ctx).Create(&dps).Error)
}

func (r *deviceProductRepo) GetByDeviceID(ctx context.Context, deviceID uuid.UUID) ([]model.DeviceProduct, error) {
	var dps []model.DeviceProduct
	err := r.db.WithContext(ctx).
		Preload("Product").
		Where("device_id = ?", deviceID).
		Find(&dps).Error
	return dps, translateError(err)
}

func (r *deviceProductRepo) GetByProductID(ctx context.Context, productID uuid.UUID) ([]model.DeviceProduct, error) {
	var dps []model.DeviceProduct
	err := r.db.WithContext(ctx).
		Preload("Device").
		Where("product_id = ?", productID).
		Find(&dps).Error
	return dps, translateError(err)
}

func (r *deviceProductRepo) Delete(ctx context.Context, deviceID, productID uuid.UUID) error {
	return translateError(r.db.WithContext(ctx).Delete(&model.DeviceProduct{}, "device_id = ? AND product_id = ?", deviceID, productID).Error)
}

func (r *deviceProductRepo) DeleteByDeviceID(ctx context.Context, deviceID uuid.UUID) error {
	return translateError(r.db.WithContext(ctx).Delete(&model.DeviceProduct{}, "device_id = ?", deviceID).Error)
}

func (r *deviceProductRepo) ReplaceForDevice(ctx context.Context, deviceID uuid.UUID, productIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete existing assignments
		if err := tx.Delete(&model.DeviceProduct{}, "device_id = ?", deviceID).Error; err != nil {
			return err
		}
		// Create new assignments
		if len(productIDs) == 0 {
			return nil
		}
		dps := make([]model.DeviceProduct, len(productIDs))
		for i, productID := range productIDs {
			dps[i] = model.DeviceProduct{
				DeviceID:  deviceID,
				ProductID: productID,
			}
		}
		return tx.Create(&dps).Error
	})
}

package postgres

import (
	"context"

	"backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DeviceRepository defines the interface for device data access.
type DeviceRepository interface {
	Create(ctx context.Context, device *model.Device) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Device, error)
	GetByDeviceKey(ctx context.Context, deviceKey string) (*model.Device, error)
	GetAll(ctx context.Context) ([]model.Device, error)
	GetByType(ctx context.Context, deviceType model.DeviceType) ([]model.Device, error)
	GetByStatus(ctx context.Context, status model.CommonStatus) ([]model.Device, error)
	GetApproved(ctx context.Context) ([]model.Device, error)
	Update(ctx context.Context, device *model.Device) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type deviceRepo struct {
	db *gorm.DB
}

// NewDeviceRepository creates a new DeviceRepository.
func NewDeviceRepository(db *gorm.DB) DeviceRepository {
	return &deviceRepo{db: db}
}

func (r *deviceRepo) Create(ctx context.Context, device *model.Device) error {
	return translateError(r.db.WithContext(ctx).Create(device).Error)
}

func (r *deviceRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Device, error) {
	var device model.Device
	err := r.db.WithContext(ctx).First(&device, "id = ?", id).Error
	return &device, translateError(err)
}

func (r *deviceRepo) GetByDeviceKey(ctx context.Context, deviceKey string) (*model.Device, error) {
	var device model.Device
	err := r.db.WithContext(ctx).First(&device, "device_key = ?", deviceKey).Error
	return &device, translateError(err)
}

func (r *deviceRepo) GetAll(ctx context.Context) ([]model.Device, error) {
	var devices []model.Device
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&devices).Error
	return devices, translateError(err)
}

func (r *deviceRepo) GetByType(ctx context.Context, deviceType model.DeviceType) ([]model.Device, error) {
	var devices []model.Device
	err := r.db.WithContext(ctx).Where("type = ?", deviceType).Order("name ASC").Find(&devices).Error
	return devices, translateError(err)
}

func (r *deviceRepo) GetByStatus(ctx context.Context, status model.CommonStatus) ([]model.Device, error) {
	var devices []model.Device
	err := r.db.WithContext(ctx).Where("status = ?", status).Order("created_at DESC").Find(&devices).Error
	return devices, translateError(err)
}

func (r *deviceRepo) GetApproved(ctx context.Context) ([]model.Device, error) {
	var devices []model.Device
	err := r.db.WithContext(ctx).
		Where("status = ?", model.CommonStatusApproved).
		Order("name ASC").
		Find(&devices).Error
	return devices, translateError(err)
}

func (r *deviceRepo) Update(ctx context.Context, device *model.Device) error {
	return translateError(r.db.WithContext(ctx).Save(device).Error)
}

func (r *deviceRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return translateError(r.db.WithContext(ctx).Delete(&model.Device{}, "id = ?", id).Error)
}

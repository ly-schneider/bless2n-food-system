package repository

import (
	"context"

	"backend/internal/generated/ent"

	"github.com/google/uuid"
)

type DeviceProductRepository interface {
	Create(ctx context.Context, deviceID, productID uuid.UUID) (*ent.DeviceProduct, error)
	CreateBatch(ctx context.Context, deviceID uuid.UUID, productIDs []uuid.UUID) ([]*ent.DeviceProduct, error)
	GetByDeviceID(ctx context.Context, deviceID uuid.UUID) ([]*ent.DeviceProduct, error)
	GetByProductID(ctx context.Context, productID uuid.UUID) ([]*ent.DeviceProduct, error)
	Delete(ctx context.Context, deviceID, productID uuid.UUID) error
	DeleteByDeviceID(ctx context.Context, deviceID uuid.UUID) error
	ReplaceForDevice(ctx context.Context, deviceID uuid.UUID, productIDs []uuid.UUID) error
}

type deviceProductRepo struct {
	client *ent.Client
}

func NewDeviceProductRepository(client *ent.Client) DeviceProductRepository {
	return &deviceProductRepo{client: client}
}

func (r *deviceProductRepo) ec(ctx context.Context) *ent.Client {
	return ClientFromContext(ctx, r.client)
}

func (r *deviceProductRepo) Create(ctx context.Context, deviceID, productID uuid.UUID) (*ent.DeviceProduct, error) {
	created, err := r.ec(ctx).DeviceProduct.Create().
		SetDeviceID(deviceID).
		SetProductID(productID).
		Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return created, nil
}

func (r *deviceProductRepo) CreateBatch(ctx context.Context, deviceID uuid.UUID, productIDs []uuid.UUID) ([]*ent.DeviceProduct, error) {
	if len(productIDs) == 0 {
		return nil, nil
	}
	builders := make([]*ent.DeviceProductCreate, len(productIDs))
	for i, pid := range productIDs {
		builders[i] = r.ec(ctx).DeviceProduct.Create().
			SetDeviceID(deviceID).
			SetProductID(pid)
	}
	created, err := r.ec(ctx).DeviceProduct.CreateBulk(builders...).Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return created, nil
}

func (r *deviceProductRepo) GetByDeviceID(ctx context.Context, deviceID uuid.UUID) ([]*ent.DeviceProduct, error) {
	rows, err := r.ec(ctx).DeviceProduct.Query().
		Where(entDeviceProductDeviceID(deviceID)).
		WithProduct().
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

func (r *deviceProductRepo) GetByProductID(ctx context.Context, productID uuid.UUID) ([]*ent.DeviceProduct, error) {
	rows, err := r.ec(ctx).DeviceProduct.Query().
		Where(entDeviceProductProductID(productID)).
		WithDevice().
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

func (r *deviceProductRepo) Delete(ctx context.Context, deviceID, productID uuid.UUID) error {
	_, err := r.ec(ctx).DeviceProduct.Delete().
		Where(
			entDeviceProductDeviceID(deviceID),
			entDeviceProductProductID(productID),
		).
		Exec(ctx)
	return translateError(err)
}

func (r *deviceProductRepo) DeleteByDeviceID(ctx context.Context, deviceID uuid.UUID) error {
	_, err := r.ec(ctx).DeviceProduct.Delete().
		Where(entDeviceProductDeviceID(deviceID)).
		Exec(ctx)
	return translateError(err)
}

func (r *deviceProductRepo) ReplaceForDevice(ctx context.Context, deviceID uuid.UUID, productIDs []uuid.UUID) error {
	// Use a transaction via the TxManager pattern
	tx, err := r.ec(ctx).Tx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	// Delete existing assignments
	_, err = tx.DeviceProduct.Delete().
		Where(entDeviceProductDeviceID(deviceID)).
		Exec(ctx)
	if err != nil {
		return translateError(err)
	}

	// Create new assignments
	if len(productIDs) == 0 {
		return tx.Commit()
	}
	builders := make([]*ent.DeviceProductCreate, len(productIDs))
	for i, pid := range productIDs {
		builders[i] = tx.DeviceProduct.Create().
			SetDeviceID(deviceID).
			SetProductID(pid)
	}
	_, err = tx.DeviceProduct.CreateBulk(builders...).Save(ctx)
	if err != nil {
		return translateError(err)
	}

	return tx.Commit()
}

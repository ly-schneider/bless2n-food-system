package repository

import (
	"context"
	"crypto/rand"
	"errors"
	"time"

	"backend/internal/generated/ent"
	"backend/internal/generated/ent/device"

	"github.com/google/uuid"
)

type DeviceRepository interface {
	Create(ctx context.Context, name, deviceKey string, deviceType device.Type, status device.Status, model *string, os *string, decidedBy *string, decidedAt, expiresAt *time.Time, pendingSessionToken, pairingCode *string, pairingCodeExpiresAt *time.Time) (*ent.Device, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.Device, error)
	GetByDeviceKey(ctx context.Context, deviceKey string) (*ent.Device, error)
	GetAll(ctx context.Context) ([]*ent.Device, error)
	GetByType(ctx context.Context, deviceType device.Type) ([]*ent.Device, error)
	GetByStatus(ctx context.Context, status device.Status) ([]*ent.Device, error)
	GetApproved(ctx context.Context) ([]*ent.Device, error)
	Update(ctx context.Context, id uuid.UUID, name, deviceKey string, deviceType device.Type, status device.Status, model *string, os *string, decidedBy *string, decidedAt, expiresAt *time.Time, pendingSessionToken, pairingCode *string, pairingCodeExpiresAt *time.Time) (*ent.Device, error)
	Delete(ctx context.Context, id uuid.UUID) error
	UpsertPendingByDeviceKey(ctx context.Context, name, deviceModel, os, deviceKey string, deviceType device.Type) (*ent.Device, error)
	ListProductIDsByDevice(ctx context.Context, deviceID uuid.UUID) ([]uuid.UUID, error)
	SetPendingSessionToken(ctx context.Context, deviceID uuid.UUID, token string) error
	ClearPendingSessionToken(ctx context.Context, deviceID uuid.UUID) error
	GeneratePairingCode(ctx context.Context, name, deviceModel, os, deviceKey string, deviceType device.Type) (*ent.Device, error)
	GetByPairingCode(ctx context.Context, code string) (*ent.Device, error)
	ClearPairingCode(ctx context.Context, deviceID uuid.UUID) error
}

type deviceRepo struct {
	client *ent.Client
}

func NewDeviceRepository(client *ent.Client) DeviceRepository {
	return &deviceRepo{client: client}
}

func (r *deviceRepo) ec(ctx context.Context) *ent.Client {
	return ClientFromContext(ctx, r.client)
}

func (r *deviceRepo) Create(ctx context.Context, name, deviceKey string, deviceType device.Type, status device.Status, deviceModel *string, os *string, decidedBy *string, decidedAt, expiresAt *time.Time, pendingSessionToken, pairingCode *string, pairingCodeExpiresAt *time.Time) (*ent.Device, error) {
	builder := r.ec(ctx).Device.Create().
		SetName(name).
		SetDeviceKey(deviceKey).
		SetType(deviceType).
		SetStatus(status)
	if deviceModel != nil {
		builder.SetModel(*deviceModel)
	}
	if os != nil {
		builder.SetOs(*os)
	}
	if decidedBy != nil {
		builder.SetDecidedBy(*decidedBy)
	}
	if decidedAt != nil {
		builder.SetDecidedAt(*decidedAt)
	}
	if expiresAt != nil {
		builder.SetExpiresAt(*expiresAt)
	}
	if pendingSessionToken != nil {
		builder.SetPendingSessionToken(*pendingSessionToken)
	}
	if pairingCode != nil {
		builder.SetPairingCode(*pairingCode)
	}
	if pairingCodeExpiresAt != nil {
		builder.SetPairingCodeExpiresAt(*pairingCodeExpiresAt)
	}
	created, err := builder.Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return created, nil
}

func (r *deviceRepo) GetByID(ctx context.Context, id uuid.UUID) (*ent.Device, error) {
	e, err := r.ec(ctx).Device.Get(ctx, id)
	if err != nil {
		return nil, translateError(err)
	}
	return e, nil
}

func (r *deviceRepo) GetByDeviceKey(ctx context.Context, deviceKey string) (*ent.Device, error) {
	e, err := r.ec(ctx).Device.Query().
		Where(device.DeviceKeyEQ(deviceKey)).
		Only(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return e, nil
}

func (r *deviceRepo) GetAll(ctx context.Context) ([]*ent.Device, error) {
	rows, err := r.ec(ctx).Device.Query().
		Order(device.ByCreatedAt(entDescOpt())).
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

func (r *deviceRepo) GetByType(ctx context.Context, deviceType device.Type) ([]*ent.Device, error) {
	rows, err := r.ec(ctx).Device.Query().
		Where(device.TypeEQ(deviceType)).
		Order(device.ByName()).
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

func (r *deviceRepo) GetByStatus(ctx context.Context, status device.Status) ([]*ent.Device, error) {
	rows, err := r.ec(ctx).Device.Query().
		Where(device.StatusEQ(status)).
		Order(device.ByCreatedAt(entDescOpt())).
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

func (r *deviceRepo) GetApproved(ctx context.Context) ([]*ent.Device, error) {
	rows, err := r.ec(ctx).Device.Query().
		Where(device.StatusEQ(device.StatusApproved)).
		Order(device.ByName()).
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

func (r *deviceRepo) Update(ctx context.Context, id uuid.UUID, name, deviceKey string, deviceType device.Type, status device.Status, deviceModel *string, os *string, decidedBy *string, decidedAt, expiresAt *time.Time, pendingSessionToken, pairingCode *string, pairingCodeExpiresAt *time.Time) (*ent.Device, error) {
	builder := r.ec(ctx).Device.UpdateOneID(id).
		SetName(name).
		SetDeviceKey(deviceKey).
		SetType(deviceType).
		SetStatus(status)
	if deviceModel != nil {
		builder.SetModel(*deviceModel)
	} else {
		builder.ClearModel()
	}
	if os != nil {
		builder.SetOs(*os)
	} else {
		builder.ClearOs()
	}
	if decidedBy != nil {
		builder.SetDecidedBy(*decidedBy)
	} else {
		builder.ClearDecidedBy()
	}
	if decidedAt != nil {
		builder.SetDecidedAt(*decidedAt)
	} else {
		builder.ClearDecidedAt()
	}
	if expiresAt != nil {
		builder.SetExpiresAt(*expiresAt)
	} else {
		builder.ClearExpiresAt()
	}
	if pendingSessionToken != nil {
		builder.SetPendingSessionToken(*pendingSessionToken)
	} else {
		builder.ClearPendingSessionToken()
	}
	if pairingCode != nil {
		builder.SetPairingCode(*pairingCode)
	} else {
		builder.ClearPairingCode()
	}
	if pairingCodeExpiresAt != nil {
		builder.SetPairingCodeExpiresAt(*pairingCodeExpiresAt)
	} else {
		builder.ClearPairingCodeExpiresAt()
	}
	updated, err := builder.Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return updated, nil
}

func (r *deviceRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return translateError(r.ec(ctx).Device.DeleteOneID(id).Exec(ctx))
}

func (r *deviceRepo) UpsertPendingByDeviceKey(ctx context.Context, name, deviceModel, os, deviceKey string, deviceType device.Type) (*ent.Device, error) {
	// Check if device exists
	existing, err := r.ec(ctx).Device.Query().
		Where(device.DeviceKeyEQ(deviceKey)).
		Only(ctx)

	if err == nil {
		// Update existing device
		builder := r.ec(ctx).Device.UpdateOneID(existing.ID).
			SetName(name)
		if deviceModel != "" {
			builder.SetModel(deviceModel)
		}
		if os != "" {
			builder.SetOs(os)
		}
		// Reset status to pending if previously rejected
		if existing.Status == device.StatusRejected {
			builder.SetStatus(device.StatusPending).
				ClearDecidedAt().
				ClearDecidedBy()
		}
		updated, err := builder.Save(ctx)
		if err != nil {
			return nil, translateError(err)
		}
		return updated, nil
	}

	if !ent.IsNotFound(err) {
		return nil, translateError(err)
	}

	// Create new device
	createBuilder := r.ec(ctx).Device.Create().
		SetName(name).
		SetDeviceKey(deviceKey).
		SetType(deviceType).
		SetStatus(device.StatusPending)
	if deviceModel != "" {
		createBuilder.SetModel(deviceModel)
	}
	if os != "" {
		createBuilder.SetOs(os)
	}
	created, err := createBuilder.Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return created, nil
}

func (r *deviceRepo) ListProductIDsByDevice(ctx context.Context, deviceID uuid.UUID) ([]uuid.UUID, error) {
	dps, err := r.ec(ctx).DeviceProduct.Query().
		Where(entDeviceProductDeviceID(deviceID)).
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	ids := make([]uuid.UUID, len(dps))
	for i, dp := range dps {
		ids[i] = dp.ProductID
	}
	return ids, nil
}

func (r *deviceRepo) SetPendingSessionToken(ctx context.Context, deviceID uuid.UUID, token string) error {
	n, err := r.ec(ctx).Device.Update().
		Where(device.ID(deviceID)).
		SetPendingSessionToken(token).
		Save(ctx)
	if err != nil {
		return translateError(err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *deviceRepo) ClearPendingSessionToken(ctx context.Context, deviceID uuid.UUID) error {
	n, err := r.ec(ctx).Device.Update().
		Where(device.ID(deviceID)).
		ClearPendingSessionToken().
		Save(ctx)
	if err != nil {
		return translateError(err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// pairingCodeCharset excludes ambiguous characters I, O, 0, 1.
const pairingCodeCharset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func generateCode() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = pairingCodeCharset[int(b[i])%len(pairingCodeCharset)]
	}
	return string(b), nil
}

func (r *deviceRepo) GeneratePairingCode(ctx context.Context, name, deviceModel, os, deviceKey string, deviceType device.Type) (*ent.Device, error) {
	// Upsert the device first
	d, err := r.UpsertPendingByDeviceKey(ctx, name, deviceModel, os, deviceKey, deviceType)
	if err != nil {
		return nil, err
	}

	// Generate a unique pairing code with retry on collision (up to 5 attempts)
	expiry := time.Now().UTC().Add(5 * time.Minute)
	for attempt := 0; attempt < 5; attempt++ {
		code, err := generateCode()
		if err != nil {
			return nil, err
		}
		updated, updateErr := r.ec(ctx).Device.UpdateOneID(d.ID).
			SetPairingCode(code).
			SetPairingCodeExpiresAt(expiry).
			Save(ctx)
		if updateErr != nil {
			// Unique constraint violation - retry with a new code
			if ent.IsConstraintError(updateErr) {
				continue
			}
			return nil, translateError(updateErr)
		}
		return updated, nil
	}
	return nil, errors.New("failed to generate unique pairing code after 5 attempts")
}

func (r *deviceRepo) GetByPairingCode(ctx context.Context, code string) (*ent.Device, error) {
	e, err := r.ec(ctx).Device.Query().
		Where(
			device.PairingCodeEQ(code),
			device.PairingCodeExpiresAtGT(time.Now().UTC()),
		).
		Only(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return e, nil
}

func (r *deviceRepo) ClearPairingCode(ctx context.Context, deviceID uuid.UUID) error {
	_, err := r.ec(ctx).Device.UpdateOneID(deviceID).
		ClearPairingCode().
		ClearPairingCodeExpiresAt().
		Save(ctx)
	return translateError(err)
}

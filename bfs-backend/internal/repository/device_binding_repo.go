package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"backend/internal/generated/ent"
	"backend/internal/generated/ent/devicebinding"

	"github.com/google/uuid"
)

type DeviceBindingRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*ent.DeviceBinding, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (*ent.DeviceBinding, error)
	Create(ctx context.Context, deviceType devicebinding.DeviceType, tokenHash, createdByUserID string, name *string, deviceID, stationID *uuid.UUID) (*ent.DeviceBinding, error)
	UpdateLastSeen(ctx context.Context, id uuid.UUID) error
	Revoke(ctx context.Context, id uuid.UUID) error
	ListActive(ctx context.Context) ([]*ent.DeviceBinding, error)
	ListByType(ctx context.Context, deviceType devicebinding.DeviceType) ([]*ent.DeviceBinding, error)
}

type deviceBindingRepo struct {
	client *ent.Client
}

func NewDeviceBindingRepository(client *ent.Client) DeviceBindingRepository {
	return &deviceBindingRepo{client: client}
}

func (r *deviceBindingRepo) ec(ctx context.Context) *ent.Client {
	return ClientFromContext(ctx, r.client)
}

func (r *deviceBindingRepo) GetByID(ctx context.Context, id uuid.UUID) (*ent.DeviceBinding, error) {
	e, err := r.ec(ctx).DeviceBinding.Get(ctx, id)
	if err != nil {
		return nil, translateError(err)
	}
	return e, nil
}

func (r *deviceBindingRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*ent.DeviceBinding, error) {
	e, err := r.ec(ctx).DeviceBinding.Query().
		Where(devicebinding.TokenHashEQ(tokenHash), devicebinding.RevokedAtIsNil()).
		Only(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return e, nil
}

func (r *deviceBindingRepo) Create(ctx context.Context, deviceType devicebinding.DeviceType, tokenHash, createdByUserID string, name *string, deviceID, stationID *uuid.UUID) (*ent.DeviceBinding, error) {
	builder := r.ec(ctx).DeviceBinding.Create().
		SetDeviceType(deviceType).
		SetTokenHash(tokenHash).
		SetCreatedByUserID(createdByUserID).
		SetLastSeenAt(time.Now().UTC())

	if name != nil {
		builder = builder.SetName(*name)
	}
	if deviceID != nil {
		builder = builder.SetDeviceID(*deviceID)
	}
	if stationID != nil {
		builder = builder.SetStationID(*stationID)
	}

	created, err := builder.Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return created, nil
}

func (r *deviceBindingRepo) UpdateLastSeen(ctx context.Context, id uuid.UUID) error {
	n, err := r.ec(ctx).DeviceBinding.Update().
		Where(devicebinding.IDEQ(id)).
		SetLastSeenAt(time.Now().UTC()).
		Save(ctx)
	if err != nil {
		return translateError(err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *deviceBindingRepo) Revoke(ctx context.Context, id uuid.UUID) error {
	now := time.Now().UTC()
	n, err := r.ec(ctx).DeviceBinding.Update().
		Where(devicebinding.IDEQ(id), devicebinding.RevokedAtIsNil()).
		SetRevokedAt(now).
		Save(ctx)
	if err != nil {
		return translateError(err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *deviceBindingRepo) ListActive(ctx context.Context) ([]*ent.DeviceBinding, error) {
	rows, err := r.ec(ctx).DeviceBinding.Query().
		Where(devicebinding.RevokedAtIsNil()).
		Order(devicebinding.ByCreatedAt(entDescOpt())).
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

func (r *deviceBindingRepo) ListByType(ctx context.Context, deviceType devicebinding.DeviceType) ([]*ent.DeviceBinding, error) {
	rows, err := r.ec(ctx).DeviceBinding.Query().
		Where(
			devicebinding.DeviceTypeEQ(deviceType),
			devicebinding.RevokedAtIsNil(),
		).
		Order(devicebinding.ByCreatedAt(entDescOpt())).
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

// HashToken computes a SHA-256 hash of the given token string.
func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

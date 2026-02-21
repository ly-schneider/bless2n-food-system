package repository

import (
	"context"

	"backend/internal/generated/ent"
	"backend/internal/generated/ent/club100freeproduct"
	"backend/internal/generated/ent/settings"

	"github.com/google/uuid"
)

type SettingsRepository interface {
	Get(ctx context.Context) (*ent.Settings, error)
	GetWithProducts(ctx context.Context) (*ent.Settings, error)
	Upsert(ctx context.Context, mode settings.PosMode) error
	UpdateClub100Settings(ctx context.Context, freeProductIDs []uuid.UUID, maxRedemptions int) error
	IsSystemEnabled(ctx context.Context) (bool, error)
	SetSystemEnabled(ctx context.Context, enabled bool) error
}

type settingsRepo struct {
	client *ent.Client
}

func NewSettingsRepository(client *ent.Client) SettingsRepository {
	return &settingsRepo{client: client}
}

func (r *settingsRepo) ec(ctx context.Context) *ent.Client {
	return ClientFromContext(ctx, r.client)
}

func (r *settingsRepo) Get(ctx context.Context) (*ent.Settings, error) {
	e, err := r.ec(ctx).Settings.Get(ctx, "default")
	if err != nil {
		if translateError(err) == ErrNotFound {
			return &ent.Settings{
				ID:                    "default",
				PosMode:               settings.PosModeQR_CODE,
				SystemEnabled:         true,
				Club100MaxRedemptions: 2,
			}, nil
		}
		return nil, translateError(err)
	}
	return e, nil
}

func (r *settingsRepo) GetWithProducts(ctx context.Context) (*ent.Settings, error) {
	e, err := r.ec(ctx).Settings.Query().
		Where(settings.IDEQ("default")).
		WithClub100FreeProducts().
		Only(ctx)
	if err != nil {
		if translateError(err) == ErrNotFound {
			return &ent.Settings{
				ID:                    "default",
				PosMode:               settings.PosModeQR_CODE,
				SystemEnabled:         true,
				Club100MaxRedemptions: 2,
			}, nil
		}
		return nil, translateError(err)
	}
	return e, nil
}

func (r *settingsRepo) Upsert(ctx context.Context, mode settings.PosMode) error {
	err := r.ec(ctx).Settings.Create().
		SetID("default").
		SetPosMode(mode).
		OnConflictColumns(settings.FieldID).
		SetPosMode(mode).
		Exec(ctx)
	return translateError(err)
}

func (r *settingsRepo) UpdateClub100Settings(ctx context.Context, freeProductIDs []uuid.UUID, maxRedemptions int) error {
	client := r.ec(ctx)

	err := client.Settings.Create().
		SetID("default").
		SetClub100MaxRedemptions(maxRedemptions).
		OnConflictColumns(settings.FieldID).
		SetClub100MaxRedemptions(maxRedemptions).
		Exec(ctx)
	if err != nil {
		return translateError(err)
	}

	_, err = client.Club100FreeProduct.Delete().
		Where(club100freeproduct.SettingsIDEQ("default")).
		Exec(ctx)
	if err != nil {
		return translateError(err)
	}

	for _, productID := range freeProductIDs {
		err = client.Club100FreeProduct.Create().
			SetSettingsID("default").
			SetProductID(productID).
			Exec(ctx)
		if err != nil {
			return translateError(err)
		}
	}

	return nil
}

func (r *settingsRepo) IsSystemEnabled(ctx context.Context) (bool, error) {
	s, err := r.Get(ctx)
	if err != nil {
		return true, err
	}
	return s.SystemEnabled, nil
}

func (r *settingsRepo) SetSystemEnabled(ctx context.Context, enabled bool) error {
	err := r.ec(ctx).Settings.Create().
		SetID("default").
		SetSystemEnabled(enabled).
		OnConflictColumns(settings.FieldID).
		SetSystemEnabled(enabled).
		Exec(ctx)
	return translateError(err)
}

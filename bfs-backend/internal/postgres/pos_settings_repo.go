package postgres

import (
	"context"

	"backend/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// PosSettingsRepository defines the interface for POS settings data access.
type PosSettingsRepository interface {
	Get(ctx context.Context) (*model.PosSettings, error)
	Upsert(ctx context.Context, settings *model.PosSettings) error
}

type posSettingsRepo struct {
	db *gorm.DB
}

// NewPosSettingsRepository creates a new PosSettingsRepository.
func NewPosSettingsRepository(db *gorm.DB) PosSettingsRepository {
	return &posSettingsRepo{db: db}
}

func (r *posSettingsRepo) Get(ctx context.Context) (*model.PosSettings, error) {
	var settings model.PosSettings
	err := r.db.WithContext(ctx).First(&settings, "id = ?", "default").Error
	if err != nil {
		if translateError(err) == ErrNotFound {
			// Return default settings if not found
			return &model.PosSettings{
				ID:   "default",
				Mode: model.PosFulfillmentModeJeton,
			}, nil
		}
		return nil, translateError(err)
	}
	return &settings, nil
}

func (r *posSettingsRepo) Upsert(ctx context.Context, settings *model.PosSettings) error {
	settings.ID = "default"
	return translateError(r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"mode", "updated_at"}),
	}).Create(settings).Error)
}

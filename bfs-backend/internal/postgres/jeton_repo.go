package postgres

import (
	"context"

	"backend/internal/model"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// JetonRepository defines the interface for jeton data access.
type JetonRepository interface {
	Create(ctx context.Context, jeton *model.Jeton) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Jeton, error)
	GetAll(ctx context.Context) ([]model.Jeton, error)
	Update(ctx context.Context, jeton *model.Jeton) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type jetonRepo struct {
	db *gorm.DB
}

// NewJetonRepository creates a new JetonRepository.
func NewJetonRepository(db *gorm.DB) JetonRepository {
	return &jetonRepo{db: db}
}

func (r *jetonRepo) Create(ctx context.Context, jeton *model.Jeton) error {
	return translateError(r.db.WithContext(ctx).Create(jeton).Error)
}

func (r *jetonRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Jeton, error) {
	var jeton model.Jeton
	err := r.db.WithContext(ctx).First(&jeton, "id = ?", id).Error
	return &jeton, translateError(err)
}

func (r *jetonRepo) GetAll(ctx context.Context) ([]model.Jeton, error) {
	var jetons []model.Jeton
	err := r.db.WithContext(ctx).Order("name ASC").Find(&jetons).Error
	return jetons, translateError(err)
}

func (r *jetonRepo) Update(ctx context.Context, jeton *model.Jeton) error {
	return translateError(r.db.WithContext(ctx).Save(jeton).Error)
}

func (r *jetonRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return translateError(r.db.WithContext(ctx).Delete(&model.Jeton{}, "id = ?", id).Error)
}

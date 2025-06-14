package repository

import (
	"backend/internal/domain"
	"context"

	"gorm.io/gorm"
)

type AuditLogRepository interface {
	List(ctx context.Context) ([]domain.AuditLog, error)
	Get(ctx context.Context, id string) (domain.AuditLog, error)
	Create(ctx context.Context, a *domain.AuditLog) error
	ListByUser(ctx context.Context, userID string) ([]domain.AuditLog, error)
}

type auditLogRepo struct{ db *gorm.DB }

func NewAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &auditLogRepo{db}
}

func (r *auditLogRepo) List(ctx context.Context) ([]domain.AuditLog, error) {
	var out []domain.AuditLog
	return out, r.db.WithContext(ctx).Preload("User").Find(&out).Error
}

func (r *auditLogRepo) Get(ctx context.Context, id string) (domain.AuditLog, error) {
	var a domain.AuditLog
	return a, r.db.WithContext(ctx).Preload("User").First(&a, "id = ?", id).Error
}

func (r *auditLogRepo) Create(ctx context.Context, a *domain.AuditLog) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *auditLogRepo) ListByUser(ctx context.Context, userID string) ([]domain.AuditLog, error) {
	var out []domain.AuditLog
	return out, r.db.WithContext(ctx).Preload("User").Where("user_id = ?", userID).Find(&out).Error
}

package repository

import (
	"backend/internal/domain"
	"context"
	"fmt"

	"gorm.io/gorm"
)

type AuditLogRepository interface {
	Create(ctx context.Context, auditLog *domain.AuditLog) error
}

type auditLogRepo struct{ db *gorm.DB }

func NewAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &auditLogRepo{db}
}

func (r *auditLogRepo) Create(ctx context.Context, auditLog *domain.AuditLog) error {
	if err := r.db.WithContext(ctx).Create(auditLog).Error; err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}
	return nil
}

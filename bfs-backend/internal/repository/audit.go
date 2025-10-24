package repository

import (
	"backend/internal/database"
	"backend/internal/domain"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type AuditRepository interface {
	Insert(ctx context.Context, log *domain.AuditLog) error
}

type auditRepository struct {
	collection *mongo.Collection
}

func NewAuditRepository(db *database.MongoDB) AuditRepository {
	return &auditRepository{collection: db.Database.Collection(database.AuditLogsCollection)}
}

func (r *auditRepository) Insert(ctx context.Context, log *domain.AuditLog) error {
	if log == nil {
		return nil
	}
	if log.ID.IsZero() {
		log.ID = bson.NewObjectID()
	}
	if log.CreatedAt.IsZero() {
		log.CreatedAt = time.Now().UTC()
	}
	_, err := r.collection.InsertOne(ctx, log)
	return err
}

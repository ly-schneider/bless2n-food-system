package repository

import (
	"context"
	"time"

	"backend/internal/database"
	"backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AuditLogRepository interface {
	Create(ctx context.Context, auditLog *domain.AuditLog) error
	GetByUserID(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*domain.AuditLog, error)
	GetByEvent(ctx context.Context, event domain.AuditEvent, limit, offset int) ([]*domain.AuditLog, error)
	List(ctx context.Context, limit, offset int) ([]*domain.AuditLog, error)
	DeleteOlderThan(ctx context.Context, cutoff time.Time) error
}

type auditLogRepository struct {
	collection *mongo.Collection
}

func NewAuditLogRepository(db *database.MongoDB) AuditLogRepository {
	return &auditLogRepository{
		collection: db.Database.Collection(database.AuditLogsCollection),
	}
}

func (r *auditLogRepository) Create(ctx context.Context, auditLog *domain.AuditLog) error {
	auditLog.ID = primitive.NewObjectID()
	auditLog.CreatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, auditLog)
	return err
}

func (r *auditLogRepository) GetByUserID(ctx context.Context, userID primitive.ObjectID, limit, offset int) ([]*domain.AuditLog, error) {
	filter := bson.M{"user_id": userID}
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []*domain.AuditLog
	for cursor.Next(ctx) {
		var log domain.AuditLog
		if err := cursor.Decode(&log); err != nil {
			return nil, err
		}
		logs = append(logs, &log)
	}

	return logs, cursor.Err()
}

func (r *auditLogRepository) GetByEvent(ctx context.Context, event domain.AuditEvent, limit, offset int) ([]*domain.AuditLog, error) {
	filter := bson.M{"event": event}
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []*domain.AuditLog
	for cursor.Next(ctx) {
		var log domain.AuditLog
		if err := cursor.Decode(&log); err != nil {
			return nil, err
		}
		logs = append(logs, &log)
	}

	return logs, cursor.Err()
}

func (r *auditLogRepository) List(ctx context.Context, limit, offset int) ([]*domain.AuditLog, error) {
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []*domain.AuditLog
	for cursor.Next(ctx) {
		var log domain.AuditLog
		if err := cursor.Decode(&log); err != nil {
			return nil, err
		}
		logs = append(logs, &log)
	}

	return logs, cursor.Err()
}

func (r *auditLogRepository) DeleteOlderThan(ctx context.Context, cutoff time.Time) error {
	filter := bson.M{"created_at": bson.M{"$lt": cutoff}}
	_, err := r.collection.DeleteMany(ctx, filter)
	return err
}
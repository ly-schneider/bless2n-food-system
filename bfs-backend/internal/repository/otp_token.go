package repository

import (
	"backend/internal/database"
	"backend/internal/domain"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type OTPTokenRepository interface {
	Create(ctx context.Context, t *domain.OTPToken) (*domain.OTPToken, error)
	CreateOTPCode(ctx context.Context, userID primitive.ObjectID, codeHash string, expiresAt time.Time) (*domain.OTPToken, error)
	FindActiveByUser(ctx context.Context, userID primitive.ObjectID) ([]domain.OTPToken, error)
	MarkUsed(ctx context.Context, id primitive.ObjectID, usedAt time.Time) error
	IncrementAttempts(ctx context.Context, id primitive.ObjectID) (int, error)
	DeleteByUser(ctx context.Context, userID primitive.ObjectID) error
}

type otpTokenRepository struct {
	collection *mongo.Collection
}

func NewOTPTokenRepository(db *database.MongoDB) OTPTokenRepository {
	return &otpTokenRepository{
		collection: db.Database.Collection(database.OTPTokensCollection),
	}
}

func (r *otpTokenRepository) Create(ctx context.Context, t *domain.OTPToken) (*domain.OTPToken, error) {
	if t.ID.IsZero() {
		t.ID = primitive.NewObjectID()
	}
	t.CreatedAt = time.Now().UTC()
	if _, err := r.collection.InsertOne(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (r *otpTokenRepository) CreateOTPCode(ctx context.Context, userID primitive.ObjectID, codeHash string, expiresAt time.Time) (*domain.OTPToken, error) {
	t := &domain.OTPToken{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		TokenHash: codeHash,
		CreatedAt: time.Now().UTC(),
		Attempts:  0,
		ExpiresAt: expiresAt,
	}
	_, err := r.collection.InsertOne(ctx, t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *otpTokenRepository) FindActiveByUser(ctx context.Context, userID primitive.ObjectID) ([]domain.OTPToken, error) {
	now := time.Now().UTC()
	cur, err := r.collection.Find(ctx, bson.M{
		"user_id":    userID,
		"used_at":    bson.M{"$exists": false},
		"expires_at": bson.M{"$gt": now},
	}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer func() { _ = cur.Close(ctx) }()
	var tokens []domain.OTPToken
	for cur.Next(ctx) {
		var t domain.OTPToken
		if err := cur.Decode(&t); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, nil
}

// FindActiveByTokenHash removed: OTP-only flow

func (r *otpTokenRepository) MarkUsed(ctx context.Context, id primitive.ObjectID, usedAt time.Time) error {
	_, err := r.collection.UpdateByID(ctx, id, bson.M{"$set": bson.M{"used_at": usedAt}})
	return err
}

func (r *otpTokenRepository) IncrementAttempts(ctx context.Context, id primitive.ObjectID) (int, error) {
	res := r.collection.FindOneAndUpdate(ctx, bson.M{"_id": id}, bson.M{"$inc": bson.M{"attempts": 1}}, options.FindOneAndUpdate().SetReturnDocument(options.After))
	var t domain.OTPToken
	if err := res.Decode(&t); err != nil {
		return 0, err
	}
	return t.Attempts, nil
}

func (r *otpTokenRepository) DeleteByUser(ctx context.Context, userID primitive.ObjectID) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"user_id": userID})
	return err
}

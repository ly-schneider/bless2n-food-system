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

type OTPTokenRepository interface {
	Create(ctx context.Context, token *domain.OTPToken) error
	GetLatestByUserAndType(ctx context.Context, userID primitive.ObjectID, tokenType domain.TokenType) (*domain.OTPToken, error)
	MarkAsUsed(ctx context.Context, id primitive.ObjectID) error
	IncrementAttempts(ctx context.Context, id primitive.ObjectID) error
	DeleteExpired(ctx context.Context) error
	DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error
}

type otpTokenRepository struct {
	collection *mongo.Collection
}

func NewOTPTokenRepository(db *database.MongoDB) OTPTokenRepository {
	return &otpTokenRepository{
		collection: db.Database.Collection(database.OTPTokensCollection),
	}
}

func (r *otpTokenRepository) Create(ctx context.Context, token *domain.OTPToken) error {
	token.ID = primitive.NewObjectID()
	token.CreatedAt = time.Now()
	token.Attempts = 0

	_, err := r.collection.InsertOne(ctx, token)
	return err
}

func (r *otpTokenRepository) GetLatestByUserAndType(ctx context.Context, userID primitive.ObjectID, tokenType domain.TokenType) (*domain.OTPToken, error) {
	filter := bson.M{
		"user_id": userID,
		"type":    tokenType,
		"used_at": nil,
		"expires_at": bson.M{"$gt": time.Now()},
	}

	opts := options.FindOne().SetSort(bson.D{{Key: "created_at", Value: -1}})

	var token domain.OTPToken
	err := r.collection.FindOne(ctx, filter, opts).Decode(&token)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &token, nil
}

func (r *otpTokenRepository) MarkAsUsed(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"used_at": time.Now()}}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *otpTokenRepository) IncrementAttempts(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$inc": bson.M{"attempts": 1}}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *otpTokenRepository) DeleteExpired(ctx context.Context) error {
	filter := bson.M{"expires_at": bson.M{"$lte": time.Now()}}
	_, err := r.collection.DeleteMany(ctx, filter)
	return err
}

func (r *otpTokenRepository) DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error {
	filter := bson.M{"user_id": userID}
	_, err := r.collection.DeleteMany(ctx, filter)
	return err
}
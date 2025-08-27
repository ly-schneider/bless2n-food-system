package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"backend/internal/database"
	"backend/internal/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	CreateWithPlainToken(ctx context.Context, token *domain.RefreshToken, plainToken string) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	GetValidTokenForUser(ctx context.Context, plainToken string) (*domain.RefreshToken, error)
	UpdateLastUsed(ctx context.Context, id primitive.ObjectID) error
	RevokeByFamilyID(ctx context.Context, familyID string, reason string) error
	RevokeByID(ctx context.Context, id primitive.ObjectID, reason string) error
	RevokeByHash(ctx context.Context, tokenHash string) error
	RevokeByClientID(ctx context.Context, userID primitive.ObjectID, clientID string, reason string) error
	DeleteExpired(ctx context.Context) error
	DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error
	GetActiveByUserID(ctx context.Context, userID primitive.ObjectID) ([]*domain.RefreshToken, error)
}

type refreshTokenRepository struct {
	collection *mongo.Collection
}

func NewRefreshTokenRepository(db *database.MongoDB) RefreshTokenRepository {
	return &refreshTokenRepository{
		collection: db.Database.Collection(database.RefreshTokensCollection),
	}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	token.ID = primitive.NewObjectID()
	token.IssuedAt = time.Now()
	token.LastUsedAt = time.Now()
	token.IsRevoked = false

	_, err := r.collection.InsertOne(ctx, token)
	return err
}

func (r *refreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	filter := bson.M{
		"token_hash": tokenHash,
		"is_revoked": false,
		"expires_at": bson.M{"$gt": time.Now()},
	}

	var token domain.RefreshToken
	err := r.collection.FindOne(ctx, filter).Decode(&token)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &token, nil
}

func (r *refreshTokenRepository) UpdateLastUsed(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"last_used_at": time.Now()}}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *refreshTokenRepository) RevokeByFamilyID(ctx context.Context, familyID string, reason string) error {
	filter := bson.M{"family_id": familyID}
	update := bson.M{
		"$set": bson.M{
			"is_revoked":      true,
			"revoked_reason":  reason,
		},
	}

	_, err := r.collection.UpdateMany(ctx, filter, update)
	return err
}

func (r *refreshTokenRepository) RevokeByID(ctx context.Context, id primitive.ObjectID, reason string) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"is_revoked":     true,
			"revoked_reason": reason,
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *refreshTokenRepository) DeleteExpired(ctx context.Context) error {
	filter := bson.M{"expires_at": bson.M{"$lte": time.Now()}}
	_, err := r.collection.DeleteMany(ctx, filter)
	return err
}

func (r *refreshTokenRepository) DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error {
	filter := bson.M{"user_id": userID}
	_, err := r.collection.DeleteMany(ctx, filter)
	return err
}

func (r *refreshTokenRepository) GetActiveByUserID(ctx context.Context, userID primitive.ObjectID) ([]*domain.RefreshToken, error) {
	filter := bson.M{
		"user_id":    userID,
		"is_revoked": false,
		"expires_at": bson.M{"$gt": time.Now()},
	}

	opts := options.Find().SetSort(bson.D{{Key: "last_used_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tokens []*domain.RefreshToken
	for cursor.Next(ctx) {
		var token domain.RefreshToken
		if err := cursor.Decode(&token); err != nil {
			return nil, err
		}
		tokens = append(tokens, &token)
	}

	return tokens, cursor.Err()
}

func (r *refreshTokenRepository) GetValidTokenForUser(ctx context.Context, plainToken string) (*domain.RefreshToken, error) {
	// Hash the plain token to compare with stored hashes
	tokenHash := hashRefreshToken(plainToken)
	
	filter := bson.M{
		"token_hash": tokenHash,
		"is_revoked": false,
		"expires_at": bson.M{"$gt": time.Now()},
	}

	var token domain.RefreshToken
	err := r.collection.FindOne(ctx, filter).Decode(&token)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, domain.ErrRefreshTokenNotFound
		}
		return nil, err
	}

	return &token, nil
}

func (r *refreshTokenRepository) RevokeByHash(ctx context.Context, tokenHash string) error {
	filter := bson.M{"token_hash": tokenHash}
	update := bson.M{
		"$set": bson.M{
			"is_revoked":     true,
			"revoked_reason": "token_rotation",
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return domain.ErrRefreshTokenNotFound
	}

	return nil
}

func (r *refreshTokenRepository) CreateWithPlainToken(ctx context.Context, token *domain.RefreshToken, plainToken string) error {
	token.ID = primitive.NewObjectID()
	token.IssuedAt = time.Now()
	token.LastUsedAt = time.Now()
	token.IsRevoked = false
	token.TokenHash = hashRefreshToken(plainToken)

	_, err := r.collection.InsertOne(ctx, token)
	return err
}

func (r *refreshTokenRepository) RevokeByClientID(ctx context.Context, userID primitive.ObjectID, clientID string, reason string) error {
	filter := bson.M{
		"user_id":    userID,
		"client_id":  clientID,
		"is_revoked": false,
	}
	update := bson.M{
		"$set": bson.M{
			"is_revoked":     true,
			"revoked_reason": reason,
		},
	}

	_, err := r.collection.UpdateMany(ctx, filter, update)
	return err
}

func hashRefreshToken(plainToken string) string {
	hash := sha256.Sum256([]byte(plainToken))
	return hex.EncodeToString(hash[:])
}
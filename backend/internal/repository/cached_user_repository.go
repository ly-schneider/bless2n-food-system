package repository

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"backend/internal/domain"
	"backend/internal/model"
	"backend/internal/redis"
)

type CachedUserRepository struct {
	repo  UserRepository
	cache *redis.CacheService
	logger *zap.Logger
}

func NewCachedUserRepository(repo UserRepository, cache *redis.CacheService, logger *zap.Logger) UserRepository {
	return &CachedUserRepository{
		repo:   repo,
		cache:  cache,
		logger: logger,
	}
}

func (r *CachedUserRepository) GetByID(ctx context.Context, id model.NanoID14) (*domain.User, error) {
	cacheKey := r.userIDKey(string(id))
	
	var user domain.User
	if err := r.cache.Get(ctx, cacheKey, &user); err == nil {
		r.logger.Debug("user cache hit", zap.String("user_id", string(id)))
		return &user, nil
	} else if err != redis.ErrCacheMiss {
		r.logger.Error("cache error getting user by ID", 
			zap.String("user_id", string(id)), 
			zap.Error(err))
	}

	user_ptr, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := r.cache.Set(ctx, cacheKey, user_ptr, 15*time.Minute); err != nil {
		r.logger.Error("failed to cache user by ID",
			zap.String("user_id", string(id)),
			zap.Error(err))
	}

	r.logger.Debug("user fetched from database and cached", zap.String("user_id", string(id)))
	return user_ptr, nil
}

func (r *CachedUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	cacheKey := r.userEmailKey(email)
	
	var user domain.User
	if err := r.cache.Get(ctx, cacheKey, &user); err == nil {
		r.logger.Debug("user cache hit by email", zap.String("email", email))
		return &user, nil
	} else if err != redis.ErrCacheMiss {
		r.logger.Error("cache error getting user by email", 
			zap.String("email", email), 
			zap.Error(err))
	}

	user_ptr, err := r.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if err := r.cache.Set(ctx, cacheKey, user_ptr, 15*time.Minute); err != nil {
		r.logger.Error("failed to cache user by email",
			zap.String("email", email),
			zap.Error(err))
	}

	userIDKey := r.userIDKey(string(user_ptr.ID))
	if err := r.cache.Set(ctx, userIDKey, user_ptr, 15*time.Minute); err != nil {
		r.logger.Error("failed to cache user by ID during email lookup",
			zap.String("user_id", string(user_ptr.ID)),
			zap.Error(err))
	}

	r.logger.Debug("user fetched from database and cached", 
		zap.String("email", email),
		zap.String("user_id", string(user_ptr.ID)))
	return user_ptr, nil
}

func (r *CachedUserRepository) Create(ctx context.Context, user *domain.User) error {
	if err := r.repo.Create(ctx, user); err != nil {
		return err
	}

	userIDKey := r.userIDKey(string(user.ID))
	userEmailKey := r.userEmailKey(user.Email)

	if err := r.cache.Set(ctx, userIDKey, user, 15*time.Minute); err != nil {
		r.logger.Error("failed to cache newly created user by ID",
			zap.String("user_id", string(user.ID)),
			zap.Error(err))
	}

	if err := r.cache.Set(ctx, userEmailKey, user, 15*time.Minute); err != nil {
		r.logger.Error("failed to cache newly created user by email",
			zap.String("email", user.Email),
			zap.Error(err))
	}

	r.logger.Debug("user created and cached", 
		zap.String("user_id", string(user.ID)),
		zap.String("email", user.Email))

	return nil
}

func (r *CachedUserRepository) InvalidateUser(ctx context.Context, userID model.NanoID14, email string) error {
	keys := []string{
		r.userIDKey(string(userID)),
		r.userEmailKey(email),
	}

	if err := r.cache.Delete(ctx, keys...); err != nil {
		r.logger.Error("failed to invalidate user cache",
			zap.String("user_id", string(userID)),
			zap.String("email", email),
			zap.Error(err))
		return err
	}

	r.logger.Debug("user cache invalidated", 
		zap.String("user_id", string(userID)),
		zap.String("email", email))

	return nil
}

func (r *CachedUserRepository) userIDKey(userID string) string {
	return fmt.Sprintf("user:id:%s", userID)
}

func (r *CachedUserRepository) userEmailKey(email string) string {
	return fmt.Sprintf("user:email:%s", email)
}
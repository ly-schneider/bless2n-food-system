package repository

import (
	"context"
	"fmt"
	"time"

	"backend/internal/domain"
	"backend/internal/model"
	"backend/internal/redis"

	"go.uber.org/zap"
)

type CachedUserRepository struct {
	base   UserRepository
	cache  *redis.CacheService
	logger *zap.Logger
}

// userCacheData represents the complete user data stored in Redis
type userCacheData struct {
	User         *domain.User `json:"user"`
	PasswordHash string       `json:"password_hash"`
}

func NewCachedUserRepository(base UserRepository, cache *redis.CacheService, logger *zap.Logger) UserRepository {
	return &CachedUserRepository{
		base:   base,
		cache:  cache,
		logger: logger,
	}
}

func (r *CachedUserRepository) GetByID(ctx context.Context, id model.NanoID14) (*domain.User, error) {
	cacheKey := fmt.Sprintf("user:id:%s", id)
	
	// Try to get from cache first
	var cacheData userCacheData
	if err := r.cache.Get(ctx, cacheKey, &cacheData); err == nil {
		// Restore password hash from cache data
		cacheData.User.PasswordHash = cacheData.PasswordHash
		r.logger.Debug("Cache hit for user by ID", zap.String("user_id", string(id)))
		return cacheData.User, nil
	}
	
	r.logger.Debug("Cache miss for user by ID", zap.String("user_id", string(id)))
	
	// Cache miss - get from database
	user, err := r.base.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	// Cache the user data
	r.cacheUser(ctx, user)
	
	return user, nil
}

func (r *CachedUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	cacheKey := fmt.Sprintf("user:email:%s", email)
	
	// Try to get from cache first
	var cacheData userCacheData
	if err := r.cache.Get(ctx, cacheKey, &cacheData); err == nil {
		// Restore password hash from cache data
		cacheData.User.PasswordHash = cacheData.PasswordHash
		r.logger.Debug("Cache hit for user by email", zap.String("email", email))
		return cacheData.User, nil
	}
	
	r.logger.Debug("Cache miss for user by email", zap.String("email", email))
	
	// Cache miss - get from database
	user, err := r.base.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	
	// Cache the user data
	r.cacheUser(ctx, user)
	
	return user, nil
}

func (r *CachedUserRepository) Create(ctx context.Context, user *domain.User) error {
	// Create in database first
	err := r.base.Create(ctx, user)
	if err != nil {
		return err
	}
	
	// Load the user with Role relationship to cache it properly
	userWithRole, err := r.base.GetByID(ctx, user.ID)
	if err != nil {
		r.logger.Error("Failed to load user with role after creation", zap.String("user_id", string(user.ID)), zap.Error(err))
		// Don't fail the create operation, just don't cache
		return nil
	}
	
	// Cache the newly created user with role loaded
	r.cacheUser(ctx, userWithRole)
	
	return nil
}

func (r *CachedUserRepository) Update(ctx context.Context, user *domain.User) error {
	// Update in database first
	err := r.base.Update(ctx, user)
	if err != nil {
		return err
	}
	
	// Invalidate cache for this user
	r.InvalidateUser(ctx, user)
	
	return nil
}

// cacheUser stores user data in Redis with password hash stored separately
func (r *CachedUserRepository) cacheUser(ctx context.Context, user *domain.User) {
	if user == nil {
		return
	}
	
	// Create cache data structure
	userCopy := *user
	userCopy.PasswordHash = "" // Clear password hash from user copy
	
	// Ensure Role relationship is preserved
	if user.Role != nil {
		roleCopy := *user.Role
		userCopy.Role = &roleCopy
	}
	
	cacheData := userCacheData{
		User:         &userCopy,
		PasswordHash: user.PasswordHash, // Store password hash separately
	}
	
	ttl := 15 * time.Minute
	
	// Cache by ID
	idKey := fmt.Sprintf("user:id:%s", user.ID)
	if err := r.cache.Set(ctx, idKey, cacheData, ttl); err != nil {
		r.logger.Error("Failed to cache user by ID", zap.String("user_id", string(user.ID)), zap.Error(err))
	}
	
	// Cache by email
	emailKey := fmt.Sprintf("user:email:%s", user.Email)
	if err := r.cache.Set(ctx, emailKey, cacheData, ttl); err != nil {
		r.logger.Error("Failed to cache user by email", zap.String("email", user.Email), zap.Error(err))
	}
	
	r.logger.Debug("Cached user data", zap.String("user_id", string(user.ID)), zap.String("email", user.Email))
}

// InvalidateUser removes user data from cache
func (r *CachedUserRepository) InvalidateUser(ctx context.Context, user *domain.User) {
	if user == nil {
		return
	}
	
	idKey := fmt.Sprintf("user:id:%s", user.ID)
	emailKey := fmt.Sprintf("user:email:%s", user.Email)
	
	if err := r.cache.Delete(ctx, idKey); err != nil {
		r.logger.Error("Failed to invalidate user cache by ID", zap.String("user_id", string(user.ID)), zap.Error(err))
	}
	
	if err := r.cache.Delete(ctx, emailKey); err != nil {
		r.logger.Error("Failed to invalidate user cache by email", zap.String("email", user.Email), zap.Error(err))
	}
	
	r.logger.Debug("Invalidated user cache", zap.String("user_id", string(user.ID)), zap.String("email", user.Email))
}
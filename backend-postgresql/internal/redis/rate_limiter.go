package redis

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type RateLimiterService struct {
	cache  *CacheService
	logger *zap.Logger
}

type RateLimit struct {
	Limit     int64         `json:"limit"`
	Window    time.Duration `json:"window"`
	Remaining int64         `json:"remaining"`
	ResetTime time.Time     `json:"reset_time"`
}

func NewRateLimiterService(cache *CacheService, logger *zap.Logger) *RateLimiterService {
	return &RateLimiterService{
		cache:  cache,
		logger: logger,
	}
}

func (r *RateLimiterService) CheckLimit(ctx context.Context, key string, limit int64, window time.Duration) (*RateLimit, error) {
	rateLimitKey := r.rateLimitKey(key)
	
	current, err := r.cache.IncrementWithExpire(ctx, rateLimitKey, window)
	if err != nil {
		r.logger.Error("failed to increment rate limit counter",
			zap.String("key", key),
			zap.Int64("limit", limit),
			zap.Duration("window", window),
			zap.Error(err))
		return nil, fmt.Errorf("increment rate limit: %w", err)
	}

	remaining := limit - current
	if remaining < 0 {
		remaining = 0
	}

	resetTime := time.Now().Add(window)
	if current > 1 {
		ttl, err := r.cache.GetTTL(ctx, rateLimitKey)
		if err == nil && ttl > 0 {
			resetTime = time.Now().Add(ttl)
		}
	}

	rateLimit := &RateLimit{
		Limit:     limit,
		Window:    window,
		Remaining: remaining,
		ResetTime: resetTime,
	}

	r.logger.Debug("rate limit checked",
		zap.String("key", key),
		zap.Int64("current", current),
		zap.Int64("limit", limit),
		zap.Int64("remaining", remaining),
		zap.Time("reset_time", resetTime))

	return rateLimit, nil
}

func (r *RateLimiterService) IsLimited(ctx context.Context, key string, limit int64, window time.Duration) (bool, *RateLimit, error) {
	rateLimit, err := r.CheckLimit(ctx, key, limit, window)
	if err != nil {
		return false, nil, err
	}

	isLimited := rateLimit.Remaining == 0
	
	if isLimited {
		r.logger.Warn("rate limit exceeded",
			zap.String("key", key),
			zap.Int64("limit", limit),
			zap.Duration("window", window))
	}

	return isLimited, rateLimit, nil
}

func (r *RateLimiterService) ResetLimit(ctx context.Context, key string) error {
	rateLimitKey := r.rateLimitKey(key)
	
	if err := r.cache.Delete(ctx, rateLimitKey); err != nil {
		r.logger.Error("failed to reset rate limit",
			zap.String("key", key),
			zap.Error(err))
		return fmt.Errorf("reset rate limit: %w", err)
	}

	r.logger.Debug("rate limit reset successfully", zap.String("key", key))
	return nil
}

func (r *RateLimiterService) GetCurrentCount(ctx context.Context, key string) (int64, error) {
	rateLimitKey := r.rateLimitKey(key)
	
	var count int64
	if err := r.cache.Get(ctx, rateLimitKey, &count); err != nil {
		if err == ErrCacheMiss {
			return 0, nil
		}
		return 0, fmt.Errorf("get current count: %w", err)
	}
	
	return count, nil
}

func (r *RateLimiterService) rateLimitKey(key string) string {
	return fmt.Sprintf("rate_limit:%s", key)
}

func (r *RateLimiterService) IPKey(ip string) string {
	return fmt.Sprintf("ip:%s", ip)
}

func (r *RateLimiterService) UserKey(userID string) string {
	return fmt.Sprintf("user:%s", userID)
}

func (r *RateLimiterService) EndpointKey(endpoint, method string) string {
	return fmt.Sprintf("endpoint:%s:%s", method, endpoint)
}

func (r *RateLimiterService) UserEndpointKey(userID, endpoint, method string) string {
	return fmt.Sprintf("user_endpoint:%s:%s:%s", userID, method, endpoint)
}
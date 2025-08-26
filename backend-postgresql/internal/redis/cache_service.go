package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type CacheService struct {
	client *Client
	logger *zap.Logger
}

func NewCacheService(client *Client, logger *zap.Logger) *CacheService {
	return &CacheService{
		client: client,
		logger: logger,
	}
}

func (c *CacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		c.logger.Error("failed to marshal value for cache", 
			zap.String("key", key), 
			zap.Error(err))
		return fmt.Errorf("marshal value: %w", err)
	}

	if err := c.client.rdb.Set(ctx, key, data, ttl).Err(); err != nil {
		c.logger.Error("failed to set cache", 
			zap.String("key", key), 
			zap.Duration("ttl", ttl),
			zap.Error(err))
		return fmt.Errorf("set cache: %w", err)
	}

	c.logger.Debug("cache set successfully", 
		zap.String("key", key), 
		zap.Duration("ttl", ttl))
	
	return nil
}

func (c *CacheService) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		c.logger.Error("failed to get from cache", 
			zap.String("key", key), 
			zap.Error(err))
		return fmt.Errorf("get cache: %w", err)
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		c.logger.Error("failed to unmarshal cached value", 
			zap.String("key", key), 
			zap.Error(err))
		return fmt.Errorf("unmarshal cached value: %w", err)
	}

	c.logger.Debug("cache hit", zap.String("key", key))
	return nil
}

func (c *CacheService) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	if err := c.client.rdb.Del(ctx, keys...).Err(); err != nil {
		c.logger.Error("failed to delete from cache", 
			zap.Strings("keys", keys), 
			zap.Error(err))
		return fmt.Errorf("delete cache: %w", err)
	}

	c.logger.Debug("cache deleted successfully", zap.Strings("keys", keys))
	return nil
}

func (c *CacheService) Exists(ctx context.Context, key string) (bool, error) {
	count, err := c.client.rdb.Exists(ctx, key).Result()
	if err != nil {
		c.logger.Error("failed to check cache existence", 
			zap.String("key", key), 
			zap.Error(err))
		return false, fmt.Errorf("check cache existence: %w", err)
	}
	return count > 0, nil
}

func (c *CacheService) SetExpire(ctx context.Context, key string, ttl time.Duration) error {
	if err := c.client.rdb.Expire(ctx, key, ttl).Err(); err != nil {
		c.logger.Error("failed to set cache expiration", 
			zap.String("key", key), 
			zap.Duration("ttl", ttl),
			zap.Error(err))
		return fmt.Errorf("set cache expiration: %w", err)
	}
	return nil
}

func (c *CacheService) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := c.client.rdb.TTL(ctx, key).Result()
	if err != nil {
		c.logger.Error("failed to get cache TTL", 
			zap.String("key", key), 
			zap.Error(err))
		return 0, fmt.Errorf("get cache TTL: %w", err)
	}
	return ttl, nil
}

func (c *CacheService) Increment(ctx context.Context, key string) (int64, error) {
	val, err := c.client.rdb.Incr(ctx, key).Result()
	if err != nil {
		c.logger.Error("failed to increment cache value", 
			zap.String("key", key), 
			zap.Error(err))
		return 0, fmt.Errorf("increment cache: %w", err)
	}
	return val, nil
}

func (c *CacheService) IncrementWithExpire(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	pipe := c.client.rdb.Pipeline()
	incrCmd := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, ttl)
	
	if _, err := pipe.Exec(ctx); err != nil {
		c.logger.Error("failed to increment with expire", 
			zap.String("key", key), 
			zap.Duration("ttl", ttl),
			zap.Error(err))
		return 0, fmt.Errorf("increment with expire: %w", err)
	}
	
	return incrCmd.Val(), nil
}

func (c *CacheService) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return false, fmt.Errorf("marshal value: %w", err)
	}

	set, err := c.client.rdb.SetNX(ctx, key, data, ttl).Result()
	if err != nil {
		c.logger.Error("failed to set cache if not exists", 
			zap.String("key", key), 
			zap.Error(err))
		return false, fmt.Errorf("set cache if not exists: %w", err)
	}

	return set, nil
}

func (c *CacheService) FlushPattern(ctx context.Context, pattern string) error {
	keys, err := c.client.rdb.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("get keys by pattern: %w", err)
	}

	if len(keys) > 0 {
		if err := c.client.rdb.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("delete keys: %w", err)
		}
		c.logger.Info("cache pattern flushed", 
			zap.String("pattern", pattern), 
			zap.Int("count", len(keys)))
	}

	return nil
}

var ErrCacheMiss = fmt.Errorf("cache miss")
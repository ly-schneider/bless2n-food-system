package redis

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type SessionData struct {
	UserID   string    `json:"user_id"`
	Role     string    `json:"role"`
	IssuedAt time.Time `json:"issued_at"`
	DeviceID string    `json:"device_id,omitempty"`
}

type SessionCacheService struct {
	cache  *CacheService
	logger *zap.Logger
}

func NewSessionCacheService(cache *CacheService, logger *zap.Logger) *SessionCacheService {
	return &SessionCacheService{
		cache:  cache,
		logger: logger,
	}
}

func (s *SessionCacheService) SetSession(ctx context.Context, jti string, sessionData *SessionData, ttl time.Duration) error {
	key := s.sessionKey(jti)
	
	if err := s.cache.Set(ctx, key, sessionData, ttl); err != nil {
		s.logger.Error("failed to set session cache",
			zap.String("jti", jti),
			zap.String("user_id", sessionData.UserID),
			zap.Error(err))
		return fmt.Errorf("set session cache: %w", err)
	}

	s.logger.Debug("session cached successfully",
		zap.String("jti", jti),
		zap.String("user_id", sessionData.UserID),
		zap.Duration("ttl", ttl))
	
	return nil
}

func (s *SessionCacheService) GetSession(ctx context.Context, jti string) (*SessionData, error) {
	key := s.sessionKey(jti)
	var sessionData SessionData
	
	if err := s.cache.Get(ctx, key, &sessionData); err != nil {
		if err == ErrCacheMiss {
			s.logger.Debug("session cache miss", zap.String("jti", jti))
			return nil, ErrSessionNotFound
		}
		s.logger.Error("failed to get session from cache",
			zap.String("jti", jti),
			zap.Error(err))
		return nil, fmt.Errorf("get session cache: %w", err)
	}

	s.logger.Debug("session cache hit", 
		zap.String("jti", jti),
		zap.String("user_id", sessionData.UserID))
	
	return &sessionData, nil
}

func (s *SessionCacheService) InvalidateSession(ctx context.Context, jti string) error {
	key := s.sessionKey(jti)
	
	if err := s.cache.Delete(ctx, key); err != nil {
		s.logger.Error("failed to invalidate session",
			zap.String("jti", jti),
			zap.Error(err))
		return fmt.Errorf("invalidate session: %w", err)
	}

	s.logger.Debug("session invalidated successfully", zap.String("jti", jti))
	return nil
}

func (s *SessionCacheService) InvalidateUserSessions(ctx context.Context, userID string) error {
	pattern := s.userSessionPattern(userID)
	
	if err := s.cache.FlushPattern(ctx, pattern); err != nil {
		s.logger.Error("failed to invalidate user sessions",
			zap.String("user_id", userID),
			zap.Error(err))
		return fmt.Errorf("invalidate user sessions: %w", err)
	}

	s.logger.Info("user sessions invalidated successfully", zap.String("user_id", userID))
	return nil
}

func (s *SessionCacheService) ExtendSession(ctx context.Context, jti string, ttl time.Duration) error {
	key := s.sessionKey(jti)
	
	exists, err := s.cache.Exists(ctx, key)
	if err != nil {
		return fmt.Errorf("check session existence: %w", err)
	}
	
	if !exists {
		return ErrSessionNotFound
	}
	
	if err := s.cache.SetExpire(ctx, key, ttl); err != nil {
		s.logger.Error("failed to extend session",
			zap.String("jti", jti),
			zap.Duration("ttl", ttl),
			zap.Error(err))
		return fmt.Errorf("extend session: %w", err)
	}

	s.logger.Debug("session extended successfully", 
		zap.String("jti", jti),
		zap.Duration("ttl", ttl))
	
	return nil
}

func (s *SessionCacheService) GetSessionTTL(ctx context.Context, jti string) (time.Duration, error) {
	key := s.sessionKey(jti)
	return s.cache.GetTTL(ctx, key)
}

func (s *SessionCacheService) IsSessionValid(ctx context.Context, jti string) (bool, error) {
	_, err := s.GetSession(ctx, jti)
	if err != nil {
		if err == ErrSessionNotFound {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *SessionCacheService) sessionKey(jti string) string {
	return fmt.Sprintf("session:%s", jti)
}

func (s *SessionCacheService) userSessionPattern(userID string) string {
	return fmt.Sprintf("session:*:user:%s", userID)
}

var ErrSessionNotFound = fmt.Errorf("session not found")
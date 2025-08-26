package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	"backend/internal/apperrors"
	"backend/internal/redis"
)

type RateLimitConfig struct {
	Limit  int64
	Window time.Duration
	KeyFn  func(*http.Request) string
}

func RateLimit(rateLimiter *redis.RateLimiterService, config RateLimitConfig, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := config.KeyFn(r)
			
			isLimited, rateLimit, err := rateLimiter.IsLimited(r.Context(), key, config.Limit, config.Window)
			if err != nil {
				logger.Error("rate limiter error",
					zap.String("key", key),
					zap.Error(err))
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(rateLimit.Limit, 10))
			w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(rateLimit.Remaining, 10))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(rateLimit.ResetTime.Unix(), 10))

			if isLimited {
				retryAfter := int(time.Until(rateLimit.ResetTime).Seconds())
				if retryAfter < 0 {
					retryAfter = 1
				}
				
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				w.Header().Set("Content-Type", "application/json")
				
				logger.Warn("rate limit exceeded",
					zap.String("key", key),
					zap.Int64("limit", config.Limit),
					zap.Duration("window", config.Window),
					zap.String("remote_addr", r.RemoteAddr))

				apiErr := apperrors.FromStatus(http.StatusTooManyRequests, "Too many requests. Please try again later.", 
					errors.New("rate limit exceeded"))
				
				w.WriteHeader(apiErr.Status)
				json.NewEncoder(w).Encode(apiErr)
				return
			}

			logger.Debug("rate limit check passed",
				zap.String("key", key),
				zap.Int64("remaining", rateLimit.Remaining))

			next.ServeHTTP(w, r)
		})
	}
}

func IPRateLimit(rateLimiter *redis.RateLimiterService, limit int64, window time.Duration, logger *zap.Logger) func(http.Handler) http.Handler {
	config := RateLimitConfig{
		Limit:  limit,
		Window: window,
		KeyFn: func(r *http.Request) string {
			return rateLimiter.IPKey(getClientIP(r))
		},
	}
	
	return RateLimit(rateLimiter, config, logger)
}

func UserRateLimit(rateLimiter *redis.RateLimiterService, limit int64, window time.Duration, logger *zap.Logger) func(http.Handler) http.Handler {
	config := RateLimitConfig{
		Limit:  limit,
		Window: window,
		KeyFn: func(r *http.Request) string {
			userID := r.Context().Value("user_id")
			if userID == nil {
				return rateLimiter.IPKey(getClientIP(r))
			}
			return rateLimiter.UserKey(userID.(string))
		},
	}
	
	return RateLimit(rateLimiter, config, logger)
}

func EndpointRateLimit(rateLimiter *redis.RateLimiterService, endpoint string, limit int64, window time.Duration, logger *zap.Logger) func(http.Handler) http.Handler {
	config := RateLimitConfig{
		Limit:  limit,
		Window: window,
		KeyFn: func(r *http.Request) string {
			return rateLimiter.EndpointKey(endpoint, r.Method)
		},
	}
	
	return RateLimit(rateLimiter, config, logger)
}

func UserEndpointRateLimit(rateLimiter *redis.RateLimiterService, endpoint string, limit int64, window time.Duration, logger *zap.Logger) func(http.Handler) http.Handler {
	config := RateLimitConfig{
		Limit:  limit,
		Window: window,
		KeyFn: func(r *http.Request) string {
			userID := r.Context().Value("user_id")
			if userID == nil {
				return rateLimiter.IPKey(getClientIP(r))
			}
			return rateLimiter.UserEndpointKey(userID.(string), endpoint, r.Method)
		},
	}
	
	return RateLimit(rateLimiter, config, logger)
}

func getClientIP(r *http.Request) string {
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		return xff
	}
	
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}
	
	return r.RemoteAddr
}
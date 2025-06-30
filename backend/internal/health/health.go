package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"backend/internal/redis"
)

type HealthService struct {
	db          *gorm.DB
	redisClient *redis.Client
	logger      *zap.Logger
}

type HealthResponse struct {
	Status   string            `json:"status"`
	Services map[string]Health `json:"services"`
}

type Health struct {
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
	Timestamp string `json:"timestamp"`
}

func NewHealthService(db *gorm.DB, redisClient *redis.Client, logger *zap.Logger) *HealthService {
	return &HealthService{
		db:          db,
		redisClient: redisClient,
		logger:      logger,
	}
}

func (h *HealthService) CheckHealth(ctx context.Context) *HealthResponse {
	response := &HealthResponse{
		Status:   "healthy",
		Services: make(map[string]Health),
	}

	dbHealth := h.checkDatabase(ctx)
	redisHealth := h.checkRedis(ctx)

	response.Services["database"] = dbHealth
	response.Services["redis"] = redisHealth

	if dbHealth.Status == "unhealthy" || redisHealth.Status == "unhealthy" {
		response.Status = "unhealthy"
	}

	return response
}

func (h *HealthService) checkDatabase(ctx context.Context) Health {
	health := Health{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	timeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	sqlDB, err := h.db.DB()
	if err != nil {
		health.Status = "unhealthy"
		health.Message = "Failed to get database connection"
		h.logger.Error("Health check: database connection failed", zap.Error(err))
		return health
	}

	if err := sqlDB.PingContext(timeout); err != nil {
		health.Status = "unhealthy"
		health.Message = "Database ping failed"
		h.logger.Error("Health check: database ping failed", zap.Error(err))
		return health
	}

	health.Status = "healthy"
	h.logger.Debug("Health check: database is healthy")
	return health
}

func (h *HealthService) checkRedis(ctx context.Context) Health {
	health := Health{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	timeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := h.redisClient.HealthCheck(timeout); err != nil {
		health.Status = "unhealthy"
		health.Message = "Redis ping failed"
		h.logger.Error("Health check: Redis ping failed", zap.Error(err))
		return health
	}

	health.Status = "healthy"
	h.logger.Debug("Health check: Redis is healthy")
	return health
}

func (h *HealthService) HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		
		health := h.CheckHealth(ctx)
		
		w.Header().Set("Content-Type", "application/json")
		
		if health.Status == "unhealthy" {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		
		if err := json.NewEncoder(w).Encode(health); err != nil {
			h.logger.Error("Failed to encode health response", zap.Error(err))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}
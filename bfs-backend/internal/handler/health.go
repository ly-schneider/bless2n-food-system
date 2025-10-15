package handler

import (
	"backend/internal/response"
	"context"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type HealthHandler struct {
	logger *zap.Logger
	db     *mongo.Database
}

type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Checks    map[string]string `json:"checks,omitempty"`
}

func NewHealthHandler(logger *zap.Logger, db *mongo.Database, jwksClient interface{}) *HealthHandler {
	return &HealthHandler{
		logger: logger,
		db:     db,
	}
}

// Healthz godoc
// @Summary Liveness probe
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /healthz [get]
func (h *HealthHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	healthResponse := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	response.WriteSuccess(w, http.StatusOK, healthResponse)
}

// Readyz godoc
// @Summary Readiness probe
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 503 {object} response.ProblemDetails
// @Router /readyz [get]
func (h *HealthHandler) Readyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	checks := make(map[string]string)
	allHealthy := true

	if err := h.db.Client().Ping(ctx, nil); err != nil {
		checks["database"] = "unhealthy: " + err.Error()
		allHealthy = false
		h.logger.Warn("Database health check failed", zap.Error(err))
	} else {
		checks["database"] = "healthy"
	}

	if !allHealthy {
		// Return RFC 9457 Problem Details for unhealthy service
		problem := response.NewProblem(
			http.StatusServiceUnavailable,
			"Service Unavailable", 
			"One or more health checks failed",
		)
		problem.Errors = []response.ValidationError{
			{
				Field:   "#/checks",
				Message: "Database health check failed",
				Value:   checks,
			},
		}
		response.WriteProblem(w, problem)
		return
	}

	healthResponse := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Checks:    checks,
	}

	response.WriteSuccess(w, http.StatusOK, healthResponse)
}

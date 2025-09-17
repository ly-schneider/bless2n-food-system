package handler

import (
	"backend/internal/service"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type StationHandler struct {
	stationService service.StationService
	validator      *validator.Validate
	logger         *zap.Logger
}

func NewStationHandler(stationService service.StationService, logger *zap.Logger) *StationHandler {
	return &StationHandler{
		stationService: stationService,
		validator:      validator.New(),
		logger:         logger,
	}
}

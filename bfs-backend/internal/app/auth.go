package app

import (
	"backend/internal/auth"
	"backend/internal/repository"

	"go.uber.org/zap"
)

func NewSessionMiddleware(sessionRepo repository.SessionRepository, bindingRepo repository.DeviceBindingRepository, logger *zap.Logger) *auth.SessionMiddleware {
	return auth.NewSessionMiddleware(sessionRepo, bindingRepo, logger)
}

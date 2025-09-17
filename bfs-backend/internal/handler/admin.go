package handler

import (
	"backend/internal/service"

	"github.com/go-playground/validator/v10"
)

type AdminHandler struct {
	adminService service.AdminService
	validator    *validator.Validate
}

func NewAdminHandler(adminService service.AdminService) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
		validator:    validator.New(),
	}
}

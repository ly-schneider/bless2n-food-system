package handler

import (
    "backend/internal/service"
    "net/http"

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

// Ping is a trivial protected endpoint for RBAC tests
func (h *AdminHandler) Ping(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write([]byte(`{"ok":true}`))
}

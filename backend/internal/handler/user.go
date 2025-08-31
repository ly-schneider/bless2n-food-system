package handler

import (
	"encoding/json"
	"net/http"

	"backend/internal/middleware"
	"backend/internal/response"
	"backend/internal/service"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type UserHandler struct {
	userService service.UserService
	validator   *validator.Validate
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
		validator:   validator.New(),
	}
}

// UpdateProfile godoc
// @Summary Update customer profile
// @Description Allow customers to update their email address
// @Tags users
// @Accept json
// @Produce json
// @Security Bearer
// @Param request body service.UpdateProfileRequest true "Profile update payload"
// @Success 200 {object} service.UpdateProfileResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /v1/users/profile [put]
func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	// Get user from JWT token
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req service.UpdateProfileRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		zap.L().Error("failed to decode request body", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		zap.L().Error("validation failed", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	svcResp, err := h.userService.UpdateProfile(r.Context(), user.Subject, req)
	if err != nil {
		zap.L().Error("failed to update profile", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.WriteJSON(w, http.StatusOK, svcResp)
}

// DeleteProfile godoc
// @Summary Delete customer profile
// @Description Allow customers to delete their own profile
// @Tags users
// @Produce json
// @Security Bearer
// @Success 200 {object} service.DeleteProfileResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /v1/users/profile [delete]
func (h *UserHandler) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	// Get user from JWT token
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	svcResp, err := h.userService.DeleteProfile(r.Context(), user.Subject)
	if err != nil {
		zap.L().Error("failed to delete profile", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.WriteJSON(w, http.StatusOK, svcResp)
}

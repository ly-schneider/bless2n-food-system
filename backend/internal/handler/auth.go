package handler

import (
	"encoding/json"
	"net/http"

	"backend/internal/response"
	"backend/internal/service"

	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type AuthHandler struct {
	authService service.AuthService
	validator   *validator.Validate
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator.New(),
	}
}

// RegisterCustomer godoc
// @Summary Register customer
// @Description Register a new customer and send an OTP to email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.RegisterCustomerRequest true "payload"
// @Success 201 {object} service.RegisterCustomerResponse
// @Failure 400 {object} response.ErrorResponse
// @Router /v1/auth/register/customer [post]
func (h *AuthHandler) RegisterCustomer(w http.ResponseWriter, r *http.Request) {
	var req service.RegisterCustomerRequest

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

	svcResp, err := h.authService.RegisterCustomer(r.Context(), req)
	if err != nil {
		zap.L().Error("failed to register customer", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.WriteJSON(w, http.StatusCreated, svcResp)
}

// VerifyOTP godoc
// @Summary Verify OTP
// @Description Verify a 6-digit OTP sent to the user's email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.VerifyOTPRequest true "payload"
// @Success 200 {object} service.VerifyOTPResponse
// @Failure 400 {object} response.ErrorResponse
// @Router /v1/auth/verify-otp [post]
func (h *AuthHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req service.VerifyOTPRequest

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

	svcResp, err := h.authService.VerifyOTP(r.Context(), req)
	if err != nil {
		zap.L().Error("failed to verify OTP", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.WriteJSON(w, http.StatusOK, svcResp)
}

// ResendOTP godoc
// @Summary Resend OTP
// @Description Resend a login/verification OTP to the user's email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.ResendOTPRequest true "payload"
// @Success 200 {object} service.ResendOTPResponse
// @Failure 400 {object} response.ErrorResponse
// @Router /v1/auth/resend-otp [post]
func (h *AuthHandler) ResendOTP(w http.ResponseWriter, r *http.Request) {
	var req service.ResendOTPRequest

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

	svcResp, err := h.authService.ResendOTP(r.Context(), req)
	if err != nil {
		zap.L().Error("failed to resend OTP", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.WriteJSON(w, http.StatusOK, svcResp)
}

// RequestLoginOTP godoc
// @Summary Request login OTP
// @Description Start passwordless login by requesting an OTP for the given email
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.RequestLoginOTPRequest true "payload"
// @Success 200 {object} service.RequestLoginOTPResponse
// @Failure 400 {object} response.ErrorResponse
// @Router /v1/auth/request-login-otp [post]
func (h *AuthHandler) RequestLoginOTP(w http.ResponseWriter, r *http.Request) {
	var req service.RequestLoginOTPRequest

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

	svcResp, err := h.authService.RequestLoginOTP(r.Context(), req)
	if err != nil {
		zap.L().Error("failed to request login OTP", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.WriteJSON(w, http.StatusOK, svcResp)
}

// Login godoc
// @Summary Login with OTP
// @Description Exchange a valid OTP for an access/refresh token pair
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.LoginRequest true "payload"
// @Success 200 {object} service.LoginResponse
// @Failure 400 {object} response.ErrorResponse
// @Router /v1/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req service.LoginRequest

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

	svcResp, err := h.authService.Login(r.Context(), req)
	if err != nil {
		zap.L().Error("failed to login", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.WriteJSON(w, http.StatusOK, svcResp)
}

// RefreshToken godoc
// @Summary Refresh tokens
// @Description Exchange a valid refresh token for a new access/refresh token pair
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.RefreshTokenRequest true "payload"
// @Success 200 {object} service.RefreshTokenResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req service.RefreshTokenRequest

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

	svcResp, err := h.authService.RefreshToken(r.Context(), req)
	if err != nil {
		zap.L().Error("failed to refresh token", zap.Error(err))
		response.WriteError(w, http.StatusUnauthorized, err.Error())
		return
	}

	response.WriteJSON(w, http.StatusOK, svcResp)
}

// Logout godoc
// @Summary Logout (invalidate refresh token)
// @Description Revoke the provided refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body service.LogoutRequest true "payload"
// @Success 200 {object} service.LogoutResponse
// @Failure 400 {object} response.ErrorResponse
// @Router /v1/auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req service.LogoutRequest

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

	svcResp, err := h.authService.Logout(r.Context(), req)
	if err != nil {
		zap.L().Error("failed to logout", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.WriteJSON(w, http.StatusOK, svcResp)
}

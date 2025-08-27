package handler

import (
	"encoding/json"
	"net/http"

	"backend/internal/service"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type AuthHandler struct {
	authService service.AuthService
	validator   *validator.Validate
	logger      *zap.Logger
}

func NewAuthHandler(authService service.AuthService, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator.New(),
		logger:      logger,
	}
}

func (h *AuthHandler) RegisterCustomer(w http.ResponseWriter, r *http.Request) {
	var req service.RegisterCustomerRequest
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request body", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		h.logger.Error("validation failed", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	response, err := h.authService.RegisterCustomer(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to register customer", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, response)
}

func (h *AuthHandler) VerifyOTP(w http.ResponseWriter, r *http.Request) {
	var req service.VerifyOTPRequest
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request body", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		h.logger.Error("validation failed", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	response, err := h.authService.VerifyOTP(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to verify OTP", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *AuthHandler) ResendOTP(w http.ResponseWriter, r *http.Request) {
	var req service.ResendOTPRequest
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request body", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		h.logger.Error("validation failed", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	response, err := h.authService.ResendOTP(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to resend OTP", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *AuthHandler) writeJSONResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req service.LoginRequest
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request body", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		h.logger.Error("validation failed", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	response, err := h.authService.Login(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to login", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req service.RefreshTokenRequest
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request body", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		h.logger.Error("validation failed", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	response, err := h.authService.RefreshToken(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to refresh token", zap.Error(err))
		h.writeErrorResponse(w, http.StatusUnauthorized, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req service.LogoutRequest
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request body", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		h.logger.Error("validation failed", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	response, err := h.authService.Logout(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to logout", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *AuthHandler) RequestLoginOTP(w http.ResponseWriter, r *http.Request) {
	var req service.RequestLoginOTPRequest
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request body", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.validator.Struct(req); err != nil {
		h.logger.Error("validation failed", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	response, err := h.authService.RequestLoginOTP(r.Context(), req)
	if err != nil {
		h.logger.Error("failed to request login OTP", zap.Error(err))
		h.writeErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

func (h *AuthHandler) writeErrorResponse(w http.ResponseWriter, status int, message string) {
	errorResponse := map[string]any{
		"error":   true,
		"message": message,
		"status":  status,
	}
	h.writeJSONResponse(w, status, errorResponse)
}
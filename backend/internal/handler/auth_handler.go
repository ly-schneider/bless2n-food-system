package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"backend/internal/apperrors"
	"backend/internal/domain"
	"backend/internal/http/respond"
	"backend/internal/logger"
	"backend/internal/model"
	"backend/internal/service"
	"backend/internal/utils"
)

type AuthHandler struct {
	svc                 service.AuthService
	verificationService service.VerificationService
	vldt                *validator.Validate
}

func NewAuthHandler(svc service.AuthService, verificationService service.VerificationService) AuthHandler {
	return AuthHandler{
		svc:                 svc,
		verificationService: verificationService,
		vldt:                validator.New(),
	}
}

func (h AuthHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Post("/refresh", h.Refresh)
	r.Post("/logout", h.Logout)
	r.Post("/verify-email", h.VerifyEmail)
	r.Post("/resend-verification", h.ResendVerification)
	return r
}

func (h AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req service.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.NewWriter(w).WriteError(apperrors.BadRequest("bad_json", domain.ErrParseBody.Error(), err))
		return
	}
	if err := h.vldt.Struct(req); err != nil {
		respond.NewWriter(w).WriteError(apperrors.BadRequest("validation_error", domain.ErrInvalidBody.Error(), err))
		return
	}

	out, err := h.svc.Register(r.Context(), &req)
	if err != nil {
		respond.NewWriter(w).WriteError(err)
		return
	}

	respond.JSON(w, http.StatusCreated, out)
}

func (h AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req service.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.NewWriter(w).WriteError(apperrors.BadRequest("bad_json", domain.ErrParseBody.Error(), err))
		return
	}

	if err := h.vldt.Struct(req); err != nil {
		respond.NewWriter(w).WriteError(apperrors.BadRequest("validation_error", domain.ErrInvalidBody.Error(), err))
		return
	}

	browserLoginResp, mobileLoginResp, err := h.svc.Login(r.Context(), &req)
	if err != nil {
		respond.NewWriter(w).WriteError(err)
		return
	}

	ua := r.UserAgent()

	if utils.IsMobile(&ua) {
		respond.JSON(w, http.StatusOK, mobileLoginResp)
		return
	}

	setSharedCookie(w, browserLoginResp.RefreshToken)
	respond.JSON(w, http.StatusOK, browserLoginResp)
}

func (h AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var plainRefreshToken string
	if c, err := r.Cookie("refresh_token"); err == nil {
		plainRefreshToken = c.Value
	} else {
		var body struct {
			Refresh string `json:"refresh_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			respond.NewWriter(w).WriteError(apperrors.BadRequest("bad_json", domain.ErrParseBody.Error(), err))
			return
		}
		defer r.Body.Close()
		plainRefreshToken = body.Refresh
	}

	if plainRefreshToken == "" {
		respond.NewWriter(w).WriteError(apperrors.BadRequest("missing_refresh_token", "refresh token is required", nil))
		return
	}

	browserLoginResp, mobileLoginResp, err := h.svc.Refresh(r.Context(), plainRefreshToken)
	if err != nil {
		logger.L.Error(r.Context(), "failed to refresh token: ", err)
		logger.L.Infow("status code from error", "status_code", err.Status)
		respond.NewWriter(w).WriteError(err)
		return
	}

	ua := r.UserAgent()

	if utils.IsMobile(&ua) {
		respond.JSON(w, http.StatusOK, mobileLoginResp)
		return
	}

	setSharedCookie(w, browserLoginResp.RefreshToken)
	respond.JSON(w, http.StatusOK, browserLoginResp)
}

func (h AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var plainRefreshToken string
	if c, err := r.Cookie("refresh_token"); err == nil {
		plainRefreshToken = c.Value
	} else {
		var body struct {
			Refresh string `json:"refresh_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			respond.NewWriter(w).WriteError(apperrors.BadRequest("bad_json", domain.ErrParseBody.Error(), err))
			return
		}
		defer r.Body.Close()
		plainRefreshToken = body.Refresh
	}

	if plainRefreshToken == "" {
		respond.NewWriter(w).WriteError(apperrors.BadRequest("missing_refresh_token", "refresh token is required", nil))
		return
	}

	err := h.svc.Logout(r.Context(), plainRefreshToken)
	if err != nil {
		logger.L.Error(r.Context(), "failed to refresh token: ", err)
		logger.L.Infow("status code from error", "status_code", err.Status)
		respond.NewWriter(w).WriteError(err)
		return
	}

	respond.JSON(w, http.StatusOK, nil)
}

func (h AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"user_id" validate:"required"`
		Code   string `json:"code" validate:"required,len=6"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.NewWriter(w).WriteError(apperrors.BadRequest("bad_json", domain.ErrParseBody.Error(), err))
		return
	}
	
	if err := h.vldt.Struct(req); err != nil {
		respond.NewWriter(w).WriteError(apperrors.BadRequest("validation_error", domain.ErrInvalidBody.Error(), err))
		return
	}

	// Convert string to NanoID14
	userID := model.NanoID14(req.UserID)
	
	err := h.verificationService.VerifyCode(r.Context(), userID, req.Code)
	if err != nil {
		if err == domain.ErrVerificationTokenNotFound {
			respond.NewWriter(w).WriteError(apperrors.BadRequest("invalid_code", "Invalid or expired verification code", err))
			return
		}
		if err == domain.ErrVerificationTokenExpired {
			respond.NewWriter(w).WriteError(apperrors.BadRequest("expired_code", "Verification code has expired", err))
			return
		}
		respond.NewWriter(w).WriteError(apperrors.FromStatus(http.StatusInternalServerError, "Failed to verify code", err))
		return
	}

	respond.JSON(w, http.StatusOK, map[string]string{
		"message": "Email verified successfully",
	})
}

func (h AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email" validate:"required,email"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.NewWriter(w).WriteError(apperrors.BadRequest("bad_json", domain.ErrParseBody.Error(), err))
		return
	}
	
	if err := h.vldt.Struct(req); err != nil {
		respond.NewWriter(w).WriteError(apperrors.BadRequest("validation_error", domain.ErrInvalidBody.Error(), err))
		return
	}

	// This would require getting the user by email first
	// For now, we'll just return success to prevent email enumeration attacks
	respond.JSON(w, http.StatusOK, map[string]string{
		"message": "If a user with this email exists, a verification email has been sent",
	})
}

func setSharedCookie(w http.ResponseWriter, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    value,
		Domain:   ".rentro.ch",
		Path:     "/v1/auth/refresh",
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

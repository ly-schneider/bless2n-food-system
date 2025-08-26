package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"backend/internal/http/middleware"
	"backend/internal/http/respond"
	"backend/internal/logger"
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
		// Respond with bad request error
		return
	}
	if err := h.vldt.Struct(req); err != nil {
		// Respond with bad request / validation error
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
		// Respond with bad request error
		return
	}

	if err := h.vldt.Struct(req); err != nil {
		// Respond with bad request / validation error
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
			// Respond with bad request error
			return
		}
		defer r.Body.Close()
		plainRefreshToken = body.Refresh
	}

	if plainRefreshToken == "" {
		// Respond with bad request error
		return
	}

	loginResponse, err := h.svc.Refresh(r.Context(), plainRefreshToken)
	if err != nil {
		logger.L.Error(r.Context(), "failed to refresh token: ", err)
		logger.L.Infow("status code from error", "status_code", err.Status)
		respond.NewWriter(w).WriteError(err)
		return
	}

	setSharedCookie(w, loginResponse.RefreshToken)
	respond.JSON(w, http.StatusOK, loginResponse)
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
			// Respond with bad request error
			return
		}
		defer r.Body.Close()
		plainRefreshToken = body.Refresh
	}

	if plainRefreshToken == "" {
		// Respond with bad request error
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
		Code string `json:"code" validate:"required,len=6"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Respond with bad request error
		return
	}

	if err := h.vldt.Struct(req); err != nil {
		// Respond with bad request / validation error
		return
	}

	userID := middleware.ExtractUserIDFromContext(r.Context())
	if userID == nil || *userID == "" {
		// Respond with unauthorized error
		return
	}

	userIDNano := model.NanoID14(*userID)

	verifyErr := h.verificationService.VerifyCode(r.Context(), userIDNano, req.Code)
	if verifyErr != nil {
		if verifyErr == domain.ErrVerificationTokenNotFound {
			// Respond with not found error
			return
		}
		if verifyErr == domain.ErrVerificationTokenExpired {
			// Respond with expired error
			return
		}
		// Respond with failed to verify code error
		return
	}

	respond.JSON(w, http.StatusOK, map[string]string{
		"message": "email verified successfully",
	})
}

func (h AuthHandler) ResendVerification(w http.ResponseWriter, r *http.Request) {
	userID := middleware.ExtractUserIDFromContext(r.Context())
	if userID == nil || *userID == "" {
		// Respond with unauthorized error
		return
	}

	userIDNano := model.NanoID14(*userID)

	verifyErr := h.verificationService.SendVerificationCode(r.Context(), userIDNano)
	if verifyErr != nil {
		if verifyErr == domain.ErrVerificationTokenNotFound {
			// Respond with not found error
			return
		}
		if verifyErr == domain.ErrVerificationTokenExpired {
			// Respond with expired error
			return
		}
		// Respond with failed to send verification code error
		return
	}

	respond.JSON(w, http.StatusOK, map[string]string{
		"message": "Email verified successfully",
	})
}

func setSharedCookie(w http.ResponseWriter, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    value,
		Domain:   ".blessthun.ch",
		Path:     "/auth/refresh",
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
}

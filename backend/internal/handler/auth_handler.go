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
	"backend/internal/service"
	"backend/internal/utils"
)

type AuthHandler struct {
	svc  service.AuthService
	vldt *validator.Validate
}

func NewAuthHandler(svc service.AuthService) AuthHandler {
	return AuthHandler{svc: svc, vldt: validator.New()}
}

func (h AuthHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/register", h.Register)
	r.Post("/login", h.Login)
	r.Post("/refresh", h.Refresh)
	r.Post("/logout", h.Logout)
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

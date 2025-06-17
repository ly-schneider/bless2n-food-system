package handler

import (
	"encoding/json"
	e "errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"backend/internal/errors"
	"backend/internal/http/respond"
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
	return r
}

func (h AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req service.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.NewWriter(w).SetError(errors.BadRequest("bad_json", "unable to parse body", err))
		return
	}
	if err := h.vldt.Struct(req); err != nil {
		respond.NewWriter(w).SetError(errors.BadRequest("validation_error", "invalid input", err))
		return
	}

	out, err := h.svc.Register(r.Context(), &req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "user already exists with this email" {
			statusCode = http.StatusConflict
		}

		respond.NewWriter(w).SetError(errors.FromStatus(statusCode, "failed to register user", err))
		return
	}

	respond.JSON(w, http.StatusCreated, out)
}

func (h AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req service.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.NewWriter(w).SetError(errors.BadRequest("bad_json", "unable to parse body", err))
		return
	}
	if err := h.vldt.Struct(req); err != nil {
		respond.NewWriter(w).SetError(errors.BadRequest("validation_error", "invalid input", err))
		return
	}

	browserLoginResp, mobileLoginResp, err := h.svc.Login(r.Context(), &req, r.UserAgent())
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "user not found" || err.Error() == "invalid credentials" {
			statusCode = http.StatusUnauthorized
		} else if err.Error() == "invalid login request: missing required fields" {
			statusCode = http.StatusBadRequest
		}

		respond.NewWriter(w).SetError(errors.FromStatus(statusCode, "failed to login user", err))
		return
	}

	if utils.IsMobile(r.UserAgent()) {
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
			respond.NewWriter(w).SetError(errors.BadRequest("bad_json", "unable to parse body", err))
			return
		}
		defer r.Body.Close()
		plainRefreshToken = body.Refresh
	}

	if plainRefreshToken == "" {
		err := e.New("refresh token is required")
		respond.NewWriter(w).SetError(errors.BadRequest("missing_refresh_token", "refresh token is required", err))
		return
	}

	browserLoginResp, mobileLoginResp, err := h.svc.Refresh(r.Context(), plainRefreshToken, r.UserAgent())
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "refresh token not found" ||
			err.Error() == "refresh token is revoked or expired" ||
			err.Error() == "user not found" {
			statusCode = http.StatusUnauthorized
		}

		respond.NewWriter(w).SetError(errors.FromStatus(statusCode, "failed to refresh tokens", err))
		return
	}

	if utils.IsMobile(r.UserAgent()) {
		respond.JSON(w, http.StatusOK, mobileLoginResp)
		return
	}

	setSharedCookie(w, browserLoginResp.RefreshToken)
	respond.JSON(w, http.StatusOK, browserLoginResp)
}

func setSharedCookie(w http.ResponseWriter, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    value,
		Domain:   ".rentro.ch", // share across sub-domains
		Path:     "/auth/refresh",
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode, // blocks CSRF on cross-site 3rd-party calls
	})
}

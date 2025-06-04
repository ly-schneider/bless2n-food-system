package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

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
	return r
}

func (h AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	logger.L.Infow("Handling register request", "method", r.Method, "path", r.URL.Path)

	var req service.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.L.Errorw("Failed to decode register request body", "error", err)
		utils.RespondError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.vldt.Struct(req); err != nil {
		logger.L.Errorw("Register validation failed", "error", err)
		utils.RespondError(w, http.StatusUnprocessableEntity, err)
		return
	}

	out, err := h.svc.Register(r.Context(), &req)
	if err != nil {
		logger.L.Errorw("Failed to register user", "error", err, "email", req.Email)
		utils.RespondError(w, http.StatusInternalServerError, err)
		return
	}

	logger.L.Infow("Successfully handled register request", "email", req.Email)
	utils.RespondJSON(w, http.StatusCreated, out)
}

func (h AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	logger.L.Infow("Handling login request", "method", r.Method, "path", r.URL.Path)

	var req service.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.L.Errorw("Failed to decode login request body", "error", err)
		utils.RespondError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.vldt.Struct(req); err != nil {
		logger.L.Errorw("Login validation failed", "error", err)
		utils.RespondError(w, http.StatusUnprocessableEntity, err)
		return
	}

	browserLoginResp, mobileLoginResp, err := h.svc.Login(r.Context(), &req, r.UserAgent())
	if err != nil {
		logger.L.Errorw("Failed to login user", "error", err, "email", req.Email)
		utils.RespondError(w, http.StatusInternalServerError, err)
		return
	}

	logger.L.Infow("Successfully handled login request", "email", req.Email)

	if utils.IsMobile(r.UserAgent()) {
		utils.RespondJSON(w, http.StatusOK, mobileLoginResp)
		return
	}

	setSharedCookie(w, browserLoginResp.RefreshToken)
	utils.RespondJSON(w, http.StatusOK, browserLoginResp)
}

func (h AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	logger.L.Infow("Handling refresh request", "method", r.Method, "path", r.URL.Path)

	var plainRefreshToken string
	if c, err := r.Cookie("refresh_token"); err == nil {
		plainRefreshToken = c.Value
	} else {
		var body struct {
			Refresh string `json:"refresh_token"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		plainRefreshToken = body.Refresh
	}

	browserLoginResp, mobileLoginResp, err := h.svc.Refresh(r.Context(), plainRefreshToken, r.UserAgent())
	if err != nil {
		logger.L.Errorw("Failed to refresh tokens", "error", err, "refresh_token", plainRefreshToken)
		utils.RespondError(w, http.StatusInternalServerError, err)
		return
	}

	if utils.IsMobile(r.UserAgent()) {
		utils.RespondJSON(w, http.StatusOK, mobileLoginResp)
		return
	}

	setSharedCookie(w, browserLoginResp.RefreshToken)
	utils.RespondJSON(w, http.StatusOK, browserLoginResp)
}

func setSharedCookie(w http.ResponseWriter, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     "refresh_token",
		Value:    value,
		Domain:   ".rentro.ch", // share across sub-domains
		Path:     "/v1/auth/refresh",
		Expires:  time.Now().Add(30 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode, // blocks CSRF on cross-site 3rd-party calls
	})
}

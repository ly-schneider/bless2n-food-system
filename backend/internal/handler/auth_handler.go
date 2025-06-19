package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"backend/internal/apperrors"
	"backend/internal/domain"
	"backend/internal/http/respond"
	"backend/internal/service"
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
	return r
}

func (h AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req service.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respond.NewWriter(w).SetError(apperrors.BadRequest("bad_json", domain.ErrParseBody.Error(), err))
		return
	}
	if err := h.vldt.Struct(req); err != nil {
		respond.NewWriter(w).SetError(apperrors.BadRequest("validation_error", domain.ErrInvalidBody.Error(), err))
		return
	}

	out, err := h.svc.Register(r.Context(), &req)
	if err != nil {
		respond.NewWriter(w).SetError(err)
		return
	}

	respond.JSON(w, http.StatusCreated, out)
}

package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"backend/internal/domain"
	"backend/internal/logger"
	"backend/internal/service"
)

type ProductHandler struct {
	svc  service.ProductService
	vldt *validator.Validate
}

func NewProductHandler(svc service.ProductService) ProductHandler {
	return ProductHandler{svc: svc, vldt: validator.New()}
}

func (h ProductHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.List)
	r.Get("/{id}", h.Get)
	r.Post("/", h.Create)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	return r
}

func (h ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	logger.L.Infow("Handling product list request", "method", r.Method, "path", r.URL.Path)

	out, err := h.svc.List(r.Context())
	if err != nil {
		logger.L.Errorw("Failed to list products", "error", err)
		respondError(w, http.StatusInternalServerError, err)
		return
	}

	logger.L.Infow("Successfully handled product list request", "count", len(out))
	respondJSON(w, http.StatusOK, out)
}

func (h ProductHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	logger.L.Infow("Handling get product request", "id", id, "method", r.Method, "path", r.URL.Path)

	out, err := h.svc.Get(r.Context(), id)
	if err != nil {
		logger.L.Errorw("Failed to get product", "id", id, "error", err)
		respondError(w, http.StatusNotFound, err)
		return
	}

	logger.L.Infow("Successfully handled get product request", "id", id, "name", out.Name)
	respondJSON(w, http.StatusOK, out)
}

func (h ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	logger.L.Infow("Handling create product request", "method", r.Method, "path", r.URL.Path)

	var in domain.Product
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		logger.L.Errorw("Failed to decode product request body", "error", err)
		respondError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.vldt.Struct(in); err != nil {
		logger.L.Errorw("Product validation failed", "error", err, "product", in)
		respondError(w, http.StatusUnprocessableEntity, err)
		return
	}

	if err := h.svc.Create(r.Context(), &in); err != nil {
		logger.L.Errorw("Failed to create product", "error", err, "product", in)
		respondError(w, http.StatusInternalServerError, err)
		return
	}

	logger.L.Infow("Successfully created product", "id", in.ID, "name", in.Name)
	respondJSON(w, http.StatusCreated, in)
}

func (h ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	logger.L.Infow("Handling update product request", "id", id, "method", r.Method, "path", r.URL.Path)

	var in domain.Product
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		logger.L.Errorw("Failed to decode update product request body", "id", id, "error", err)
		respondError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.vldt.Struct(in); err != nil {
		logger.L.Errorw("Product update validation failed", "id", id, "error", err, "product", in)
		respondError(w, http.StatusUnprocessableEntity, err)
		return
	}

	out, err := h.svc.Update(r.Context(), id, &in)
	if err != nil {
		logger.L.Errorw("Failed to update product", "id", id, "error", err)
		respondError(w, http.StatusInternalServerError, err)
		return
	}

	logger.L.Infow("Successfully updated product", "id", id, "name", out.Name)
	respondJSON(w, http.StatusOK, out)
}

func (h ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}

	logger.L.Infow("Handling delete product request", "id", id, "method", r.Method, "path", r.URL.Path)

	if err := h.svc.Delete(r.Context(), id); err != nil {
		logger.L.Errorw("Failed to delete product", "id", id, "error", err)
		respondError(w, http.StatusInternalServerError, err)
		return
	}

	logger.L.Infow("Successfully deleted product", "id", id)
	w.WriteHeader(http.StatusNoContent)
}

func parseID(w http.ResponseWriter, r *http.Request) (uint, bool) {
	raw := chi.URLParam(r, "id")
	n, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		logger.L.Errorw("Failed to parse ID from URL", "raw_id", raw, "error", err)
		respondError(w, http.StatusBadRequest, err)
		return 0, false
	}
	return uint(n), true
}

func respondJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func respondError(w http.ResponseWriter, status int, err error) {
	respondJSON(w, status, map[string]string{"error": err.Error()})
}

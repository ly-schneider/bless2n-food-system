package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"backend/internal/middleware"
	"backend/internal/response"
	"backend/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type CategoryHandler struct {
	categoryService service.CategoryService
	validator       *validator.Validate
}

func NewCategoryHandler(categoryService service.CategoryService) *CategoryHandler {
	return &CategoryHandler{
		categoryService: categoryService,
		validator:       validator.New(),
	}
}

// CreateCategory godoc
// @Summary Create a new category
// @Description Allow admins to create a new category
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body service.CreateCategoryRequest true "Category creation payload"
// @Success 201 {object} service.CreateCategoryResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Router /v1/admin/categories [post]
func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req service.CreateCategoryRequest

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

	svcResp, err := h.categoryService.CreateCategory(r.Context(), req)
	if err != nil {
		zap.L().Error("failed to create category", zap.Error(err))
		response.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	response.WriteJSON(w, http.StatusCreated, svcResp)
}

// GetCategory godoc
// @Summary Get category by ID
// @Description Allow admins to get a category by ID
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Success 200 {object} service.GetCategoryResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /v1/admin/categories/{id} [get]
func (h *CategoryHandler) GetCategory(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	categoryID := chi.URLParam(r, "id")
	if categoryID == "" {
		response.WriteError(w, http.StatusBadRequest, "Category ID is required")
		return
	}

	svcResp, err := h.categoryService.GetCategory(r.Context(), categoryID)
	if err != nil {
		zap.L().Error("failed to get category", zap.Error(err))
		if err.Error() == "category not found" {
			response.WriteError(w, http.StatusNotFound, err.Error())
		} else {
			response.WriteError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	response.WriteJSON(w, http.StatusOK, svcResp)
}

// UpdateCategory godoc
// @Summary Update category
// @Description Allow admins to update a category
// @Tags admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Param request body service.UpdateCategoryRequest true "Category update payload"
// @Success 200 {object} service.UpdateCategoryResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /v1/admin/categories/{id} [put]
func (h *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	categoryID := chi.URLParam(r, "id")
	if categoryID == "" {
		response.WriteError(w, http.StatusBadRequest, "Category ID is required")
		return
	}

	var req service.UpdateCategoryRequest

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

	svcResp, err := h.categoryService.UpdateCategory(r.Context(), categoryID, req)
	if err != nil {
		zap.L().Error("failed to update category", zap.Error(err))
		if err.Error() == "category not found" {
			response.WriteError(w, http.StatusNotFound, err.Error())
		} else {
			response.WriteError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	response.WriteJSON(w, http.StatusOK, svcResp)
}

// DeleteCategory godoc
// @Summary Delete category
// @Description Allow admins to delete a category
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Success 200 {object} service.DeleteCategoryResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /v1/admin/categories/{id} [delete]
func (h *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	categoryID := chi.URLParam(r, "id")
	if categoryID == "" {
		response.WriteError(w, http.StatusBadRequest, "Category ID is required")
		return
	}

	svcResp, err := h.categoryService.DeleteCategory(r.Context(), categoryID)
	if err != nil {
		zap.L().Error("failed to delete category", zap.Error(err))
		if err.Error() == "category not found" {
			response.WriteError(w, http.StatusNotFound, err.Error())
		} else {
			response.WriteError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	response.WriteJSON(w, http.StatusOK, svcResp)
}

// ListCategories godoc
// @Summary List categories
// @Description Allow admins to list all categories
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param active_only query boolean false "Show only active categories" default(false)
// @Param limit query int false "Limit size" minimum(1) maximum(100) default(50)
// @Param offset query int false "Offset" minimum(0) default(0)
// @Success 200 {object} service.ListCategoriesResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Router /v1/admin/categories [get]
func (h *CategoryHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	activeOnlyStr := r.URL.Query().Get("active_only")

	limit := 50
	offset := 0
	activeOnly := false

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	if activeOnlyStr != "" {
		activeOnly = activeOnlyStr == "true"
	}

	svcResp, err := h.categoryService.ListCategories(r.Context(), activeOnly, limit, offset)
	if err != nil {
		zap.L().Error("failed to list categories", zap.Error(err))
		response.WriteError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.WriteJSON(w, http.StatusOK, svcResp)
}

// SetCategoryActive godoc
// @Summary Set category active status
// @Description Allow admins to activate or deactivate a category
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Param active query boolean true "Active status" default(true)
// @Success 200 {object} service.SetCategoryActiveResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /v1/admin/categories/{id}/status [put]
func (h *CategoryHandler) SetCategoryActive(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		response.WriteError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	categoryID := chi.URLParam(r, "id")
	if categoryID == "" {
		response.WriteError(w, http.StatusBadRequest, "Category ID is required")
		return
	}

	activeStr := r.URL.Query().Get("active")
	if activeStr == "" {
		response.WriteError(w, http.StatusBadRequest, "Active status is required")
		return
	}

	isActive := activeStr == "true"

	svcResp, err := h.categoryService.SetCategoryActive(r.Context(), categoryID, isActive)
	if err != nil {
		zap.L().Error("failed to set category active status", zap.Error(err))
		if err.Error() == "category not found" {
			response.WriteError(w, http.StatusNotFound, err.Error())
		} else {
			response.WriteError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	response.WriteJSON(w, http.StatusOK, svcResp)
}
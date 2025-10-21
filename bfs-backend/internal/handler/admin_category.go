package handler

import (
	"backend/internal/domain"
	"backend/internal/repository"
	"backend/internal/response"
	"encoding/json"
	"net/http"

	"go.uber.org/zap"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AdminCategoryHandler struct {
	categories repository.CategoryRepository
	audit      repository.AuditRepository
}

func NewAdminCategoryHandler(categories repository.CategoryRepository, audit repository.AuditRepository) *AdminCategoryHandler {
	return &AdminCategoryHandler{categories: categories, audit: audit}
}

// List godoc
// @Summary List categories
// @Tags admin-categories
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /v1/admin/categories [get]
func (h *AdminCategoryHandler) List(w http.ResponseWriter, r *http.Request) {
    // Minimal list; add filters later as needed
    items, total, err := h.categories.List(r.Context(), nil, nil, 200, 0)
    if err != nil {
        // Log underlying error for traceability of 5xx
        zap.L().Error("admin list categories failed", zap.Error(err), zap.String("method", r.Method), zap.String("path", r.URL.Path))
        response.WriteError(w, http.StatusInternalServerError, "failed to list categories")
        return
    }
	type CatDTO struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		IsActive bool   `json:"isActive"`
		Position int    `json:"position"`
	}
	out := make([]CatDTO, 0, len(items))
	for _, c := range items {
		out = append(out, CatDTO{ID: c.ID.Hex(), Name: c.Name, IsActive: c.IsActive, Position: c.Position})
	}
	response.WriteJSON(w, http.StatusOK, map[string]any{"items": out, "count": total})
}

// Create godoc
// @Summary Create category
// @Tags admin-categories
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param payload body createCategoryBody true "Category payload"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} response.ProblemDetails
// @Router /v1/admin/categories [post]
type createCategoryBody struct {
	Name     string `json:"name"`
	Position int    `json:"position"`
	IsActive *bool  `json:"isActive,omitempty"`
}

func (h *AdminCategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body createCategoryBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		response.WriteError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	if body.Position < 0 {
		response.WriteError(w, http.StatusBadRequest, "invalid position")
		return
	}
	c := &domain.Category{Name: body.Name, IsActive: true, Position: body.Position}
	if body.IsActive != nil {
		c.IsActive = *body.IsActive
	}
	id, err := h.categories.Insert(r.Context(), c)
	if err != nil {
		response.WriteError(w, http.StatusInternalServerError, "create failed")
		return
	}
	response.WriteJSON(w, http.StatusCreated, map[string]any{"id": id.Hex()})
}

// Update godoc
// @Summary Update category
// @Tags admin-categories
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Param payload body updateCategoryBody true "Category payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} response.ProblemDetails
// @Failure 404 {object} response.ProblemDetails
// @Router /v1/admin/categories/{id} [patch]
type updateCategoryBody struct {
	Name     *string `json:"name,omitempty"`
	IsActive *bool   `json:"isActive,omitempty"`
	Position *int    `json:"position,omitempty"`
}

func (h *AdminCategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chiURLParam(r, "id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body updateCategoryBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	set := primitive.M{}
	if body.Name != nil {
		set["name"] = *body.Name
	}
	if body.IsActive != nil {
		set["is_active"] = *body.IsActive
	}
	if body.Position != nil {
		if *body.Position < 0 {
			response.WriteError(w, http.StatusBadRequest, "invalid position")
			return
		}
		set["position"] = *body.Position
	}
	if len(set) == 0 {
		response.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}
	if err := h.categories.UpdateFields(r.Context(), oid, set); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "update failed")
		return
	}
	response.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// Delete godoc
// @Summary Delete category
// @Tags admin-categories
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.ProblemDetails
// @Router /v1/admin/categories/{id} [delete]
func (h *AdminCategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chiURLParam(r, "id")
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.categories.DeleteByID(r.Context(), oid); err != nil {
		response.WriteError(w, http.StatusInternalServerError, "delete failed")
		return
	}
	response.WriteNoContent(w)
}

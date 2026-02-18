package api

import (
	"encoding/json"
	"net/http"

	"backend/internal/generated/api/generated"
	"backend/internal/response"

	openapi_types "github.com/oapi-codegen/runtime/types"
)

// ListCategories returns all categories.
// (GET /categories)
func (h *Handlers) ListCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := h.categories.GetAll(r.Context())
	if err != nil {
		writeEntError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, generated.CategoryList{
		Items: toAPICategories(cats),
	})
}

// CreateCategory creates a new category.
// (POST /categories)
func (h *Handlers) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var body generated.CategoryCreate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	position := 0
	if body.Position != nil {
		position = *body.Position
	}

	cat, err := h.categories.Create(r.Context(), body.Name, position)
	if err != nil {
		writeEntError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusCreated, toAPICategory(cat))
}

// GetCategory returns a single category by ID.
// (GET /categories/{categoryId})
func (h *Handlers) GetCategory(w http.ResponseWriter, r *http.Request, categoryId openapi_types.UUID) {
	cat, err := h.categories.GetByID(r.Context(), categoryId)
	if err != nil {
		writeEntError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, toAPICategory(cat))
}

// UpdateCategory partially updates a category.
// (PATCH /categories/{categoryId})
func (h *Handlers) UpdateCategory(w http.ResponseWriter, r *http.Request, categoryId openapi_types.UUID) {
	// Fetch current state so we can merge partial updates.
	existing, err := h.categories.GetByID(r.Context(), categoryId)
	if err != nil {
		writeEntError(w, err)
		return
	}

	var body generated.CategoryUpdate
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_body", "Invalid request body")
		return
	}

	name := existing.Name
	if body.Name != nil {
		name = *body.Name
	}
	position := existing.Position
	if body.Position != nil {
		position = *body.Position
	}
	isActive := existing.IsActive
	if body.IsActive != nil {
		isActive = *body.IsActive
	}

	cat, err := h.categories.Update(r.Context(), categoryId, name, position, isActive)
	if err != nil {
		writeEntError(w, err)
		return
	}
	response.WriteJSON(w, http.StatusOK, toAPICategory(cat))
}

// DeleteCategory removes a category.
// (DELETE /categories/{categoryId})
func (h *Handlers) DeleteCategory(w http.ResponseWriter, r *http.Request, categoryId openapi_types.UUID) {
	if err := h.categories.Delete(r.Context(), categoryId); err != nil {
		writeEntError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

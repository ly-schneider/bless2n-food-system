package controllers

import (
	"net/http"

	"backend/models/models"
	"backend/services"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type CategoryController struct {
	service  *services.CategoryService
	validate *validator.Validate
}

func NewCategoryController(service *services.CategoryService) *CategoryController {
	return &CategoryController{
		service:  service,
		validate: validator.New(),
	}
}

func (c *CategoryController) GetCategories(ctx echo.Context) error {
	var isActive *bool
	if activeQuery := ctx.QueryParam("active"); activeQuery != "" {
		active := activeQuery == "true"
		isActive = &active
	}

	categories, err := c.service.GetAllCategories(isActive)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, categories)
}

func (c *CategoryController) GetCategory(ctx echo.Context) error {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid category ID"})
	}

	category, err := c.service.GetCategoryByID(id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Category not found"})
	}

	return ctx.JSON(http.StatusOK, category)
}

func (c *CategoryController) CreateCategory(ctx echo.Context) error {
	var category models.Category
	if err := ctx.Bind(&category); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if err := c.validate.Struct(category); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		errorMessages := make(map[string]string)

		for _, e := range validationErrors {
			errorMessages[e.Field()] = "This field is required"
		}

		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation failed",
			"details": errorMessages,
		})
	}

	if err := c.service.CreateCategory(&category); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusCreated, category)
}

func (c *CategoryController) UpdateCategory(ctx echo.Context) error {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid category ID"})
	}

	var category models.Category
	if err := ctx.Bind(&category); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	// Validate category fields
	if err := c.validate.Struct(category); err != nil {
		validationErrors := err.(validator.ValidationErrors)
		errorMessages := make(map[string]string)

		for _, e := range validationErrors {
			errorMessages[e.Field()] = "This field is required"
		}

		return ctx.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "Validation failed",
			"details": errorMessages,
		})
	}

	category.ID = id

	if err := c.service.UpdateCategory(&category); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusOK, category)
}

func (c *CategoryController) DeleteCategory(ctx echo.Context) error {
	idStr := ctx.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid category ID"})
	}

	if err := c.service.DeleteCategory(id); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return ctx.JSON(http.StatusNoContent, nil)
}

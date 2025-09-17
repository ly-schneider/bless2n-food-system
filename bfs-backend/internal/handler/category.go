package handler

import (
	"backend/internal/service"

	"github.com/go-playground/validator/v10"
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

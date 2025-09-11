package service

import (
	"context"
	"errors"
	"fmt"

	"backend/internal/domain"
	"backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CategoryService interface {
	CreateCategory(ctx context.Context, req CreateCategoryRequest) (*CreateCategoryResponse, error)
	GetCategory(ctx context.Context, categoryID string) (*GetCategoryResponse, error)
	UpdateCategory(ctx context.Context, categoryID string, req UpdateCategoryRequest) (*UpdateCategoryResponse, error)
	DeleteCategory(ctx context.Context, categoryID string) (*DeleteCategoryResponse, error)
	ListCategories(ctx context.Context, activeOnly bool, limit, offset int) (*ListCategoriesResponse, error)
	SetCategoryActive(ctx context.Context, categoryID string, isActive bool) (*SetCategoryActiveResponse, error)
}

type CreateCategoryRequest struct {
	Name string `json:"name" validate:"required"`
}

type CreateCategoryResponse struct {
	Category CategoryDTO `json:"category"`
	Message  string      `json:"message"`
	Success  bool        `json:"success"`
}

type GetCategoryResponse struct {
	Category CategoryDTO `json:"category"`
}

type UpdateCategoryRequest struct {
	Name string `json:"name" validate:"required"`
}

type UpdateCategoryResponse struct {
	Category CategoryDTO `json:"category"`
	Message  string      `json:"message"`
	Success  bool        `json:"success"`
}

type DeleteCategoryResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type ListCategoriesResponse struct {
	Categories []CategoryDTO `json:"categories"`
	Total      int           `json:"total"`
}

type SetCategoryActiveResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type CategoryDTO struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type categoryService struct {
	categoryRepo repository.CategoryRepository
}

func NewCategoryService(categoryRepo repository.CategoryRepository) CategoryService {
	return &categoryService{
		categoryRepo: categoryRepo,
	}
}

func (s *categoryService) CreateCategory(ctx context.Context, req CreateCategoryRequest) (*CreateCategoryResponse, error) {
	existingCategory, err := s.categoryRepo.GetByName(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing category: %w", err)
	}
	if existingCategory != nil {
		return nil, errors.New("category with this name already exists")
	}

	category := &domain.Category{
		Name: req.Name,
	}

	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return &CreateCategoryResponse{
		Category: s.toCategoryDTO(category),
		Message:  "Category created successfully",
		Success:  true,
	}, nil
}

func (s *categoryService) GetCategory(ctx context.Context, categoryID string) (*GetCategoryResponse, error) {
	objectID, err := primitive.ObjectIDFromHex(categoryID)
	if err != nil {
		return nil, errors.New("invalid category ID format")
	}

	category, err := s.categoryRepo.GetByID(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	if category == nil {
		return nil, errors.New("category not found")
	}

	return &GetCategoryResponse{
		Category: s.toCategoryDTO(category),
	}, nil
}

func (s *categoryService) UpdateCategory(ctx context.Context, categoryID string, req UpdateCategoryRequest) (*UpdateCategoryResponse, error) {
	objectID, err := primitive.ObjectIDFromHex(categoryID)
	if err != nil {
		return nil, errors.New("invalid category ID format")
	}

	category, err := s.categoryRepo.GetByID(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	if category == nil {
		return nil, errors.New("category not found")
	}

	if category.Name != req.Name {
		existingCategory, err := s.categoryRepo.GetByName(ctx, req.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing category: %w", err)
		}
		if existingCategory != nil {
			return nil, errors.New("category with this name already exists")
		}
	}

	category.Name = req.Name

	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return &UpdateCategoryResponse{
		Category: s.toCategoryDTO(category),
		Message:  "Category updated successfully",
		Success:  true,
	}, nil
}

func (s *categoryService) DeleteCategory(ctx context.Context, categoryID string) (*DeleteCategoryResponse, error) {
	objectID, err := primitive.ObjectIDFromHex(categoryID)
	if err != nil {
		return nil, errors.New("invalid category ID format")
	}

	category, err := s.categoryRepo.GetByID(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	if category == nil {
		return nil, errors.New("category not found")
	}

	if err := s.categoryRepo.Delete(ctx, objectID); err != nil {
		return nil, fmt.Errorf("failed to delete category: %w", err)
	}

	return &DeleteCategoryResponse{
		Message: "Category deleted successfully",
		Success: true,
	}, nil
}

func (s *categoryService) ListCategories(ctx context.Context, activeOnly bool, limit, offset int) (*ListCategoriesResponse, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	categories, err := s.categoryRepo.List(ctx, activeOnly, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}

	categoryDTOs := make([]CategoryDTO, len(categories))
	for i, category := range categories {
		categoryDTOs[i] = s.toCategoryDTO(category)
	}

	return &ListCategoriesResponse{
		Categories: categoryDTOs,
		Total:      len(categoryDTOs),
	}, nil
}

func (s *categoryService) SetCategoryActive(ctx context.Context, categoryID string, isActive bool) (*SetCategoryActiveResponse, error) {
	objectID, err := primitive.ObjectIDFromHex(categoryID)
	if err != nil {
		return nil, errors.New("invalid category ID format")
	}

	category, err := s.categoryRepo.GetByID(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	if category == nil {
		return nil, errors.New("category not found")
	}

	if err := s.categoryRepo.SetActive(ctx, objectID, isActive); err != nil {
		return nil, fmt.Errorf("failed to update category status: %w", err)
	}

	var action string
	if isActive {
		action = "activated"
	} else {
		action = "deactivated"
	}

	return &SetCategoryActiveResponse{
		Message: fmt.Sprintf("Category %s successfully", action),
		Success: true,
	}, nil
}

func (s *categoryService) toCategoryDTO(category *domain.Category) CategoryDTO {
	return CategoryDTO{
		ID:        category.ID.Hex(),
		Name:      category.Name,
		IsActive:  category.IsActive,
		CreatedAt: category.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: category.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}
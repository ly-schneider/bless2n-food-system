package services

import (
	"backend/models/models"
	"backend/repositories"

	"github.com/google/uuid"
)

type CategoryService struct {
	repo *repositories.CategoryRepository
}

func NewCategoryService(repo *repositories.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) GetAllCategories(isActive *bool) ([]models.Category, error) {
	return s.repo.FindAll(isActive)
}

func (s *CategoryService) GetCategoryByID(id uuid.UUID) (models.Category, error) {
	return s.repo.FindByID(id)
}

func (s *CategoryService) CreateCategory(category *models.Category) error {
	return s.repo.Create(category)
}

func (s *CategoryService) UpdateCategory(category *models.Category) error {
	return s.repo.Update(category)
}

func (s *CategoryService) DeleteCategory(id uuid.UUID) error {
	return s.repo.Delete(id)
}

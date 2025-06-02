package repositories

import (
	"backend/models/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CategoryRepository struct {
	DB *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *CategoryRepository {
	return &CategoryRepository{DB: db}
}

func (r *CategoryRepository) FindAll(isActive *bool) ([]models.Category, error) {
	var categories []models.Category
	query := r.DB

	if isActive != nil {
		query = query.Where("is_active = ?", *isActive)
	}

	result := query.Find(&categories)
	return categories, result.Error
}

func (r *CategoryRepository) FindByID(id uuid.UUID) (models.Category, error) {
	var category models.Category
	result := r.DB.First(&category, "id = ?", id)
	return category, result.Error
}

func (r *CategoryRepository) Create(category *models.Category) error {
	return r.DB.Create(category).Error
}

func (r *CategoryRepository) Update(category *models.Category) error {
	return r.DB.Save(category).Error
}

func (r *CategoryRepository) Delete(id uuid.UUID) error {
	return r.DB.Delete(&models.Category{}, "id = ?", id).Error
}

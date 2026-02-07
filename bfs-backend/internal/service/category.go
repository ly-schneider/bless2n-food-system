package service

import (
	"context"

	"backend/internal/generated/ent"
	"backend/internal/repository"

	"github.com/google/uuid"
)

type CategoryService interface {
	GetAll(ctx context.Context) ([]*ent.Category, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.Category, error)
	GetActive(ctx context.Context) ([]*ent.Category, error)
	List(ctx context.Context, limit, offset int) ([]*ent.Category, int64, error)
	Create(ctx context.Context, name string, position int) (*ent.Category, error)
	Update(ctx context.Context, id uuid.UUID, name string, position int, isActive bool) (*ent.Category, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type categoryService struct {
	repo repository.CategoryRepository
}

func NewCategoryService(repo repository.CategoryRepository) CategoryService {
	return &categoryService{repo: repo}
}

func (s *categoryService) GetAll(ctx context.Context) ([]*ent.Category, error) {
	return s.repo.GetAll(ctx)
}

func (s *categoryService) GetByID(ctx context.Context, id uuid.UUID) (*ent.Category, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *categoryService) GetActive(ctx context.Context) ([]*ent.Category, error) {
	return s.repo.GetAllActive(ctx)
}

func (s *categoryService) List(ctx context.Context, limit, offset int) ([]*ent.Category, int64, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *categoryService) Create(ctx context.Context, name string, position int) (*ent.Category, error) {
	return s.repo.Create(ctx, name, position, true)
}

func (s *categoryService) Update(ctx context.Context, id uuid.UUID, name string, position int, isActive bool) (*ent.Category, error) {
	return s.repo.Update(ctx, id, name, position, isActive)
}

func (s *categoryService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}

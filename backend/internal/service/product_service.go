package service

import (
	"context"
	"encoding/json"
	"fmt"

	"backend/internal/domain"
	"backend/internal/logger"
	"backend/internal/repository"

	"github.com/hibiken/asynq"
)

// ProductService defines the application-layer contract.
type ProductService interface {
	List(ctx context.Context) ([]domain.Product, error)
	Get(ctx context.Context, id uint) (domain.Product, error)
	Create(ctx context.Context, p *domain.Product) error
	Update(ctx context.Context, id uint, in *domain.Product) (domain.Product, error)
	Delete(ctx context.Context, id uint) error
}

type productService struct {
	repo   repository.ProductRepository
	asynqC *asynq.Client
}

// NewProductService is provided to Fx.
func NewProductService(r repository.ProductRepository, a *asynq.Client) ProductService {
	return &productService{repo: r, asynqC: a}
}

func (s *productService) List(ctx context.Context) ([]domain.Product, error) {
	logger.L.Info("Listing products")
	products, err := s.repo.List(ctx)
	if err != nil {
		logger.L.Errorw("Failed to list products", "error", err)
		return nil, err
	}
	logger.L.Infow("Successfully listed products", "count", len(products))
	return products, nil
}

func (s *productService) Get(ctx context.Context, id uint) (domain.Product, error) {
	logger.L.Infow("Getting product", "id", id)
	product, err := s.repo.Get(ctx, id)
	if err != nil {
		logger.L.Errorw("Failed to get product", "id", id, "error", err)
		return product, err
	}
	logger.L.Infow("Successfully retrieved product", "id", id, "name", product.Name)
	return product, nil
}

func (s *productService) Create(ctx context.Context, p *domain.Product) error {
	logger.L.Infow("Creating product", "name", p.Name, "price", p.Price)

	if err := s.repo.Create(ctx, p); err != nil {
		logger.L.Errorw("Failed to create product", "name", p.Name, "error", err)
		return err
	}

	logger.L.Infow("Product created successfully", "id", p.ID, "name", p.Name)

	// Enqueue background task
	payload, _ := json.Marshal(map[string]any{"product_id": p.ID})
	task := asynq.NewTask("product:created", payload)

	info, err := s.asynqC.Enqueue(task, asynq.Queue("default"))
	if err != nil {
		logger.L.Errorw("Failed to enqueue product created task",
			"product_id", p.ID,
			"task_type", task.Type(),
			"error", err)
		return fmt.Errorf("enqueue: %w", err)
	}

	logger.L.Infow("Enqueued product created task",
		"product_id", p.ID,
		"task_type", task.Type(),
		"task_id", info.ID)

	return nil
}

func (s *productService) Update(ctx context.Context, id uint, in *domain.Product) (domain.Product, error) {
	logger.L.Infow("Updating product", "id", id, "name", in.Name, "price", in.Price)

	// ensure the record exists
	p, err := s.repo.Get(ctx, id)
	if err != nil {
		logger.L.Errorw("Product not found for update", "id", id, "error", err)
		return p, err
	}

	// copy mutable fields
	oldName, oldPrice := p.Name, p.Price
	p.Name = in.Name
	p.Price = in.Price

	if err := s.repo.Update(ctx, &p); err != nil {
		logger.L.Errorw("Failed to update product", "id", id, "error", err)
		return p, err
	}

	logger.L.Infow("Product updated successfully",
		"id", id,
		"old_name", oldName,
		"new_name", p.Name,
		"old_price", oldPrice,
		"new_price", p.Price)

	return p, nil
}

func (s *productService) Delete(ctx context.Context, id uint) error {
	logger.L.Infow("Deleting product", "id", id)

	if err := s.repo.Delete(ctx, id); err != nil {
		logger.L.Errorw("Failed to delete product", "id", id, "error", err)
		return err
	}

	logger.L.Infow("Product deleted successfully", "id", id)
	return nil
}

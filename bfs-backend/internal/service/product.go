package service

import (
	"context"
	"errors"
	"fmt"

	"backend/internal/domain"
	"backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProductService interface {
	CreateProduct(ctx context.Context, req CreateProductRequest) (*CreateProductResponse, error)
	GetProduct(ctx context.Context, productID string) (*GetProductResponse, error)
	UpdateProduct(ctx context.Context, productID string, req UpdateProductRequest) (*UpdateProductResponse, error)
	DeleteProduct(ctx context.Context, productID string) (*DeleteProductResponse, error)
	ListProducts(ctx context.Context, categoryID *string, activeOnly bool, limit, offset int) (*ListProductsResponse, error)
	SetProductActive(ctx context.Context, productID string, isActive bool) (*SetProductActiveResponse, error)
	UpdateProductStock(ctx context.Context, productID string, req UpdateProductStockRequest) (*UpdateProductStockResponse, error)
	CreateProductBundle(ctx context.Context, req CreateProductBundleRequest) (*CreateProductBundleResponse, error)
	UpdateProductBundle(ctx context.Context, bundleID string, req UpdateProductBundleRequest) (*UpdateProductBundleResponse, error)
	AssignProductToStations(ctx context.Context, productID string, stationIDs []primitive.ObjectID) (*AssignProductToStationsResponse, error)
}

type CreateProductRequest struct {
	CategoryID string             `json:"category_id" validate:"required"`
	Type       domain.ProductType `json:"type" validate:"required,oneof=simple bundle"`
	Name       string             `json:"name" validate:"required"`
	Image      *string            `json:"image,omitempty"`
	Price      float64            `json:"price" validate:"required,gte=0"`
}

type CreateProductResponse struct {
	Product ProductDTO `json:"product"`
	Message string     `json:"message"`
	Success bool       `json:"success"`
}

type GetProductResponse struct {
	Product ProductDTO `json:"product"`
}

type UpdateProductRequest struct {
	CategoryID *string             `json:"category_id,omitempty"`
	Type       *domain.ProductType `json:"type,omitempty" validate:"omitempty,oneof=simple bundle"`
	Name       *string             `json:"name,omitempty"`
	Image      *string             `json:"image,omitempty"`
	Price      *float64            `json:"price,omitempty" validate:"omitempty,gte=0"`
}

type UpdateProductResponse struct {
	Product ProductDTO `json:"product"`
	Message string     `json:"message"`
	Success bool       `json:"success"`
}

type DeleteProductResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type ListProductsResponse struct {
	Products []ProductDTO `json:"products"`
	Total    int          `json:"total"`
}

type SetProductActiveResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type UpdateProductStockRequest struct {
	Action string `json:"action" validate:"required,oneof=add subtract set"`
	Amount int    `json:"amount" validate:"required,gt=0"`
}

type UpdateProductStockResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type CreateProductBundleRequest struct {
	CategoryID string                      `json:"category_id" validate:"required"`
	Name       string                      `json:"name" validate:"required"`
	Image      *string                     `json:"image,omitempty"`
	Price      float64                     `json:"price" validate:"required,gte=0"`
	Components []ProductBundleComponentDTO `json:"components" validate:"required,dive"`
}

type CreateProductBundleResponse struct {
	Bundle     ProductDTO                  `json:"bundle"`
	Components []ProductBundleComponentDTO `json:"components"`
	Message    string                      `json:"message"`
	Success    bool                        `json:"success"`
}

type UpdateProductBundleRequest struct {
	CategoryID *string                     `json:"category_id,omitempty"`
	Name       *string                     `json:"name,omitempty"`
	Image      *string                     `json:"image,omitempty"`
	Price      *float64                    `json:"price,omitempty" validate:"omitempty,gte=0"`
	Components []ProductBundleComponentDTO `json:"components,omitempty" validate:"omitempty,dive"`
}

type UpdateProductBundleResponse struct {
	Bundle     ProductDTO                  `json:"bundle"`
	Components []ProductBundleComponentDTO `json:"components"`
	Message    string                      `json:"message"`
	Success    bool                        `json:"success"`
}

type ProductBundleComponentDTO struct {
	ComponentProductID string `json:"component_product_id" validate:"required"`
	Quantity           int    `json:"quantity" validate:"required,gt=0"`
}

type AssignProductToStationsResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type ProductDTO struct {
	ID         string  `json:"id"`
	CategoryID string  `json:"category_id"`
	Type       string  `json:"type"`
	Name       string  `json:"name"`
	Image      *string `json:"image,omitempty"`
	Price      float64 `json:"price"`
	IsActive   bool    `json:"is_active"`
	CreatedAt  string  `json:"created_at"`
	UpdatedAt  string  `json:"updated_at"`
}

type productService struct {
	productRepo         repository.ProductRepository
	categoryRepo        repository.CategoryRepository
	bundleComponentRepo repository.ProductBundleComponentRepository
	stationProductRepo  repository.StationProductRepository
}

func NewProductService(
	productRepo repository.ProductRepository,
	categoryRepo repository.CategoryRepository,
	bundleComponentRepo repository.ProductBundleComponentRepository,
	stationProductRepo repository.StationProductRepository,
) ProductService {
	return &productService{
		productRepo:         productRepo,
		categoryRepo:        categoryRepo,
		bundleComponentRepo: bundleComponentRepo,
		stationProductRepo:  stationProductRepo,
	}
}

func (s *productService) CreateProduct(ctx context.Context, req CreateProductRequest) (*CreateProductResponse, error) {
	categoryID, err := primitive.ObjectIDFromHex(req.CategoryID)
	if err != nil {
		return nil, errors.New("invalid category ID format")
	}

	category, err := s.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate category: %w", err)
	}
	if category == nil {
		return nil, errors.New("category not found")
	}

	product := &domain.Product{
		CategoryID: categoryID,
		Type:       req.Type,
		Name:       req.Name,
		Image:      req.Image,
		Price:      req.Price,
	}

	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return &CreateProductResponse{
		Product: s.toProductDTO(product),
		Message: "Product created successfully",
		Success: true,
	}, nil
}

func (s *productService) GetProduct(ctx context.Context, productID string) (*GetProductResponse, error) {
	objectID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return nil, errors.New("invalid product ID format")
	}

	product, err := s.productRepo.GetByID(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	if product == nil {
		return nil, errors.New("product not found")
	}

	return &GetProductResponse{
		Product: s.toProductDTO(product),
	}, nil
}

func (s *productService) UpdateProduct(ctx context.Context, productID string, req UpdateProductRequest) (*UpdateProductResponse, error) {
	objectID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return nil, errors.New("invalid product ID format")
	}

	product, err := s.productRepo.GetByID(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	if product == nil {
		return nil, errors.New("product not found")
	}

	if req.CategoryID != nil {
		categoryID, err := primitive.ObjectIDFromHex(*req.CategoryID)
		if err != nil {
			return nil, errors.New("invalid category ID format")
		}
		category, err := s.categoryRepo.GetByID(ctx, categoryID)
		if err != nil {
			return nil, fmt.Errorf("failed to validate category: %w", err)
		}
		if category == nil {
			return nil, errors.New("category not found")
		}
		product.CategoryID = categoryID
	}

	if req.Type != nil {
		product.Type = *req.Type
	}
	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.Image != nil {
		product.Image = req.Image
	}
	if req.Price != nil {
		product.Price = *req.Price
	}

	if err := s.productRepo.Update(ctx, product); err != nil {
		return nil, fmt.Errorf("failed to update product: %w", err)
	}

	return &UpdateProductResponse{
		Product: s.toProductDTO(product),
		Message: "Product updated successfully",
		Success: true,
	}, nil
}

func (s *productService) DeleteProduct(ctx context.Context, productID string) (*DeleteProductResponse, error) {
	objectID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return nil, errors.New("invalid product ID format")
	}

	product, err := s.productRepo.GetByID(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	if product == nil {
		return nil, errors.New("product not found")
	}

	if err := s.productRepo.Delete(ctx, objectID); err != nil {
		return nil, fmt.Errorf("failed to delete product: %w", err)
	}

	return &DeleteProductResponse{
		Message: "Product deleted successfully",
		Success: true,
	}, nil
}

func (s *productService) ListProducts(ctx context.Context, categoryID *string, activeOnly bool, limit, offset int) (*ListProductsResponse, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	var products []*domain.Product
	var err error

	if categoryID != nil {
		categoryObjID, err := primitive.ObjectIDFromHex(*categoryID)
		if err != nil {
			return nil, errors.New("invalid category ID format")
		}
		products, err = s.productRepo.GetByCategoryID(ctx, categoryObjID, activeOnly, limit, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to list products: %w", err)
		}
	} else {
		// List all products regardless of type
		var simpleProducts []*domain.Product
		simpleProducts, err = s.productRepo.GetByType(ctx, domain.ProductTypeSimple, limit, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to get simple products: %w", err)
		}
		var bundleProducts []*domain.Product
		bundleProducts, err = s.productRepo.GetByType(ctx, domain.ProductTypeBundle, limit, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to get bundle products: %w", err)
		}
		products = append(simpleProducts, bundleProducts...)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list products: %w", err)
	}

	productDTOs := make([]ProductDTO, len(products))
	for i, product := range products {
		productDTOs[i] = s.toProductDTO(product)
	}

	return &ListProductsResponse{
		Products: productDTOs,
		Total:    len(productDTOs),
	}, nil
}

func (s *productService) SetProductActive(ctx context.Context, productID string, isActive bool) (*SetProductActiveResponse, error) {
	objectID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return nil, errors.New("invalid product ID format")
	}

	product, err := s.productRepo.GetByID(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	if product == nil {
		return nil, errors.New("product not found")
	}

	if err := s.productRepo.SetActive(ctx, objectID, isActive); err != nil {
		return nil, fmt.Errorf("failed to update product status: %w", err)
	}

	var action string
	if isActive {
		action = "activated"
	} else {
		action = "deactivated"
	}

	return &SetProductActiveResponse{
		Message: fmt.Sprintf("Product %s successfully", action),
		Success: true,
	}, nil
}

func (s *productService) UpdateProductStock(ctx context.Context, productID string, req UpdateProductStockRequest) (*UpdateProductStockResponse, error) {
	return &UpdateProductStockResponse{
		Message: "Stock update functionality not yet implemented",
		Success: false,
	}, nil
}

func (s *productService) CreateProductBundle(ctx context.Context, req CreateProductBundleRequest) (*CreateProductBundleResponse, error) {
	categoryID, err := primitive.ObjectIDFromHex(req.CategoryID)
	if err != nil {
		return nil, errors.New("invalid category ID format")
	}

	category, err := s.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate category: %w", err)
	}
	if category == nil {
		return nil, errors.New("category not found")
	}

	bundle := &domain.Product{
		CategoryID: categoryID,
		Type:       domain.ProductTypeBundle,
		Name:       req.Name,
		Image:      req.Image,
		Price:      req.Price,
	}

	if err := s.productRepo.Create(ctx, bundle); err != nil {
		return nil, fmt.Errorf("failed to create bundle: %w", err)
	}

	for _, comp := range req.Components {
		componentID, err := primitive.ObjectIDFromHex(comp.ComponentProductID)
		if err != nil {
			return nil, fmt.Errorf("invalid component product ID format: %s", comp.ComponentProductID)
		}

		component, err := s.productRepo.GetByID(ctx, componentID)
		if err != nil {
			return nil, fmt.Errorf("failed to validate component product: %w", err)
		}
		if component == nil {
			return nil, fmt.Errorf("component product not found: %s", comp.ComponentProductID)
		}

		bundleComponent := &domain.ProductBundleComponent{
			BundleID:           bundle.ID,
			ComponentProductID: componentID,
			Quantity:           comp.Quantity,
		}

		if err := s.bundleComponentRepo.Create(ctx, bundleComponent); err != nil {
			return nil, fmt.Errorf("failed to create bundle component: %w", err)
		}
	}

	return &CreateProductBundleResponse{
		Bundle:     s.toProductDTO(bundle),
		Components: req.Components,
		Message:    "Product bundle created successfully",
		Success:    true,
	}, nil
}

func (s *productService) UpdateProductBundle(ctx context.Context, bundleID string, req UpdateProductBundleRequest) (*UpdateProductBundleResponse, error) {
	objectID, err := primitive.ObjectIDFromHex(bundleID)
	if err != nil {
		return nil, errors.New("invalid bundle ID format")
	}

	bundle, err := s.productRepo.GetByID(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bundle: %w", err)
	}
	if bundle == nil {
		return nil, errors.New("bundle not found")
	}
	if bundle.Type != domain.ProductTypeBundle {
		return nil, errors.New("product is not a bundle")
	}

	if req.CategoryID != nil {
		categoryID, err := primitive.ObjectIDFromHex(*req.CategoryID)
		if err != nil {
			return nil, errors.New("invalid category ID format")
		}
		category, err := s.categoryRepo.GetByID(ctx, categoryID)
		if err != nil {
			return nil, fmt.Errorf("failed to validate category: %w", err)
		}
		if category == nil {
			return nil, errors.New("category not found")
		}
		bundle.CategoryID = categoryID
	}

	if req.Name != nil {
		bundle.Name = *req.Name
	}
	if req.Image != nil {
		bundle.Image = req.Image
	}
	if req.Price != nil {
		bundle.Price = *req.Price
	}

	if err := s.productRepo.Update(ctx, bundle); err != nil {
		return nil, fmt.Errorf("failed to update bundle: %w", err)
	}

	if req.Components != nil {
		if err := s.bundleComponentRepo.DeleteByBundleID(ctx, objectID); err != nil {
			return nil, fmt.Errorf("failed to remove existing components: %w", err)
		}

		for _, comp := range req.Components {
			componentID, err := primitive.ObjectIDFromHex(comp.ComponentProductID)
			if err != nil {
				return nil, fmt.Errorf("invalid component product ID format: %s", comp.ComponentProductID)
			}

			component, err := s.productRepo.GetByID(ctx, componentID)
			if err != nil {
				return nil, fmt.Errorf("failed to validate component product: %w", err)
			}
			if component == nil {
				return nil, fmt.Errorf("component product not found: %s", comp.ComponentProductID)
			}

			bundleComponent := &domain.ProductBundleComponent{
				BundleID:           bundle.ID,
				ComponentProductID: componentID,
				Quantity:           comp.Quantity,
			}

			if err := s.bundleComponentRepo.Create(ctx, bundleComponent); err != nil {
				return nil, fmt.Errorf("failed to create bundle component: %w", err)
			}
		}
	}

	return &UpdateProductBundleResponse{
		Bundle:     s.toProductDTO(bundle),
		Components: req.Components,
		Message:    "Product bundle updated successfully",
		Success:    true,
	}, nil
}

func (s *productService) AssignProductToStations(ctx context.Context, productID string, stationIDs []primitive.ObjectID) (*AssignProductToStationsResponse, error) {
	objectID, err := primitive.ObjectIDFromHex(productID)
	if err != nil {
		return nil, errors.New("invalid product ID format")
	}

	product, err := s.productRepo.GetByID(ctx, objectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}
	if product == nil {
		return nil, errors.New("product not found")
	}

	for _, stationID := range stationIDs {
		stationProduct := &domain.StationProduct{
			StationID: stationID,
			ProductID: objectID,
		}

		if err := s.stationProductRepo.Create(ctx, stationProduct); err != nil {
			return nil, fmt.Errorf("failed to assign product to station %s: %w", stationID.Hex(), err)
		}
	}

	return &AssignProductToStationsResponse{
		Message: fmt.Sprintf("Product assigned to %d stations successfully", len(stationIDs)),
		Success: true,
	}, nil
}

func (s *productService) toProductDTO(product *domain.Product) ProductDTO {
	return ProductDTO{
		ID:         product.ID.Hex(),
		CategoryID: product.CategoryID.Hex(),
		Type:       string(product.Type),
		Name:       product.Name,
		Image:      product.Image,
		Price:      product.Price,
		IsActive:   product.IsActive,
		CreatedAt:  product.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  product.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

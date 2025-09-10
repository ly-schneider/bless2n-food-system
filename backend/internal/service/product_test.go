package service

import (
    "errors"
    "testing"

	"backend/internal/domain"
	"backend/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func setupProductService() (*productService, *testutil.MockProductRepository, *testutil.MockCategoryRepository, *testutil.MockProductBundleComponentRepository, *testutil.MockStationProductRepository) {
	mockProductRepo := &testutil.MockProductRepository{}
	mockCategoryRepo := &testutil.MockCategoryRepository{}
	mockBundleComponentRepo := &testutil.MockProductBundleComponentRepository{}
	mockStationProductRepo := &testutil.MockStationProductRepository{}

	service := &productService{
		productRepo:         mockProductRepo,
		categoryRepo:        mockCategoryRepo,
		bundleComponentRepo: mockBundleComponentRepo,
		stationProductRepo:  mockStationProductRepo,
	}

	return service, mockProductRepo, mockCategoryRepo, mockBundleComponentRepo, mockStationProductRepo
}

func TestProductService_CreateProduct(t *testing.T) {
	categoryID := primitive.NewObjectID()

	tests := []struct {
		name          string
		request       CreateProductRequest
		setupMocks    func(*testutil.MockProductRepository, *testutil.MockCategoryRepository, *testutil.MockProductBundleComponentRepository, *testutil.MockStationProductRepository)
		expectedError string
		expectSuccess bool
	}{
		{
			name: "successful product creation",
			request: CreateProductRequest{
				CategoryID: categoryID.Hex(),
				Type:       domain.ProductTypeSimple,
				Name:       "Test Product",
				Price:      19.99,
			},
			setupMocks: func(productRepo *testutil.MockProductRepository, categoryRepo *testutil.MockCategoryRepository, bundleRepo *testutil.MockProductBundleComponentRepository, stationRepo *testutil.MockStationProductRepository) {
				category := testutil.CreateTestCategory("Test Category")
				category.ID = categoryID

				categoryRepo.On("GetByID", mock.Anything, categoryID).Return(category, nil)
				productRepo.On("Create", mock.Anything, mock.MatchedBy(func(p *domain.Product) bool {
					return p.CategoryID == categoryID && p.Type == domain.ProductTypeSimple && p.Name == "Test Product" && p.Price == 19.99
				})).Return(nil)
			},
			expectSuccess: true,
		},
		{
			name: "invalid category ID format",
			request: CreateProductRequest{
				CategoryID: "invalid-id",
				Type:       domain.ProductTypeSimple,
				Name:       "Test Product",
				Price:      19.99,
			},
			setupMocks:    func(productRepo *testutil.MockProductRepository, categoryRepo *testutil.MockCategoryRepository, bundleRepo *testutil.MockProductBundleComponentRepository, stationRepo *testutil.MockStationProductRepository) {},
			expectedError: "invalid category ID format",
		},
		{
			name: "category not found",
			request: CreateProductRequest{
				CategoryID: categoryID.Hex(),
				Type:       domain.ProductTypeSimple,
				Name:       "Test Product",
				Price:      19.99,
			},
			setupMocks: func(productRepo *testutil.MockProductRepository, categoryRepo *testutil.MockCategoryRepository, bundleRepo *testutil.MockProductBundleComponentRepository, stationRepo *testutil.MockStationProductRepository) {
				categoryRepo.On("GetByID", mock.Anything, categoryID).Return(nil, nil)
			},
			expectedError: "category not found",
		},
		{
			name: "database error when validating category",
			request: CreateProductRequest{
				CategoryID: categoryID.Hex(),
				Type:       domain.ProductTypeSimple,
				Name:       "Test Product",
				Price:      19.99,
			},
			setupMocks: func(productRepo *testutil.MockProductRepository, categoryRepo *testutil.MockCategoryRepository, bundleRepo *testutil.MockProductBundleComponentRepository, stationRepo *testutil.MockStationProductRepository) {
				categoryRepo.On("GetByID", mock.Anything, categoryID).Return(nil, errors.New("db error"))
			},
			expectedError: "failed to validate category",
		},
		{
			name: "database error when creating product",
			request: CreateProductRequest{
				CategoryID: categoryID.Hex(),
				Type:       domain.ProductTypeSimple,
				Name:       "Test Product",
				Price:      19.99,
			},
			setupMocks: func(productRepo *testutil.MockProductRepository, categoryRepo *testutil.MockCategoryRepository, bundleRepo *testutil.MockProductBundleComponentRepository, stationRepo *testutil.MockStationProductRepository) {
				category := testutil.CreateTestCategory("Test Category")
				category.ID = categoryID

				categoryRepo.On("GetByID", mock.Anything, categoryID).Return(category, nil)
				productRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedError: "failed to create product",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, productRepo, categoryRepo, bundleRepo, stationRepo := setupProductService()
			tt.setupMocks(productRepo, categoryRepo, bundleRepo, stationRepo)

			ctx := testutil.TestContext()
			response, err := service.CreateProduct(ctx, tt.request)

			if tt.expectedError != "" {
				testutil.AssertErrorContains(t, err, tt.expectedError)
				assert.Nil(t, response)
			} else {
				testutil.AssertNoError(t, err)
				require.NotNil(t, response)
				if tt.expectSuccess {
					assert.True(t, response.Success)
					assert.Equal(t, "Product created successfully", response.Message)
					assert.NotEmpty(t, response.Product.ID)
				}
			}

			productRepo.AssertExpectations(t)
			categoryRepo.AssertExpectations(t)
			bundleRepo.AssertExpectations(t)
			stationRepo.AssertExpectations(t)
		})
	}
}

func TestProductService_GetProduct(t *testing.T) {
	productID := primitive.NewObjectID()

	tests := []struct {
		name          string
		productID     string
		setupMocks    func(*testutil.MockProductRepository, *testutil.MockCategoryRepository, *testutil.MockProductBundleComponentRepository, *testutil.MockStationProductRepository)
		expectedError string
		expectProduct bool
	}{
		{
			name:      "successful product retrieval",
			productID: productID.Hex(),
			setupMocks: func(productRepo *testutil.MockProductRepository, categoryRepo *testutil.MockCategoryRepository, bundleRepo *testutil.MockProductBundleComponentRepository, stationRepo *testutil.MockStationProductRepository) {
				product := testutil.CreateTestProduct(primitive.NewObjectID(), "Test Product", 19.99)
				product.ID = productID

				productRepo.On("GetByID", mock.Anything, productID).Return(product, nil)
			},
			expectProduct: true,
		},
		{
			name:          "invalid product ID format",
			productID:     "invalid-id",
			setupMocks:    func(productRepo *testutil.MockProductRepository, categoryRepo *testutil.MockCategoryRepository, bundleRepo *testutil.MockProductBundleComponentRepository, stationRepo *testutil.MockStationProductRepository) {},
			expectedError: "invalid product ID format",
		},
		{
			name:      "product not found",
			productID: productID.Hex(),
			setupMocks: func(productRepo *testutil.MockProductRepository, categoryRepo *testutil.MockCategoryRepository, bundleRepo *testutil.MockProductBundleComponentRepository, stationRepo *testutil.MockStationProductRepository) {
				productRepo.On("GetByID", mock.Anything, productID).Return(nil, nil)
			},
			expectedError: "product not found",
		},
		{
			name:      "database error",
			productID: productID.Hex(),
			setupMocks: func(productRepo *testutil.MockProductRepository, categoryRepo *testutil.MockCategoryRepository, bundleRepo *testutil.MockProductBundleComponentRepository, stationRepo *testutil.MockStationProductRepository) {
				productRepo.On("GetByID", mock.Anything, productID).Return(nil, errors.New("db error"))
			},
			expectedError: "failed to get product",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, productRepo, categoryRepo, bundleRepo, stationRepo := setupProductService()
			tt.setupMocks(productRepo, categoryRepo, bundleRepo, stationRepo)

			ctx := testutil.TestContext()
			response, err := service.GetProduct(ctx, tt.productID)

			if tt.expectedError != "" {
				testutil.AssertErrorContains(t, err, tt.expectedError)
				assert.Nil(t, response)
			} else {
				testutil.AssertNoError(t, err)
				require.NotNil(t, response)
				if tt.expectProduct {
					assert.NotEmpty(t, response.Product.ID)
					assert.NotEmpty(t, response.Product.Name)
				}
			}

			productRepo.AssertExpectations(t)
			categoryRepo.AssertExpectations(t)
			bundleRepo.AssertExpectations(t)
			stationRepo.AssertExpectations(t)
		})
	}
}

func TestProductService_UpdateProduct(t *testing.T) {
	productID := primitive.NewObjectID()
	categoryID := primitive.NewObjectID()
	newCategoryID := primitive.NewObjectID()

	tests := []struct {
		name          string
		productID     string
		request       UpdateProductRequest
		setupMocks    func(*testutil.MockProductRepository, *testutil.MockCategoryRepository, *testutil.MockProductBundleComponentRepository, *testutil.MockStationProductRepository)
		expectedError string
		expectSuccess bool
	}{
		{
			name:      "successful product update",
			productID: productID.Hex(),
			request: UpdateProductRequest{
				Name:  testutil.StringPtr("Updated Product"),
				Price: testutil.Float64Ptr(29.99),
			},
			setupMocks: func(productRepo *testutil.MockProductRepository, categoryRepo *testutil.MockCategoryRepository, bundleRepo *testutil.MockProductBundleComponentRepository, stationRepo *testutil.MockStationProductRepository) {
				product := testutil.CreateTestProduct(categoryID, "Test Product", 19.99)
				product.ID = productID

				productRepo.On("GetByID", mock.Anything, productID).Return(product, nil)
				productRepo.On("Update", mock.Anything, mock.MatchedBy(func(p *domain.Product) bool {
					return p.Name == "Updated Product" && p.Price == 29.99
				})).Return(nil)
			},
			expectSuccess: true,
		},
		{
			name:      "successful product update with category change",
			productID: productID.Hex(),
			request: UpdateProductRequest{
				CategoryID: testutil.StringPtr(newCategoryID.Hex()),
				Name:       testutil.StringPtr("Updated Product"),
			},
			setupMocks: func(productRepo *testutil.MockProductRepository, categoryRepo *testutil.MockCategoryRepository, bundleRepo *testutil.MockProductBundleComponentRepository, stationRepo *testutil.MockStationProductRepository) {
				product := testutil.CreateTestProduct(categoryID, "Test Product", 19.99)
				product.ID = productID

				newCategory := testutil.CreateTestCategory("New Category")
				newCategory.ID = newCategoryID

				productRepo.On("GetByID", mock.Anything, productID).Return(product, nil)
				categoryRepo.On("GetByID", mock.Anything, newCategoryID).Return(newCategory, nil)
				productRepo.On("Update", mock.Anything, mock.MatchedBy(func(p *domain.Product) bool {
					return p.CategoryID == newCategoryID && p.Name == "Updated Product"
				})).Return(nil)
			},
			expectSuccess: true,
		},
		{
			name:          "invalid product ID format",
			productID:     "invalid-id",
			request:       UpdateProductRequest{},
			setupMocks:    func(productRepo *testutil.MockProductRepository, categoryRepo *testutil.MockCategoryRepository, bundleRepo *testutil.MockProductBundleComponentRepository, stationRepo *testutil.MockStationProductRepository) {},
			expectedError: "invalid product ID format",
		},
		{
			name:      "product not found",
			productID: productID.Hex(),
			request:   UpdateProductRequest{},
			setupMocks: func(productRepo *testutil.MockProductRepository, categoryRepo *testutil.MockCategoryRepository, bundleRepo *testutil.MockProductBundleComponentRepository, stationRepo *testutil.MockStationProductRepository) {
				productRepo.On("GetByID", mock.Anything, productID).Return(nil, nil)
			},
			expectedError: "product not found",
		},
		{
			name:      "invalid category ID format in update",
			productID: productID.Hex(),
			request: UpdateProductRequest{
				CategoryID: testutil.StringPtr("invalid-id"),
			},
			setupMocks: func(productRepo *testutil.MockProductRepository, categoryRepo *testutil.MockCategoryRepository, bundleRepo *testutil.MockProductBundleComponentRepository, stationRepo *testutil.MockStationProductRepository) {
				product := testutil.CreateTestProduct(categoryID, "Test Product", 19.99)
				product.ID = productID

				productRepo.On("GetByID", mock.Anything, productID).Return(product, nil)
			},
			expectedError: "invalid category ID format",
		},
		{
			name:      "category not found in update",
			productID: productID.Hex(),
			request: UpdateProductRequest{
				CategoryID: testutil.StringPtr(newCategoryID.Hex()),
			},
			setupMocks: func(productRepo *testutil.MockProductRepository, categoryRepo *testutil.MockCategoryRepository, bundleRepo *testutil.MockProductBundleComponentRepository, stationRepo *testutil.MockStationProductRepository) {
				product := testutil.CreateTestProduct(categoryID, "Test Product", 19.99)
				product.ID = productID

				productRepo.On("GetByID", mock.Anything, productID).Return(product, nil)
				categoryRepo.On("GetByID", mock.Anything, newCategoryID).Return(nil, nil)
			},
			expectedError: "category not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, productRepo, categoryRepo, bundleRepo, stationRepo := setupProductService()
			tt.setupMocks(productRepo, categoryRepo, bundleRepo, stationRepo)

			ctx := testutil.TestContext()
			response, err := service.UpdateProduct(ctx, tt.productID, tt.request)

			if tt.expectedError != "" {
				testutil.AssertErrorContains(t, err, tt.expectedError)
				assert.Nil(t, response)
			} else {
				testutil.AssertNoError(t, err)
				require.NotNil(t, response)
				if tt.expectSuccess {
					assert.True(t, response.Success)
					assert.Equal(t, "Product updated successfully", response.Message)
				}
			}

			productRepo.AssertExpectations(t)
			categoryRepo.AssertExpectations(t)
			bundleRepo.AssertExpectations(t)
			stationRepo.AssertExpectations(t)
		})
	}
}

func TestProductService_DeleteProduct(t *testing.T) {
	productID := primitive.NewObjectID()

	tests := []struct {
		name          string
		productID     string
		setupMocks    func(*testutil.MockProductRepository, *testutil.MockCategoryRepository, *testutil.MockProductBundleComponentRepository, *testutil.MockStationProductRepository)
		expectedError string
		expectSuccess bool
	}{
		{
			name:      "successful product deletion",
			productID: productID.Hex(),
			setupMocks: func(productRepo *testutil.MockProductRepository, categoryRepo *testutil.MockCategoryRepository, bundleRepo *testutil.MockProductBundleComponentRepository, stationRepo *testutil.MockStationProductRepository) {
				product := testutil.CreateTestProduct(primitive.NewObjectID(), "Test Product", 19.99)
				product.ID = productID

				productRepo.On("GetByID", mock.Anything, productID).Return(product, nil)
				productRepo.On("Delete", mock.Anything, productID).Return(nil)
			},
			expectSuccess: true,
		},
		{
			name:          "invalid product ID format",
			productID:     "invalid-id",
			setupMocks:    func(productRepo *testutil.MockProductRepository, categoryRepo *testutil.MockCategoryRepository, bundleRepo *testutil.MockProductBundleComponentRepository, stationRepo *testutil.MockStationProductRepository) {},
			expectedError: "invalid product ID format",
		},
		{
			name:      "product not found",
			productID: productID.Hex(),
			setupMocks: func(productRepo *testutil.MockProductRepository, categoryRepo *testutil.MockCategoryRepository, bundleRepo *testutil.MockProductBundleComponentRepository, stationRepo *testutil.MockStationProductRepository) {
				productRepo.On("GetByID", mock.Anything, productID).Return(nil, nil)
			},
			expectedError: "product not found",
		},
		{
			name:      "database error when deleting",
			productID: productID.Hex(),
			setupMocks: func(productRepo *testutil.MockProductRepository, categoryRepo *testutil.MockCategoryRepository, bundleRepo *testutil.MockProductBundleComponentRepository, stationRepo *testutil.MockStationProductRepository) {
				product := testutil.CreateTestProduct(primitive.NewObjectID(), "Test Product", 19.99)
				product.ID = productID

				productRepo.On("GetByID", mock.Anything, productID).Return(product, nil)
				productRepo.On("Delete", mock.Anything, productID).Return(errors.New("db error"))
			},
			expectedError: "failed to delete product",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, productRepo, categoryRepo, bundleRepo, stationRepo := setupProductService()
			tt.setupMocks(productRepo, categoryRepo, bundleRepo, stationRepo)

			ctx := testutil.TestContext()
			response, err := service.DeleteProduct(ctx, tt.productID)

			if tt.expectedError != "" {
				testutil.AssertErrorContains(t, err, tt.expectedError)
				assert.Nil(t, response)
			} else {
				testutil.AssertNoError(t, err)
				require.NotNil(t, response)
				if tt.expectSuccess {
					assert.True(t, response.Success)
					assert.Equal(t, "Product deleted successfully", response.Message)
				}
			}

			productRepo.AssertExpectations(t)
			categoryRepo.AssertExpectations(t)
			bundleRepo.AssertExpectations(t)
			stationRepo.AssertExpectations(t)
		})
	}
}

func TestProductService_SetProductActive(t *testing.T) {
	productID := primitive.NewObjectID()

	tests := []struct {
		name          string
		productID     string
		isActive      bool
		setupMocks    func(*testutil.MockProductRepository, *testutil.MockCategoryRepository, *testutil.MockProductBundleComponentRepository, *testutil.MockStationProductRepository)
		expectedError string
		expectSuccess bool
		expectedMsg   string
	}{
		{
			name:      "successful product activation",
			productID: productID.Hex(),
			isActive:  true,
			setupMocks: func(productRepo *testutil.MockProductRepository, categoryRepo *testutil.MockCategoryRepository, bundleRepo *testutil.MockProductBundleComponentRepository, stationRepo *testutil.MockStationProductRepository) {
				product := testutil.CreateTestProduct(primitive.NewObjectID(), "Test Product", 19.99)
				product.ID = productID
				product.IsActive = false

				productRepo.On("GetByID", mock.Anything, productID).Return(product, nil)
				productRepo.On("SetActive", mock.Anything, productID, true).Return(nil)
			},
			expectSuccess: true,
			expectedMsg:   "Product activated successfully",
		},
		{
			name:      "successful product deactivation",
			productID: productID.Hex(),
			isActive:  false,
			setupMocks: func(productRepo *testutil.MockProductRepository, categoryRepo *testutil.MockCategoryRepository, bundleRepo *testutil.MockProductBundleComponentRepository, stationRepo *testutil.MockStationProductRepository) {
				product := testutil.CreateTestProduct(primitive.NewObjectID(), "Test Product", 19.99)
				product.ID = productID
				product.IsActive = true

				productRepo.On("GetByID", mock.Anything, productID).Return(product, nil)
				productRepo.On("SetActive", mock.Anything, productID, false).Return(nil)
			},
			expectSuccess: true,
			expectedMsg:   "Product deactivated successfully",
		},
		{
			name:          "invalid product ID format",
			productID:     "invalid-id",
			isActive:      true,
			setupMocks:    func(productRepo *testutil.MockProductRepository, categoryRepo *testutil.MockCategoryRepository, bundleRepo *testutil.MockProductBundleComponentRepository, stationRepo *testutil.MockStationProductRepository) {},
			expectedError: "invalid product ID format",
		},
		{
			name:      "product not found",
			productID: productID.Hex(),
			isActive:  true,
			setupMocks: func(productRepo *testutil.MockProductRepository, categoryRepo *testutil.MockCategoryRepository, bundleRepo *testutil.MockProductBundleComponentRepository, stationRepo *testutil.MockStationProductRepository) {
				productRepo.On("GetByID", mock.Anything, productID).Return(nil, nil)
			},
			expectedError: "product not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, productRepo, categoryRepo, bundleRepo, stationRepo := setupProductService()
			tt.setupMocks(productRepo, categoryRepo, bundleRepo, stationRepo)

			ctx := testutil.TestContext()
			response, err := service.SetProductActive(ctx, tt.productID, tt.isActive)

			if tt.expectedError != "" {
				testutil.AssertErrorContains(t, err, tt.expectedError)
				assert.Nil(t, response)
			} else {
				testutil.AssertNoError(t, err)
				require.NotNil(t, response)
				if tt.expectSuccess {
					assert.True(t, response.Success)
					assert.Equal(t, tt.expectedMsg, response.Message)
				}
			}

			productRepo.AssertExpectations(t)
			categoryRepo.AssertExpectations(t)
			bundleRepo.AssertExpectations(t)
			stationRepo.AssertExpectations(t)
		})
	}
}

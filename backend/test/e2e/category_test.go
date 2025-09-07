package e2e

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"backend/internal/domain"
	"backend/internal/service"
	"backend/test/fixtures"
	"backend/test/helpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type CategoryTestSuite struct {
	suite.Suite
	server  *helpers.TestServer
	client  *helpers.HTTPClient
	testDB  *helpers.TestDB
	ctx     context.Context
	baseURL string
}

func (suite *CategoryTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	server, err := helpers.NewTestServer()
	require.NoError(suite.T(), err, "Failed to create test server")

	err = server.Start()
	require.NoError(suite.T(), err, "Failed to start test server")

	suite.server = server
	suite.baseURL = server.BaseURL

	suite.client = helpers.NewHTTPClient(suite.baseURL)

	mongoURI := os.Getenv("MONGO_URI_LOCAL")
	dbName := os.Getenv("MONGO_DATABASE")

	testDB, err := helpers.NewTestDB(mongoURI, dbName)
	require.NoError(suite.T(), err, "Failed to create test database")
	suite.testDB = testDB
}

func (suite *CategoryTestSuite) TearDownSuite() {
	if suite.testDB != nil {
		suite.testDB.Close()
	}
	if suite.server != nil {
		suite.server.Stop()
	}
}

func (suite *CategoryTestSuite) SetupTest() {
	err := suite.testDB.CleanAll()
	require.NoError(suite.T(), err, "Failed to clean database")

	suite.server.EmailMock.Reset()

	suite.client.ClearAuth()
}

func (suite *CategoryTestSuite) TearDownTest() {
	suite.testDB.CleanOTPs()
}

// Helper method to get authentication token for a user
func (suite *CategoryTestSuite) getAuthToken(email string) string {
	// Request OTP
	otpReq := map[string]any{"email": email}
	_, err := suite.client.POST(suite.ctx, "/v1/auth/request-otp", otpReq)
	require.NoError(suite.T(), err, "Should be able to request OTP")

	// Get OTP from email mock
	sentOTP := suite.server.EmailMock.GetLastSentOTP(email)
	require.NotNil(suite.T(), sentOTP, "OTP should have been sent")

	// Login with OTP
	loginReq := map[string]any{
		"email":     email,
		"otp":       sentOTP.OTP,
		"client_id": "test-client-123",
	}
	loginResp, err := suite.client.POST(suite.ctx, "/v1/auth/login", loginReq)
	require.NoError(suite.T(), err, "Login should not fail")
	require.Equal(suite.T(), http.StatusOK, loginResp.StatusCode, "Login should return 200")

	var loginResult service.LoginResponse
	err = loginResp.JSON(&loginResult)
	require.NoError(suite.T(), err, "Should be able to parse login response")

	return loginResult.AccessToken
}

// TestCreateCategoryFlow tests admin category creation
func (suite *CategoryTestSuite) TestCreateCategoryFlow() {
	t := suite.T()

	// Seed admin user
	adminUser := fixtures.TestUsers.AdminUser
	err := suite.testDB.SeedUser(adminUser)
	require.NoError(t, err, "Should be able to seed admin user")

	// Get admin token
	adminToken := suite.getAuthToken(adminUser.Email)
	suite.client.SetAuthToken(adminToken)

	// Test category creation
	createReq := map[string]any{"name": "Test Beverages"}
	resp, err := suite.client.POST(suite.ctx, "/v1/admin/categories", createReq)
	require.NoError(t, err, "Category creation should not fail")
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Category creation should return 201")

	var createResp service.CreateCategoryResponse
	err = resp.JSON(&createResp)
	require.NoError(t, err, "Should be able to parse category response")
	assert.NotEmpty(t, createResp.Category.ID, "Category ID should be returned")
	assert.Equal(t, "Test Beverages", createResp.Category.Name, "Category name should match")
	assert.True(t, createResp.Category.IsActive, "Category should be active by default")
	assert.True(t, createResp.Success, "Creation should be successful")
	assert.Contains(t, createResp.Message, "successfully", "Should indicate successful creation")

	// Verify category was created in database
	categoryObjectID, err := primitive.ObjectIDFromHex(createResp.Category.ID)
	require.NoError(t, err, "Should be able to convert category ID")
	category, err := suite.testDB.GetCategoryByID(categoryObjectID)
	require.NoError(t, err, "Should be able to get category from database")
	require.NotNil(t, category, "Category should exist in database")
	assert.Equal(t, createResp.Category.ID, category.ID.Hex(), "Category ID should match")
	assert.Equal(t, "Test Beverages", category.Name, "Category name should match")
	assert.True(t, category.IsActive, "Category should be active")
}

// TestGetCategoryFlow tests retrieving a specific category
func (suite *CategoryTestSuite) TestGetCategoryFlow() {
	t := suite.T()

	// Seed admin user
	adminUser := fixtures.TestUsers.AdminUser
	err := suite.testDB.SeedUser(adminUser)
	require.NoError(t, err, "Should be able to seed admin user")

	// Seed a test category
	testCategory := &domain.Category{
		ID:       primitive.NewObjectID(),
		Name:     "Test Snacks",
		IsActive: true,
	}
	err = suite.testDB.SeedCategory(testCategory)
	require.NoError(t, err, "Should be able to seed category")

	// Get admin token
	adminToken := suite.getAuthToken(adminUser.Email)
	suite.client.SetAuthToken(adminToken)

	// Get category by ID
	categoryID := testCategory.ID.Hex()
	resp, err := suite.client.GET(suite.ctx, "/v1/admin/categories/"+categoryID)
	require.NoError(t, err, "Get category should not fail")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Get category should return 200")

	var categoryResp service.GetCategoryResponse
	err = resp.JSON(&categoryResp)
	require.NoError(t, err, "Should be able to parse category response")
	assert.Equal(t, categoryID, categoryResp.Category.ID, "Category ID should match")
	assert.Equal(t, "Test Snacks", categoryResp.Category.Name, "Category name should match")
	assert.True(t, categoryResp.Category.IsActive, "Category should be active")
}

// TestUpdateCategoryFlow tests category updates
func (suite *CategoryTestSuite) TestUpdateCategoryFlow() {
	t := suite.T()

	// Seed admin user
	adminUser := fixtures.TestUsers.AdminUser
	err := suite.testDB.SeedUser(adminUser)
	require.NoError(t, err, "Should be able to seed admin user")

	// Seed a test category
	testCategory := &domain.Category{
		ID:       primitive.NewObjectID(),
		Name:     "Old Category Name",
		IsActive: true,
	}
	err = suite.testDB.SeedCategory(testCategory)
	require.NoError(t, err, "Should be able to seed category")

	// Get admin token
	adminToken := suite.getAuthToken(adminUser.Email)
	suite.client.SetAuthToken(adminToken)

	// Update category
	updateReq := map[string]any{"name": "Updated Category Name"}
	categoryID := testCategory.ID.Hex()
	resp, err := suite.client.PUT(suite.ctx, "/v1/admin/categories/"+categoryID, updateReq)
	require.NoError(t, err, "Category update should not fail")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Category update should return 200")

	var updateResp service.UpdateCategoryResponse
	err = resp.JSON(&updateResp)
	require.NoError(t, err, "Should be able to parse update response")
	assert.Equal(t, categoryID, updateResp.Category.ID, "Category ID should match")
	assert.Equal(t, "Updated Category Name", updateResp.Category.Name, "Category name should be updated")
	assert.True(t, updateResp.Success, "Update should be successful")
	assert.Contains(t, updateResp.Message, "successfully", "Should indicate successful update")

	// Verify category was updated in database
	category, err := suite.testDB.GetCategoryByID(testCategory.ID)
	require.NoError(t, err, "Should be able to get category from database")
	assert.Equal(t, "Updated Category Name", category.Name, "Category name should be updated in database")
}

// TestDeleteCategoryFlow tests category deletion
func (suite *CategoryTestSuite) TestDeleteCategoryFlow() {
	t := suite.T()

	// Seed admin user
	adminUser := fixtures.TestUsers.AdminUser
	err := suite.testDB.SeedUser(adminUser)
	require.NoError(t, err, "Should be able to seed admin user")

	// Seed a test category
	testCategory := &domain.Category{
		ID:       primitive.NewObjectID(),
		Name:     "Category To Delete",
		IsActive: true,
	}
	err = suite.testDB.SeedCategory(testCategory)
	require.NoError(t, err, "Should be able to seed category")

	// Get admin token
	adminToken := suite.getAuthToken(adminUser.Email)
	suite.client.SetAuthToken(adminToken)

	// Delete category
	categoryID := testCategory.ID.Hex()
	resp, err := suite.client.DELETE(suite.ctx, "/v1/admin/categories/"+categoryID)
	require.NoError(t, err, "Category deletion should not fail")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Category deletion should return 200")

	var deleteResp service.DeleteCategoryResponse
	err = resp.JSON(&deleteResp)
	require.NoError(t, err, "Should be able to parse delete response")
	assert.True(t, deleteResp.Success, "Deletion should be successful")
	assert.Contains(t, deleteResp.Message, "successfully", "Should indicate successful deletion")

	// Verify category was deleted from database
	category, err := suite.testDB.GetCategoryByID(testCategory.ID)
	require.NoError(t, err, "Database query should not fail")
	assert.Nil(t, category, "Category should be deleted from database")
}

// TestListCategoriesFlow tests category listing with filtering and pagination
func (suite *CategoryTestSuite) TestListCategoriesFlow() {
	t := suite.T()

	// Seed admin user
	adminUser := fixtures.TestUsers.AdminUser
	err := suite.testDB.SeedUser(adminUser)
	require.NoError(t, err, "Should be able to seed admin user")

	// Seed multiple test categories
	testCategories := []*domain.Category{
		{
			ID:       primitive.NewObjectID(),
			Name:     "Active Category 1",
			IsActive: true,
		},
		{
			ID:       primitive.NewObjectID(),
			Name:     "Active Category 2",
			IsActive: true,
		},
		{
			ID:       primitive.NewObjectID(),
			Name:     "Inactive Category",
			IsActive: false,
		},
	}
	err = suite.testDB.SeedCategories(testCategories)
	require.NoError(t, err, "Should be able to seed categories")

	// Get admin token
	adminToken := suite.getAuthToken(adminUser.Email)
	suite.client.SetAuthToken(adminToken)

	// Test list all categories
	resp, err := suite.client.GET(suite.ctx, "/v1/admin/categories")
	require.NoError(t, err, "List categories should not fail")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "List categories should return 200")

	var listResp service.ListCategoriesResponse
	err = resp.JSON(&listResp)
	require.NoError(t, err, "Should be able to parse list response")
	assert.GreaterOrEqual(t, len(listResp.Categories), 3, "Should return at least 3 categories")
	assert.Equal(t, len(listResp.Categories), listResp.Total, "Total should match categories count")

	// Test filter by active status
	resp, err = suite.client.GET(suite.ctx, "/v1/admin/categories?active_only=true")
	require.NoError(t, err, "Filter by active should not fail")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Filter by active should return 200")

	err = resp.JSON(&listResp)
	require.NoError(t, err, "Should be able to parse filtered response")
	assert.GreaterOrEqual(t, len(listResp.Categories), 2, "Should return at least 2 active categories")
	for _, category := range listResp.Categories {
		assert.True(t, category.IsActive, "All returned categories should be active")
	}

	// Test pagination
	resp, err = suite.client.GET(suite.ctx, "/v1/admin/categories?limit=1&offset=0")
	require.NoError(t, err, "Pagination should not fail")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Pagination should return 200")

	err = resp.JSON(&listResp)
	require.NoError(t, err, "Should be able to parse paginated response")
	assert.LessOrEqual(t, len(listResp.Categories), 1, "Should return at most 1 category")
}

// TestSetCategoryActiveFlow tests category activation/deactivation
func (suite *CategoryTestSuite) TestSetCategoryActiveFlow() {
	t := suite.T()

	// Seed admin user
	adminUser := fixtures.TestUsers.AdminUser
	err := suite.testDB.SeedUser(adminUser)
	require.NoError(t, err, "Should be able to seed admin user")

	// Seed a test category
	testCategory := &domain.Category{
		ID:       primitive.NewObjectID(),
		Name:     "Category To Toggle",
		IsActive: true,
	}
	err = suite.testDB.SeedCategory(testCategory)
	require.NoError(t, err, "Should be able to seed category")

	// Get admin token
	adminToken := suite.getAuthToken(adminUser.Email)
	suite.client.SetAuthToken(adminToken)

	categoryID := testCategory.ID.Hex()

	// Test deactivating category
	resp, err := suite.client.PUT(suite.ctx, "/v1/admin/categories/"+categoryID+"/status?active=false", nil)
	require.NoError(t, err, "Category deactivation should not fail")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Category deactivation should return 200")

	var statusResp service.SetCategoryActiveResponse
	err = resp.JSON(&statusResp)
	require.NoError(t, err, "Should be able to parse status response")
	assert.True(t, statusResp.Success, "Status update should be successful")
	assert.Contains(t, statusResp.Message, "deactivated", "Should indicate deactivation")

	// Verify category status changed in database
	category, err := suite.testDB.GetCategoryByID(testCategory.ID)
	require.NoError(t, err, "Should be able to get category from database")
	assert.False(t, category.IsActive, "Category should be inactive")

	// Test reactivating category
	resp, err = suite.client.PUT(suite.ctx, "/v1/admin/categories/"+categoryID+"/status?active=true", nil)
	require.NoError(t, err, "Category activation should not fail")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Category activation should return 200")

	err = resp.JSON(&statusResp)
	require.NoError(t, err, "Should be able to parse status response")
	assert.True(t, statusResp.Success, "Status update should be successful")
	assert.Contains(t, statusResp.Message, "activated", "Should indicate activation")

	// Verify category status changed in database
	category, err = suite.testDB.GetCategoryByID(testCategory.ID)
	require.NoError(t, err, "Should be able to get category from database")
	assert.True(t, category.IsActive, "Category should be active")
}

// TestDuplicateCategoryName tests that duplicate category names are rejected
func (suite *CategoryTestSuite) TestDuplicateCategoryName() {
	t := suite.T()

	// Seed admin user
	adminUser := fixtures.TestUsers.AdminUser
	err := suite.testDB.SeedUser(adminUser)
	require.NoError(t, err, "Should be able to seed admin user")

	// Seed existing category
	existingCategory := &domain.Category{
		ID:       primitive.NewObjectID(),
		Name:     "Existing Category",
		IsActive: true,
	}
	err = suite.testDB.SeedCategory(existingCategory)
	require.NoError(t, err, "Should be able to seed existing category")

	// Get admin token
	adminToken := suite.getAuthToken(adminUser.Email)
	suite.client.SetAuthToken(adminToken)

	// Try to create category with duplicate name
	createReq := map[string]any{"name": "Existing Category"}
	resp, err := suite.client.POST(suite.ctx, "/v1/admin/categories", createReq)
	require.NoError(t, err, "Request should not fail")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Should return 400 for duplicate name")

	var errorResp map[string]any
	err = resp.JSON(&errorResp)
	require.NoError(t, err, "Should be able to parse error response")
	assert.Contains(t, errorResp, "error", "Error response should contain error field")

	// Create another category and try to update it to duplicate name
	anotherCategory := &domain.Category{
		ID:       primitive.NewObjectID(),
		Name:     "Another Category",
		IsActive: true,
	}
	err = suite.testDB.SeedCategory(anotherCategory)
	require.NoError(t, err, "Should be able to seed another category")

	// Try to update with duplicate name
	updateReq := map[string]any{"name": "Existing Category"}
	resp, err = suite.client.PUT(suite.ctx, "/v1/admin/categories/"+anotherCategory.ID.Hex(), updateReq)
	require.NoError(t, err, "Request should not fail")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Should return 400 for duplicate name on update")
}

// TestInvalidRequests tests various invalid request scenarios
func (suite *CategoryTestSuite) TestInvalidRequests() {
	t := suite.T()

	// Seed users for authentication tests
	adminUser := fixtures.TestUsers.AdminUser
	customerUser := fixtures.TestUsers.CustomerUser
	err := suite.testDB.SeedUsers([]*domain.User{adminUser, customerUser})
	require.NoError(t, err, "Should be able to seed users")

	// Get tokens
	adminToken := suite.getAuthToken(adminUser.Email)
	userToken := suite.getAuthToken(customerUser.Email)

	testCases := []struct {
		name           string
		endpoint       string
		method         string
		payload        map[string]any
		requireAuth    bool
		requireAdmin   bool
		expectedStatus int
	}{
		{
			name:           "Create category with empty name",
			endpoint:       "/v1/admin/categories",
			method:         "POST",
			payload:        map[string]any{"name": ""},
			requireAuth:    true,
			requireAdmin:   true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Create category with missing name",
			endpoint:       "/v1/admin/categories",
			method:         "POST",
			payload:        map[string]any{},
			requireAuth:    true,
			requireAdmin:   true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Get category with invalid ID",
			endpoint:       "/v1/admin/categories/invalid-id",
			method:         "GET",
			payload:        nil,
			requireAuth:    true,
			requireAdmin:   true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Update category with empty name",
			endpoint:       "/v1/admin/categories/" + primitive.NewObjectID().Hex(),
			method:         "PUT",
			payload:        map[string]any{"name": ""},
			requireAuth:    true,
			requireAdmin:   true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Set category status without active parameter",
			endpoint:       "/v1/admin/categories/" + primitive.NewObjectID().Hex() + "/status",
			method:         "PUT",
			payload:        nil,
			requireAuth:    true,
			requireAdmin:   true,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			if tc.requireAuth {
				if tc.requireAdmin {
					suite.client.SetAuthToken(adminToken)
				} else {
					suite.client.SetAuthToken(userToken)
				}
			} else {
				suite.client.ClearAuth()
			}

			var resp *helpers.HTTPResponse
			var err error

			switch tc.method {
			case "POST":
				resp, err = suite.client.POST(suite.ctx, tc.endpoint, tc.payload)
			case "PUT":
				resp, err = suite.client.PUT(suite.ctx, tc.endpoint, tc.payload)
			case "GET":
				resp, err = suite.client.GET(suite.ctx, tc.endpoint)
			case "DELETE":
				resp, err = suite.client.DELETE(suite.ctx, tc.endpoint)
			}

			require.NoError(t, err, "Request should not fail")
			assert.Equal(t, tc.expectedStatus, resp.StatusCode,
				fmt.Sprintf("Expected status %d for %s", tc.expectedStatus, tc.name))

			// Verify error response structure for error cases
			if resp.StatusCode >= 400 {
				var errorResp map[string]any
				err = resp.JSON(&errorResp)
				require.NoError(t, err, "Should be able to parse error response")
				assert.Contains(t, errorResp, "error", "Error response should contain error field")
			}
		})
	}
}

// TestAuthenticationRequirements tests authentication requirements for category endpoints
func (suite *CategoryTestSuite) TestAuthenticationRequirements() {
	t := suite.T()

	// Seed users for authentication tests
	adminUser := fixtures.TestUsers.AdminUser
	customerUser := fixtures.TestUsers.CustomerUser
	err := suite.testDB.SeedUsers([]*domain.User{adminUser, customerUser})
	require.NoError(t, err, "Should be able to seed users")

	// Get tokens
	adminToken := suite.getAuthToken(adminUser.Email)
	userToken := suite.getAuthToken(customerUser.Email)

	// Seed a test category for operations
	testCategory := &domain.Category{
		ID:       primitive.NewObjectID(),
		Name:     "Test Category",
		IsActive: true,
	}
	err = suite.testDB.SeedCategory(testCategory)
	require.NoError(t, err, "Should be able to seed category")

	categoryID := testCategory.ID.Hex()

	testCases := []struct {
		name           string
		endpoint       string
		method         string
		payload        map[string]any
		requireAuth    bool
		requireAdmin   bool
		expectedStatus int
	}{
		{
			name:           "Create category requires admin",
			endpoint:       "/v1/admin/categories",
			method:         "POST",
			payload:        map[string]any{"name": "Admin Test Category"},
			requireAuth:    true,
			requireAdmin:   true,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "List categories requires admin",
			endpoint:       "/v1/admin/categories",
			method:         "GET",
			payload:        nil,
			requireAuth:    true,
			requireAdmin:   true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get category requires admin",
			endpoint:       "/v1/admin/categories/" + categoryID,
			method:         "GET",
			payload:        nil,
			requireAuth:    true,
			requireAdmin:   true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Update category requires admin",
			endpoint:       "/v1/admin/categories/" + categoryID,
			method:         "PUT",
			payload:        map[string]any{"name": "Updated Admin Category"},
			requireAuth:    true,
			requireAdmin:   true,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Delete category requires admin",
			endpoint:       "/v1/admin/categories/" + categoryID,
			method:         "DELETE",
			payload:        nil,
			requireAuth:    true,
			requireAdmin:   true,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Test without authentication (should fail)
			if tc.requireAuth {
				suite.client.ClearAuth()

				var resp *helpers.HTTPResponse
				var err error
				switch tc.method {
				case "POST":
					resp, err = suite.client.POST(suite.ctx, tc.endpoint, tc.payload)
				case "PUT":
					resp, err = suite.client.PUT(suite.ctx, tc.endpoint, tc.payload)
				case "GET":
					resp, err = suite.client.GET(suite.ctx, tc.endpoint)
				case "DELETE":
					resp, err = suite.client.DELETE(suite.ctx, tc.endpoint)
				}

				require.NoError(t, err, "Request should not fail")
				assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Should require authentication")
			}

			// Test with customer auth (should fail for admin endpoints)
			if tc.requireAdmin {
				suite.client.SetAuthToken(userToken)

				var resp *helpers.HTTPResponse
				var err error
				switch tc.method {
				case "POST":
					resp, err = suite.client.POST(suite.ctx, tc.endpoint, tc.payload)
				case "PUT":
					resp, err = suite.client.PUT(suite.ctx, tc.endpoint, tc.payload)
				case "GET":
					resp, err = suite.client.GET(suite.ctx, tc.endpoint)
				case "DELETE":
					resp, err = suite.client.DELETE(suite.ctx, tc.endpoint)
				}

				require.NoError(t, err, "Request should not fail")
				assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Should require admin role")
			}

			// Test with proper authentication
			if tc.requireAuth {
				if tc.requireAdmin {
					suite.client.SetAuthToken(adminToken)
				} else {
					suite.client.SetAuthToken(userToken)
				}
			} else {
				suite.client.ClearAuth()
			}

			var resp *helpers.HTTPResponse
			var err error
			switch tc.method {
			case "POST":
				resp, err = suite.client.POST(suite.ctx, tc.endpoint, tc.payload)
			case "PUT":
				resp, err = suite.client.PUT(suite.ctx, tc.endpoint, tc.payload)
			case "GET":
				resp, err = suite.client.GET(suite.ctx, tc.endpoint)
			case "DELETE":
				resp, err = suite.client.DELETE(suite.ctx, tc.endpoint)
			}

			require.NoError(t, err, "Request should not fail")
			assert.Equal(t, tc.expectedStatus, resp.StatusCode,
				fmt.Sprintf("Expected status %d for %s with proper auth", tc.expectedStatus, tc.name))
		})
	}
}

// TestNonExistentCategory tests scenarios with non-existent categories
func (suite *CategoryTestSuite) TestNonExistentCategory() {
	t := suite.T()

	// Seed admin user
	adminUser := fixtures.TestUsers.AdminUser
	err := suite.testDB.SeedUser(adminUser)
	require.NoError(t, err, "Should be able to seed admin user")

	// Get admin token
	adminToken := suite.getAuthToken(adminUser.Email)
	suite.client.SetAuthToken(adminToken)

	nonExistentID := primitive.NewObjectID().Hex()

	// Try to get non-existent category
	resp, err := suite.client.GET(suite.ctx, "/v1/admin/categories/"+nonExistentID)
	require.NoError(t, err, "Request should not fail")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Should return 404 for non-existent category")

	// Try to update non-existent category
	updateReq := map[string]any{"name": "Updated Name"}
	resp, err = suite.client.PUT(suite.ctx, "/v1/admin/categories/"+nonExistentID, updateReq)
	require.NoError(t, err, "Request should not fail")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Should return 404 for non-existent category")

	// Try to delete non-existent category
	resp, err = suite.client.DELETE(suite.ctx, "/v1/admin/categories/"+nonExistentID)
	require.NoError(t, err, "Request should not fail")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Should return 404 for non-existent category")

	// Try to set status of non-existent category
	resp, err = suite.client.PUT(suite.ctx, "/v1/admin/categories/"+nonExistentID+"/status?active=false", nil)
	require.NoError(t, err, "Request should not fail")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Should return 404 for non-existent category")
}

// TestCategoryEndToEnd tests complete category lifecycle from creation to deletion
func (suite *CategoryTestSuite) TestCategoryEndToEnd() {
	t := suite.T()

	// Seed admin user
	adminUser := fixtures.TestUsers.AdminUser
	err := suite.testDB.SeedUser(adminUser)
	require.NoError(t, err, "Should be able to seed admin user")

	// Get admin token
	adminToken := suite.getAuthToken(adminUser.Email)
	suite.client.SetAuthToken(adminToken)

	categoryName := "End-to-End Test Category"

	// Step 1: Create category
	createReq := map[string]any{"name": categoryName}
	resp, err := suite.client.POST(suite.ctx, "/v1/admin/categories", createReq)
	require.NoError(t, err, "Category creation should not fail")
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Category creation should return 201")

	var createResp service.CreateCategoryResponse
	err = resp.JSON(&createResp)
	require.NoError(t, err, "Should be able to parse category response")
	categoryID := createResp.Category.ID

	// Step 2: List categories (should include our new category)
	resp, err = suite.client.GET(suite.ctx, "/v1/admin/categories")
	require.NoError(t, err, "List categories should not fail")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "List categories should return 200")

	var listResp service.ListCategoriesResponse
	err = resp.JSON(&listResp)
	require.NoError(t, err, "Should be able to parse list response")

	found := false
	for _, category := range listResp.Categories {
		if category.ID == categoryID {
			found = true
			assert.Equal(t, categoryName, category.Name, "Category name should match")
			assert.True(t, category.IsActive, "Category should be active")
			break
		}
	}
	assert.True(t, found, "Category should be found in list")

	// Step 3: Get category by ID
	resp, err = suite.client.GET(suite.ctx, "/v1/admin/categories/"+categoryID)
	require.NoError(t, err, "Get category should not fail")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Get category should return 200")

	var getResp service.GetCategoryResponse
	err = resp.JSON(&getResp)
	require.NoError(t, err, "Should be able to parse get response")
	assert.Equal(t, categoryName, getResp.Category.Name, "Category name should match")

	// Step 4: Update category name
	newName := "Updated " + categoryName
	updateReq := map[string]any{"name": newName}
	resp, err = suite.client.PUT(suite.ctx, "/v1/admin/categories/"+categoryID, updateReq)
	require.NoError(t, err, "Category update should not fail")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Category update should return 200")

	var updateResp service.UpdateCategoryResponse
	err = resp.JSON(&updateResp)
	require.NoError(t, err, "Should be able to parse update response")
	assert.Equal(t, newName, updateResp.Category.Name, "Category name should be updated")

	// Step 5: Deactivate category
	resp, err = suite.client.PUT(suite.ctx, "/v1/admin/categories/"+categoryID+"/status?active=false", nil)
	require.NoError(t, err, "Category deactivation should not fail")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Category deactivation should return 200")

	// Step 6: Verify category is not in active-only list
	resp, err = suite.client.GET(suite.ctx, "/v1/admin/categories?active_only=true")
	require.NoError(t, err, "List active categories should not fail")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "List active categories should return 200")

	err = resp.JSON(&listResp)
	require.NoError(t, err, "Should be able to parse list response")

	found = false
	for _, category := range listResp.Categories {
		if category.ID == categoryID {
			found = true
			break
		}
	}
	assert.False(t, found, "Inactive category should not be found in active-only list")

	// Step 7: Reactivate category
	resp, err = suite.client.PUT(suite.ctx, "/v1/admin/categories/"+categoryID+"/status?active=true", nil)
	require.NoError(t, err, "Category activation should not fail")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Category activation should return 200")

	// Step 8: Delete category
	resp, err = suite.client.DELETE(suite.ctx, "/v1/admin/categories/"+categoryID)
	require.NoError(t, err, "Category deletion should not fail")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Category deletion should return 200")

	// Step 9: Verify category is deleted
	resp, err = suite.client.GET(suite.ctx, "/v1/admin/categories/"+categoryID)
	require.NoError(t, err, "Request should not fail")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Deleted category should return 404")
}

func TestCategoryTestSuite(t *testing.T) {
	suite.Run(t, new(CategoryTestSuite))
}
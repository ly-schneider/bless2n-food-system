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

type StationTestSuite struct {
	suite.Suite
	server  *helpers.TestServer
	client  *helpers.HTTPClient
	testDB  *helpers.TestDB
	ctx     context.Context
	baseURL string
}

func (suite *StationTestSuite) SetupSuite() {
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

func (suite *StationTestSuite) TearDownSuite() {
	if suite.testDB != nil {
		suite.testDB.Close()
	}
	if suite.server != nil {
		suite.server.Stop()
	}
}

func (suite *StationTestSuite) SetupTest() {
	err := suite.testDB.CleanAll()
	require.NoError(suite.T(), err, "Failed to clean database")

	suite.server.EmailMock.Reset()

	suite.client.ClearAuth()
}

func (suite *StationTestSuite) TearDownTest() {
	suite.testDB.CleanOTPs()
}

// Helper method to get authentication token for a user (following auth test pattern)
func (suite *StationTestSuite) getAuthToken(email string) string {
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

// TestStationRequestFlow tests the public station request functionality
func (suite *StationTestSuite) TestStationRequestFlow() {
	t := suite.T()

	// Test public station request (no authentication required)
	resp, err := suite.client.POST(suite.ctx, "/v1/stations/request", fixtures.ValidStationRequest)
	require.NoError(t, err, "Station request should not fail")
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Station request should return 201")

	var stationResp service.StationResponse
	err = resp.JSON(&stationResp)
	require.NoError(t, err, "Should be able to parse station response")
	assert.NotEmpty(t, stationResp.ID, "Station ID should be returned")
	assert.Equal(t, fixtures.ValidStationRequest["name"], stationResp.Name, "Station name should match")
	assert.Equal(t, string(domain.StationStatusPending), string(stationResp.Status), "Station should be pending")

	// Verify station was created in database
	stationObjectID, err := primitive.ObjectIDFromHex(stationResp.ID)
	require.NoError(t, err, "Should be able to convert station ID")
	station, err := suite.testDB.GetStationByID(stationObjectID)
	require.NoError(t, err, "Should be able to get station from database")
	require.NotNil(t, station, "Station should exist in database")
	assert.Equal(t, stationResp.ID, station.ID.Hex(), "Station ID should match")
	assert.Equal(t, domain.StationStatusPending, station.Status, "Station should be pending")
}

// TestAdminCreateStationFlow tests admin-only station creation
func (suite *StationTestSuite) TestAdminCreateStationFlow() {
	t := suite.T()

	// First, let's test that a customer can't access admin endpoints
	customerUser := fixtures.TestUsers.CustomerUser
	err := suite.testDB.SeedUser(customerUser)
	require.NoError(t, err, "Should be able to seed customer user")

	customerToken := suite.getAuthToken(customerUser.Email)
	suite.client.SetAuthToken(customerToken)

	// Try admin endpoint with customer token - should fail
	resp, err := suite.client.POST(suite.ctx, "/v1/admin/stations", fixtures.ValidStationRequest)
	require.NoError(t, err, "Request should not fail")
	assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Customer should not be able to create stations")

	// Now test with admin user
	adminUser := fixtures.TestUsers.AdminUser
	err = suite.testDB.SeedUser(adminUser)
	require.NoError(t, err, "Should be able to seed admin user")

	// Get admin token through complete login flow
	adminToken := suite.getAuthToken(adminUser.Email)

	// Test admin station creation
	suite.client.SetAuthToken(adminToken)
	resp, err = suite.client.POST(suite.ctx, "/v1/admin/stations", fixtures.ValidStationRequest)
	require.NoError(t, err, "Admin station creation should not fail")

	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Admin station creation should return 201")

	var stationResp service.StationResponse
	err = resp.JSON(&stationResp)
	require.NoError(t, err, "Should be able to parse station response")
	assert.NotEmpty(t, stationResp.ID, "Station ID should be returned")
	assert.Equal(t, fixtures.ValidStationRequest["name"], stationResp.Name, "Station name should match")

	// Verify station was created in database
	stationObjectID, err := primitive.ObjectIDFromHex(stationResp.ID)
	require.NoError(t, err, "Should be able to convert station ID")
	station, err := suite.testDB.GetStationByID(stationObjectID)
	require.NoError(t, err, "Should be able to get station from database")
	require.NotNil(t, station, "Station should exist in database")
}

// TestListStationsFlow tests station listing with filtering
func (suite *StationTestSuite) TestListStationsFlow() {
	t := suite.T()

	// Seed customer user
	user := fixtures.TestUsers.CustomerUser
	err := suite.testDB.SeedUser(user)
	require.NoError(t, err, "Should be able to seed user")

	// Seed test stations
	err = suite.testDB.SeedStations([]*domain.Station{
		fixtures.TestStations.PendingStation,
		fixtures.TestStations.ApprovedStation,
		fixtures.TestStations.RejectedStation,
	})
	require.NoError(t, err, "Should be able to seed stations")

	// Get user token
	userToken := suite.getAuthToken(user.Email)
	suite.client.SetAuthToken(userToken)

	// Test list all stations
	resp, err := suite.client.GET(suite.ctx, "/v1/stations")
	require.NoError(t, err, "List stations should not fail")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "List stations should return 200")

	var listResp service.ListStationsResponse
	err = resp.JSON(&listResp)
	require.NoError(t, err, "Should be able to parse list response")
	assert.GreaterOrEqual(t, len(listResp.Stations), 3, "Should return at least 3 stations")

	// Test filter by status - approved stations
	resp, err = suite.client.GET(suite.ctx, "/v1/stations?status=approved")
	require.NoError(t, err, "Filter by status should not fail")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Filter by status should return 200")

	err = resp.JSON(&listResp)
	require.NoError(t, err, "Should be able to parse filtered response")
	assert.GreaterOrEqual(t, len(listResp.Stations), 1, "Should return at least 1 approved station")
	for _, station := range listResp.Stations {
		assert.Equal(t, string(domain.StationStatusApproved), string(station.Status), "All stations should be approved")
	}

	// Test pagination
	resp, err = suite.client.GET(suite.ctx, "/v1/stations?limit=1&offset=0")
	require.NoError(t, err, "Pagination should not fail")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Pagination should return 200")

	err = resp.JSON(&listResp)
	require.NoError(t, err, "Should be able to parse paginated response")
	assert.LessOrEqual(t, len(listResp.Stations), 1, "Should return at most 1 station")
}

// TestGetStationFlow tests retrieving a specific station
func (suite *StationTestSuite) TestGetStationFlow() {
	t := suite.T()

	// Seed customer user
	user := fixtures.TestUsers.CustomerUser
	err := suite.testDB.SeedUser(user)
	require.NoError(t, err, "Should be able to seed user")

	// Seed a test station
	err = suite.testDB.SeedStation(fixtures.TestStations.ApprovedStation)
	require.NoError(t, err, "Should be able to seed station")

	// Get user token
	userToken := suite.getAuthToken(user.Email)
	suite.client.SetAuthToken(userToken)

	// Get station by ID
	stationID := fixtures.TestStations.ApprovedStation.ID.Hex()
	resp, err := suite.client.GET(suite.ctx, "/v1/stations/"+stationID)
	require.NoError(t, err, "Get station should not fail")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Get station should return 200")

	var stationResp service.StationResponse
	err = resp.JSON(&stationResp)
	require.NoError(t, err, "Should be able to parse station response")
	assert.Equal(t, stationID, stationResp.ID, "Station ID should match")
	assert.Equal(t, fixtures.TestStations.ApprovedStation.Name, stationResp.Name, "Station name should match")
}

// TestStationStatusUpdateFlow tests station approval and rejection by admin
func (suite *StationTestSuite) TestStationStatusUpdateFlow() {
	t := suite.T()

	// Seed admin user
	adminUser := fixtures.TestUsers.AdminUser
	err := suite.testDB.SeedUser(adminUser)
	require.NoError(t, err, "Should be able to seed admin user")

	// Seed a pending station
	err = suite.testDB.SeedStation(fixtures.TestStations.PendingStation)
	require.NoError(t, err, "Should be able to seed pending station")

	// Get admin token
	adminToken := suite.getAuthToken(adminUser.Email)
	suite.client.SetAuthToken(adminToken)

	stationID := fixtures.TestStations.PendingStation.ID.Hex()

	// Test station approval
	statusResp, err := suite.client.PUT(suite.ctx, "/v1/admin/stations/"+stationID+"/status", fixtures.ValidStationStatusApprovalRequest)
	require.NoError(t, err, "Station status update should not fail")
	assert.Equal(t, http.StatusOK, statusResp.StatusCode, "Station status update should return 200")

	var statusResult service.StatusResponse
	err = statusResp.JSON(&statusResult)
	require.NoError(t, err, "Should be able to parse status response")
	assert.True(t, statusResult.Success, "Status update should be successful")

	// Verify station status changed in database
	station, err := suite.testDB.GetStationByID(fixtures.TestStations.PendingStation.ID)
	require.NoError(t, err, "Should be able to get station from database")
	assert.Equal(t, domain.StationStatusApproved, station.Status, "Station should be approved")
	assert.NotNil(t, station.ApprovedBy, "Approved by should be set")
	assert.NotNil(t, station.ApprovedAt, "Approved at should be set")

	// Test station rejection (create another pending station)
	pendingStation2 := &domain.Station{
		ID:        primitive.NewObjectID(),
		Name:      "Another Pending Station",
		Status:    domain.StationStatusPending,
		CreatedAt: fixtures.TestStations.PendingStation.CreatedAt,
		UpdatedAt: fixtures.TestStations.PendingStation.UpdatedAt,
	}
	err = suite.testDB.SeedStation(pendingStation2)
	require.NoError(t, err, "Should be able to seed second pending station")

	rejectionResp, err := suite.client.PUT(suite.ctx, "/v1/admin/stations/"+pendingStation2.ID.Hex()+"/status", fixtures.ValidStationStatusRejectionRequest)
	require.NoError(t, err, "Station rejection should not fail")
	assert.Equal(t, http.StatusOK, rejectionResp.StatusCode, "Station rejection should return 200")

	err = rejectionResp.JSON(&statusResult)
	require.NoError(t, err, "Should be able to parse rejection response")
	assert.True(t, statusResult.Success, "Rejection should be successful")

	// Verify station status changed in database
	station, err = suite.testDB.GetStationByID(pendingStation2.ID)
	require.NoError(t, err, "Should be able to get station from database")
	assert.Equal(t, domain.StationStatusRejected, station.Status, "Station should be rejected")
	assert.NotNil(t, station.RejectedBy, "Rejected by should be set")
	assert.NotNil(t, station.RejectedAt, "Rejected at should be set")
	assert.NotNil(t, station.RejectionReason, "Rejection reason should be set")
}

// TestStationProductAssignmentFlow tests assigning and removing products from stations
func (suite *StationTestSuite) TestStationProductAssignmentFlow() {
	t := suite.T()

	// Seed admin user
	adminUser := fixtures.TestUsers.AdminUser
	err := suite.testDB.SeedUser(adminUser)
	require.NoError(t, err, "Should be able to seed admin user")

	// Seed an approved station
	err = suite.testDB.SeedStation(fixtures.TestStations.ApprovedStation)
	require.NoError(t, err, "Should be able to seed approved station")

	// Get admin token
	adminToken := suite.getAuthToken(adminUser.Email)
	suite.client.SetAuthToken(adminToken)

	stationID := fixtures.TestStations.ApprovedStation.ID.Hex()

	// Create some mock product IDs for testing
	productID1 := primitive.NewObjectID().Hex()
	productID2 := primitive.NewObjectID().Hex()
	productIDs := []string{productID1, productID2}

	// Test assigning non-existent products to station
	assignResp, err := suite.client.POST(suite.ctx, "/v1/admin/stations/"+stationID+"/products", productIDs)
	require.NoError(t, err, "Product assignment request should not fail")
	assert.Equal(t, http.StatusBadRequest, assignResp.StatusCode, "Product assignment should fail with 400 when products don't exist")

	var assignResponse service.AssignProductsResponse
	err = assignResp.JSON(&assignResponse)
	require.NoError(t, err, "Should be able to parse assignment response")
	assert.False(t, assignResponse.Success, "Assignment should not be successful")
	assert.Contains(t, assignResponse.Message, "not found", "Error message should indicate products not found")

	// Test getting station products
	resp, err := suite.client.GET(suite.ctx, "/v1/stations/"+stationID+"/products")
	require.NoError(t, err, "Get station products should not fail")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Get station products should return 200")

	var productsResp service.StationProductsResponse
	err = resp.JSON(&productsResp)
	require.NoError(t, err, "Should be able to parse products response")
	assert.Empty(t, productsResp.Products, "Station should have no products assigned")

	// Test removing non-existent product from station
	removeResp, err := suite.client.DELETE(suite.ctx, "/v1/admin/stations/"+stationID+"/products/"+productID1)
	require.NoError(t, err, "Product removal request should not fail")
	assert.Equal(t, http.StatusOK, removeResp.StatusCode, "Remove product should return 200 even if product wasn't assigned")
}

// TestInvalidRequests tests various invalid request scenarios
func (suite *StationTestSuite) TestInvalidRequests() {
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
			name:           "Station request with invalid name",
			endpoint:       "/v1/stations/request",
			method:         "POST",
			payload:        fixtures.InvalidRequests.InvalidStationName,
			requireAuth:    false,
			requireAdmin:   false,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Station request with missing name",
			endpoint:       "/v1/stations/request",
			method:         "POST",
			payload:        fixtures.InvalidRequests.MissingStationName,
			requireAuth:    false,
			requireAdmin:   false,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Admin station creation with invalid name",
			endpoint:       "/v1/admin/stations",
			method:         "POST",
			payload:        fixtures.InvalidRequests.InvalidStationName,
			requireAuth:    true,
			requireAdmin:   true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Get station with invalid ID",
			endpoint:       "/v1/stations/invalid-id",
			method:         "GET",
			payload:        nil,
			requireAuth:    true,
			requireAdmin:   false,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "List stations with invalid status filter",
			endpoint:       "/v1/stations?status=invalid",
			method:         "GET",
			payload:        nil,
			requireAuth:    true,
			requireAdmin:   false,
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

// TestAuthenticationRequirements tests authentication requirements for different endpoints
func (suite *StationTestSuite) TestAuthenticationRequirements() {
	t := suite.T()

	// Seed users for authentication tests
	adminUser := fixtures.TestUsers.AdminUser
	customerUser := fixtures.TestUsers.CustomerUser
	err := suite.testDB.SeedUsers([]*domain.User{adminUser, customerUser})
	require.NoError(t, err, "Should be able to seed users")

	// Get tokens
	adminToken := suite.getAuthToken(adminUser.Email)
	userToken := suite.getAuthToken(customerUser.Email)

	// Seed a test station
	err = suite.testDB.SeedStation(fixtures.TestStations.PendingStation)
	require.NoError(t, err, "Should be able to seed station")

	stationID := fixtures.TestStations.PendingStation.ID.Hex()

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
			name:           "Public station request (no auth required)",
			endpoint:       "/v1/stations/request",
			method:         "POST",
			payload:        fixtures.ValidStationRequest,
			requireAuth:    false,
			requireAdmin:   false,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "List stations requires auth",
			endpoint:       "/v1/stations",
			method:         "GET",
			payload:        nil,
			requireAuth:    true,
			requireAdmin:   false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get station requires auth",
			endpoint:       "/v1/stations/" + stationID,
			method:         "GET",
			payload:        nil,
			requireAuth:    true,
			requireAdmin:   false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Admin station creation requires admin",
			endpoint:       "/v1/admin/stations",
			method:         "POST",
			payload:        map[string]any{"name": "Admin Created Test Station"},
			requireAuth:    true,
			requireAdmin:   true,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Station status update requires admin",
			endpoint:       "/v1/admin/stations/" + stationID + "/status",
			method:         "PUT",
			payload:        fixtures.ValidStationStatusApprovalRequest,
			requireAuth:    true,
			requireAdmin:   true,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Test without authentication first (if auth is required)
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
				}

				require.NoError(t, err, "Request should not fail")
				assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Should require authentication")
			}

			// Test with customer auth (if admin is required)
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
			}

			require.NoError(t, err, "Request should not fail")
			assert.Equal(t, tc.expectedStatus, resp.StatusCode,
				fmt.Sprintf("Expected status %d for %s with proper auth", tc.expectedStatus, tc.name))
		})
	}
}

// TestNonExistentStation tests scenarios with non-existent stations
func (suite *StationTestSuite) TestNonExistentStation() {
	t := suite.T()

	// Seed customer user
	user := fixtures.TestUsers.CustomerUser
	err := suite.testDB.SeedUser(user)
	require.NoError(t, err, "Should be able to seed user")

	// Get user token
	userToken := suite.getAuthToken(user.Email)
	suite.client.SetAuthToken(userToken)

	nonExistentID := primitive.NewObjectID().Hex()

	// Try to get non-existent station
	resp, err := suite.client.GET(suite.ctx, "/v1/stations/"+nonExistentID)
	require.NoError(t, err, "Request should not fail")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Should return 404 for non-existent station")
}

// TestStationEndToEnd tests complete station lifecycle from request to approval
func (suite *StationTestSuite) TestStationEndToEnd() {
	t := suite.T()

	// Seed users for the test
	adminUser := fixtures.TestUsers.AdminUser
	customerUser := fixtures.TestUsers.CustomerUser
	err := suite.testDB.SeedUsers([]*domain.User{adminUser, customerUser})
	require.NoError(t, err, "Should be able to seed users")

	// Get tokens
	adminToken := suite.getAuthToken(adminUser.Email)
	userToken := suite.getAuthToken(customerUser.Email)

	stationName := "End-to-End Test Station"

	// Step 1: Public station request
	stationReq := map[string]any{"name": stationName}
	resp, err := suite.client.POST(suite.ctx, "/v1/stations/request", stationReq)
	require.NoError(t, err, "Station request should not fail")
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Station request should return 201")

	var stationResp service.StationResponse
	err = resp.JSON(&stationResp)
	require.NoError(t, err, "Should be able to parse station response")
	stationID := stationResp.ID

	// Step 2: List stations as user (should see the pending station)
	suite.client.SetAuthToken(userToken)
	resp, err = suite.client.GET(suite.ctx, "/v1/stations?status=pending")
	require.NoError(t, err, "List pending stations should not fail")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "List stations should return 200")

	var listResp service.ListStationsResponse
	err = resp.JSON(&listResp)
	require.NoError(t, err, "Should be able to parse list response")

	found := false
	for _, station := range listResp.Stations {
		if station.ID == stationID {
			found = true
			assert.Equal(t, stationName, station.Name, "Station name should match")
			assert.Equal(t, string(domain.StationStatusPending), string(station.Status), "Station should be pending")
			break
		}
	}
	assert.True(t, found, "Station should be found in pending list")

	// Step 3: Admin approves the station
	suite.client.SetAuthToken(adminToken)
	statusReq := map[string]any{"approve": true}
	resp, err = suite.client.PUT(suite.ctx, "/v1/admin/stations/"+stationID+"/status", statusReq)
	require.NoError(t, err, "Station status update should not fail")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Station status update should return 200")

	// Step 4: Verify station is now approved
	suite.client.SetAuthToken(userToken)
	resp, err = suite.client.GET(suite.ctx, "/v1/stations/"+stationID)
	require.NoError(t, err, "Get station should not fail")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Get station should return 200")

	err = resp.JSON(&stationResp)
	require.NoError(t, err, "Should be able to parse station response")
	assert.Equal(t, string(domain.StationStatusApproved), string(stationResp.Status), "Station should be approved")

	// Step 5: Check that station appears in approved list
	resp, err = suite.client.GET(suite.ctx, "/v1/stations?status=approved")
	require.NoError(t, err, "List approved stations should not fail")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "List approved stations should return 200")

	err = resp.JSON(&listResp)
	require.NoError(t, err, "Should be able to parse list response")

	found = false
	for _, station := range listResp.Stations {
		if station.ID == stationID {
			found = true
			assert.Equal(t, string(domain.StationStatusApproved), string(station.Status), "Station should be approved")
			break
		}
	}
	assert.True(t, found, "Station should be found in approved list")
}

func TestStationTestSuite(t *testing.T) {
	suite.Run(t, new(StationTestSuite))
}

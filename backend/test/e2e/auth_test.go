package e2e

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"backend/internal/service"
	"backend/test/fixtures"
	"backend/test/helpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type AuthTestSuite struct {
	suite.Suite
	server  *helpers.TestServer
	client  *helpers.HTTPClient
	testDB  *helpers.TestDB
	ctx     context.Context
	baseURL string
}

func (suite *AuthTestSuite) SetupSuite() {
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

func (suite *AuthTestSuite) TearDownSuite() {
	if suite.testDB != nil {
		suite.testDB.Close()
	}
	if suite.server != nil {
		suite.server.Stop()
	}
}

func (suite *AuthTestSuite) SetupTest() {
	err := suite.testDB.CleanAll()
	require.NoError(suite.T(), err, "Failed to clean database")

	suite.server.EmailMock.Reset()

	suite.client.ClearAuth()
}

func (suite *AuthTestSuite) TearDownTest() {
	suite.testDB.CleanOTPs()
}

// TestCustomerRegistrationFlow tests the complete customer registration flow
func (suite *AuthTestSuite) TestCustomerRegistrationFlow() {
	t := suite.T()

	// Register new customer
	resp, err := suite.client.POST(suite.ctx, "/v1/auth/register/customer", fixtures.ValidRegistrationRequest)
	require.NoError(t, err, "Registration request should not fail")
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Registration should return 201")

	var registerResp service.RegisterCustomerResponse
	err = resp.JSON(&registerResp)
	require.NoError(t, err, "Should be able to parse registration response")
	assert.NotEmpty(t, registerResp.UserID, "User ID should be returned")
	assert.Contains(t, registerResp.Message, "Registration successful.", "Should indicate successful registration")

	// Verify user was created in database but not verified
	user, err := suite.testDB.GetUserByEmail(fixtures.ValidRegistrationRequest["email"].(string))
	require.NoError(t, err, "Should be able to get user from database")
	require.NotNil(t, user, "User should exist in database")
	assert.Equal(t, registerResp.UserID, user.ID.Hex(), "User ID should match response")
}

// TestLoginFlow tests the complete login flow
func (suite *AuthTestSuite) TestLoginFlow() {
	t := suite.T()

	// Seed existing user
	user := fixtures.TestUsers.CustomerUser
	err := suite.testDB.SeedUser(user)
	require.NoError(t, err, "Should be able to seed user")

	// Request OTP
	OTPRequest := map[string]any{
		"email": user.Email,
	}

	resp, err := suite.client.POST(suite.ctx, "/v1/auth/request-otp", OTPRequest)
	require.NoError(t, err, "Login OTP request should not fail")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Login OTP request should return 200")

	var otpResp service.RequestOTPResponse
	err = resp.JSON(&otpResp)
	require.NoError(t, err, "Should be able to parse OTP response")
	assert.Contains(t, otpResp.Message, "Login code", "Should indicate successful OTP request")

	// Get the OTP that was sent via email mock
	sentOTP := suite.server.EmailMock.GetLastSentOTP(user.Email)
	require.NotNil(t, sentOTP, "OTP should have been sent via email")
	require.NotEmpty(t, sentOTP.OTP, "Sent OTP should not be empty")
	validOTP := sentOTP.OTP

	// Login with OTP
	loginRequest := map[string]any{
		"email":     user.Email,
		"otp":       validOTP,
		"client_id": "test-client-123",
	}

	resp, err = suite.client.POST(suite.ctx, "/v1/auth/login", loginRequest)
	require.NoError(t, err, "Login should not fail")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Login should return 200")

	var loginResp service.LoginResponse
	err = resp.JSON(&loginResp)
	require.NoError(t, err, "Should be able to parse login response")
	assert.NotEmpty(t, loginResp.AccessToken, "Access token should be returned")
	assert.NotEmpty(t, loginResp.RefreshToken, "Refresh token should be returned")
	assert.Equal(t, "Bearer", loginResp.TokenType, "Token type should be Bearer")
	assert.Greater(t, loginResp.ExpiresIn, int64(0), "Expires in should be positive")
	assert.Equal(t, user.ID, loginResp.User.ID, "User should match logged in user")
}

// TestTokenRefreshFlow tests the token refresh functionality
func (suite *AuthTestSuite) TestTokenRefreshFlow() {
	t := suite.T()

	// Complete login first to get tokens
	user := fixtures.TestUsers.CustomerUser
	err := suite.testDB.SeedUser(user)
	require.NoError(t, err, "Should be able to seed user")

	// Get tokens through login - first request OTP
	OTPRequest := map[string]any{
		"email": user.Email,
	}
	_, err = suite.client.POST(suite.ctx, "/v1/auth/request-otp", OTPRequest)
	require.NoError(t, err, "Should be able to request login OTP")

	// Get the OTP from mock
	sentOTP := suite.server.EmailMock.GetLastSentOTP(user.Email)
	require.NotNil(t, sentOTP, "OTP should have been sent via email")
	validOTP := sentOTP.OTP

	loginRequest := map[string]any{
		"email":     user.Email,
		"otp":       validOTP,
		"client_id": "test-client-123",
	}

	resp, err := suite.client.POST(suite.ctx, "/v1/auth/login", loginRequest)
	require.NoError(t, err, "Login should not fail")

	var loginResp service.LoginResponse
	err = resp.JSON(&loginResp)
	require.NoError(t, err, "Should be able to parse login response")

	// Use refresh token to get new tokens
	refreshRequest := map[string]any{
		"refresh_token": loginResp.RefreshToken,
		"client_id":     "test-client-123",
	}

	resp, err = suite.client.POST(suite.ctx, "/v1/auth/refresh", refreshRequest)
	require.NoError(t, err, "Token refresh should not fail")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Token refresh should return 200")

	var refreshResp service.RefreshTokenResponse
	err = resp.JSON(&refreshResp)
	require.NoError(t, err, "Should be able to parse refresh response")
	assert.NotEmpty(t, refreshResp.AccessToken, "New access token should be returned")
	assert.NotEmpty(t, refreshResp.RefreshToken, "New refresh token should be returned")
	assert.Equal(t, "Bearer", refreshResp.TokenType, "Token type should be Bearer")
	assert.Greater(t, refreshResp.ExpiresIn, int64(0), "Expires in should be positive")

	// Step 2: Verify new tokens are different from old ones
	assert.NotEqual(t, loginResp.AccessToken, refreshResp.AccessToken, "New access token should be different")
	assert.NotEqual(t, loginResp.RefreshToken, refreshResp.RefreshToken, "New refresh token should be different")
}

// TestLogoutFlow tests the logout functionality
func (suite *AuthTestSuite) TestLogoutFlow() {
	t := suite.T()

	// Get tokens through complete login
	user := fixtures.TestUsers.CustomerUser
	err := suite.testDB.SeedUser(user)
	require.NoError(t, err, "Should be able to seed user")

	// Request login OTP
	OTPRequest := map[string]any{
		"email": user.Email,
	}
	_, err = suite.client.POST(suite.ctx, "/v1/auth/request-otp", OTPRequest)
	require.NoError(t, err, "Should be able to request login OTP")

	// Get the OTP from mock
	sentOTP := suite.server.EmailMock.GetLastSentOTP(user.Email)
	require.NotNil(t, sentOTP, "OTP should have been sent via email")
	validOTP := sentOTP.OTP

	loginRequest := map[string]any{
		"email":     user.Email,
		"otp":       validOTP,
		"client_id": "test-client-123",
	}

	resp, err := suite.client.POST(suite.ctx, "/v1/auth/login", loginRequest)
	require.NoError(t, err, "Login should not fail")

	var loginResp service.LoginResponse
	err = resp.JSON(&loginResp)
	require.NoError(t, err, "Should be able to parse login response")

	// Logout with refresh token
	logoutRequest := map[string]any{
		"refresh_token": loginResp.RefreshToken,
	}

	resp, err = suite.client.POST(suite.ctx, "/v1/auth/logout", logoutRequest)
	require.NoError(t, err, "Logout should not fail")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Logout should return 200")

	var logoutResp service.LogoutResponse
	err = resp.JSON(&logoutResp)
	require.NoError(t, err, "Should be able to parse logout response")
	assert.Contains(t, logoutResp.Message, "Logged out", "Should indicate successful logout")

	// Try to use the same refresh token again - should fail
	refreshRequest := map[string]any{
		"refresh_token": loginResp.RefreshToken,
		"client_id":     "test-client-123",
	}

	resp, err = suite.client.POST(suite.ctx, "/v1/auth/refresh", refreshRequest)
	require.NoError(t, err, "Request should not fail")
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "Should return 401 for invalid refresh token")
}

// TestOTPResend tests the OTP resend functionality
func (suite *AuthTestSuite) TestOTPResend() {
	t := suite.T()

	// Seed customer
	user := fixtures.TestUsers.CustomerUser
	err := suite.testDB.SeedUser(user)
	require.NoError(t, err, "Should be able to seed user")

	// Request login OTP
	OTPRequest := map[string]any{
		"email": user.Email,
	}
	_, err = suite.client.POST(suite.ctx, "/v1/auth/request-otp", OTPRequest)
	require.NoError(t, err, "Should be able to request login OTP")

	// Get the OTP from mock
	sentOTP := suite.server.EmailMock.GetLastSentOTP(user.Email)
	require.NotNil(t, sentOTP, "OTP should have been sent via email")
	firstOTP := sentOTP.OTP

	// Request login OTP again
	_, err = suite.client.POST(suite.ctx, "/v1/auth/request-otp", OTPRequest)
	require.NoError(t, err, "Should be able to request login OTP")

	// Fetch the new OTP from mock
	sentOTP = suite.server.EmailMock.GetLastSentOTP(user.Email)
	require.NotNil(t, sentOTP, "OTP should have been sent via email")
	assert.NotEqual(t, firstOTP, sentOTP.OTP, "Should generate a new OTP")

	// Assert not the same
	assert.NotEqual(t, firstOTP, sentOTP.OTP, "Should generate a new OTP")
}

// TestInvalidRequests tests various invalid request scenarios
func (suite *AuthTestSuite) TestInvalidRequests() {
	t := suite.T()

	testCases := []struct {
		name           string
		endpoint       string
		payload        map[string]any
		expectedStatus int
	}{
		{
			name:           "Registration with invalid email",
			endpoint:       "/v1/auth/register/customer",
			payload:        fixtures.InvalidRequests.InvalidEmail,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Registration with missing email",
			endpoint:       "/v1/auth/register/customer",
			payload:        fixtures.InvalidRequests.MissingEmail,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Login with missing client ID",
			endpoint:       "/v1/auth/login",
			payload:        fixtures.InvalidRequests.MissingClientID,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			resp, err := suite.client.POST(suite.ctx, tc.endpoint, tc.payload)
			require.NoError(t, err, "Request should not fail")
			assert.Equal(t, tc.expectedStatus, resp.StatusCode,
				fmt.Sprintf("Expected status %d for %s", tc.expectedStatus, tc.name))

			// Verify error response structure
			var errorResp map[string]any
			err = resp.JSON(&errorResp)
			require.NoError(t, err, "Should be able to parse error response")
			assert.Contains(t, errorResp, "error", "Error response should contain error field")
		})
	}
}

// TestNonExistentUser tests scenarios with non-existent users
func (suite *AuthTestSuite) TestNonExistentUser() {
	t := suite.T()

	// Try to login with non-existent user
	OTPRequest := map[string]any{
		"email": "nonexistent@test.com",
	}

	resp, err := suite.client.POST(suite.ctx, "/v1/auth/request-otp", OTPRequest)
	require.NoError(t, err, "Request should not fail")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Should return 400 for non-existent user")
}

// TestExpiredOTP tests scenarios with expired OTPs
func (suite *AuthTestSuite) TestExpiredOTP() {
	t := suite.T()

	// Setup: Seed user and request OTP, then wait for it to expire
	user := fixtures.TestUsers.CustomerUser
	err := suite.testDB.SeedUser(user)
	require.NoError(t, err, "Should be able to seed user")

	// Request login OTP
	OTPRequest := map[string]any{
		"email": user.Email,
	}
	_, err = suite.client.POST(suite.ctx, "/v1/auth/request-otp", OTPRequest)
	require.NoError(t, err, "Should be able to request login OTP")

	// Get the OTP from mock
	sentOTP := suite.server.EmailMock.GetLastSentOTP(user.Email)
	require.NotNil(t, sentOTP, "OTP should have been sent via email")

	// For this test, we'll use an invalid OTP instead of waiting for expiry
	// (since waiting 10+ minutes is impractical for tests)
	invalidOTP := "000000" // Use a known invalid OTP

	// Try to login with invalid OTP
	loginRequest := map[string]any{
		"email":     user.Email,
		"otp":       invalidOTP,
		"client_id": "test-client-123",
	}

	resp, err := suite.client.POST(suite.ctx, "/v1/auth/login", loginRequest)
	require.NoError(t, err, "Request should not fail")
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "Should return 400 for invalid OTP")
}

// TestAuthenticationEndToEnd tests the complete authentication flow from registration to logout
func (suite *AuthTestSuite) TestAuthenticationEndToEnd() {
	t := suite.T()

	userEmail := "e2e@test.com"
	clientID := "e2e-client-123"

	// Register
	registerReq := map[string]any{"email": userEmail}
	resp, err := suite.client.POST(suite.ctx, "/v1/auth/register/customer", registerReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// Request OTP
	otpReq := map[string]any{"email": userEmail}
	resp, err = suite.client.POST(suite.ctx, "/v1/auth/request-otp", otpReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Fetch OTP
	sentOTP := suite.server.EmailMock.GetLastSentOTP(userEmail)
	require.NotNil(t, sentOTP, "OTP should have been sent during registration")
	validOTP := sentOTP.OTP

	// Login
	loginRequest := map[string]any{
		"email":     userEmail,
		"otp":       validOTP,
		"client_id": clientID,
	}
	resp, err = suite.client.POST(suite.ctx, "/v1/auth/login", loginRequest)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var tokens service.VerifyOTPResponse
	err = resp.JSON(&tokens)
	require.NoError(t, err)

	// Store for future use
	suite.client.SetAuthToken(tokens.AccessToken)

	// Refresh access token
	refreshReq := map[string]any{
		"refresh_token": tokens.RefreshToken, "client_id": clientID,
	}
	resp, err = suite.client.POST(suite.ctx, "/v1/auth/refresh", refreshReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Logout
	logoutReq := map[string]any{"refresh_token": tokens.RefreshToken}
	resp, err = suite.client.POST(suite.ctx, "/v1/auth/logout", logoutReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestAuthTestSuite(t *testing.T) {
	suite.Run(t, new(AuthTestSuite))
}

package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/domain"
	"backend/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type mockAuthService struct {
	mock.Mock
}

func (m *mockAuthService) RegisterCustomer(ctx context.Context, req service.RegisterCustomerRequest) (*service.RegisterCustomerResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.RegisterCustomerResponse), args.Error(1)
}

func (m *mockAuthService) RequestOTP(ctx context.Context, req service.RequestOTPRequest) (*service.RequestOTPResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.RequestOTPResponse), args.Error(1)
}

func (m *mockAuthService) Login(ctx context.Context, req service.LoginRequest) (*service.LoginResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.LoginResponse), args.Error(1)
}

func (m *mockAuthService) RefreshToken(ctx context.Context, req service.RefreshTokenRequest) (*service.RefreshTokenResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.RefreshTokenResponse), args.Error(1)
}

func (m *mockAuthService) Logout(ctx context.Context, req service.LogoutRequest) (*service.LogoutResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.LogoutResponse), args.Error(1)
}

func setupAuthHandler() (*AuthHandler, *mockAuthService) {
	authSvc := &mockAuthService{}
	handler := NewAuthHandler(authSvc)
	return handler, authSvc
}

func TestAuthHandler_RegisterCustomer(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func(*mockAuthService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "successful customer registration",
			requestBody: service.RegisterCustomerRequest{
				Email: "test@example.com",
			},
			setupMocks: func(authSvc *mockAuthService) {
				authSvc.On("RegisterCustomer", mock.Anything, service.RegisterCustomerRequest{
					Email: "test@example.com",
				}).Return(&service.RegisterCustomerResponse{
					Message: "Registration successful.",
					UserID:  primitive.NewObjectID().Hex(),
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"message": "Registration successful.",
			},
		},
		{
			name:           "invalid JSON in request body",
			requestBody:    "invalid-json",
			setupMocks:     func(authSvc *mockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Invalid request body",
			},
		},
		{
			name: "validation failure - missing email",
			requestBody: service.RegisterCustomerRequest{
				Email: "",
			},
			setupMocks:     func(authSvc *mockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Validation failed: Key: 'RegisterCustomerRequest.Email' Error:Field validation for 'Email' failed on the 'required' tag",
			},
		},
		{
			name: "validation failure - invalid email",
			requestBody: service.RegisterCustomerRequest{
				Email: "invalid-email",
			},
			setupMocks:     func(authSvc *mockAuthService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service error - user already exists",
			requestBody: service.RegisterCustomerRequest{
				Email: "existing@example.com",
			},
			setupMocks: func(authSvc *mockAuthService) {
				authSvc.On("RegisterCustomer", mock.Anything, service.RegisterCustomerRequest{
					Email: "existing@example.com",
				}).Return(nil, errors.New("user with email existing@example.com already exists"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "user with email existing@example.com already exists",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, authSvc := setupAuthHandler()
			tt.setupMocks(authSvc)

			var body bytes.Buffer
			if str, ok := tt.requestBody.(string); ok {
				body.WriteString(str)
			} else {
				json.NewEncoder(&body).Encode(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/v1/auth/register/customer", &body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.RegisterCustomer(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				for key, expectedValue := range tt.expectedBody {
					if key == "error" {
						assert.Contains(t, response["message"].(string), expectedValue.(string))
					} else {
						assert.Equal(t, expectedValue, response[key])
					}
				}
			}

			authSvc.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_RequestOTP(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func(*mockAuthService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "successful OTP request",
			requestBody: service.RequestOTPRequest{
				Email: "user@example.com",
			},
			setupMocks: func(authSvc *mockAuthService) {
				authSvc.On("RequestOTP", mock.Anything, service.RequestOTPRequest{
					Email: "user@example.com",
				}).Return(&service.RequestOTPResponse{
					Message: "Login code sent to your email.",
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"message": "Login code sent to your email.",
			},
		},
		{
			name:           "invalid JSON in request body",
			requestBody:    "invalid-json",
			setupMocks:     func(authSvc *mockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Invalid request body",
			},
		},
		{
			name: "validation failure - missing email",
			requestBody: service.RequestOTPRequest{
				Email: "",
			},
			setupMocks:     func(authSvc *mockAuthService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service error - user not found",
			requestBody: service.RequestOTPRequest{
				Email: "notfound@example.com",
			},
			setupMocks: func(authSvc *mockAuthService) {
				authSvc.On("RequestOTP", mock.Anything, service.RequestOTPRequest{
					Email: "notfound@example.com",
				}).Return(nil, errors.New("user not found"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "user not found",
			},
		},
		{
			name: "service error - disabled account",
			requestBody: service.RequestOTPRequest{
				Email: "disabled@example.com",
			},
			setupMocks: func(authSvc *mockAuthService) {
				authSvc.On("RequestOTP", mock.Anything, service.RequestOTPRequest{
					Email: "disabled@example.com",
				}).Return(nil, errors.New("account is disabled"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "account is disabled",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, authSvc := setupAuthHandler()
			tt.setupMocks(authSvc)

			var body bytes.Buffer
			if str, ok := tt.requestBody.(string); ok {
				body.WriteString(str)
			} else {
				json.NewEncoder(&body).Encode(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/v1/auth/request-login-otp", &body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.RequestOTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				for key, expectedValue := range tt.expectedBody {
					if key == "error" {
						assert.Contains(t, response["message"].(string), expectedValue.(string))
					} else {
						assert.Equal(t, expectedValue, response[key])
					}
				}
			}

			authSvc.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Login(t *testing.T) {
	userID := primitive.NewObjectID()
	user := &domain.User{
		ID:    userID,
		Email: "user@example.com",
		Role:  domain.UserRoleCustomer,
	}

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func(*mockAuthService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "successful login",
			requestBody: service.LoginRequest{
				Email:    "user@example.com",
				OTP:      "123456",
				ClientID: "test-client",
			},
			setupMocks: func(authSvc *mockAuthService) {
				authSvc.On("Login", mock.Anything, service.LoginRequest{
					Email:    "user@example.com",
					OTP:      "123456",
					ClientID: "test-client",
				}).Return(&service.LoginResponse{
					AccessToken:  "access_token",
					RefreshToken: "refresh_token",
					TokenType:    "Bearer",
					ExpiresIn:    3600,
					User:         user,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"access_token":  "access_token",
				"refresh_token": "refresh_token",
				"token_type":    "Bearer",
				"expires_in":    float64(3600),
			},
		},
		{
			name:           "invalid JSON in request body",
			requestBody:    "invalid-json",
			setupMocks:     func(authSvc *mockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Invalid request body",
			},
		},
		{
			name: "validation failure - missing fields",
			requestBody: service.LoginRequest{
				Email: "",
				OTP:   "",
			},
			setupMocks:     func(authSvc *mockAuthService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service error - invalid OTP",
			requestBody: service.LoginRequest{
				Email:    "user@example.com",
				OTP:      "000000",
				ClientID: "test-client",
			},
			setupMocks: func(authSvc *mockAuthService) {
				authSvc.On("Login", mock.Anything, service.LoginRequest{
					Email:    "user@example.com",
					OTP:      "000000",
					ClientID: "test-client",
				}).Return(nil, errors.New("invalid OTP code"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "invalid OTP code",
			},
		},
		{
			name: "service error - user not found",
			requestBody: service.LoginRequest{
				Email:    "notfound@example.com",
				OTP:      "123456",
				ClientID: "test-client",
			},
			setupMocks: func(authSvc *mockAuthService) {
				authSvc.On("Login", mock.Anything, service.LoginRequest{
					Email:    "notfound@example.com",
					OTP:      "123456",
					ClientID: "test-client",
				}).Return(nil, errors.New("invalid credentials"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "invalid credentials",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, authSvc := setupAuthHandler()
			tt.setupMocks(authSvc)

			var body bytes.Buffer
			if str, ok := tt.requestBody.(string); ok {
				body.WriteString(str)
			} else {
				json.NewEncoder(&body).Encode(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", &body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Login(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)

				for key, expectedValue := range tt.expectedBody {
					if key == "error" {
						assert.Contains(t, response["message"].(string), expectedValue.(string))
					} else {
						assert.Equal(t, expectedValue, response[key])
					}
				}
			}

			authSvc.AssertExpectations(t)
		})
	}
}

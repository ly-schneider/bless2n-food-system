package auth

import (
	"context"
	"errors"
	"testing"

	"backend/internal/domain"
	"backend/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	if args.Get(0) == nil {
		user.ID = primitive.NewObjectID()
	}
	return args.Error(0)
}

func (m *mockUserRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepository) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUserRepository) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*domain.User), args.Error(1)
}

func (m *mockUserRepository) ListCustomers(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*domain.User), args.Error(1)
}

func (m *mockUserRepository) CountCustomers(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

func (m *mockUserRepository) Disable(ctx context.Context, id primitive.ObjectID, reason string) error {
	args := m.Called(ctx, id, reason)
	return args.Error(0)
}

func (m *mockUserRepository) Enable(ctx context.Context, id primitive.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type mockOTPService struct {
	mock.Mock
}

func (m *mockOTPService) GenerateAndSend(ctx context.Context, userID primitive.ObjectID, email string, tokenType domain.TokenType) error {
	args := m.Called(ctx, userID, email, tokenType)
	return args.Error(0)
}

func (m *mockOTPService) Verify(ctx context.Context, userID primitive.ObjectID, otp string, tokenType domain.TokenType) error {
	args := m.Called(ctx, userID, otp, tokenType)
	return args.Error(0)
}

type mockTokenService struct {
	mock.Mock
}

func (m *mockTokenService) GenerateTokenPair(ctx context.Context, user *domain.User, clientID string) (*service.TokenPairResponse, error) {
	args := m.Called(ctx, user, clientID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.TokenPairResponse), args.Error(1)
}

func (m *mockTokenService) RefreshTokenPair(ctx context.Context, refreshToken, clientID string) (*service.TokenPairResponse, error) {
	args := m.Called(ctx, refreshToken, clientID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.TokenPairResponse), args.Error(1)
}

func (m *mockTokenService) RevokeTokenFamily(ctx context.Context, refreshToken, reason string) error {
	args := m.Called(ctx, refreshToken, reason)
	return args.Error(0)
}

func setupAuthService() (*authService, *mockUserRepository, *mockOTPService, *mockTokenService) {
	userRepo := &mockUserRepository{}
	otpService := &mockOTPService{}
	tokenService := &mockTokenService{}

	authSvc := &authService{
		userRepo:     userRepo,
		otpService:   otpService,
		tokenService: tokenService,
	}

	return authSvc, userRepo, otpService, tokenService
}

func TestAuthService_RegisterCustomer(t *testing.T) {
	tests := []struct {
		name           string
		request        service.RegisterCustomerRequest
		setupMocks     func(*mockUserRepository, *mockOTPService, *mockTokenService)
		expectedError  string
		expectedUserID string
	}{
		{
			name: "successful registration",
			request: service.RegisterCustomerRequest{
				Email: "test@example.com",
			},
			setupMocks: func(userRepo *mockUserRepository, otpService *mockOTPService, tokenService *mockTokenService) {
				userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, nil)
				userRepo.On("Create", mock.Anything, mock.MatchedBy(func(user *domain.User) bool {
					return user.Email == "test@example.com" &&
						user.Role == domain.UserRoleCustomer &&
						!user.IsVerified &&
						!user.IsDisabled
				})).Return(nil)
			},
			expectedError:  "",
			expectedUserID: "not-empty",
		},
		{
			name: "user already exists",
			request: service.RegisterCustomerRequest{
				Email: "existing@example.com",
			},
			setupMocks: func(userRepo *mockUserRepository, otpService *mockOTPService, tokenService *mockTokenService) {
				existingUser := &domain.User{
					ID:    primitive.NewObjectID(),
					Email: "existing@example.com",
				}
				userRepo.On("GetByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil)
			},
			expectedError: "user with email existing@example.com already exists",
		},
		{
			name: "database error when checking existing user",
			request: service.RegisterCustomerRequest{
				Email: "test@example.com",
			},
			setupMocks: func(userRepo *mockUserRepository, otpService *mockOTPService, tokenService *mockTokenService) {
				userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, errors.New("db error"))
			},
			expectedError: "failed to check existing user",
		},
		{
			name: "database error when creating user",
			request: service.RegisterCustomerRequest{
				Email: "test@example.com",
			},
			setupMocks: func(userRepo *mockUserRepository, otpService *mockOTPService, tokenService *mockTokenService) {
				userRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, nil)
				userRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedError: "failed to create user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authSvc, userRepo, otpService, tokenService := setupAuthService()
			tt.setupMocks(userRepo, otpService, tokenService)

			ctx := context.Background()
			response, err := authSvc.RegisterCustomer(ctx, tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, "Registration successful.", response.Message)
				if tt.expectedUserID == "not-empty" {
					assert.NotEmpty(t, response.UserID)
				}
			}

			userRepo.AssertExpectations(t)
			otpService.AssertExpectations(t)
			tokenService.AssertExpectations(t)
		})
	}
}

func TestAuthService_RequestOTP(t *testing.T) {
	userID := primitive.NewObjectID()
	disabledReason := "account suspended"

	tests := []struct {
		name          string
		request       service.RequestOTPRequest
		setupMocks    func(*mockUserRepository, *mockOTPService, *mockTokenService)
		expectedError string
	}{
		{
			name: "successful OTP request for verified user",
			request: service.RequestOTPRequest{
				Email: "user@example.com",
			},
			setupMocks: func(userRepo *mockUserRepository, otpService *mockOTPService, tokenService *mockTokenService) {
				user := &domain.User{
					ID:         userID,
					Email:      "user@example.com",
					IsVerified: true,
					IsDisabled: false,
				}
				userRepo.On("GetByEmail", mock.Anything, "user@example.com").Return(user, nil)
				otpService.On("GenerateAndSend", mock.Anything, userID, "user@example.com", domain.TokenTypeLogin).Return(nil)
			},
		},
		{
			name: "successful OTP request for unverified user",
			request: service.RequestOTPRequest{
				Email: "newuser@example.com",
			},
			setupMocks: func(userRepo *mockUserRepository, otpService *mockOTPService, tokenService *mockTokenService) {
				user := &domain.User{
					ID:         userID,
					Email:      "newuser@example.com",
					IsVerified: false,
					IsDisabled: false,
				}
				userRepo.On("GetByEmail", mock.Anything, "newuser@example.com").Return(user, nil)
				otpService.On("GenerateAndSend", mock.Anything, userID, "newuser@example.com", domain.TokenTypeLogin).Return(nil)
			},
		},
		{
			name: "user not found",
			request: service.RequestOTPRequest{
				Email: "notfound@example.com",
			},
			setupMocks: func(userRepo *mockUserRepository, otpService *mockOTPService, tokenService *mockTokenService) {
				userRepo.On("GetByEmail", mock.Anything, "notfound@example.com").Return(nil, nil)
			},
			expectedError: "user not found",
		},
		{
			name: "disabled user",
			request: service.RequestOTPRequest{
				Email: "disabled@example.com",
			},
			setupMocks: func(userRepo *mockUserRepository, otpService *mockOTPService, tokenService *mockTokenService) {
				user := &domain.User{
					ID:             userID,
					Email:          "disabled@example.com",
					IsDisabled:     true,
					DisabledReason: &disabledReason,
				}
				userRepo.On("GetByEmail", mock.Anything, "disabled@example.com").Return(user, nil)
			},
			expectedError: "account is disabled: account suspended",
		},
		{
			name: "database error when getting user",
			request: service.RequestOTPRequest{
				Email: "user@example.com",
			},
			setupMocks: func(userRepo *mockUserRepository, otpService *mockOTPService, tokenService *mockTokenService) {
				userRepo.On("GetByEmail", mock.Anything, "user@example.com").Return(nil, errors.New("db error"))
			},
			expectedError: "failed to get user",
		},
		{
			name: "error generating OTP",
			request: service.RequestOTPRequest{
				Email: "user@example.com",
			},
			setupMocks: func(userRepo *mockUserRepository, otpService *mockOTPService, tokenService *mockTokenService) {
				user := &domain.User{
					ID:         userID,
					Email:      "user@example.com",
					IsVerified: true,
					IsDisabled: false,
				}
				userRepo.On("GetByEmail", mock.Anything, "user@example.com").Return(user, nil)
				otpService.On("GenerateAndSend", mock.Anything, userID, "user@example.com", domain.TokenTypeLogin).Return(errors.New("otp error"))
			},
			expectedError: "failed to send login code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authSvc, userRepo, otpService, tokenService := setupAuthService()
			tt.setupMocks(userRepo, otpService, tokenService)

			ctx := context.Background()
			response, err := authSvc.RequestOTP(ctx, tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, "Login code sent to your email.", response.Message)
			}

			userRepo.AssertExpectations(t)
			otpService.AssertExpectations(t)
			tokenService.AssertExpectations(t)
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	userID := primitive.NewObjectID()
	disabledReason := "account suspended"
	clientID := "test-client"

	tests := []struct {
		name          string
		request       service.LoginRequest
		setupMocks    func(*mockUserRepository, *mockOTPService, *mockTokenService)
		expectedError string
	}{
		{
			name: "successful login for verified user",
			request: service.LoginRequest{
				Email:    "user@example.com",
				OTP:      "123456",
				ClientID: clientID,
			},
			setupMocks: func(userRepo *mockUserRepository, otpService *mockOTPService, tokenService *mockTokenService) {
				user := &domain.User{
					ID:         userID,
					Email:      "user@example.com",
					IsVerified: true,
					IsDisabled: false,
				}
				userRepo.On("GetByEmail", mock.Anything, "user@example.com").Return(user, nil)
				otpService.On("Verify", mock.Anything, userID, "123456", domain.TokenTypeLogin).Return(nil)
				tokenService.On("GenerateTokenPair", mock.Anything, user, clientID).Return(&service.TokenPairResponse{
					AccessToken:  "access_token",
					RefreshToken: "refresh_token",
					TokenType:    "Bearer",
					ExpiresIn:    3600,
				}, nil)
			},
		},
		{
			name: "successful first-time login for unverified user",
			request: service.LoginRequest{
				Email:    "newuser@example.com",
				OTP:      "123456",
				ClientID: clientID,
			},
			setupMocks: func(userRepo *mockUserRepository, otpService *mockOTPService, tokenService *mockTokenService) {
				user := &domain.User{
					ID:         userID,
					Email:      "newuser@example.com",
					IsVerified: false,
					IsDisabled: false,
				}
				userRepo.On("GetByEmail", mock.Anything, "newuser@example.com").Return(user, nil)
				otpService.On("Verify", mock.Anything, userID, "123456", domain.TokenTypeLogin).Return(nil)
				userRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
					return u.ID == userID && u.IsVerified
				})).Return(nil)
				tokenService.On("GenerateTokenPair", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
					return u.ID == userID && u.IsVerified
				}), clientID).Return(&service.TokenPairResponse{
					AccessToken:  "access_token",
					RefreshToken: "refresh_token",
					TokenType:    "Bearer",
					ExpiresIn:    3600,
				}, nil)
			},
		},
		{
			name: "user not found",
			request: service.LoginRequest{
				Email:    "notfound@example.com",
				OTP:      "123456",
				ClientID: clientID,
			},
			setupMocks: func(userRepo *mockUserRepository, otpService *mockOTPService, tokenService *mockTokenService) {
				userRepo.On("GetByEmail", mock.Anything, "notfound@example.com").Return(nil, nil)
			},
			expectedError: "invalid credentials",
		},
		{
			name: "disabled user",
			request: service.LoginRequest{
				Email:    "disabled@example.com",
				OTP:      "123456",
				ClientID: clientID,
			},
			setupMocks: func(userRepo *mockUserRepository, otpService *mockOTPService, tokenService *mockTokenService) {
				user := &domain.User{
					ID:             userID,
					Email:          "disabled@example.com",
					IsDisabled:     true,
					DisabledReason: &disabledReason,
				}
				userRepo.On("GetByEmail", mock.Anything, "disabled@example.com").Return(user, nil)
			},
			expectedError: "account is disabled: account suspended",
		},
		{
			name: "invalid OTP",
			request: service.LoginRequest{
				Email:    "user@example.com",
				OTP:      "invalid",
				ClientID: clientID,
			},
			setupMocks: func(userRepo *mockUserRepository, otpService *mockOTPService, tokenService *mockTokenService) {
				user := &domain.User{
					ID:         userID,
					Email:      "user@example.com",
					IsVerified: true,
					IsDisabled: false,
				}
				userRepo.On("GetByEmail", mock.Anything, "user@example.com").Return(user, nil)
				otpService.On("Verify", mock.Anything, userID, "invalid", domain.TokenTypeLogin).Return(errors.New("invalid OTP"))
			},
			expectedError: "invalid OTP",
		},
		{
			name: "error updating user verification status",
			request: service.LoginRequest{
				Email:    "newuser@example.com",
				OTP:      "123456",
				ClientID: clientID,
			},
			setupMocks: func(userRepo *mockUserRepository, otpService *mockOTPService, tokenService *mockTokenService) {
				user := &domain.User{
					ID:         userID,
					Email:      "newuser@example.com",
					IsVerified: false,
					IsDisabled: false,
				}
				userRepo.On("GetByEmail", mock.Anything, "newuser@example.com").Return(user, nil)
				otpService.On("Verify", mock.Anything, userID, "123456", domain.TokenTypeLogin).Return(nil)
				userRepo.On("Update", mock.Anything, mock.Anything).Return(errors.New("update error"))
			},
			expectedError: "failed to verify user",
		},
		{
			name: "error generating token pair",
			request: service.LoginRequest{
				Email:    "user@example.com",
				OTP:      "123456",
				ClientID: clientID,
			},
			setupMocks: func(userRepo *mockUserRepository, otpService *mockOTPService, tokenService *mockTokenService) {
				user := &domain.User{
					ID:         userID,
					Email:      "user@example.com",
					IsVerified: true,
					IsDisabled: false,
				}
				userRepo.On("GetByEmail", mock.Anything, "user@example.com").Return(user, nil)
				otpService.On("Verify", mock.Anything, userID, "123456", domain.TokenTypeLogin).Return(nil)
				tokenService.On("GenerateTokenPair", mock.Anything, user, clientID).Return(nil, errors.New("token error"))
			},
			expectedError: "token error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authSvc, userRepo, otpService, tokenService := setupAuthService()
			tt.setupMocks(userRepo, otpService, tokenService)

			ctx := context.Background()
			response, err := authSvc.Login(ctx, tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, "access_token", response.AccessToken)
				assert.Equal(t, "refresh_token", response.RefreshToken)
				assert.Equal(t, "Bearer", response.TokenType)
				assert.Equal(t, int64(3600), response.ExpiresIn)
				assert.NotNil(t, response.User)
			}

			userRepo.AssertExpectations(t)
			otpService.AssertExpectations(t)
			tokenService.AssertExpectations(t)
		})
	}
}

func TestAuthService_RefreshToken(t *testing.T) {
	clientID := "test-client"
	refreshToken := "refresh_token"

	tests := []struct {
		name          string
		request       service.RefreshTokenRequest
		setupMocks    func(*mockUserRepository, *mockOTPService, *mockTokenService)
		expectedError string
	}{
		{
			name: "successful token refresh",
			request: service.RefreshTokenRequest{
				RefreshToken: refreshToken,
				ClientID:     clientID,
			},
			setupMocks: func(userRepo *mockUserRepository, otpService *mockOTPService, tokenService *mockTokenService) {
				tokenService.On("RefreshTokenPair", mock.Anything, refreshToken, clientID).Return(&service.TokenPairResponse{
					AccessToken:  "new_access_token",
					RefreshToken: "new_refresh_token",
					TokenType:    "Bearer",
					ExpiresIn:    3600,
				}, nil)
			},
		},
		{
			name: "error refreshing token",
			request: service.RefreshTokenRequest{
				RefreshToken: "invalid_token",
				ClientID:     clientID,
			},
			setupMocks: func(userRepo *mockUserRepository, otpService *mockOTPService, tokenService *mockTokenService) {
				tokenService.On("RefreshTokenPair", mock.Anything, "invalid_token", clientID).Return(nil, errors.New("invalid token"))
			},
			expectedError: "invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authSvc, userRepo, otpService, tokenService := setupAuthService()
			tt.setupMocks(userRepo, otpService, tokenService)

			ctx := context.Background()
			response, err := authSvc.RefreshToken(ctx, tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, "new_access_token", response.AccessToken)
				assert.Equal(t, "new_refresh_token", response.RefreshToken)
				assert.Equal(t, "Bearer", response.TokenType)
				assert.Equal(t, int64(3600), response.ExpiresIn)
			}

			userRepo.AssertExpectations(t)
			otpService.AssertExpectations(t)
			tokenService.AssertExpectations(t)
		})
	}
}

func TestAuthService_Logout(t *testing.T) {
	refreshToken := "refresh_token"

	tests := []struct {
		name          string
		request       service.LogoutRequest
		setupMocks    func(*mockUserRepository, *mockOTPService, *mockTokenService)
		expectedError string
	}{
		{
			name: "successful logout",
			request: service.LogoutRequest{
				RefreshToken: refreshToken,
			},
			setupMocks: func(userRepo *mockUserRepository, otpService *mockOTPService, tokenService *mockTokenService) {
				tokenService.On("RevokeTokenFamily", mock.Anything, refreshToken, "user_logout").Return(nil)
			},
		},
		{
			name: "error revoking token",
			request: service.LogoutRequest{
				RefreshToken: "invalid_token",
			},
			setupMocks: func(userRepo *mockUserRepository, otpService *mockOTPService, tokenService *mockTokenService) {
				tokenService.On("RevokeTokenFamily", mock.Anything, "invalid_token", "user_logout").Return(errors.New("revoke error"))
			},
			expectedError: "revoke error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authSvc, userRepo, otpService, tokenService := setupAuthService()
			tt.setupMocks(userRepo, otpService, tokenService)

			ctx := context.Background()
			response, err := authSvc.Logout(ctx, tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, response)
			} else {
				require.NoError(t, err)
				require.NotNil(t, response)
				assert.Equal(t, "Logged out successfully", response.Message)
			}

			userRepo.AssertExpectations(t)
			otpService.AssertExpectations(t)
			tokenService.AssertExpectations(t)
		})
	}
}

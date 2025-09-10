package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"backend/internal/domain"
	"backend/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type mockRefreshTokenRepository struct {
	mock.Mock
}

func (m *mockRefreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		token.ID = primitive.NewObjectID()
	}
	return args.Error(0)
}

func (m *mockRefreshTokenRepository) CreateWithPlainToken(ctx context.Context, token *domain.RefreshToken, plainToken string) error {
	args := m.Called(ctx, token, plainToken)
	if args.Get(0) == nil {
		token.ID = primitive.NewObjectID()
	}
	return args.Error(0)
}

func (m *mockRefreshTokenRepository) GetValidTokenForUser(ctx context.Context, plainToken string) (*domain.RefreshToken, error) {
	args := m.Called(ctx, plainToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RefreshToken), args.Error(1)
}

func (m *mockRefreshTokenRepository) RevokeByID(ctx context.Context, id primitive.ObjectID, reason string) error {
	args := m.Called(ctx, id, reason)
	return args.Error(0)
}

func (m *mockRefreshTokenRepository) RevokeByClientID(ctx context.Context, userID primitive.ObjectID, clientID, reason string) error {
	args := m.Called(ctx, userID, clientID, reason)
	return args.Error(0)
}

func (m *mockRefreshTokenRepository) RevokeByFamilyID(ctx context.Context, familyID, reason string) error {
	args := m.Called(ctx, familyID, reason)
	return args.Error(0)
}

func (m *mockRefreshTokenRepository) UpdateLastUsed(ctx context.Context, id primitive.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockRefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockRefreshTokenRepository) DeleteByUserID(ctx context.Context, userID primitive.ObjectID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *mockRefreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	args := m.Called(ctx, tokenHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RefreshToken), args.Error(1)
}

func (m *mockRefreshTokenRepository) RevokeByHash(ctx context.Context, tokenHash string) error {
	args := m.Called(ctx, tokenHash)
	return args.Error(0)
}

func (m *mockRefreshTokenRepository) GetActiveByUserID(ctx context.Context, userID primitive.ObjectID) ([]*domain.RefreshToken, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.RefreshToken), args.Error(1)
}

type mockJWTService struct {
	mock.Mock
}

func (m *mockJWTService) GenerateTokenPair(user *domain.User, clientID string) (*service.TokenPairResponse, error) {
	args := m.Called(user, clientID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.TokenPairResponse), args.Error(1)
}

func (m *mockJWTService) GenerateAccessToken(user *domain.User) (string, error) {
	args := m.Called(user)
	return args.String(0), args.Error(1)
}

func (m *mockJWTService) GenerateRefreshToken() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *mockJWTService) ValidateAccessToken(tokenString string) (*service.TokenClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.TokenClaims), args.Error(1)
}

func setupTokenService() (*tokenService, *mockUserRepository, *mockRefreshTokenRepository, *mockJWTService) {
	userRepo := &mockUserRepository{}
	refreshTokenRepo := &mockRefreshTokenRepository{}
	jwtSvc := &mockJWTService{}
	
	tokenSvc := &tokenService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtService:       jwtSvc,
	}
	
	return tokenSvc, userRepo, refreshTokenRepo, jwtSvc
}

func TestTokenService_GenerateTokenPair(t *testing.T) {
	userID := primitive.NewObjectID()
	clientID := "test-client"
	
	user := &domain.User{
		ID:    userID,
		Email: "test@example.com",
		Role:  domain.UserRoleCustomer,
	}
	
	tests := []struct {
		name          string
		user          *domain.User
		clientID      string
		setupMocks    func(*mockUserRepository, *mockRefreshTokenRepository, *mockJWTService)
		expectedError string
	}{
		{
			name:     "successful token pair generation",
			user:     user,
			clientID: clientID,
			setupMocks: func(userRepo *mockUserRepository, refreshTokenRepo *mockRefreshTokenRepository, jwtSvc *mockJWTService) {
				refreshTokenRepo.On("RevokeByClientID", mock.Anything, userID, clientID, "new_login").Return(nil)
				jwtSvc.On("GenerateTokenPair", user, clientID).Return(&service.TokenPairResponse{
					AccessToken:  "access_token",
					RefreshToken: "refresh_token",
					TokenType:    "Bearer",
					ExpiresIn:    3600,
				}, nil)
				refreshTokenRepo.On("CreateWithPlainToken", mock.Anything, mock.MatchedBy(func(token *domain.RefreshToken) bool {
					return token.UserID == userID &&
						token.ClientID == clientID &&
						len(token.FamilyID) > 0 &&
						token.ExpiresAt.After(time.Now())
				}), "refresh_token").Return(nil)
			},
		},
		{
			name:     "error revoking existing tokens",
			user:     user,
			clientID: clientID,
			setupMocks: func(userRepo *mockUserRepository, refreshTokenRepo *mockRefreshTokenRepository, jwtSvc *mockJWTService) {
				refreshTokenRepo.On("RevokeByClientID", mock.Anything, userID, clientID, "new_login").Return(errors.New("db error"))
			},
			expectedError: "failed to revoke existing refresh tokens",
		},
		{
			name:     "error generating JWT token pair",
			user:     user,
			clientID: clientID,
			setupMocks: func(userRepo *mockUserRepository, refreshTokenRepo *mockRefreshTokenRepository, jwtSvc *mockJWTService) {
				refreshTokenRepo.On("RevokeByClientID", mock.Anything, userID, clientID, "new_login").Return(nil)
				jwtSvc.On("GenerateTokenPair", user, clientID).Return(nil, errors.New("jwt error"))
			},
			expectedError: "failed to generate tokens",
		},
		{
			name:     "error storing refresh token",
			user:     user,
			clientID: clientID,
			setupMocks: func(userRepo *mockUserRepository, refreshTokenRepo *mockRefreshTokenRepository, jwtSvc *mockJWTService) {
				refreshTokenRepo.On("RevokeByClientID", mock.Anything, userID, clientID, "new_login").Return(nil)
				jwtSvc.On("GenerateTokenPair", user, clientID).Return(&service.TokenPairResponse{
					AccessToken:  "access_token",
					RefreshToken: "refresh_token",
					TokenType:    "Bearer",
					ExpiresIn:    3600,
				}, nil)
				refreshTokenRepo.On("CreateWithPlainToken", mock.Anything, mock.Anything, "refresh_token").Return(errors.New("db error"))
			},
			expectedError: "failed to store refresh token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenSvc, userRepo, refreshTokenRepo, jwtSvc := setupTokenService()
			tt.setupMocks(userRepo, refreshTokenRepo, jwtSvc)

			ctx := context.Background()
			response, err := tokenSvc.GenerateTokenPair(ctx, tt.user, tt.clientID)

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
			}

			userRepo.AssertExpectations(t)
			refreshTokenRepo.AssertExpectations(t)
			jwtSvc.AssertExpectations(t)
		})
	}
}

func TestTokenService_RefreshTokenPair(t *testing.T) {
	userID := primitive.NewObjectID()
	tokenID := primitive.NewObjectID()
	clientID := "test-client"
	familyID := "family-123"
	refreshToken := "refresh_token"
	disabledReason := "account suspended"
	
	storedToken := &domain.RefreshToken{
		ID:       tokenID,
		UserID:   userID,
		ClientID: clientID,
		FamilyID: familyID,
	}
	
	user := &domain.User{
		ID:    userID,
		Email: "test@example.com",
		Role:  domain.UserRoleCustomer,
	}
	
	tests := []struct {
		name          string
		refreshToken  string
		clientID      string
		setupMocks    func(*mockUserRepository, *mockRefreshTokenRepository, *mockJWTService)
		expectedError string
	}{
		{
			name:         "successful token refresh",
			refreshToken: refreshToken,
			clientID:     clientID,
			setupMocks: func(userRepo *mockUserRepository, refreshTokenRepo *mockRefreshTokenRepository, jwtSvc *mockJWTService) {
				refreshTokenRepo.On("GetValidTokenForUser", mock.Anything, refreshToken).Return(storedToken, nil)
				userRepo.On("GetByID", mock.Anything, userID).Return(user, nil)
				jwtSvc.On("GenerateTokenPair", user, clientID).Return(&service.TokenPairResponse{
					AccessToken:  "new_access_token",
					RefreshToken: "new_refresh_token",
					TokenType:    "Bearer",
					ExpiresIn:    3600,
				}, nil)
				refreshTokenRepo.On("RevokeByID", mock.Anything, tokenID, "token_rotation").Return(nil)
				refreshTokenRepo.On("CreateWithPlainToken", mock.Anything, mock.MatchedBy(func(token *domain.RefreshToken) bool {
					return token.UserID == userID &&
						token.ClientID == clientID &&
						token.FamilyID == familyID
				}), "new_refresh_token").Return(nil)
				refreshTokenRepo.On("UpdateLastUsed", mock.Anything, tokenID).Return(nil)
			},
		},
		{
			name:         "refresh token not found",
			refreshToken: "invalid_token",
			clientID:     clientID,
			setupMocks: func(userRepo *mockUserRepository, refreshTokenRepo *mockRefreshTokenRepository, jwtSvc *mockJWTService) {
				refreshTokenRepo.On("GetValidTokenForUser", mock.Anything, "invalid_token").Return(nil, domain.ErrRefreshTokenNotFound)
			},
			expectedError: "invalid refresh token",
		},
		{
			name:         "client ID mismatch",
			refreshToken: refreshToken,
			clientID:     "wrong-client",
			setupMocks: func(userRepo *mockUserRepository, refreshTokenRepo *mockRefreshTokenRepository, jwtSvc *mockJWTService) {
				refreshTokenRepo.On("GetValidTokenForUser", mock.Anything, refreshToken).Return(storedToken, nil)
			},
			expectedError: "invalid client",
		},
		{
			name:         "user not found",
			refreshToken: refreshToken,
			clientID:     clientID,
			setupMocks: func(userRepo *mockUserRepository, refreshTokenRepo *mockRefreshTokenRepository, jwtSvc *mockJWTService) {
				refreshTokenRepo.On("GetValidTokenForUser", mock.Anything, refreshToken).Return(storedToken, nil)
				userRepo.On("GetByID", mock.Anything, userID).Return(nil, nil)
			},
			expectedError: "user not found",
		},
		{
			name:         "disabled user",
			refreshToken: refreshToken,
			clientID:     clientID,
			setupMocks: func(userRepo *mockUserRepository, refreshTokenRepo *mockRefreshTokenRepository, jwtSvc *mockJWTService) {
				disabledUser := &domain.User{
					ID:             userID,
					Email:          "test@example.com",
					Role:           domain.UserRoleCustomer,
					IsDisabled:     true,
					DisabledReason: &disabledReason,
				}
				refreshTokenRepo.On("GetValidTokenForUser", mock.Anything, refreshToken).Return(storedToken, nil)
				userRepo.On("GetByID", mock.Anything, userID).Return(disabledUser, nil)
			},
			expectedError: "account is disabled",
		},
		{
			name:         "error generating new token pair",
			refreshToken: refreshToken,
			clientID:     clientID,
			setupMocks: func(userRepo *mockUserRepository, refreshTokenRepo *mockRefreshTokenRepository, jwtSvc *mockJWTService) {
				refreshTokenRepo.On("GetValidTokenForUser", mock.Anything, refreshToken).Return(storedToken, nil)
				userRepo.On("GetByID", mock.Anything, userID).Return(user, nil)
				jwtSvc.On("GenerateTokenPair", user, clientID).Return(nil, errors.New("jwt error"))
			},
			expectedError: "failed to generate tokens",
		},
		{
			name:         "error revoking old token",
			refreshToken: refreshToken,
			clientID:     clientID,
			setupMocks: func(userRepo *mockUserRepository, refreshTokenRepo *mockRefreshTokenRepository, jwtSvc *mockJWTService) {
				refreshTokenRepo.On("GetValidTokenForUser", mock.Anything, refreshToken).Return(storedToken, nil)
				userRepo.On("GetByID", mock.Anything, userID).Return(user, nil)
				jwtSvc.On("GenerateTokenPair", user, clientID).Return(&service.TokenPairResponse{
					AccessToken:  "new_access_token",
					RefreshToken: "new_refresh_token",
					TokenType:    "Bearer",
					ExpiresIn:    3600,
				}, nil)
				refreshTokenRepo.On("RevokeByID", mock.Anything, tokenID, "token_rotation").Return(errors.New("db error"))
			},
			expectedError: "failed to rotate token",
		},
		{
			name:         "error storing new refresh token",
			refreshToken: refreshToken,
			clientID:     clientID,
			setupMocks: func(userRepo *mockUserRepository, refreshTokenRepo *mockRefreshTokenRepository, jwtSvc *mockJWTService) {
				refreshTokenRepo.On("GetValidTokenForUser", mock.Anything, refreshToken).Return(storedToken, nil)
				userRepo.On("GetByID", mock.Anything, userID).Return(user, nil)
				jwtSvc.On("GenerateTokenPair", user, clientID).Return(&service.TokenPairResponse{
					AccessToken:  "new_access_token",
					RefreshToken: "new_refresh_token",
					TokenType:    "Bearer",
					ExpiresIn:    3600,
				}, nil)
				refreshTokenRepo.On("RevokeByID", mock.Anything, tokenID, "token_rotation").Return(nil)
				refreshTokenRepo.On("CreateWithPlainToken", mock.Anything, mock.Anything, "new_refresh_token").Return(errors.New("db error"))
			},
			expectedError: "failed to store refresh token",
		},
		{
			name:         "database error when getting refresh token",
			refreshToken: refreshToken,
			clientID:     clientID,
			setupMocks: func(userRepo *mockUserRepository, refreshTokenRepo *mockRefreshTokenRepository, jwtSvc *mockJWTService) {
				refreshTokenRepo.On("GetValidTokenForUser", mock.Anything, refreshToken).Return(nil, errors.New("db error"))
			},
			expectedError: "failed to validate refresh token",
		},
		{
			name:         "database error when getting user",
			refreshToken: refreshToken,
			clientID:     clientID,
			setupMocks: func(userRepo *mockUserRepository, refreshTokenRepo *mockRefreshTokenRepository, jwtSvc *mockJWTService) {
				refreshTokenRepo.On("GetValidTokenForUser", mock.Anything, refreshToken).Return(storedToken, nil)
				userRepo.On("GetByID", mock.Anything, userID).Return(nil, errors.New("db error"))
			},
			expectedError: "failed to get user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenSvc, userRepo, refreshTokenRepo, jwtSvc := setupTokenService()
			tt.setupMocks(userRepo, refreshTokenRepo, jwtSvc)

			ctx := context.Background()
			response, err := tokenSvc.RefreshTokenPair(ctx, tt.refreshToken, tt.clientID)

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
			refreshTokenRepo.AssertExpectations(t)
			jwtSvc.AssertExpectations(t)
		})
	}
}

func TestTokenService_RevokeTokenFamily(t *testing.T) {
	userID := primitive.NewObjectID()
	tokenID := primitive.NewObjectID()
	familyID := "family-123"
	refreshToken := "refresh_token"
	
	storedToken := &domain.RefreshToken{
		ID:       tokenID,
		UserID:   userID,
		ClientID: "test-client",
		FamilyID: familyID,
	}
	
	tests := []struct {
		name          string
		refreshToken  string
		reason        string
		setupMocks    func(*mockUserRepository, *mockRefreshTokenRepository, *mockJWTService)
		expectedError string
	}{
		{
			name:         "successful token family revocation",
			refreshToken: refreshToken,
			reason:       "user_logout",
			setupMocks: func(userRepo *mockUserRepository, refreshTokenRepo *mockRefreshTokenRepository, jwtSvc *mockJWTService) {
				refreshTokenRepo.On("GetValidTokenForUser", mock.Anything, refreshToken).Return(storedToken, nil)
				refreshTokenRepo.On("RevokeByFamilyID", mock.Anything, familyID, "user_logout").Return(nil)
			},
		},
		{
			name:         "token not found - should not error (logout case)",
			refreshToken: "non_existent_token",
			reason:       "user_logout",
			setupMocks: func(userRepo *mockUserRepository, refreshTokenRepo *mockRefreshTokenRepository, jwtSvc *mockJWTService) {
				refreshTokenRepo.On("GetValidTokenForUser", mock.Anything, "non_existent_token").Return(nil, domain.ErrRefreshTokenNotFound)
			},
		},
		{
			name:         "database error when getting token",
			refreshToken: refreshToken,
			reason:       "user_logout",
			setupMocks: func(userRepo *mockUserRepository, refreshTokenRepo *mockRefreshTokenRepository, jwtSvc *mockJWTService) {
				refreshTokenRepo.On("GetValidTokenForUser", mock.Anything, refreshToken).Return(nil, errors.New("db error"))
			},
			expectedError: "failed to validate refresh token",
		},
		{
			name:         "error revoking token family",
			refreshToken: refreshToken,
			reason:       "user_logout",
			setupMocks: func(userRepo *mockUserRepository, refreshTokenRepo *mockRefreshTokenRepository, jwtSvc *mockJWTService) {
				refreshTokenRepo.On("GetValidTokenForUser", mock.Anything, refreshToken).Return(storedToken, nil)
				refreshTokenRepo.On("RevokeByFamilyID", mock.Anything, familyID, "user_logout").Return(errors.New("db error"))
			},
			expectedError: "failed to revoke tokens",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenSvc, userRepo, refreshTokenRepo, jwtSvc := setupTokenService()
			tt.setupMocks(userRepo, refreshTokenRepo, jwtSvc)

			ctx := context.Background()
			err := tokenSvc.RevokeTokenFamily(ctx, tt.refreshToken, tt.reason)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			userRepo.AssertExpectations(t)
			refreshTokenRepo.AssertExpectations(t)
			jwtSvc.AssertExpectations(t)
		})
	}
}
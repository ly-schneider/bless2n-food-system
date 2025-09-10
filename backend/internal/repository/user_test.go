package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"backend/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// mockUserRepository implements the UserRepository interface for testing
type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	if args.Error(0) == nil {
		user.ID = primitive.NewObjectID()
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
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
	if args.Error(0) == nil {
		user.UpdatedAt = time.Now()
	}
	return args.Error(0)
}

func (m *mockUserRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockUserRepository) List(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.User), args.Error(1)
}

func (m *mockUserRepository) ListCustomers(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
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

// Test the UserRepository interface behavior through mocks
func TestUserRepositoryInterface_Create(t *testing.T) {
	tests := []struct {
		name          string
		user          *domain.User
		setupMocks    func(*mockUserRepository)
		expectedError bool
	}{
		{
			name: "successful user creation",
			user: &domain.User{
				Email: "test@example.com",
				Role:  domain.UserRoleCustomer,
			},
			setupMocks: func(mockRepo *mockUserRepository) {
				mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(user *domain.User) bool {
					return user.Email == "test@example.com" && user.Role == domain.UserRoleCustomer
				})).Return(nil)
			},
		},
		{
			name: "database error during creation",
			user: &domain.User{
				Email: "test@example.com",
				Role:  domain.UserRoleCustomer,
			},
			setupMocks: func(mockRepo *mockUserRepository) {
				mockRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockUserRepository{}
			tt.setupMocks(mockRepo)

			ctx := context.Background()
			err := mockRepo.Create(ctx, tt.user)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				// Verify that the mock properly sets the ID and timestamps
				assert.False(t, tt.user.ID.IsZero())
				assert.False(t, tt.user.CreatedAt.IsZero())
				assert.False(t, tt.user.UpdatedAt.IsZero())
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserRepositoryInterface_GetByID(t *testing.T) {
	userID := primitive.NewObjectID()
	expectedUser := &domain.User{
		ID:    userID,
		Email: "test@example.com",
		Role:  domain.UserRoleCustomer,
	}

	tests := []struct {
		name          string
		userID        primitive.ObjectID
		setupMocks    func(*mockUserRepository)
		expectedUser  *domain.User
		expectedError bool
	}{
		{
			name:   "successful user retrieval",
			userID: userID,
			setupMocks: func(mockRepo *mockUserRepository) {
				mockRepo.On("GetByID", mock.Anything, userID).Return(expectedUser, nil)
			},
			expectedUser: expectedUser,
		},
		{
			name:   "user not found",
			userID: userID,
			setupMocks: func(mockRepo *mockUserRepository) {
				mockRepo.On("GetByID", mock.Anything, userID).Return(nil, nil)
			},
			expectedUser: nil,
		},
		{
			name:   "database error",
			userID: userID,
			setupMocks: func(mockRepo *mockUserRepository) {
				mockRepo.On("GetByID", mock.Anything, userID).Return(nil, errors.New("db error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockUserRepository{}
			tt.setupMocks(mockRepo)

			ctx := context.Background()
			user, err := mockRepo.GetByID(ctx, tt.userID)

			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				if tt.expectedUser == nil {
					assert.Nil(t, user)
				} else {
					require.NotNil(t, user)
					assert.Equal(t, tt.expectedUser.ID, user.ID)
					assert.Equal(t, tt.expectedUser.Email, user.Email)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserRepositoryInterface_GetByEmail(t *testing.T) {
	email := "test@example.com"
	expectedUser := &domain.User{
		ID:    primitive.NewObjectID(),
		Email: email,
		Role:  domain.UserRoleCustomer,
	}

	tests := []struct {
		name          string
		email         string
		setupMocks    func(*mockUserRepository)
		expectedUser  *domain.User
		expectedError bool
	}{
		{
			name:  "successful user retrieval by email",
			email: email,
			setupMocks: func(mockRepo *mockUserRepository) {
				mockRepo.On("GetByEmail", mock.Anything, email).Return(expectedUser, nil)
			},
			expectedUser: expectedUser,
		},
		{
			name:  "user not found by email",
			email: "notfound@example.com",
			setupMocks: func(mockRepo *mockUserRepository) {
				mockRepo.On("GetByEmail", mock.Anything, "notfound@example.com").Return(nil, nil)
			},
			expectedUser: nil,
		},
		{
			name:  "database error when finding by email",
			email: email,
			setupMocks: func(mockRepo *mockUserRepository) {
				mockRepo.On("GetByEmail", mock.Anything, email).Return(nil, errors.New("db error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockUserRepository{}
			tt.setupMocks(mockRepo)

			ctx := context.Background()
			user, err := mockRepo.GetByEmail(ctx, tt.email)

			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				if tt.expectedUser == nil {
					assert.Nil(t, user)
				} else {
					require.NotNil(t, user)
					assert.Equal(t, tt.expectedUser.ID, user.ID)
					assert.Equal(t, tt.expectedUser.Email, user.Email)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserRepositoryInterface_CountCustomers(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*mockUserRepository)
		expectedCount int
		expectedError bool
	}{
		{
			name: "successful customer count",
			setupMocks: func(mockRepo *mockUserRepository) {
				mockRepo.On("CountCustomers", mock.Anything).Return(5, nil)
			},
			expectedCount: 5,
		},
		{
			name: "database error during count",
			setupMocks: func(mockRepo *mockUserRepository) {
				mockRepo.On("CountCustomers", mock.Anything).Return(0, errors.New("db error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockUserRepository{}
			tt.setupMocks(mockRepo)

			ctx := context.Background()
			count, err := mockRepo.CountCustomers(ctx)

			if tt.expectedError {
				require.Error(t, err)
				assert.Zero(t, count)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedCount, count)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserRepositoryInterface_Disable(t *testing.T) {
	userID := primitive.NewObjectID()
	reason := "account violation"

	tests := []struct {
		name          string
		userID        primitive.ObjectID
		reason        string
		setupMocks    func(*mockUserRepository)
		expectedError bool
	}{
		{
			name:   "successful user disable",
			userID: userID,
			reason: reason,
			setupMocks: func(mockRepo *mockUserRepository) {
				mockRepo.On("Disable", mock.Anything, userID, reason).Return(nil)
			},
		},
		{
			name:   "database error during disable",
			userID: userID,
			reason: reason,
			setupMocks: func(mockRepo *mockUserRepository) {
				mockRepo.On("Disable", mock.Anything, userID, reason).Return(errors.New("db error"))
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockUserRepository{}
			tt.setupMocks(mockRepo)

			ctx := context.Background()
			err := mockRepo.Disable(ctx, tt.userID, tt.reason)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
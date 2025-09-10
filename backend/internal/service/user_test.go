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

func setupUserService() (*userService, *testutil.MockUserRepository) {
	mockRepo := &testutil.MockUserRepository{}
	service := &userService{
		userRepo: mockRepo,
	}
	return service, mockRepo
}

func TestUserService_UpdateProfile(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		request       UpdateProfileRequest
		setupMocks    func(*testutil.MockUserRepository)
		expectedError string
		expectSuccess bool
	}{
		{
			name:   "successful profile update",
			userID: primitive.NewObjectID().Hex(),
			request: UpdateProfileRequest{
				Email: "updated@example.com",
			},
			setupMocks: func(mockRepo *testutil.MockUserRepository) {
				userID, _ := primitive.ObjectIDFromHex(primitive.NewObjectID().Hex())
				user := testutil.CreateTestCustomer("old@example.com")
				user.ID = userID

                mockRepo.On("GetByID", mock.Anything, mock.Anything).Return(user, nil)
				mockRepo.On("GetByEmail", mock.Anything, "updated@example.com").Return(nil, nil)
				mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
					return u.Email == "updated@example.com"
				})).Return(nil)
			},
			expectSuccess: true,
		},
		{
			name:   "invalid user ID format",
			userID: "invalid-id",
			request: UpdateProfileRequest{
				Email: "test@example.com",
			},
			setupMocks:    func(mockRepo *testutil.MockUserRepository) {},
			expectedError: "invalid user ID format",
		},
		{
			name:   "user not found",
			userID: primitive.NewObjectID().Hex(),
			request: UpdateProfileRequest{
				Email: "test@example.com",
			},
			setupMocks: func(mockRepo *testutil.MockUserRepository) {
                mockRepo.On("GetByID", mock.Anything, mock.Anything).Return(nil, nil)
			},
			expectedError: "user not found",
		},
		{
			name:   "disabled user",
			userID: primitive.NewObjectID().Hex(),
			request: UpdateProfileRequest{
				Email: "test@example.com",
			},
			setupMocks: func(mockRepo *testutil.MockUserRepository) {
				userID, _ := primitive.ObjectIDFromHex(primitive.NewObjectID().Hex())
				user := testutil.CreateTestCustomer("old@example.com")
				user.ID = userID
				user.IsDisabled = true

                mockRepo.On("GetByID", mock.Anything, mock.Anything).Return(user, nil)
			},
			expectedError: "account is disabled",
		},
		{
			name:   "email already in use by another user",
			userID: primitive.NewObjectID().Hex(),
			request: UpdateProfileRequest{
				Email: "existing@example.com",
			},
			setupMocks: func(mockRepo *testutil.MockUserRepository) {
				userID, _ := primitive.ObjectIDFromHex(primitive.NewObjectID().Hex())
				user := testutil.CreateTestCustomer("old@example.com")
				user.ID = userID

				existingUser := testutil.CreateTestCustomer("existing@example.com")

                mockRepo.On("GetByID", mock.Anything, mock.Anything).Return(user, nil)
				mockRepo.On("GetByEmail", mock.Anything, "existing@example.com").Return(existingUser, nil)
			},
			expectedError: "email address is already in use",
		},
		{
			name:   "database error when checking user",
			userID: primitive.NewObjectID().Hex(),
			request: UpdateProfileRequest{
				Email: "test@example.com",
			},
			setupMocks: func(mockRepo *testutil.MockUserRepository) {
                mockRepo.On("GetByID", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
			},
			expectedError: "failed to get user",
		},
		{
			name:   "database error when checking email availability",
			userID: primitive.NewObjectID().Hex(),
			request: UpdateProfileRequest{
				Email: "test@example.com",
			},
			setupMocks: func(mockRepo *testutil.MockUserRepository) {
				userID, _ := primitive.ObjectIDFromHex(primitive.NewObjectID().Hex())
				user := testutil.CreateTestCustomer("old@example.com")
				user.ID = userID

                mockRepo.On("GetByID", mock.Anything, mock.Anything).Return(user, nil)
				mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, errors.New("db error"))
			},
			expectedError: "failed to check email availability",
		},
		{
			name:   "database error when updating",
			userID: primitive.NewObjectID().Hex(),
			request: UpdateProfileRequest{
				Email: "test@example.com",
			},
			setupMocks: func(mockRepo *testutil.MockUserRepository) {
				userID, _ := primitive.ObjectIDFromHex(primitive.NewObjectID().Hex())
				user := testutil.CreateTestCustomer("old@example.com")
				user.ID = userID

                mockRepo.On("GetByID", mock.Anything, mock.Anything).Return(user, nil)
				mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, nil)
				mockRepo.On("Update", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedError: "failed to update profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockRepo := setupUserService()
			tt.setupMocks(mockRepo)

    ctx := testutil.TestContext()
    response, err := service.UpdateProfile(ctx, tt.userID, tt.request)

			if tt.expectedError != "" {
				testutil.AssertErrorContains(t, err, tt.expectedError)
				assert.Nil(t, response)
			} else {
				testutil.AssertNoError(t, err)
				require.NotNil(t, response)
				if tt.expectSuccess {
					assert.True(t, response.Success)
					assert.Equal(t, "Profile updated successfully", response.Message)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_DeleteProfile(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		setupMocks    func(*testutil.MockUserRepository)
		expectedError string
		expectSuccess bool
	}{
		{
			name:   "successful profile deletion",
			userID: primitive.NewObjectID().Hex(),
			setupMocks: func(mockRepo *testutil.MockUserRepository) {
				userID, _ := primitive.ObjectIDFromHex(primitive.NewObjectID().Hex())
				user := testutil.CreateTestCustomer("customer@example.com")
				user.ID = userID

                mockRepo.On("GetByID", mock.Anything, mock.Anything).Return(user, nil)
                mockRepo.On("Delete", mock.Anything, mock.Anything).Return(nil)
			},
			expectSuccess: true,
		},
		{
			name:          "invalid user ID format",
			userID:        "invalid-id",
			setupMocks:    func(mockRepo *testutil.MockUserRepository) {},
			expectedError: "invalid user ID format",
		},
		{
			name:   "user not found",
			userID: primitive.NewObjectID().Hex(),
			setupMocks: func(mockRepo *testutil.MockUserRepository) {
                mockRepo.On("GetByID", mock.Anything, mock.Anything).Return(nil, nil)
			},
			expectedError: "user not found",
		},
		{
			name:   "admin cannot delete profile",
			userID: primitive.NewObjectID().Hex(),
			setupMocks: func(mockRepo *testutil.MockUserRepository) {
				userID, _ := primitive.ObjectIDFromHex(primitive.NewObjectID().Hex())
				user := testutil.CreateTestAdmin("admin@example.com")
				user.ID = userID

                mockRepo.On("GetByID", mock.Anything, mock.Anything).Return(user, nil)
			},
			expectedError: "only customer accounts can be self-deleted",
		},
		{
			name:   "database error when checking user",
			userID: primitive.NewObjectID().Hex(),
			setupMocks: func(mockRepo *testutil.MockUserRepository) {
                mockRepo.On("GetByID", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
			},
			expectedError: "failed to get user",
		},
		{
			name:   "database error when deleting",
			userID: primitive.NewObjectID().Hex(),
			setupMocks: func(mockRepo *testutil.MockUserRepository) {
				userID, _ := primitive.ObjectIDFromHex(primitive.NewObjectID().Hex())
				user := testutil.CreateTestCustomer("customer@example.com")
				user.ID = userID

                mockRepo.On("GetByID", mock.Anything, mock.Anything).Return(user, nil)
                mockRepo.On("Delete", mock.Anything, mock.Anything).Return(errors.New("db error"))
			},
			expectedError: "failed to delete profile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockRepo := setupUserService()
			tt.setupMocks(mockRepo)

			ctx := testutil.TestContext()
			response, err := service.DeleteProfile(ctx, tt.userID)

			if tt.expectedError != "" {
				testutil.AssertErrorContains(t, err, tt.expectedError)
				assert.Nil(t, response)
			} else {
				testutil.AssertNoError(t, err)
				require.NotNil(t, response)
				if tt.expectSuccess {
					assert.True(t, response.Success)
					assert.Equal(t, "Profile deleted successfully", response.Message)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_GetProfile(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		setupMocks    func(*testutil.MockUserRepository)
		expectedError string
		expectUser    bool
	}{
		{
			name:   "successful profile retrieval",
			userID: primitive.NewObjectID().Hex(),
			setupMocks: func(mockRepo *testutil.MockUserRepository) {
				userID, _ := primitive.ObjectIDFromHex(primitive.NewObjectID().Hex())
				user := testutil.CreateTestCustomer("customer@example.com")
				user.ID = userID

                mockRepo.On("GetByID", mock.Anything, mock.Anything).Return(user, nil)
			},
			expectUser: true,
		},
		{
			name:          "invalid user ID format",
			userID:        "invalid-id",
			setupMocks:    func(mockRepo *testutil.MockUserRepository) {},
			expectedError: "invalid user ID format",
		},
		{
			name:   "user not found",
			userID: primitive.NewObjectID().Hex(),
			setupMocks: func(mockRepo *testutil.MockUserRepository) {
                mockRepo.On("GetByID", mock.Anything, mock.Anything).Return(nil, nil)
			},
			expectedError: "user not found",
		},
		{
			name:   "database error",
			userID: primitive.NewObjectID().Hex(),
			setupMocks: func(mockRepo *testutil.MockUserRepository) {
                mockRepo.On("GetByID", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
			},
			expectedError: "failed to get user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, mockRepo := setupUserService()
			tt.setupMocks(mockRepo)

			ctx := testutil.TestContext()
			user, err := service.GetProfile(ctx, tt.userID)

			if tt.expectedError != "" {
				testutil.AssertErrorContains(t, err, tt.expectedError)
				assert.Nil(t, user)
			} else {
				testutil.AssertNoError(t, err)
				if tt.expectUser {
					require.NotNil(t, user)
					assert.NotEmpty(t, user.Email)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

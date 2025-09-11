package service

import (
    "context"
    "errors"
    "testing"

    "backend/internal/domain"
    "backend/internal/testutil"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
)

func setupStationService() (*stationService, *testutil.MockStationRepository, *testutil.MockStationProductRepository, *testutil.MockProductRepository, *testutil.MockUserRepository) {
    stationRepo := &testutil.MockStationRepository{}
    stationProductRepo := &testutil.MockStationProductRepository{}
    productRepo := &testutil.MockProductRepository{}
    userRepo := &testutil.MockUserRepository{}

    svc := &stationService{
        stationRepo:        stationRepo,
        stationProductRepo: stationProductRepo,
        productRepo:        productRepo,
        userRepo:           userRepo,
    }
    return svc, stationRepo, stationProductRepo, productRepo, userRepo
}

func TestStationService_CreateStation(t *testing.T) {
    tests := []struct {
        name        string
        req         *domain.StationRequest
        setupMocks  func(*testutil.MockStationRepository)
        expectError string
    }{
        {
            name: "successful creation when name is unique",
            req:  &domain.StationRequest{Name: "Station A"},
            setupMocks: func(stationRepo *testutil.MockStationRepository) {
                stationRepo.On("GetByName", mock.Anything, "Station A").Return(nil, nil)
                stationRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Station")).Return(nil)
            },
        },
        {
            name: "duplicate name returns error",
            req:  &domain.StationRequest{Name: "Dup"},
            setupMocks: func(stationRepo *testutil.MockStationRepository) {
                existing := &domain.Station{ID: primitive.NewObjectID(), Name: "Dup"}
                stationRepo.On("GetByName", mock.Anything, "Dup").Return(existing, nil)
            },
            expectError: "already exists",
        },
        {
            name: "repo error on GetByName bubbles up",
            req:  &domain.StationRequest{Name: "X"},
            setupMocks: func(stationRepo *testutil.MockStationRepository) {
                stationRepo.On("GetByName", mock.Anything, "X").Return(nil, errors.New("db error"))
            },
            expectError: "db error",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            svc, stationRepo, _, _, _ := setupStationService()
            tt.setupMocks(stationRepo)
            ctx := context.Background()
            resp, err := svc.CreateStation(ctx, tt.req)
            if tt.expectError != "" {
                require.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectError)
                assert.Nil(t, resp)
            } else {
                require.NoError(t, err)
                require.NotNil(t, resp)
                assert.Equal(t, tt.req.Name, resp.Name)
                assert.Equal(t, string(domain.StationStatusPending), resp.Status)
            }
            stationRepo.AssertExpectations(t)
        })
    }
}

func TestStationService_ApproveReject(t *testing.T) {
    stationID := primitive.NewObjectID()
    adminID := primitive.NewObjectID()

    t.Run("approve success", func(t *testing.T) {
        svc, stationRepo, _, _, _ := setupStationService()
        stationRepo.On("ApproveStation", mock.Anything, stationID, adminID).Return(nil)
        resp, err := svc.ApproveStation(context.Background(), stationID, adminID)
        require.NoError(t, err)
        require.NotNil(t, resp)
        assert.True(t, resp.Success)
        assert.Contains(t, resp.Message, "approved")
        stationRepo.AssertExpectations(t)
    })

    t.Run("approve not pending returns false success", func(t *testing.T) {
        svc, stationRepo, _, _, _ := setupStationService()
        stationRepo.On("ApproveStation", mock.Anything, stationID, adminID).Return(mongo.ErrNoDocuments)
        resp, err := svc.ApproveStation(context.Background(), stationID, adminID)
        require.NoError(t, err)
        require.NotNil(t, resp)
        assert.False(t, resp.Success)
        assert.Contains(t, resp.Message, "not pending")
        stationRepo.AssertExpectations(t)
    })

    t.Run("reject success", func(t *testing.T) {
        svc, stationRepo, _, _, _ := setupStationService()
        stationRepo.On("RejectStation", mock.Anything, stationID, adminID, "nope").Return(nil)
        resp, err := svc.RejectStation(context.Background(), stationID, adminID, "nope")
        require.NoError(t, err)
        require.NotNil(t, resp)
        assert.True(t, resp.Success)
        assert.Contains(t, resp.Message, "rejected")
        stationRepo.AssertExpectations(t)
    })
}

func TestStationService_AssignAndRemoveProducts(t *testing.T) {
    stationID := primitive.NewObjectID()
    prod1 := primitive.NewObjectID()
    prod2 := primitive.NewObjectID()

    t.Run("assign requires approved station and existing products", func(t *testing.T) {
        svc, stationRepo, stationProductRepo, productRepo, _ := setupStationService()
        station := &domain.Station{ID: stationID, Name: "S", Status: domain.StationStatusApproved}
        stationRepo.On("GetByID", mock.Anything, stationID).Return(station, nil)
        productRepo.On("GetByID", mock.Anything, prod1).Return(&domain.Product{ID: prod1, Name: "P1"}, nil)
        productRepo.On("GetByID", mock.Anything, prod2).Return(&domain.Product{ID: prod2, Name: "P2"}, nil)
        stationProductRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.StationProduct")).Return(nil).Twice()

        resp, err := svc.AssignProductsToStation(context.Background(), stationID, []primitive.ObjectID{prod1, prod2})
        require.NoError(t, err)
        require.NotNil(t, resp)
        assert.True(t, resp.Success)
        stationRepo.AssertExpectations(t)
        productRepo.AssertExpectations(t)
        stationProductRepo.AssertExpectations(t)
    })

    t.Run("assign fails when station not approved", func(t *testing.T) {
        svc, stationRepo, _, _, _ := setupStationService()
        station := &domain.Station{ID: stationID, Name: "S", Status: domain.StationStatusPending}
        stationRepo.On("GetByID", mock.Anything, stationID).Return(station, nil)
        resp, err := svc.AssignProductsToStation(context.Background(), stationID, []primitive.ObjectID{prod1})
        require.NoError(t, err)
        require.NotNil(t, resp)
        assert.False(t, resp.Success)
        assert.Contains(t, resp.Message, "approved")
        stationRepo.AssertExpectations(t)
    })

    t.Run("remove product from station", func(t *testing.T) {
        svc, _, stationProductRepo, _, _ := setupStationService()
        stationProductRepo.On("Delete", mock.Anything, stationID, prod1).Return(nil)
        resp, err := svc.RemoveProductFromStation(context.Background(), stationID, prod1)
        require.NoError(t, err)
        require.NotNil(t, resp)
        assert.True(t, resp.Success)
        stationProductRepo.AssertExpectations(t)
    })
}


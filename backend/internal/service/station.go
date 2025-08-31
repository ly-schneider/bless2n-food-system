package service

import (
	"context"
	"errors"

	"backend/internal/domain"
	"backend/internal/repository"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type StationService interface {
	CreateStation(ctx context.Context, req *domain.StationRequest) (*StationResponse, error)
	ListStations(ctx context.Context, limit, offset int) (*ListStationsResponse, error)
	ListStationsByStatus(ctx context.Context, status domain.StationStatus, limit, offset int) (*ListStationsResponse, error)
	GetStation(ctx context.Context, stationID primitive.ObjectID) (*StationResponse, error)
	ApproveStation(ctx context.Context, stationID, adminID primitive.ObjectID) (*ApprovalResponse, error)
	RejectStation(ctx context.Context, stationID, adminID primitive.ObjectID, reason string) (*ApprovalResponse, error)
	AssignProductsToStation(ctx context.Context, stationID primitive.ObjectID, productIDs []primitive.ObjectID) (*AssignProductsResponse, error)
	GetStationProducts(ctx context.Context, stationID primitive.ObjectID) (*StationProductsResponse, error)
	RemoveProductFromStation(ctx context.Context, stationID, productID primitive.ObjectID) (*AssignProductsResponse, error)
}

type StationResponse struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Status          string  `json:"status"`
	ApprovedBy      *string `json:"approved_by,omitempty"`
	ApprovedAt      *string `json:"approved_at,omitempty"`
	RejectedBy      *string `json:"rejected_by,omitempty"`
	RejectedAt      *string `json:"rejected_at,omitempty"`
	RejectionReason *string `json:"rejection_reason,omitempty"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

type ListStationsResponse struct {
	Stations []StationResponse `json:"stations"`
	Total    int               `json:"total"`
}

type ApprovalResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type AssignProductsResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

type Product struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type StationProductsResponse struct {
	StationID string    `json:"station_id"`
	Products  []Product `json:"products"`
}

type stationService struct {
	stationRepo        repository.StationRepository
	stationProductRepo repository.StationProductRepository
	productRepo        repository.ProductRepository
	userRepo           repository.UserRepository
}

func NewStationService(
	stationRepo repository.StationRepository,
	stationProductRepo repository.StationProductRepository,
	productRepo repository.ProductRepository,
	userRepo repository.UserRepository,
) StationService {
	return &stationService{
		stationRepo:        stationRepo,
		stationProductRepo: stationProductRepo,
		productRepo:        productRepo,
		userRepo:           userRepo,
	}
}

func (s *stationService) CreateStation(ctx context.Context, req *domain.StationRequest) (*StationResponse, error) {
	// Check if station name already exists
	existingStation, err := s.stationRepo.GetByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if existingStation != nil {
		return nil, errors.New("station with this name already exists")
	}

	station := &domain.Station{
		Name:   req.Name,
		Status: domain.StationStatusPending,
	}

	err = s.stationRepo.Create(ctx, station)
	if err != nil {
		return nil, err
	}

	return s.mapStationToResponse(station), nil
}

func (s *stationService) ListStations(ctx context.Context, limit, offset int) (*ListStationsResponse, error) {
	stations, err := s.stationRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]StationResponse, len(stations))
	for i, station := range stations {
		responses[i] = *s.mapStationToResponse(station)
	}

	return &ListStationsResponse{
		Stations: responses,
		Total:    len(responses),
	}, nil
}

func (s *stationService) ListStationsByStatus(ctx context.Context, status domain.StationStatus, limit, offset int) (*ListStationsResponse, error) {
	stations, err := s.stationRepo.ListByStatus(ctx, status, limit, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]StationResponse, len(stations))
	for i, station := range stations {
		responses[i] = *s.mapStationToResponse(station)
	}

	return &ListStationsResponse{
		Stations: responses,
		Total:    len(responses),
	}, nil
}


func (s *stationService) GetStation(ctx context.Context, stationID primitive.ObjectID) (*StationResponse, error) {
	station, err := s.stationRepo.GetByID(ctx, stationID)
	if err != nil {
		return nil, err
	}
	if station == nil {
		return nil, errors.New("station not found")
	}

	return s.mapStationToResponse(station), nil
}

func (s *stationService) ApproveStation(ctx context.Context, stationID, adminID primitive.ObjectID) (*ApprovalResponse, error) {
	err := s.stationRepo.ApproveStation(ctx, stationID, adminID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &ApprovalResponse{
				Message: "Station not found or not pending approval",
				Success: false,
			}, nil
		}
		return nil, err
	}

	return &ApprovalResponse{
		Message: "Station approved successfully",
		Success: true,
	}, nil
}

func (s *stationService) RejectStation(ctx context.Context, stationID, adminID primitive.ObjectID, reason string) (*ApprovalResponse, error) {
	err := s.stationRepo.RejectStation(ctx, stationID, adminID, reason)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return &ApprovalResponse{
				Message: "Station not found or not pending approval",
				Success: false,
			}, nil
		}
		return nil, err
	}

	return &ApprovalResponse{
		Message: "Station rejected successfully",
		Success: true,
	}, nil
}

func (s *stationService) AssignProductsToStation(ctx context.Context, stationID primitive.ObjectID, productIDs []primitive.ObjectID) (*AssignProductsResponse, error) {
	// Verify station exists and is approved
	station, err := s.stationRepo.GetByID(ctx, stationID)
	if err != nil {
		return nil, err
	}
	if station == nil {
		return &AssignProductsResponse{
			Message: "Station not found",
			Success: false,
		}, nil
	}
	if station.Status != domain.StationStatusApproved {
		return &AssignProductsResponse{
			Message: "Station must be approved to assign products",
			Success: false,
		}, nil
	}

	// Verify all products exist
	for _, productID := range productIDs {
		product, err := s.productRepo.GetByID(ctx, productID)
		if err != nil {
			return nil, err
		}
		if product == nil {
			return &AssignProductsResponse{
				Message: "One or more products not found",
				Success: false,
			}, nil
		}
	}

	// Create station-product associations
	for _, productID := range productIDs {
		stationProduct := &domain.StationProduct{
			StationID: stationID,
			ProductID: productID,
		}
		err := s.stationProductRepo.Create(ctx, stationProduct)
		if err != nil {
			return nil, err
		}
	}

	return &AssignProductsResponse{
		Message: "Products assigned to station successfully",
		Success: true,
	}, nil
}

func (s *stationService) GetStationProducts(ctx context.Context, stationID primitive.ObjectID) (*StationProductsResponse, error) {
	stationProducts, err := s.stationProductRepo.GetByStationID(ctx, stationID)
	if err != nil {
		return nil, err
	}

	products := make([]Product, 0, len(stationProducts))
	for _, sp := range stationProducts {
		product, err := s.productRepo.GetByID(ctx, sp.ProductID)
		if err != nil {
			return nil, err
		}
		if product != nil {
			products = append(products, Product{
				ID:    product.ID.Hex(),
				Name:  product.Name,
				Price: product.Price,
			})
		}
	}

	return &StationProductsResponse{
		StationID: stationID.Hex(),
		Products:  products,
	}, nil
}

func (s *stationService) RemoveProductFromStation(ctx context.Context, stationID, productID primitive.ObjectID) (*AssignProductsResponse, error) {
	err := s.stationProductRepo.Delete(ctx, stationID, productID)
	if err != nil {
		return nil, err
	}

	return &AssignProductsResponse{
		Message: "Product removed from station successfully",
		Success: true,
	}, nil
}

func (s *stationService) mapStationToResponse(station *domain.Station) *StationResponse {
	response := &StationResponse{
		ID:        station.ID.Hex(),
		Name:      station.Name,
		Status:    string(station.Status),
		CreatedAt: station.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: station.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if station.ApprovedBy != nil {
		approvedBy := station.ApprovedBy.Hex()
		response.ApprovedBy = &approvedBy
	}
	if station.ApprovedAt != nil {
		approvedAt := station.ApprovedAt.Format("2006-01-02T15:04:05Z")
		response.ApprovedAt = &approvedAt
	}
	if station.RejectedBy != nil {
		rejectedBy := station.RejectedBy.Hex()
		response.RejectedBy = &rejectedBy
	}
	if station.RejectedAt != nil {
		rejectedAt := station.RejectedAt.Format("2006-01-02T15:04:05Z")
		response.RejectedAt = &rejectedAt
	}
	if station.RejectionReason != nil {
		response.RejectionReason = station.RejectionReason
	}

	return response
}
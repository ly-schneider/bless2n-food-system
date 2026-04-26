package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backend/internal/config"
	"backend/internal/generated/ent"
	"backend/internal/generated/ent/device"
	"backend/internal/generated/ent/orderline"
	"backend/internal/repository"

	"github.com/google/uuid"
)

type StationService interface {
	ListStations(ctx context.Context, status *string) ([]*ent.Device, error)
	GetStationByKey(ctx context.Context, deviceKey string) (*ent.Device, error)
	GetStationByID(ctx context.Context, id uuid.UUID) (*ent.Device, error)
	SetStationProducts(ctx context.Context, stationID uuid.UUID, productIDs []uuid.UUID) error
	AddStationProduct(ctx context.Context, stationID uuid.UUID, productID uuid.UUID) error
	RemoveStationProduct(ctx context.Context, stationID uuid.UUID, productID uuid.UUID) error
	ListStationProductIDs(ctx context.Context, stationID uuid.UUID) ([]uuid.UUID, error)
	AssignedItemsForOrder(ctx context.Context, stationID, orderID uuid.UUID) ([]*ent.OrderLine, error)
	RedeemAssigned(ctx context.Context, stationID, orderID uuid.UUID, idemKey string) (map[string]any, error)
	RenameStation(ctx context.Context, stationID uuid.UUID, name string) (*ent.Device, error)
}

type stationService struct {
	cfg            config.Config
	client         *ent.Client
	devices        repository.DeviceRepository
	deviceProducts repository.DeviceProductRepository
	orderLineRepo  repository.OrderLineRepository
	redemptionRepo repository.OrderLineRedemptionRepository
	idempotency    repository.IdempotencyRepository
}

func NewStationService(
	cfg config.Config,
	client *ent.Client,
	devices repository.DeviceRepository,
	deviceProducts repository.DeviceProductRepository,
	orderLineRepo repository.OrderLineRepository,
	redemptionRepo repository.OrderLineRedemptionRepository,
	idempotency repository.IdempotencyRepository,
) StationService {
	return &stationService{
		cfg:            cfg,
		client:         client,
		devices:        devices,
		deviceProducts: deviceProducts,
		orderLineRepo:  orderLineRepo,
		redemptionRepo: redemptionRepo,
		idempotency:    idempotency,
	}
}

func (s *stationService) ListStations(ctx context.Context, status *string) ([]*ent.Device, error) {
	q := s.client.Device.Query().
		Where(device.TypeEQ(device.TypeSTATION)).
		WithDeviceProducts(func(dpq *ent.DeviceProductQuery) {
			dpq.WithProduct()
		})
	if status != nil && *status != "" {
		q = q.Where(device.StatusEQ(device.Status(*status)))
	}
	return q.Order(device.ByCreatedAt()).All(ctx)
}

func (s *stationService) GetStationByKey(ctx context.Context, deviceKey string) (*ent.Device, error) {
	d, err := s.devices.GetByDeviceKey(ctx, deviceKey)
	if err != nil {
		return nil, err
	}
	if d.Type != device.TypeSTATION {
		return nil, errors.New("device_not_station")
	}
	return d, nil
}

func (s *stationService) GetStationByID(ctx context.Context, id uuid.UUID) (*ent.Device, error) {
	d, err := s.getStationWithProducts(ctx, id)
	if err != nil {
		return nil, err
	}
	return d, nil
}

// getStationWithProducts loads a station device by ID with its device products
// (and each product) eagerly loaded via ent edges.
func (s *stationService) getStationWithProducts(ctx context.Context, id uuid.UUID) (*ent.Device, error) {
	return s.client.Device.Query().
		Where(device.ID(id), device.TypeEQ(device.TypeSTATION)).
		WithDeviceProducts(func(q *ent.DeviceProductQuery) {
			q.WithProduct()
		}).
		Only(ctx)
}

func (s *stationService) SetStationProducts(ctx context.Context, stationID uuid.UUID, productIDs []uuid.UUID) error {
	return s.deviceProducts.ReplaceForDevice(ctx, stationID, productIDs)
}

func (s *stationService) AddStationProduct(ctx context.Context, stationID uuid.UUID, productID uuid.UUID) error {
	_, err := s.deviceProducts.Create(ctx, stationID, productID)
	return err
}

func (s *stationService) RemoveStationProduct(ctx context.Context, stationID uuid.UUID, productID uuid.UUID) error {
	return s.deviceProducts.Delete(ctx, stationID, productID)
}

func (s *stationService) ListStationProductIDs(ctx context.Context, stationID uuid.UUID) ([]uuid.UUID, error) {
	return s.devices.ListProductIDsByDevice(ctx, stationID)
}

func (s *stationService) RenameStation(ctx context.Context, stationID uuid.UUID, name string) (*ent.Device, error) {
	_, err := s.client.Device.UpdateOneID(stationID).
		Where(device.TypeEQ(device.TypeSTATION)).
		SetName(name).
		Save(ctx)
	if err != nil {
		return nil, err
	}
	return s.getStationWithProducts(ctx, stationID)
}

func (s *stationService) AssignedItemsForOrder(ctx context.Context, stationID, orderID uuid.UUID) ([]*ent.OrderLine, error) {
	pids, err := s.devices.ListProductIDsByDevice(ctx, stationID)
	if err != nil {
		return nil, err
	}
	if len(pids) == 0 {
		return []*ent.OrderLine{}, nil
	}
	lines, err := s.orderLineRepo.GetByOrderAndProductIDs(ctx, orderID, pids)
	if err != nil {
		return nil, err
	}

	// Bundle lines are not redeemable — expand them to their component children.
	var bundleIDs []uuid.UUID
	var result []*ent.OrderLine
	for _, line := range lines {
		if line.LineType == orderline.LineTypeBundle {
			bundleIDs = append(bundleIDs, line.ID)
		} else {
			result = append(result, line)
		}
	}
	if len(bundleIDs) > 0 {
		children, err := s.orderLineRepo.GetByParentLineIDs(ctx, bundleIDs)
		if err != nil {
			return nil, err
		}
		result = append(result, children...)
	}

	return result, nil
}

func (s *stationService) RedeemAssigned(ctx context.Context, stationID, orderID uuid.UUID, idemKey string) (map[string]any, error) {
	scope := fmt.Sprintf("station:%s:order:%s", stationID.String(), orderID.String())

	// Check idempotency
	if idemKey != "" {
		if rec, err := s.idempotency.Get(ctx, scope, idemKey); err == nil && rec != nil {
			response, _ := repository.GetResponseMap(rec)
			return response, nil
		}
	}

	// Get assigned items
	assigned, err := s.AssignedItemsForOrder(ctx, stationID, orderID)
	if err != nil {
		return nil, err
	}

	// Filter unredeemed items, skipping bundle parents (only components are redeemable)
	var unredeemedIDs []uuid.UUID
	for _, line := range assigned {
		if line.Edges.Redemption == nil && line.LineType != orderline.LineTypeBundle {
			unredeemedIDs = append(unredeemedIDs, line.ID)
		}
	}

	// Redeem unredeemed items
	var redeemed int64
	if len(unredeemedIDs) > 0 {
		redeemed, err = s.redemptionRepo.RedeemUnredeemedByOrderLineIDs(ctx, unredeemedIDs)
		if err != nil {
			return nil, err
		}
	}

	now := time.Now().UTC()

	// Build response
	resp := map[string]any{
		"orderId":    orderID.String(),
		"stationId":  stationID.String(),
		"matched":    len(assigned),
		"redeemed":   redeemed,
		"items":      toPublicOrderLines(assigned),
		"redeemedAt": now.Format(time.RFC3339),
	}

	// Save idempotency record
	if idemKey != "" {
		_, _ = s.idempotency.SaveIfAbsent(ctx, scope, idemKey, resp, 24*time.Hour)
	}

	return resp, nil
}

func toPublicOrderLines(lines []*ent.OrderLine) []map[string]any {
	out := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		var parentID *string
		if line.ParentLineID != nil {
			s := line.ParentLineID.String()
			parentID = &s
		}
		var msID *string
		if line.MenuSlotID != nil {
			s := line.MenuSlotID.String()
			msID = &s
		}
		isRedeemed := line.Edges.Redemption != nil
		out = append(out, map[string]any{
			"id":           line.ID.String(),
			"orderId":      line.OrderID.String(),
			"productId":    line.ProductID.String(),
			"title":        line.Title,
			"quantity":     line.Quantity,
			"isRedeemed":   isRedeemed,
			"parentItemId": parentID,
			"menuSlotId":   msID,
			"menuSlotName": line.MenuSlotName,
		})
	}
	return out
}

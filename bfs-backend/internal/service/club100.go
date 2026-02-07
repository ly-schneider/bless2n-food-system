package service

import (
	"context"
	"fmt"

	"backend/internal/repository"

	"github.com/google/uuid"
)

var ErrProductNotFreeForClub100 = fmt.Errorf("product_not_free_for_club100")

type Club100Service interface {
	GetPeopleWithRedemptions(ctx context.Context) ([]Club100Person, error)
	GetRemainingRedemptions(ctx context.Context, elvantoPersonID string) (remaining int, max int, err error)
	RecordRedemption(ctx context.Context, elvantoPersonID, elvantoPersonName string, orderID uuid.UUID, qty int) error
	GetFreeProductIDs(ctx context.Context) ([]uuid.UUID, error)
	GetMaxRedemptions(ctx context.Context) (int, error)
	ValidateOrderForRedemption(ctx context.Context, orderID uuid.UUID) error
}

type Club100Person struct {
	ID               string `json:"id"`
	FirstName        string `json:"firstName"`
	LastName         string `json:"lastName"`
	TotalRedemptions int    `json:"totalRedemptions"`
	Remaining        int    `json:"remaining"`
	Max              int    `json:"max"`
}

type club100Service struct {
	elvanto     ElvantoService
	redemptions repository.Club100RedemptionRepository
	settings    repository.SettingsRepository
	orderLines  repository.OrderLineRepository
}

func NewClub100Service(
	elvanto ElvantoService,
	redemptions repository.Club100RedemptionRepository,
	settings repository.SettingsRepository,
	orderLines repository.OrderLineRepository,
) Club100Service {
	return &club100Service{
		elvanto:     elvanto,
		redemptions: redemptions,
		settings:    settings,
		orderLines:  orderLines,
	}
}

func (s *club100Service) GetPeopleWithRedemptions(ctx context.Context) ([]Club100Person, error) {
	if !s.elvanto.IsConfigured() {
		return []Club100Person{}, nil
	}

	settingsData, err := s.settings.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}
	maxRedemptions := settingsData.Club100MaxRedemptions

	people, err := s.elvanto.SearchPeople(ctx)
	if err != nil {
		return nil, fmt.Errorf("search people: %w", err)
	}

	result := make([]Club100Person, 0, len(people))
	for _, p := range people {
		total, err := s.redemptions.GetTotalRedemptions(ctx, p.ID)
		if err != nil {
			return nil, fmt.Errorf("get redemptions for %s: %w", p.ID, err)
		}
		remaining := maxRedemptions - total
		if remaining < 0 {
			remaining = 0
		}
		result = append(result, Club100Person{
			ID:              p.ID,
			FirstName:       p.FirstName,
			LastName:        p.LastName,
			TotalRedemptions: total,
			Remaining:       remaining,
			Max:             maxRedemptions,
		})
	}

	return result, nil
}

func (s *club100Service) GetRemainingRedemptions(ctx context.Context, elvantoPersonID string) (remaining int, max int, err error) {
	settingsData, err := s.settings.Get(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("get settings: %w", err)
	}
	max = settingsData.Club100MaxRedemptions

	total, err := s.redemptions.GetTotalRedemptions(ctx, elvantoPersonID)
	if err != nil {
		return 0, 0, fmt.Errorf("get total redemptions: %w", err)
	}

	remaining = max - total
	if remaining < 0 {
		remaining = 0
	}
	return remaining, max, nil
}

func (s *club100Service) RecordRedemption(ctx context.Context, elvantoPersonID, elvantoPersonName string, orderID uuid.UUID, qty int) error {
	if qty <= 0 {
		return nil
	}

	remaining, _, err := s.GetRemainingRedemptions(ctx, elvantoPersonID)
	if err != nil {
		return fmt.Errorf("check remaining: %w", err)
	}
	if remaining < qty {
		return fmt.Errorf("insufficient remaining redemptions: have %d, need %d", remaining, qty)
	}

	_, err = s.redemptions.Create(ctx, elvantoPersonID, elvantoPersonName, orderID, qty)
	if err != nil {
		return fmt.Errorf("create redemption: %w", err)
	}
	return nil
}

func (s *club100Service) GetFreeProductIDs(ctx context.Context) ([]uuid.UUID, error) {
	settingsWithProducts, err := s.settings.GetWithProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}
	if settingsWithProducts.Edges.Club100FreeProducts == nil {
		return []uuid.UUID{}, nil
	}
	ids := make([]uuid.UUID, 0, len(settingsWithProducts.Edges.Club100FreeProducts))
	for _, p := range settingsWithProducts.Edges.Club100FreeProducts {
		ids = append(ids, p.ID)
	}
	return ids, nil
}

func (s *club100Service) GetMaxRedemptions(ctx context.Context) (int, error) {
	settingsData, err := s.settings.Get(ctx)
	if err != nil {
		return 0, fmt.Errorf("get settings: %w", err)
	}
	return settingsData.Club100MaxRedemptions, nil
}

func (s *club100Service) ValidateOrderForRedemption(ctx context.Context, orderID uuid.UUID) error {
	freeProductIDs, err := s.GetFreeProductIDs(ctx)
	if err != nil {
		return fmt.Errorf("get free product IDs: %w", err)
	}

	if len(freeProductIDs) == 0 {
		return fmt.Errorf("%w: no free products configured", ErrProductNotFreeForClub100)
	}

	freeProductSet := make(map[uuid.UUID]bool, len(freeProductIDs))
	for _, id := range freeProductIDs {
		freeProductSet[id] = true
	}

	lines, err := s.orderLines.GetByOrderID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("get order lines: %w", err)
	}

	for _, line := range lines {
		if !freeProductSet[line.ProductID] {
			return fmt.Errorf("%w: product %s is not in the free products list", ErrProductNotFreeForClub100, line.ProductID)
		}
	}

	return nil
}

package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"backend/internal/generated/ent"
	"backend/internal/generated/ent/settings"
	"backend/internal/repository"

	"github.com/google/uuid"
)

var (
	ErrJetonRequired = errors.New("jeton_required")
)

type MissingJetonForActiveProductsError struct {
	Count int64
}

func (e MissingJetonForActiveProductsError) Error() string {
	return "active_products_missing_jeton"
}

type JetonInUseError struct {
	Count int64
}

func (e JetonInUseError) Error() string {
	return "jeton_in_use"
}

type SettingsService interface {
	GetSettings(ctx context.Context) (*ent.Settings, error)
	GetSettingsWithProducts(ctx context.Context) (*ent.Settings, error)
	SetPosMode(ctx context.Context, mode settings.PosMode) error
	SetClub100Settings(ctx context.Context, freeProductIDs []uuid.UUID, maxRedemptions *int) error
	ListJetons(ctx context.Context) ([]*ent.Jeton, error)
	CreateJeton(ctx context.Context, name, color string) (*ent.Jeton, error)
	UpdateJeton(ctx context.Context, id uuid.UUID, name, color string) (*ent.Jeton, error)
	DeleteJeton(ctx context.Context, id uuid.UUID) error
	SetProductJeton(ctx context.Context, productID uuid.UUID, jetonID *uuid.UUID) error
	GetFreeProductIDs(ctx context.Context) ([]uuid.UUID, error)
}

type settingsService struct {
	settings repository.SettingsRepository
	jetons   repository.JetonRepository
	products *repository.ProductRepository
}

func NewSettingsService(
	settings repository.SettingsRepository,
	jetons repository.JetonRepository,
	products *repository.ProductRepository,
) SettingsService {
	return &settingsService{settings: settings, jetons: jetons, products: products}
}

var settingsHexPattern = regexp.MustCompile(`^#?[0-9a-fA-F]{6}$`)

func normalizeSettingsColor(raw string) (string, error) {
	h := strings.TrimSpace(raw)
	if h == "" {
		return "", fmt.Errorf("color_required")
	}
	if !settingsHexPattern.MatchString(h) {
		return "", fmt.Errorf("invalid_hex")
	}
	h = strings.ToUpper(h)
	if !strings.HasPrefix(h, "#") {
		h = "#" + h
	}
	return h, nil
}

func (s *settingsService) GetSettings(ctx context.Context) (*ent.Settings, error) {
	return s.settings.Get(ctx)
}

func (s *settingsService) GetSettingsWithProducts(ctx context.Context) (*ent.Settings, error) {
	return s.settings.GetWithProducts(ctx)
}

func (s *settingsService) SetPosMode(ctx context.Context, mode settings.PosMode) error {
	if mode != settings.PosModeQR_CODE && mode != settings.PosModeJETON {
		return fmt.Errorf("invalid_mode")
	}
	if mode == settings.PosModeJETON {
		if missing, err := s.products.CountActiveWithoutJeton(ctx); err == nil && missing > 0 {
			return MissingJetonForActiveProductsError{Count: missing}
		} else if err != nil {
			return err
		}
	}
	return s.settings.Upsert(ctx, mode)
}

func (s *settingsService) SetClub100Settings(ctx context.Context, freeProductIDs []uuid.UUID, maxRedemptions *int) error {
	current, err := s.settings.Get(ctx)
	if err != nil {
		return err
	}

	max := current.Club100MaxRedemptions
	if maxRedemptions != nil {
		if *maxRedemptions < 0 {
			return fmt.Errorf("max redemptions must be non-negative")
		}
		max = *maxRedemptions
	}

	productIDs := freeProductIDs
	if productIDs == nil {
		settingsWithProducts, err := s.settings.GetWithProducts(ctx)
		if err != nil {
			return err
		}
		if settingsWithProducts.Edges.Club100FreeProducts != nil {
			productIDs = make([]uuid.UUID, 0, len(settingsWithProducts.Edges.Club100FreeProducts))
			for _, p := range settingsWithProducts.Edges.Club100FreeProducts {
				productIDs = append(productIDs, p.ID)
			}
		}
	}

	for _, productID := range productIDs {
		if productID == uuid.Nil {
			continue
		}
		if _, err := s.products.GetByID(ctx, productID); err != nil {
			return fmt.Errorf("product %s not found", productID)
		}
	}

	validProductIDs := make([]uuid.UUID, 0, len(productIDs))
	for _, id := range productIDs {
		if id != uuid.Nil {
			validProductIDs = append(validProductIDs, id)
		}
	}

	return s.settings.UpdateClub100Settings(ctx, validProductIDs, max)
}

func (s *settingsService) ListJetons(ctx context.Context) ([]*ent.Jeton, error) {
	return s.jetons.GetAll(ctx)
}

func (s *settingsService) CreateJeton(ctx context.Context, name, color string) (*ent.Jeton, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("name_required")
	}
	normColor, err := normalizeSettingsColor(color)
	if err != nil {
		return nil, err
	}
	return s.jetons.Create(ctx, name, normColor)
}

func (s *settingsService) UpdateJeton(ctx context.Context, id uuid.UUID, name, color string) (*ent.Jeton, error) {
	if _, err := s.jetons.GetByID(ctx, id); err != nil {
		return nil, err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("name_required")
	}
	normColor, err := normalizeSettingsColor(color)
	if err != nil {
		return nil, err
	}
	return s.jetons.Update(ctx, id, name, normColor)
}

func (s *settingsService) DeleteJeton(ctx context.Context, id uuid.UUID) error {
	usage, err := s.products.CountByJetonIDs(ctx, []uuid.UUID{id})
	if err == nil {
		if c := usage[id]; c > 0 {
			return JetonInUseError{Count: c}
		}
	}
	return s.jetons.Delete(ctx, id)
}

func (s *settingsService) SetProductJeton(ctx context.Context, productID uuid.UUID, jetonID *uuid.UUID) error {
	p, err := s.products.GetByID(ctx, productID)
	if err != nil {
		return err
	}
	if jetonID != nil {
		if _, err := s.jetons.GetByID(ctx, *jetonID); err != nil {
			return err
		}
	}
	settingsData, err := s.settings.Get(ctx)
	if err != nil {
		return err
	}
	if settingsData != nil && settingsData.PosMode == settings.PosModeJETON && p.IsActive && jetonID == nil {
		if p.Type == "simple" {
			return ErrJetonRequired
		}
	}
	return s.products.UpdateJeton(ctx, productID, jetonID)
}

func (s *settingsService) GetFreeProductIDs(ctx context.Context) ([]uuid.UUID, error) {
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

package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"backend/internal/domain"
	"backend/internal/repository"

	"go.mongodb.org/mongo-driver/v2/bson"
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

type POSConfigService interface {
	GetSettings(ctx context.Context) (*domain.PosSettings, error)
	SetMode(ctx context.Context, mode domain.PosFulfillmentMode) error
	ListJetons(ctx context.Context) ([]domain.JetonDTO, error)
	CreateJeton(ctx context.Context, name, paletteColor string, hexColor *string) (*domain.JetonDTO, error)
	UpdateJeton(ctx context.Context, id bson.ObjectID, name, paletteColor string, hexColor *string) (*domain.JetonDTO, error)
	DeleteJeton(ctx context.Context, id bson.ObjectID) error
	SetProductJeton(ctx context.Context, productID bson.ObjectID, jetonID *bson.ObjectID) error
}

type posConfigService struct {
	settings repository.PosSettingsRepository
	jetons   repository.JetonRepository
	products repository.ProductRepository
}

func NewPOSConfigService(
	settings repository.PosSettingsRepository,
	jetons repository.JetonRepository,
	products repository.ProductRepository,
) POSConfigService {
	return &posConfigService{settings: settings, jetons: jetons, products: products}
}

var jetonPaletteDefaults = map[string]string{
	"yellow": "#FACC15",
	"blue":   "#3B82F6",
	"red":    "#EF4444",
	"green":  "#22C55E",
	"purple": "#A855F7",
	"orange": "#F97316",
	"gray":   "#6B7280",
}

var hexPattern = regexp.MustCompile(`^#?[0-9a-fA-F]{6}$`)

func resolveJetonColorHex(palette string, hexOverride *string) string {
	if hexOverride != nil && hexPattern.MatchString(*hexOverride) {
		h := strings.ToUpper(*hexOverride)
		if !strings.HasPrefix(h, "#") {
			h = "#" + h
		}
		return h
	}
	if v, ok := jetonPaletteDefaults[strings.ToLower(palette)]; ok && v != "" {
		return strings.ToUpper(v)
	}
	return "#9CA3AF"
}

func normalizeHex(hex *string) (*string, error) {
	if hex == nil {
		return nil, nil
	}
	h := strings.TrimSpace(*hex)
	if h == "" {
		return nil, nil
	}
	if !hexPattern.MatchString(h) {
		return nil, fmt.Errorf("invalid_hex")
	}
	h = strings.ToUpper(h)
	if !strings.HasPrefix(h, "#") {
		h = "#" + h
	}
	return &h, nil
}

func (s *posConfigService) GetSettings(ctx context.Context) (*domain.PosSettings, error) {
	return s.settings.Get(ctx)
}

func (s *posConfigService) SetMode(ctx context.Context, mode domain.PosFulfillmentMode) error {
	if mode != domain.PosModeQRCode && mode != domain.PosModeJeton {
		return fmt.Errorf("invalid_mode")
	}
	if mode == domain.PosModeJeton {
		if missing, err := s.products.CountActiveWithoutJeton(ctx); err == nil && missing > 0 {
			return MissingJetonForActiveProductsError{Count: missing}
		} else if err != nil {
			return err
		}
	}
	return s.settings.UpsertMode(ctx, mode)
}

func (s *posConfigService) ListJetons(ctx context.Context) ([]domain.JetonDTO, error) {
	items, err := s.jetons.List(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]bson.ObjectID, 0, len(items))
	for _, j := range items {
		ids = append(ids, j.ID)
	}
	usage, _ := s.products.CountByJetonIDs(ctx, ids)
	out := make([]domain.JetonDTO, 0, len(items))
	for _, j := range items {
		dto := domain.JetonDTO{
			ID:           j.ID.Hex(),
			Name:         j.Name,
			PaletteColor: j.PaletteColor,
			HexColor:     j.HexColor,
			ColorHex:     resolveJetonColorHex(j.PaletteColor, j.HexColor),
		}
		c := usage[j.ID]
		dto.UsageCount = &c
		out = append(out, dto)
	}
	return out, nil
}

func (s *posConfigService) CreateJeton(ctx context.Context, name, paletteColor string, hexColor *string) (*domain.JetonDTO, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("name_required")
	}
	if paletteColor == "" && (hexColor == nil || strings.TrimSpace(*hexColor) == "") {
		return nil, fmt.Errorf("color_required")
	}
	normHex, err := normalizeHex(hexColor)
	if err != nil {
		return nil, err
	}
	j := &domain.Jeton{
		Name:         name,
		PaletteColor: paletteColor,
		HexColor:     normHex,
	}
	if _, err := s.jetons.Insert(ctx, j); err != nil {
		return nil, err
	}
	dto := domain.JetonDTO{
		ID:           j.ID.Hex(),
		Name:         j.Name,
		PaletteColor: j.PaletteColor,
		HexColor:     j.HexColor,
		ColorHex:     resolveJetonColorHex(j.PaletteColor, j.HexColor),
	}
	zero := int64(0)
	dto.UsageCount = &zero
	return &dto, nil
}

func (s *posConfigService) UpdateJeton(ctx context.Context, id bson.ObjectID, name, paletteColor string, hexColor *string) (*domain.JetonDTO, error) {
	existing, err := s.jetons.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, mongoNotFoundErr("jeton_not_found")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("name_required")
	}
	if paletteColor == "" && (hexColor == nil || strings.TrimSpace(*hexColor) == "") {
		return nil, fmt.Errorf("color_required")
	}
	normHex, err := normalizeHex(hexColor)
	if err != nil {
		return nil, err
	}
	set := bson.M{"name": name, "palette_color": paletteColor, "hex_color": normHex}
	if err := s.jetons.Update(ctx, id, set); err != nil {
		return nil, err
	}
	existing.Name = name
	existing.PaletteColor = paletteColor
	existing.HexColor = normHex

	dto := domain.JetonDTO{
		ID:           existing.ID.Hex(),
		Name:         existing.Name,
		PaletteColor: existing.PaletteColor,
		HexColor:     existing.HexColor,
		ColorHex:     resolveJetonColorHex(existing.PaletteColor, existing.HexColor),
	}
	if usage, err := s.products.CountByJetonIDs(ctx, []bson.ObjectID{id}); err == nil {
		if c, ok := usage[id]; ok {
			dto.UsageCount = &c
		}
	}
	return &dto, nil
}

func (s *posConfigService) DeleteJeton(ctx context.Context, id bson.ObjectID) error {
	usage, err := s.products.CountByJetonIDs(ctx, []bson.ObjectID{id})
	if err == nil {
		if c := usage[id]; c > 0 {
			return JetonInUseError{Count: c}
		}
	}
	return s.jetons.Delete(ctx, id)
}

func (s *posConfigService) SetProductJeton(ctx context.Context, productID bson.ObjectID, jetonID *bson.ObjectID) error {
	p, err := s.products.FindByID(ctx, productID)
	if err != nil {
		return err
	}
	if p == nil {
		return mongoNotFoundErr("product_not_found")
	}
	// Validate jeton existence when provided
	if jetonID != nil {
		if j, err := s.jetons.FindByID(ctx, *jetonID); err != nil {
			return err
		} else if j == nil {
			return mongoNotFoundErr("jeton_not_found")
		}
	}
	settings, err := s.settings.Get(ctx)
	if err != nil {
		return err
	}
	if settings != nil && settings.Mode == domain.PosModeJeton && p.IsActive && jetonID == nil {
		if p.Type == domain.ProductTypeSimple {
			return ErrJetonRequired
		}
	}
	return s.products.UpdateJeton(ctx, productID, jetonID)
}

func mongoNotFoundErr(msg string) error {
	return errors.New(msg)
}

package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"backend/internal/generated/ent"
	"backend/internal/generated/ent/order"
	"backend/internal/generated/ent/orderline"
	"backend/internal/generated/ent/orderpayment"
	"backend/internal/generated/ent/volunteercampaign"
	"backend/internal/repository"

	"github.com/google/uuid"
)

const (
	volunteerSessionBytes   = 24
	volunteerAccessCodeLen  = 4
	volunteerAccessCodeAlph = "ABCDEFGHIJKLMNOPQRSTUVWXYZ123456789"
	volunteerQRPayloadPfx   = "CAMP:"
)

var (
	ErrVolunteerCampaignNotFound      = errors.New("volunteer_campaign_not_found")
	ErrVolunteerCampaignInactive      = errors.New("volunteer_campaign_inactive")
	ErrVolunteerCampaignOutsideValid  = errors.New("volunteer_campaign_outside_validity")
	ErrVolunteerAccessCodeInvalid     = errors.New("volunteer_access_code_invalid")
	ErrVolunteerCampaignHasNoProducts = errors.New("volunteer_campaign_has_no_products")
	ErrVolunteerCampaignInvalidAccess = errors.New("volunteer_access_code_must_be_4_digits")
	ErrVolunteerMaxRedemptionsReached = errors.New("volunteer_max_redemptions_reached")
	ErrVolunteerMaxBelowCount         = errors.New("volunteer_max_redemptions_below_current_count")
)

type VolunteerService interface {
	CreateCampaign(ctx context.Context, input CreateVolunteerCampaignInput) (*ent.VolunteerCampaign, error)
	ListCampaigns(ctx context.Context) ([]VolunteerCampaignSummary, error)
	GetCampaign(ctx context.Context, id uuid.UUID) (*VolunteerCampaignDetail, error)
	UpdateCampaign(ctx context.Context, id uuid.UUID, input UpdateVolunteerCampaignInput) (*ent.VolunteerCampaign, error)
	EndCampaign(ctx context.Context, id uuid.UUID) error
	RotateClaimToken(ctx context.Context, id uuid.UUID) (uuid.UUID, error)

	VerifyAccess(ctx context.Context, token uuid.UUID, code string) (sessionID string, err error)
	NewSessionID() string

	GetClaimView(ctx context.Context, token uuid.UUID) (*VolunteerClaimView, error)
	RedeemSharedQR(ctx context.Context, claimToken uuid.UUID, stationID uuid.UUID, idempotencyKey string) (*VolunteerRedemptionResult, error)
}

type CreateVolunteerCampaignInput struct {
	Name           string
	ValidFrom      *time.Time
	ValidUntil     *time.Time
	Products       []repository.VolunteerCampaignProductInput
	MaxRedemptions int
}

type UpdateVolunteerCampaignInput struct {
	Name           string
	ValidFrom      *time.Time
	ValidUntil     *time.Time
	Status         volunteercampaign.Status
	Products       []repository.VolunteerCampaignProductInput
	MaxRedemptions *int
}

type VolunteerCampaignSummary struct {
	Campaign        *ent.VolunteerCampaign
	MaxRedemptions  int
	RedemptionCount int
}

type VolunteerCampaignProductView struct {
	ProductID    uuid.UUID
	ProductName  string
	ProductImage *string
	Quantity     int
}

type VolunteerRedemptionView struct {
	ID        uuid.UUID
	OrderID   uuid.UUID
	CreatedAt time.Time
}

type VolunteerCampaignDetail struct {
	Campaign    *ent.VolunteerCampaign
	Products    []VolunteerCampaignProductView
	Redemptions []VolunteerRedemptionView
}

type VolunteerClaimView struct {
	Campaign  *ent.VolunteerCampaign
	Products  []VolunteerCampaignProductView
	QRPayload string
}

type VolunteerRedemptionResult struct {
	OrderID         uuid.UUID
	RedemptionCount int
	MaxRedemptions  int
	StationResult   map[string]any
}

type volunteerService struct {
	client      *ent.Client
	campaigns   repository.VolunteerCampaignRepository
	redemptions repository.VolunteerRedemptionRepository
	orders      repository.OrderRepository
	lines       repository.OrderLineRepository
	payments    repository.OrderPaymentRepository
	stations    StationService
}

func NewVolunteerService(
	client *ent.Client,
	campaigns repository.VolunteerCampaignRepository,
	redemptions repository.VolunteerRedemptionRepository,
	orders repository.OrderRepository,
	lines repository.OrderLineRepository,
	payments repository.OrderPaymentRepository,
	stations StationService,
) VolunteerService {
	return &volunteerService{
		client:      client,
		campaigns:   campaigns,
		redemptions: redemptions,
		orders:      orders,
		lines:       lines,
		payments:    payments,
		stations:    stations,
	}
}

func (s *volunteerService) NewSessionID() string {
	b := make([]byte, volunteerSessionBytes)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func generateAccessCode() string {
	buf := make([]byte, volunteerAccessCodeLen)
	raw := make([]byte, volunteerAccessCodeLen)
	_, _ = rand.Read(raw)
	n := byte(len(volunteerAccessCodeAlph))
	for i := 0; i < volunteerAccessCodeLen; i++ {
		buf[i] = volunteerAccessCodeAlph[raw[i]%n]
	}
	return string(buf)
}

// BuildQRPayload returns the string encoded in the shared campaign QR code.
// Station scanner detects the CAMP: prefix and routes to the campaign-redeem endpoint.
func BuildQRPayload(claimToken uuid.UUID) string {
	return volunteerQRPayloadPfx + claimToken.String()
}

func (s *volunteerService) CreateCampaign(ctx context.Context, input CreateVolunteerCampaignInput) (*ent.VolunteerCampaign, error) {
	if strings.TrimSpace(input.Name) == "" {
		return nil, errors.New("name_required")
	}
	if len(input.Products) == 0 {
		return nil, ErrVolunteerCampaignHasNoProducts
	}
	if input.MaxRedemptions <= 0 {
		return nil, errors.New("max_redemptions_must_be_positive")
	}

	accessCode := generateAccessCode()

	tx, err := s.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	txCtx := repository.ContextWithClient(ctx, tx.Client())

	campaign, err := s.campaigns.Create(txCtx, input.Name, accessCode, input.ValidFrom, input.ValidUntil, volunteercampaign.StatusActive, input.MaxRedemptions)
	if err != nil {
		return nil, fmt.Errorf("create campaign: %w", err)
	}

	if err := s.campaigns.ReplaceProducts(txCtx, campaign.ID, input.Products); err != nil {
		return nil, fmt.Errorf("attach products: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return campaign, nil
}

type productSnapshot struct {
	ID       uuid.UUID
	Name     string
	Quantity int
}

func (s *volunteerService) createGratisOrder(ctx context.Context, products []productSnapshot) (uuid.UUID, error) {
	ord, err := s.orders.Create(ctx, 0, order.StatusPaid, order.OriginShop, nil, nil, nil, nil, nil)
	if err != nil {
		return uuid.Nil, fmt.Errorf("create order: %w", err)
	}

	for _, p := range products {
		if _, err := s.lines.Create(ctx, ord.ID, orderline.LineTypeSimple, p.ID, truncatedTitle(p.Name), p.Quantity, 0, nil, nil, nil); err != nil {
			return uuid.Nil, fmt.Errorf("create order line: %w", err)
		}
	}

	if _, err := s.payments.Create(ctx, ord.ID, orderpayment.MethodGRATIS_STAFF, 0, time.Now(), nil); err != nil {
		return uuid.Nil, fmt.Errorf("create payment: %w", err)
	}
	return ord.ID, nil
}

func truncatedTitle(s string) string {
	if len(s) > 20 {
		return s[:20]
	}
	return s
}

func (s *volunteerService) ListCampaigns(ctx context.Context) ([]VolunteerCampaignSummary, error) {
	campaigns, err := s.campaigns.List(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]VolunteerCampaignSummary, 0, len(campaigns))
	for _, c := range campaigns {
		out = append(out, VolunteerCampaignSummary{
			Campaign:        c,
			MaxRedemptions:  c.MaxRedemptions,
			RedemptionCount: c.RedemptionCount,
		})
	}
	return out, nil
}

func (s *volunteerService) GetCampaign(ctx context.Context, id uuid.UUID) (*VolunteerCampaignDetail, error) {
	campaign, err := s.campaigns.GetByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrVolunteerCampaignNotFound
		}
		return nil, err
	}
	cps, err := s.campaigns.ListProducts(ctx, id)
	if err != nil {
		return nil, err
	}
	products := campaignProductsToViews(cps)

	reds, err := s.redemptions.ListByCampaign(ctx, id, 50)
	if err != nil {
		return nil, err
	}
	redViews := make([]VolunteerRedemptionView, 0, len(reds))
	for _, r := range reds {
		redViews = append(redViews, VolunteerRedemptionView{
			ID:        r.ID,
			OrderID:   r.OrderID,
			CreatedAt: r.CreatedAt,
		})
	}
	return &VolunteerCampaignDetail{
		Campaign:    campaign,
		Products:    products,
		Redemptions: redViews,
	}, nil
}

func (s *volunteerService) UpdateCampaign(ctx context.Context, id uuid.UUID, input UpdateVolunteerCampaignInput) (*ent.VolunteerCampaign, error) {
	if strings.TrimSpace(input.Name) == "" {
		return nil, errors.New("name_required")
	}

	tx, err := s.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := repository.ContextWithClient(ctx, tx.Client())

	existing, err := s.campaigns.GetByID(txCtx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrVolunteerCampaignNotFound
		}
		return nil, err
	}

	campaign, err := s.campaigns.Update(txCtx, id, input.Name, existing.AccessCode, input.ValidFrom, input.ValidUntil, input.Status)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrVolunteerCampaignNotFound
		}
		return nil, err
	}

	if input.MaxRedemptions != nil {
		if *input.MaxRedemptions <= 0 {
			return nil, errors.New("max_redemptions_must_be_positive")
		}
		updated, ok, err := s.campaigns.UpdateMaxRedemptions(txCtx, id, *input.MaxRedemptions)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrVolunteerMaxBelowCount
		}
		campaign = updated
	}

	if len(input.Products) > 0 {
		if err := s.campaigns.ReplaceProducts(txCtx, id, input.Products); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return campaign, nil
}

func (s *volunteerService) EndCampaign(ctx context.Context, id uuid.UUID) error {
	if err := s.campaigns.SetStatus(ctx, id, volunteercampaign.StatusEnded); err != nil {
		if ent.IsNotFound(err) {
			return ErrVolunteerCampaignNotFound
		}
		return err
	}
	return nil
}

func (s *volunteerService) RotateClaimToken(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	return s.campaigns.RotateClaimToken(ctx, id)
}

func (s *volunteerService) VerifyAccess(ctx context.Context, token uuid.UUID, code string) (string, error) {
	campaign, err := s.campaigns.GetByClaimToken(ctx, token)
	if err != nil {
		if ent.IsNotFound(err) {
			return "", ErrVolunteerCampaignNotFound
		}
		return "", err
	}
	if campaign.Status != volunteercampaign.StatusActive {
		return "", ErrVolunteerCampaignInactive
	}
	if !campaignWithinValidity(campaign, time.Now()) {
		return "", ErrVolunteerCampaignOutsideValid
	}
	if code != campaign.AccessCode {
		return "", ErrVolunteerAccessCodeInvalid
	}
	return s.NewSessionID(), nil
}

func (s *volunteerService) GetClaimView(ctx context.Context, token uuid.UUID) (*VolunteerClaimView, error) {
	campaign, err := s.campaigns.GetByClaimToken(ctx, token)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrVolunteerCampaignNotFound
		}
		return nil, err
	}
	cps, err := s.campaigns.ListProducts(ctx, campaign.ID)
	if err != nil {
		return nil, err
	}
	return &VolunteerClaimView{
		Campaign:  campaign,
		Products:  campaignProductsToViews(cps),
		QRPayload: BuildQRPayload(campaign.ClaimToken),
	}, nil
}

func (s *volunteerService) RedeemSharedQR(ctx context.Context, claimToken uuid.UUID, stationID uuid.UUID, idempotencyKey string) (*VolunteerRedemptionResult, error) {
	campaign, err := s.campaigns.GetByClaimToken(ctx, claimToken)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrVolunteerCampaignNotFound
		}
		return nil, err
	}

	if idempotencyKey != "" {
		if prev, err := s.redemptions.GetByIdempotencyKey(ctx, campaign.ID, idempotencyKey); err == nil && prev != nil {
			stationResp, redeemErr := s.stations.RedeemAssigned(ctx, stationID, prev.OrderID, idempotencyKey)
			if redeemErr != nil {
				return nil, redeemErr
			}
			return &VolunteerRedemptionResult{
				OrderID:         prev.OrderID,
				RedemptionCount: campaign.RedemptionCount,
				MaxRedemptions:  campaign.MaxRedemptions,
				StationResult:   stationResp,
			}, nil
		}
	}

	if campaign.Status != volunteercampaign.StatusActive {
		return nil, ErrVolunteerCampaignInactive
	}
	if !campaignWithinValidity(campaign, time.Now()) {
		return nil, ErrVolunteerCampaignOutsideValid
	}

	cps, err := s.campaigns.ListProducts(ctx, campaign.ID)
	if err != nil {
		return nil, err
	}
	if len(cps) == 0 {
		return nil, ErrVolunteerCampaignHasNoProducts
	}
	products := make([]productSnapshot, 0, len(cps))
	for _, cp := range cps {
		p, _ := cp.Edges.ProductOrErr()
		name := ""
		if p != nil {
			name = p.Name
		}
		qty := cp.Quantity
		if qty <= 0 {
			qty = 1
		}
		products = append(products, productSnapshot{
			ID:       cp.ProductID,
			Name:     name,
			Quantity: qty,
		})
	}

	incremented, err := s.campaigns.IncrementRedemptionAtomic(ctx, campaign.ID)
	if err != nil {
		return nil, err
	}
	if !incremented {
		return nil, ErrVolunteerMaxRedemptionsReached
	}

	tx, err := s.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := repository.ContextWithClient(ctx, tx.Client())

	orderID, err := s.createGratisOrder(txCtx, products)
	if err != nil {
		return nil, err
	}

	var idemPtr *string
	if idempotencyKey != "" {
		idemPtr = &idempotencyKey
	}
	var stationPtr *uuid.UUID
	if stationID != uuid.Nil {
		stationPtr = &stationID
	}
	if _, err := s.redemptions.Create(txCtx, campaign.ID, orderID, stationPtr, idemPtr); err != nil {
		return nil, fmt.Errorf("record redemption: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	stationResp, err := s.stations.RedeemAssigned(ctx, stationID, orderID, idempotencyKey)
	if err != nil {
		return nil, err
	}

	return &VolunteerRedemptionResult{
		OrderID:         orderID,
		RedemptionCount: campaign.RedemptionCount + 1,
		MaxRedemptions:  campaign.MaxRedemptions,
		StationResult:   stationResp,
	}, nil
}

func campaignWithinValidity(c *ent.VolunteerCampaign, now time.Time) bool {
	if c.ValidFrom != nil && now.Before(*c.ValidFrom) {
		return false
	}
	if c.ValidUntil != nil && now.After(*c.ValidUntil) {
		return false
	}
	return true
}

func campaignProductsToViews(cps []*ent.VolunteerCampaignProduct) []VolunteerCampaignProductView {
	out := make([]VolunteerCampaignProductView, 0, len(cps))
	for _, cp := range cps {
		name := ""
		var image *string
		if p, _ := cp.Edges.ProductOrErr(); p != nil {
			name = p.Name
			image = p.Image
		}
		out = append(out, VolunteerCampaignProductView{
			ProductID:    cp.ProductID,
			ProductName:  name,
			ProductImage: image,
			Quantity:     cp.Quantity,
		})
	}
	return out
}

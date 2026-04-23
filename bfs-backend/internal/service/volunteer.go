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
	volunteerReservationTTL = 10 * time.Minute
	volunteerSessionBytes   = 24
	volunteerAccessCodeLen  = 4
	volunteerAccessCodeAlph = "ABCDEFGHIJKLMNOPQRSTUVWXYZ123456789"
)

var (
	ErrVolunteerCampaignNotFound      = errors.New("volunteer_campaign_not_found")
	ErrVolunteerCampaignInactive      = errors.New("volunteer_campaign_inactive")
	ErrVolunteerCampaignOutsideValid  = errors.New("volunteer_campaign_outside_validity")
	ErrVolunteerAccessCodeInvalid     = errors.New("volunteer_access_code_invalid")
	ErrVolunteerSlotNotAvailable      = errors.New("volunteer_slot_not_available")
	ErrVolunteerSlotNotYours          = errors.New("volunteer_slot_not_yours")
	ErrVolunteerSlotAlreadyRedeemed   = errors.New("volunteer_slot_already_redeemed")
	ErrVolunteerCampaignHasNoProducts = errors.New("volunteer_campaign_has_no_products")
	ErrVolunteerCampaignInvalidAccess = errors.New("volunteer_access_code_must_be_4_digits")
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

	GetClaimCampaign(ctx context.Context, token uuid.UUID) (*ent.VolunteerCampaign, error)
	ListClaimSlots(ctx context.Context, token uuid.UUID, sessionID string) (*VolunteerClaimView, error)
	GetClaimSlot(ctx context.Context, token uuid.UUID, slotID uuid.UUID, sessionID string) (*VolunteerSlotView, error)
	ReserveSlot(ctx context.Context, token uuid.UUID, slotID uuid.UUID, sessionID string) error
	ReleaseSlot(ctx context.Context, token uuid.UUID, slotID uuid.UUID, sessionID string) error
}

type CreateVolunteerCampaignInput struct {
	Name       string
	AccessCode string
	ValidFrom  *time.Time
	ValidUntil *time.Time
	Products   []repository.VolunteerCampaignProductInput
	SlotCount  int
}

type UpdateVolunteerCampaignInput struct {
	Name       string
	AccessCode string
	ValidFrom  *time.Time
	ValidUntil *time.Time
	Status     volunteercampaign.Status
	Products   []repository.VolunteerCampaignProductInput
}

type VolunteerCampaignSummary struct {
	Campaign      *ent.VolunteerCampaign
	TotalSlots    int
	RedeemedSlots int
	ReservedSlots int
}

type VolunteerCampaignDetail struct {
	Campaign *ent.VolunteerCampaign
	Products []*ent.VolunteerCampaignProduct
	Slots    []*ent.VolunteerSlot
}

type VolunteerClaimView struct {
	Campaign       *ent.VolunteerCampaign
	AvailableSlots []VolunteerSlotView
	ReservedByMe   []VolunteerSlotView
	TotalSlots     int
}

type VolunteerSlotView struct {
	Slot          *ent.VolunteerSlot
	OrderID       uuid.UUID
	IsReservedBy  string
	ReservedUntil *time.Time
	RedeemedAt    *time.Time
	Lines         []VolunteerSlotLine
}

type VolunteerSlotLine struct {
	ProductName  string
	ProductImage *string
	Quantity     int
	IsRedeemed   bool
	RedeemedAt   *time.Time
}

type volunteerService struct {
	client    *ent.Client
	campaigns repository.VolunteerCampaignRepository
	slots     repository.VolunteerSlotRepository
	orders    repository.OrderRepository
	lines     repository.OrderLineRepository
	payments  repository.OrderPaymentRepository
	products  *repository.ProductRepository
}

func NewVolunteerService(
	client *ent.Client,
	campaigns repository.VolunteerCampaignRepository,
	slots repository.VolunteerSlotRepository,
	orders repository.OrderRepository,
	lines repository.OrderLineRepository,
	payments repository.OrderPaymentRepository,
	products *repository.ProductRepository,
) VolunteerService {
	return &volunteerService{
		client:    client,
		campaigns: campaigns,
		slots:     slots,
		orders:    orders,
		lines:     lines,
		payments:  payments,
		products:  products,
	}
}

func (s *volunteerService) NewSessionID() string {
	b := make([]byte, volunteerSessionBytes)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func validateAccessCode(code string) error {
	if len(code) != volunteerAccessCodeLen {
		return ErrVolunteerCampaignInvalidAccess
	}
	for _, c := range code {
		if !strings.ContainsRune(volunteerAccessCodeAlph, c) {
			return ErrVolunteerCampaignInvalidAccess
		}
	}
	return nil
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

func (s *volunteerService) CreateCampaign(ctx context.Context, input CreateVolunteerCampaignInput) (*ent.VolunteerCampaign, error) {
	if strings.TrimSpace(input.Name) == "" {
		return nil, errors.New("name_required")
	}
	if len(input.Products) == 0 {
		return nil, ErrVolunteerCampaignHasNoProducts
	}
	if input.SlotCount <= 0 {
		return nil, errors.New("slot_count_must_be_positive")
	}

	accessCode := generateAccessCode()

	productSnapshots, err := s.loadProductSnapshots(ctx, input.Products)
	if err != nil {
		return nil, err
	}

	tx, err := s.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	txCtx := repository.ContextWithClient(ctx, tx.Client())

	campaign, err := s.campaigns.Create(txCtx, input.Name, accessCode, input.ValidFrom, input.ValidUntil, volunteercampaign.StatusActive)
	if err != nil {
		return nil, fmt.Errorf("create campaign: %w", err)
	}

	if err := s.campaigns.ReplaceProducts(txCtx, campaign.ID, input.Products); err != nil {
		return nil, fmt.Errorf("attach products: %w", err)
	}

	for i := 0; i < input.SlotCount; i++ {
		if err := s.createSlotOrder(txCtx, campaign.ID, productSnapshots); err != nil {
			return nil, fmt.Errorf("create slot %d: %w", i, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return campaign, nil
}

type productSnapshot struct {
	ID         uuid.UUID
	Name       string
	PriceCents int64
	Quantity   int
}

func (s *volunteerService) loadProductSnapshots(ctx context.Context, items []repository.VolunteerCampaignProductInput) ([]productSnapshot, error) {
	out := make([]productSnapshot, 0, len(items))
	for _, it := range items {
		p, err := s.products.GetByID(ctx, it.ProductID)
		if err != nil {
			return nil, fmt.Errorf("load product %s: %w", it.ProductID, err)
		}
		qty := it.Quantity
		if qty <= 0 {
			qty = 1
		}
		out = append(out, productSnapshot{
			ID:         p.ID,
			Name:       p.Name,
			PriceCents: p.PriceCents,
			Quantity:   qty,
		})
	}
	return out, nil
}

func (s *volunteerService) createSlotOrder(ctx context.Context, campaignID uuid.UUID, products []productSnapshot) error {
	ord, err := s.orders.Create(ctx, 0, order.StatusPaid, order.OriginShop, nil, nil, nil, nil, nil)
	if err != nil {
		return fmt.Errorf("create order: %w", err)
	}

	for _, p := range products {
		if _, err := s.lines.Create(ctx, ord.ID, orderline.LineTypeSimple, p.ID, truncatedTitle(p.Name), p.Quantity, 0, nil, nil, nil); err != nil {
			return fmt.Errorf("create order line: %w", err)
		}
	}

	if _, err := s.payments.Create(ctx, ord.ID, orderpayment.MethodGRATIS_STAFF, 0, time.Now(), nil); err != nil {
		return fmt.Errorf("create payment: %w", err)
	}

	if _, err := s.slots.Create(ctx, campaignID, ord.ID); err != nil {
		return fmt.Errorf("create slot: %w", err)
	}
	return nil
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
		summary, err := s.summarize(ctx, c)
		if err != nil {
			return nil, err
		}
		out = append(out, summary)
	}
	return out, nil
}

func (s *volunteerService) summarize(ctx context.Context, c *ent.VolunteerCampaign) (VolunteerCampaignSummary, error) {
	slots, err := s.slots.ListByCampaign(ctx, c.ID)
	if err != nil {
		return VolunteerCampaignSummary{}, err
	}
	total, redeemed, reserved := 0, 0, 0
	now := time.Now()
	for _, slot := range slots {
		total++
		if slotRedeemed(slot) {
			redeemed++
		} else if slot.ReservedBySession != nil && slot.ReservedUntil != nil && slot.ReservedUntil.After(now) {
			reserved++
		}
	}
	return VolunteerCampaignSummary{
		Campaign:      c,
		TotalSlots:    total,
		RedeemedSlots: redeemed,
		ReservedSlots: reserved,
	}, nil
}

func (s *volunteerService) GetCampaign(ctx context.Context, id uuid.UUID) (*VolunteerCampaignDetail, error) {
	campaign, err := s.campaigns.GetByID(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrVolunteerCampaignNotFound
		}
		return nil, err
	}
	products, err := s.campaigns.ListProducts(ctx, id)
	if err != nil {
		return nil, err
	}
	slots, err := s.slots.ListByCampaign(ctx, id)
	if err != nil {
		return nil, err
	}
	return &VolunteerCampaignDetail{
		Campaign: campaign,
		Products: products,
		Slots:    slots,
	}, nil
}

func (s *volunteerService) UpdateCampaign(ctx context.Context, id uuid.UUID, input UpdateVolunteerCampaignInput) (*ent.VolunteerCampaign, error) {
	if strings.TrimSpace(input.Name) == "" {
		return nil, errors.New("name_required")
	}
	if err := validateAccessCode(input.AccessCode); err != nil {
		return nil, err
	}

	tx, err := s.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := repository.ContextWithClient(ctx, tx.Client())

	campaign, err := s.campaigns.Update(txCtx, id, input.Name, input.AccessCode, input.ValidFrom, input.ValidUntil, input.Status)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrVolunteerCampaignNotFound
		}
		return nil, err
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
	tx, err := s.client.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := repository.ContextWithClient(ctx, tx.Client())

	if err := s.campaigns.SetStatus(txCtx, id, volunteercampaign.StatusEnded); err != nil {
		if ent.IsNotFound(err) {
			return ErrVolunteerCampaignNotFound
		}
		return err
	}

	slots, err := s.slots.ListByCampaign(txCtx, id)
	if err != nil {
		return err
	}
	for _, slot := range slots {
		if slotRedeemed(slot) {
			continue
		}
		if err := s.orders.UpdateStatus(txCtx, slot.OrderID, order.StatusCancelled); err != nil {
			return fmt.Errorf("cancel slot order %s: %w", slot.OrderID, err)
		}
	}
	return tx.Commit()
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

func (s *volunteerService) GetClaimCampaign(ctx context.Context, token uuid.UUID) (*ent.VolunteerCampaign, error) {
	campaign, err := s.campaigns.GetByClaimToken(ctx, token)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrVolunteerCampaignNotFound
		}
		return nil, err
	}
	return campaign, nil
}

func (s *volunteerService) ListClaimSlots(ctx context.Context, token uuid.UUID, sessionID string) (*VolunteerClaimView, error) {
	campaign, err := s.GetClaimCampaign(ctx, token)
	if err != nil {
		return nil, err
	}
	if campaign.Status != volunteercampaign.StatusActive {
		return nil, ErrVolunteerCampaignInactive
	}
	slots, err := s.slots.ListRedeemableByCampaign(ctx, campaign.ID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	view := &VolunteerClaimView{
		Campaign:       campaign,
		AvailableSlots: []VolunteerSlotView{},
		ReservedByMe:   []VolunteerSlotView{},
		TotalSlots:     len(slots),
	}

	for _, slot := range slots {
		if slotRedeemed(slot) {
			continue
		}
		v := toSlotView(slot)
		reservedByMe := slot.ReservedBySession != nil && *slot.ReservedBySession == sessionID &&
			slot.ReservedUntil != nil && slot.ReservedUntil.After(now)
		if reservedByMe {
			view.ReservedByMe = append(view.ReservedByMe, v)
			continue
		}
		available := slot.ReservedBySession == nil || slot.ReservedUntil == nil || slot.ReservedUntil.Before(now)
		if available {
			view.AvailableSlots = append(view.AvailableSlots, v)
		}
	}
	return view, nil
}

func (s *volunteerService) GetClaimSlot(ctx context.Context, token uuid.UUID, slotID uuid.UUID, sessionID string) (*VolunteerSlotView, error) {
	campaign, err := s.GetClaimCampaign(ctx, token)
	if err != nil {
		return nil, err
	}
	slot, err := s.slots.GetByID(ctx, slotID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrVolunteerSlotNotAvailable
		}
		return nil, err
	}
	if slot.CampaignID != campaign.ID {
		return nil, ErrVolunteerSlotNotAvailable
	}
	ownedSlots, err := s.slots.ListRedeemableByCampaign(ctx, campaign.ID)
	if err != nil {
		return nil, err
	}
	for _, full := range ownedSlots {
		if full.ID == slot.ID {
			view := toSlotView(full)
			return &view, nil
		}
	}
	v := VolunteerSlotView{Slot: slot, OrderID: slot.OrderID}
	return &v, nil
}

func (s *volunteerService) ReserveSlot(ctx context.Context, token uuid.UUID, slotID uuid.UUID, sessionID string) error {
	campaign, err := s.GetClaimCampaign(ctx, token)
	if err != nil {
		return err
	}
	if campaign.Status != volunteercampaign.StatusActive {
		return ErrVolunteerCampaignInactive
	}
	if !campaignWithinValidity(campaign, time.Now()) {
		return ErrVolunteerCampaignOutsideValid
	}
	slot, err := s.slots.GetByID(ctx, slotID)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrVolunteerSlotNotAvailable
		}
		return err
	}
	if slot.CampaignID != campaign.ID {
		return ErrVolunteerSlotNotAvailable
	}

	until := time.Now().Add(volunteerReservationTTL)
	_, ok, err := s.slots.ReserveAtomic(ctx, slotID, sessionID, until)
	if err != nil {
		return err
	}
	if !ok {
		return ErrVolunteerSlotNotAvailable
	}
	return nil
}

func (s *volunteerService) ReleaseSlot(ctx context.Context, token uuid.UUID, slotID uuid.UUID, sessionID string) error {
	campaign, err := s.GetClaimCampaign(ctx, token)
	if err != nil {
		return err
	}
	slot, err := s.slots.GetByID(ctx, slotID)
	if err != nil {
		if ent.IsNotFound(err) {
			return ErrVolunteerSlotNotAvailable
		}
		return err
	}
	if slot.CampaignID != campaign.ID {
		return ErrVolunteerSlotNotAvailable
	}
	ok, err := s.slots.Release(ctx, slotID, sessionID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrVolunteerSlotNotYours
	}
	return nil
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

func slotRedeemed(slot *ent.VolunteerSlot) bool {
	ord, err := slot.Edges.OrderOrErr()
	if err != nil || ord == nil {
		return false
	}
	lines := ord.Edges.Lines
	if len(lines) == 0 {
		return false
	}
	for _, l := range lines {
		r, _ := l.Edges.RedemptionOrErr()
		if r == nil {
			return false
		}
	}
	return true
}

func toSlotView(slot *ent.VolunteerSlot) VolunteerSlotView {
	v := VolunteerSlotView{
		Slot:          slot,
		OrderID:       slot.OrderID,
		ReservedUntil: slot.ReservedUntil,
	}
	if slot.ReservedBySession != nil {
		v.IsReservedBy = *slot.ReservedBySession
	}
	ord, err := slot.Edges.OrderOrErr()
	if err == nil && ord != nil {
		allRedeemed := len(ord.Edges.Lines) > 0
		var earliest *time.Time
		for _, l := range ord.Edges.Lines {
			r, _ := l.Edges.RedemptionOrErr()
			productName := ""
			var productImage *string
			if p, _ := l.Edges.ProductOrErr(); p != nil {
				productName = p.Name
				productImage = p.Image
			} else {
				productName = l.Title
			}
			var redeemedAt *time.Time
			if r != nil {
				ra := r.RedeemedAt
				redeemedAt = &ra
				if earliest == nil || ra.Before(*earliest) {
					earliest = &ra
				}
			} else {
				allRedeemed = false
			}
			v.Lines = append(v.Lines, VolunteerSlotLine{
				ProductName:  productName,
				ProductImage: productImage,
				Quantity:     l.Quantity,
				IsRedeemed:   r != nil,
				RedeemedAt:   redeemedAt,
			})
		}
		if allRedeemed {
			v.RedeemedAt = earliest
		}
	}
	return v
}

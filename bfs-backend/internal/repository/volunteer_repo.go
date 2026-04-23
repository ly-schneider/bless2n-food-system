package repository

import (
	"context"
	"time"

	"backend/internal/generated/ent"
	"backend/internal/generated/ent/order"
	"backend/internal/generated/ent/volunteercampaign"
	"backend/internal/generated/ent/volunteercampaignproduct"
	"backend/internal/generated/ent/volunteerslot"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
)

type VolunteerCampaignRepository interface {
	Create(ctx context.Context, name, accessCode string, validFrom, validUntil *time.Time, status volunteercampaign.Status) (*ent.VolunteerCampaign, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.VolunteerCampaign, error)
	GetByIDWithProducts(ctx context.Context, id uuid.UUID) (*ent.VolunteerCampaign, error)
	GetByClaimToken(ctx context.Context, token uuid.UUID) (*ent.VolunteerCampaign, error)
	List(ctx context.Context) ([]*ent.VolunteerCampaign, error)
	Update(ctx context.Context, id uuid.UUID, name, accessCode string, validFrom, validUntil *time.Time, status volunteercampaign.Status) (*ent.VolunteerCampaign, error)
	RotateClaimToken(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
	SetStatus(ctx context.Context, id uuid.UUID, status volunteercampaign.Status) error

	ReplaceProducts(ctx context.Context, campaignID uuid.UUID, items []VolunteerCampaignProductInput) error
	ListProducts(ctx context.Context, campaignID uuid.UUID) ([]*ent.VolunteerCampaignProduct, error)
}

type VolunteerCampaignProductInput struct {
	ProductID uuid.UUID
	Quantity  int
}

type VolunteerSlotRepository interface {
	Create(ctx context.Context, campaignID, orderID uuid.UUID) (*ent.VolunteerSlot, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.VolunteerSlot, error)
	ListByCampaign(ctx context.Context, campaignID uuid.UUID) ([]*ent.VolunteerSlot, error)
	ListRedeemableByCampaign(ctx context.Context, campaignID uuid.UUID) ([]*ent.VolunteerSlot, error)
	ReserveAtomic(ctx context.Context, slotID uuid.UUID, sessionID string, until time.Time) (*ent.VolunteerSlot, bool, error)
	Release(ctx context.Context, slotID uuid.UUID, sessionID string) (bool, error)
	ListUnredeemedOrderIDs(ctx context.Context, campaignID uuid.UUID) ([]uuid.UUID, error)
}

type volunteerCampaignRepo struct {
	client *ent.Client
}

type volunteerSlotRepo struct {
	client *ent.Client
}

func NewVolunteerCampaignRepository(client *ent.Client) VolunteerCampaignRepository {
	return &volunteerCampaignRepo{client: client}
}

func NewVolunteerSlotRepository(client *ent.Client) VolunteerSlotRepository {
	return &volunteerSlotRepo{client: client}
}

func (r *volunteerCampaignRepo) ec(ctx context.Context) *ent.Client {
	return ClientFromContext(ctx, r.client)
}

func (r *volunteerSlotRepo) ec(ctx context.Context) *ent.Client {
	return ClientFromContext(ctx, r.client)
}

func (r *volunteerCampaignRepo) Create(ctx context.Context, name, accessCode string, validFrom, validUntil *time.Time, status volunteercampaign.Status) (*ent.VolunteerCampaign, error) {
	b := r.ec(ctx).VolunteerCampaign.Create().
		SetName(name).
		SetAccessCode(accessCode).
		SetStatus(status)
	if validFrom != nil {
		b.SetValidFrom(*validFrom)
	}
	if validUntil != nil {
		b.SetValidUntil(*validUntil)
	}
	created, err := b.Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return created, nil
}

func (r *volunteerCampaignRepo) GetByID(ctx context.Context, id uuid.UUID) (*ent.VolunteerCampaign, error) {
	e, err := r.ec(ctx).VolunteerCampaign.Get(ctx, id)
	if err != nil {
		return nil, translateError(err)
	}
	return e, nil
}

func (r *volunteerCampaignRepo) GetByIDWithProducts(ctx context.Context, id uuid.UUID) (*ent.VolunteerCampaign, error) {
	e, err := r.ec(ctx).VolunteerCampaign.Query().
		Where(volunteercampaign.ID(id)).
		WithCampaignProducts(func(q *ent.VolunteerCampaignProductQuery) {
			q.WithProduct()
		}).
		Only(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return e, nil
}

func (r *volunteerCampaignRepo) GetByClaimToken(ctx context.Context, token uuid.UUID) (*ent.VolunteerCampaign, error) {
	e, err := r.ec(ctx).VolunteerCampaign.Query().
		Where(volunteercampaign.ClaimTokenEQ(token)).
		Only(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return e, nil
}

func (r *volunteerCampaignRepo) List(ctx context.Context) ([]*ent.VolunteerCampaign, error) {
	rows, err := r.ec(ctx).VolunteerCampaign.Query().
		Order(volunteercampaign.ByCreatedAt(entDescOpt())).
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

func (r *volunteerCampaignRepo) Update(ctx context.Context, id uuid.UUID, name, accessCode string, validFrom, validUntil *time.Time, status volunteercampaign.Status) (*ent.VolunteerCampaign, error) {
	b := r.ec(ctx).VolunteerCampaign.UpdateOneID(id).
		SetName(name).
		SetAccessCode(accessCode).
		SetStatus(status)
	if validFrom != nil {
		b.SetValidFrom(*validFrom)
	} else {
		b.ClearValidFrom()
	}
	if validUntil != nil {
		b.SetValidUntil(*validUntil)
	} else {
		b.ClearValidUntil()
	}
	updated, err := b.Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return updated, nil
}

func (r *volunteerCampaignRepo) RotateClaimToken(ctx context.Context, id uuid.UUID) (uuid.UUID, error) {
	newToken := uuid.Must(uuid.NewV7())
	_, err := r.ec(ctx).VolunteerCampaign.UpdateOneID(id).
		SetClaimToken(newToken).
		Save(ctx)
	if err != nil {
		return uuid.Nil, translateError(err)
	}
	return newToken, nil
}

func (r *volunteerCampaignRepo) SetStatus(ctx context.Context, id uuid.UUID, status volunteercampaign.Status) error {
	_, err := r.ec(ctx).VolunteerCampaign.UpdateOneID(id).
		SetStatus(status).
		Save(ctx)
	return translateError(err)
}

func (r *volunteerCampaignRepo) ReplaceProducts(ctx context.Context, campaignID uuid.UUID, items []VolunteerCampaignProductInput) error {
	client := r.ec(ctx)
	_, err := client.VolunteerCampaignProduct.Delete().
		Where(volunteercampaignproduct.CampaignIDEQ(campaignID)).
		Exec(ctx)
	if err != nil {
		return translateError(err)
	}
	if len(items) == 0 {
		return nil
	}
	builders := make([]*ent.VolunteerCampaignProductCreate, len(items))
	for i, it := range items {
		builders[i] = client.VolunteerCampaignProduct.Create().
			SetCampaignID(campaignID).
			SetProductID(it.ProductID).
			SetQuantity(it.Quantity)
	}
	_, err = client.VolunteerCampaignProduct.CreateBulk(builders...).Save(ctx)
	return translateError(err)
}

func (r *volunteerCampaignRepo) ListProducts(ctx context.Context, campaignID uuid.UUID) ([]*ent.VolunteerCampaignProduct, error) {
	rows, err := r.ec(ctx).VolunteerCampaignProduct.Query().
		Where(volunteercampaignproduct.CampaignIDEQ(campaignID)).
		WithProduct().
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

func (r *volunteerSlotRepo) Create(ctx context.Context, campaignID, orderID uuid.UUID) (*ent.VolunteerSlot, error) {
	created, err := r.ec(ctx).VolunteerSlot.Create().
		SetCampaignID(campaignID).
		SetOrderID(orderID).
		Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return created, nil
}

func (r *volunteerSlotRepo) GetByID(ctx context.Context, id uuid.UUID) (*ent.VolunteerSlot, error) {
	e, err := r.ec(ctx).VolunteerSlot.Get(ctx, id)
	if err != nil {
		return nil, translateError(err)
	}
	return e, nil
}

func (r *volunteerSlotRepo) ListByCampaign(ctx context.Context, campaignID uuid.UUID) ([]*ent.VolunteerSlot, error) {
	rows, err := r.ec(ctx).VolunteerSlot.Query().
		Where(volunteerslot.CampaignIDEQ(campaignID)).
		WithOrder(func(q *ent.OrderQuery) {
			q.WithLines(func(lq *ent.OrderLineQuery) {
				lq.WithRedemption()
			})
		}).
		Order(volunteerslot.ByCreatedAt()).
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

func (r *volunteerSlotRepo) ListRedeemableByCampaign(ctx context.Context, campaignID uuid.UUID) ([]*ent.VolunteerSlot, error) {
	rows, err := r.ec(ctx).VolunteerSlot.Query().
		Where(
			volunteerslot.CampaignIDEQ(campaignID),
			volunteerslot.HasOrderWith(order.StatusEQ(order.StatusPaid)),
		).
		WithOrder(func(q *ent.OrderQuery) {
			q.WithLines(func(lq *ent.OrderLineQuery) {
				lq.WithProduct()
				lq.WithRedemption()
			})
		}).
		Order(volunteerslot.ByCreatedAt()).
		All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

func (r *volunteerSlotRepo) ReserveAtomic(ctx context.Context, slotID uuid.UUID, sessionID string, until time.Time) (*ent.VolunteerSlot, bool, error) {
	client := r.ec(ctx)
	now := time.Now()

	n, err := client.VolunteerSlot.Update().
		Where(
			volunteerslot.IDEQ(slotID),
			volunteerslot.Or(
				volunteerslot.ReservedBySessionIsNil(),
				volunteerslot.ReservedBySessionEQ(sessionID),
				volunteerslot.ReservedUntilLT(now),
			),
		).
		SetReservedBySession(sessionID).
		SetReservedAt(now).
		SetReservedUntil(until).
		Save(ctx)
	if err != nil {
		return nil, false, translateError(err)
	}
	if n == 0 {
		return nil, false, nil
	}
	slot, err := client.VolunteerSlot.Get(ctx, slotID)
	if err != nil {
		return nil, false, translateError(err)
	}
	return slot, true, nil
}

func (r *volunteerSlotRepo) Release(ctx context.Context, slotID uuid.UUID, sessionID string) (bool, error) {
	n, err := r.ec(ctx).VolunteerSlot.Update().
		Where(
			volunteerslot.IDEQ(slotID),
			volunteerslot.ReservedBySessionEQ(sessionID),
		).
		ClearReservedBySession().
		ClearReservedAt().
		ClearReservedUntil().
		Save(ctx)
	if err != nil {
		return false, translateError(err)
	}
	return n > 0, nil
}

func (r *volunteerSlotRepo) ListUnredeemedOrderIDs(ctx context.Context, campaignID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.ec(ctx).VolunteerSlot.Query().
		Where(volunteerslot.CampaignIDEQ(campaignID)).
		Modify(func(s *sql.Selector) {
			s.Select(s.C(volunteerslot.FieldOrderID))
		}).
		Scan(ctx, &ids)
	if err != nil {
		return nil, translateError(err)
	}
	return ids, nil
}

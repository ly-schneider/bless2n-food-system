package repository

import (
	"context"
	"time"

	"backend/internal/generated/ent"
	"backend/internal/generated/ent/volunteercampaign"
	"backend/internal/generated/ent/volunteercampaignproduct"
	"backend/internal/generated/ent/volunteerredemption"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
)

type VolunteerCampaignRepository interface {
	Create(ctx context.Context, name, accessCode string, validFrom, validUntil *time.Time, status volunteercampaign.Status, maxRedemptions int) (*ent.VolunteerCampaign, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.VolunteerCampaign, error)
	GetByIDWithProducts(ctx context.Context, id uuid.UUID) (*ent.VolunteerCampaign, error)
	GetByClaimToken(ctx context.Context, token uuid.UUID) (*ent.VolunteerCampaign, error)
	List(ctx context.Context) ([]*ent.VolunteerCampaign, error)
	Update(ctx context.Context, id uuid.UUID, name, accessCode string, validFrom, validUntil *time.Time, status volunteercampaign.Status) (*ent.VolunteerCampaign, error)
	UpdateMaxRedemptions(ctx context.Context, id uuid.UUID, newMax int) (*ent.VolunteerCampaign, bool, error)
	RotateClaimToken(ctx context.Context, id uuid.UUID) (uuid.UUID, error)
	SetStatus(ctx context.Context, id uuid.UUID, status volunteercampaign.Status) error

	ReplaceProducts(ctx context.Context, campaignID uuid.UUID, items []VolunteerCampaignProductInput) error
	ListProducts(ctx context.Context, campaignID uuid.UUID) ([]*ent.VolunteerCampaignProduct, error)

	// IncrementRedemptionAtomic conditionally increments redemption_count iff the campaign
	// is active, inside its validity window, and under max_redemptions. Returns true when
	// the increment succeeded.
	IncrementRedemptionAtomic(ctx context.Context, campaignID uuid.UUID) (bool, error)
}

type VolunteerCampaignProductInput struct {
	ProductID uuid.UUID
	Quantity  int
}

type VolunteerRedemptionRepository interface {
	Create(ctx context.Context, campaignID, orderID uuid.UUID, stationDeviceID *uuid.UUID, idempotencyKey *string) (*ent.VolunteerRedemption, error)
	GetByIdempotencyKey(ctx context.Context, campaignID uuid.UUID, key string) (*ent.VolunteerRedemption, error)
	ListByCampaign(ctx context.Context, campaignID uuid.UUID, limit int) ([]*ent.VolunteerRedemption, error)
}

type volunteerCampaignRepo struct {
	client *ent.Client
}

type volunteerRedemptionRepo struct {
	client *ent.Client
}

func NewVolunteerCampaignRepository(client *ent.Client) VolunteerCampaignRepository {
	return &volunteerCampaignRepo{client: client}
}

func NewVolunteerRedemptionRepository(client *ent.Client) VolunteerRedemptionRepository {
	return &volunteerRedemptionRepo{client: client}
}

func (r *volunteerCampaignRepo) ec(ctx context.Context) *ent.Client {
	return ClientFromContext(ctx, r.client)
}

func (r *volunteerRedemptionRepo) ec(ctx context.Context) *ent.Client {
	return ClientFromContext(ctx, r.client)
}

func (r *volunteerCampaignRepo) Create(ctx context.Context, name, accessCode string, validFrom, validUntil *time.Time, status volunteercampaign.Status, maxRedemptions int) (*ent.VolunteerCampaign, error) {
	b := r.ec(ctx).VolunteerCampaign.Create().
		SetName(name).
		SetAccessCode(accessCode).
		SetStatus(status).
		SetMaxRedemptions(maxRedemptions)
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

// UpdateMaxRedemptions updates max_redemptions only if newMax >= current redemption_count.
// Returns (updated, ok, err). ok=false means the new value is below the current redemption count.
func (r *volunteerCampaignRepo) UpdateMaxRedemptions(ctx context.Context, id uuid.UUID, newMax int) (*ent.VolunteerCampaign, bool, error) {
	client := r.ec(ctx)
	n, err := client.VolunteerCampaign.Update().
		Where(
			volunteercampaign.IDEQ(id),
			volunteercampaign.RedemptionCountLTE(newMax),
		).
		SetMaxRedemptions(newMax).
		Save(ctx)
	if err != nil {
		return nil, false, translateError(err)
	}
	if n == 0 {
		return nil, false, nil
	}
	updated, err := client.VolunteerCampaign.Get(ctx, id)
	if err != nil {
		return nil, false, translateError(err)
	}
	return updated, true, nil
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

func (r *volunteerCampaignRepo) IncrementRedemptionAtomic(ctx context.Context, campaignID uuid.UUID) (bool, error) {
	now := time.Now()
	n, err := r.ec(ctx).VolunteerCampaign.Update().
		Where(
			volunteercampaign.IDEQ(campaignID),
			volunteercampaign.StatusEQ(volunteercampaign.StatusActive),
			func(s *sql.Selector) {
				s.Where(sql.ColumnsLT(
					s.C(volunteercampaign.FieldRedemptionCount),
					s.C(volunteercampaign.FieldMaxRedemptions),
				))
				s.Where(sql.Or(
					sql.IsNull(s.C(volunteercampaign.FieldValidFrom)),
					sql.LTE(s.C(volunteercampaign.FieldValidFrom), now),
				))
				s.Where(sql.Or(
					sql.IsNull(s.C(volunteercampaign.FieldValidUntil)),
					sql.GT(s.C(volunteercampaign.FieldValidUntil), now),
				))
			},
		).
		AddRedemptionCount(1).
		Save(ctx)
	if err != nil {
		return false, translateError(err)
	}
	return n > 0, nil
}

func (r *volunteerRedemptionRepo) Create(ctx context.Context, campaignID, orderID uuid.UUID, stationDeviceID *uuid.UUID, idempotencyKey *string) (*ent.VolunteerRedemption, error) {
	b := r.ec(ctx).VolunteerRedemption.Create().
		SetCampaignID(campaignID).
		SetOrderID(orderID)
	if stationDeviceID != nil {
		b.SetStationDeviceID(*stationDeviceID)
	}
	if idempotencyKey != nil {
		b.SetIdempotencyKey(*idempotencyKey)
	}
	created, err := b.Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return created, nil
}

func (r *volunteerRedemptionRepo) GetByIdempotencyKey(ctx context.Context, campaignID uuid.UUID, key string) (*ent.VolunteerRedemption, error) {
	row, err := r.ec(ctx).VolunteerRedemption.Query().
		Where(
			volunteerredemption.CampaignIDEQ(campaignID),
			volunteerredemption.IdempotencyKeyEQ(key),
		).
		Only(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return row, nil
}

func (r *volunteerRedemptionRepo) ListByCampaign(ctx context.Context, campaignID uuid.UUID, limit int) ([]*ent.VolunteerRedemption, error) {
	q := r.ec(ctx).VolunteerRedemption.Query().
		Where(volunteerredemption.CampaignIDEQ(campaignID)).
		Order(volunteerredemption.ByCreatedAt(entDescOpt()))
	if limit > 0 {
		q = q.Limit(limit)
	}
	rows, err := q.All(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return rows, nil
}

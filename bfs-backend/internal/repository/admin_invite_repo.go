package repository

import (
	"context"
	"time"

	"backend/internal/generated/ent"
	"backend/internal/generated/ent/admininvite"
	"backend/internal/generated/ent/predicate"

	"github.com/google/uuid"
)

type AdminInviteRepository interface {
	Create(ctx context.Context, invitedByUserID, inviteeEmail, tokenHash string, status admininvite.Status, expiresAt time.Time) (*ent.AdminInvite, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ent.AdminInvite, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (*ent.AdminInvite, error)
	List(ctx context.Context, status *admininvite.Status, email *string, limit, offset int) ([]*ent.AdminInvite, int64, error)
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, status admininvite.Status) error
	UpdateTokenAndExpiry(ctx context.Context, id uuid.UUID, tokenHash string, expiresAt time.Time) error
	MarkAccepted(ctx context.Context, id uuid.UUID) error
}

type adminInviteRepo struct {
	client *ent.Client
}

func NewAdminInviteRepository(client *ent.Client) AdminInviteRepository {
	return &adminInviteRepo{client: client}
}

func (r *adminInviteRepo) ec(ctx context.Context) *ent.Client {
	return ClientFromContext(ctx, r.client)
}

func (r *adminInviteRepo) Create(ctx context.Context, invitedByUserID, inviteeEmail, tokenHash string, status admininvite.Status, expiresAt time.Time) (*ent.AdminInvite, error) {
	created, err := r.ec(ctx).AdminInvite.Create().
		SetInvitedByUserID(invitedByUserID).
		SetInviteeEmail(inviteeEmail).
		SetTokenHash(tokenHash).
		SetStatus(status).
		SetExpiresAt(expiresAt).
		Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return created, nil
}

func (r *adminInviteRepo) GetByID(ctx context.Context, id uuid.UUID) (*ent.AdminInvite, error) {
	e, err := r.ec(ctx).AdminInvite.Get(ctx, id)
	if err != nil {
		return nil, translateError(err)
	}
	return e, nil
}

func (r *adminInviteRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*ent.AdminInvite, error) {
	e, err := r.ec(ctx).AdminInvite.Query().
		Where(admininvite.TokenHashEQ(tokenHash)).
		Only(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return e, nil
}

func (r *adminInviteRepo) List(ctx context.Context, status *admininvite.Status, email *string, limit, offset int) ([]*ent.AdminInvite, int64, error) {
	var filters []predicate.AdminInvite
	if status != nil {
		filters = append(filters, admininvite.StatusEQ(*status))
	}
	if email != nil && *email != "" {
		filters = append(filters, admininvite.InviteeEmailContainsFold(*email))
	}

	total, err := r.ec(ctx).AdminInvite.Query().Where(filters...).Count(ctx)
	if err != nil {
		return nil, 0, translateError(err)
	}

	rows, err := r.ec(ctx).AdminInvite.Query().
		Where(filters...).
		WithInviter().
		Order(admininvite.ByCreatedAt(entDescOpt())).
		Limit(limit).
		Offset(offset).
		All(ctx)
	if err != nil {
		return nil, 0, translateError(err)
	}

	return rows, int64(total), nil
}

func (r *adminInviteRepo) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.ec(ctx).AdminInvite.DeleteOneID(id).Exec(ctx)
	return translateError(err)
}

func (r *adminInviteRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status admininvite.Status) error {
	n, err := r.ec(ctx).AdminInvite.Update().
		Where(admininvite.IDEQ(id)).
		SetStatus(status).
		Save(ctx)
	if err != nil {
		return translateError(err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *adminInviteRepo) UpdateTokenAndExpiry(ctx context.Context, id uuid.UUID, tokenHash string, expiresAt time.Time) error {
	n, err := r.ec(ctx).AdminInvite.Update().
		Where(admininvite.IDEQ(id)).
		SetTokenHash(tokenHash).
		SetExpiresAt(expiresAt).
		Save(ctx)
	if err != nil {
		return translateError(err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *adminInviteRepo) MarkAccepted(ctx context.Context, id uuid.UUID) error {
	n, err := r.ec(ctx).AdminInvite.Update().
		Where(admininvite.IDEQ(id)).
		SetStatus(admininvite.StatusAccepted).
		SetUsedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return translateError(err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

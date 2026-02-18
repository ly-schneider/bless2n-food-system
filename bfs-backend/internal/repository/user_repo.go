package repository

import (
	"context"
	cryptoRand "crypto/rand"

	"backend/internal/generated/ent"
	"backend/internal/generated/ent/user"
)

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*ent.User, error)
	GetByID(ctx context.Context, id string) (*ent.User, error)
	List(ctx context.Context, limit, offset int) ([]*ent.User, int64, error)
	UpdateRole(ctx context.Context, id string, role user.Role) error
	UpdateRoleAndName(ctx context.Context, id string, role user.Role, name string) error
	CreateAdminUser(ctx context.Context, email, name string) (*ent.User, error)
	Delete(ctx context.Context, id string) error
}

type userRepo struct {
	client *ent.Client
}

func NewUserRepository(client *ent.Client) UserRepository {
	return &userRepo{client: client}
}

func (r *userRepo) ec(ctx context.Context) *ent.Client {
	return ClientFromContext(ctx, r.client)
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*ent.User, error) {
	e, err := r.ec(ctx).User.Query().
		Where(user.EmailEQ(email)).
		Only(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return e, nil
}

func (r *userRepo) GetByID(ctx context.Context, id string) (*ent.User, error) {
	e, err := r.ec(ctx).User.Get(ctx, id)
	if err != nil {
		return nil, translateError(err)
	}
	return e, nil
}

func (r *userRepo) List(ctx context.Context, limit, offset int) ([]*ent.User, int64, error) {
	total, err := r.ec(ctx).User.Query().Count(ctx)
	if err != nil {
		return nil, 0, translateError(err)
	}

	rows, err := r.ec(ctx).User.Query().
		Order(user.ByCreatedAt(entDescOpt())).
		Limit(limit).
		Offset(offset).
		All(ctx)
	if err != nil {
		return nil, 0, translateError(err)
	}

	return rows, int64(total), nil
}

func (r *userRepo) UpdateRole(ctx context.Context, id string, role user.Role) error {
	n, err := r.ec(ctx).User.Update().
		Where(user.IDEQ(id)).
		SetRole(role).
		Save(ctx)
	if err != nil {
		return translateError(err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *userRepo) UpdateRoleAndName(ctx context.Context, id string, role user.Role, name string) error {
	n, err := r.ec(ctx).User.Update().
		Where(user.IDEQ(id)).
		SetRole(role).
		SetName(name).
		Save(ctx)
	if err != nil {
		return translateError(err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *userRepo) CreateAdminUser(ctx context.Context, email, name string) (*ent.User, error) {
	e, err := r.ec(ctx).User.Create().
		SetID(generateUserID()).
		SetEmail(email).
		SetName(name).
		SetEmailVerified(true).
		SetIsAnonymous(false).
		SetRole(user.RoleAdmin).
		Save(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	return e, nil
}

func (r *userRepo) Delete(ctx context.Context, id string) error {
	n, err := r.ec(ctx).User.Delete().
		Where(user.IDEQ(id)).
		Exec(ctx)
	if err != nil {
		return translateError(err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// generateUserID creates a unique ID for a new user.
// Better Auth typically uses IDs with a prefix.
func generateUserID() string {
	return "u_" + generateRandomID(24)
}

// generateRandomID creates a random alphanumeric ID of the given length.
func generateRandomID(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	randBytes := make([]byte, length)
	_, _ = cryptoRand.Read(randBytes)
	for i := range b {
		b[i] = chars[int(randBytes[i])%len(chars)]
	}
	return string(b)
}

package service

import (
	"context"

	"backend/internal/generated/ent"
	"backend/internal/generated/ent/user"
	"backend/internal/repository"

	"entgo.io/ent/dialect/sql"
)

type UserService interface {
	List(ctx context.Context, role *string, limit, offset int) ([]*ent.User, int64, error)
	GetByID(ctx context.Context, id string) (*ent.User, error)
	UpdateRole(ctx context.Context, id string, role string) error
	UpdateName(ctx context.Context, id string, name string) error
	Delete(ctx context.Context, id string) error
}

type userService struct {
	client   *ent.Client
	userRepo repository.UserRepository
}

func NewUserService(
	client *ent.Client,
	userRepo repository.UserRepository,
) UserService {
	return &userService{
		client:   client,
		userRepo: userRepo,
	}
}

func (s *userService) List(ctx context.Context, role *string, limit, offset int) ([]*ent.User, int64, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	q := s.client.User.Query()
	if role != nil && *role != "" {
		q = q.Where(user.RoleEQ(user.Role(*role)))
	}

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	users, err := q.
		Order(user.ByCreatedAt(entDescOrder())).
		Limit(limit).
		Offset(offset).
		All(ctx)
	if err != nil {
		return nil, 0, err
	}

	return users, int64(total), nil
}

func (s *userService) GetByID(ctx context.Context, id string) (*ent.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

func (s *userService) UpdateRole(ctx context.Context, id string, role string) error {
	return s.userRepo.UpdateRole(ctx, id, user.Role(role))
}

func (s *userService) UpdateName(ctx context.Context, id string, name string) error {
	n, err := s.client.User.Update().
		Where(user.IDEQ(id)).
		SetName(name).
		Save(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func (s *userService) Delete(ctx context.Context, id string) error {
	return s.userRepo.Delete(ctx, id)
}

func entDescOrder() sql.OrderTermOption {
	return sql.OrderDesc()
}

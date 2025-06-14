package service

import (
	"context"
	"errors"

	"backend/internal/domain"
	"backend/internal/logger"
	"backend/internal/repository"
)

type RoleService interface {
	List(ctx context.Context) ([]domain.Role, error)
	Get(ctx context.Context, id string) (domain.Role, error)
	GetByName(ctx context.Context, name string) (domain.Role, error)
	Create(ctx context.Context, r *domain.Role) error
	Update(ctx context.Context, id string, in *domain.Role) (domain.Role, error)
	Delete(ctx context.Context, id string) error
}

type roleService struct {
	repo repository.RoleRepository
}

func NewRoleService(r repository.RoleRepository) RoleService {
	return &roleService{repo: r}
}

func (s *roleService) List(ctx context.Context) ([]domain.Role, error) {
	logger.L.Info("Listing roles")
	roles, err := s.repo.List(ctx)
	if err != nil {
		logger.L.Errorw("Failed to list roles", "error", err)
		return nil, err
	}
	logger.L.Infow("Successfully listed roles", "count", len(roles))
	return roles, nil
}

func (s *roleService) Get(ctx context.Context, id string) (domain.Role, error) {
	logger.L.Infow("Getting role", "id", id)

	if id == "" {
		err := errors.New("role ID cannot be empty")
		logger.L.Error(err.Error())
		return domain.Role{}, err
	}

	role, err := s.repo.Get(ctx, id)
	if err != nil {
		logger.L.Errorw("Failed to get role", "id", id, "error", err)
		return role, err
	}
	logger.L.Infow("Successfully retrieved role", "id", id, "name", role.Name)
	return role, nil
}

func (s *roleService) GetByName(ctx context.Context, name string) (domain.Role, error) {
	logger.L.Infow("Getting role by name", "name", name)

	if name == "" {
		err := errors.New("role name cannot be empty")
		logger.L.Error(err.Error())
		return domain.Role{}, err
	}

	role, err := s.repo.GetByName(ctx, name)
	if err != nil {
		logger.L.Errorw("Failed to get role by name", "name", name, "error", err)
		return role, err
	}
	logger.L.Infow("Successfully retrieved role by name", "id", role.ID, "name", name)
	return role, nil
}

func (s *roleService) Create(ctx context.Context, r *domain.Role) error {
	logger.L.Infow("Creating role", "name", r.Name)

	if r.Name == "" {
		err := errors.New("role name cannot be empty")
		logger.L.Error(err.Error())
		return err
	}

	if err := s.repo.Create(ctx, r); err != nil {
		logger.L.Errorw("Failed to create role", "name", r.Name, "error", err)
		return err
	}

	logger.L.Infow("Role created successfully", "id", r.ID, "name", r.Name)
	return nil
}

func (s *roleService) Update(ctx context.Context, id string, in *domain.Role) (domain.Role, error) {
	logger.L.Infow("Updating role", "id", id, "name", in.Name)

	if id == "" {
		err := errors.New("role ID cannot be empty")
		logger.L.Error(err.Error())
		return domain.Role{}, err
	}

	r, err := s.repo.Get(ctx, id)
	if err != nil {
		logger.L.Errorw("Role not found for update", "id", id, "error", err)
		return r, err
	}

	r.Name = in.Name

	if err := s.repo.Update(ctx, &r); err != nil {
		logger.L.Errorw("Failed to update role", "id", id, "error", err)
		return r, err
	}

	logger.L.Infow("Role updated successfully", "id", id, "name", r.Name)
	return r, nil
}

func (s *roleService) Delete(ctx context.Context, id string) error {
	logger.L.Infow("Deleting role", "id", id)

	if id == "" {
		err := errors.New("role ID cannot be empty")
		logger.L.Error(err.Error())
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		logger.L.Errorw("Failed to delete role", "id", id, "error", err)
		return err
	}

	logger.L.Infow("Role deleted successfully", "id", id)
	return nil
}

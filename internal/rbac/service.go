package rbac

import (
	"context"
	"errors"
	"time"
	"pos_bos/internal/core/domain"
	"pos_bos/pkg/validation"
	"github.com/go-playground/validator/v10"
)

var (
	ErrPermissionAlreadyExists = errors.New("permission name already exists")
)

type rbacService struct {
	rbacRepo domain.RBACRepository
	validate *validator.Validate
}

func NewRBACService(repo domain.RBACRepository) domain.RBACService {
	return &rbacService{
		rbacRepo: repo,
		validate: validator.New(),
	}
}

func (s *rbacService) CreatePermission(ctx context.Context, req domain.PermissionRequest) (*domain.Permission, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, validation.FormatValidationError(err)
	}

	existing, err := s.rbacRepo.GetPermissionByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrPermissionAlreadyExists
	}

	return s.rbacRepo.CreatePermission(ctx, req.Name)
}

func (s *rbacService) GetAllPermissions(ctx context.Context, req domain.PaginationRequest) (domain.PaginationResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}

	permissions, total, err := s.rbacRepo.GetAllPermissions(ctx, req)
	if err != nil {
		return domain.PaginationResponse{}, err
	}

	totalPages := total / req.Limit
	if total%req.Limit > 0 {
		totalPages++
	}

	return domain.PaginationResponse{
		Data: permissions,
		Meta: domain.PaginationMeta{
			Page:         req.Page,
			Limit:        req.Limit,
			TotalRecords: total,
			TotalPages:   totalPages,
		},
	}, nil
}

func (s *rbacService) UpdatePermission(ctx context.Context, id int, req domain.PermissionRequest) (*domain.Permission, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, validation.FormatValidationError(err)
	}
	
	existing, err := s.rbacRepo.GetPermissionByName(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if existing != nil && existing.ID != id {
		return nil, ErrPermissionAlreadyExists
	}

	return s.rbacRepo.UpdatePermission(ctx, id, req.Name)
}

func (s *rbacService) DeletePermission(ctx context.Context, id int) error {
	return s.rbacRepo.DeletePermission(ctx, id)
}

func (s *rbacService) CreateRole(ctx context.Context, req domain.RoleRequest) (*domain.RoleResponse, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, validation.FormatValidationError(err)
	}

	exists, err := s.rbacRepo.CheckRoleNameExists(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("role name already exists")
	}

	role, err := s.rbacRepo.CreateRole(ctx, req.Name)
	if err != nil {
		return nil, err
	}

	if err := s.rbacRepo.AssignPermissionToRole(ctx, role.ID, req.PermissionIDs); err != nil {
		return nil, err
	}

	perms, _ := s.rbacRepo.GetRolePermissions(ctx, role.ID)

	return &domain.RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Permissions: perms,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}, nil
}

func (s *rbacService) GetAllRoles(ctx context.Context, req domain.PaginationRequest) (domain.PaginationResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}

	roles, total, err := s.rbacRepo.GetAllRoles(ctx, req)
	if err != nil {
		return domain.PaginationResponse{}, err
	}

	var roleResponses []domain.RoleResponse
	for _, role := range roles {
		perms, _ := s.rbacRepo.GetRolePermissions(ctx, role.ID)
		roleResponses = append(roleResponses, domain.RoleResponse{
			ID:          role.ID,
			Name:        role.Name,
			Permissions: perms,
			CreatedAt:   role.CreatedAt,
			UpdatedAt:   role.UpdatedAt,
		})
	}

	totalPages := total / req.Limit
	if total%req.Limit > 0 {
		totalPages++
	}

	return domain.PaginationResponse{
		Data: roleResponses,
		Meta: domain.PaginationMeta{
			Page:         req.Page,
			Limit:        req.Limit,
			TotalRecords: total,
			TotalPages:   totalPages,
		},
	}, nil
}

func (s *rbacService) UpdateRole(ctx context.Context, id int, req domain.RoleRequest) (*domain.RoleResponse, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, validation.FormatValidationError(err)
	}

	role, err := s.rbacRepo.UpdateRole(ctx, id, req.Name)
	if err != nil {
		return nil, err
	}

	if err := s.rbacRepo.AssignPermissionToRole(ctx, role.ID, req.PermissionIDs); err != nil {
		return nil, err
	}

	perms, _ := s.rbacRepo.GetRolePermissions(ctx, role.ID)

	return &domain.RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Permissions: perms,
	}, nil
}

func (s *rbacService) DeleteRole(ctx context.Context, id int) error {
	return s.rbacRepo.DeleteRole(ctx, id)
}

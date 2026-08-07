package domain

import (
	"context"
	"time"
)

type Role struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Permission struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RBACRepository interface {
	HasPermission(ctx context.Context, userID int, permissionName string) (bool, error)
	HasRole(ctx context.Context, userID int, roleName string) (bool, error)
	GetPermissionByName(ctx context.Context, name string) (*Permission, error)
	CreatePermission(ctx context.Context, name string) (*Permission, error)
	GetAllPermissions(ctx context.Context, req PaginationRequest) ([]Permission, int, error)
	UpdatePermission(ctx context.Context, id int, name string) (*Permission, error)
	DeletePermission(ctx context.Context, id int) error

	CreateRole(ctx context.Context, name, description string) (*Role, error)
	GetAllRoles(ctx context.Context, req PaginationRequest) ([]Role, int, error)
	UpdateRole(ctx context.Context, id int, name, description string) (*Role, error)
	DeleteRole(ctx context.Context, id int) error
	AssignPermissionToRole(ctx context.Context, roleID int, permissionIDs []int) error
	GetRolePermissions(ctx context.Context, roleID int) ([]Permission, error)
	CheckRoleNameExists(ctx context.Context, name string) (bool, error)
	GetUserPermissions(ctx context.Context, userID int) ([]string, error)
}

type RBACService interface {
	CreatePermission(ctx context.Context, req PermissionRequest) (*Permission, error)
	GetAllPermissions(ctx context.Context, req PaginationRequest) (PaginationResponse, error)
	UpdatePermission(ctx context.Context, id int, req PermissionRequest) (*Permission, error)
	DeletePermission(ctx context.Context, id int) error

	CreateRole(ctx context.Context, req RoleRequest) (*RoleResponse, error)
	GetAllRoles(ctx context.Context, req PaginationRequest) (PaginationResponse, error)
	UpdateRole(ctx context.Context, id int, req RoleRequest) (*RoleResponse, error)
	DeleteRole(ctx context.Context, id int) error
	GetUserPermissions(ctx context.Context, userID int) ([]string, error)
}

type PermissionRequest struct {
	Name string `json:"name" validate:"required"`
}

type RoleRequest struct {
	Name          string `json:"name" validate:"required"`
	Description   string `json:"description"`
	PermissionIDs []int  `json:"permission_ids" validate:"required,min=1"`
}

type RoleResponse struct {
	ID          int          `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

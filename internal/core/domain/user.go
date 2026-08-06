package domain

import (
	"context"
	"time"

)

// User represents the users table in the database
type User struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	Password     string    `json:"-"` // Never send password in JSON response
	Photo        *string   `json:"photo,omitempty"`
	AuthProvider string    `json:"auth_provider,omitempty"`
	ProviderID   *string   `json:"provider_id,omitempty"`
	Roles        []Role    `json:"roles,omitempty"`
	Permissions  []string  `json:"permissions,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// AddUserRequest represents the incoming request
type AddUserRequest struct {
	Name     string  `json:"name" validate:"required"`
	Email    string  `json:"email" validate:"required,email"`
	Password string  `json:"password" validate:"required"`
	Photo    *string `json:"photo,omitempty"`
	RoleIDs  []int   `json:"role_ids,omitempty"`
}

type UpdateUserRequest struct {
	Name     string  `json:"name" validate:"required"`
	Email    string  `json:"email" validate:"required,email"`
	Password *string `json:"password,omitempty"`
	Photo    *string `json:"photo,omitempty"`
	RoleIDs  []int   `json:"role_ids,omitempty"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id int) (*User, error)
	GetAll(ctx context.Context, request PaginationRequest) ([]*User, int, error)
	Update(ctx context.Context, id int, user *User)error
	Delete(ctx context.Context, id int) error
	AssignRolesToUser(ctx context.Context, userID int, roleIDs []int) error
	GetUserRoles(ctx context.Context, userID int) ([]Role, error)
	GetUsersRoles(ctx context.Context, userIDs []int) (map[int][]Role, error)
}

// UserService defines the business logic operations for User
type UserService interface {
	AddUser(ctx context.Context, request AddUserRequest) (*User, error)
	GetByID(ctx context.Context, id int) (*User, error)
	GetAllUsers(ctx context.Context, request PaginationRequest) (PaginationResponse, error)
	UpdateUser(ctx context.Context, id int, request UpdateUserRequest)(*User, error)
	DeleteUser(ctx context.Context, id int) error
}

type OAuthLoginRequest struct {
	Email      string
	Name       string
	Photo      string
	Provider   string
	ProviderID string
}

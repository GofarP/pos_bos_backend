package domain

import (
	"context"
	"time"
)

type Category struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CategoryRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

type CategoryRepository interface {
	Create(ctx context.Context, name, description string) (*Category, error)
	GetAll(ctx context.Context, req PaginationRequest) ([]Category, int, error)
	GetByID(ctx context.Context, id int) (*Category, error)
	Update(ctx context.Context, id int, name, description string) (*Category, error)
	Delete(ctx context.Context, id int) error
	CheckNameExists(ctx context.Context, name string) (bool, error)
}

type CategoryService interface {
	CreateCategory(ctx context.Context, req CategoryRequest) (*Category, error)
	GetAllCategories(ctx context.Context, req PaginationRequest) (PaginationResponse, error)
	GetCategoryByID(ctx context.Context, id int) (*Category, error)
	UpdateCategory(ctx context.Context, id int, req CategoryRequest) (*Category, error)
	DeleteCategory(ctx context.Context, id int) error
}

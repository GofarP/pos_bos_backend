package domain

import (
	"context"
	"time"
)

type Product struct {
	ID          int       `json:"id"`
	CategoryID  *int      `json:"category_id"`
	Category    *Category `json:"category,omitempty"`
	SKU         *string   `json:"sku"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	Price       int64   `json:"price"`
	Stock       int       `json:"stock"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProductRequest struct {
	CategoryID  *int    `json:"category_id"`
	SKU         *string `json:"sku" validate:"omitempty,max=100"`
	Name        string  `json:"name" validate:"required,max=255"`
	Description *string `json:"description"`
	Price       int64 `json:"price" validate:"required,min=0"`
	Stock       int     `json:"stock" validate:"min=0"`
}

type ProductRepository interface {
	CheckSKUExists(ctx context.Context, sku string) (bool, error)
	Create(ctx context.Context, req ProductRequest) (*Product, error)
	GetAll(ctx context.Context, req PaginationRequest) ([]Product, int, error)
	GetByID(ctx context.Context, id int) (*Product, error)
	Update(ctx context.Context, id int, req ProductRequest) (*Product, error)
	Delete(ctx context.Context, id int) error
}

type ProductService interface {
	CreateProduct(ctx context.Context, req ProductRequest) (*Product, error)
	GetAllProducts(ctx context.Context, req PaginationRequest) (PaginationResponse, error)
	GetProductByID(ctx context.Context, id int) (*Product, error)
	UpdateProduct(ctx context.Context, id int, req ProductRequest) (*Product, error)
	DeleteProduct(ctx context.Context, id int) error
}

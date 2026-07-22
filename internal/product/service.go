package product

import (
	"context"
	"errors"
	"pos_bos/internal/core/domain"
	"pos_bos/pkg/validation"

	"github.com/go-playground/validator/v10"
)

type productService struct {
	repo     domain.ProductRepository
	catRepo  domain.CategoryRepository
	validate *validator.Validate
}

func NewProductService(repo domain.ProductRepository, catRepo domain.CategoryRepository) domain.ProductService {
	return &productService{
		repo:     repo,
		catRepo:  catRepo,
		validate: validator.New(),
	}
}

func (s *productService) CreateProduct(ctx context.Context, req domain.ProductRequest) (*domain.Product, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, validation.FormatValidationError(err)
	}

	if req.SKU != nil {
		exists, err := s.repo.CheckSKUExists(ctx, *req.SKU)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("sku already exists")
		}
	}

	if req.CategoryID != nil {
		cat, err := s.catRepo.GetByID(ctx, *req.CategoryID)
		if err != nil || cat == nil {
			return nil, errors.New("category not found")
		}
	}

	return s.repo.Create(ctx, req)
}

func (s *productService) GetAllProducts(ctx context.Context, req domain.PaginationRequest) (domain.PaginationResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}

	products, total, err := s.repo.GetAll(ctx, req)
	if err != nil {
		return domain.PaginationResponse{}, err
	}

	totalPages := total / req.Limit
	if total%req.Limit > 0 {
		totalPages++
	}

	return domain.PaginationResponse{
		Data: products,
		Meta: domain.PaginationMeta{
			Page:         req.Page,
			Limit:        req.Limit,
			TotalRecords: total,
			TotalPages:   totalPages,
		},
	}, nil
}

func (s *productService) GetProductByID(ctx context.Context, id int) (*domain.Product, error) {
	p, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, errors.New("product not found")
	}
	return p, nil
}

func (s *productService) UpdateProduct(ctx context.Context, id int, req domain.ProductRequest) (*domain.Product, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, validation.FormatValidationError(err)
	}

	existing, err := s.repo.GetByID(ctx, id)
	if err != nil || existing == nil {
		return nil, errors.New("product not found")
	}

	if req.SKU != nil && (existing.SKU == nil || *existing.SKU != *req.SKU) {
		exists, _ := s.repo.CheckSKUExists(ctx, *req.SKU)
		if exists {
			return nil, errors.New("sku already exists")
		}
	}

	if req.CategoryID != nil && (existing.CategoryID == nil || *existing.CategoryID != *req.CategoryID) {
		cat, err := s.catRepo.GetByID(ctx, *req.CategoryID)
		if err != nil || cat == nil {
			return nil, errors.New("category not found")
		}
	}

	return s.repo.Update(ctx, id, req)
}

func (s *productService) DeleteProduct(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

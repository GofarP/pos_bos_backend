package category

import (
	"context"
	"errors"
	"pos_bos/internal/core/domain"
	"pos_bos/pkg/validation"

	"github.com/go-playground/validator/v10"
)

type categoryService struct {
	repo     domain.CategoryRepository
	validate *validator.Validate
}

func NewCategoryService(repo domain.CategoryRepository) domain.CategoryService {
	return &categoryService{
		repo:     repo,
		validate: validator.New(),
	}
}

func (s *categoryService) CreateCategory(ctx context.Context, req domain.CategoryRequest) (*domain.Category, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, validation.FormatValidationError(err)
	}
	exists, err := s.repo.CheckNameExists(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("category name already exists")
	}

	return s.repo.Create(ctx, req.Name, req.Description)
}

func (s *categoryService) GetAllCategories(ctx context.Context, req domain.PaginationRequest) (domain.PaginationResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}

	if req.Limit <= 0 {
		req.Limit = 10
	}

	categories, total, err := s.repo.GetAll(ctx, req)
	if err != nil {
		return domain.PaginationResponse{}, err
	}

	totalPages := total / req.Limit

	if total%req.Limit > 0 {
		totalPages++
	}

	return domain.PaginationResponse{
		Data: categories,
		Meta: domain.PaginationMeta{
			Page:         req.Page,
			Limit:        req.Limit,
			TotalRecords: total,
			TotalPages:   totalPages,
		},
	}, nil
}

func (s *categoryService) GetCategoryByID(ctx context.Context, id int) (*domain.Category, error) {
	cat, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if cat == nil {
		return nil, errors.New("category not found")
	}
	return cat, nil
}

func (s *categoryService) UpdateCategory(ctx context.Context, id int, req domain.CategoryRequest) (*domain.Category, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, validation.FormatValidationError(err)
	}
	// Cek apakah ada yang namanya sama (tapi ID-nya berbeda)
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil || existing == nil {
		return nil, errors.New("category not found")
	}
	if existing.Name != req.Name {
		exists, _ := s.repo.CheckNameExists(ctx, req.Name)
		if exists {
			return nil, errors.New("category name already exists")
		}
	}
	return s.repo.Update(ctx, id, req.Name, req.Description)
}

func (s *categoryService) DeleteCategory(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

package category

import (
	"context"
	"database/sql"
	"pos_bos/internal/core/domain"
)

type mysqlCategoryRepository struct {
	db *sql.DB
}

func NewCategoryRepository(db *sql.DB) domain.CategoryRepository {
	return &mysqlCategoryRepository{db: db}
}

func (r *mysqlCategoryRepository) CheckNameExists(ctx context.Context, name string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM categories WHERE name = ?)`
	err := r.db.QueryRowContext(ctx, query, name).Scan(&exists)
	return exists, err
}

func (r *mysqlCategoryRepository) Create(ctx context.Context, name, description string) (*domain.Category, error) {
	query := `INSERT INTO categories (name, description) VALUES (?, ?)`

	res, err := r.db.ExecContext(ctx, query, name, description)

	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &domain.Category{
		ID:          int(id),
		Name:        name,
		Description: description,
	}, nil
}

func (r *mysqlCategoryRepository) GetAll(ctx context.Context, req domain.PaginationRequest) ([]domain.Category, int, error) {
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(id) FROM categories`).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (req.Page - 1) * req.Limit

	if offset < 0 {
		offset = 0
	}

	query := `SELECT id, name, description, created_at, updated_at from categories order by id desc limit ? offset ?`
	rows, err := r.db.QueryContext(ctx, query, req.Limit, offset)

	if err != nil {
		return nil, 0, err
	}

	defer rows.Close()

	categories := []domain.Category{}
	for rows.Next() {
		var cat domain.Category
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Description, &cat.CreatedAt, &cat.UpdatedAt); err != nil {
			return nil, 0, err
		}
		categories = append(categories, cat)
	}
	return categories, total, nil

}

func (r *mysqlCategoryRepository) GetByID(ctx context.Context, id int) (*domain.Category, error) {
	query := `SELECT id, name, description, created_at, updated_at from categories where id=?`

	var cat domain.Category
	err := r.db.QueryRowContext(ctx, query, id).Scan(&cat.ID, &cat.Name, &cat.Description, &cat.CreatedAt, &cat.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &cat, nil
}

func (r *mysqlCategoryRepository) Update(ctx context.Context, id int, name, description string) (*domain.Category, error) {
	query := `UPDATE categories SET name = ?, description = ? WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, name, description, id)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *mysqlCategoryRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM categories WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

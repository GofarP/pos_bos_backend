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
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	err = stmt.QueryRowContext(ctx, name).Scan(&exists)
	return exists, err
}

func (r *mysqlCategoryRepository) Create(ctx context.Context, name, description string) (*domain.Category, error) {
	query := `INSERT INTO categories (name, description) VALUES (?, ?)`
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	res, err := stmt.ExecContext(ctx, name, description)
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
	countQuery := `SELECT COUNT(id) FROM categories WHERE 1=1`
	var args []interface{}

	if req.Search != "" {
		searchPattern := "%" + req.Search + "%"
		countQuery += ` AND (name LIKE ? OR description LIKE ?)`
		args = append(args, searchPattern, searchPattern)
	}

	var total int
	stmtCount, err := r.db.PrepareContext(ctx, countQuery)
	if err != nil {
		return nil, 0, err
	}
	defer stmtCount.Close()

	if err := stmtCount.QueryRowContext(ctx, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (req.Page - 1) * req.Limit
	if offset < 0 {
		offset = 0
	}

	query := `SELECT id, name, description, created_at, updated_at FROM categories WHERE 1=1`
	var queryArgs []interface{}
	if req.Search != "" {
		searchPattern := "%" + req.Search + "%"
		query += ` AND (name LIKE ? OR description LIKE ?)`
		queryArgs = append(queryArgs, searchPattern, searchPattern)
	}
	query += ` ORDER BY id DESC LIMIT ? OFFSET ?`
	queryArgs = append(queryArgs, req.Limit, offset)

	stmtSelect, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	defer stmtSelect.Close()

	rows, err := stmtSelect.QueryContext(ctx, queryArgs...)
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
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var cat domain.Category
	err = stmt.QueryRowContext(ctx, id).Scan(&cat.ID, &cat.Name, &cat.Description, &cat.CreatedAt, &cat.UpdatedAt)
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
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, name, description, id)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *mysqlCategoryRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM categories WHERE id = ?`
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, id)
	return err
}

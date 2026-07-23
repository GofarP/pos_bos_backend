package product

import (
	"context"
	"database/sql"
	"pos_bos/internal/core/domain"
)

type mysqlProductRepository struct {
	db *sql.DB
}

func NewProductRepository(db *sql.DB) domain.ProductRepository {
	return &mysqlProductRepository{db: db}
}

func (r *mysqlProductRepository) CheckSKUExists(ctx context.Context, sku string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM products WHERE sku = ?)`
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	err = stmt.QueryRowContext(ctx, sku).Scan(&exists)
	return exists, err
}

func (r *mysqlProductRepository) Create(ctx context.Context, req domain.ProductRequest) (*domain.Product, error) {
	query := `INSERT INTO products (category_id, sku, name, description, price, stock) VALUES (?, ?, ?, ?, ?, ?)`
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	res, err := stmt.ExecContext(ctx, req.CategoryID, req.SKU, req.Name, req.Description, req.Price, req.Stock)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, int(id))
}

func (r *mysqlProductRepository) GetAll(ctx context.Context, req domain.PaginationRequest) ([]domain.Product, int, error) {
	countQuery := `SELECT COUNT(id) FROM products WHERE 1=1`
	var args []interface{}

	if req.Search != "" {
		searchPattern := "%" + req.Search + "%"
		countQuery += ` AND (name LIKE ? OR sku LIKE ? OR description LIKE ?)`
		args = append(args, searchPattern, searchPattern, searchPattern)
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

	// Join dengan tabel categories untuk mendapatkan nama kategorinya juga
	query := `
		SELECT p.id, p.category_id, p.sku, p.name, p.description, p.price, p.stock, p.created_at, p.updated_at,
		       c.id, c.name, c.description, c.created_at, c.updated_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE 1=1
	`
	var queryArgs []interface{}
	if req.Search != "" {
		searchPattern := "%" + req.Search + "%"
		query += ` AND (p.name LIKE ? OR p.sku LIKE ? OR p.description LIKE ?)`
		queryArgs = append(queryArgs, searchPattern, searchPattern, searchPattern)
	}
	query += ` ORDER BY p.id DESC LIMIT ? OFFSET ?`
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

	products := []domain.Product{}
	for rows.Next() {
		var p domain.Product
		var c domain.Category
		var cID sql.NullInt64
		var cName, cDesc sql.NullString
		var cCreated, cUpdated sql.NullTime

		err := rows.Scan(
			&p.ID, &p.CategoryID, &p.SKU, &p.Name, &p.Description, &p.Price, &p.Stock, &p.CreatedAt, &p.UpdatedAt,
			&cID, &cName, &cDesc, &cCreated, &cUpdated,
		)
		if err != nil {
			return nil, 0, err
		}

		if cID.Valid {
			c.ID = int(cID.Int64)
			c.Name = cName.String
			if cDesc.Valid {
				c.Description = cDesc.String
			}
			c.CreatedAt = cCreated.Time
			c.UpdatedAt = cUpdated.Time
			p.Category = &c
		}

		products = append(products, p)
	}
	return products, total, nil
}

func (r *mysqlProductRepository) GetByID(ctx context.Context, id int) (*domain.Product, error) {
	query := `
		SELECT p.id, p.category_id, p.sku, p.name, p.description, p.price, p.stock, p.created_at, p.updated_at,
		       c.id, c.name, c.description, c.created_at, c.updated_at
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		WHERE p.id = ?
	`
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, id)

	var p domain.Product
	var c domain.Category
	var cID sql.NullInt64
	var cName, cDesc sql.NullString
	var cCreated, cUpdated sql.NullTime

	err = row.Scan(
		&p.ID, &p.CategoryID, &p.SKU, &p.Name, &p.Description, &p.Price, &p.Stock, &p.CreatedAt, &p.UpdatedAt,
		&cID, &cName, &cDesc, &cCreated, &cUpdated,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if cID.Valid {
		c.ID = int(cID.Int64)
		c.Name = cName.String
		if cDesc.Valid {
			c.Description = cDesc.String
		}
		c.CreatedAt = cCreated.Time
		c.UpdatedAt = cUpdated.Time
		p.Category = &c
	}
	return &p, nil
}

func (r *mysqlProductRepository) Update(ctx context.Context, id int, req domain.ProductRequest) (*domain.Product, error) {
	query := `UPDATE products SET category_id=?, sku=?, name=?, description=?, price=?, stock=? WHERE id=?`
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, req.CategoryID, req.SKU, req.Name, req.Description, req.Price, req.Stock, id)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *mysqlProductRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM products WHERE id = ?`
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, id)
	return err
}

package user

import (
	"context"
	"database/sql"
	"errors"

	"pos_bos/internal/core/domain"
)

type mysqlUserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new MySQL user repository
func NewUserRepository(db *sql.DB) domain.UserRepository {
	return &mysqlUserRepository{
		db: db,
	}
}

func (repository *mysqlUserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (name, email, password) VALUES (?, ?, ?)`

	result, err := repository.db.ExecContext(ctx, query, user.Name, user.Email, user.Password)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	user.ID = int(id)
	return nil
}

func (repository *mysqlUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, name, email, password, photo, created_at, updated_at FROM users WHERE email = ?`

	var user domain.User
	err := repository.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Photo,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Return nil if user not found, without error
		}
		return nil, err
	}

	return &user, nil
}

func (r *mysqlUserRepository) GetByID(ctx context.Context, id int) (*domain.User, error) {
	query := `SELECT id, name, email, password, photo, created_at, updated_at FROM users WHERE id = ?`
	row := r.db.QueryRowContext(ctx, query, id)

	var user domain.User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Photo, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil if user not found
		}
		return nil, err
	}

	return &user, nil
}

func (repository *mysqlUserRepository) GetAll(ctx context.Context, request domain.PaginationRequest) ([]*domain.User, int, error) {
	countQuery := "SELECT COUNT(id) FROM users WHERE 1=1"
	selectQuery := "SELECT id, name, email, password, photo, created_at, updated_at FROM users WHERE 1=1"
	var args []interface{}

	if request.Search != "" {
		searchPattern := "%" + request.Search + "%"
		countQuery += " AND (name LIKE ? OR email LIKE ?)"
		selectQuery += " AND (name LIKE ? OR email LIKE ?)"
		args = append(args, searchPattern, searchPattern)
	}

	// Hitung total data
	var total int
	err := repository.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Hitung offset
	offset := (request.Page - 1) * request.Limit
	selectQuery += " LIMIT ? OFFSET ?"

	// Tambahkan limit dan offset ke argument
	args = append(args, request.Limit, offset)

	rows, err := repository.db.QueryContext(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := make([]*domain.User, 0)
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Photo, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (repository *mysqlUserRepository) Update(ctx context.Context, id int, user *domain.User)error{
	query:=`UPDATE users SET name=?, email=?, photo=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`

	_,err:=repository.db.ExecContext(ctx, query, user.Name, user.Email, user.Password, user.Photo, id)

	return err
}

func (repository *mysqlUserRepository) Delete(ctx context.Context, id int) error {
	query := "DELETE FROM users WHERE id = ?"
	_, err := repository.db.ExecContext(ctx, query, id)
	return err
}

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
	query := `INSERT INTO users (name, email, password, photo) VALUES (?, ?, ?, ?)`
	stmt, err := repository.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, user.Name, user.Email, user.Password, user.Photo)
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
	stmt, err := repository.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var user domain.User
	err = stmt.QueryRowContext(ctx, email).Scan(
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
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, id)

	var user domain.User
	err = row.Scan(&user.ID, &user.Name, &user.Email, &user.Password, &user.Photo, &user.CreatedAt, &user.UpdatedAt)
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
	stmtCount, err := repository.db.PrepareContext(ctx, countQuery)
	if err != nil {
		return nil, 0, err
	}
	defer stmtCount.Close()

	err = stmtCount.QueryRowContext(ctx, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Hitung offset
	offset := (request.Page - 1) * request.Limit
	selectQuery += " LIMIT ? OFFSET ?"

	// Tambahkan limit dan offset ke argument
	selectArgs := append(args, request.Limit, offset)

	stmtSelect, err := repository.db.PrepareContext(ctx, selectQuery)
	if err != nil {
		return nil, 0, err
	}
	defer stmtSelect.Close()

	rows, err := stmtSelect.QueryContext(ctx, selectArgs...)
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

func (repository *mysqlUserRepository) Update(ctx context.Context, id int, user *domain.User) error {
	query := `UPDATE users SET name=?, email=?, password=?, photo=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`
	stmt, err := repository.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, user.Name, user.Email, user.Password, user.Photo, id)
	return err
}

func (repository *mysqlUserRepository) Delete(ctx context.Context, id int) error {
	query := "DELETE FROM users WHERE id = ?"
	stmt, err := repository.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, id)
	return err
}

func (r *mysqlUserRepository) AssignRolesToUser(ctx context.Context, userID int, roleIDs []int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	deleteQuery := `DELETE FROM user_roles WHERE user_id = ?`
	if _, err := tx.ExecContext(ctx, deleteQuery, userID); err != nil {
		return err
	}

	insertQuery := `INSERT INTO user_roles (user_id, role_id) VALUES (?, ?)`
	stmtInsert, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return err
	}
	defer stmtInsert.Close()

	for _, roleID := range roleIDs {
		if _, err := stmtInsert.ExecContext(ctx, userID, roleID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *mysqlUserRepository) GetUserRoles(ctx context.Context, userID int) ([]domain.Role, error) {
	query := `
		SELECT r.id, r.name, r.created_at, r.updated_at 
		FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = ?
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []domain.Role
	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func (r *mysqlUserRepository) GetUsersRoles(ctx context.Context, userIDs []int) (map[int][]domain.Role, error) {
	if len(userIDs) == 0 {
		return make(map[int][]domain.Role), nil
	}

	args := make([]interface{}, len(userIDs))
	placeholders := ""
	for i, id := range userIDs {
		args[i] = id
		if i > 0 {
			placeholders += ", "
		}
		placeholders += "?"
	}

	query := `
		SELECT ur.user_id, r.id, r.name, r.created_at, r.updated_at 
		FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id IN (` + placeholders + `)
	`
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rolesMap := make(map[int][]domain.Role)
	for rows.Next() {
		var userID int
		var role domain.Role
		if err := rows.Scan(&userID, &role.ID, &role.Name, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, err
		}
		rolesMap[userID] = append(rolesMap[userID], role)
	}
	return rolesMap, nil
}

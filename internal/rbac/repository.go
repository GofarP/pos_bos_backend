package rbac

import (
	"context"
	"database/sql"
	"pos_bos/internal/core/domain"
)

type rbacRepository struct {
	db *sql.DB
}

func NewRBACRepository(db *sql.DB) domain.RBACRepository {
	return &rbacRepository{db: db}
}

func (r *rbacRepository) HasPermission(ctx context.Context, userID int, permissionName string) (bool, error) {
	query := `
		SELECT 1 FROM permissions p
		LEFT JOIN user_permissions up ON p.id = up.permission_id AND up.user_id = ?
		LEFT JOIN role_permissions rp ON p.id = rp.permission_id
		LEFT JOIN user_roles ur ON rp.role_id = ur.role_id AND ur.user_id = ?
		WHERE p.name = ? AND (up.user_id IS NOT NULL OR ur.user_id IS NOT NULL)
		LIMIT 1
	`
	var exists int
	err := r.db.QueryRowContext(ctx, query, userID, userID, permissionName).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return exists == 1, nil
}

func (r *rbacRepository) HasRole(ctx context.Context, userID int, roleName string) (bool, error) {
	query := `
		SELECT 1 FROM roles r
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = ? AND r.name = ?
		LIMIT 1
	`
	var exists int
	err := r.db.QueryRowContext(ctx, query, userID, roleName).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return exists == 1, nil
}

func (r *rbacRepository) GetPermissionByName(ctx context.Context, name string) (*domain.Permission, error) {
	query := `SELECT id, name, created_at, updated_at FROM permissions WHERE name = ?`
	var p domain.Permission
	err := r.db.QueryRowContext(ctx, query, name).Scan(&p.ID, &p.Name, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, err
	}
	return &p, nil
}

func (r *rbacRepository) CreatePermission(ctx context.Context, name string) (*domain.Permission, error) {
	query := `INSERT INTO permissions (name) VALUES (?)`
	result, err := r.db.ExecContext(ctx, query, name)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return &domain.Permission{
		ID:   int(id),
		Name: name,
	}, nil
}

func (r *rbacRepository) GetAllPermissions(ctx context.Context, req domain.PaginationRequest) ([]domain.Permission, int, error) {
	var total int
	countQuery := `SELECT COUNT(*) FROM permissions`
	args := []interface{}{}

	if req.Search != "" {
		countQuery += ` WHERE name LIKE ?`
		args = append(args, "%"+req.Search+"%")
	}

	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `SELECT id, name, created_at, updated_at FROM permissions`
	if req.Search != "" {
		query += ` WHERE name LIKE ?`
	}
	query += ` ORDER BY id DESC LIMIT ? OFFSET ?`

	offset := (req.Page - 1) * req.Limit
	args = append(args, req.Limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	permissions := []domain.Permission{}
	for rows.Next() {
		var p domain.Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, err
		}
		permissions = append(permissions, p)
	}
	return permissions, total, nil
}

func (r *rbacRepository) UpdatePermission(ctx context.Context, id int, name string) (*domain.Permission, error) {
	query := `UPDATE permissions SET name = ?, updated_at = NOW() WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, name, id)
	if err != nil {
		return nil, err
	}

	return &domain.Permission{
		ID:   id,
		Name: name,
	}, nil
}

func (r *rbacRepository) DeletePermission(ctx context.Context, id int) error {
	query := `DELETE FROM permissions WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}


func (r *rbacRepository) CheckRoleNameExists(ctx context.Context, name string) (bool, error) {
   var exists bool
   query:=`select exists(select 1 from roles where name=?)`
   err:=r.db.QueryRowContext(ctx, query, name).Scan(&exists)
   return exists, err
}

func (r *rbacRepository) CreateRole(ctx context.Context, name string) (*domain.Role, error) {
	query := `INSERT INTO roles (name) VALUES (?)`
	res, err := r.db.ExecContext(ctx, query, name)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &domain.Role{
		ID:   int(id),
		Name: name,
	}, nil
}

func (r *rbacRepository) GetAllRoles(ctx context.Context, req domain.PaginationRequest) ([]domain.Role, int, error) {
	var total int
	countQuery := `SELECT COUNT(id) FROM roles`
	if err := r.db.QueryRowContext(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (req.Page - 1) * req.Limit
	if offset < 0 {
		offset = 0
	}
	query := `SELECT id, name, created_at, updated_at FROM roles ORDER BY id DESC LIMIT ? OFFSET ?`
	rows, err := r.db.QueryContext(ctx, query, req.Limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var roles []domain.Role
	for rows.Next() {
		var role domain.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.CreatedAt, &role.UpdatedAt); err != nil {
			return nil, 0, err
		}
		roles = append(roles, role)
	}

	return roles, total, nil
}

func (r *rbacRepository) UpdateRole(ctx context.Context, id int, name string) (*domain.Role, error) {
	query := `UPDATE roles SET name = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, name, id)
	if err != nil {
		return nil, err
	}

	return &domain.Role{
		ID:   id,
		Name: name,
	}, nil
}

func (r *rbacRepository) DeleteRole(ctx context.Context, id int) error {
	query := `DELETE FROM roles WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *rbacRepository) AssignPermissionToRole(ctx context.Context, roleID int, permissionIDs []int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM role_permissions WHERE role_id = ?`, roleID)
	if err != nil {
		tx.Rollback()
		return err
	}

	insertQuery := `INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)`
	for _, permID := range permissionIDs {
		_, err = tx.ExecContext(ctx, insertQuery, roleID, permID)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *rbacRepository) GetRolePermissions(ctx context.Context, roleID int) ([]domain.Permission, error) {
	query := `
		SELECT p.id, p.name, p.created_at, p.updated_at 
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = ?
	`
	rows, err := r.db.QueryContext(ctx, query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []domain.Permission
	for rows.Next() {
		var p domain.Permission
		if err := rows.Scan(&p.ID, &p.Name, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		perms = append(perms, p)
	}
	return perms, nil
}
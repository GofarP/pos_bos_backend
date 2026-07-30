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
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	var exists int
	err = stmt.QueryRowContext(ctx, userID, userID, permissionName).Scan(&exists)
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
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	var exists int
	err = stmt.QueryRowContext(ctx, userID, roleName).Scan(&exists)
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
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var p domain.Permission
	err = stmt.QueryRowContext(ctx, name).Scan(&p.ID, &p.Name, &p.CreatedAt, &p.UpdatedAt)
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
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, name)
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
	var args []interface{}

	if req.Search != "" {
		countQuery += ` WHERE name LIKE ?`
		args = append(args, "%"+req.Search+"%")
	}

	stmtCount, err := r.db.PrepareContext(ctx, countQuery)
	if err != nil {
		return nil, 0, err
	}
	defer stmtCount.Close()

	err = stmtCount.QueryRowContext(ctx, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `SELECT id, name, created_at, updated_at FROM permissions`
	var queryArgs []interface{}
	if req.Search != "" {
		query += ` WHERE name LIKE ?`
		queryArgs = append(queryArgs, "%"+req.Search+"%")
	}
	query += ` ORDER BY id DESC LIMIT ? OFFSET ?`

	offset := (req.Page - 1) * req.Limit
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
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, name, id)
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
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, id)
	return err
}

func (r *rbacRepository) CheckRoleNameExists(ctx context.Context, name string) (bool, error) {
	var exists bool
	query := `select exists(select 1 from roles where name=?)`
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	err = stmt.QueryRowContext(ctx, name).Scan(&exists)
	return exists, err
}

func (r *rbacRepository) CreateRole(ctx context.Context, name string) (*domain.Role, error) {
	query := `INSERT INTO roles (name) VALUES (?)`
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	res, err := stmt.ExecContext(ctx, name)
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
	stmtCount, err := r.db.PrepareContext(ctx, countQuery)
	if err != nil {
		return nil, 0, err
	}
	defer stmtCount.Close()

	if err := stmtCount.QueryRowContext(ctx).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (req.Page - 1) * req.Limit
	if offset < 0 {
		offset = 0
	}
	query := `SELECT id, name, created_at, updated_at FROM roles ORDER BY id DESC LIMIT ? OFFSET ?`
	stmtSelect, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	defer stmtSelect.Close()

	rows, err := stmtSelect.QueryContext(ctx, req.Limit, offset)
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
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, name, id)
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
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, id)
	return err
}

func (r *rbacRepository) AssignPermissionToRole(ctx context.Context, roleID int, permissionIDs []int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	deleteQuery := `DELETE FROM role_permissions WHERE role_id = ?`
	stmtDelete, err := tx.PrepareContext(ctx, deleteQuery)
	if err != nil {
		return err
	}
	defer stmtDelete.Close()

	_, err = stmtDelete.ExecContext(ctx, roleID)
	if err != nil {
		return err
	}

	insertQuery := `INSERT INTO role_permissions (role_id, permission_id) VALUES (?, ?)`
	stmtInsert, err := tx.PrepareContext(ctx, insertQuery)
	if err != nil {
		return err
	}
	defer stmtInsert.Close()

	for _, permID := range permissionIDs {
		_, err = stmtInsert.ExecContext(ctx, roleID, permID)
		if err != nil {
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
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, roleID)
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

func (r *rbacRepository) GetUserPermissions(ctx context.Context, userID int) ([]string, error) {
	query := `
		SELECT DISTINCT p.name
		FROM permissions p
		LEFT JOIN role_permissions rp ON p.id = rp.permission_id
		LEFT JOIN user_roles ur ON rp.role_id = ur.role_id
		LEFT JOIN user_permissions up ON p.id = up.permission_id
		WHERE up.user_id = ? OR ur.user_id = ?
		ORDER BY p.name ASC
	`
	rows, err := r.db.QueryContext(ctx, query, userID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		permissions = append(permissions, name)
	}
	if permissions == nil {
		permissions = []string{}
	}
	return permissions, nil
}
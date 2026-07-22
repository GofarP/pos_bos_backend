package auth

import (
	"context"
	"database/sql"
	"errors"
	"pos_bos/internal/core/domain"
	"time"
)

type mysqlAuthRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) domain.AuthRepository {
	return &mysqlAuthRepository{
		db: db,
	}
}

func (r *mysqlAuthRepository) StoreRefreshToken(ctx context.Context, userID int, token string, expiresAt time.Time) error {
	query := `INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES (?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query, userID, token, expiresAt)
	return err
}

func (r *mysqlAuthRepository) GetRefreshToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	query := `SELECT id, user_id, token, expires_at, revoked_at, created_at FROM refresh_tokens WHERE token = ?`
	row := r.db.QueryRowContext(ctx, query, token)
	
	var rt domain.RefreshToken
	err := row.Scan(&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt, &rt.RevokedAt, &rt.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &rt, nil
}

func (r *mysqlAuthRepository) RevokeRefreshToken(ctx context.Context, token string) error {
	query := `UPDATE refresh_tokens SET revoked_at = CURRENT_TIMESTAMP WHERE token = ? AND revoked_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, token)
	if err != nil {
		return err
	}
	
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	
	if rows == 0 {
		return errors.New("invalid or already revoked refresh token")
	}
	
	return nil
}

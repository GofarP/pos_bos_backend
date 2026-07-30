package domain
import (
	"context"
	"time"
)

type LoginRequest struct{
	Email string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct{
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RefreshToken struct {
	ID        int
	UserID    int
	Token     string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

type AuthRepository interface {
	StoreRefreshToken(ctx context.Context, userID int, token string, expiresAt time.Time) error
	GetRefreshToken(ctx context.Context, token string) (*RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, token string) error
}

type AuthService interface {
	Login(ctx context.Context, request LoginRequest) (LoginResponse, error)
	Logout(ctx context.Context, request LogoutRequest) error
	Refresh(ctx context.Context, request RefreshRequest) (LoginResponse, error)
	GetMe(ctx context.Context, userID int) (*User, error)
}
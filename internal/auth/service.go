package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"os"
	"pos_bos/internal/core/domain"
	"pos_bos/pkg/validation"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
)

type authService struct {
	userRepo domain.UserRepository
	authRepo domain.AuthRepository
	validate *validator.Validate
}

func NewAuthService(userRepo domain.UserRepository, authRepo domain.AuthRepository) domain.AuthService {
	return &authService{
		userRepo: userRepo,
		authRepo: authRepo,
		validate: validator.New(),
	}
}

func (service *authService) Login(ctx context.Context, request domain.LoginRequest) (domain.LoginResponse, error) {
	if err := service.validate.Struct(request); err != nil {
		return domain.LoginResponse{}, validation.FormatValidationError(err)
	}

	user, err := service.userRepo.GetByEmail(ctx, request.Email)
	if err != nil || user == nil {
		return domain.LoginResponse{}, ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password))
	if err != nil {
		return domain.LoginResponse{}, ErrInvalidCredentials
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Minute * 15).Unix(), // Berlaku 15 Menit
	})

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "rahasia_default"
	}

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return domain.LoginResponse{}, err
	}

	// Generate Refresh Token
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return domain.LoginResponse{}, err
	}
	refreshToken := base64.URLEncoding.EncodeToString(b)

	// Hash token before storing in DB for security
	hash := sha256.Sum256([]byte(refreshToken))
	hashedToken := hex.EncodeToString(hash[:])

	// Store hashed token in DB
	expiresAt := time.Now().Add(time.Hour * 24 * 7) // Berlaku 7 hari
	if err := service.authRepo.StoreRefreshToken(ctx, user.ID, hashedToken, expiresAt); err != nil {
		return domain.LoginResponse{}, err
	}

	return domain.LoginResponse{
		Token:        tokenString,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}

func (service *authService) Logout(ctx context.Context, request domain.LogoutRequest) error {
	if err := service.validate.Struct(request); err != nil {
		return validation.FormatValidationError(err)
	}

	// Hash the incoming token to match the database
	hash := sha256.Sum256([]byte(request.RefreshToken))
	hashedToken := hex.EncodeToString(hash[:])

	return service.authRepo.RevokeRefreshToken(ctx, hashedToken)
}

func (service *authService) Refresh(ctx context.Context, request domain.RefreshRequest) (domain.LoginResponse, error) {
	if err := service.validate.Struct(request); err != nil {
		return domain.LoginResponse{}, validation.FormatValidationError(err)
	}

	hash := sha256.Sum256([]byte(request.RefreshToken))
	hashedToken := hex.EncodeToString(hash[:])

	rt, err := service.authRepo.GetRefreshToken(ctx, hashedToken)
	if err != nil {
		return domain.LoginResponse{}, err
	}
	if rt == nil {
		return domain.LoginResponse{}, errors.New("invalid refresh token")
	}

	if rt.RevokedAt != nil {
		return domain.LoginResponse{}, errors.New("refresh token has been revoked")
	}

	if time.Now().After(rt.ExpiresAt) {
		return domain.LoginResponse{}, errors.New("refresh token has expired")
	}

	user, err := service.userRepo.GetByID(ctx, rt.UserID)
	if err != nil || user == nil {
		return domain.LoginResponse{}, errors.New("user not found")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Minute * 15).Unix(), // Berlaku 15 Menit lagi
	})

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "rahasia_default"
	}

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return domain.LoginResponse{}, err
	}

	// We can choose to rotate the refresh token or keep using the same one until it expires.
	// For simplicity, we return the same refresh token (it will expire in up to 7 days from initial login).
	return domain.LoginResponse{
		Token:        tokenString,
		RefreshToken: request.RefreshToken,
		User:         *user,
	}, nil
}

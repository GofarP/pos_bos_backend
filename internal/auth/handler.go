package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"pos_bos/internal/core/domain"
	"pos_bos/pkg/response"
	"pos_bos/pkg/validation"
	"time"

	"github.com/go-chi/chi/v5"
)

type AuthHandler struct {
	service domain.AuthService
}

func NewAuthHandler(service domain.AuthService) *AuthHandler {
	return &AuthHandler{
		service: service,
	}
}

func (handler *AuthHandler) RegisterRoutes(router chi.Router) {
	router.Post("/login", handler.Login)
	router.Post("/logout", handler.Logout)
	router.Post("/refresh", handler.Refresh)
}

func (handler *AuthHandler) Login(responseWriter http.ResponseWriter, request *http.Request) {
	var loginReq domain.LoginRequest
	if err := json.NewDecoder(request.Body).Decode(&loginReq); err != nil {
		response.Error(responseWriter, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	res, err := handler.service.Login(request.Context(), loginReq)
	if err != nil {
		var valErr *validation.ValidationError
		if errors.As(err, &valErr) {
			response.ValidationErrors(responseWriter, http.StatusBadRequest, valErr.Errors)
			return
		}
		if errors.Is(err, ErrInvalidCredentials) {
			response.Error(responseWriter, http.StatusUnauthorized, err.Error())
			return
		}
		response.Error(responseWriter, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Set HttpOnly Cookies
	http.SetCookie(responseWriter, &http.Cookie{
		Name:     "access_token",
		Value:    res.Token,
		Expires:  time.Now().Add(15 * time.Minute),
		HttpOnly: true,
		Secure:   false, // Set true in production (HTTPS)
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(responseWriter, &http.Cookie{
		Name:     "refresh_token",
		Value:    res.RefreshToken,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   false, 
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(responseWriter, &http.Cookie{
		Name:     "is_logged_in",
		Value:    "1",
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: false,
		Secure:   false,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	response.JSON(responseWriter, http.StatusOK, map[string]interface{}{
		"message": "Login successful",
		"data":    res,
	})
}

func (handler *AuthHandler) Logout(responseWriter http.ResponseWriter, request *http.Request) {
	var req domain.LogoutRequest
	_ = json.NewDecoder(request.Body).Decode(&req)

	if req.RefreshToken == "" {
		if cookie, err := request.Cookie("refresh_token"); err == nil {
			req.RefreshToken = cookie.Value
		}
	}

	if req.RefreshToken == "" {
		response.Error(responseWriter, http.StatusBadRequest, "Missing refresh token")
		return
	}

	if err := handler.service.Logout(request.Context(), req); err != nil {
		var valErr *validation.ValidationError
		if errors.As(err, &valErr) {
			response.ValidationErrors(responseWriter, http.StatusBadRequest, valErr.Errors)
			return
		}
		
		response.Error(responseWriter, http.StatusBadRequest, err.Error())
		return
	}

	// Clear Cookies
	http.SetCookie(responseWriter, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(responseWriter, &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(responseWriter, &http.Cookie{
		Name:     "is_logged_in",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: false,
		Secure:   false,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	response.JSON(responseWriter, http.StatusOK, map[string]string{"message": "Successfully logged out"})
}

func (handler *AuthHandler) Refresh(responseWriter http.ResponseWriter, request *http.Request) {
	var req domain.RefreshRequest
	_ = json.NewDecoder(request.Body).Decode(&req)

	if req.RefreshToken == "" {
		if cookie, err := request.Cookie("refresh_token"); err == nil {
			req.RefreshToken = cookie.Value
		}
	}

	if req.RefreshToken == "" {
		response.Error(responseWriter, http.StatusBadRequest, "Missing refresh token")
		return
	}

	res, err := handler.service.Refresh(request.Context(), req)
	if err != nil {
		var valErr *validation.ValidationError
		if errors.As(err, &valErr) {
			response.ValidationErrors(responseWriter, http.StatusBadRequest, valErr.Errors)
			return
		}

		response.Error(responseWriter, http.StatusUnauthorized, err.Error())
		return
	}

	// Set New Cookies
	http.SetCookie(responseWriter, &http.Cookie{
		Name:     "access_token",
		Value:    res.Token,
		Expires:  time.Now().Add(15 * time.Minute),
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(responseWriter, &http.Cookie{
		Name:     "refresh_token",
		Value:    res.RefreshToken,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: true,
		Secure:   false,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(responseWriter, &http.Cookie{
		Name:     "is_logged_in",
		Value:    "1",
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		HttpOnly: false,
		Secure:   false,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	response.JSON(responseWriter, http.StatusOK, res)
}

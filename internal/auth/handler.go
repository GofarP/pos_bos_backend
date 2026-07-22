package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"pos_bos/internal/core/domain"
	"pos_bos/pkg/response"
	"pos_bos/pkg/validation"

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

	response.JSON(responseWriter, http.StatusOK, map[string]interface{}{
		"message": "Login successful",
		"data":    res,
	})
}

func (handler *AuthHandler) Logout(responseWriter http.ResponseWriter, request *http.Request) {
	var req domain.LogoutRequest
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		response.Error(responseWriter, http.StatusBadRequest, "Invalid JSON payload")
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

	response.JSON(responseWriter, http.StatusOK, map[string]string{"message": "Successfully logged out"})
}

func (handler *AuthHandler) Refresh(responseWriter http.ResponseWriter, request *http.Request) {
	var req domain.RefreshRequest
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		response.Error(responseWriter, http.StatusBadRequest, "Invalid JSON payload")
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

	response.JSON(responseWriter, http.StatusOK, res)
}

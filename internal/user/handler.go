package user

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"pos_bos/internal/core/domain"
	"pos_bos/pkg/response"
	"pos_bos/pkg/validation"

	"github.com/go-chi/chi/v5"
)

var ErrFileTooLarge = errors.New("ukuran file melebihi batas maksimal 2MB")

type UserHandler struct {
	service domain.UserService
}

func NewUserHandler(service domain.UserService) *UserHandler {
	return &UserHandler{
		service: service,
	}
}

func (handler *UserHandler) RegisterRoutes(router chi.Router) {
	router.Post("/users", handler.AddUser)
	router.Get("/users", handler.GetUsers)
	router.Put("/users/{id}", handler.UpdateUser)
	router.Delete("/users/{id}", handler.DeleteUser)
}

func (handler *UserHandler) GetUsers(responseWriter http.ResponseWriter, request *http.Request) {
	page, _ := strconv.Atoi(request.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(request.URL.Query().Get("limit"))
	search := request.URL.Query().Get("search")

	paginationRequest := domain.PaginationRequest{
		Page:   page,
		Limit:  limit,
		Search: search,
	}

	res, err := handler.service.GetAllUsers(request.Context(), paginationRequest)
	if err != nil {
		response.Error(responseWriter, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(responseWriter, http.StatusOK, res)
}

func (handler *UserHandler) AddUser(responseWriter http.ResponseWriter, request *http.Request) {
	if err := request.ParseMultipartForm(10 << 20); err != nil {
		response.Error(responseWriter, http.StatusBadRequest, "Invalid form data")
		return
	}

	photoPath, err := handler.handlePhotoUpload(request, "photo")
	if err != nil {
		if errors.Is(err, ErrFileTooLarge) {
			response.Error(responseWriter, http.StatusBadRequest, err.Error())
		} else {
			response.Error(responseWriter, http.StatusInternalServerError, "Failed to upload photo")
		}
		return
	}

	userRequest := domain.AddUserRequest{
		Name:     request.FormValue("name"),
		Email:    request.FormValue("email"),
		Password: request.FormValue("password"),
		Photo:    photoPath,
	}

	user, err := handler.service.AddUser(request.Context(), userRequest)
	if err != nil {
		var valErr *validation.ValidationError
		if errors.As(err, &valErr) {
			response.ValidationErrors(responseWriter, http.StatusBadRequest, valErr.Errors)
			return
		}
		if errors.Is(err, ErrEmailAlreadyExists) {
			response.Error(responseWriter, http.StatusConflict, err.Error())
			return
		}
		response.Error(responseWriter, http.StatusInternalServerError, "Failed to create user")
		return
	}

	response.JSON(responseWriter, http.StatusCreated, map[string]interface{}{
		"message": "User created successfully",
		"data":    user,
	})
}

func (handler *UserHandler) UpdateUser(responseWriter http.ResponseWriter, request *http.Request) {
	idStr := chi.URLParam(request, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(responseWriter, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if err := request.ParseMultipartForm(10 << 20); err != nil {
		response.Error(responseWriter, http.StatusBadRequest, "Invalid form data")
		return
	}

	photoPath, err := handler.handlePhotoUpload(request, "photo")
	if err != nil {
		if errors.Is(err, ErrFileTooLarge) {
			response.Error(responseWriter, http.StatusBadRequest, err.Error())
		} else {
			response.Error(responseWriter, http.StatusInternalServerError, "Failed to upload photo")
		}
		return
	}

	password := request.FormValue("password")
	var passwordPtr *string
	if password != "" {
		passwordPtr = &password
	}

	updateRequest := domain.UpdateUserRequest{
		Name:     request.FormValue("name"),
		Email:    request.FormValue("email"),
		Password: passwordPtr,
		Photo:    photoPath,
	}

	user, err := handler.service.UpdateUser(request.Context(), id, updateRequest)
	if err != nil {
		var valErr *validation.ValidationError
		if errors.As(err, &valErr) {
			response.ValidationErrors(responseWriter, http.StatusBadRequest, valErr.Errors)
			return
		}
		response.Error(responseWriter, http.StatusInternalServerError, "Failed to update user")
		return
	}

	response.JSON(responseWriter, http.StatusOK, map[string]interface{}{
		"message": "User updated successfully",
		"data":    user,
	})
}

func (handler *UserHandler) DeleteUser(responseWriter http.ResponseWriter, request *http.Request) {
	idStr := chi.URLParam(request, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(responseWriter, http.StatusBadRequest, "Invalid user ID")
		return
	}

	err = handler.service.DeleteUser(request.Context(), id)
	if err != nil {
		response.Error(responseWriter, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	response.JSON(responseWriter, http.StatusOK, map[string]string{"message": "User deleted successfully"})
}


func (handler *UserHandler) handlePhotoUpload(request *http.Request, fieldName string) (*string, error) {
	file, header, err := request.FormFile(fieldName)
	if err != nil {
		if err == http.ErrMissingFile {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	const maxUploadSize = 2 << 20
	if header.Size > maxUploadSize {
		return nil, ErrFileTooLarge
	}

	uploadDir := "uploads/photos"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return nil, err
	}

	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	filepath := filepath.Join(uploadDir, filename)

	out, err := os.Create(filepath)
	if err != nil {
		return nil, err
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		return nil, err
	}

	relativePath := filepath
	return &relativePath, nil
}

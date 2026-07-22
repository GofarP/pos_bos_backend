package category

import (
	"encoding/json"
	"errors"
	"net/http"
	"pos_bos/internal/core/domain"
	"pos_bos/pkg/response"
	"pos_bos/pkg/validation"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type CategoryHandler struct {
	service domain.CategoryService
}

func NewCategoryHandler(service domain.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

func (handler *CategoryHandler) RegisterRoutes(router chi.Router) {
	router.Post("/categories", handler.CreateCategory)
	router.Get("/categories", handler.GetAllCategories)
	router.Get("/categories/{id}", handler.GetCategoryByID)
	router.Put("/categories/{id}", handler.UpdateCategory)
	router.Delete("/categories/{id}", handler.DeleteCategory)
}

func (handler *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req domain.CategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}
	res, err := handler.service.CreateCategory(r.Context(), req)
	if err != nil {
		var valErr *validation.ValidationError
		if errors.As(err, &valErr) {
			response.ValidationErrors(w, http.StatusBadRequest, valErr.Errors)
			return
		}
		if err.Error() == "category name already exists" {
			response.Error(w, http.StatusConflict, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusCreated, res)
}

func (handler *CategoryHandler) GetAllCategories(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	req := domain.PaginationRequest{
		Page:  page,
		Limit: limit,
	}
	res, err := handler.service.GetAllCategories(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, res)
}

func (handler *CategoryHandler) GetCategoryByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	res, err := handler.service.GetCategoryByID(r.Context(), id)
	if err != nil {
		if err.Error() == "category not found" {
			response.Error(w, http.StatusNotFound, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, res)
}

func (handler *CategoryHandler) UpdateCategory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	var req domain.CategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}
	res, err := handler.service.UpdateCategory(r.Context(), id, req)
	if err != nil {
		if err.Error() == "category not found" {
			response.Error(w, http.StatusNotFound, err.Error())
			return
		}
		if err.Error() == "category name already exists" {
			response.Error(w, http.StatusConflict, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, res)
}

func (handler *CategoryHandler) DeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	if err := handler.service.DeleteCategory(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "Category deleted successfully"})
}

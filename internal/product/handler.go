package product

import (
	"encoding/json"
	"errors"
	"net/http"
	"pos_bos/internal/core/domain"
	"pos_bos/pkg/middleware"
	"pos_bos/pkg/response"
	"pos_bos/pkg/validation"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type ProductHandler struct {
	service domain.ProductService
}

func NewProductHandler(service domain.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) RegisterRoutes(router chi.Router, rbacRepo domain.RBACRepository) {
	router.Route("/products", func(r chi.Router) {
		r.With(middleware.RequirePermission("create.product", rbacRepo)).Post("/", h.CreateProduct)
		r.With(middleware.RequirePermission("view.product", rbacRepo)).Get("/", h.GetAllProducts)
		r.With(middleware.RequirePermission("view.product", rbacRepo)).Get("/{id}", h.GetProductByID)
		r.With(middleware.RequirePermission("edit.product", rbacRepo)).Put("/{id}", h.UpdateProduct)
		r.With(middleware.RequirePermission("delete.product", rbacRepo)).Delete("/{id}", h.DeleteProduct)
	})
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req domain.ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	res, err := h.service.CreateProduct(r.Context(), req)
	if err != nil {
		var valErr *validation.ValidationError
		if errors.As(err, &valErr) {
			response.ValidationErrors(w, http.StatusBadRequest, valErr.Errors)
			return
		}
		if err.Error() == "sku already exists" || err.Error() == "category not found" {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusCreated, res)
}

func (h *ProductHandler) GetAllProducts(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	search := r.URL.Query().Get("search")

	req := domain.PaginationRequest{
		Page:   page,
		Limit:  limit,
		Search: search,
	}

	res, err := h.service.GetAllProducts(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, res)
}

func (h *ProductHandler) GetProductByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	res, err := h.service.GetProductByID(r.Context(), id)
	if err != nil {
		if err.Error() == "product not found" {
			response.Error(w, http.StatusNotFound, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, res)
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var req domain.ProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	res, err := h.service.UpdateProduct(r.Context(), id, req)
	if err != nil {
		var valErr *validation.ValidationError
		if errors.As(err, &valErr) {
			response.ValidationErrors(w, http.StatusBadRequest, valErr.Errors)
			return
		}
		if err.Error() == "product not found" {
			response.Error(w, http.StatusNotFound, err.Error())
			return
		}
		if err.Error() == "sku already exists" || err.Error() == "category not found" {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, res)
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	err = h.service.DeleteProduct(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "Product deleted successfully"})
}

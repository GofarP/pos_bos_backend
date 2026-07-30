package rbac

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"pos_bos/internal/core/domain"
	"pos_bos/pkg/middleware"
	"pos_bos/pkg/response"
	"pos_bos/pkg/validation"

	"github.com/go-chi/chi/v5"
)

type RBACHandler struct {
	service domain.RBACService
}

func NewRBACHandler(service domain.RBACService) *RBACHandler {
	return &RBACHandler{
		service: service,
	}
}

func (handler *RBACHandler) RegisterRoutes(router chi.Router, rbacRepo domain.RBACRepository) {
	router.With(middleware.RequirePermission("create.permission", rbacRepo)).Post("/permissions", handler.CreatePermission)
	router.With(middleware.RequirePermission("view.permission", rbacRepo)).Get("/permissions", handler.GetPermissions)
	router.With(middleware.RequirePermission("edit.permission", rbacRepo)).Put("/permissions/{id}", handler.UpdatePermission)
	router.With(middleware.RequirePermission("delete.permission", rbacRepo)).Delete("/permissions/{id}", handler.DeletePermission)

	router.With(middleware.RequirePermission("create.role", rbacRepo)).Post("/roles", handler.CreateRole)
	router.With(middleware.RequirePermission("view.role", rbacRepo)).Get("/roles", handler.GetAllRoles)
	router.With(middleware.RequirePermission("edit.role", rbacRepo)).Put("/roles/{id}", handler.UpdateRole)
	router.With(middleware.RequirePermission("delete.role", rbacRepo)).Delete("/roles/{id}", handler.DeleteRole)
}

func (handler *RBACHandler) GetPermissions(responseWriter http.ResponseWriter, request *http.Request) {
	page, _ := strconv.Atoi(request.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(request.URL.Query().Get("limit"))
	search := request.URL.Query().Get("search")

	paginationRequest := domain.PaginationRequest{
		Page:   page,
		Limit:  limit,
		Search: search,
	}

	res, err := handler.service.GetAllPermissions(request.Context(), paginationRequest)
	if err != nil {
		response.Error(responseWriter, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(responseWriter, http.StatusOK, res)
}

func (handler *RBACHandler) CreatePermission(responseWriter http.ResponseWriter, request *http.Request) {
	var req domain.PermissionRequest
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		response.Error(responseWriter, http.StatusBadRequest, "Invalid request payload")
		return
	}

	permission, err := handler.service.CreatePermission(request.Context(), req)
	if err != nil {
		var valErr *validation.ValidationError
		if errors.As(err, &valErr) {
			response.ValidationErrors(responseWriter, http.StatusBadRequest, valErr.Errors)
			return
		}
		if errors.Is(err, ErrPermissionAlreadyExists) {
			response.Error(responseWriter, http.StatusConflict, err.Error())
			return
		}
		response.Error(responseWriter, http.StatusInternalServerError, "Failed to create permission")
		return
	}

	response.JSON(responseWriter, http.StatusCreated, map[string]interface{}{
		"message": "Permission created successfully",
		"data":    permission,
	})
}

func (handler *RBACHandler) UpdatePermission(responseWriter http.ResponseWriter, request *http.Request) {
	idStr := chi.URLParam(request, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(responseWriter, http.StatusBadRequest, "Invalid permission ID")
		return
	}

	var req domain.PermissionRequest
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		response.Error(responseWriter, http.StatusBadRequest, "Invalid request payload")
		return
	}

	permission, err := handler.service.UpdatePermission(request.Context(), id, req)
	if err != nil {
		var valErr *validation.ValidationError
		if errors.As(err, &valErr) {
			response.ValidationErrors(responseWriter, http.StatusBadRequest, valErr.Errors)
			return
		}
		if errors.Is(err, ErrPermissionAlreadyExists) {
			response.Error(responseWriter, http.StatusConflict, err.Error())
			return
		}
		response.Error(responseWriter, http.StatusInternalServerError, "Failed to update permission")
		return
	}

	response.JSON(responseWriter, http.StatusOK, map[string]interface{}{
		"message": "Permission updated successfully",
		"data":    permission,
	})
}

func (handler *RBACHandler) DeletePermission(responseWriter http.ResponseWriter, request *http.Request) {
	idStr := chi.URLParam(request, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.Error(responseWriter, http.StatusBadRequest, "Invalid permission ID")
		return
	}

	err = handler.service.DeletePermission(request.Context(), id)
	if err != nil {
		response.Error(responseWriter, http.StatusInternalServerError, "Failed to delete permission")
		return
	}

	response.JSON(responseWriter, http.StatusOK, map[string]string{"message": "Permission deleted successfully"})
}

func (handler *RBACHandler) CreateRole(responseWriter http.ResponseWriter, request *http.Request) {
	var req domain.RoleRequest
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		response.Error(responseWriter, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	res, err := handler.service.CreateRole(request.Context(), req)
	if err != nil {
		var valErr *validation.ValidationError
		if errors.As(err, &valErr) {
			response.ValidationErrors(responseWriter, http.StatusBadRequest, valErr.Errors)
			return
		}

		if err.Error() == "role name already exists" {
			response.Error(responseWriter, http.StatusConflict, err.Error())
			return
		}

		response.Error(responseWriter, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(responseWriter, http.StatusCreated, res)
}

func (handler *RBACHandler) GetAllRoles(responseWriter http.ResponseWriter, request *http.Request) {
	page, _ := strconv.Atoi(request.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(request.URL.Query().Get("limit"))

	req := domain.PaginationRequest{
		Page:  page,
		Limit: limit,
	}

	res, err := handler.service.GetAllRoles(request.Context(), req)
	if err != nil {
		response.Error(responseWriter, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(responseWriter, http.StatusOK, res)
}

func (handler *RBACHandler) UpdateRole(responseWriter http.ResponseWriter, request *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(request, "id"))
	if err != nil {
		response.Error(responseWriter, http.StatusBadRequest, "Invalid ID")
		return
	}

	var req domain.RoleRequest
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		response.Error(responseWriter, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	res, err := handler.service.UpdateRole(request.Context(), id, req)
	if err != nil {
		response.Error(responseWriter, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(responseWriter, http.StatusOK, res)
}

func (handler *RBACHandler) DeleteRole(responseWriter http.ResponseWriter, request *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(request, "id"))
	if err != nil {
		response.Error(responseWriter, http.StatusBadRequest, "Invalid ID")
		return
	}

	if err := handler.service.DeleteRole(request.Context(), id); err != nil {
		response.Error(responseWriter, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(responseWriter, http.StatusOK, map[string]string{"message": "Role deleted successfully"})
}

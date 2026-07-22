package transaction

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

type TransactionHandler struct {
	service domain.TransactionService
}

func NewTransactionHandler(service domain.TransactionService) *TransactionHandler {
	return &TransactionHandler{service: service}
}

func (h *TransactionHandler) RegisterRoutes(router chi.Router) {
	router.Post("/transactions", h.Checkout)
	router.Get("/transactions", h.GetAllTransactions)
	router.Get("/transactions/{id}", h.GetTransactionByID)

}

func (h *TransactionHandler) Checkout(w http.ResponseWriter, r *http.Request) {
	var req domain.TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	// Ambil UserID dari JWT Context
	userIDFloat, ok := r.Context().Value(middleware.UserIDKey).(float64)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "Unauthorized: Invalid token data")
		return
	}
	req.UserID = int(userIDFloat)

	res, err := h.service.ProcessTransaction(r.Context(), req)
	if err != nil {
		var valErr *validation.ValidationError
		if errors.As(err, &valErr) {
			response.ValidationErrors(w, http.StatusBadRequest, valErr.Errors)
			return
		}
		if err.Error() == "insufficient stock for product ID" || err.Error() == "transaction with this idempotency key already processed" {
			response.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, res)
}

func (h *TransactionHandler) GetAllTransactions(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	if pageStr != "" {
		p, err := strconv.Atoi(pageStr)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "Invalid page parameter")
			return
		}
		page = p
	}

	limit := 10
	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil {
			response.Error(w, http.StatusBadRequest, "Invalid limit parameter")
			return
		}
		limit = l
	}

	req := domain.PaginationRequest{
		Page:  page,
		Limit: limit,
	}

	res, err := h.service.GetAllTransactions(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(w, http.StatusOK, res)
}

func (h *TransactionHandler) GetTransactionByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid ID")
		return
	}
	res, err := h.service.GetTransactionByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrTransactionNotFound) {
			response.Error(w, http.StatusNotFound, err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	response.JSON(w, http.StatusOK, res)
}

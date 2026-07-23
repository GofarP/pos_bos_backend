package domain

import (
	"context"
	"errors"
	"time"
)

var ErrTransactionNotFound = errors.New("transaction not found")

type Transaction struct {
	ID             int               `json:"id"`
	UserID         int               `json:"user_id"`
	InvoiceNumber  string            `json:"invoice_number"`
	TotalAmount    int64             `json:"total_amount"`
	IdempotencyKey string            `json:"idempotency_key"`
	Status         string            `json:"status"`
	Items          []TransactionItem `json:"items"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

type TransactionFilterRequest struct {
	Page      int    `json:"page"`
	Limit     int    `json:"limit"`
	UserID    int    `json:"user_id,omitempty"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
}

type TransactionItem struct {
	ID            int     `json:"id"`
	TransactionID int     `json:"transaction_id"`
	ProductID     int     `json:"product_id"`
	Quantity      int     `json:"quantity"`
	Price         int64   `json:"price"`
	Subtotal      int64   `json:"subtotal"`
}

type TransactionItemRequest struct {
	ProductID int `json:"product_id" validate:"required"`
	Quantity  int `json:"quantity" validate:"required,min=1"`
}

type TransactionRequest struct {
	IdempotencyKey string                   `json:"idempotency_key" validate:"required"`
	UserID         int                      `json:"-"` // Diambil dari JWT
	Items          []TransactionItemRequest `json:"items" validate:"required,min=1,dive"`
}

type TransactionRepository interface {
	CheckIdempotencyKey(ctx context.Context, key string) (bool, error)
	CreateTransaction(ctx context.Context, txReq *Transaction) error
	GetAllTransactions(ctx context.Context, req TransactionFilterRequest) ([]Transaction, int, error)
	GetTransactionByID(ctx context.Context, id int) (*Transaction, error)
	CancelTransaction(ctx context.Context, id int) error
}

type TransactionService interface {
	ProcessTransaction(ctx context.Context, req TransactionRequest) (*Transaction, error)
	GetAllTransactions(ctx context.Context, req TransactionFilterRequest) (PaginationResponse, error)
	GetTransactionByID(ctx context.Context, id int) (*Transaction, error)
	CancelTransaction(ctx context.Context, id int) error
}

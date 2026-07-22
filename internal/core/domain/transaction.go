package domain

import (
	"context"
	"time"
)

type Transaction struct {
	ID             int               `json:"id"`
	UserID         int               `json:"user_id"`
	InvoiceNumber  string            `json:"invoice_number"`
	TotalAmount    float64           `json:"total_amount"`
	IdempotencyKey string            `json:"idempotency_key"`
	Items          []TransactionItem `json:"items"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
}

type TransactionItem struct {
	ID            int     `json:"id"`
	TransactionID int     `json:"transaction_id"`
	ProductID     int     `json:"product_id"`
	Quantity      int     `json:"quantity"`
	Price         float64 `json:"price"`
	Subtotal      float64 `json:"subtotal"`
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
}

type TransactionService interface {
	ProcessTransaction(ctx context.Context, req TransactionRequest) (*Transaction, error)
}

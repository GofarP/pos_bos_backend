package transaction

import (
	"context"
	"errors"
	"fmt"
	"pos_bos/internal/core/domain"
	"pos_bos/pkg/validation"
	"time"

	"github.com/go-playground/validator/v10"
)

type transactionService struct {
	repo     domain.TransactionRepository
	validate *validator.Validate
}

func NewTransactionService(repo domain.TransactionRepository) domain.TransactionService {
	return &transactionService{
		repo:     repo,
		validate: validator.New(),
	}
}

func (s *transactionService) ProcessTransaction(ctx context.Context, req domain.TransactionRequest) (*domain.Transaction, error) {
	if err := s.validate.Struct(req); err != nil {
		return nil, validation.FormatValidationError(err)
	}

	exists, err := s.repo.CheckIdempotencyKey(ctx, req.IdempotencyKey)

	if err != nil {
		return nil, err
	}

	if exists {
		return nil, errors.New("transaction with this idempotency key already processed")
	}

	invoiceNumber := fmt.Sprintf("TRX-%d", time.Now().Unix())
	var items []domain.TransactionItem
	for _, reqItem := range req.Items {
		items = append(items, domain.TransactionItem{
			ProductID: reqItem.ProductID,
			Quantity:  reqItem.Quantity,
		})
	}

	txData := &domain.Transaction{
		UserID:         req.UserID,
		InvoiceNumber:  invoiceNumber,
		IdempotencyKey: req.IdempotencyKey,
		Status:         "COMPLETED",
		Items:          items,
	}

	err = s.repo.CreateTransaction(ctx, txData)
	if err != nil {
		return nil, err
	}

	return txData, nil

}

func (s *transactionService) GetAllTransactions(ctx context.Context, req domain.TransactionFilterRequest) (domain.PaginationResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}

	if req.Limit <= 0 {
		req.Limit = 10
	}

	transactions, total, err := s.repo.GetAllTransactions(ctx, req)
	if err != nil {
		return domain.PaginationResponse{}, err
	}
	totalPages := total / req.Limit
	if total%req.Limit > 0 {
		totalPages++
	}

	return domain.PaginationResponse{
		Data: transactions,
		Meta: domain.PaginationMeta{
			Page:         req.Page,
			Limit:        req.Limit,
			TotalRecords: total,
			TotalPages:   totalPages,
		},
	}, nil
}

func (s *transactionService) GetTransactionByID(ctx context.Context, id int) (*domain.Transaction, error) {
	tx, err := s.repo.GetTransactionByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if tx == nil {
		return nil, domain.ErrTransactionNotFound
	}

	return tx, nil
}

func (s *transactionService) CancelTransaction(ctx context.Context, id int) error {
	err := s.repo.CancelTransaction(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

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
		Items:          items,
	}

	err = s.repo.CreateTransaction(ctx, txData)
	if err != nil {
		return nil, err
	}

	return txData, nil

}

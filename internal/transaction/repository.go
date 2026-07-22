package transaction

import (
	"context"
	"database/sql"
	"errors"
	"pos_bos/internal/core/domain"
)

type mysqlTransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) domain.TransactionRepository {
	return &mysqlTransactionRepository{db: db}
}

func (r *mysqlTransactionRepository) CheckIdempotencyKey(ctx context.Context, key string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM transactions WHERE idempotency_key = ?)`
	err := r.db.QueryRowContext(ctx, query, key).Scan(&exists)
	return exists, err
}

func (r *mysqlTransactionRepository) CreateTransaction(ctx context.Context, txReq *domain.Transaction) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	queryTx := `INSERT INTO transactions(user_id, invoice_number, total_amount, idempotency_key) values(?,?,?,?)`
	res, err := tx.ExecContext(ctx, queryTx, txReq.UserID, txReq.InvoiceNumber, 0, txReq.IdempotencyKey)
	if err != nil {
		return err
	}

	txID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	txReq.ID = int(txID)

	var totalAmount float64

	for i, item := range txReq.Items {
		var currentStock int
		var price float64
		err := tx.QueryRowContext(ctx, "SELECT stock, price FROM products WHERE id = ? FOR UPDATE", item.ProductID).Scan(&currentStock, &price)
		if err != nil {
			return errors.New(("Product not found or failed to lock"))
		}

		if currentStock < item.Quantity {
			return errors.New("insufficient stock for product ID")
		}

		subtotal := price * float64(item.Quantity)
		txReq.Items[i].Price = price
		txReq.Items[i].Subtotal = subtotal
		totalAmount += subtotal

		_, err = tx.ExecContext(ctx, "UPDATE products SET stock = stock - ? WHERE id = ?", item.Quantity, item.ProductID)
		if err != nil {
			return err
		}

		queryItem := `INSERT INTO transaction_items (transaction_id, product_id, quantity, price, subtotal) VALUES (?, ?, ?, ?, ?)`

		itemRes, err := tx.ExecContext(ctx, queryItem, txReq.ID, item.ProductID, item.Quantity, price, subtotal)
		if err != nil {
			return err
		}

		itemID, _ := itemRes.LastInsertId()
		txReq.Items[i].ID = int(itemID)
		txReq.Items[i].TransactionID = txReq.ID

	}

	_, err = tx.ExecContext(ctx, "UPDATE transactions SET total_amount=? where id=?", totalAmount, txReq.ID)

	if err != nil {
		return err
	}

	txReq.TotalAmount = totalAmount

	return tx.Commit()
}

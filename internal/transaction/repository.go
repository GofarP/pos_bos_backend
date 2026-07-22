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

func (r *mysqlTransactionRepository) GetAllTransactions(ctx context.Context, req domain.TransactionFilterRequest) ([]domain.Transaction, int, error) {
	// Build dynamic query
	whereQuery := "WHERE 1=1"
	args := []interface{}{}

	if req.UserID != 0 {
		whereQuery += " AND user_id = ?"
		args = append(args, req.UserID)
	}
	if req.StartDate != "" && req.EndDate != "" {
		whereQuery += " AND DATE(created_at) BETWEEN ? AND ?"
		args = append(args, req.StartDate, req.EndDate)
	}

	var total int
	countQuery := "SELECT COUNT(id) FROM transactions " + whereQuery
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (req.Page - 1) * req.Limit
	if offset < 0 {
		offset = 0
	}

	query := `SELECT id, user_id, invoice_number, total_amount, idempotency_key, status, created_at, updated_at 
	          FROM transactions ` + whereQuery + ` ORDER BY id DESC LIMIT ? OFFSET ?`

	args = append(args, req.Limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var transactions []domain.Transaction
	for rows.Next() {
		var tx domain.Transaction
		err := rows.Scan(&tx.ID, &tx.UserID, &tx.InvoiceNumber, &tx.TotalAmount, &tx.IdempotencyKey, &tx.Status, &tx.CreatedAt, &tx.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		transactions = append(transactions, tx)
	}
	return transactions, total, nil
}

func (r *mysqlTransactionRepository) GetTransactionByID(ctx context.Context, id int) (*domain.Transaction, error) {
	// 1. Ambil Data Header Transaksi
	queryTx := `SELECT id, user_id, invoice_number, total_amount, idempotency_key, status, created_at, updated_at 
	            FROM transactions WHERE id = ?`

	var tx domain.Transaction
	err := r.db.QueryRowContext(ctx, queryTx, id).Scan(
		&tx.ID, &tx.UserID, &tx.InvoiceNumber, &tx.TotalAmount, &tx.IdempotencyKey, &tx.Status, &tx.CreatedAt, &tx.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil jika tidak ditemukan
		}
		return nil, err
	}
	// 2. Ambil Data Detail Item-nya
	queryItems := `SELECT id, transaction_id, product_id, quantity, price, subtotal 
	               FROM transaction_items WHERE transaction_id = ?`

	rows, err := r.db.QueryContext(ctx, queryItems, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item domain.TransactionItem
		err := rows.Scan(&item.ID, &item.TransactionID, &item.ProductID, &item.Quantity, &item.Price, &item.Subtotal)
		if err != nil {
			return nil, err
		}
		tx.Items = append(tx.Items, item)
	}
	return &tx, nil
}

func (r *mysqlTransactionRepository) CancelTransaction(ctx context.Context, id int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Check if transaction can be cancelled
	var currentStatus string
	err = tx.QueryRowContext(ctx, "SELECT status FROM transactions WHERE id = ? FOR UPDATE", id).Scan(&currentStatus)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.ErrTransactionNotFound
		}
		return err
	}

	if currentStatus == "CANCELLED" {
		return errors.New("transaction is already cancelled")
	}

	// 2. Fetch items to return stock
	rows, err := tx.QueryContext(ctx, "SELECT product_id, quantity FROM transaction_items WHERE transaction_id = ?", id)
	if err != nil {
		return err
	}
	
	type item struct {
		productID int
		quantity  int
	}
	var itemsToReturn []item
	for rows.Next() {
		var i item
		if err := rows.Scan(&i.productID, &i.quantity); err != nil {
			rows.Close()
			return err
		}
		itemsToReturn = append(itemsToReturn, i)
	}
	rows.Close()

	// 3. Return stock to products
	for _, i := range itemsToReturn {
		_, err := tx.ExecContext(ctx, "UPDATE products SET stock = stock + ? WHERE id = ?", i.quantity, i.productID)
		if err != nil {
			return err
		}
	}

	// 4. Update status to CANCELLED
	_, err = tx.ExecContext(ctx, "UPDATE transactions SET status = 'CANCELLED' WHERE id = ?", id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

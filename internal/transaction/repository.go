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
	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	err = stmt.QueryRowContext(ctx, key).Scan(&exists)
	return exists, err
}

func (r *mysqlTransactionRepository) CreateTransaction(ctx context.Context, txReq *domain.Transaction) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queryTx := `INSERT INTO transactions(user_id, invoice_number, total_amount, idempotency_key) values(?,?,?,?)`
	stmtTx, err := tx.PrepareContext(ctx, queryTx)
	if err != nil {
		return err
	}
	defer stmtTx.Close()

	res, err := stmtTx.ExecContext(ctx, txReq.UserID, txReq.InvoiceNumber, 0, txReq.IdempotencyKey)
	if err != nil {
		return err
	}

	txID, err := res.LastInsertId()
	if err != nil {
		return err
	}

	txReq.ID = int(txID)

	var totalAmount int64

	// Prepare statements once outside the loop
	queryProduct := "SELECT stock, price FROM products WHERE id = ? FOR UPDATE"
	stmtProduct, err := tx.PrepareContext(ctx, queryProduct)
	if err != nil {
		return err
	}
	defer stmtProduct.Close()

	queryUpdateProduct := "UPDATE products SET stock = stock - ? WHERE id = ?"
	stmtUpdateProduct, err := tx.PrepareContext(ctx, queryUpdateProduct)
	if err != nil {
		return err
	}
	defer stmtUpdateProduct.Close()

	queryItem := `INSERT INTO transaction_items (transaction_id, product_id, quantity, price, subtotal) VALUES (?, ?, ?, ?, ?)`
	stmtItem, err := tx.PrepareContext(ctx, queryItem)
	if err != nil {
		return err
	}
	defer stmtItem.Close()

	for i, item := range txReq.Items {
		var currentStock int
		var price int64
		err := stmtProduct.QueryRowContext(ctx, item.ProductID).Scan(&currentStock, &price)
		if err != nil {
			return errors.New("Product not found or failed to lock")
		}

		if currentStock < item.Quantity {
			return errors.New("insufficient stock for product ID")
		}

		subtotal := price * int64(item.Quantity)
		txReq.Items[i].Price = price
		txReq.Items[i].Subtotal = subtotal
		totalAmount += subtotal

		_, err = stmtUpdateProduct.ExecContext(ctx, item.Quantity, item.ProductID)
		if err != nil {
			return err
		}

		itemRes, err := stmtItem.ExecContext(ctx, txReq.ID, item.ProductID, item.Quantity, price, subtotal)
		if err != nil {
			return err
		}

		itemID, _ := itemRes.LastInsertId()
		txReq.Items[i].ID = int(itemID)
		txReq.Items[i].TransactionID = txReq.ID
	}

	queryUpdateTx := "UPDATE transactions SET total_amount=? where id=?"
	stmtUpdateTx, err := tx.PrepareContext(ctx, queryUpdateTx)
	if err != nil {
		return err
	}
	defer stmtUpdateTx.Close()

	_, err = stmtUpdateTx.ExecContext(ctx, totalAmount, txReq.ID)
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
	stmtCount, err := r.db.PrepareContext(ctx, countQuery)
	if err != nil {
		return nil, 0, err
	}
	defer stmtCount.Close()

	if err := stmtCount.QueryRowContext(ctx, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (req.Page - 1) * req.Limit
	if offset < 0 {
		offset = 0
	}

	query := `SELECT t.id, t.user_id, COALESCE(u.name, '') as user_name, t.invoice_number, t.total_amount, t.idempotency_key, t.status, t.created_at, t.updated_at 
	          FROM transactions t
	          LEFT JOIN users u ON t.user_id = u.id ` + whereQuery + ` ORDER BY t.id DESC LIMIT ? OFFSET ?`

	selectArgs := append(args, req.Limit, offset)

	stmtSelect, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	defer stmtSelect.Close()

	rows, err := stmtSelect.QueryContext(ctx, selectArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var transactions []domain.Transaction
	for rows.Next() {
		var tx domain.Transaction
		err := rows.Scan(&tx.ID, &tx.UserID, &tx.UserName, &tx.InvoiceNumber, &tx.TotalAmount, &tx.IdempotencyKey, &tx.Status, &tx.CreatedAt, &tx.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		transactions = append(transactions, tx)
	}
	return transactions, total, nil
}

func (r *mysqlTransactionRepository) GetTransactionByID(ctx context.Context, id int) (*domain.Transaction, error) {
	queryTx := `SELECT t.id, t.user_id, COALESCE(u.name, '') as user_name, t.invoice_number, t.total_amount, t.idempotency_key, t.status, t.created_at, t.updated_at 
	            FROM transactions t
	            LEFT JOIN users u ON t.user_id = u.id
	            WHERE t.id = ?`

	stmtTx, err := r.db.PrepareContext(ctx, queryTx)
	if err != nil {
		return nil, err
	}
	defer stmtTx.Close()

	var tx domain.Transaction
	err = stmtTx.QueryRowContext(ctx, id).Scan(
		&tx.ID, &tx.UserID, &tx.UserName, &tx.InvoiceNumber, &tx.TotalAmount, &tx.IdempotencyKey, &tx.Status, &tx.CreatedAt, &tx.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	queryItems := `SELECT ti.id, ti.transaction_id, ti.product_id, COALESCE(p.name, '') as product_name, COALESCE(p.sku, '') as product_sku, ti.quantity, ti.price, ti.subtotal 
	               FROM transaction_items ti
	               LEFT JOIN products p ON ti.product_id = p.id
	               WHERE ti.transaction_id = ?`

	stmtItems, err := r.db.PrepareContext(ctx, queryItems)
	if err != nil {
		return nil, err
	}
	defer stmtItems.Close()

	rows, err := stmtItems.QueryContext(ctx, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tx.Items = []domain.TransactionItem{}
	for rows.Next() {
		var item domain.TransactionItem
		err := rows.Scan(&item.ID, &item.TransactionID, &item.ProductID, &item.ProductName, &item.ProductSKU, &item.Quantity, &item.Price, &item.Subtotal)
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
	queryStatus := "SELECT status FROM transactions WHERE id = ? FOR UPDATE"
	stmtStatus, err := tx.PrepareContext(ctx, queryStatus)
	if err != nil {
		return err
	}
	defer stmtStatus.Close()

	err = stmtStatus.QueryRowContext(ctx, id).Scan(&currentStatus)
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
	queryItems := "SELECT product_id, quantity FROM transaction_items WHERE transaction_id = ?"
	stmtItems, err := tx.PrepareContext(ctx, queryItems)
	if err != nil {
		return err
	}
	defer stmtItems.Close()

	rows, err := stmtItems.QueryContext(ctx, id)
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
	queryReturnStock := "UPDATE products SET stock = stock + ? WHERE id = ?"
	stmtReturnStock, err := tx.PrepareContext(ctx, queryReturnStock)
	if err != nil {
		return err
	}
	defer stmtReturnStock.Close()

	for _, i := range itemsToReturn {
		_, err := stmtReturnStock.ExecContext(ctx, i.quantity, i.productID)
		if err != nil {
			return err
		}
	}

	// 4. Update status to CANCELLED
	queryCancelTx := "UPDATE transactions SET status = 'CANCELLED' WHERE id = ?"
	stmtCancelTx, err := tx.PrepareContext(ctx, queryCancelTx)
	if err != nil {
		return err
	}
	defer stmtCancelTx.Close()

	_, err = stmtCancelTx.ExecContext(ctx, id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *mysqlTransactionRepository) GetDashboardSummary(ctx context.Context) (*domain.DashboardSummary, error) {
	query := `
		SELECT 
			COALESCE(SUM(total_amount), 0) as total_sales,
			COUNT(id) as new_orders,
			COALESCE((
				SELECT SUM(ti.quantity) 
				FROM transaction_items ti 
				JOIN transactions t2 ON ti.transaction_id = t2.id 
				WHERE DATE(t2.created_at) = CURDATE() AND t2.status = 'COMPLETED'
			), 0) as products_sold
		FROM transactions
		WHERE DATE(created_at) = CURDATE() AND status = 'COMPLETED'
	`

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var summary domain.DashboardSummary
	err = stmt.QueryRowContext(ctx).Scan(&summary.TotalSales, &summary.NewOrders, &summary.ProductsSold)
	if err != nil {
		return nil, err
	}

	return &summary, nil
}

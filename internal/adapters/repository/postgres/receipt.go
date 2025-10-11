package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/jmoiron/sqlx"

	"library-service/internal/domain/payment"
	"library-service/pkg/sqlutil"
)

// ReceiptRepository implements payment.ReceiptRepository for PostgreSQL
type ReceiptRepository struct {
	db *sqlx.DB
}

// NewReceiptRepository creates a new PostgreSQL receipt repository
func NewReceiptRepository(db *sqlx.DB) *ReceiptRepository {
	return &ReceiptRepository{db: db}
}

// receiptRow represents a receipt row in the database
type receiptRow struct {
	ID            string         `db:"id"`
	PaymentID     string         `db:"payment_id"`
	ReceiptNumber string         `db:"receipt_number"`
	MemberID      string         `db:"member_id"`
	Amount        int64          `db:"amount"`
	Currency      string         `db:"currency"`
	PaymentType   string         `db:"payment_type"`
	PaymentMethod string         `db:"payment_method"`
	TransactionID string         `db:"transaction_id"`
	PaymentDate   sql.NullTime   `db:"payment_date"`
	ReceiptDate   sql.NullTime   `db:"receipt_date"`
	Status        string         `db:"status"`
	Description   sql.NullString `db:"description"`
	MemberName    string         `db:"member_name"`
	MemberEmail   string         `db:"member_email"`
	CardMask      sql.NullString `db:"card_mask"`
	Items         []byte         `db:"items"`
	TaxAmount     int64          `db:"tax_amount"`
	TotalAmount   int64          `db:"total_amount"`
	Notes         sql.NullString `db:"notes"`
	CreatedAt     sql.NullTime   `db:"created_at"`
	UpdatedAt     sql.NullTime   `db:"updated_at"`
}

// Create inserts a new receipt record
func (r *ReceiptRepository) Create(receipt payment.Receipt) (string, error) {
	// Marshal items to JSON
	itemsJSON, err := json.Marshal(receipt.Items)
	if err != nil {
		return "", fmt.Errorf("failed to marshal items: %w", err)
	}

	query := `
		INSERT INTO receipts (
			id, payment_id, receipt_number, member_id, amount, currency,
			payment_type, payment_method, transaction_id, payment_date, receipt_date,
			status, description, member_name, member_email, card_mask,
			items, tax_amount, total_amount, notes, created_at, updated_at
		) VALUES (
			:id, :payment_id, :receipt_number, :member_id, :amount, :currency,
			:payment_type, :payment_method, :transaction_id, :payment_date, :receipt_date,
			:status, :description, :member_name, :member_email, :card_mask,
			:items, :tax_amount, :total_amount, :notes, :created_at, :updated_at
		)
		RETURNING id
	`

	rows, err := r.db.NamedQuery(query, map[string]interface{}{
		"id":             receipt.ID,
		"payment_id":     receipt.PaymentID,
		"receipt_number": receipt.ReceiptNumber,
		"member_id":      receipt.MemberID,
		"amount":         receipt.Amount,
		"currency":       receipt.Currency,
		"payment_type":   receipt.PaymentType,
		"payment_method": receipt.PaymentMethod,
		"transaction_id": receipt.TransactionID,
		"payment_date":   receipt.PaymentDate,
		"receipt_date":   receipt.ReceiptDate,
		"status":         receipt.Status,
		"description":    receipt.Description,
		"member_name":    receipt.MemberName,
		"member_email":   receipt.MemberEmail,
		"card_mask":      receipt.CardMask,
		"items":          itemsJSON,
		"tax_amount":     receipt.TaxAmount,
		"total_amount":   receipt.TotalAmount,
		"notes":          receipt.Notes,
		"created_at":     receipt.CreatedAt,
		"updated_at":     receipt.UpdatedAt,
	})

	if err != nil {
		return "", fmt.Errorf("failed to create receipt: %w", err)
	}
	defer rows.Close()

	var id string
	if rows.Next() {
		if err := rows.Scan(&id); err != nil {
			return "", fmt.Errorf("failed to scan receipt ID: %w", err)
		}
	}

	return id, nil
}

// GetByID retrieves a receipt by ID
func (r *ReceiptRepository) GetByID(id string) (payment.Receipt, error) {
	query := `
		SELECT id, payment_id, receipt_number, member_id, amount, currency,
		       payment_type, payment_method, transaction_id, payment_date, receipt_date,
		       status, description, member_name, member_email, card_mask,
		       items, tax_amount, total_amount, notes, created_at, updated_at
		FROM receipts
		WHERE id = $1
	`

	var row receiptRow
	err := r.db.Get(&row, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return payment.Receipt{}, fmt.Errorf("receipt not found: %s", id)
		}
		return payment.Receipt{}, fmt.Errorf("failed to get receipt: %w", err)
	}

	return r.rowToReceipt(row)
}

// GetByPaymentID retrieves a receipt by payment ID
func (r *ReceiptRepository) GetByPaymentID(paymentID string) (payment.Receipt, error) {
	query := `
		SELECT id, payment_id, receipt_number, member_id, amount, currency,
		       payment_type, payment_method, transaction_id, payment_date, receipt_date,
		       status, description, member_name, member_email, card_mask,
		       items, tax_amount, total_amount, notes, created_at, updated_at
		FROM receipts
		WHERE payment_id = $1
	`

	var row receiptRow
	err := r.db.Get(&row, query, paymentID)
	if err != nil {
		if err == sql.ErrNoRows {
			return payment.Receipt{}, fmt.Errorf("receipt not found for payment: %s", paymentID)
		}
		return payment.Receipt{}, fmt.Errorf("failed to get receipt: %w", err)
	}

	return r.rowToReceipt(row)
}

// GetByReceiptNumber retrieves a receipt by receipt number
func (r *ReceiptRepository) GetByReceiptNumber(receiptNumber string) (payment.Receipt, error) {
	query := `
		SELECT id, payment_id, receipt_number, member_id, amount, currency,
		       payment_type, payment_method, transaction_id, payment_date, receipt_date,
		       status, description, member_name, member_email, card_mask,
		       items, tax_amount, total_amount, notes, created_at, updated_at
		FROM receipts
		WHERE receipt_number = $1
	`

	var row receiptRow
	err := r.db.Get(&row, query, receiptNumber)
	if err != nil {
		if err == sql.ErrNoRows {
			return payment.Receipt{}, fmt.Errorf("receipt not found: %s", receiptNumber)
		}
		return payment.Receipt{}, fmt.Errorf("failed to get receipt: %w", err)
	}

	return r.rowToReceipt(row)
}

// ListByMemberID retrieves all receipts for a member
func (r *ReceiptRepository) ListByMemberID(memberID string) ([]payment.Receipt, error) {
	query := `
		SELECT id, payment_id, receipt_number, member_id, amount, currency,
		       payment_type, payment_method, transaction_id, payment_date, receipt_date,
		       status, description, member_name, member_email, card_mask,
		       items, tax_amount, total_amount, notes, created_at, updated_at
		FROM receipts
		WHERE member_id = $1
		ORDER BY created_at DESC
	`

	var rows []receiptRow
	err := r.db.Select(&rows, query, memberID)
	if err != nil {
		return nil, fmt.Errorf("failed to list receipts: %w", err)
	}

	receipts := make([]payment.Receipt, len(rows))
	for i, row := range rows {
		receipt, err := r.rowToReceipt(row)
		if err != nil {
			return nil, fmt.Errorf("converting row to receipt: %w", err)
		}
		receipts[i] = receipt
	}

	return receipts, nil
}

// Update updates an existing receipt
func (r *ReceiptRepository) Update(receipt payment.Receipt) error {
	// Marshal items to JSON
	itemsJSON, err := json.Marshal(receipt.Items)
	if err != nil {
		return fmt.Errorf("failed to marshal items: %w", err)
	}

	query := `
		UPDATE receipts
		SET status = :status,
		    description = :description,
		    items = :items,
		    notes = :notes,
		    updated_at = :updated_at
		WHERE id = :id
	`

	result, err := r.db.NamedExec(query, map[string]interface{}{
		"id":          receipt.ID,
		"status":      receipt.Status,
		"description": receipt.Description,
		"items":       itemsJSON,
		"notes":       receipt.Notes,
		"updated_at":  receipt.UpdatedAt,
	})

	if err != nil {
		return fmt.Errorf("failed to update receipt: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("receipt not found: %s", receipt.ID)
	}

	return nil
}

// rowToReceipt converts a database row to a Receipt entity
func (r *ReceiptRepository) rowToReceipt(row receiptRow) (payment.Receipt, error) {
	// Unmarshal items
	var items []payment.ReceiptItem
	if len(row.Items) > 0 {
		if err := json.Unmarshal(row.Items, &items); err != nil {
			return payment.Receipt{}, fmt.Errorf("failed to unmarshal items: %w", err)
		}
	}

	return payment.Receipt{
		ID:            row.ID,
		PaymentID:     row.PaymentID,
		ReceiptNumber: row.ReceiptNumber,
		MemberID:      row.MemberID,
		Amount:        row.Amount,
		Currency:      row.Currency,
		PaymentType:   payment.PaymentType(row.PaymentType),
		PaymentMethod: payment.PaymentMethod(row.PaymentMethod),
		TransactionID: row.TransactionID,
		PaymentDate:   sqlutil.NullTimeToTime(row.PaymentDate),
		ReceiptDate:   sqlutil.NullTimeToTime(row.ReceiptDate),
		Status:        payment.Status(row.Status),
		Description:   sqlutil.NullStringToString(row.Description),
		MemberName:    row.MemberName,
		MemberEmail:   row.MemberEmail,
		CardMask:      sqlutil.NullStringToPtr(row.CardMask),
		Items:         items,
		TaxAmount:     row.TaxAmount,
		TotalAmount:   row.TotalAmount,
		Notes:         sqlutil.NullStringToString(row.Notes),
		CreatedAt:     sqlutil.NullTimeToTime(row.CreatedAt),
		UpdatedAt:     sqlutil.NullTimeToTime(row.UpdatedAt),
	}, nil
}

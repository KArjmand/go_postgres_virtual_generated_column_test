package postgres

import (
	"context"
	"database/sql"

	"github.com/KArjmand/go_postgres_virtual_generated_column_test/pkg/invoices/domain/invoices"
)

// Repository implements invoices.Repository using PostgreSQL
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new PostgreSQL repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// FindAllWithVirtual returns invoices with pre-computed total from DB
func (r *Repository) FindAllWithVirtual(ctx context.Context, limit int) ([]*invoices.Invoice, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, customer_id, amount_cents, tax_rate, total_cents
		FROM invoices_with_virtual
		ORDER BY id
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*invoices.Invoice, 0, limit)
	for rows.Next() {
		var id int64
		var customerID int64
		var amountCents int64
		var taxRate float64
		var totalCents int64

		if err := rows.Scan(&id, &customerID, &amountCents, &taxRate, &totalCents); err != nil {
			return nil, err
		}

		result = append(result, invoices.NewInvoice(
			invoices.ID(id),
			customerID,
			amountCents,
			taxRate,
			totalCents,
		))
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// FindAllWithoutVirtual returns invoices without pre-computed total
func (r *Repository) FindAllWithoutVirtual(ctx context.Context, limit int) ([]*invoices.Invoice, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, customer_id, amount_cents, tax_rate
		FROM invoices_without_virtual
		ORDER BY id
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]*invoices.Invoice, 0, limit)
	for rows.Next() {
		var id int64
		var customerID int64
		var amountCents int64
		var taxRate float64

		if err := rows.Scan(&id, &customerID, &amountCents, &taxRate); err != nil {
			return nil, err
		}

		// Calculate total in Go
		result = append(result, invoices.NewInvoiceWithCalculation(
			invoices.ID(id),
			customerID,
			amountCents,
			taxRate,
		))
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// CountWithVirtual returns the count of invoices in the virtual table
func (r *Repository) CountWithVirtual(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM invoices_with_virtual").Scan(&count)
	return count, err
}

// CountWithoutVirtual returns the count of invoices in the non-virtual table
func (r *Repository) CountWithoutVirtual(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM invoices_without_virtual").Scan(&count)
	return count, err
}

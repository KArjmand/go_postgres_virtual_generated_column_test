package invoices

import "context"

// Repository defines the interface for invoice persistence
type Repository interface {
	// FindAllWithVirtual returns invoices with pre-computed total from DB
	FindAllWithVirtual(ctx context.Context, limit int) ([]*Invoice, error)

	// FindAllWithoutVirtual returns invoices without pre-computed total
	// The caller is responsible for calculating the total
	FindAllWithoutVirtual(ctx context.Context, limit int) ([]*Invoice, error)

	// CountWithVirtual returns the count of invoices in the virtual table
	CountWithVirtual(ctx context.Context) (int64, error)

	// CountWithoutVirtual returns the count of invoices in the non-virtual table
	CountWithoutVirtual(ctx context.Context) (int64, error)
}

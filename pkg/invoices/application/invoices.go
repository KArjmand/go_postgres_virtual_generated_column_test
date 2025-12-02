package application

import (
	"context"
	"runtime"
	"time"

	"github.com/KArjmand/go_postgres_virtual_generated_column_test/pkg/invoices/domain/invoices"
)

// InvoicesService handles invoice-related business logic
type InvoicesService struct {
	repository invoices.Repository
}

// NewInvoicesService creates a new InvoicesService
func NewInvoicesService(repository invoices.Repository) InvoicesService {
	return InvoicesService{repository: repository}
}

// InvoicesResult contains invoices and performance metrics
type InvoicesResult struct {
	Invoices []*invoices.Invoice
	Metrics  invoices.QueryMetrics
}

// GetInvoicesWithVirtual retrieves invoices using PostgreSQL virtual generated column
func (s InvoicesService) GetInvoicesWithVirtual(ctx context.Context, limit int) (InvoicesResult, error) {
	totalStart := time.Now()

	var memStart runtime.MemStats
	runtime.ReadMemStats(&memStart)

	queryStart := time.Now()
	invs, err := s.repository.FindAllWithVirtual(ctx, limit)
	queryDuration := time.Since(queryStart)

	if err != nil {
		return InvoicesResult{}, err
	}

	var memEnd runtime.MemStats
	runtime.ReadMemStats(&memEnd)

	totalDuration := time.Since(totalStart)

	// Handle GC causing memEnd < memStart
	var memUsed uint64
	if memEnd.Alloc > memStart.Alloc {
		memUsed = memEnd.Alloc - memStart.Alloc
	}

	return InvoicesResult{
		Invoices: invs,
		Metrics:  invoices.NewQueryMetrics(queryDuration, totalDuration, memUsed),
	}, nil
}

// GetInvoicesWithCalculation retrieves invoices and calculates total in Go
func (s InvoicesService) GetInvoicesWithCalculation(ctx context.Context, limit int) (InvoicesResult, error) {
	totalStart := time.Now()

	var memStart runtime.MemStats
	runtime.ReadMemStats(&memStart)

	queryStart := time.Now()
	invs, err := s.repository.FindAllWithoutVirtual(ctx, limit)
	queryDuration := time.Since(queryStart)

	if err != nil {
		return InvoicesResult{}, err
	}

	var memEnd runtime.MemStats
	runtime.ReadMemStats(&memEnd)

	totalDuration := time.Since(totalStart)

	// Handle GC causing memEnd < memStart
	var memUsed uint64
	if memEnd.Alloc > memStart.Alloc {
		memUsed = memEnd.Alloc - memStart.Alloc
	}

	return InvoicesResult{
		Invoices: invs,
		Metrics:  invoices.NewQueryMetrics(queryDuration, totalDuration, memUsed),
	}, nil
}

// StatsResult contains table statistics
type StatsResult struct {
	WithVirtualCount    int64
	WithoutVirtualCount int64
}

// GetStats returns row counts for both tables
func (s InvoicesService) GetStats(ctx context.Context) (StatsResult, error) {
	withCount, err := s.repository.CountWithVirtual(ctx)
	if err != nil {
		return StatsResult{}, err
	}

	withoutCount, err := s.repository.CountWithoutVirtual(ctx)
	if err != nil {
		return StatsResult{}, err
	}

	return StatsResult{
		WithVirtualCount:    withCount,
		WithoutVirtualCount: withoutCount,
	}, nil
}

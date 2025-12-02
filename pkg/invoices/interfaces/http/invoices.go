package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	common_http "github.com/KArjmand/go_postgres_virtual_generated_column_test/pkg/common/http"
	"github.com/KArjmand/go_postgres_virtual_generated_column_test/pkg/invoices/application"
	"github.com/KArjmand/go_postgres_virtual_generated_column_test/pkg/invoices/domain/invoices"
)

const defaultLimit = 10000

// AddRoutes registers invoice routes on the router
func AddRoutes(mux *http.ServeMux, service application.InvoicesService) {
	resource := invoicesResource{service: service}

	mux.HandleFunc("/api/invoices/virtual", resource.GetWithVirtual)
	mux.HandleFunc("/api/invoices/calculated", resource.GetWithCalculation)
	mux.HandleFunc("/api/benchmark", resource.Benchmark)
	mux.HandleFunc("/api/stats", resource.GetStats)
	mux.HandleFunc("/health", resource.HealthCheck)
}

type invoicesResource struct {
	service application.InvoicesService
}

// InvoiceView represents an invoice in API responses
type InvoiceView struct {
	ID          int64   `json:"id"`
	CustomerID  int64   `json:"customer_id"`
	AmountCents int64   `json:"amount_cents"`
	TaxRate     float64 `json:"tax_rate"`
	TotalCents  int64   `json:"total_cents"`
}

// InvoicesResponse wraps the API response with metrics
type InvoicesResponse struct {
	Data        []InvoiceView `json:"data"`
	Count       int           `json:"count"`
	QueryTimeMs float64       `json:"query_time_ms"`
	TotalTimeMs float64       `json:"total_time_ms"`
	CPUTimeNs   int64         `json:"cpu_time_ns"`
	MemoryBytes uint64        `json:"memory_bytes"`
}

func toInvoiceViews(invs []*invoices.Invoice) []InvoiceView {
	views := make([]InvoiceView, len(invs))
	for i, inv := range invs {
		views[i] = InvoiceView{
			ID:          int64(inv.ID()),
			CustomerID:  inv.CustomerID(),
			AmountCents: inv.AmountCents(),
			TaxRate:     inv.TaxRate(),
			TotalCents:  inv.TotalCents(),
		}
	}
	return views
}

func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func (r invoicesResource) GetWithVirtual(w http.ResponseWriter, req *http.Request) {
	result, err := r.service.GetInvoicesWithVirtual(req.Context(), defaultLimit)
	if err != nil {
		common_http.ErrInternal(w, err)
		return
	}

	writeJSON(w, InvoicesResponse{
		Data:        toInvoiceViews(result.Invoices),
		Count:       len(result.Invoices),
		QueryTimeMs: result.Metrics.QueryTimeMs(),
		TotalTimeMs: result.Metrics.TotalTimeMs(),
		CPUTimeNs:   result.Metrics.CPUTimeNs(),
		MemoryBytes: result.Metrics.MemoryBytes,
	})
}

func (r invoicesResource) GetWithCalculation(w http.ResponseWriter, req *http.Request) {
	result, err := r.service.GetInvoicesWithCalculation(req.Context(), defaultLimit)
	if err != nil {
		common_http.ErrInternal(w, err)
		return
	}

	writeJSON(w, InvoicesResponse{
		Data:        toInvoiceViews(result.Invoices),
		Count:       len(result.Invoices),
		QueryTimeMs: result.Metrics.QueryTimeMs(),
		TotalTimeMs: result.Metrics.TotalTimeMs(),
		CPUTimeNs:   result.Metrics.CPUTimeNs(),
		MemoryBytes: result.Metrics.MemoryBytes,
	})
}

// StatsResponse represents table statistics
type StatsResponse struct {
	InvoicesWithVirtualCount    int64 `json:"invoices_with_virtual_count"`
	InvoicesWithoutVirtualCount int64 `json:"invoices_without_virtual_count"`
}

func (r invoicesResource) GetStats(w http.ResponseWriter, req *http.Request) {
	stats, err := r.service.GetStats(req.Context())
	if err != nil {
		common_http.ErrInternal(w, err)
		return
	}

	writeJSON(w, StatsResponse{
		InvoicesWithVirtualCount:    stats.WithVirtualCount,
		InvoicesWithoutVirtualCount: stats.WithoutVirtualCount,
	})
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status string `json:"status"`
}

func (r invoicesResource) HealthCheck(w http.ResponseWriter, req *http.Request) {
	writeJSON(w, HealthResponse{Status: "ok"})
}

// BenchmarkMetrics holds metrics for a single test run
type BenchmarkMetrics struct {
	QueryTimeMs   float64 `json:"query_time_ms"`
	TotalTimeMs   float64 `json:"total_time_ms"`
	CPUTimeNs     int64   `json:"cpu_time_ns"`
	MemoryBytes   uint64  `json:"memory_bytes"`
	RowCount      int     `json:"row_count"`
	ResponseBytes int     `json:"response_bytes"`
}

// BenchmarkComparison compares virtual vs calculated approaches
type BenchmarkComparison struct {
	Virtual    BenchmarkMetrics `json:"virtual"`
	Calculated BenchmarkMetrics `json:"calculated"`
	Comparison struct {
		QueryTimeDiffMs   float64 `json:"query_time_diff_ms"`
		QueryTimeDiffPct  float64 `json:"query_time_diff_pct"`
		TotalTimeDiffMs   float64 `json:"total_time_diff_ms"`
		TotalTimeDiffPct  float64 `json:"total_time_diff_pct"`
		MemoryDiffBytes   int64   `json:"memory_diff_bytes"`
		MemoryDiffPct     float64 `json:"memory_diff_pct"`
		ResponseDiffBytes int     `json:"response_diff_bytes"`
		ResponseDiffPct   float64 `json:"response_diff_pct"`
		Winner            string  `json:"winner"`
		Summary           string  `json:"summary"`
	} `json:"comparison"`
}

func (r invoicesResource) Benchmark(w http.ResponseWriter, req *http.Request) {
	// Run virtual column test
	virtualResult, err := r.service.GetInvoicesWithVirtual(req.Context(), defaultLimit)
	if err != nil {
		common_http.ErrInternal(w, err)
		return
	}
	virtualViews := toInvoiceViews(virtualResult.Invoices)
	virtualJSON, _ := json.Marshal(virtualViews)

	// Run calculated test
	calcResult, err := r.service.GetInvoicesWithCalculation(req.Context(), defaultLimit)
	if err != nil {
		common_http.ErrInternal(w, err)
		return
	}
	calcViews := toInvoiceViews(calcResult.Invoices)
	calcJSON, _ := json.Marshal(calcViews)

	// Build comparison
	result := BenchmarkComparison{
		Virtual: BenchmarkMetrics{
			QueryTimeMs:   virtualResult.Metrics.QueryTimeMs(),
			TotalTimeMs:   virtualResult.Metrics.TotalTimeMs(),
			CPUTimeNs:     virtualResult.Metrics.CPUTimeNs(),
			MemoryBytes:   virtualResult.Metrics.MemoryBytes,
			RowCount:      len(virtualResult.Invoices),
			ResponseBytes: len(virtualJSON),
		},
		Calculated: BenchmarkMetrics{
			QueryTimeMs:   calcResult.Metrics.QueryTimeMs(),
			TotalTimeMs:   calcResult.Metrics.TotalTimeMs(),
			CPUTimeNs:     calcResult.Metrics.CPUTimeNs(),
			MemoryBytes:   calcResult.Metrics.MemoryBytes,
			RowCount:      len(calcResult.Invoices),
			ResponseBytes: len(calcJSON),
		},
	}

	// Calculate differences (positive = virtual is better)
	result.Comparison.QueryTimeDiffMs = calcResult.Metrics.QueryTimeMs() - virtualResult.Metrics.QueryTimeMs()
	result.Comparison.TotalTimeDiffMs = calcResult.Metrics.TotalTimeMs() - virtualResult.Metrics.TotalTimeMs()
	result.Comparison.MemoryDiffBytes = int64(calcResult.Metrics.MemoryBytes) - int64(virtualResult.Metrics.MemoryBytes)
	result.Comparison.ResponseDiffBytes = len(calcJSON) - len(virtualJSON)

	// Calculate percentages
	if virtualResult.Metrics.QueryTimeMs() > 0 {
		result.Comparison.QueryTimeDiffPct = (result.Comparison.QueryTimeDiffMs / virtualResult.Metrics.QueryTimeMs()) * 100
	}
	if virtualResult.Metrics.TotalTimeMs() > 0 {
		result.Comparison.TotalTimeDiffPct = (result.Comparison.TotalTimeDiffMs / virtualResult.Metrics.TotalTimeMs()) * 100
	}
	if virtualResult.Metrics.MemoryBytes > 0 {
		result.Comparison.MemoryDiffPct = (float64(result.Comparison.MemoryDiffBytes) / float64(virtualResult.Metrics.MemoryBytes)) * 100
	}
	if len(virtualJSON) > 0 {
		result.Comparison.ResponseDiffPct = (float64(result.Comparison.ResponseDiffBytes) / float64(len(virtualJSON))) * 100
	}

	// Determine winner
	virtualScore := 0
	calcScore := 0

	if result.Comparison.QueryTimeDiffMs > 0 {
		virtualScore++
	} else {
		calcScore++
	}
	if result.Comparison.TotalTimeDiffMs > 0 {
		virtualScore++
	} else {
		calcScore++
	}
	if result.Comparison.MemoryDiffBytes > 0 {
		virtualScore++
	} else {
		calcScore++
	}

	if virtualScore > calcScore {
		result.Comparison.Winner = "virtual"
	} else if calcScore > virtualScore {
		result.Comparison.Winner = "calculated"
	} else {
		result.Comparison.Winner = "tie"
	}

	result.Comparison.Summary = fmt.Sprintf(
		"Virtual: query=%.2fms, total=%.2fms, mem=%dKB, response=%dKB | "+
			"Calculated: query=%.2fms, total=%.2fms, mem=%dKB, response=%dKB | "+
			"Winner: %s",
		result.Virtual.QueryTimeMs, result.Virtual.TotalTimeMs,
		result.Virtual.MemoryBytes/1024, result.Virtual.ResponseBytes/1024,
		result.Calculated.QueryTimeMs, result.Calculated.TotalTimeMs,
		result.Calculated.MemoryBytes/1024, result.Calculated.ResponseBytes/1024,
		result.Comparison.Winner,
	)

	writeJSON(w, result)
}

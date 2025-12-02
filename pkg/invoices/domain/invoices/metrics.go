package invoices

import "time"

// QueryMetrics holds performance metrics for a query operation
type QueryMetrics struct {
	QueryDuration time.Duration
	TotalDuration time.Duration
	MemoryBytes   uint64
}

// NewQueryMetrics creates a new QueryMetrics instance
func NewQueryMetrics(queryDuration, totalDuration time.Duration, memoryBytes uint64) QueryMetrics {
	return QueryMetrics{
		QueryDuration: queryDuration,
		TotalDuration: totalDuration,
		MemoryBytes:   memoryBytes,
	}
}

// QueryTimeMs returns query time in milliseconds
func (m QueryMetrics) QueryTimeMs() float64 {
	return float64(m.QueryDuration.Microseconds()) / 1000
}

// TotalTimeMs returns total time in milliseconds
func (m QueryMetrics) TotalTimeMs() float64 {
	return float64(m.TotalDuration.Microseconds()) / 1000
}

// CPUTimeNs returns CPU time in nanoseconds
func (m QueryMetrics) CPUTimeNs() int64 {
	return m.TotalDuration.Nanoseconds()
}

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"sync/atomic"
	"time"
)

// RunSchema creates the database tables
func RunSchema(ctx context.Context, db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS invoices_with_virtual (
		id           BIGSERIAL PRIMARY KEY,
		customer_id  BIGINT NOT NULL,
		amount_cents BIGINT NOT NULL,
		tax_rate     NUMERIC(4,2) NOT NULL,
		total_cents  BIGINT GENERATED ALWAYS AS (
			ROUND(amount_cents * (1 + tax_rate))
		) STORED
	);

	CREATE TABLE IF NOT EXISTS invoices_without_virtual (
		id           BIGSERIAL PRIMARY KEY,
		customer_id  BIGINT NOT NULL,
		amount_cents BIGINT NOT NULL,
		tax_rate     NUMERIC(4,2) NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_invoices_with_virtual_customer ON invoices_with_virtual(customer_id);
	CREATE INDEX IF NOT EXISTS idx_invoices_without_virtual_customer ON invoices_without_virtual(customer_id);
	`

	_, err := db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to run schema: %w", err)
	}

	log.Println("Schema created successfully")
	return nil
}

// Seed populates the database with random data using concurrent workers
func Seed(ctx context.Context, db *sql.DB) error {
	seedCountStr := os.Getenv("SEED_COUNT")
	seedCount := 1000000000 // 1 billion default
	if seedCountStr != "" {
		if n, err := strconv.Atoi(seedCountStr); err == nil {
			seedCount = n
		}
	}

	// Number of concurrent workers (adjust based on DB capacity)
	numWorkers := 10
	if w := os.Getenv("SEED_WORKERS"); w != "" {
		if n, err := strconv.Atoi(w); err == nil {
			numWorkers = n
		}
	}

	var count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM invoices_with_virtual").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing data: %w", err)
	}

	if count >= seedCount {
		log.Printf("Data already seeded (%d rows), skipping...", count)
		return nil
	}

	remaining := seedCount - count
	log.Printf("Seeding %d rows with %d workers (existing: %d, target: %d)...", remaining, numWorkers, count, seedCount)
	start := time.Now()

	batchSize := 5000
	jobs := make(chan int, numWorkers*2)
	results := make(chan error, numWorkers)
	var inserted int64

	// Start workers
	for w := 0; w < numWorkers; w++ {
		go func(workerID int) {
			r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))
			for size := range jobs {
				if err := insertBatch(ctx, db, r, size); err != nil {
					results <- err
					return
				}
				atomic.AddInt64(&inserted, int64(size))
			}
			results <- nil
		}(w)
	}

	// Progress reporter
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				ins := atomic.LoadInt64(&inserted)
				elapsed := time.Since(start)
				rate := float64(ins) / elapsed.Seconds()
				eta := time.Duration(float64(remaining-int(ins)) / rate * float64(time.Second))
				log.Printf("Inserted %d/%d rows (%.0f rows/sec, ETA: %v)", ins, remaining, rate, eta.Round(time.Second))
			case <-done:
				return
			}
		}
	}()

	// Send jobs
	for i := 0; i < remaining; i += batchSize {
		jobs <- min(batchSize, remaining-i)
	}
	close(jobs)

	// Wait for workers
	var firstErr error
	for w := 0; w < numWorkers; w++ {
		if err := <-results; err != nil && firstErr == nil {
			firstErr = err
		}
	}
	close(done)

	if firstErr != nil {
		return firstErr
	}

	log.Printf("Seeding completed: %d rows in %v", inserted, time.Since(start))
	return nil
}

func insertBatch(ctx context.Context, db *sql.DB, r *rand.Rand, count int) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmtWith, err := tx.PrepareContext(ctx, `
		INSERT INTO invoices_with_virtual (customer_id, amount_cents, tax_rate)
		VALUES ($1, $2, $3)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmtWith.Close()

	stmtWithout, err := tx.PrepareContext(ctx, `
		INSERT INTO invoices_without_virtual (customer_id, amount_cents, tax_rate)
		VALUES ($1, $2, $3)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmtWithout.Close()

	for i := 0; i < count; i++ {
		customerID := r.Int63n(10000) + 1
		amountCents := r.Int63n(1000000) + 100
		taxRate := float64(r.Intn(25)+1) / 100

		if _, err := stmtWith.ExecContext(ctx, customerID, amountCents, taxRate); err != nil {
			return fmt.Errorf("failed to insert into invoices_with_virtual: %w", err)
		}

		if _, err := stmtWithout.ExecContext(ctx, customerID, amountCents, taxRate); err != nil {
			return fmt.Errorf("failed to insert into invoices_without_virtual: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

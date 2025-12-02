package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/KArjmand/go_postgres_virtual_generated_column_test/pkg/common/cmd"
	"github.com/KArjmand/go_postgres_virtual_generated_column_test/pkg/invoices/application"
	"github.com/KArjmand/go_postgres_virtual_generated_column_test/pkg/invoices/infrastructure/postgres"
	invoices_http "github.com/KArjmand/go_postgres_virtual_generated_column_test/pkg/invoices/interfaces/http"
)

func main() {
	log.Println("Starting PostgreSQL Virtual Generated Column Test Server")

	// Load .env file if exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	ctx := cmd.Context()

	// Initialize database connection
	dbConfig := postgres.ConfigFromEnv()
	db, err := postgres.NewConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run schema
	if err := postgres.RunSchema(context.Background(), db); err != nil {
		log.Fatalf("Failed to run schema: %v", err)
	}

	// Seed data
	if err := postgres.Seed(context.Background(), db); err != nil {
		log.Fatalf("Failed to seed data: %v", err)
	}

	// Create repository and service
	invoicesRepo := postgres.NewRepository(db)
	invoicesService := application.NewInvoicesService(invoicesRepo)

	// Create router and add routes
	mux := cmd.CreateRouter()
	invoices_http.AddRoutes(mux, invoicesService)

	// Get port
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: cmd.WithMiddleware(mux),
	}

	go func() {
		log.Printf("Server listening on :%s", port)
		log.Println("Routes:")
		log.Println("  GET /api/invoices/virtual    - Uses PostgreSQL virtual generated column")
		log.Println("  GET /api/invoices/calculated - Calculates total_cents in Go")
		log.Println("  GET /api/benchmark           - Compare both approaches (CPU, RAM, network)")
		log.Println("  GET /api/stats               - Table statistics")
		log.Println("  GET /health                  - Health check")

		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down server...")

	if err := server.Close(); err != nil {
		log.Printf("Error closing server: %v", err)
	}
}

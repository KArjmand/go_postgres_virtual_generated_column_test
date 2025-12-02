.PHONY: build run test up down clean

# Build the application
build:
	go build -o server ./cmd/monolith

# Run the application locally
run:
	go run ./cmd/monolith

# Run tests
test:
	go test -v ./...

# Start all services with docker compose
up:
	docker compose up --build

# Start services in background
up-d:
	docker compose up --build -d

# Stop all services
down:
	docker compose down

# Stop and remove volumes
clean:
	docker compose down -v

# Run linter
lint:
	golangci-lint run ./...

# Format code
fmt:
	go fmt ./...

# Download dependencies
deps:
	go mod download
	go mod tidy

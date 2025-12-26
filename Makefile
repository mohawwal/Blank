# Project Variables
BINARY_NAME=whatsapp-bot

# Default target
.PHONY: all
all: build

# Run the application
.PHONY: run
run:
	@echo "Running the application..."
	go run ./cmd/api

# Build the application
.PHONY: build
build:
	@echo "Building the application..."
	go build -o bin/$(BINARY_NAME) ./cmd/api

# Run database migrations
.PHONY: migrate-up
migrate-up:
	@echo "Running database migrations..."
	go run internal/database/migrate_up.go

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning up..."
	rm -rf bin
	go clean

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Tidy options
.PHONY: tidy
tidy:
	@echo "Tidying module dependencies..."
	go mod tidy

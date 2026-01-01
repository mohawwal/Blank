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

# Run terminal chat mode
.PHONY: chat
chat:
	@echo "Starting terminal chat mode..."
	go run ./cmd/terminal

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

# Ngrok tunnel for webhooks (requires ngrok installed)
.PHONY: ngrok
ngrok:
	@echo "Starting ngrok tunnel on port 2342..."
	ngrok http 2342

# Ngrok with custom port
.PHONY: ngrok-port
ngrok-port:
	@echo "Starting ngrok tunnel on custom port..."
	@read -p "Enter port number: " port; \
	ngrok http $$port

# Run app and ngrok in parallel (requires tmux or run in separate terminals)
.PHONY: dev
dev:
	@echo "Starting development environment..."
	@echo "Run 'make run' in one terminal and 'make ngrok' in another"
	@echo "Or install tmux and use 'make dev-tmux'"

# Development with tmux (requires tmux installed)
.PHONY: dev-tmux
dev-tmux:
	@echo "Starting app and ngrok in tmux..."
	tmux new-session -d -s whatsapp-bot 'make run'
	tmux split-window -h 'make ngrok'
	tmux attach-session -t whatsapp-bot

# Stop tmux session
.PHONY: stop-tmux
stop-tmux:
	@echo "Stopping tmux session..."
	tmux kill-session -t whatsapp-bot || echo "No tmux session found"

# Kill the running server on port 2342
.PHONY: kill
kill:
	@echo "Stopping server on port 2342..."
	@lsof -ti:2342 | xargs kill -9 2>/dev/null && echo "Server stopped successfully" || echo "No server running on port 2342"

# Restart the server (kill + build + run in background)
.PHONY: restart
restart: kill build
	@echo "Starting server in background..."
	@./bin/$(BINARY_NAME) &
	@sleep 1
	@echo "Server restarted successfully on port 2342"

# Makefile for Project Referee

.PHONY: all build test run clean docker-up docker-down

# Set Go binary name
BINARY_NAME=referee

all: test build

# Build the Go application
build:
	@echo "Building Go binary..."
	@go build -o ./bin/$(BINARY_NAME) ./cmd/referee

# Run tests with coverage
test:
	@echo "Running tests..."
	@go test -v -cover ./...

# Run the application (requires config.yaml)
run: build
	@echo "Starting Referee..."
	@./bin/$(BINARY_NAME)

# Clean up binaries
clean:
	@echo "Cleaning up..."
	@go clean
	@rm -f ./bin/$(BINARY_NAME)

# Start docker-compose services in the background
docker-up:
	@echo "Starting Docker services (Postgres & Metabase)..."
	@docker-compose up -d

# Stop and remove docker-compose services
docker-down:
	@echo "Stopping Docker services..."
	@docker-compose down

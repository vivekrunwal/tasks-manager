.PHONY: build dev test lint migrate-up migrate-down compose-up compose-down

# Default target
all: build

# Build the application
build:
	go build -o ./bin/task-svc ./cmd/task-svc

# Run the application for development
dev:
	@if command -v air > /dev/null; then \
		air; \
	else \
		go run ./cmd/task-svc; \
	fi

# Run tests
test:
	go test -v -race -cover ./...

# Run linting
lint:
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Please install it: https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi

# Run database migrations up
migrate-up:
	@echo "Running migrations up..."
	cat db/migrations/0001_init.up.sql | docker exec -i task-svc-postgres psql -U postgres -d tasks

# Run database migrations down
migrate-down:
	@echo "Running migrations down..."
	cat db/migrations/0001_init.down.sql | docker exec -i task-svc-postgres psql -U postgres -d tasks

# Start docker-compose services
compose-up:
	docker-compose up -d

# Stop docker-compose services
compose-down:
	docker-compose down

# Clean up build artifacts
clean:
	rm -rf ./bin

# Set default DB_DSN if not already set
DB_DSN ?= postgres://postgres:postgres@localhost:5432/tasks?sslmode=disable

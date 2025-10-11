# ABAC System Makefile

.PHONY: help setup-db migrate test test-storage test-integration test-all benchmark clean docker-up docker-down

# Default target
help:
	@echo "ABAC System - Make Commands"
	@echo "=========================="
	@echo ""
	@echo "Database Setup:"
	@echo "  docker-up      - Start PostgreSQL with Docker Compose"
	@echo "  docker-down    - Stop PostgreSQL containers"
	@echo "  setup-db       - Create databases (main and test)"
	@echo "  migrate        - Run database migration and seed data"
	@echo ""
	@echo "Testing:"
	@echo "  test           - Run all tests"
	@echo "  test-storage   - Run storage layer tests only"
	@echo "  test-integration - Run integration tests only"
	@echo "  test-postgresql  - Run PostgreSQL specific tests"
	@echo "  benchmark      - Run benchmark tests"
	@echo ""
	@echo "Development:"
	@echo "  run            - Run the main application"
	@echo "  clean          - Clean test databases and temporary files"
	@echo "  deps           - Install/update dependencies"

# Docker commands
docker-up:
	@echo "🐳 Starting PostgreSQL with Docker..."
	docker-compose up -d
	@echo "⏳ Waiting for PostgreSQL to be ready..."
	@sleep 5
	@docker-compose exec postgres pg_isready -U postgres || (echo "❌ PostgreSQL not ready" && exit 1)
	@echo "✅ PostgreSQL is ready"

docker-down:
	@echo "🐳 Stopping PostgreSQL containers..."
	docker-compose down

# Database setup
setup-db:
	@echo "🗄️ Setting up databases..."
	@./scripts/setup-test-db.sh

migrate: setup-db
	@echo "🔄 Running database migration and seeding..."
	@go run cmd/migrate/main.go

# Testing
deps:
	@echo "📦 Installing dependencies..."
	@go mod tidy
	@go mod download

test: deps
	@echo "🧪 Running all tests..."
	@go test ./... -v

test-storage: deps
	@echo "🧪 Running storage tests..."
	@go test ./storage -v

test-integration: deps
	@echo "🧪 Running integration tests..."
	@go test -run Integration -v

test-postgresql: deps
	@echo "🧪 Running PostgreSQL specific tests..."
	@go test -run PostgreSQL -v

benchmark: deps
	@echo "⚡ Running benchmarks..."
	@go test -bench=. -benchmem -v

# Development
run: migrate
	@echo "🚀 Running ABAC application..."
	@go run main.go

# Cleanup
clean:
	@echo "🧹 Cleaning up..."
	@if command -v psql >/dev/null 2>&1; then \
		PGPASSWORD=postgres psql -h localhost -U postgres -c "DROP DATABASE IF EXISTS abac_test;" 2>/dev/null || true; \
	fi
	@go clean -testcache
	@rm -f *.log
	@echo "✅ Cleanup complete"

# Full setup from scratch
setup: docker-up setup-db migrate
	@echo "🎉 Full setup complete! Ready to run tests and application."

# Quick test cycle
test-quick: setup-db
	@echo "⚡ Quick test cycle..."
	@go test ./storage -v -short
	@go test -run PostgreSQL -v -short

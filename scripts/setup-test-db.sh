#!/bin/bash

# Test Database Setup Script for ABAC System
# This script sets up test databases for running tests

set -e

# Configuration
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_USER=${DB_USER:-postgres}
DB_PASSWORD=${DB_PASSWORD:-postgres}
TEST_DB_NAME=${TEST_DB_NAME:-abac_test}
MAIN_DB_NAME=${MAIN_DB_NAME:-abac_system}

echo "🧪 Setting up ABAC test databases..."
echo "Host: $DB_HOST:$DB_PORT"
echo "User: $DB_USER"
echo "Test DB: $TEST_DB_NAME"
echo "Main DB: $MAIN_DB_NAME"

# Function to run SQL command
run_sql() {
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -c "$1"
}

# Function to check if database exists
db_exists() {
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -tAc "SELECT 1 FROM pg_database WHERE datname='$1'" | grep -q 1
}

# Check PostgreSQL connection
echo "🔍 Checking PostgreSQL connection..."
if ! PGPASSWORD=$DB_PASSWORD pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER; then
    echo "❌ Cannot connect to PostgreSQL at $DB_HOST:$DB_PORT"
    echo "Please ensure PostgreSQL is running and accessible."
    echo "You can start it with: docker-compose up -d"
    exit 1
fi
echo "✅ PostgreSQL connection successful"

# Create main database if it doesn't exist
echo "🗄️  Setting up main database: $MAIN_DB_NAME"
if db_exists $MAIN_DB_NAME; then
    echo "✅ Main database $MAIN_DB_NAME already exists"
else
    run_sql "CREATE DATABASE $MAIN_DB_NAME;"
    echo "✅ Created main database: $MAIN_DB_NAME"
fi

# Create test database if it doesn't exist
echo "🧪 Setting up test database: $TEST_DB_NAME"
if db_exists $TEST_DB_NAME; then
    echo "⚠️  Test database $TEST_DB_NAME already exists, dropping and recreating..."
    run_sql "DROP DATABASE $TEST_DB_NAME;"
fi

run_sql "CREATE DATABASE $TEST_DB_NAME;"
echo "✅ Created test database: $TEST_DB_NAME"

# Set environment variables for tests
export DB_HOST=$DB_HOST
export DB_PORT=$DB_PORT
export DB_USER=$DB_USER
export DB_PASSWORD=$DB_PASSWORD
export DB_NAME=$MAIN_DB_NAME
export TEST_DB_HOST=$DB_HOST
export TEST_DB_PORT=$DB_PORT
export TEST_DB_USER=$DB_USER
export TEST_DB_PASSWORD=$DB_PASSWORD
export TEST_DB_NAME=$TEST_DB_NAME

echo ""
echo "🎯 Database setup complete!"
echo ""
echo "Environment variables set:"
echo "  DB_NAME=$MAIN_DB_NAME"
echo "  TEST_DB_NAME=$TEST_DB_NAME"
echo ""
echo "Next steps:"
echo "1. Run migration for main database:"
echo "   go run cmd/migrate/main.go"
echo ""
echo "2. Run tests:"
echo "   go test ./..."
echo ""
echo "3. Run specific PostgreSQL tests:"
echo "   go test ./storage -v"
echo "   go test -run PostgreSQL -v"
echo ""
echo "4. Run benchmarks:"
echo "   go test -bench=. -benchmem"

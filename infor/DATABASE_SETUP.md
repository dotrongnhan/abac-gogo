# PostgreSQL Database Setup

This document describes how to set up and use PostgreSQL database with the ABAC system.

## Prerequisites

1. PostgreSQL server installed and running
2. Go modules initialized with GORM dependencies

## Environment Variables

Create a `.env` file or set the following environment variables:

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=abac_system
DB_SSL_MODE=disable
DB_TIMEZONE=UTC

# Database Connection Pool Settings (optional)
DB_MAX_IDLE_CONNS=10
DB_MAX_OPEN_CONNS=100
DB_CONN_MAX_LIFETIME=3600

# Database Logging (optional)
DB_LOG_LEVEL=info
```

## Database Setup

1. **Create Database**:
   ```sql
   CREATE DATABASE abac_system;
   ```

2. **Run Migration and Seed Data**:
   ```bash
   go run cmd/migrate/main.go
   ```

   This will:
   - Create all necessary tables with proper indexes
   - Seed data from existing JSON files (subjects.json, resources.json, actions.json, policies.json)

3. **Run the Application**:
   ```bash
   go run main.go
   ```

## Database Schema

The system creates the following tables:

- **subjects**: User, service, or application entities
- **resources**: API endpoints, documents, or data objects  
- **actions**: Operations that can be performed
- **policies**: Access control policies with rules
- **audit_logs**: Audit trail of all evaluations

## Features

- **JSONB Support**: Complex attributes and metadata stored as JSONB for efficient querying
- **Indexing**: Proper indexes on frequently queried fields
- **Connection Pooling**: Configurable connection pool settings
- **Auto-Migration**: GORM auto-migration ensures schema is up to date
- **Audit Logging**: Built-in audit trail with PostgreSQL storage

## Development

To reset the database and reseed data:

```bash
# Drop and recreate database
psql -U postgres -c "DROP DATABASE IF EXISTS abac_system;"
psql -U postgres -c "CREATE DATABASE abac_system;"

# Run migration and seeding again
go run cmd/migrate/main.go
```

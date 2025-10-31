# Database Migrations

This directory contains SQL migration files for the ABAC database schema.

## Migration Files

### 002 - User Schema
**File**: `002_user_schema.sql`

**Purpose**: Creates the user management schema including users, profiles, companies, departments, positions, and roles.

**Rollback**: `002_user_schema_rollback.sql`

### 003 - User Seed Data
**File**: `003_user_seed_data.sql`

**Purpose**: Seeds initial user data for testing and development.

## Running Migrations

### Using Make (Recommended)

```bash
# Run all migrations and seeding
make migrate
```

### Using Docker

```bash
# If database is in Docker container
docker-compose exec postgres psql -U postgres -d abac_db -f /docker-entrypoint-initdb.d/migrations/002_user_schema.sql
docker-compose exec postgres psql -U postgres -d abac_db -f /docker-entrypoint-initdb.d/migrations/003_user_seed_data.sql
```

### Using psql directly

```bash
# If database is on localhost
psql -h localhost -U postgres -d abac_db -f migrations/002_user_schema.sql
psql -h localhost -U postgres -d abac_db -f migrations/003_user_seed_data.sql
```

## Migration Order

1. **`002_user_schema.sql`** - User management schema
2. **`003_user_seed_data.sql`** - Initial user data

## Rollback

To rollback the user schema:

```bash
psql -h localhost -U postgres -d abac_db -f migrations/002_user_schema_rollback.sql
```

## Testing After Migration

Run tests to verify:

```bash
go test ./... -v
```

All tests should pass after migrations are applied.

## Notes

- Migrations are applied in order (002, 003, etc.)
- The system uses **Deny-Override algorithm** for policy evaluation (any Deny = immediate Deny)
- No priority field needed for policy ordering as Deny-Override doesn't use it

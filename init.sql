-- Initialize PostgreSQL for ABAC System
-- This script is executed when the PostgreSQL container starts

-- Enable JSONB extensions if needed
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Set timezone
SET timezone = 'UTC';

-- Create database if it doesn't exist (though it should be created by POSTGRES_DB)
-- The database is already created by the POSTGRES_DB environment variable

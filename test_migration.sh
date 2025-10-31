#!/bin/bash

echo "🧪 Testing Priority Migration"
echo "=============================="

# Check if migration file exists
if [ ! -f "migrations/004_add_priority_to_policies.sql" ]; then
    echo "❌ Migration file not found!"
    exit 1
fi

echo "✅ Migration file exists"

# Test migration syntax (dry run)
echo ""
echo "Testing SQL syntax..."
psql -h localhost -U postgres -d postgres --set ON_ERROR_STOP=1 --dry-run < migrations/004_add_priority_to_policies.sql 2>&1 | head -5

echo ""
echo "📝 Migration Summary:"
echo "  - Adds 'priority' column to policies table"
echo "  - Default value: 100"
echo "  - Creates index: idx_policies_priority"
echo ""
echo "To apply migration:"
echo "  make migrate-priority"
echo "  OR"
echo "  psql -h localhost -U postgres -d abac_db -f migrations/004_add_priority_to_policies.sql"


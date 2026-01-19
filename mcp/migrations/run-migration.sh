#!/usr/bin/env bash
# Run a database migration on the team-memory.db

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DB_PATH="$SCRIPT_DIR/../team-memory.db"
MIGRATION_FILE="$1"

if [ -z "$MIGRATION_FILE" ]; then
    echo "Usage: $0 <migration-file.sql>"
    echo "Example: $0 001_add_implementation_plan_categories.sql"
    exit 1
fi

MIGRATION_PATH="$SCRIPT_DIR/$MIGRATION_FILE"

if [ ! -f "$MIGRATION_PATH" ]; then
    echo "Error: Migration file not found: $MIGRATION_PATH"
    exit 1
fi

if [ ! -f "$DB_PATH" ]; then
    echo "Error: Database file not found: $DB_PATH"
    exit 1
fi

echo "Running migration: $MIGRATION_FILE"
echo "Database: $DB_PATH"
echo ""

# Create a backup before running migration
BACKUP_PATH="$DB_PATH.backup-$(date +%Y%m%d-%H%M%S)"
echo "Creating backup: $BACKUP_PATH"
cp "$DB_PATH" "$BACKUP_PATH"

# Run the migration
echo "Executing migration..."
sqlite3 "$DB_PATH" < "$MIGRATION_PATH"

if [ $? -eq 0 ]; then
    echo ""
    echo "✓ Migration completed successfully!"
    echo "Backup saved at: $BACKUP_PATH"
else
    echo ""
    echo "✗ Migration failed!"
    echo "Restoring from backup..."
    cp "$BACKUP_PATH" "$DB_PATH"
    echo "Database restored from backup"
    exit 1
fi

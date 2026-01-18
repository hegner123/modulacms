#!/usr/bin/env bash

# Script to create a fresh SQLite database from schema and demo data
# Usage: ./create_db.sh [database_name]
# Default database name: modula.db

set -e  # Exit on error

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Configuration
DB_NAME="${1:-modula.db}"
SCHEMA_FILE="sql/all_schema.sql"
DEMO_DATA_FILE="all_schema/demo_data.sql"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Change to project directory
cd "$SCRIPT_DIR"

# Validate files exist
if [ ! -f "$SCHEMA_FILE" ]; then
    echo -e "${RED}Error: Schema file not found: $SCHEMA_FILE${NC}"
    exit 1
fi

if [ ! -f "$DEMO_DATA_FILE" ]; then
    echo -e "${RED}Error: Demo data file not found: $DEMO_DATA_FILE${NC}"
    exit 1
fi

# Check if database exists and remove it
if [ -f "$DB_NAME" ]; then
    echo -e "${YELLOW}Database $DB_NAME already exists. Removing...${NC}"
    rm "$DB_NAME"
fi

# Create database and load schema
echo -e "${GREEN}Creating database: $DB_NAME${NC}"
echo -e "${GREEN}Loading schema from: $SCHEMA_FILE${NC}"
sqlite3 "$DB_NAME" < "$SCHEMA_FILE"

if [ $? -ne 0 ]; then
    echo -e "${RED}Error: Failed to load schema${NC}"
    exit 1
fi

# Load demo data
echo -e "${GREEN}Loading demo data from: $DEMO_DATA_FILE${NC}"
sqlite3 "$DB_NAME" < "$DEMO_DATA_FILE"

if [ $? -ne 0 ]; then
    echo -e "${RED}Error: Failed to load demo data${NC}"
    exit 1
fi

# Verify database was created
if [ -f "$DB_NAME" ]; then
    DB_SIZE=$(du -h "$DB_NAME" | cut -f1)
    TABLE_COUNT=$(sqlite3 "$DB_NAME" "SELECT COUNT(*) FROM sqlite_master WHERE type='table';")
    echo -e "${GREEN}✓ Database created successfully!${NC}"
    echo -e "  Database: $DB_NAME"
    echo -e "  Size: $DB_SIZE"
    echo -e "  Tables: $TABLE_COUNT"
else
    echo -e "${RED}Error: Database file was not created${NC}"
    exit 1
fi

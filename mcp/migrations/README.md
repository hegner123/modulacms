# Team Memory Database Migrations

This directory contains database migration scripts for the team-memory MCP server.

## Running a Migration

To run a migration:

```bash
cd /Users/home/Documents/Code/Go_dev/modulacms/mcp/migrations
./run-migration.sh 001_add_implementation_plan_categories.sql
```

The migration script will:
1. Create a timestamped backup of your database
2. Run the migration
3. If the migration fails, automatically restore from backup

## Current Migrations

### 001_add_implementation_plan_categories.sql
Adds two new category options to the `memories` table:
- `implementation-plans` - For approved implementation plans for features
- `active-implementation-plan` - For the current plan being actively implemented

**Status**: Ready to run

## Creating New Migrations

When creating a new migration:
1. Name it with a sequential number prefix: `XXX_description.sql`
2. Always use transactions (`BEGIN TRANSACTION` ... `COMMIT`)
3. Test on a copy of the database first
4. Document what the migration does in this README

## Backup Location

Backups are created in the parent directory with the format:
`team-memory.db.backup-YYYYMMDD-HHMMSS`

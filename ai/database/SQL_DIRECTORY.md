# SQL_DIRECTORY.md

Guidelines for working with the SQL directory containing schema definitions and SQL query files.

**Directory Path:** `/Users/home/Documents/Code/Go_dev/modulacms/sql`

---

## Directory Purpose

The `sql/` directory contains:
- Database schema migration files organized in numbered directories
- SQL query files for MySQL and PostgreSQL (processed by sqlc)
- Combined schema files for each database type
- sqlc configuration
- Database-related documentation

**This directory contains .sql files, NOT .go files.**

---

## Directory Structure

```
sql/
├── sqlc.yml                    # sqlc configuration file
├── all_schema.sql              # Combined SQLite schema
├── all_schema_mysql.sql        # Combined MySQL schema
├── all_schema_psql.sql         # Combined PostgreSQL schema
├── docker-compose.yml          # Database containers for testing
├── permissions.json            # Permission definitions
├── create_order.md             # Schema creation order documentation
│
├── mysql/                      # MySQL-specific query files
│   └── *.sql                   # Query files with sqlc annotations
│
├── postgres/                   # PostgreSQL-specific query files
│   └── *.sql                   # Query files with sqlc annotations
│
└── schema/                     # Schema migration directories (numbered 1-22)
    ├── instructions.md         # Schema migration instructions
    ├── list.md                 # Schema list documentation
    ├── utility/                # Schema utility scripts
    ├── 1_permissions/          # Permissions table schema
    ├── 2_roles/                # Roles table schema
    ├── 3_media_dimension/      # Media dimensions schema
    ├── 4_users/                # Users table schema
    ├── 5_admin_routes/         # Admin routes schema
    ├── 6_routes/               # Routes table schema
    ├── 7_datatypes/            # Datatypes table schema
    ├── 8_fields/               # Fields table schema
    ├── 9_admin_datatypes/      # Admin datatypes schema
    ├── 10_admin_fields/        # Admin fields schema
    ├── 11_tokens/              # Tokens table schema
    ├── 12_user_oauth/          # OAuth table schema
    ├── 13_tables/              # Tables metadata schema
    ├── 14_media/               # Media table schema
    ├── 15_sessions/            # Sessions table schema
    ├── 16_content_data/        # Content data schema
    ├── 17_content_fields/      # Content fields schema
    ├── 18_admin_content_data/  # Admin content data schema
    ├── 19_admin_content_fields/# Admin content fields schema
    ├── 20_datatypes_fields/    # Datatypes-fields junction
    ├── 21_admin_datatypes_fields/# Admin datatypes-fields junction
    └── 22_joins/               # Join definitions and views
```

---

## Working with Schema Migrations

### Schema Organization

Schema migrations are organized in numbered directories (1-22) in `sql/schema/`. The numbering enforces dependency order.

**Rules:**
1. Lower numbers run before higher numbers
2. Each directory represents one or more related tables
3. Tables with dependencies must come after their dependencies
4. Foreign key relationships determine ordering

### Creating a New Schema Migration

**Step 1: Determine the correct number**
- Check existing schema directories (1-22 currently exist)
- New migrations should be numbered 23+
- Consider table dependencies - your new table must come after any tables it references

**Step 2: Create the directory**
```bash
mkdir -p /Users/home/Documents/Code/Go_dev/modulacms/sql/schema/23_your_table_name
```

**Step 3: Create schema files**
Each schema directory should contain SQL files for all three database types:
- `schema.sql` - SQLite version
- `schema_mysql.sql` - MySQL version
- `schema_psql.sql` - PostgreSQL version

**Example structure:**
```
sql/schema/23_your_table_name/
├── schema.sql          # SQLite CREATE TABLE
├── schema_mysql.sql    # MySQL CREATE TABLE
└── schema_psql.sql     # PostgreSQL CREATE TABLE
```

**Step 4: Write the schema**
Include:
- CREATE TABLE statement
- All columns with appropriate types
- Primary keys
- Foreign keys with proper references
- Indexes for performance
- Default values where appropriate
- NOT NULL constraints
- ON DELETE/ON UPDATE behaviors for foreign keys

**SQLite Example:**
```sql
-- sql/schema/23_example/schema.sql
CREATE TABLE IF NOT EXISTS example (
    example_id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE INDEX idx_example_user_id ON example(user_id);
```

**MySQL Example:**
```sql
-- sql/schema/23_example/schema_mysql.sql
CREATE TABLE IF NOT EXISTS example (
    example_id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at BIGINT NOT NULL DEFAULT (UNIX_TIMESTAMP()),
    updated_at BIGINT,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    INDEX idx_example_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**PostgreSQL Example:**
```sql
-- sql/schema/23_example/schema_psql.sql
CREATE TABLE IF NOT EXISTS example (
    example_id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at BIGINT NOT NULL DEFAULT EXTRACT(EPOCH FROM NOW()),
    updated_at BIGINT,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);

CREATE INDEX idx_example_user_id ON example(user_id);
```

**Step 5: Update combined schema files**
After creating the schema directory, regenerate the combined schema files:
```bash
# From project root
cd sql
# Concatenate all schema files
cat schema/*/schema.sql > all_schema.sql
cat schema/*/schema_mysql.sql > all_schema_mysql.sql
cat schema/*/schema_psql.sql > all_schema_psql.sql
```

**Step 6: Update schema documentation**
Update `sql/schema/list.md` to include your new table.

### Modifying Existing Schema

**NEVER modify existing schema files directly in production.**

For development/testing:
1. Drop your local database
2. Modify the schema file
3. Rebuild database from migrations

For production:
1. Create a new migration directory with a higher number
2. Write ALTER TABLE statements
3. Test thoroughly before deploying

---

## Working with SQL Queries (sqlc)

### What is sqlc?

sqlc generates type-safe Go code from SQL queries. You write SQL with special annotations, sqlc generates Go functions.

**Configuration:** `sql/sqlc.yml`

### Query File Organization

Queries are organized by database type:
- `sql/mysql/*.sql` - MySQL queries
- `sql/postgres/*.sql` - PostgreSQL queries

**Important:** SQLite queries are typically embedded in the db-sqlite package, not in separate .sql files.

### Writing Queries for sqlc

**Query Annotation Format:**
```sql
-- name: QueryName :returnType
```

**Return Types:**
- `:many` - Returns slice of rows (SELECT returning multiple rows)
- `:one` - Returns single row (SELECT returning one row)
- `:exec` - Returns execution result (INSERT, UPDATE, DELETE)
- `:execrows` - Returns number of rows affected

### Query Examples

**SELECT Multiple Rows:**
```sql
-- name: ListUsersByRole :many
SELECT user_id, username, email, role_id
FROM users
WHERE role_id = ?
ORDER BY username;
```

**SELECT Single Row:**
```sql
-- name: GetUserByID :one
SELECT user_id, username, email, role_id, created_at
FROM users
WHERE user_id = ?;
```

**INSERT:**
```sql
-- name: CreateUser :exec
INSERT INTO users (username, email, password_hash, role_id)
VALUES (?, ?, ?, ?);
```

**UPDATE:**
```sql
-- name: UpdateUserEmail :exec
UPDATE users
SET email = ?, updated_at = UNIX_TIMESTAMP()
WHERE user_id = ?;
```

**DELETE:**
```sql
-- name: DeleteUser :exec
DELETE FROM users
WHERE user_id = ?;
```

**Complex Query with Joins:**
```sql
-- name: GetContentTreeByRoute :many
SELECT
    cd.content_data_id,
    cd.parent_id,
    cd.first_child_id,
    cd.next_sibling_id,
    cd.prev_sibling_id,
    cd.datatype_id,
    dt.label as datatype_label,
    dt.type as datatype_type
FROM content_data cd
JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
WHERE cd.route_id = ?
ORDER BY cd.parent_id NULLS FIRST, cd.content_data_id;
```

### Parameter Placeholders

**MySQL:**
- Use `?` for parameters
- Parameters are positional

**PostgreSQL:**
- Use `$1`, `$2`, `$3` for parameters
- Parameters are numbered

**Example PostgreSQL:**
```sql
-- name: GetUserByEmailAndRole :one
SELECT user_id, username, email
FROM users
WHERE email = $1 AND role_id = $2;
```

### Generating Go Code from SQL

**Command:**
```bash
# From project root
make sqlc

# Or manually:
cd sql
sqlc generate
```

**What this does:**
1. Reads `sqlc.yml` configuration
2. Parses all .sql files in mysql/ and postgres/ directories
3. Generates type-safe Go functions
4. Outputs to locations specified in sqlc.yml

**Generated Code Location:**
Check `sqlc.yml` for output paths. Typically generates into:
- `internal/db-mysql/` for MySQL queries
- `internal/db-psql/` for PostgreSQL queries

### Adding a New Query

**Step 1: Identify the database type**
Determine if query is MySQL or PostgreSQL specific.

**Step 2: Choose or create a query file**
```bash
# For MySQL
touch sql/mysql/your_feature.sql

# For PostgreSQL
touch sql/postgres/your_feature.sql
```

**Step 3: Write the query with sqlc annotation**
```sql
-- name: YourQueryName :returnType
SELECT ...
```

**Step 4: Generate Go code**
```bash
make sqlc
```

**Step 5: Use generated function**
The generated function will be available in the database driver package (db-mysql or db-psql).

---

## Database-Specific Considerations

### SQLite
- Uses `INTEGER PRIMARY KEY AUTOINCREMENT`
- Timestamp storage: UNIX timestamp as INTEGER
- String comparison is case-insensitive by default
- Limited ALTER TABLE support
- Foreign keys must be enabled: `PRAGMA foreign_keys = ON;`

### MySQL
- Uses `INT AUTO_INCREMENT PRIMARY KEY`
- Timestamp storage: BIGINT for UNIX timestamps
- String comparison depends on collation
- Full ALTER TABLE support
- InnoDB engine recommended (supports foreign keys)
- Use `ENGINE=InnoDB DEFAULT CHARSET=utf8mb4`

### PostgreSQL
- Uses `SERIAL PRIMARY KEY` or `BIGSERIAL`
- Timestamp storage: BIGINT or native TIMESTAMP type
- Case-sensitive by default
- Most feature-rich (arrays, JSON, CTEs, window functions)
- Full ALTER TABLE support
- Best for complex queries and large datasets

---

## Common Tasks

### Task: Add a new table

1. Determine migration number (next available: 23)
2. Create schema directory: `sql/schema/23_your_table/`
3. Create three schema files (SQLite, MySQL, PostgreSQL)
4. Write CREATE TABLE statements
5. Update combined schema files
6. Update documentation

### Task: Add a new query

1. Decide database type (MySQL or PostgreSQL)
2. Create or edit .sql file in appropriate directory
3. Write query with sqlc annotation
4. Run `make sqlc` to generate Go code
5. Use generated function in Go code

### Task: Modify existing table

Development:
1. Modify schema file
2. Drop and recreate database
3. Test changes

Production:
1. Create new migration with ALTER TABLE statements
2. Test migration on staging
3. Apply to production

### Task: View all queries

```bash
# MySQL queries
ls -la /Users/home/Documents/Code/Go_dev/modulacms/sql/mysql/

# PostgreSQL queries
ls -la /Users/home/Documents/Code/Go_dev/modulacms/sql/postgres/
```

### Task: Test schema locally

```bash
# Start database containers
cd /Users/home/Documents/Code/Go_dev/modulacms/sql
docker-compose up -d

# Run application
cd /Users/home/Documents/Code/Go_dev/modulacms
make dev
./modulacms-x86 --cli
```

---

## Best Practices

### Schema Design

1. **Use appropriate data types**
   - INTEGER for IDs and counters
   - TEXT/VARCHAR for strings
   - BIGINT for UNIX timestamps
   - BLOB for binary data (avoid if possible, use S3)

2. **Always include timestamps**
   - `created_at` for record creation time
   - `updated_at` for last modification time

3. **Use foreign keys**
   - Define relationships explicitly
   - Use ON DELETE CASCADE or ON DELETE SET NULL appropriately
   - Add indexes on foreign key columns

4. **Add indexes strategically**
   - Index columns used in WHERE clauses
   - Index columns used in JOIN conditions
   - Index columns used in ORDER BY
   - Don't over-index (slows writes)

5. **Use consistent naming**
   - Table names: lowercase, plural (users, routes, content_data)
   - Primary key: `table_name_id` (user_id, route_id)
   - Foreign key: same as referenced column (user_id references users.user_id)
   - Indexes: `idx_table_column` (idx_users_email)

### Query Writing

1. **Be explicit**
   - List all columns instead of SELECT *
   - Use table aliases for clarity
   - Always specify ORDER BY if order matters

2. **Use parameters**
   - Never concatenate strings into SQL
   - Always use ? or $1 placeholders
   - Let sqlc generate safe code

3. **Optimize for performance**
   - Use indexes effectively
   - Avoid SELECT * in production
   - Limit result sets when possible
   - Use EXPLAIN to analyze query plans

4. **Handle NULL properly**
   - Use IS NULL, not = NULL
   - Use COALESCE for default values
   - Consider NULLS FIRST/LAST in ORDER BY

### Documentation

1. **Comment complex queries**
   - Explain business logic
   - Note performance considerations
   - Document assumptions

2. **Keep schema documentation updated**
   - Update list.md when adding tables
   - Document foreign key relationships
   - Note migration dependencies

---

## Troubleshooting

### sqlc generation fails

**Error: "query name must be unique"**
- Two queries have the same name
- Search for duplicate `-- name:` annotations

**Error: "column doesn't exist"**
- Schema file might not be loaded
- Check schema migration order
- Verify table/column name spelling

**Error: "syntax error"**
- SQL syntax error in query
- Check database-specific syntax (MySQL vs PostgreSQL)
- Validate SQL outside sqlc first

### Schema migration issues

**Foreign key constraint fails**
- Parent table doesn't exist yet
- Check migration numbering
- Parent table must have lower number than child

**Table already exists**
- Use `CREATE TABLE IF NOT EXISTS`
- Or check for existing tables before creating

### Database-specific issues

**SQLite: foreign key constraint failed**
```sql
PRAGMA foreign_keys = ON;
```

**MySQL: incorrect string value**
- Use `utf8mb4` charset
- Check character encoding

**PostgreSQL: permission denied**
- Check database user permissions
- Verify connection string

---

## Related Documentation

- **DB_PACKAGE.md** - Guide for working with internal/db/ Go code
- **FILE_TREE.md** - Complete directory structure
- **CLAUDE.md** - Project-wide development guidelines
- **sql/schema/instructions.md** - Schema migration instructions
- **sql/schema/list.md** - Complete table list

---

## Quick Reference

### File Extensions
- `.sql` - SQL schema or query files
- `.yml` - sqlc configuration
- `.md` - Documentation

### Key Commands
```bash
# Generate Go code from SQL
make sqlc

# Update combined schemas
cat sql/schema/*/schema.sql > sql/all_schema.sql

# Start database containers
cd sql && docker-compose up -d

# Run tests
make test
```

### Key Paths
- Schema migrations: `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/`
- MySQL queries: `/Users/home/Documents/Code/Go_dev/modulacms/sql/mysql/`
- PostgreSQL queries: `/Users/home/Documents/Code/Go_dev/modulacms/sql/postgres/`
- sqlc config: `/Users/home/Documents/Code/Go_dev/modulacms/sql/sqlc.yml`

# SQLC.md

Comprehensive documentation for sqlc - the SQL compiler that generates type-safe Go code from SQL queries.

**sqlc Repository:** https://github.com/sqlc-dev/sqlc
**sqlc Documentation:** https://docs.sqlc.dev/

---

## What is sqlc?

sqlc generates type-safe Go code from SQL. You write SQL queries with special annotations, sqlc parses them and generates Go functions with proper types, error handling, and database/sql integration.

**Benefits:**
- Type-safe database queries at compile time
- No runtime reflection or string concatenation
- Write actual SQL, not a query builder DSL
- Catches SQL errors before runtime
- Generates idiomatic Go code
- Supports multiple databases (MySQL, PostgreSQL, SQLite)

**Workflow:**
1. Write SQL queries in `.sql` files with sqlc annotations
2. Run `sqlc generate` (or `just sqlc`)
3. sqlc generates Go code with type-safe functions
4. Import and use generated functions in your code

---

## sqlc Configuration

### Configuration File

**Location:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/sqlc.yml`

**Format:** YAML

### Configuration Structure

```yaml
version: "2"
sql:
  - engine: "mysql"
    queries: "./mysql"
    schema: "./schema"
    gen:
      go:
        package: "dbmysql"
        out: "../internal/db-mysql"
        sql_package: "database/sql"
        emit_json_tags: true
        emit_interface: false
        emit_empty_slices: true
  - engine: "postgresql"
    queries: "./postgres"
    schema: "./schema"
    gen:
      go:
        package: "dbpsql"
        out: "../internal/db-psql"
        sql_package: "database/sql"
        emit_json_tags: true
        emit_interface: false
        emit_empty_slices: true
```

### Configuration Fields Explained

#### Top-Level Fields

**`version`** (required)
- Type: string
- Values: "1" or "2"
- Current: "2" (recommended)
- Description: sqlc configuration version

**`sql`** (required)
- Type: array
- Description: List of SQL generation configurations (one per database engine)

#### SQL Configuration Block

**`engine`** (required)
- Type: string
- Values: "mysql", "postgresql", "sqlite"
- Description: Database engine to target

**`queries`** (required)
- Type: string (path)
- Description: Directory containing `.sql` query files
- Example: `"./mysql"` for MySQL queries

**`schema`** (required)
- Type: string (path)
- Description: Directory containing schema definitions
- Example: `"./schema"` for migration files
- Note: Can be array of paths: `["./schema", "./other_schema"]`

**`gen`** (required)
- Type: object
- Description: Code generation configuration

#### Generation Configuration (`gen.go`)

**`package`** (required)
- Type: string
- Description: Go package name for generated code
- Example: `"dbmysql"`, `"dbpsql"`, `"dbsqlite"`

**`out`** (required)
- Type: string (path)
- Description: Output directory for generated Go files
- Example: `"../internal/db-mysql"`

**`sql_package`** (optional)
- Type: string
- Default: `"database/sql"`
- Description: Which SQL package to use
- Options:
  - `"database/sql"` - Standard library (default)
  - `"pgx/v5"` - Use pgx for PostgreSQL

**`emit_json_tags`** (optional)
- Type: boolean
- Default: false
- Description: Add JSON struct tags to generated types
- Example: `json:"user_id"`

**`emit_interface`** (optional)
- Type: boolean
- Default: false
- Description: Generate interface type for Queries struct

**`emit_empty_slices`** (optional)
- Type: boolean
- Default: false
- Description: Return empty slices instead of nil for empty results

**`emit_exact_table_names`** (optional)
- Type: boolean
- Default: false
- Description: Use exact table names for struct names (don't singularize)

**`emit_prepared_queries`** (optional)
- Type: boolean
- Default: false
- Description: Generate code using prepared statements

**`emit_pointers_for_null_types`** (optional)
- Type: boolean
- Default: false
- Description: Use pointers (*string) instead of sql.NullString

**`query_parameter_limit`** (optional)
- Type: integer
- Default: unlimited
- Description: Maximum number of parameters per query

**`rename`** (optional)
- Type: object
- Description: Override generated type/field names
- Example:
```yaml
rename:
  user_id: "ID"
  created_at: "CreatedAt"
```

**`overrides`** (optional)
- Type: array
- Description: Override column types
- Example:
```yaml
overrides:
  - column: "users.id"
    go_type: "github.com/google/uuid.UUID"
```

### Example Full Configuration

```yaml
version: "2"
sql:
  # MySQL Configuration
  - engine: "mysql"
    queries: "./mysql"
    schema:
      - "./schema"
      - "./schema/utility"
    gen:
      go:
        package: "dbmysql"
        out: "../internal/db-mysql"
        sql_package: "database/sql"
        emit_json_tags: true
        emit_interface: false
        emit_empty_slices: true
        emit_exact_table_names: false
        emit_prepared_queries: true
        query_parameter_limit: 50
        rename:
          user_id: "UserID"
          created_at: "CreatedAt"
          updated_at: "UpdatedAt"

  # PostgreSQL Configuration
  - engine: "postgresql"
    queries: "./postgres"
    schema: "./schema"
    gen:
      go:
        package: "dbpsql"
        out: "../internal/db-psql"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_interface: true
        emit_empty_slices: true
        overrides:
          - column: "users.metadata"
            go_type: "encoding/json.RawMessage"

  # SQLite Configuration (if using sqlc for SQLite)
  - engine: "sqlite"
    queries: "./sqlite"
    schema: "./schema"
    gen:
      go:
        package: "dbsqlite"
        out: "../internal/db-sqlite"
        sql_package: "database/sql"
        emit_json_tags: false
        emit_interface: false
```

---

## sqlc Query Annotations

### Annotation Format

Annotations are SQL comments that start with `-- name:` followed by the query name and return type.

**Format:**
```sql
-- name: QueryName :returnType
```

**Rules:**
- Must be on its own line
- Must come immediately before the query
- Query name must be valid Go function name (PascalCase)
- Return type determines function signature

### Return Types

#### `:many` - Multiple Rows

Returns a slice of structs. Use for SELECT queries that return multiple rows.

**Generated Signature:**
```go
func (q *Queries) QueryName(ctx context.Context, arg ArgType) ([]ReturnType, error)
```

**Example:**
```sql
-- name: ListUsers :many
SELECT user_id, username, email, role_id
FROM users
ORDER BY username;
```

**Generated:**
```go
type ListUsersRow struct {
    UserID   int64
    Username string
    Email    string
    RoleID   int64
}

func (q *Queries) ListUsers(ctx context.Context) ([]ListUsersRow, error)
```

**Use Case:**
- Getting all records
- Searching with filters
- Listing with pagination

#### `:one` - Single Row

Returns a single struct. Use for SELECT queries that return one row.

**Generated Signature:**
```go
func (q *Queries) QueryName(ctx context.Context, arg ArgType) (ReturnType, error)
```

**Example:**
```sql
-- name: GetUserByID :one
SELECT user_id, username, email, role_id, created_at
FROM users
WHERE user_id = ?;
```

**Generated:**
```go
type GetUserByIDRow struct {
    UserID    int64
    Username  string
    Email     string
    RoleID    int64
    CreatedAt int64
}

func (q *Queries) GetUserByID(ctx context.Context, userID int64) (GetUserByIDRow, error)
```

**Use Case:**
- Getting by primary key
- Finding unique record
- Fetching single result

**Important:** Returns error if no rows found (`sql.ErrNoRows`)

#### `:exec` - Execute Statement

Returns execution result. Use for INSERT, UPDATE, DELETE without returning data.

**Generated Signature:**
```go
func (q *Queries) QueryName(ctx context.Context, arg ArgType) error
```

**Example:**
```sql
-- name: DeleteUser :exec
DELETE FROM users
WHERE user_id = ?;
```

**Generated:**
```go
func (q *Queries) DeleteUser(ctx context.Context, userID int64) error
```

**Use Case:**
- DELETE operations
- UPDATE without RETURNING
- INSERT without needing the ID

#### `:execresult` - Execute with Result

Returns `sql.Result`. Use when you need LastInsertId() or RowsAffected().

**Generated Signature:**
```go
func (q *Queries) QueryName(ctx context.Context, arg ArgType) (sql.Result, error)
```

**Example:**
```sql
-- name: CreateUser :execresult
INSERT INTO users (username, email, password_hash, role_id)
VALUES (?, ?, ?, ?);
```

**Generated:**
```go
func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (sql.Result, error)
```

**Usage:**
```go
result, err := queries.CreateUser(ctx, CreateUserParams{...})
if err != nil {
    return err
}

userID, err := result.LastInsertId()
rowsAffected, err := result.RowsAffected()
```

**Use Case:**
- INSERT and get generated ID
- UPDATE and check affected rows
- DELETE and verify deletion

#### `:execrows` - Execute Returning Row Count

Returns number of affected rows as int64.

**Generated Signature:**
```go
func (q *Queries) QueryName(ctx context.Context, arg ArgType) (int64, error)
```

**Example:**
```sql
-- name: UpdateUserEmail :execrows
UPDATE users
SET email = ?, updated_at = UNIX_TIMESTAMP()
WHERE user_id = ?;
```

**Generated:**
```go
func (q *Queries) UpdateUserEmail(ctx context.Context, arg UpdateUserEmailParams) (int64, error)
```

**Use Case:**
- Verify UPDATE affected rows
- Check DELETE removed records
- Conditional logic based on affected count

#### `:batchexec` - Batch Execute

For batch operations (PostgreSQL COPY, bulk inserts).

**Generated Signature:**
```go
func (q *Queries) QueryName(ctx context.Context, args []ArgType) error
```

**Use Case:**
- Bulk inserts
- Batch updates
- High-performance bulk operations

#### `:batchone` - Batch Returning Values

Batch operation returning values per row.

**Use Case:**
- Batch insert with RETURNING clause
- Bulk operations needing results

#### `:batchmany` - Batch Returning Multiple Values

Batch operation returning multiple values per operation.

**Use Case:**
- Complex batch operations
- Bulk operations with complex returns

---

## Query Parameters

### Parameter Placeholders

**MySQL:**
- Use `?` for all parameters
- Parameters are positional

**PostgreSQL:**
- Use `$1`, `$2`, `$3`, etc.
- Parameters are numbered

**SQLite:**
- Use `?` for all parameters
- Can also use `?NNN`, `:name`, `@name`, `$name`

### Single Parameter

**MySQL:**
```sql
-- name: GetUserByID :one
SELECT * FROM users WHERE user_id = ?;
```

**Generated:**
```go
func (q *Queries) GetUserByID(ctx context.Context, userID int64) (User, error)
```

**PostgreSQL:**
```sql
-- name: GetUserByID :one
SELECT * FROM users WHERE user_id = $1;
```

**Generated:**
```go
func (q *Queries) GetUserByID(ctx context.Context, userID int64) (User, error)
```

### Multiple Parameters

**MySQL:**
```sql
-- name: GetUserByEmailAndRole :one
SELECT * FROM users
WHERE email = ? AND role_id = ?;
```

**Generated:**
```go
type GetUserByEmailAndRoleParams struct {
    Email  string
    RoleID int64
}

func (q *Queries) GetUserByEmailAndRole(ctx context.Context, arg GetUserByEmailAndRoleParams) (User, error)
```

**PostgreSQL:**
```sql
-- name: GetUserByEmailAndRole :one
SELECT * FROM users
WHERE email = $1 AND role_id = $2;
```

**Generated (same):**
```go
type GetUserByEmailAndRoleParams struct {
    Email  string
    RoleID int64
}

func (q *Queries) GetUserByEmailAndRole(ctx context.Context, arg GetUserByEmailAndRoleParams) (User, error)
```

### Named Parameters (PostgreSQL)

**Query:**
```sql
-- name: CreateUser :one
INSERT INTO users (username, email, password_hash, role_id)
VALUES ($1, $2, $3, $4)
RETURNING user_id;
```

**Generated:**
```go
type CreateUserParams struct {
    Username     string
    Email        string
    PasswordHash string
    RoleID       int64
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (int64, error)
```

### Parameter Type Inference

sqlc infers parameter types from:
1. WHERE clause comparisons
2. INSERT VALUES
3. UPDATE SET statements
4. JOIN conditions

**Example:**
```sql
-- name: UpdateUserStatus :exec
UPDATE users
SET status = ?
WHERE user_id = ? AND role_id = ?;
```

**Generated:**
```go
type UpdateUserStatusParams struct {
    Status string  // Inferred from users.status column type
    UserID int64   // Inferred from users.user_id column type
    RoleID int64   // Inferred from users.role_id column type
}
```

---

## Column Selection

### Selecting Specific Columns

**Query:**
```sql
-- name: GetUserInfo :one
SELECT user_id, username, email
FROM users
WHERE user_id = ?;
```

**Generated:**
```go
type GetUserInfoRow struct {
    UserID   int64
    Username string
    Email    string
}
```

### Selecting All Columns

**Query:**
```sql
-- name: GetUser :one
SELECT * FROM users WHERE user_id = ?;
```

**Generated:**
```go
// Uses User struct defined from schema
func (q *Queries) GetUser(ctx context.Context, userID int64) (User, error)
```

**Note:** Using `SELECT *` generates a struct with all table columns based on schema.

### Column Aliases

**Query:**
```sql
-- name: GetContentWithType :many
SELECT
    cd.content_data_id,
    cd.parent_id,
    dt.label as datatype_label,
    dt.type as datatype_type
FROM content_data cd
JOIN datatypes dt ON cd.datatype_id = dt.datatype_id
WHERE cd.route_id = ?;
```

**Generated:**
```go
type GetContentWithTypeRow struct {
    ContentDataID  int64
    ParentID       sql.NullInt64
    DatatypeLabel  string
    DatatypeType   string
}
```

**Note:** Column aliases are converted to PascalCase in Go struct.

---

## Complex Queries

### Joins

**Query:**
```sql
-- name: GetUsersWithRoles :many
SELECT
    u.user_id,
    u.username,
    u.email,
    r.role_id,
    r.role_name
FROM users u
JOIN roles r ON u.role_id = r.role_id
WHERE r.role_name = ?
ORDER BY u.username;
```

**Generated:**
```go
type GetUsersWithRolesRow struct {
    UserID   int64
    Username string
    Email    string
    RoleID   int64
    RoleName string
}

func (q *Queries) GetUsersWithRoles(ctx context.Context, roleName string) ([]GetUsersWithRolesRow, error)
```

### Subqueries

**Query:**
```sql
-- name: GetUsersInRole :many
SELECT user_id, username, email
FROM users
WHERE role_id IN (
    SELECT role_id
    FROM roles
    WHERE role_name = ?
);
```

**Generated:**
```go
type GetUsersInRoleRow struct {
    UserID   int64
    Username string
    Email    string
}

func (q *Queries) GetUsersInRole(ctx context.Context, roleName string) ([]GetUsersInRoleRow, error)
```

### Aggregates

**Query:**
```sql
-- name: CountUsersByRole :one
SELECT COUNT(*) as user_count
FROM users
WHERE role_id = ?;
```

**Generated:**
```go
func (q *Queries) CountUsersByRole(ctx context.Context, roleID int64) (int64, error)
```

### CTEs (Common Table Expressions)

**PostgreSQL:**
```sql
-- name: GetContentHierarchy :many
WITH RECURSIVE content_tree AS (
    SELECT content_data_id, parent_id, title, 0 as depth
    FROM content_data
    WHERE parent_id IS NULL

    UNION ALL

    SELECT cd.content_data_id, cd.parent_id, cd.title, ct.depth + 1
    FROM content_data cd
    JOIN content_tree ct ON cd.parent_id = ct.content_data_id
)
SELECT * FROM content_tree
ORDER BY depth, content_data_id;
```

**Generated:**
```go
type GetContentHierarchyRow struct {
    ContentDataID int64
    ParentID      sql.NullInt64
    Title         string
    Depth         int32
}

func (q *Queries) GetContentHierarchy(ctx context.Context) ([]GetContentHierarchyRow, error)
```

### Window Functions

**PostgreSQL:**
```sql
-- name: GetUsersWithRank :many
SELECT
    user_id,
    username,
    created_at,
    ROW_NUMBER() OVER (ORDER BY created_at) as user_rank
FROM users
ORDER BY user_rank;
```

**Generated:**
```go
type GetUsersWithRankRow struct {
    UserID    int64
    Username  string
    CreatedAt int64
    UserRank  int64
}
```

---

## Type Mappings

### MySQL to Go Types

| MySQL Type | Go Type |
|------------|---------|
| TINYINT, SMALLINT, MEDIUMINT, INT | int32 |
| BIGINT | int64 |
| FLOAT | float32 |
| DOUBLE | float64 |
| DECIMAL, NUMERIC | string or custom |
| VARCHAR, TEXT | string |
| CHAR | string |
| BLOB, BINARY | []byte |
| DATE, DATETIME, TIMESTAMP | time.Time or int64 |
| BOOLEAN | bool |
| JSON | json.RawMessage |

### PostgreSQL to Go Types

| PostgreSQL Type | Go Type |
|----------------|---------|
| SMALLINT, INT, INTEGER | int32 |
| BIGINT | int64 |
| SERIAL | int32 |
| BIGSERIAL | int64 |
| REAL | float32 |
| DOUBLE PRECISION | float64 |
| NUMERIC, DECIMAL | string or custom |
| VARCHAR, TEXT | string |
| CHAR | string |
| BYTEA | []byte |
| TIMESTAMP, DATE, TIME | time.Time or custom |
| BOOLEAN | bool |
| JSON, JSONB | json.RawMessage |
| UUID | uuid.UUID or string |
| ARRAY | pq.Array or []T |

### SQLite to Go Types

| SQLite Type | Go Type |
|------------|---------|
| INTEGER | int64 |
| REAL | float64 |
| TEXT | string |
| BLOB | []byte |
| NULL | sql.Null* types |

### Nullable Types

**Database NULL values map to:**
- `sql.NullInt64` for nullable integers
- `sql.NullString` for nullable strings
- `sql.NullFloat64` for nullable floats
- `sql.NullBool` for nullable booleans
- `sql.NullTime` for nullable timestamps

**Example:**
```sql
CREATE TABLE content_data (
    content_data_id INTEGER PRIMARY KEY,
    parent_id INTEGER,  -- Nullable
    title TEXT NOT NULL
);
```

**Generated:**
```go
type ContentData struct {
    ContentDataID int64
    ParentID      sql.NullInt64  // Nullable
    Title         string          // Not null
}
```

---

## Generated Code Structure

### Files Generated

sqlc generates these files in the output directory:

**`db.go`**
- DBTX interface
- Queries struct
- New() constructor
- WithTx() method

**`models.go`**
- Struct definitions from schema
- One struct per table

**`querier.go`** (if emit_interface: true)
- Interface with all query methods

**`<queryfile>.sql.go`**
- Generated functions for queries
- One file per .sql query file
- Parameter structs
- Return type structs

### Example Generated Structure

**db.go:**
```go
package dbmysql

import (
    "context"
    "database/sql"
)

type DBTX interface {
    ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
    PrepareContext(context.Context, string) (*sql.Stmt, error)
    QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
    QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

func New(db DBTX) *Queries {
    return &Queries{db: db}
}

type Queries struct {
    db DBTX
}

func (q *Queries) WithTx(tx *sql.Tx) *Queries {
    return &Queries{db: tx}
}
```

**models.go:**
```go
package dbmysql

import "database/sql"

type User struct {
    UserID       int64
    Username     string
    Email        string
    PasswordHash string
    RoleID       int64
    CreatedAt    int64
    UpdatedAt    sql.NullInt64
}

type Role struct {
    RoleID    int64
    RoleName  string
    CreatedAt int64
}
```

**users.sql.go:**
```go
package dbmysql

import "context"

const getUserByID = `-- name: GetUserByID :one
SELECT user_id, username, email, password_hash, role_id, created_at, updated_at
FROM users
WHERE user_id = ?
`

func (q *Queries) GetUserByID(ctx context.Context, userID int64) (User, error) {
    row := q.db.QueryRowContext(ctx, getUserByID, userID)
    var i User
    err := row.Scan(
        &i.UserID,
        &i.Username,
        &i.Email,
        &i.PasswordHash,
        &i.RoleID,
        &i.CreatedAt,
        &i.UpdatedAt,
    )
    return i, err
}
```

---

## Using Generated Code

### Initialize Queries

**With sql.DB:**
```go
import (
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    "github.com/hegner123/modulacms/internal/db-mysql"
)

db, err := sql.Open("mysql", "user:pass@tcp(localhost:3306)/dbname")
if err != nil {
    log.Fatal(err)
}

queries := dbmysql.New(db)
```

**With Transaction:**
```go
tx, err := db.Begin()
if err != nil {
    return err
}
defer tx.Rollback()

queries := dbmysql.New(db).WithTx(tx)

// Use queries...

if err := tx.Commit(); err != nil {
    return err
}
```

### Execute Queries

**Single Row:**
```go
ctx := context.Background()

user, err := queries.GetUserByID(ctx, 123)
if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
        return fmt.Errorf("user not found")
    }
    return err
}

fmt.Printf("User: %s (%s)\n", user.Username, user.Email)
```

**Multiple Rows:**
```go
users, err := queries.ListUsers(ctx)
if err != nil {
    return err
}

for _, user := range users {
    fmt.Printf("%s: %s\n", user.Username, user.Email)
}
```

**Insert with Params:**
```go
result, err := queries.CreateUser(ctx, dbmysql.CreateUserParams{
    Username:     "johndoe",
    Email:        "john@example.com",
    PasswordHash: "hashed_password",
    RoleID:       2,
})
if err != nil {
    return err
}

userID, err := result.LastInsertId()
```

**Update:**
```go
rowsAffected, err := queries.UpdateUserEmail(ctx, dbmysql.UpdateUserEmailParams{
    Email:  "newemail@example.com",
    UserID: 123,
})
if err != nil {
    return err
}

if rowsAffected == 0 {
    return fmt.Errorf("user not found")
}
```

**Delete:**
```go
err := queries.DeleteUser(ctx, 123)
if err != nil {
    return err
}
```

---

## Best Practices

### Query Organization

1. **Group related queries in same file**
   ```
   mysql/users.sql        - User CRUD operations
   mysql/content.sql      - Content operations
   mysql/auth.sql         - Authentication queries
   ```

2. **One query per annotation**
   ```sql
   -- name: GetUser :one
   SELECT * FROM users WHERE user_id = ?;

   -- name: ListUsers :many
   SELECT * FROM users ORDER BY username;
   ```

3. **Use descriptive names**
   - `GetUserByID` not `GetUser`
   - `ListActiveUsers` not `GetUsers`
   - `UpdateUserEmail` not `Update`

### Performance

1. **Select only needed columns**
   ```sql
   -- Good
   SELECT user_id, username, email FROM users;

   -- Avoid
   SELECT * FROM users;
   ```

2. **Use appropriate return types**
   - `:one` for single row (faster than :many)
   - `:exec` when you don't need results
   - `:execrows` only when you need affected count

3. **Index columns in WHERE clauses**
   ```sql
   -- Ensure user_id has an index
   SELECT * FROM users WHERE user_id = ?;
   ```

### Error Handling

1. **Check for sql.ErrNoRows**
   ```go
   user, err := queries.GetUserByID(ctx, id)
   if err != nil {
       if errors.Is(err, sql.ErrNoRows) {
           return nil, ErrUserNotFound
       }
       return nil, err
   }
   ```

2. **Verify affected rows**
   ```go
   rows, err := queries.DeleteUser(ctx, id)
   if err != nil {
       return err
   }
   if rows == 0 {
       return ErrUserNotFound
   }
   ```

### Type Safety

1. **Use parameter structs**
   ```go
   // Generated automatically for queries with multiple params
   params := CreateUserParams{
       Username: "john",
       Email:    "john@example.com",
       Password: "hash",
       RoleID:   2,
   }
   ```

2. **Leverage NULL types**
   ```go
   if user.UpdatedAt.Valid {
       lastUpdate := user.UpdatedAt.Int64
       // Use lastUpdate
   }
   ```

### Documentation

1. **Add comments to complex queries**
   ```sql
   -- name: GetContentHierarchy :many
   -- Returns full content tree with depth calculation
   -- Only includes published content
   SELECT ...
   ```

2. **Document business logic**
   ```go
   // GetActiveUsersByRole returns users with given role
   // who have logged in within the last 30 days
   ```

---

## Troubleshooting

### Common Errors

**"query name must be unique"**
- Two queries have the same name in different files
- Rename one of the queries

**"unable to parse query"**
- SQL syntax error
- Test SQL outside sqlc first
- Check database-specific syntax

**"column doesn't exist"**
- Schema not loaded properly
- Verify schema path in sqlc.yml
- Check column name spelling

**"type mismatch"**
- Generated type doesn't match usage
- Check schema definition
- Verify parameter types

### Debugging

**Validate configuration:**
```bash
cd /Users/home/Documents/Code/Go_dev/modulacms/sql
sqlc validate
```

**Check version:**
```bash
sqlc version
```

**Verbose output:**
```bash
sqlc generate --verbose
```

**Verify schema loading:**
```yaml
# Add multiple schema paths if needed
schema:
  - "./schema"
  - "./schema/1_permissions"
  - "./schema/2_roles"
```

---

## Running sqlc

### Via Makefile

```bash
# From project root
just sqlc
```

### Manually

```bash
# From sql directory
cd /Users/home/Documents/Code/Go_dev/modulacms/sql
sqlc generate

# From project root
cd /Users/home/Documents/Code/Go_dev/modulacms
cd sql && sqlc generate && cd ..
```

### CI/CD Integration

```bash
# Install sqlc
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Generate in CI
cd sql && sqlc generate
```

---

## Advanced Features

### Custom Type Overrides

Override generated types for specific columns:

```yaml
overrides:
  - column: "users.metadata"
    go_type:
      import: "encoding/json"
      type: "RawMessage"

  - column: "users.user_id"
    go_type: "github.com/google/uuid.UUID"
```

### Rename Generated Names

```yaml
rename:
  # Table renames
  user_account: "User"

  # Column renames
  user_id: "ID"
  created_at: "CreatedAt"
  updated_at: "UpdatedAt"
```

### Emit Interface

Generate interface for mocking:

```yaml
emit_interface: true
```

**Generated:**
```go
type Querier interface {
    GetUserByID(ctx context.Context, userID int64) (User, error)
    ListUsers(ctx context.Context) ([]User, error)
    // ... all query methods
}
```

### Prepared Statements

Enable prepared statements:

```yaml
emit_prepared_queries: true
```

Generates code that prepares statements once and reuses them.

---

## Related Documentation

- **SQL_DIRECTORY.md** - Guide for working with sql/ directory
- **DB_PACKAGE.md** - Guide for database package and drivers
- **FILE_TREE.md** - Complete directory structure
- **CLAUDE.md** - Project-wide guidelines

---

## Quick Reference

### sqlc Commands
```bash
sqlc generate      # Generate Go code
sqlc validate      # Validate configuration
sqlc version       # Show version
sqlc help          # Show help
```

### Return Types Summary
- `:many` - Multiple rows → `[]Struct`
- `:one` - Single row → `Struct`
- `:exec` - Execute → `error`
- `:execresult` - Execute → `sql.Result`
- `:execrows` - Execute → `int64`

### Parameter Placeholders
- MySQL: `?`
- PostgreSQL: `$1`, `$2`, `$3`
- SQLite: `?` or `?NNN`

### Key Paths
- Config: `/Users/home/Documents/Code/Go_dev/modulacms/sql/sqlc.yml`
- MySQL queries: `/Users/home/Documents/Code/Go_dev/modulacms/sql/mysql/`
- PostgreSQL queries: `/Users/home/Documents/Code/Go_dev/modulacms/sql/postgres/`
- Schema: `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/`

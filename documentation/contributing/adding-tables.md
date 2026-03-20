# Adding Database Tables

Add a new database table that works across all three backends (SQLite, MySQL, PostgreSQL) using the "comments" table as an example.

## Steps at a Glance

Adding a table involves these steps:

1. Determine the migration number
2. Create schema files for all three databases
3. Write sqlc query annotations for all three databases
4. Update the combined schema files
5. Generate Go code with `just sqlc`
6. Create application-level data structures
7. Add methods to the DbDriver interface
8. Implement methods on all three driver structs
9. Write tests

## Step 1: Find the Next Migration Number

Schema directories in `sql/schema/` are numbered sequentially. List the existing ones to find the next available number:

```bash
ls -1 sql/schema/
```

If the highest is `41_admin_media/`, your new table is `42_comments/`.

```bash
mkdir sql/schema/42_comments
```

## Step 2: Create Schema Files

Each migration directory contains six files -- three schema files and three query files:

```
42_comments/
  schema.sql           # SQLite
  schema_mysql.sql     # MySQL
  schema_psql.sql      # PostgreSQL
  queries.sql          # SQLite queries
  queries_mysql.sql    # MySQL queries
  queries_psql.sql     # PostgreSQL queries
```

### SQLite Schema (schema.sql)

```sql
CREATE TABLE IF NOT EXISTS comments (
    comment_id TEXT NOT NULL
        PRIMARY KEY CHECK (length(comment_id) = 26),
    content_data_id TEXT NOT NULL
        REFERENCES content_data(content_data_id)
            ON DELETE CASCADE,
    author_id TEXT NOT NULL
        REFERENCES users(user_id)
            ON DELETE SET NULL,
    comment_text TEXT NOT NULL,
    status TEXT DEFAULT 'pending',
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);
```

All primary keys use ULID format: 26-character lexicographically sortable unique identifiers stored as TEXT.

### MySQL Schema (schema_mysql.sql)

```sql
CREATE TABLE IF NOT EXISTS comments (
    comment_id VARCHAR(26) NOT NULL
        PRIMARY KEY,
    content_data_id VARCHAR(26) NOT NULL,
    author_id VARCHAR(26) NOT NULL,
    comment_text TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_comments_content_data
        FOREIGN KEY (content_data_id) REFERENCES content_data (content_data_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_comments_author
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE SET NULL
);
```

### PostgreSQL Schema (schema_psql.sql)

```sql
CREATE TABLE IF NOT EXISTS comments (
    comment_id VARCHAR(26) NOT NULL
        PRIMARY KEY,
    content_data_id VARCHAR(26) NOT NULL
        CONSTRAINT fk_comments_content_data
            REFERENCES content_data(content_data_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    author_id VARCHAR(26) NOT NULL
        CONSTRAINT fk_comments_author
            REFERENCES users(user_id)
            ON UPDATE CASCADE ON DELETE SET NULL,
    comment_text TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Key SQL Dialect Differences

| Feature | SQLite | MySQL | PostgreSQL |
|---------|--------|-------|------------|
| Primary key | `TEXT NOT NULL PRIMARY KEY CHECK (length(...) = 26)` | `VARCHAR(26) NOT NULL PRIMARY KEY` | `VARCHAR(26) NOT NULL PRIMARY KEY` |
| Timestamps | `TEXT` | `TIMESTAMP` | `TIMESTAMP` |
| Auto-update timestamp | Application-side | `ON UPDATE CURRENT_TIMESTAMP` | Application-side or trigger |
| Placeholders | `?` | `?` | `$1, $2, $3` |
| RETURNING clause | Supported (3.35+) | Not supported | Supported |

## Step 3: Write sqlc Query Annotations

sqlc annotations define how SQL queries map to Go functions. Write one query file per database.

### SQLite Queries (queries.sql)

```sql
-- name: GetComment :one
SELECT * FROM comments WHERE comment_id = ? LIMIT 1;

-- name: ListComments :many
SELECT * FROM comments ORDER BY date_created DESC;

-- name: CreateComment :one
INSERT INTO comments(content_data_id, author_id, comment_text, status)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: UpdateComment :exec
UPDATE comments SET comment_text=?, status=?, date_modified=CURRENT_TIMESTAMP
WHERE comment_id = ?;

-- name: DeleteComment :exec
DELETE FROM comments WHERE comment_id = ?;
```

### MySQL Queries (queries_mysql.sql)

Same queries, but MySQL does not support `RETURNING`, so `CreateComment` uses `:exec` instead of `:one`:

```sql
-- name: CreateComment :exec
INSERT INTO comments(content_data_id, author_id, comment_text, status)
VALUES (?, ?, ?, ?);
```

### PostgreSQL Queries (queries_psql.sql)

Same queries, but with numbered placeholders:

```sql
-- name: GetComment :one
SELECT * FROM comments WHERE comment_id = $1 LIMIT 1;

-- name: CreateComment :one
INSERT INTO comments(content_data_id, author_id, comment_text, status)
VALUES ($1, $2, $3, $4)
RETURNING *;
```

### sqlc Annotation Reference

| Annotation | Returns |
|-----------|---------|
| `:one` | Single row (struct) |
| `:many` | Slice of rows |
| `:exec` | No return value |
| `:execrows` | Number of affected rows |

## Step 4: Update Combined Schema Files

Fresh installations use combined schema files. Add your table's CREATE statement to the end of each:

- `sql/all_schema.sql` (SQLite)
- `sql/all_schema_mysql.sql` (MySQL)
- `sql/all_schema_psql.sql` (PostgreSQL)

A helper script in `sql/` regenerates all three from the individual files:

```bash
cd sql
./generate_combined.sh
```

## Step 5: Generate Go Code

```bash
just sqlc
```

This produces type-safe Go code in three packages (`internal/db-sqlite/`, `internal/db-mysql/`, `internal/db-psql/`). Each package gets updated struct definitions and query function files.

> **Good to know**: Never edit files in these three directories by hand -- they are overwritten on each generation.

### Code Generation Tools

ModulaCMS has three code generators that eliminate most hand-written boilerplate:

| Tool | Command | Source | Output |
|------|---------|--------|--------|
| **sqlcgen** | `just sqlc-config` | `tools/sqlcgen/definitions.go` | `sql/sqlc.yml` (type overrides for cross-DB mismatches) |
| **dbgen** | `just dbgen` | `tools/dbgen/definitions.go` | `internal/db/{entity}_gen.go` (wrapper structs, mappers, CRUD methods for all 3 drivers) |
| **drivergen** | `just drivergen` | `internal/db/*_custom.go` | MySQL/PostgreSQL method variants from the SQLite (canonical) custom methods |

After generating sqlc code, add an entity definition to `tools/dbgen/definitions.go` and run `just dbgen` to generate the wrapper code. If you add custom methods in `_custom.go` files, only edit the SQLite (`Database` receiver) method, then run `just drivergen` to generate the MySQL and PostgreSQL variants automatically.

## Step 6: Create Application-Level Types

sqlc generates database-specific types with `sql.Null*` fields. Create a new file in `internal/db/` (e.g., `comment.go`) with application-level types:

- **Entity struct** -- Clean types (string, int64) instead of `sql.Null*` types
- **CreateParams and UpdateParams** -- Input structs for create and update operations
- **Mapping functions** -- Convert between sqlc-generated types and your entity struct for each database driver

The mapping functions handle NULL conversions using helpers from `convert.go`. MySQL and PostgreSQL use `int32` where SQLite uses `int64`, so the mappers handle type width conversion as well.

## Step 7: Add Methods to the DbDriver Interface

Add your new query methods to the `DbDriver` interface in `internal/db/db.go`:

```go
type DbDriver interface {
    // ... existing methods ...

    // Comments
    CountComments() (*int64, error)
    CreateCommentTable() error
    CreateComment(context.Context, audited.AuditContext, CreateCommentParams) (*Comments, error)
    DeleteComment(context.Context, audited.AuditContext, types.CommentID) error
    GetComment(types.CommentID) (*Comments, error)
    ListComments() (*[]Comments, error)
    UpdateComment(context.Context, audited.AuditContext, UpdateCommentParams) (*string, error)
}
```

Mutating operations (`Create`, `Update`, `Delete`) take `context.Context` and `audited.AuditContext` parameters for audit trail recording. Read operations (`Get`, `List`, `Count`) do not.

## Step 8: Implement on All Three Drivers

Implement the interface methods on each driver struct (`Database`, `MysqlDatabase`, `PsqlDatabase`). Each implementation:

1. Creates a `Queries` instance from the sqlc-generated package
2. Maps application params to sqlc params
3. Calls the generated query function
4. Maps the sqlc result back to your application type

Also add the `CreateCommentTable` call to each driver's `CreateAllTables()` method in `internal/db/db.go` so the table is created on fresh installations. Add the corresponding `DropCommentTable` call to `DropAllTables()` in `internal/db/wipe.go` in reverse dependency order.

## Step 9: Write Tests

Create `internal/db/comment_test.go` with CRUD tests:

```go
func TestCommentCRUD(t *testing.T) {
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    // Create
    comment := db.CreateComment(CreateCommentParams{
        ContentDataID: 1,
        AuthorID:      1,
        CommentText:   "Test comment",
        Status:        "pending",
    })
    if comment.CommentID == 0 {
        t.Fatal("CreateComment failed: comment_id is 0")
    }

    // Read
    fetched, err := db.GetComment(comment.CommentID)
    if err != nil {
        t.Fatalf("GetComment failed: %v", err)
    }
    if fetched.CommentText != "Test comment" {
        t.Errorf("expected 'Test comment', got '%s'", fetched.CommentText)
    }

    // Update
    err = db.UpdateComment(UpdateCommentParams{
        CommentText: "Updated comment",
        Status:      "approved",
        CommentID:   comment.CommentID,
    })
    if err != nil {
        t.Fatalf("UpdateComment failed: %v", err)
    }

    // Delete
    err = db.DeleteComment(comment.CommentID)
    if err != nil {
        t.Fatalf("DeleteComment failed: %v", err)
    }
}
```

Run the tests:

```bash
just test
```

## Avoid Common Pitfalls

**Forgetting a database backend.** Every schema, query, and driver implementation must exist for all three databases. A table that works in SQLite but is missing from MySQL fails in production.

**Not updating combined schemas.** The combined schema files (`all_schema*.sql`) are used for fresh installations. If your table is only in the migration directory, new installs won't have it.

**SQL dialect differences.** MySQL uses `?` for placeholders and does not support `RETURNING`. PostgreSQL uses `$1, $2, $3` and supports `RETURNING`. Test queries against all three backends.

**Type width mismatches.** SQLite uses `int64` for all integer types. MySQL and PostgreSQL generated code uses `int32` for `INT`/`INTEGER` columns. Your mapping functions must handle the conversion.

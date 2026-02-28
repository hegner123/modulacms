# Adding Database Tables

ModulaCMS supports SQLite, MySQL, and PostgreSQL interchangeably. Every new table must work with all three backends. This guide walks through the full workflow -- from schema design to working Go code -- using a hypothetical "comments" table as the example.

## Overview

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

## Step 1: Determine the Migration Number

Schema directories in `sql/schema/` are numbered sequentially. List the existing ones to find the next available number:

```bash
ls -1 sql/schema/
```

If the highest is `22_joins/`, your new table is `23_comments/`.

```bash
mkdir sql/schema/23_comments
```

## Step 2: Create Schema Files

Each migration directory contains six files -- three schema files and three query files:

```
23_comments/
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
    comment_id INTEGER PRIMARY KEY,
    content_data_id INTEGER NOT NULL
        REFERENCES content_data ON DELETE CASCADE,
    author_id INTEGER NOT NULL
        REFERENCES users ON DELETE SET DEFAULT,
    comment_text TEXT NOT NULL,
    status TEXT DEFAULT 'pending',
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);
```

### MySQL Schema (schema_mysql.sql)

```sql
CREATE TABLE IF NOT EXISTS comments (
    comment_id INT AUTO_INCREMENT PRIMARY KEY,
    content_data_id INT NOT NULL,
    author_id INT DEFAULT 1 NOT NULL,
    comment_text TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT fk_comments_content_data
        FOREIGN KEY (content_data_id) REFERENCES content_data (content_data_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_comments_author
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE ON DELETE SET DEFAULT
);
```

### PostgreSQL Schema (schema_psql.sql)

```sql
CREATE TABLE IF NOT EXISTS comments (
    comment_id SERIAL PRIMARY KEY,
    content_data_id INTEGER NOT NULL
        CONSTRAINT fk_comments_content_data
            REFERENCES content_data
            ON UPDATE CASCADE ON DELETE CASCADE,
    author_id INTEGER NOT NULL
        CONSTRAINT fk_comments_author
            REFERENCES users
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    comment_text TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Key SQL Dialect Differences

| Feature | SQLite | MySQL | PostgreSQL |
|---------|--------|-------|------------|
| Auto-increment | `INTEGER PRIMARY KEY` | `INT AUTO_INCREMENT` | `SERIAL` |
| Timestamps | `TEXT` | `TIMESTAMP` | `TIMESTAMP` |
| Auto-update timestamp | Application-side | `ON UPDATE CURRENT_TIMESTAMP` | Application-side or trigger |
| Placeholders | `?` | `?` | `$1, $2, $3` |
| RETURNING clause | Supported (3.35+) | Not supported | Supported |

## Step 3: Write sqlc Query Annotations

sqlc annotations define how SQL queries become Go functions. Write one query file per database.

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

ModulaCMS uses combined schema files for fresh installations. Add your table's CREATE statement to the end of each:

- `sql/all_schema.sql` (SQLite)
- `sql/all_schema_mysql.sql` (MySQL)
- `sql/all_schema_psql.sql` (PostgreSQL)

Helper scripts in `sql/schema/` can regenerate these from the individual files:

```bash
cd sql/schema
./read_sql.sh > ../all_schema.sql
./read_mysql.sh > ../all_schema_mysql.sql
./read_psql.sh > ../all_schema_psql.sql
```

## Step 5: Generate Go Code

```bash
just sqlc
```

This runs `sqlc generate` and produces type-safe Go code in three packages:

- `internal/db-sqlite/` (package `mdb`)
- `internal/db-mysql/` (package `mdbm`)
- `internal/db-psql/` (package `mdbp`)

Each package gets updated `models.go` (struct definitions) and query function files. Never edit these files by hand -- they are overwritten on each generation.

## Step 6: Create Application-Level Data Structures

sqlc generates database-specific types with `sql.Null*` fields. You need application-level types with clean Go types that your handlers and business logic use.

Create a new file in `internal/db/` (e.g., `comment.go`) containing:

- **Entity struct** -- Clean types (string, int64) instead of sql.Null* types
- **CreateParams and UpdateParams** -- Input structs for create and update operations
- **Mapping functions** -- Convert between sqlc-generated types and your entity struct for each database driver

The mapping functions handle NULL conversions using helpers like `NullStringToString` and `StringToNullString` from the convert utilities. MySQL and PostgreSQL use `int32` where SQLite uses `int64`, so the mappers handle type width conversion as well.

## Step 7: Add to the DbDriver Interface

Add your new query methods to the `DbDriver` interface in `internal/db/db.go`:

```go
type DbDriver interface {
    // ... existing methods ...

    // Comments
    CountComments() (*int64, error)
    CreateComment(s CreateCommentParams) Comment
    DeleteComment(id int64) error
    GetComment(id int64) (*Comment, error)
    ListComments() ([]Comment, error)
    UpdateComment(s UpdateCommentParams) error
}
```

## Step 8: Implement on All Three Drivers

Implement the interface methods on each driver struct: `Database` (SQLite), `MysqlDatabase`, and `PsqlDatabase`. Each implementation:

1. Creates a `Queries` instance from the sqlc-generated package
2. Maps application params to sqlc params
3. Calls the generated query function
4. Maps the sqlc result back to your application type

Also add the `CreateCommentTable` call to each driver's `CreateAllTables()` method so the table is created on fresh installations.

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

## Common Pitfalls

**Forgetting a database backend.** Every schema, query, and driver implementation must exist for all three databases. A table that works in SQLite but is missing from MySQL will fail in production.

**Not updating combined schemas.** The combined schema files (`all_schema*.sql`) are used for fresh installations. If your table is only in the migration directory, new installs will be missing it.

**SQL dialect differences.** MySQL uses `?` for placeholders and does not support `RETURNING`. PostgreSQL uses `$1, $2, $3` and supports `RETURNING`. Test your queries against all three backends.

**Type width mismatches.** SQLite uses `int64` for all integer types. MySQL and PostgreSQL generated code uses `int32` for `INT`/`INTEGER` columns. Your mapping functions must handle the conversion.

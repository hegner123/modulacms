# ADDING_TABLES.md

Complete workflow for adding new database tables to ModulaCMS.

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/ADDING_TABLES.md`
**Purpose:** Step-by-step guide for creating database tables, from schema design to working Go code
**Last Updated:** 2026-01-12

---

## Overview

Adding a new table in ModulaCMS involves:

1. **Determining the migration number** for your new table
2. **Creating schema files** for all three databases (SQLite, MySQL, PostgreSQL)
3. **Writing sqlc query annotations** for CRUD operations
4. **Generating Go code** with `just sqlc`
5. **Adding methods to DbDriver interface** for abstraction
6. **Implementing methods** in all three database drivers
7. **Creating data structures** in internal/db
8. **Writing tests** to verify functionality
9. **Using the new table** in your application code

This document walks through each step with a complete example: adding a **comments** table.

---

## Prerequisites

Before adding a new table, you should understand:

- **Multi-database support**: ModulaCMS supports SQLite, MySQL, and PostgreSQL. Every table must work with all three.
- **sqlc code generation**: SQL queries are compiled to type-safe Go code. See [SQLC.md](../SQLC.md).
- **DbDriver abstraction**: Database operations go through a common interface. See [DB_PACKAGE.md](../DB_PACKAGE.md).
- **Schema numbering**: Migrations are numbered sequentially in the `sql/schema/` directory.

**Related Documentation:**
- [SQL_DIRECTORY.md](../SQL_DIRECTORY.md) - SQL directory organization
- [SQLC.md](../SQLC.md) - sqlc annotations and code generation
- [DB_PACKAGE.md](../DB_PACKAGE.md) - Database abstraction layer
- [ADDING_FEATURES.md](ADDING_FEATURES.md) - Complete feature workflow

---

## Step 1: Determine Migration Number

Schema migrations are numbered sequentially in the `sql/schema/` directory.

### Check Existing Migrations

```bash
ls -1 /Users/home/Documents/Code/Go_dev/modulacms/sql/schema/
```

**Current migrations (as of 2026-01-12):**
```
1_permissions/
2_roles/
3_media_dimension/
4_users/
5_admin_routes/
6_routes/
7_datatypes/
8_fields/
9_admin_datatypes/
10_admin_fields/
11_tokens/
12_user_oauth/
13_tables/
14_media/
15_sessions/
16_content_data/
17_content_fields/
18_admin_content_data/
19_admin_content_fields/
20_datatypes_fields/
21_admin_datatypes_fields/
22_joins/
```

**Next number:** 23

### Choose a Descriptive Name

The directory name should be: `{number}_{table_name_plural}`

**Example:** `23_comments/`

### Create the Directory

```bash
mkdir /Users/home/Documents/Code/Go_dev/modulacms/sql/schema/23_comments
```

---

## Step 2: Create Schema Files

You must create **three schema files** and **three query files** for each database engine.

### Required Files

```
23_comments/
├── schema.sql          # SQLite schema
├── schema_mysql.sql    # MySQL schema
├── schema_psql.sql     # PostgreSQL schema
├── queries.sql         # SQLite queries
├── queries_mysql.sql   # MySQL queries
└── queries_psql.sql    # PostgreSQL queries
```

### 2.1 SQLite Schema (schema.sql)

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/23_comments/schema.sql`

```sql
CREATE TABLE IF NOT EXISTS comments (
    comment_id INTEGER
        PRIMARY KEY,
    content_data_id INTEGER NOT NULL
        REFERENCES content_data
            ON DELETE CASCADE,
    author_id INTEGER NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    comment_text TEXT NOT NULL,
    status TEXT DEFAULT 'pending',
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (content_data_id) REFERENCES content_data(content_data_id) ON DELETE CASCADE,
    FOREIGN KEY (author_id) REFERENCES users(user_id) ON DELETE SET DEFAULT
);
```

**SQLite specifics:**
- Primary key: `INTEGER PRIMARY KEY` (auto-increment)
- Timestamps: `TEXT` with `CURRENT_TIMESTAMP`
- Foreign keys: Defined inline and repeated at bottom (SQLite requirement)

### 2.2 MySQL Schema (schema_mysql.sql)

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/23_comments/schema_mysql.sql`

```sql
CREATE TABLE IF NOT EXISTS comments (
    comment_id INT AUTO_INCREMENT
        PRIMARY KEY,
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

**MySQL specifics:**
- Primary key: `INT AUTO_INCREMENT PRIMARY KEY`
- Timestamps: `TIMESTAMP` with `ON UPDATE CURRENT_TIMESTAMP` for auto-update
- Foreign keys: Named constraints with `CONSTRAINT fk_*`
- Text fields: Consider `VARCHAR` for shorter strings, `TEXT` for long content

### 2.3 PostgreSQL Schema (schema_psql.sql)

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/23_comments/schema_psql.sql`

```sql
CREATE TABLE IF NOT EXISTS comments (
    comment_id SERIAL
        PRIMARY KEY,
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

**PostgreSQL specifics:**
- Primary key: `SERIAL PRIMARY KEY` (auto-increment)
- Timestamps: `TIMESTAMP`
- Foreign keys: Inline `CONSTRAINT` definitions
- No `ON UPDATE CURRENT_TIMESTAMP` (handle in application or use triggers)

---

## Step 3: Write sqlc Queries

sqlc annotations define how SQL queries are converted to Go code. You need queries for each database.

### 3.1 SQLite Queries (queries.sql)

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/23_comments/queries.sql`

```sql
-- name: DropCommentTable :exec
DROP TABLE comments;

-- name: CreateCommentTable :exec
CREATE TABLE IF NOT EXISTS comments (
    comment_id INTEGER
        PRIMARY KEY,
    content_data_id INTEGER NOT NULL
        REFERENCES content_data
            ON DELETE CASCADE,
    author_id INTEGER NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    comment_text TEXT NOT NULL,
    status TEXT DEFAULT 'pending',
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (content_data_id) REFERENCES content_data(content_data_id) ON DELETE CASCADE,
    FOREIGN KEY (author_id) REFERENCES users(user_id) ON DELETE SET DEFAULT
);

-- name: GetComment :one
SELECT * FROM comments
WHERE comment_id = ? LIMIT 1;

-- name: CountComments :one
SELECT COUNT(*)
FROM comments;

-- name: ListComments :many
SELECT * FROM comments
ORDER BY date_created DESC;

-- name: ListCommentsByContent :many
SELECT * FROM comments
WHERE content_data_id = ?
ORDER BY date_created DESC;

-- name: CreateComment :one
INSERT INTO comments(
    content_data_id,
    author_id,
    comment_text,
    status
) VALUES (
    ?,
    ?,
    ?,
    ?
)
RETURNING *;

-- name: UpdateComment :exec
UPDATE comments
SET comment_text=?,
    status=?,
    date_modified=CURRENT_TIMESTAMP
WHERE comment_id = ?;

-- name: DeleteComment :exec
DELETE FROM comments
WHERE comment_id = ?;

-- name: ApproveComment :exec
UPDATE comments
SET status='approved',
    date_modified=CURRENT_TIMESTAMP
WHERE comment_id = ?;

-- name: RejectComment :exec
UPDATE comments
SET status='rejected',
    date_modified=CURRENT_TIMESTAMP
WHERE comment_id = ?;
```

**Query annotations explained:**
- `-- name: {FunctionName} :{returnType}` defines the generated Go function
- Return types:
  - `:exec` - Execute with no return (for INSERT/UPDATE/DELETE without RETURNING)
  - `:one` - Return single row
  - `:many` - Return multiple rows
- `?` placeholders for parameters (SQLite style)
- `RETURNING *` returns the inserted row (SQLite 3.35+)

### 3.2 MySQL Queries (queries_mysql.sql)

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/23_comments/queries_mysql.sql`

```sql
-- name: DropCommentTable :exec
DROP TABLE comments;

-- name: CreateCommentTable :exec
CREATE TABLE IF NOT EXISTS comments (
    comment_id INT AUTO_INCREMENT
        PRIMARY KEY,
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

-- name: GetComment :one
SELECT * FROM comments
WHERE comment_id = ? LIMIT 1;

-- name: CountComments :one
SELECT COUNT(*)
FROM comments;

-- name: ListComments :many
SELECT * FROM comments
ORDER BY date_created DESC;

-- name: ListCommentsByContent :many
SELECT * FROM comments
WHERE content_data_id = ?
ORDER BY date_created DESC;

-- name: CreateComment :exec
INSERT INTO comments(
    content_data_id,
    author_id,
    comment_text,
    status
) VALUES (
    ?,
    ?,
    ?,
    ?
);

-- name: UpdateComment :exec
UPDATE comments
SET comment_text=?,
    status=?
WHERE comment_id = ?;

-- name: DeleteComment :exec
DELETE FROM comments
WHERE comment_id = ?;

-- name: ApproveComment :exec
UPDATE comments
SET status='approved'
WHERE comment_id = ?;

-- name: RejectComment :exec
UPDATE comments
SET status='rejected'
WHERE comment_id = ?;
```

**MySQL differences:**
- No `RETURNING *` clause (MySQL doesn't support it)
- `:exec` return type for INSERT (use `LastInsertId()` to get generated ID)
- `date_modified` updates automatically via `ON UPDATE CURRENT_TIMESTAMP`

### 3.3 PostgreSQL Queries (queries_psql.sql)

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/23_comments/queries_psql.sql`

```sql
-- name: DropCommentTable :exec
DROP TABLE comments;

-- name: CreateCommentTable :exec
CREATE TABLE IF NOT EXISTS comments (
    comment_id SERIAL
        PRIMARY KEY,
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

-- name: GetComment :one
SELECT * FROM comments
WHERE comment_id = $1 LIMIT 1;

-- name: CountComments :one
SELECT COUNT(*)
FROM comments;

-- name: ListComments :many
SELECT * FROM comments
ORDER BY date_created DESC;

-- name: ListCommentsByContent :many
SELECT * FROM comments
WHERE content_data_id = $1
ORDER BY date_created DESC;

-- name: CreateComment :one
INSERT INTO comments(
    content_data_id,
    author_id,
    comment_text,
    status
) VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

-- name: UpdateComment :exec
UPDATE comments
SET comment_text=$1,
    status=$2,
    date_modified=CURRENT_TIMESTAMP
WHERE comment_id = $3;

-- name: DeleteComment :exec
DELETE FROM comments
WHERE comment_id = $1;

-- name: ApproveComment :exec
UPDATE comments
SET status='approved',
    date_modified=CURRENT_TIMESTAMP
WHERE comment_id = $1;

-- name: RejectComment :exec
UPDATE comments
SET status='rejected',
    date_modified=CURRENT_TIMESTAMP
WHERE comment_id = $1;
```

**PostgreSQL differences:**
- `$1, $2, $3` numbered placeholders instead of `?`
- `RETURNING *` supported (returns inserted row)
- Manual `date_modified=CURRENT_TIMESTAMP` in UPDATE queries

---

## Step 4: Update Combined Schema Files

ModulaCMS maintains combined schema files for database initialization. Update these after adding your table.

### 4.1 Update all_schema.sql (SQLite)

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/all_schema.sql`

Add your table's schema to the end:

```sql
-- (existing tables above...)

-- Table: comments
CREATE TABLE IF NOT EXISTS comments (
    comment_id INTEGER
        PRIMARY KEY,
    content_data_id INTEGER NOT NULL
        REFERENCES content_data
            ON DELETE CASCADE,
    author_id INTEGER NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    comment_text TEXT NOT NULL,
    status TEXT DEFAULT 'pending',
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (content_data_id) REFERENCES content_data(content_data_id) ON DELETE CASCADE,
    FOREIGN KEY (author_id) REFERENCES users(user_id) ON DELETE SET DEFAULT
);
```

### 4.2 Update all_schema_mysql.sql

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/all_schema_mysql.sql`

Add MySQL schema to the end.

### 4.3 Update all_schema_psql.sql

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/all_schema_psql.sql`

Add PostgreSQL schema to the end.

**Tip:** Use the helper scripts to regenerate combined schemas:

```bash
cd /Users/home/Documents/Code/Go_dev/modulacms/sql/schema
./read_sql.sh > ../all_schema.sql
./read_mysql.sh > ../all_schema_mysql.sql
./read_psql.sh > ../all_schema_psql.sql
```

---

## Step 5: Generate Go Code with sqlc

Once your schema and queries are written, generate type-safe Go code.

### Run sqlc Code Generation

```bash
cd /Users/home/Documents/Code/Go_dev/modulacms/sql
just sqlc
```

Or directly:

```bash
cd /Users/home/Documents/Code/Go_dev/modulacms/sql
sqlc generate
```

### What Gets Generated

sqlc generates Go code in three packages:

**SQLite:**
- **Package:** `internal/db-sqlite` (package name: `mdb`)
- **Files:**
  - `models.go` - Updated with `Comments` struct
  - `queries.sql.go` - Updated with comment query functions

**MySQL:**
- **Package:** `internal/db-mysql` (package name: `mdbm`)
- **Files:** Same structure

**PostgreSQL:**
- **Package:** `internal/db-psql` (package name: `mdbp`)
- **Files:** Same structure

### Generated Struct Example

After running `just sqlc`, the `Comments` struct will be generated in `internal/db-sqlite/models.go`:

```go
type Comments struct {
    CommentID      int64          `json:"comment_id"`
    ContentDataID  int64          `json:"content_data_id"`
    AuthorID       int64          `json:"author_id"`
    CommentText    string         `json:"comment_text"`
    Status         sql.NullString `json:"status"`
    DateCreated    sql.NullString `json:"date_created"`
    DateModified   sql.NullString `json:"date_modified"`
}
```

### Generated Query Functions Example

In `internal/db-sqlite/queries.sql.go`:

```go
const getComment = `-- name: GetComment :one
SELECT comment_id, content_data_id, author_id, comment_text, status, date_created, date_modified
FROM comments
WHERE comment_id = ? LIMIT 1
`

func (q *Queries) GetComment(ctx context.Context, commentID int64) (Comments, error) {
    row := q.db.QueryRowContext(ctx, getComment, commentID)
    var i Comments
    err := row.Scan(
        &i.CommentID,
        &i.ContentDataID,
        &i.AuthorID,
        &i.CommentText,
        &i.Status,
        &i.DateCreated,
        &i.DateModified,
    )
    return i, err
}
```

**See:** [SQLC.md](../SQLC.md) for detailed sqlc documentation

---

## Step 6: Create Data Structures in internal/db

Create a new file for your table's data structures and mapping functions.

### Create comment.go

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/comment.go`

```go
package db

import (
	"fmt"
	"strconv"

	mdbm "github.com/hegner123/modulacms/internal/db-mysql"
	mdbp "github.com/hegner123/modulacms/internal/db-psql"
	mdb "github.com/hegner123/modulacms/internal/db-sqlite"
)

///////////////////////////////
// STRUCTS
///////////////////////////////

type Comment struct {
	CommentID     int64  `json:"comment_id"`
	ContentDataID int64  `json:"content_data_id"`
	AuthorID      int64  `json:"author_id"`
	CommentText   string `json:"comment_text"`
	Status        string `json:"status"`
	DateCreated   string `json:"date_created"`
	DateModified  string `json:"date_modified"`
}

type CreateCommentParams struct {
	ContentDataID int64  `json:"content_data_id"`
	AuthorID      int64  `json:"author_id"`
	CommentText   string `json:"comment_text"`
	Status        string `json:"status"`
}

type UpdateCommentParams struct {
	CommentText string `json:"comment_text"`
	Status      string `json:"status"`
	CommentID   int64  `json:"comment_id"`
}

type CreateCommentFormParams struct {
	ContentDataID string `json:"content_data_id"`
	AuthorID      string `json:"author_id"`
	CommentText   string `json:"comment_text"`
	Status        string `json:"status"`
}

type UpdateCommentFormParams struct {
	CommentText string `json:"comment_text"`
	Status      string `json:"status"`
	CommentID   string `json:"comment_id"`
}

///////////////////////////////
// GENERIC (Form mapping)
///////////////////////////////

func MapCreateCommentParams(a CreateCommentFormParams) CreateCommentParams {
	return CreateCommentParams{
		ContentDataID: StringToInt64(a.ContentDataID),
		AuthorID:      StringToInt64(a.AuthorID),
		CommentText:   a.CommentText,
		Status:        a.Status,
	}
}

func MapUpdateCommentParams(a UpdateCommentFormParams) UpdateCommentParams {
	return UpdateCommentParams{
		CommentText: a.CommentText,
		Status:      a.Status,
		CommentID:   StringToInt64(a.CommentID),
	}
}

///////////////////////////////
// SQLITE
///////////////////////////////

/// MAPS

func (d Database) MapComment(a mdb.Comments) Comment {
	return Comment{
		CommentID:     a.CommentID,
		ContentDataID: a.ContentDataID,
		AuthorID:      a.AuthorID,
		CommentText:   a.CommentText,
		Status:        NullStringToString(a.Status),
		DateCreated:   NullStringToString(a.DateCreated),
		DateModified:  NullStringToString(a.DateModified),
	}
}

func (d Database) MapCreateCommentParams(a CreateCommentParams) mdb.CreateCommentParams {
	return mdb.CreateCommentParams{
		ContentDataID: a.ContentDataID,
		AuthorID:      a.AuthorID,
		CommentText:   a.CommentText,
		Status:        StringToNullString(a.Status),
	}
}

func (d Database) MapUpdateCommentParams(a UpdateCommentParams) mdb.UpdateCommentParams {
	return mdb.UpdateCommentParams{
		CommentText: a.CommentText,
		Status:      StringToNullString(a.Status),
		CommentID:   a.CommentID,
	}
}

/// QUERIES

func (d Database) CountComments() (*int64, error) {
	queries := mdb.New(d.Connection)
	c, err := queries.CountComments(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d Database) CreateCommentTable() error {
	queries := mdb.New(d.Connection)
	err := queries.CreateCommentTable(d.Context)
	return err
}

func (d Database) CreateComment(s CreateCommentParams) Comment {
	params := d.MapCreateCommentParams(s)
	queries := mdb.New(d.Connection)
	row, err := queries.CreateComment(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateComment: %v\n", err)
	}
	return d.MapComment(row)
}

func (d Database) DeleteComment(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.DeleteComment(d.Context, id)
	if err != nil {
		return fmt.Errorf("Failed to Delete Comment: %v ", id)
	}
	return nil
}

func (d Database) GetComment(id int64) (*Comment, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetComment(d.Context, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get comment: %v", err)
	}
	comment := d.MapComment(row)
	return &comment, nil
}

func (d Database) ListComments() ([]Comment, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListComments(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list comments: %v", err)
	}
	comments := make([]Comment, len(rows))
	for i, row := range rows {
		comments[i] = d.MapComment(row)
	}
	return comments, nil
}

func (d Database) ListCommentsByContent(contentID int64) ([]Comment, error) {
	queries := mdb.New(d.Connection)
	rows, err := queries.ListCommentsByContent(d.Context, contentID)
	if err != nil {
		return nil, fmt.Errorf("failed to list comments by content: %v", err)
	}
	comments := make([]Comment, len(rows))
	for i, row := range rows {
		comments[i] = d.MapComment(row)
	}
	return comments, nil
}

func (d Database) UpdateComment(s UpdateCommentParams) error {
	params := d.MapUpdateCommentParams(s)
	queries := mdb.New(d.Connection)
	err := queries.UpdateComment(d.Context, params)
	if err != nil {
		return fmt.Errorf("failed to update comment: %v", err)
	}
	return nil
}

func (d Database) ApproveComment(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.ApproveComment(d.Context, id)
	if err != nil {
		return fmt.Errorf("failed to approve comment: %v", err)
	}
	return nil
}

func (d Database) RejectComment(id int64) error {
	queries := mdb.New(d.Connection)
	err := queries.RejectComment(d.Context, id)
	if err != nil {
		return fmt.Errorf("failed to reject comment: %v", err)
	}
	return nil
}

///////////////////////////////
// MYSQL
///////////////////////////////

/// MAPS

func (d MysqlDatabase) MapComment(a mdbm.Comments) Comment {
	return Comment{
		CommentID:     int64(a.CommentID),
		ContentDataID: int64(a.ContentDataID),
		AuthorID:      int64(a.AuthorID),
		CommentText:   a.CommentText,
		Status:        NullStringToString(a.Status),
		DateCreated:   NullTimeToString(a.DateCreated),
		DateModified:  NullTimeToString(a.DateModified),
	}
}

func (d MysqlDatabase) MapCreateCommentParams(a CreateCommentParams) mdbm.CreateCommentParams {
	return mdbm.CreateCommentParams{
		ContentDataID: int32(a.ContentDataID),
		AuthorID:      int32(a.AuthorID),
		CommentText:   a.CommentText,
		Status:        StringToNullString(a.Status),
	}
}

func (d MysqlDatabase) MapUpdateCommentParams(a UpdateCommentParams) mdbm.UpdateCommentParams {
	return mdbm.UpdateCommentParams{
		CommentText: a.CommentText,
		Status:      StringToNullString(a.Status),
		CommentID:   int32(a.CommentID),
	}
}

/// QUERIES

func (d MysqlDatabase) CountComments() (*int64, error) {
	queries := mdbm.New(d.Connection)
	c, err := queries.CountComments(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d MysqlDatabase) CreateCommentTable() error {
	queries := mdbm.New(d.Connection)
	err := queries.CreateCommentTable(d.Context)
	return err
}

func (d MysqlDatabase) CreateComment(s CreateCommentParams) Comment {
	params := d.MapCreateCommentParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.CreateComment(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateComment: %v\n", err)
	}
	// MySQL doesn't support RETURNING, so we need to fetch the comment
	// In practice, you'd use LastInsertId() here
	return Comment{}
}

func (d MysqlDatabase) DeleteComment(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.DeleteComment(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Comment: %v ", id)
	}
	return nil
}

func (d MysqlDatabase) GetComment(id int64) (*Comment, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetComment(d.Context, int32(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get comment: %v", err)
	}
	comment := d.MapComment(row)
	return &comment, nil
}

func (d MysqlDatabase) ListComments() ([]Comment, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListComments(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list comments: %v", err)
	}
	comments := make([]Comment, len(rows))
	for i, row := range rows {
		comments[i] = d.MapComment(row)
	}
	return comments, nil
}

func (d MysqlDatabase) ListCommentsByContent(contentID int64) ([]Comment, error) {
	queries := mdbm.New(d.Connection)
	rows, err := queries.ListCommentsByContent(d.Context, int32(contentID))
	if err != nil {
		return nil, fmt.Errorf("failed to list comments by content: %v", err)
	}
	comments := make([]Comment, len(rows))
	for i, row := range rows {
		comments[i] = d.MapComment(row)
	}
	return comments, nil
}

func (d MysqlDatabase) UpdateComment(s UpdateCommentParams) error {
	params := d.MapUpdateCommentParams(s)
	queries := mdbm.New(d.Connection)
	err := queries.UpdateComment(d.Context, params)
	if err != nil {
		return fmt.Errorf("failed to update comment: %v", err)
	}
	return nil
}

func (d MysqlDatabase) ApproveComment(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.ApproveComment(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("failed to approve comment: %v", err)
	}
	return nil
}

func (d MysqlDatabase) RejectComment(id int64) error {
	queries := mdbm.New(d.Connection)
	err := queries.RejectComment(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("failed to reject comment: %v", err)
	}
	return nil
}

///////////////////////////////
// POSTGRESQL
///////////////////////////////

/// MAPS

func (d PsqlDatabase) MapComment(a mdbp.Comments) Comment {
	return Comment{
		CommentID:     int64(a.CommentID),
		ContentDataID: int64(a.ContentDataID),
		AuthorID:      int64(a.AuthorID),
		CommentText:   a.CommentText,
		Status:        NullStringToString(a.Status),
		DateCreated:   NullTimeToString(a.DateCreated),
		DateModified:  NullTimeToString(a.DateModified),
	}
}

func (d PsqlDatabase) MapCreateCommentParams(a CreateCommentParams) mdbp.CreateCommentParams {
	return mdbp.CreateCommentParams{
		ContentDataID: int32(a.ContentDataID),
		AuthorID:      int32(a.AuthorID),
		CommentText:   a.CommentText,
		Status:        StringToNullString(a.Status),
	}
}

func (d PsqlDatabase) MapUpdateCommentParams(a UpdateCommentParams) mdbp.UpdateCommentParams {
	return mdbp.UpdateCommentParams{
		CommentText: a.CommentText,
		Status:      StringToNullString(a.Status),
		CommentID:   int32(a.CommentID),
	}
}

/// QUERIES

func (d PsqlDatabase) CountComments() (*int64, error) {
	queries := mdbp.New(d.Connection)
	c, err := queries.CountComments(d.Context)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return &c, nil
}

func (d PsqlDatabase) CreateCommentTable() error {
	queries := mdbp.New(d.Connection)
	err := queries.CreateCommentTable(d.Context)
	return err
}

func (d PsqlDatabase) CreateComment(s CreateCommentParams) Comment {
	params := d.MapCreateCommentParams(s)
	queries := mdbp.New(d.Connection)
	row, err := queries.CreateComment(d.Context, params)
	if err != nil {
		fmt.Printf("Failed to CreateComment: %v\n", err)
	}
	return d.MapComment(row)
}

func (d PsqlDatabase) DeleteComment(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.DeleteComment(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("Failed to Delete Comment: %v ", id)
	}
	return nil
}

func (d PsqlDatabase) GetComment(id int64) (*Comment, error) {
	queries := mdbp.New(d.Connection)
	row, err := queries.GetComment(d.Context, int32(id))
	if err != nil {
		return nil, fmt.Errorf("failed to get comment: %v", err)
	}
	comment := d.MapComment(row)
	return &comment, nil
}

func (d PsqlDatabase) ListComments() ([]Comment, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListComments(d.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to list comments: %v", err)
	}
	comments := make([]Comment, len(rows))
	for i, row := range rows {
		comments[i] = d.MapComment(row)
	}
	return comments, nil
}

func (d PsqlDatabase) ListCommentsByContent(contentID int64) ([]Comment, error) {
	queries := mdbp.New(d.Connection)
	rows, err := queries.ListCommentsByContent(d.Context, int32(contentID))
	if err != nil {
		return nil, fmt.Errorf("failed to list comments by content: %v", err)
	}
	comments := make([]Comment, len(rows))
	for i, row := range rows {
		comments[i] = d.MapComment(row)
	}
	return comments, nil
}

func (d PsqlDatabase) UpdateComment(s UpdateCommentParams) error {
	params := d.MapUpdateCommentParams(s)
	queries := mdbp.New(d.Connection)
	err := queries.UpdateComment(d.Context, params)
	if err != nil {
		return fmt.Errorf("failed to update comment: %v", err)
	}
	return nil
}

func (d PsqlDatabase) ApproveComment(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.ApproveComment(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("failed to approve comment: %v", err)
	}
	return nil
}

func (d PsqlDatabase) RejectComment(id int64) error {
	queries := mdbp.New(d.Connection)
	err := queries.RejectComment(d.Context, int32(id))
	if err != nil {
		return fmt.Errorf("failed to reject comment: %v", err)
	}
	return nil
}
```

**Key patterns:**
- **Structs:** Define common structs once, map to/from database-specific types
- **Mapping functions:** Convert between sqlc-generated types and your types
- **NULL handling:** Use helper functions like `NullStringToString`, `StringToNullString`
- **Type conversions:** MySQL/PostgreSQL use `int32`, SQLite uses `int64`

---

## Step 7: Add Methods to DbDriver Interface

Update the DbDriver interface to include your new table's methods.

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/db.go`

Add methods to the `DbDriver` interface (around line 69):

```go
type DbDriver interface {
    // Database Connection
    CreateAllTables() error
    InitDB(v *bool) error
    // ... existing methods ...

    // Count operations
    CountComments() (*int64, error)
    // ... existing Count methods ...

    // Create table operations
    CreateCommentTable() error
    // ... existing CreateTable methods ...

    // CRUD operations for comments
    CreateComment(CreateCommentParams) Comment
    DeleteComment(int64) error
    GetComment(int64) (*Comment, error)
    ListComments() ([]Comment, error)
    ListCommentsByContent(int64) ([]Comment, error)
    UpdateComment(UpdateCommentParams) error

    // Comment-specific operations
    ApproveComment(int64) error
    RejectComment(int64) error

    // ... rest of existing methods ...
}
```

**Important:** All three database implementations (Database, MysqlDatabase, PsqlDatabase) must implement these methods, or Go will fail to compile.

---

## Step 8: Add Table Creation to InitDB

Update the `CreateAllTables()` function to include your new table.

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/db.go`

Find the `CreateAllTables()` implementations for each database (around lines 300-600):

```go
// SQLite Database.CreateAllTables()
func (d Database) CreateAllTables() error {
    // ... existing tables ...

    err = d.CreateCommentTable()
    if err != nil {
        return fmt.Errorf("failed to create comment table: %v", err)
    }

    // ... rest of function ...
}

// MysqlDatabase.CreateAllTables()
func (d MysqlDatabase) CreateAllTables() error {
    // ... existing tables ...

    err = d.CreateCommentTable()
    if err != nil {
        return fmt.Errorf("failed to create comment table: %v", err)
    }

    // ... rest of function ...
}

// PsqlDatabase.CreateAllTables()
func (d PsqlDatabase) CreateAllTables() error {
    // ... existing tables ...

    err = d.CreateCommentTable()
    if err != nil {
        return fmt.Errorf("failed to create comment table: %v", err)
    }

    // ... rest of function ...
}
```

---

## Step 9: Write Tests

Create tests to verify your table operations work correctly.

### Create comment_test.go

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/comment_test.go`

```go
package db

import (
	"testing"
)

func TestCommentCRUD(t *testing.T) {
	// Setup test database
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	// Create test user and content data
	userParams := CreateUserParams{
		Username: "testuser",
		Password: "hashedpassword",
	}
	user := db.CreateUser(userParams)

	// Assuming you have content_data setup...
	// contentID := createTestContent(db, user.UserID)
	contentID := int64(1) // Placeholder

	// Test Create
	createParams := CreateCommentParams{
		ContentDataID: contentID,
		AuthorID:      user.UserID,
		CommentText:   "This is a test comment",
		Status:        "pending",
	}
	comment := db.CreateComment(createParams)

	if comment.CommentID == 0 {
		t.Fatal("CreateComment failed: comment_id is 0")
	}
	if comment.CommentText != "This is a test comment" {
		t.Errorf("Expected comment text 'This is a test comment', got '%s'", comment.CommentText)
	}

	// Test Get
	fetchedComment, err := db.GetComment(comment.CommentID)
	if err != nil {
		t.Fatalf("GetComment failed: %v", err)
	}
	if fetchedComment.CommentID != comment.CommentID {
		t.Errorf("Expected comment_id %d, got %d", comment.CommentID, fetchedComment.CommentID)
	}

	// Test Update
	updateParams := UpdateCommentParams{
		CommentText: "Updated comment text",
		Status:      "approved",
		CommentID:   comment.CommentID,
	}
	err = db.UpdateComment(updateParams)
	if err != nil {
		t.Fatalf("UpdateComment failed: %v", err)
	}

	updatedComment, _ := db.GetComment(comment.CommentID)
	if updatedComment.CommentText != "Updated comment text" {
		t.Errorf("Comment text not updated. Expected 'Updated comment text', got '%s'", updatedComment.CommentText)
	}
	if updatedComment.Status != "approved" {
		t.Errorf("Comment status not updated. Expected 'approved', got '%s'", updatedComment.Status)
	}

	// Test Approve
	err = db.ApproveComment(comment.CommentID)
	if err != nil {
		t.Fatalf("ApproveComment failed: %v", err)
	}

	// Test List
	comments, err := db.ListComments()
	if err != nil {
		t.Fatalf("ListComments failed: %v", err)
	}
	if len(comments) == 0 {
		t.Error("ListComments returned empty list")
	}

	// Test ListByContent
	contentComments, err := db.ListCommentsByContent(contentID)
	if err != nil {
		t.Fatalf("ListCommentsByContent failed: %v", err)
	}
	if len(contentComments) == 0 {
		t.Error("ListCommentsByContent returned empty list")
	}

	// Test Delete
	err = db.DeleteComment(comment.CommentID)
	if err != nil {
		t.Fatalf("DeleteComment failed: %v", err)
	}

	// Verify deletion
	_, err = db.GetComment(comment.CommentID)
	if err == nil {
		t.Error("Comment still exists after deletion")
	}

	// Test Count
	count, err := db.CountComments()
	if err != nil {
		t.Fatalf("CountComments failed: %v", err)
	}
	if *count != 0 {
		t.Errorf("Expected 0 comments after deletion, got %d", *count)
	}
}
```

### Run Tests

```bash
cd /Users/home/Documents/Code/Go_dev/modulacms
just test
```

Or test only your new code:

```bash
go test -v ./internal/db -run TestCommentCRUD
```

---

## Step 10: Use the Table in Application Code

Now that your table is fully integrated, use it in your application.

### Example: HTTP Handler

```go
package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/hegner123/modulacms/internal/db"
)

func (s *Server) HandleCreateComment(w http.ResponseWriter, r *http.Request) {
	var params db.CreateCommentFormParams
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	createParams := db.MapCreateCommentParams(params)
	comment := s.DB.CreateComment(createParams)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comment)
}

func (s *Server) HandleGetComment(w http.ResponseWriter, r *http.Request) {
	commentID, err := strconv.ParseInt(r.URL.Query().Get("id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	comment, err := s.DB.GetComment(commentID)
	if err != nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comment)
}

func (s *Server) HandleListCommentsByContent(w http.ResponseWriter, r *http.Request) {
	contentID, err := strconv.ParseInt(r.URL.Query().Get("content_id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid content ID", http.StatusBadRequest)
		return
	}

	comments, err := s.DB.ListCommentsByContent(contentID)
	if err != nil {
		http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(comments)
}
```

---

## Common Pitfalls

### 1. Foreign Key Constraint Violations

**Problem:** Creating a comment fails with "FOREIGN KEY constraint failed"

**Solution:** Ensure referenced records exist:
- `content_data_id` must exist in `content_data` table
- `author_id` must exist in `users` table

### 2. Type Mismatches (int32 vs int64)

**Problem:** Compiler errors about type mismatches between `int32` and `int64`

**Solution:** MySQL and PostgreSQL use `int32` for integer IDs, SQLite uses `int64`. Always cast:
```go
queries.GetComment(d.Context, int32(id)) // MySQL/PostgreSQL
queries.GetComment(d.Context, id)        // SQLite
```

### 3. NULL Handling

**Problem:** Timestamps and optional fields return `sql.Null*` types

**Solution:** Use helper functions:
```go
Status: NullStringToString(a.Status)
DateCreated: NullTimeToString(a.DateCreated)
```

### 4. Missing Table in CreateAllTables

**Problem:** New table doesn't exist after database initialization

**Solution:** Add `CreateCommentTable()` call to all three `CreateAllTables()` implementations

### 5. sqlc Generation Errors

**Problem:** `just sqlc` fails with parsing errors

**Solution:**
- Check SQL syntax in schema files
- Ensure query annotations match format: `-- name: FunctionName :returnType`
- Verify parameter placeholders (`?` for SQLite/MySQL, `$1, $2` for PostgreSQL)

### 6. MySQL INSERT Not Returning Row

**Problem:** `CreateComment` returns empty Comment struct in MySQL

**Solution:** MySQL doesn't support `RETURNING *`. Use `:exec` and fetch the inserted row:
```go
// Option 1: Use LastInsertId()
result, err := queries.CreateComment(d.Context, params)
id, _ := result.LastInsertId()
return db.GetComment(id)

// Option 2: Accept that Create returns incomplete data
// and require a follow-up GetComment() call
```

---

## Database-Specific Considerations

### SQLite

**Strengths:**
- File-based, easy for development
- `RETURNING *` supported (SQLite 3.35+)
- Simpler foreign key syntax

**Limitations:**
- Foreign keys must be enabled: `PRAGMA foreign_keys = ON;`
- No `ON UPDATE CASCADE` (usually not needed)
- Limited concurrent write support

### MySQL

**Strengths:**
- Production-ready, scalable
- `ON UPDATE CURRENT_TIMESTAMP` auto-updates timestamps
- Good concurrent write performance

**Limitations:**
- No `RETURNING *` clause
- Use `INT` (4 bytes) instead of `BIGINT` for most IDs
- Named constraints required for foreign keys

### PostgreSQL

**Strengths:**
- Enterprise features
- `RETURNING *` supported
- `SERIAL` for auto-increment is clear

**Limitations:**
- No auto-update timestamps (use triggers or application logic)
- Numbered parameters: `$1, $2, $3`
- Type strictness can require more casts

---

## Checklist

Use this checklist when adding a new table:

- [ ] Determine next migration number
- [ ] Create migration directory: `sql/schema/{number}_{table}/`
- [ ] Write schema.sql (SQLite)
- [ ] Write schema_mysql.sql (MySQL)
- [ ] Write schema_psql.sql (PostgreSQL)
- [ ] Write queries.sql with sqlc annotations (SQLite)
- [ ] Write queries_mysql.sql (MySQL)
- [ ] Write queries_psql.sql (PostgreSQL)
- [ ] Update all_schema.sql
- [ ] Update all_schema_mysql.sql
- [ ] Update all_schema_psql.sql
- [ ] Run `just sqlc` to generate Go code
- [ ] Create `internal/db/{table}.go` with structs and implementations
- [ ] Add methods to DbDriver interface in `internal/db/db.go`
- [ ] Add table creation to `CreateAllTables()` for all three databases
- [ ] Write tests in `internal/db/{table}_test.go`
- [ ] Run `just test` to verify
- [ ] Use table in application code
- [ ] Test with all three databases
- [ ] Document any special considerations

---

## Related Documentation

**Essential Reading:**
- [SQLC.md](../SQLC.md) - sqlc annotations, configuration, and code generation
- [DB_PACKAGE.md](../DB_PACKAGE.md) - DbDriver interface and database abstraction
- [SQL_DIRECTORY.md](../SQL_DIRECTORY.md) - SQL directory structure and conventions

**Workflow Guides:**
- [ADDING_FEATURES.md](ADDING_FEATURES.md) - Complete feature development workflow
- [CREATING_TUI_SCREENS.md](CREATING_TUI_SCREENS.md) - Adding TUI screens to interact with your table

**Architecture:**
- [CONTENT_MODEL.md](../architecture/CONTENT_MODEL.md) - Domain model relationships
- [DATABASE_LAYER.md](../architecture/DATABASE_LAYER.md) - Database abstraction philosophy

---

## Quick Reference

### File Structure for New Table

```
sql/schema/{number}_{table}/
├── schema.sql          # SQLite CREATE TABLE
├── schema_mysql.sql    # MySQL CREATE TABLE
├── schema_psql.sql     # PostgreSQL CREATE TABLE
├── queries.sql         # SQLite CRUD queries
├── queries_mysql.sql   # MySQL CRUD queries
└── queries_psql.sql    # PostgreSQL CRUD queries

internal/db/
└── {table}.go         # Structs, mappings, implementations

internal/db-sqlite/
├── models.go          # Generated: {Table} struct
└── queries.sql.go     # Generated: query functions

internal/db-mysql/
├── models.go          # Generated
└── queries.sql.go     # Generated

internal/db-psql/
├── models.go          # Generated
└── queries.sql.go     # Generated
```

### Key Commands

```bash
# Generate code from SQL
cd sql && just sqlc

# Run tests
just test

# Test specific package
go test -v ./internal/db -run TestCommentCRUD

# Build and run
just dev
./modulacms-x86 --cli
```

### sqlc Annotations Quick Reference

```sql
-- :exec  - Execute with no return (INSERT/UPDATE/DELETE)
-- :one   - Return single row (SELECT ... LIMIT 1)
-- :many  - Return multiple rows (SELECT without LIMIT)

-- SQLite/MySQL: ? placeholders
-- PostgreSQL: $1, $2, $3 placeholders

-- RETURNING * (SQLite, PostgreSQL only)
INSERT INTO table(...) VALUES (...) RETURNING *;
```

### Type Mapping

| SQL Type | SQLite Go | MySQL Go | PostgreSQL Go |
|----------|-----------|----------|---------------|
| INTEGER/INT | int64 | int32 | int32 |
| TEXT/VARCHAR | string | string | string |
| TIMESTAMP | sql.NullString | sql.NullTime | sql.NullTime |
| NULL columns | sql.Null* | sql.Null* | sql.Null* |

---

**Next Steps:**
- Review [ADDING_FEATURES.md](ADDING_FEATURES.md) for integrating your table into a complete feature
- See [CREATING_TUI_SCREENS.md](CREATING_TUI_SCREENS.md) to add TUI management screens
- Check [TESTING.md](TESTING.md) for comprehensive testing strategies

---

**Last Updated:** 2026-01-12
**Status:** Complete
**Part of:** Phase 2 High Priority Documentation

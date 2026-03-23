# ADDING_TABLES.md

Complete workflow for adding new database tables to ModulaCMS.

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/ADDING_TABLES.md`
**Purpose:** Step-by-step guide for creating database tables, from schema design to working Go code
**Last Updated:** 2026-02-28

---

## Overview

Adding a new table in ModulaCMS involves:

1. **Determining the migration number** for your new table
2. **Creating schema files** for all three databases (SQLite, MySQL, PostgreSQL)
3. **Writing sqlc query annotations** for CRUD operations
4. **Adding sqlc overrides** in `tools/sqlcgen/definitions.go` (if needed for type mismatches)
5. **Generating sqlc Go code** with `just sqlc-config && just sqlc`
6. **Adding a dbgen entity definition** in `tools/dbgen/definitions.go`
7. **Generating wrapper code** with `just dbgen`
8. **Adding methods to DbDriver interface** in `internal/db/db.go`
9. **Adding the table to CreateAllTables** in `internal/db/db.go` **and DropAllTables** in `internal/db/wipe.go`
10. **Optionally creating `{entity}_custom.go`** for extra types or custom queries
11. **Writing tests** to verify functionality

This document walks through each step with a complete example: adding a **comments** table.

### What the code generators handle

Two code generators eliminate most hand-written boilerplate:

| Generator | Source | Output | What it generates |
|-----------|--------|--------|-------------------|
| **sqlcgen** | `tools/sqlcgen/definitions.go` | `sql/sqlc.yml` | sqlc config with type overrides for cross-database type mismatches |
| **dbgen** | `tools/dbgen/definitions.go` | `internal/db/{entity}_gen.go` | Wrapper struct, Map functions, audited commands, CRUD methods for all 3 drivers |

Before these generators existed, each entity required ~400-500 lines of hand-written Go code (3 mappers x 3 drivers, 3 audited commands x 3 drivers, CRUD methods x 3 drivers). Now you define the entity once and generate everything.

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

Find the highest-numbered directory and increment by one. If the highest is `41_admin_media/`, your new table is `42_comments/`.

### Choose a Descriptive Name

The directory name should be: `{number}_{table_name_plural}`

**Example:** `42_comments/`

### Create the Directory

```bash
mkdir /Users/home/Documents/Code/Go_dev/modulacms/sql/schema/42_comments
```

---

## Step 2: Create Schema Files

You must create **three schema files** and **three query files** for each database engine.

### Required Files

```
38_comments/
├── schema.sql          # SQLite schema
├── schema_mysql.sql    # MySQL schema
├── schema_psql.sql     # PostgreSQL schema
├── queries.sql         # SQLite queries
├── queries_mysql.sql   # MySQL queries
└── queries_psql.sql    # PostgreSQL queries
```

### 2.1 SQLite Schema (schema.sql)

**File:** `sql/schema/38_comments/schema.sql`

```sql
CREATE TABLE IF NOT EXISTS comments (
    comment_id TEXT NOT NULL
        PRIMARY KEY,
    content_data_id TEXT NOT NULL
        REFERENCES content_data(content_data_id)
            ON DELETE CASCADE,
    author_id TEXT NOT NULL
        REFERENCES users(user_id)
            ON DELETE SET DEFAULT,
    comment_text TEXT NOT NULL,
    status TEXT DEFAULT 'pending',
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);
```

**SQLite specifics:**
- Primary key: `TEXT NOT NULL PRIMARY KEY` (ULID-based IDs are text)
- Timestamps: `TEXT` with `CURRENT_TIMESTAMP`
- Foreign keys: Inline references

### 2.2 MySQL Schema (schema_mysql.sql)

**File:** `sql/schema/38_comments/schema_mysql.sql`

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
            ON UPDATE CASCADE ON DELETE SET DEFAULT
);
```

**MySQL specifics:**
- Primary key: `VARCHAR(26) NOT NULL PRIMARY KEY` (ULIDs are 26-character strings)
- Timestamps: `TIMESTAMP` with `ON UPDATE CURRENT_TIMESTAMP` for auto-update
- Foreign keys: Named constraints with `CONSTRAINT fk_*`

### 2.3 PostgreSQL Schema (schema_psql.sql)

**File:** `sql/schema/38_comments/schema_psql.sql`

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
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    comment_text TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**PostgreSQL specifics:**
- Primary key: `VARCHAR(26) NOT NULL PRIMARY KEY`
- Timestamps: `TIMESTAMP`
- Foreign keys: Inline `CONSTRAINT` definitions

---

## Step 3: Write sqlc Queries

sqlc annotations define how SQL queries are converted to Go code. You need queries for each database.

### Standard Query Names

dbgen expects specific query names for standard CRUD operations. By default these follow the pattern:

| Operation | Default sqlc name | Override field |
|-----------|-------------------|---------------|
| CreateTable | `Create{Singular}Table` | `SqlcCreateTableName` |
| DropTable | `Drop{Singular}Table` | (none, used in wipe.go) |
| Get | `Get{Singular}` | `SqlcGetName` |
| List | `List{Singular}` | `SqlcListName` |
| ListPaginated | `List{Singular}Paginated` | `SqlcListPaginatedName` |
| Count | `Count{Singular}` | `SqlcCountName` |
| Create | `Create{Singular}` | (none) |
| Update | `Update{Singular}` | (none) |
| Delete | `Delete{Singular}` | (none) |

For our `comments` example with `Singular: "Comment"`, the query names would be `CreateCommentTable`, `GetComment`, `ListComment`, `CountComment`, `CreateComment`, `UpdateComment`, `DeleteComment`.

### 3.1 SQLite Queries (queries.sql)

**File:** `sql/schema/38_comments/queries.sql`

```sql
-- name: DropCommentTable :exec
DROP TABLE comments;

-- name: CreateCommentTable :exec
CREATE TABLE IF NOT EXISTS comments (
    comment_id TEXT NOT NULL
        PRIMARY KEY,
    content_data_id TEXT NOT NULL
        REFERENCES content_data(content_data_id)
            ON DELETE CASCADE,
    author_id TEXT NOT NULL
        REFERENCES users(user_id)
            ON DELETE SET DEFAULT,
    comment_text TEXT NOT NULL,
    status TEXT DEFAULT 'pending',
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

-- name: GetComment :one
SELECT * FROM comments
WHERE comment_id = ? LIMIT 1;

-- name: CountComment :one
SELECT COUNT(*)
FROM comments;

-- name: ListComment :many
SELECT * FROM comments
ORDER BY date_created DESC;

-- name: CreateComment :one
INSERT INTO comments(
    comment_id,
    content_data_id,
    author_id,
    comment_text,
    status,
    date_created,
    date_modified
) VALUES (
    ?, ?, ?, ?, ?, ?, ?
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
```

**Query annotations explained:**
- `-- name: {FunctionName} :{returnType}` defines the generated Go function
- Return types: `:exec` (no return), `:one` (single row), `:many` (multiple rows)
- `?` placeholders for parameters (SQLite style)
- `RETURNING *` returns the inserted row (SQLite 3.35+)
- The `comment_id` is now a parameter (ULID generated in Go, not auto-increment)

### 3.2 MySQL Queries (queries_mysql.sql)

**File:** `sql/schema/38_comments/queries_mysql.sql`

```sql
-- name: DropCommentTable :exec
DROP TABLE comments;

-- name: CreateCommentTable :exec
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
            ON UPDATE CASCADE ON DELETE SET DEFAULT
);

-- name: GetComment :one
SELECT * FROM comments
WHERE comment_id = ? LIMIT 1;

-- name: CountComment :one
SELECT COUNT(*)
FROM comments;

-- name: ListComment :many
SELECT * FROM comments
ORDER BY date_created DESC;

-- name: CreateComment :exec
INSERT INTO comments(
    comment_id,
    content_data_id,
    author_id,
    comment_text,
    status,
    date_created,
    date_modified
) VALUES (
    ?, ?, ?, ?, ?, ?, ?
);

-- name: UpdateComment :exec
UPDATE comments
SET comment_text=?,
    status=?
WHERE comment_id = ?;

-- name: DeleteComment :exec
DELETE FROM comments
WHERE comment_id = ?;
```

**MySQL differences:**
- No `RETURNING *` clause (MySQL doesn't support it). dbgen handles the MySQL "exec then get" pattern automatically.
- `date_modified` updates automatically via `ON UPDATE CURRENT_TIMESTAMP`

### 3.3 PostgreSQL Queries (queries_psql.sql)

**File:** `sql/schema/38_comments/queries_psql.sql`

```sql
-- name: DropCommentTable :exec
DROP TABLE comments;

-- name: CreateCommentTable :exec
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
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    comment_text TEXT NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- name: GetComment :one
SELECT * FROM comments
WHERE comment_id = $1 LIMIT 1;

-- name: CountComment :one
SELECT COUNT(*)
FROM comments;

-- name: ListComment :many
SELECT * FROM comments
ORDER BY date_created DESC;

-- name: CreateComment :one
INSERT INTO comments(
    comment_id,
    content_data_id,
    author_id,
    comment_text,
    status,
    date_created,
    date_modified
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
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
```

**PostgreSQL differences:**
- `$1, $2, $3` numbered placeholders instead of `?`
- `RETURNING *` supported (returns inserted row)
- Manual `date_modified=CURRENT_TIMESTAMP` in UPDATE queries

---

## Step 4: Update Combined Schema Files

ModulaCMS maintains combined schema files for database initialization. Use the helper scripts to regenerate them:

```bash
cd /Users/home/Documents/Code/Go_dev/modulacms/sql
./generate_combined.sh
```

This single script regenerates all three combined schema files (`all_schema.sql`, `all_schema_mysql.sql`, `all_schema_psql.sql`) from the individual schema directories.

---

## Step 5: Add sqlc Overrides (if needed)

Before running sqlc, check whether your table has columns that need type overrides. The overrides live in `tools/sqlcgen/definitions.go` and handle cross-database type mismatches.

### When You Need Overrides

Overrides are needed when:
- A column uses a ULID-based typed ID (maps `TEXT`/`VARCHAR(26)` to `types.CommentID`)
- A column uses a foreign key typed ID (maps to `types.ContentID`, `types.UserID`, etc.)
- Timestamps differ across drivers (SQLite `TEXT` vs MySQL/PG `TIMESTAMP` -- mapped to `types.Timestamp`)
- Booleans differ across drivers (SQLite `int64` vs MySQL/PG `bool` -- mapped to `types.SafeBool`)
- Nullable integers differ (SQLite `int64` vs MySQL/PG `int32` -- mapped to `types.NullableInt64`)

### Adding Overrides

Edit `tools/sqlcgen/definitions.go` and add entries to the `Overrides` slice.

**For the comments example, add:**

```go
// COMMENTS
{Comment: "COMMENTS", Column: "comments.comment_id", Import: typesImport, Type: "CommentID"},
{Column: "comments.content_data_id", Import: typesImport, Type: "ContentID"},
{Column: "comments.author_id", Import: typesImport, Type: "UserID"},
```

**Note:** `date_created` and `date_modified` are already covered by the wildcard overrides `*.date_created` and `*.date_modified` which map to `types.Timestamp`. Similarly, `*.author_id` (non-nullable) already maps to `types.UserID`. Only add table-specific overrides when the wildcard doesn't cover your case or when you need to override a wildcard.

### Adding Renames (if needed)

If sqlc would singularize your table name incorrectly (e.g., `media` becomes `medium`), add a rename:

```go
{From: "comment", To: "Comments"},
```

Also add any ID column renames to ensure uppercase `ID` suffix:

```go
{From: "comment_id", To: "CommentID", Quoted: true},
```

### Available Override Types

| Type | Purpose | When to use |
|------|---------|-------------|
| `types.SafeBool` | Unifies `bool` (MySQL/PG) vs `int64` (SQLite) | Boolean columns |
| `types.NullableInt64` | Unifies nullable `int32` (MySQL/PG) vs nullable `int64` (SQLite) | Nullable integer columns |
| `types.Timestamp` | Unifies `time.Time` (MySQL/PG) vs `string` (SQLite) | Temporal columns not covered by `*.date_created`/`*.date_modified` wildcards |
| `types.{TypeName}ID` | ULID-based typed ID | Primary key and foreign key columns |
| `types.Nullable{TypeName}ID` | Nullable ULID-based typed ID | Nullable foreign key columns |

### Generating sqlc Config and Code

After adding overrides, regenerate the sqlc config and then the Go code:

```bash
just sqlc-config   # Regenerates sql/sqlc.yml from definitions
just sqlc          # Runs sqlc-config then sqlc generate (shortcut)
```

Or use the combined command:

```bash
just sqlc   # Does both steps
```

### What Gets Generated

sqlc generates Go code in three packages:

| Package | Location | Package name |
|---------|----------|-------------|
| SQLite | `internal/db-sqlite/` | `mdb` |
| MySQL | `internal/db-mysql/` | `mdbm` |
| PostgreSQL | `internal/db-psql/` | `mdbp` |

Each gets updated `models.go` (with the `Comments` struct) and `queries*.sql.go` (with query functions).

**Do not edit files in these directories by hand** -- they are overwritten by sqlc.

---

## Step 6: Add Typed ID (if needed)

If your entity uses a new ULID-based typed ID, add it to `internal/db/types/types_ids.go`. Existing entities use types like `ContentID`, `UserID`, `MediaID`, etc.

For our comments example, if `CommentID` doesn't already exist:

```go
type CommentID string

func NewCommentID() CommentID { return CommentID(newULID()) }
```

The typed ID types implement `driver.Valuer`, `sql.Scanner`, and `json.Marshaler` via shared patterns already in `types_ids.go`. Follow the existing pattern for the new type.

---

## Step 7: Add dbgen Entity Definition

This is the key step that replaces hundreds of lines of hand-written boilerplate. Add an entity definition to `tools/dbgen/definitions.go`.

### Entity Definition for Comments

Add to the `Entities` slice:

```go
// Comments
{
    Name:               "Comments",
    Singular:           "Comment",
    Plural:             "Comments",
    SqlcTypeName:       "Comments",
    TableName:          "comments",
    IDType:             "types.CommentID",
    IDField:            "CommentID",
    NewIDFunc:          "types.NewCommentID()",
    UpdateSuccessField: "s.CommentText",
    StringTypeName:     "StringComments",
    Fields: []Field{
        {AppName: "CommentID", Type: "types.CommentID", JSONTag: "comment_id", IsPrimaryID: true, InCreate: false, InUpdate: true, StringConvert: "toString"},
        {AppName: "ContentDataID", Type: "types.ContentID", JSONTag: "content_data_id", InCreate: true, InUpdate: false, StringConvert: "toString"},
        {AppName: "AuthorID", Type: "types.UserID", JSONTag: "author_id", InCreate: true, InUpdate: false, StringConvert: "toString"},
        {AppName: "CommentText", Type: "string", JSONTag: "comment_text", InCreate: true, InUpdate: true, StringConvert: "string"},
        {AppName: "Status", Type: "string", JSONTag: "status", InCreate: true, InUpdate: true, StringConvert: "string"},
        {AppName: "DateCreated", Type: "types.Timestamp", JSONTag: "date_created", InCreate: true, InUpdate: true, StringConvert: "toString"},
        {AppName: "DateModified", Type: "types.Timestamp", JSONTag: "date_modified", InCreate: true, InUpdate: true, StringConvert: "toString"},
    },
    OutputFile: "comment_gen.go",
},
```

### Entity Definition Fields Explained

**Top-level fields:**

| Field | Purpose | Example |
|-------|---------|---------|
| `Name` | Struct name in generated code | `"Comments"` |
| `Singular` | Used in method names: `Create{Singular}`, `Get{Singular}` | `"Comment"` |
| `Plural` | Used in list method names: `List{Plural}` | `"Comments"` |
| `SqlcTypeName` | sqlc struct name (usually same as `Name`) | `"Comments"` |
| `TableName` | SQL table name | `"comments"` |
| `IDType` | Go type for the primary key | `"types.CommentID"` |
| `IDField` | Field name for the primary key | `"CommentID"` |
| `NewIDFunc` | Expression to generate a new ID | `"types.NewCommentID()"` |
| `HasPaginated` | Generate `ListPaginated` method | `true` (default `false`) |
| `CallerSuppliedID` | ID in CreateParams, generate-if-empty pattern | `true` (default `false`) |
| `UpdateSuccessField` | Field shown in update success message | `"s.CommentText"` |
| `StringTypeName` | String struct name for TUI display (empty = skip) | `"StringComments"` |
| `OutputFile` | Generated file name | `"comment_gen.go"` |

**Per-field flags:**

| Flag | Purpose | Example |
|------|---------|---------|
| `AppName` | Field name in wrapper struct | `"CommentText"` |
| `SqlcName` | sqlc field name if different from AppName | `"Roles"` (for a field named `Role`) |
| `Type` | Go type | `"types.CommentID"`, `"string"`, `"bool"` |
| `JSONTag` | JSON tag value | `"comment_text"` |
| `IsPrimaryID` | This is the entity's primary key | `true` |
| `InCreate` | Include in `CreateParams` struct | `true` |
| `InUpdate` | Include in `UpdateParams` struct | `true` |
| `NarrowInt` | Generate `int32`/`int64` casts for MySQL/PG | `true` |
| `SafeBool` | Generate `.Bool()` / `{Val: x}` conversions | `true` |
| `StringConvert` | Conversion for `MapString` function | See table below |

**StringConvert values:**

| Value | Generated expression | Use case |
|-------|---------------------|----------|
| `"toString"` | `a.Field.String()` | Typed IDs, timestamps, enums |
| `"string"` | `a.Field` (identity) | Plain strings |
| `"sprintf"` | `fmt.Sprintf("%d", a.Field)` | Integers |
| `"sprintfBool"` | `fmt.Sprintf("%t", a.Field)` | Booleans |
| `"sprintfFloat64"` | `fmt.Sprintf("%v", a.Field.Float64)` | Nullable floats |
| `"nullToString"` | `utility.NullToString(a.Field)` | sql.NullString |
| `"wrapperNullToString"` | `utility.NullToString(a.Field)` | Wrapper NullString |
| `""` (empty) | Skip field | Not shown in TUI |

### Skip Flags

Use sparingly, only for genuinely custom behavior:

| Flag | What it skips | When to use |
|------|--------------|-------------|
| `SkipMappers` | `MapEntity`, `MapCreate`, `MapUpdate` functions | Per-driver conversions too complex for generation |
| `SkipAuditedCommands` | Audited command structs and factory methods | Entity doesn't use audited mutations |
| `SkipGet` | `Get` CRUD method | No matching sqlc Get query, or ID field name differs |

When you skip a portion, you must hand-write the skipped code in `{entity}_custom.go`.

### Query Name Overrides

If your sqlc query names don't follow the default `{Operation}{Singular}` pattern, use these overrides:

```go
SqlcCreateTableName:   "CreateDatatypesFieldsTable",  // when default would be wrong
SqlcCountName:         "CountAdminRoute",              // when sqlc lowercases differently
SqlcGetName:           "GetDatatypeField",             // override Get query name
SqlcListName:          "ListDatatypeField",            // override List query name
SqlcListPaginatedName: "ListContentFieldsPaginated",   // override ListPaginated query name
```

### Extra Queries

For queries beyond standard CRUD, use `ExtraQueries`:

```go
ExtraQueries: []ExtraQuery{
    {
        MethodName:  "ListCommentsByContent",
        SqlcName:    "ListCommentsByContent",
        ReturnsList: true,
        Params: []ExtraQueryParam{
            {ParamName: "contentID", ParamType: "types.ContentID", SqlcField: "ContentDataID"},
        },
    },
},
```

For paginated extra queries, use `PaginatedExtraQueries`. For additional param structs, use `ExtraParamStructs`.

### Generate Wrapper Code

```bash
just dbgen                    # Generate all entities
just dbgen-entity Comments    # Generate just the Comments entity
```

### What dbgen Generates

The generated `internal/db/comment_gen.go` file contains:

1. **Wrapper struct** (`Comments`) with app-level types
2. **CreateParams / UpdateParams** structs
3. **MapString function** (for TUI display, if `StringTypeName` is set)
4. **For each of the 3 drivers (SQLite, MySQL, PostgreSQL):**
   - `MapComment` -- converts sqlc struct to wrapper struct
   - `MapCreateCommentParams` -- converts wrapper params to sqlc params
   - `MapUpdateCommentParams` -- converts wrapper params to sqlc params
   - Audited command structs (`NewCommentCmd`, `UpdateCommentCmd`, `DeleteCommentCmd`)
   - `CountComments()` -- count records
   - `CreateCommentTable()` -- create the table
   - `CreateComment(ctx, ac, params)` -- audited create
   - `UpdateComment(ctx, ac, params)` -- audited update
   - `DeleteComment(ctx, ac, id)` -- audited delete
   - `GetComment(id)` -- get by ID
   - `ListComments()` -- list all

**Never edit `_gen.go` files by hand** -- they are overwritten by `just dbgen`.

---

## Step 8: Add Methods to DbDriver Interface

Update the `DbDriver` interface in `internal/db/db.go` to include your new entity's methods. The method signatures must match what dbgen generated.

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

    // ... rest of existing methods ...
}
```

**Important:** All three database implementations (`Database`, `MysqlDatabase`, `PsqlDatabase`) must implement these methods, or Go will fail to compile. Since dbgen generated methods for all three, the interface is already satisfied.

### Method Signature Patterns

dbgen generates methods with these signatures:

| Method | Signature |
|--------|-----------|
| Count | `Count{Plural}() (*int64, error)` |
| CreateTable | `Create{Singular}Table() error` |
| Create | `Create{Singular}(context.Context, audited.AuditContext, Create{Singular}Params) (*{Name}, error)` |
| Update | `Update{Singular}(context.Context, audited.AuditContext, Update{Singular}Params) (*string, error)` |
| Delete | `Delete{Singular}(context.Context, audited.AuditContext, {IDType}) error` |
| Get | `Get{Singular}({IDType}) (*{Name}, error)` |
| List | `List{Plural}() (*[]{Name}, error)` |
| ListPaginated | `List{Plural}Paginated(PaginationParams) (*[]{Name}, error)` (if `HasPaginated: true`) |

---

## Step 9: Add Table to CreateAllTables / DropAllTables

### CreateAllTables

Update the `CreateAllTables()` function for all three driver structs in `internal/db/db.go`. Add your table in the correct dependency order:

```go
func (d Database) CreateAllTables() error {
    // ... existing tables ...

    err = d.CreateCommentTable()
    if err != nil {
        return fmt.Errorf("failed to create comment table: %v", err)
    }

    // ... rest of function ...
}
```

Repeat for `MysqlDatabase` and `PsqlDatabase`.

### DropAllTables

Update `DropAllTables()` in `internal/db/wipe.go`. Add the drop in **reverse dependency order** (tables that depend on others must be dropped first):

```go
func (d Database) DropAllTables() error {
    queries := mdb.New(d.Connection)

    ops := []dropOp{
        // ... higher-tier drops ...

        // Comments depend on content_data and users, so drop before those
        {"comments", func() error { return queries.DropCommentTable(d.Context) }},

        // ... rest of drops ...
    }

    return runDropOps(ops)
}
```

The `runDropOps` helper executes each drop sequentially, logging warnings for failures and continuing to the next table. Tables that don't exist are skipped. A combined error is returned if any drops failed.

Repeat for `MysqlDatabase` and `PsqlDatabase` (all three use the same `dropOp`/`runDropOps` pattern).

---

## Step 10: Create Custom File (optional)

For types and queries not covered by dbgen, create `internal/db/comment_custom.go`:

```go
package db

// Custom types, form params, or additional queries that aren't generated.
// dbgen handles standard CRUD. Put everything else here.
```

Common things that go in `_custom.go`:

- **Form params** (`CreateCommentFormParams`, `UpdateCommentFormParams`) for TUI/web forms that use string types
- **Paginated queries** with custom filter fields (if not using `PaginatedExtraQueries` in dbgen)
- **Complex queries** that join multiple tables or have non-standard logic
- **Custom map functions** (when using `SkipMappers: true`)

**Important:** After editing `_custom.go` files, run `just drivergen` to regenerate the MySQL and PostgreSQL method variants from the SQLite (canonical) version. Only edit the `Database` (SQLite) receiver methods -- the `MysqlDatabase` and `PsqlDatabase` variants are generated automatically. See the [Custom Wrapper Replication](#custom-wrapper-replication) section in CLAUDE.md for details on how drivergen works.

---

## Step 11: Write Tests

Create tests to verify your table operations work correctly.

### Create comment_test.go

**File:** `internal/db/comment_test.go`

```go
package db

import (
    "testing"

    "github.com/hegner123/modulacms/internal/db/audited"
    "github.com/hegner123/modulacms/internal/db/types"
)

func TestCommentCRUD(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    ctx := context.Background()
    ac := audited.Ctx(types.NewNodeID(), types.NewUserID(), "test", "127.0.0.1")

    // Test Create
    createParams := CreateCommentParams{
        ContentDataID: someContentID, // from test fixtures
        AuthorID:      someUserID,    // from test fixtures
        CommentText:   "This is a test comment",
        Status:        "pending",
        DateCreated:   types.TimestampNow(),
        DateModified:  types.TimestampNow(),
    }
    comment, err := db.CreateComment(ctx, ac, createParams)
    if err != nil {
        t.Fatalf("CreateComment failed: %v", err)
    }
    if comment.CommentID.IsZero() {
        t.Fatal("CreateComment returned zero ID")
    }
    if comment.CommentText != "This is a test comment" {
        t.Errorf("Expected comment text 'This is a test comment', got '%s'", comment.CommentText)
    }

    // Test Get
    fetched, err := db.GetComment(comment.CommentID)
    if err != nil {
        t.Fatalf("GetComment failed: %v", err)
    }
    if fetched.CommentID != comment.CommentID {
        t.Errorf("Expected comment_id %s, got %s", comment.CommentID, fetched.CommentID)
    }

    // Test Update
    updateParams := UpdateCommentParams{
        CommentText:  "Updated comment text",
        Status:       "approved",
        DateCreated:  comment.DateCreated,
        DateModified: types.TimestampNow(),
        CommentID:    comment.CommentID,
    }
    _, err = db.UpdateComment(ctx, ac, updateParams)
    if err != nil {
        t.Fatalf("UpdateComment failed: %v", err)
    }

    updated, err := db.GetComment(comment.CommentID)
    if err != nil {
        t.Fatalf("GetComment after update failed: %v", err)
    }
    if updated.CommentText != "Updated comment text" {
        t.Errorf("Comment text not updated. Expected 'Updated comment text', got '%s'", updated.CommentText)
    }

    // Test List
    comments, err := db.ListComments()
    if err != nil {
        t.Fatalf("ListComments failed: %v", err)
    }
    if len(*comments) == 0 {
        t.Error("ListComments returned empty list")
    }

    // Test Count
    count, err := db.CountComments()
    if err != nil {
        t.Fatalf("CountComments failed: %v", err)
    }
    if *count != 1 {
        t.Errorf("Expected 1 comment, got %d", *count)
    }

    // Test Delete
    err = db.DeleteComment(ctx, ac, comment.CommentID)
    if err != nil {
        t.Fatalf("DeleteComment failed: %v", err)
    }

    _, err = db.GetComment(comment.CommentID)
    if err == nil {
        t.Error("Comment still exists after deletion")
    }
}
```

### Run Tests

```bash
just test                                           # Run all tests
go test -v ./internal/db -run TestCommentCRUD       # Run only your new test
```

---

## Complete Command Sequence

For reference, here is the full sequence of commands from start to finish:

```bash
# 1. Create schema directory
mkdir sql/schema/42_comments

# 2. Create 6 SQL files (schema + queries for each DB)
# (create files as shown in Steps 2-3)

# 3. Regenerate combined schemas
cd sql && ./generate_combined.sh

# 4. Add overrides to tools/sqlcgen/definitions.go (if needed)
# 5. Add typed ID to internal/db/types/types_ids.go (if needed)

# 6. Generate sqlc config and Go code
just sqlc

# 7. Add entity definition to tools/dbgen/definitions.go
# 8. Generate wrapper code
just dbgen

# 9. Add methods to DbDriver interface in internal/db/db.go
# 10. Add table to CreateAllTables (db.go) and DropAllTables (wipe.go)

# 11. If you created _custom.go files, regenerate driver variants
just drivergen

# 12. Write tests and run them
go test -v ./internal/db -run TestCommentCRUD

# 13. Verify everything compiles
just check
```

---

## Common Pitfalls

### 1. sqlc Query Names Don't Match dbgen Expectations

**Problem:** dbgen generates code that calls `queries.GetComment(...)` but sqlc generated `queries.GetComments(...)`.

**Solution:** Either rename the sqlc query to match the expected pattern (`Get{Singular}`), or use the `SqlcGetName` override in the entity definition:

```go
SqlcGetName: "GetComments",
```

### 2. Type Mismatches (int32 vs int64)

**Problem:** MySQL/PostgreSQL sqlc types use `int32` for integer columns while SQLite uses `int64`.

**Solution:** Use `NarrowInt: true` on the field in the dbgen definition. The generator will insert the appropriate casts automatically. For columns that are already handled by sqlc overrides (like typed IDs or `types.Timestamp`), no `NarrowInt` flag is needed.

### 3. SafeBool Conversions

**Problem:** SQLite uses `int64` for boolean columns while MySQL/PostgreSQL use `bool`. sqlc generates different types.

**Solution:** Two steps:
1. Add a `types.SafeBool` override in `tools/sqlcgen/definitions.go`
2. Set `SafeBool: true` on the field in the dbgen entity definition

### 4. Missing Table in CreateAllTables

**Problem:** New table doesn't exist after database initialization.

**Solution:** Add `Create{Singular}Table()` call to all three `CreateAllTables()` implementations in `internal/db/db.go`.

### 5. sqlc Generation Errors

**Problem:** `just sqlc` fails with parsing errors.

**Solution:**
- Check SQL syntax in schema files
- Ensure query annotations match format: `-- name: FunctionName :returnType`
- Verify parameter placeholders (`?` for SQLite/MySQL, `$1, $2` for PostgreSQL)
- Run `just sqlc-config` first to regenerate `sql/sqlc.yml`, then `just sqlc`

### 6. MySQL INSERT Without RETURNING

**Problem:** MySQL doesn't support `RETURNING *`.

**Solution:** dbgen handles this automatically. When `MysqlReturningGap` is `true` in the driver config, the generated `Create` method uses `:exec` and then calls `Get` to fetch the inserted row. You just need to ensure the MySQL query uses `:exec` (not `:one`) for `CreateComment`.

### 7. Editing Generated Files

**Problem:** Changes to `_gen.go` files are lost on next `just dbgen`.

**Solution:** Put custom code in `{entity}_custom.go`. The `_gen.go` files have a `// Code generated by tools/dbgen; DO NOT EDIT.` header.

---

## Database-Specific Considerations

### SQLite

- File-based, easy for development
- `RETURNING *` supported (SQLite 3.35+)
- Primary keys: `TEXT NOT NULL PRIMARY KEY` for ULID columns
- Timestamps stored as `TEXT`, mapped to `types.Timestamp` via sqlc override
- Booleans stored as `INTEGER`, mapped to `types.SafeBool` via sqlc override

### MySQL

- No `RETURNING *` clause -- dbgen generates exec-then-get pattern automatically
- `ON UPDATE CURRENT_TIMESTAMP` auto-updates `date_modified`
- Primary keys: `VARCHAR(26) NOT NULL PRIMARY KEY` for ULID columns
- Named constraints required for foreign keys

### PostgreSQL

- `RETURNING *` supported
- No auto-update timestamps (handle in UPDATE queries or application logic)
- Numbered parameters: `$1, $2, $3`
- Primary keys: `VARCHAR(26) NOT NULL PRIMARY KEY` for ULID columns

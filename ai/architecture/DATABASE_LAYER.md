# DATABASE_LAYER.md

Comprehensive guide to ModulaCMS's database abstraction layer and the DbDriver interface.

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/DATABASE_LAYER.md`
**Related Code:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/`
**Implementations:** SQLite, MySQL, PostgreSQL

---

## Overview

ModulaCMS uses a **database abstraction layer** via the `DbDriver` interface. This allows the application to work with three different database systems (SQLite, MySQL, PostgreSQL) through a single, consistent API.

**Why This Matters:**
- Application code doesn't know which database it's using
- Switching databases is a configuration change, not a code change
- Testing uses SQLite in-memory databases
- Production typically uses MySQL or PostgreSQL
- Same codebase works for all database types

**Key Principle:** Program to an interface, not an implementation.

---

## Why Database Abstraction?

### Problem Without Abstraction

**Without abstraction, every database query is vendor-specific:**

```go
// MySQL version
func getContent(db *sql.DB) error {
    rows, err := db.Query("SELECT * FROM content LIMIT ? OFFSET ?", limit, offset)
    // ...
}

// PostgreSQL version
func getContent(db *sql.DB) error {
    rows, err := db.Query("SELECT * FROM content LIMIT $1 OFFSET $2", limit, offset)
    // ...
}
```

**Problems:**
- Code duplication for each database
- Hard to switch databases
- Can't test with different databases
- Business logic mixed with database specifics
- Error-prone (easy to forget differences)

### Solution: DbDriver Interface

**With abstraction:**

```go
// Define interface
type DbDriver interface {
    GetContent(ctx context.Context, routeID int64) ([]ContentData, error)
}

// Use interface
func publishContent(db DbDriver, contentID int64) error {
    content, err := db.GetContent(ctx, contentID)  // Works with any driver
    // ...
}
```

**Benefits:**
- ✅ Single API for all databases
- ✅ Easy to switch databases (configuration)
- ✅ Easy to test (mock interface)
- ✅ Consistent error handling
- ✅ Type-safe operations
- ✅ Clear separation of concerns

---

## The DbDriver Interface

The DbDriver interface defines all database operations available to the application.

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/driver.go`

### Interface Structure

```go
type DbDriver interface {
    // Connection management
    Close() error
    Ping(ctx context.Context) error

    // Transaction support
    BeginTx(ctx context.Context) (Tx, error)

    // Content operations
    GetContentTree(ctx context.Context, routeID int64) ([]GetContentTreeRow, error)
    GetContentData(ctx context.Context, contentID int64) (ContentData, error)
    CreateContentData(ctx context.Context, params CreateContentDataParams) (int64, error)
    UpdateContentData(ctx context.Context, params UpdateContentDataParams) error
    DeleteContentData(ctx context.Context, contentID int64) error

    // Datatype operations
    GetDatatypes(ctx context.Context) ([]Datatypes, error)
    GetDatatypeByID(ctx context.Context, datatypeID int64) (Datatypes, error)
    CreateDatatype(ctx context.Context, params CreateDatatypeParams) (int64, error)

    // Field operations
    GetFields(ctx context.Context) ([]Fields, error)
    GetFieldsByDatatype(ctx context.Context, datatypeID int64) ([]Fields, error)
    CreateField(ctx context.Context, params CreateFieldParams) (int64, error)

    // ... many more methods
}
```

**Key characteristics:**
- **Context-aware** - Every method takes `context.Context` for cancellation/timeouts
- **Error returns** - All methods return error as last value
- **Type-safe** - Uses generated types from sqlc
- **Consistent naming** - Get, Create, Update, Delete prefixes
- **Domain-focused** - Methods match business operations

### Design Principles

**1. Domain Operations, Not SQL**

**Bad (SQL-focused):**
```go
ExecuteQuery(query string, args ...interface{}) error
```

**Good (Domain-focused):**
```go
PublishContent(ctx context.Context, contentID int64) error
GetContentByStatus(ctx context.Context, routeID int64, status int32) ([]ContentData, error)
```

**2. Return Domain Types**

**Bad (Database types):**
```go
GetContent() (*sql.Rows, error)  // Caller deals with rows
```

**Good (Domain types):**
```go
GetContent() ([]ContentData, error)  // Returns structured data
```

**3. Parameters as Structs**

**Bad (Many parameters):**
```go
CreateContent(ctx context.Context, routeID int64, datatypeID int64,
              authorID int64, status int32, title string, ...) error
```

**Good (Struct parameter):**
```go
CreateContent(ctx context.Context, params CreateContentParams) (int64, error)

type CreateContentParams struct {
    RouteID    int64
    DatatypeID int64
    AuthorID   int64
    Status     int32
    // ...
}
```

---

## The Three Implementations

ModulaCMS implements DbDriver three times, once for each database.

### Implementation Overview

| Driver | Package | Use Case | File |
|--------|---------|----------|------|
| SQLite | `internal/db-sqlite` | Development, small sites | `driver.go` |
| MySQL | `internal/db-mysql` | Production (most common) | `driver.go` |
| PostgreSQL | `internal/db-psql` | Enterprise, large scale | `driver.go` |

### Common Structure

Each implementation has the same structure:

```
internal/db-sqlite/
├── driver.go           # Driver struct and interface implementation
├── models.go           # Generated by sqlc
├── content.sql.go      # Generated by sqlc (content queries)
├── datatype.sql.go     # Generated by sqlc (datatype queries)
├── field.sql.go        # Generated by sqlc (field queries)
└── db.go               # Generated by sqlc (database setup)
```

### Driver Struct

**Each driver has a similar struct:**

```go
// SQLite driver
type Driver struct {
    db      *sql.DB
    queries *Queries  // Generated by sqlc
}

// MySQL driver
type Driver struct {
    db      *sql.DB
    queries *Queries  // Generated by sqlc
}

// PostgreSQL driver
type Driver struct {
    db      *sql.DB
    queries *Queries  // Generated by sqlc
}
```

**Key components:**
- `db` - Standard library `*sql.DB` connection
- `queries` - sqlc-generated struct with type-safe query methods

### Example Implementation

**SQLite Driver (`internal/db-sqlite/driver.go`):**

```go
func (d *Driver) GetContentData(ctx context.Context, contentID int64) (db.ContentData, error) {
    // Call sqlc-generated method
    row, err := d.queries.GetContentData(ctx, contentID)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return db.ContentData{}, fmt.Errorf("content %d not found", contentID)
        }
        return db.ContentData{}, fmt.Errorf("failed to get content: %w", err)
    }

    // Convert from sqlc type to db type
    return db.ContentData{
        ContentDataID: row.ContentDataID,
        ParentID:      row.ParentID,
        FirstChildID:  row.FirstChildID,
        NextSiblingID: row.NextSiblingID,
        PrevSiblingID: row.PrevSiblingID,
        RouteID:       row.RouteID,
        DatatypeID:    row.DatatypeID,
        AuthorID:      row.AuthorID,
        DateCreated:   row.DateCreated,
        DateModified:  row.DateModified,
        History:       row.History,
    }, nil
}
```

**Key steps:**
1. Call sqlc-generated query method
2. Check for errors (including not found)
3. Convert from sqlc type to domain type
4. Return domain type

---

## Type Conversions

sqlc generates types specific to each database. The driver converts these to common domain types.

### Why Type Conversion?

**Problem:** sqlc generates different types for each database.

```go
// SQLite generated type
type ContentData struct {
    ContentDataID int64
    ParentID      sql.NullInt64
    // ...
}

// MySQL generated type (different package)
type ContentData struct {
    ContentDataID int64
    ParentID      sql.NullInt64
    // ...
}
```

**Solution:** Common domain types in `internal/db/`.

```go
// Domain type (internal/db/content_data.go)
package db

type ContentData struct {
    ContentDataID int64
    ParentID      sql.NullInt64
    // ...
}
```

**All drivers convert to this common type.**

### Conversion Patterns

**Pattern 1: Direct Mapping**

```go
// sqlc type → domain type
domainContent := db.ContentData{
    ContentDataID: sqlcContent.ContentDataID,
    RouteID:       sqlcContent.RouteID,
    DatatypeID:    sqlcContent.DatatypeID,
}
```

**Pattern 2: Slice Conversion**

```go
// Convert []sqlcType to []domainType
func (d *Driver) GetContentList(ctx context.Context, routeID int64) ([]db.ContentData, error) {
    rows, err := d.queries.GetContentList(ctx, routeID)
    if err != nil {
        return nil, err
    }

    result := make([]db.ContentData, len(rows))
    for i, row := range rows {
        result[i] = db.ContentData{
            ContentDataID: row.ContentDataID,
            RouteID:       row.RouteID,
            // ... map all fields
        }
    }

    return result, nil
}
```

**Pattern 3: Nested Structures**

```go
// Query returns joined data
type GetContentTreeRow struct {
    ContentDataID  int64
    RouteID        int64
    DatatypeLabel  string  // From JOIN with datatypes
    DatatypeType   string
}

// Convert to domain type with nested struct
domainTree := db.ContentTreeNode{
    Content: db.ContentData{
        ContentDataID: row.ContentDataID,
        RouteID:       row.RouteID,
    },
    Datatype: db.DatatypeInfo{
        Label: row.DatatypeLabel,
        Type:  row.DatatypeType,
    },
}
```

### NULL Handling

SQL NULL values are represented as `sql.Null*` types.

```go
// Database allows NULL
type ContentData struct {
    ParentID sql.NullInt64  // Can be NULL
    History  sql.NullString // Can be NULL
}

// Check if NULL
if content.ParentID.Valid {
    parentID := content.ParentID.Int64
    // Use parentID
} else {
    // No parent (NULL)
}

// Set to NULL
content.ParentID = sql.NullInt64{Valid: false}

// Set to value
content.ParentID = sql.NullInt64{Int64: 123, Valid: true}
```

---

## Error Handling

Consistent error handling across all drivers.

### Error Patterns

**Pattern 1: Not Found Errors**

```go
func (d *Driver) GetContentData(ctx context.Context, id int64) (db.ContentData, error) {
    row, err := d.queries.GetContentData(ctx, id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return db.ContentData{}, fmt.Errorf("content %d not found", id)
        }
        return db.ContentData{}, fmt.Errorf("failed to get content: %w", err)
    }
    // ...
}
```

**Pattern 2: Constraint Violations**

```go
func (d *Driver) CreateContent(ctx context.Context, params CreateContentParams) (int64, error) {
    id, err := d.queries.CreateContent(ctx, ...)
    if err != nil {
        // Check for foreign key violation
        if strings.Contains(err.Error(), "FOREIGN KEY constraint") {
            return 0, fmt.Errorf("invalid route_id or datatype_id")
        }
        return 0, fmt.Errorf("failed to create content: %w", err)
    }
    return id, nil
}
```

**Pattern 3: Context Cancellation**

```go
func (d *Driver) GetContentTree(ctx context.Context, routeID int64) ([]db.ContentData, error) {
    rows, err := d.queries.GetContentTree(ctx, routeID)
    if err != nil {
        if errors.Is(err, context.Canceled) {
            return nil, fmt.Errorf("query canceled")
        }
        if errors.Is(err, context.DeadlineExceeded) {
            return nil, fmt.Errorf("query timeout")
        }
        return nil, fmt.Errorf("failed to get content tree: %w", err)
    }
    // ...
}
```

### Error Wrapping

Always wrap errors with context using `%w`:

```go
// Good - preserves error chain
return fmt.Errorf("failed to get content: %w", err)

// Bad - loses error chain
return fmt.Errorf("failed to get content: %v", err)

// Bad - no context
return err
```

**Benefits of wrapping:**
- Preserve error chain for `errors.Is()` and `errors.As()`
- Add context for debugging
- Log meaningful error messages

---

## Transaction Support

Transactions ensure atomic operations across multiple queries.

### Transaction Interface

```go
type Tx interface {
    Commit() error
    Rollback() error

    // Same methods as DbDriver
    CreateContent(ctx context.Context, params CreateContentParams) (int64, error)
    UpdateContent(ctx context.Context, params UpdateContentParams) error
    // ...
}
```

### Transaction Pattern

```go
func (d *Driver) BeginTx(ctx context.Context) (db.Tx, error) {
    tx, err := d.db.BeginTx(ctx, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to begin transaction: %w", err)
    }

    return &Transaction{
        tx:      tx,
        queries: d.queries.WithTx(tx),
    }, nil
}

type Transaction struct {
    tx      *sql.Tx
    queries *Queries
}

func (t *Transaction) Commit() error {
    return t.tx.Commit()
}

func (t *Transaction) Rollback() error {
    return t.tx.Rollback()
}

func (t *Transaction) CreateContent(ctx context.Context, params CreateContentParams) (int64, error) {
    // Same implementation as driver, but uses t.queries (transaction)
    return t.queries.CreateContent(ctx, params)
}
```

### Using Transactions

```go
// Start transaction
tx, err := db.BeginTx(ctx)
if err != nil {
    return err
}
defer tx.Rollback()  // Rollback if not committed

// Perform operations
contentID, err := tx.CreateContent(ctx, contentParams)
if err != nil {
    return err  // Automatic rollback via defer
}

for _, field := range fields {
    err := tx.CreateFieldValue(ctx, contentID, field)
    if err != nil {
        return err  // Automatic rollback
    }
}

// Commit transaction
if err := tx.Commit(); err != nil {
    return err
}
```

---

## Connection Management

### Opening Connections

**Each driver has a constructor:**

```go
// SQLite
func NewDriver(dataSource string) (db.DbDriver, error) {
    sqlDB, err := sql.Open("sqlite3", dataSource)
    if err != nil {
        return nil, err
    }

    return &Driver{
        db:      sqlDB,
        queries: New(sqlDB),
    }, nil
}

// MySQL
func NewDriver(dataSource string) (db.DbDriver, error) {
    sqlDB, err := sql.Open("mysql", dataSource)
    if err != nil {
        return nil, err
    }

    // MySQL-specific configuration
    sqlDB.SetMaxOpenConns(25)
    sqlDB.SetMaxIdleConns(5)
    sqlDB.SetConnMaxLifetime(5 * time.Minute)

    return &Driver{
        db:      sqlDB,
        queries: New(sqlDB),
    }, nil
}
```

### Connection Pooling

**SQLite (no pooling needed):**
```go
sqlDB.SetMaxOpenConns(1)  // Single connection for SQLite
```

**MySQL (connection pooling):**
```go
sqlDB.SetMaxOpenConns(25)           // Max 25 concurrent connections
sqlDB.SetMaxIdleConns(5)            // Keep 5 idle connections
sqlDB.SetConnMaxLifetime(5 * time.Minute)  // Recycle after 5 minutes
```

**PostgreSQL (connection pooling):**
```go
sqlDB.SetMaxOpenConns(50)           // PostgreSQL handles more
sqlDB.SetMaxIdleConns(10)
sqlDB.SetConnMaxLifetime(10 * time.Minute)
```

### Closing Connections

```go
func (d *Driver) Close() error {
    if d.db != nil {
        return d.db.Close()
    }
    return nil
}

// Usage
defer db.Close()
```

### Health Checks

```go
func (d *Driver) Ping(ctx context.Context) error {
    return d.db.PingContext(ctx)
}

// Usage
if err := db.Ping(ctx); err != nil {
    log.Fatal("Database connection failed:", err)
}
```

---

## Adding New Methods

When adding new database operations, follow this process:

### Step 1: Write SQL Query

**File:** `sql/mysql/content.sql` (and postgres version)

```sql
-- name: GetContentByStatus :many
SELECT cd.*
FROM content_data cd
WHERE cd.route_id = ?
  AND cd.status = ?
ORDER BY cd.date_created DESC;
```

### Step 2: Generate Code

```bash
cd sql
make sqlc
```

This generates methods in `internal/db-mysql/content.sql.go`:

```go
func (q *Queries) GetContentByStatus(ctx context.Context, arg GetContentByStatusParams) ([]ContentData, error)
```

### Step 3: Add to DbDriver Interface

**File:** `internal/db/driver.go`

```go
type DbDriver interface {
    // ... existing methods ...

    GetContentByStatus(ctx context.Context, routeID int64, status int32) ([]ContentData, error)
}
```

### Step 4: Implement in All Drivers

**SQLite (`internal/db-sqlite/driver.go`):**

```go
func (d *Driver) GetContentByStatus(ctx context.Context, routeID int64, status int32) ([]db.ContentData, error) {
    rows, err := d.queries.GetContentByStatus(ctx, GetContentByStatusParams{
        RouteID: routeID,
        Status:  status,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to get content by status: %w", err)
    }

    result := make([]db.ContentData, len(rows))
    for i, row := range rows {
        result[i] = db.ContentData{
            ContentDataID: row.ContentDataID,
            RouteID:       row.RouteID,
            Status:        row.Status,
            // ... map all fields
        }
    }

    return result, nil
}
```

**Repeat for MySQL and PostgreSQL drivers.**

### Step 5: Test

```go
func TestGetContentByStatus(t *testing.T) {
    // Use SQLite in-memory for testing
    driver, err := sqlite.NewDriver(":memory:")
    if err != nil {
        t.Fatal(err)
    }
    defer driver.Close()

    // Create test data
    contentID, _ := driver.CreateContent(ctx, CreateContentParams{
        RouteID:    1,
        DatatypeID: 1,
        Status:     1,
    })

    // Test query
    results, err := driver.GetContentByStatus(ctx, 1, 1)
    if err != nil {
        t.Fatal(err)
    }

    if len(results) != 1 {
        t.Errorf("Expected 1 result, got %d", len(results))
    }

    if results[0].ContentDataID != contentID {
        t.Error("Wrong content returned")
    }
}
```

---

## Testing with Mock Drivers

The DbDriver interface enables easy testing with mocks.

### Creating Mock Driver

```go
type MockDriver struct {
    GetContentFunc func(ctx context.Context, id int64) (db.ContentData, error)
    // ... mock functions for all methods
}

func (m *MockDriver) GetContent(ctx context.Context, id int64) (db.ContentData, error) {
    if m.GetContentFunc != nil {
        return m.GetContentFunc(ctx, id)
    }
    return db.ContentData{}, nil
}

// Implement all interface methods...
```

### Using Mock in Tests

```go
func TestPublishContent(t *testing.T) {
    // Create mock
    mock := &MockDriver{
        GetContentFunc: func(ctx context.Context, id int64) (db.ContentData, error) {
            return db.ContentData{
                ContentDataID: id,
                Status:        0,  // Draft
            }, nil
        },
        UpdateContentFunc: func(ctx context.Context, params UpdateContentParams) error {
            if params.Status != 1 {
                t.Error("Expected status 1 (published)")
            }
            return nil
        },
    }

    // Test function with mock
    err := PublishContent(context.Background(), mock, 123)
    if err != nil {
        t.Error(err)
    }
}
```

---

## Database-Specific Considerations

### SQLite Specifics

**Advantages:**
- Single file database
- No server required
- Perfect for development and testing
- Fast for small datasets

**Limitations:**
- Limited concurrency (single writer)
- No true foreign key cascades (must enable)
- Case-insensitive LIKE by default

**Configuration:**
```go
dataSource := "file:modulacms.db?cache=shared&mode=rwc&_foreign_keys=1"
```

### MySQL Specifics

**Advantages:**
- Mature, well-tested
- Good performance at scale
- Wide ecosystem support
- Excellent for production

**Considerations:**
- Placeholders use `?` not `$1, $2`
- AUTO_INCREMENT for IDs
- Different date/time handling
- LIMIT syntax: `LIMIT ? OFFSET ?`

**Configuration:**
```go
dataSource := "user:password@tcp(localhost:3306)/modulacms?parseTime=true"
```

### PostgreSQL Specifics

**Advantages:**
- Advanced features (JSON, arrays, full-text search)
- Excellent concurrency
- Strong consistency guarantees
- Best for large scale

**Considerations:**
- Placeholders use `$1, $2, $3` not `?`
- SERIAL for auto-increment
- Case-sensitive by default
- LIMIT syntax: `LIMIT $1 OFFSET $2`

**Configuration:**
```go
dataSource := "host=localhost port=5432 user=postgres password=pass dbname=modulacms sslmode=disable"
```

---

## Runtime Database Selection

Database is selected at runtime via configuration.

### Configuration

**File:** `config.json`

```json
{
  "database": {
    "driver": "mysql",
    "dataSource": "user:pass@tcp(localhost:3306)/modulacms"
  }
}
```

### Driver Factory

```go
func NewDriverFromConfig(config Config) (db.DbDriver, error) {
    switch config.Database.Driver {
    case "sqlite", "sqlite3":
        return sqlite.NewDriver(config.Database.DataSource)

    case "mysql":
        return mysql.NewDriver(config.Database.DataSource)

    case "postgres", "postgresql":
        return psql.NewDriver(config.Database.DataSource)

    default:
        return nil, fmt.Errorf("unsupported database driver: %s", config.Database.Driver)
    }
}
```

### Usage

```go
// Load config
config := loadConfig()

// Create driver based on config
driver, err := NewDriverFromConfig(config)
if err != nil {
    log.Fatal(err)
}
defer driver.Close()

// Use driver (works regardless of which database)
content, err := driver.GetContent(ctx, contentID)
```

---

## Benefits of This Architecture

### 1. Database Portability

**Easy to switch databases:**
```json
// Development
{"database": {"driver": "sqlite", "dataSource": "dev.db"}}

// Production
{"database": {"driver": "mysql", "dataSource": "..."}}
```

No code changes needed.

### 2. Testing Flexibility

**Tests use SQLite in-memory:**
```go
driver, _ := sqlite.NewDriver(":memory:")
// Run tests...
```

Fast, isolated, no cleanup needed.

### 3. Type Safety

**Compile-time verification:**
```go
// Compiler catches errors
content, err := db.GetContent(ctx, "invalid")  // Compile error: string not int64
```

### 4. Consistent Error Handling

**All drivers handle errors the same way:**
```go
content, err := driver.GetContent(ctx, id)
if err != nil {
    // Same error handling for all databases
}
```

### 5. Clear Boundaries

**Business logic separated from database:**
```go
// Business logic (internal/model/)
func PublishContent(db db.DbDriver, id int64) error {
    // ...
}

// Database operations (internal/db-*/)
func (d *Driver) UpdateStatus(ctx context.Context, id int64, status int32) error {
    // ...
}
```

### 6. Simplified Application Code

**Application code is database-agnostic:**
```go
// This works with any database
func handlePublish(w http.ResponseWriter, r *http.Request) {
    content, err := app.db.GetContent(ctx, id)  // Don't care which DB
    // ...
}
```

---

## Trade-offs and Considerations

### Advantages

✅ **Portability** - Easy to switch databases
✅ **Testability** - Mock interface for unit tests
✅ **Type safety** - Compile-time checking
✅ **Consistency** - Single API for all databases
✅ **Maintainability** - Changes in one place

### Disadvantages

❌ **Abstraction overhead** - Extra layer of code
❌ **Conversion cost** - sqlc types → domain types
❌ **Feature limitations** - Can't use database-specific features
❌ **More verbose** - More code than raw SQL
❌ **Initial complexity** - Steeper learning curve

### When It's Worth It

**Good fit (ModulaCMS):**
- Supporting multiple databases is a requirement
- Testing with different databases
- Clear domain boundaries
- Type safety is important
- Long-term maintainability

**Not a good fit:**
- Single database forever
- Performance is critical (every microsecond counts)
- Heavy use of database-specific features
- Very simple application

---

## Related Documentation

**Database:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/database/SQL_DIRECTORY.md` - SQL file organization
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/database/SQLC.md` - sqlc code generation
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/database/DB_PACKAGE.md` - Database package details

**Architecture:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/CONTENT_MODEL.md` - Domain model
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/TREE_STRUCTURE.md` - Tree implementation

**Workflows:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/ADDING_FEATURES.md` - Feature development
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/ADDING_TABLES.md` - Creating tables

---

## Quick Reference

### DbDriver Interface Pattern

```go
type DbDriver interface {
    // Context-aware
    GetContent(ctx context.Context, id int64) (ContentData, error)

    // Return domain types
    // Error as last return
}
```

### Implementation Pattern

```go
func (d *Driver) GetContent(ctx context.Context, id int64) (db.ContentData, error) {
    // 1. Call sqlc query
    row, err := d.queries.GetContent(ctx, id)
    if err != nil {
        return db.ContentData{}, fmt.Errorf("context: %w", err)
    }

    // 2. Convert to domain type
    return db.ContentData{
        ID:    row.ID,
        Title: row.Title,
        // ...
    }, nil
}
```

### Adding New Methods

1. Write SQL query (`sql/mysql/*.sql`, `sql/postgres/*.sql`)
2. Run `make sqlc`
3. Add to DbDriver interface
4. Implement in all three drivers
5. Write tests

### Database Selection

```json
{"database": {"driver": "sqlite|mysql|postgres", "dataSource": "..."}}
```

### Key Files

```
internal/db/
├── driver.go           # DbDriver interface
├── content_data.go     # Domain types
└── ...

internal/db-sqlite/
├── driver.go           # SQLite implementation
└── ...

internal/db-mysql/
├── driver.go           # MySQL implementation
└── ...

internal/db-psql/
├── driver.go           # PostgreSQL implementation
└── ...
```

# DB_PACKAGE.md

Guidelines for working with the database package containing Go code and database abstraction layer.

**Directory Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db`

---

## Package Purpose

The `internal/db/` package contains:
- Database abstraction layer (DbDriver interface)
- Common database utilities and types
- Database interface definitions for all operations
- Helper functions for database management

**This package contains .go files, NOT .sql files.**

**Related packages:**
- `internal/db-sqlite/` - SQLite driver implementation
- `internal/db-mysql/` - MySQL driver implementation
- `internal/db-psql/` - PostgreSQL driver implementation

---

## Package Structure

```
internal/db/
├── *.go                        # Database interface definitions
├── old/                        # Legacy code (reference only)
└── sql/                        # SQL query definitions (for abstraction)
    ├── mysql/                  # MySQL-specific Go SQL
    ├── psql/                   # PostgreSQL-specific Go SQL
    └── sqlite/                 # SQLite-specific Go SQL
```

**Driver Implementations:**
```
internal/db-sqlite/             # SQLite driver (DbDriver implementation)
├── *.go                        # SQLite-specific Go code
└── [sqlc generated files]      # Type-safe query functions

internal/db-mysql/              # MySQL driver (DbDriver implementation)
├── *.go                        # MySQL-specific Go code
└── [sqlc generated files]      # Type-safe query functions

internal/db-psql/               # PostgreSQL driver (DbDriver implementation)
├── *.go                        # PostgreSQL-specific Go code
└── [sqlc generated files]      # Type-safe query functions
```

---

## DbDriver Interface

### Interface Definition

The `DbDriver` interface defines all database operations. All three database implementations (SQLite, MySQL, PostgreSQL) implement this interface.

**Location:** `internal/db/driver.go` (or similar)

**Purpose:**
- Provides unified API for database operations
- Allows switching databases via configuration
- Ensures consistent behavior across database types
- Enables testing with mock implementations

### Interface Methods

The DbDriver interface contains 150+ methods organized by domain:

**Connection Management:**
- `Connect() error`
- `Close() error`
- `Ping() error`

**User Operations:**
- `CreateUser(username, email, password string, roleID int64) error`
- `GetUserByID(userID int64) (*User, error)`
- `GetUserByEmail(email string) (*User, error)`
- `UpdateUser(userID int64, updates UserUpdates) error`
- `DeleteUser(userID int64) error`
- `ListUsers() ([]*User, error)`

**Content Operations:**
- `CreateContent(data ContentData) (int64, error)`
- `GetContentByID(contentID int64) (*ContentData, error)`
- `UpdateContent(contentID int64, updates ContentUpdates) error`
- `DeleteContent(contentID int64) error`
- `GetContentTree(routeID int64) ([]*ContentNode, error)`

**Datatype Operations:**
- `CreateDatatype(label, typeStr string) (int64, error)`
- `GetDatatype(datatypeID int64) (*Datatype, error)`
- `ListDatatypes() ([]*Datatype, error)`
- `UpdateDatatype(datatypeID int64, updates DatatypeUpdates) error`
- `DeleteDatatype(datatypeID int64) error`

**Field Operations:**
- `CreateField(label, data, typeStr string) (int64, error)`
- `GetField(fieldID int64) (*Field, error)`
- `GetFieldsForDatatype(datatypeID int64) ([]*Field, error)`
- `UpdateField(fieldID int64, updates FieldUpdates) error`
- `DeleteField(fieldID int64) error`

**Route Operations:**
- `CreateRoute(domain, path string) (int64, error)`
- `GetRoute(routeID int64) (*Route, error)`
- `ListRoutes() ([]*Route, error)`
- `UpdateRoute(routeID int64, updates RouteUpdates) error`
- `DeleteRoute(routeID int64) error`

**Media Operations:**
- `CreateMedia(bucket, key, filename string, metadata MediaMetadata) (int64, error)`
- `GetMedia(mediaID int64) (*Media, error)`
- `ListMedia() ([]*Media, error)`
- `DeleteMedia(mediaID int64) error`

**Session Operations:**
- `CreateSession(userID int64, token string, expiresAt int64) error`
- `GetSession(token string) (*Session, error)`
- `DeleteSession(token string) error`
- `CleanExpiredSessions() error`

**Permission & Role Operations:**
- `GetPermission(permissionID int64) (*Permission, error)`
- `ListPermissions() ([]*Permission, error)`
- `GetRole(roleID int64) (*Role, error)`
- `ListRoles() ([]*Role, error)`

**Transaction Support:**
- `BeginTx() (*Tx, error)`
- `CommitTx(tx *Tx) error`
- `RollbackTx(tx *Tx) error`

### Interface Example

```go
// internal/db/driver.go
package db

import "context"

type DbDriver interface {
    // Connection
    Connect() error
    Close() error
    Ping() error

    // Users
    CreateUser(ctx context.Context, username, email, password string, roleID int64) (int64, error)
    GetUserByID(ctx context.Context, userID int64) (*User, error)
    GetUserByEmail(ctx context.Context, email string) (*User, error)
    UpdateUserEmail(ctx context.Context, userID int64, email string) error
    DeleteUser(ctx context.Context, userID int64) error
    ListUsersByRole(ctx context.Context, roleID int64) ([]*User, error)

    // Content
    GetContentTree(ctx context.Context, routeID int64) ([]*ContentNode, error)
    CreateContent(ctx context.Context, data ContentData) (int64, error)
    UpdateContentField(ctx context.Context, fieldID int64, value string) error
    DeleteContent(ctx context.Context, contentID int64) error

    // ... 150+ more methods
}
```

---

## Working with Database Drivers

### Driver Selection

The application selects a database driver based on configuration:

```go
// Example from cmd/main.go or similar
import (
    "github.com/hegner123/modulacms/internal/config"
    "github.com/hegner123/modulacms/internal/db"
    "github.com/hegner123/modulacms/internal/db-sqlite"
    "github.com/hegner123/modulacms/internal/db-mysql"
    "github.com/hegner123/modulacms/internal/db-psql"
)

func initDatabase(cfg *config.Config) (db.DbDriver, error) {
    var driver db.DbDriver
    var err error

    switch cfg.DbDriver {
    case "sqlite":
        driver, err = dbsqlite.NewDriver(cfg.DbURL)
    case "mysql":
        driver, err = dbmysql.NewDriver(cfg.DbURL)
    case "postgres":
        driver, err = dbpsql.NewDriver(cfg.DbURL)
    default:
        return nil, fmt.Errorf("unknown database driver: %s", cfg.DbDriver)
    }

    if err != nil {
        return nil, err
    }

    if err := driver.Connect(); err != nil {
        return nil, err
    }

    return driver, nil
}
```

### Using the Driver

```go
// Example: Creating a user
func CreateNewUser(driver db.DbDriver, username, email, password string) error {
    ctx := context.Background()

    // Hash password
    passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return fmt.Errorf("failed to hash password: %w", err)
    }

    // Create user
    userID, err := driver.CreateUser(ctx, username, email, string(passwordHash), 2) // roleID 2 = user
    if err != nil {
        return fmt.Errorf("failed to create user: %w", err)
    }

    utility.DefaultLogger.Info("Created user", "userID", userID, "username", username)
    return nil
}
```

---

## Implementing a Driver

### Driver Implementation Structure

Each driver implementation:
1. Defines a struct that holds database connection
2. Implements all DbDriver interface methods
3. Uses sqlc-generated functions for type-safe queries
4. Handles database-specific behavior

### Example Driver Implementation

```go
// internal/db-sqlite/driver.go
package dbsqlite

import (
    "context"
    "database/sql"

    _ "github.com/mattn/go-sqlite3"
    "github.com/hegner123/modulacms/internal/db"
)

type SQLiteDriver struct {
    db      *sql.DB
    queries *Queries // sqlc-generated
}

func NewDriver(dbPath string) (*SQLiteDriver, error) {
    return &SQLiteDriver{}, nil
}

func (d *SQLiteDriver) Connect() error {
    db, err := sql.Open("sqlite3", d.dbPath)
    if err != nil {
        return err
    }

    // Enable foreign keys
    _, err = db.Exec("PRAGMA foreign_keys = ON;")
    if err != nil {
        return err
    }

    d.db = db
    d.queries = New(db) // sqlc-generated constructor
    return nil
}

func (d *SQLiteDriver) Close() error {
    if d.db != nil {
        return d.db.Close()
    }
    return nil
}

func (d *SQLiteDriver) GetUserByID(ctx context.Context, userID int64) (*db.User, error) {
    // Use sqlc-generated function
    row, err := d.queries.GetUserByID(ctx, userID)
    if err != nil {
        return nil, err
    }

    // Convert sqlc type to db package type
    return &db.User{
        UserID:   row.UserID,
        Username: row.Username,
        Email:    row.Email,
        RoleID:   row.RoleID,
        CreatedAt: row.CreatedAt,
    }, nil
}

func (d *SQLiteDriver) CreateUser(ctx context.Context, username, email, password string, roleID int64) (int64, error) {
    result, err := d.queries.CreateUser(ctx, CreateUserParams{
        Username:     username,
        Email:        email,
        PasswordHash: password,
        RoleID:       roleID,
    })
    if err != nil {
        return 0, err
    }

    userID, err := result.LastInsertId()
    if err != nil {
        return 0, err
    }

    return userID, nil
}

// ... implement all other DbDriver methods
```

---

## Adding New Database Operations

### Step 1: Write SQL Query

First, write the SQL query in the appropriate location:

**For SQLite:** Create query in `internal/db/sql/sqlite/` or `sql/sqlite/`
**For MySQL:** Create query in `sql/mysql/`
**For PostgreSQL:** Create query in `sql/postgres/`

See **SQL_DIRECTORY.md** for details on writing SQL queries.

### Step 2: Generate Go Code

Run sqlc to generate type-safe Go functions:
```bash
make sqlc
```

### Step 3: Add Method to DbDriver Interface

Add the new method to the DbDriver interface:

```go
// internal/db/driver.go
type DbDriver interface {
    // ... existing methods

    // New method
    GetUsersByStatus(ctx context.Context, status string) ([]*User, error)
}
```

### Step 4: Implement in Each Driver

Implement the method in all three drivers:

**SQLite:**
```go
// internal/db-sqlite/users.go
func (d *SQLiteDriver) GetUsersByStatus(ctx context.Context, status string) ([]*db.User, error) {
    rows, err := d.queries.GetUsersByStatus(ctx, status)
    if err != nil {
        return nil, err
    }

    users := make([]*db.User, len(rows))
    for i, row := range rows {
        users[i] = &db.User{
            UserID:   row.UserID,
            Username: row.Username,
            Email:    row.Email,
            Status:   row.Status,
        }
    }

    return users, nil
}
```

**MySQL:**
```go
// internal/db-mysql/users.go
func (d *MySQLDriver) GetUsersByStatus(ctx context.Context, status string) ([]*db.User, error) {
    // Similar implementation using MySQL-specific queries
}
```

**PostgreSQL:**
```go
// internal/db-psql/users.go
func (d *PostgreSQLDriver) GetUsersByStatus(ctx context.Context, status string) ([]*db.User, error) {
    // Similar implementation using PostgreSQL-specific queries
}
```

### Step 5: Use the New Method

```go
// Somewhere in application code
users, err := driver.GetUsersByStatus(ctx, "active")
if err != nil {
    utility.DefaultLogger.Error("Failed to get users", "error", err)
    return err
}

for _, user := range users {
    fmt.Printf("User: %s (%s)\n", user.Username, user.Email)
}
```

---

## Data Types and Models

### Defining Data Models

Data models are defined in `internal/db/types.go` (or similar):

```go
// internal/db/types.go
package db

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

type ContentNode struct {
    ContentDataID  int64
    ParentID       sql.NullInt64
    FirstChildID   sql.NullInt64
    NextSiblingID  sql.NullInt64
    PrevSiblingID  sql.NullInt64
    DatatypeID     int64
    DatatypeLabel  string
    DatatypeType   string
    RouteID        int64
    CreatedAt      int64
    UpdatedAt      sql.NullInt64
}

type Datatype struct {
    DatatypeID int64
    ParentID   sql.NullInt64
    Label      string
    Type       string
    CreatedAt  int64
    UpdatedAt  sql.NullInt64
}

type Field struct {
    FieldID    int64
    ParentID   sql.NullInt64
    Label      string
    Data       string
    Type       string
    CreatedAt  int64
    UpdatedAt  sql.NullInt64
}
```

### Handling NULL Values

Use `sql.NullInt64`, `sql.NullString`, etc. for nullable columns:

```go
type ContentData struct {
    ContentDataID int64
    ParentID      sql.NullInt64  // Can be NULL
    Title         string          // NOT NULL
    Description   sql.NullString // Can be NULL
}

// Checking for NULL
if contentData.ParentID.Valid {
    parentID := contentData.ParentID.Int64
    // Use parentID
} else {
    // ParentID is NULL
}

// Setting NULL
contentData.ParentID = sql.NullInt64{
    Int64: 0,
    Valid: false, // This makes it NULL
}

// Setting a value
contentData.ParentID = sql.NullInt64{
    Int64: 42,
    Valid: true,
}
```

---

## Transaction Support

### Using Transactions

```go
// Example with transaction
func TransferContent(driver db.DbDriver, fromRouteID, toRouteID, contentID int64) error {
    ctx := context.Background()

    // Begin transaction
    tx, err := driver.BeginTx()
    if err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer driver.RollbackTx(tx) // Rollback if not committed

    // Get content
    content, err := driver.GetContentByID(ctx, contentID)
    if err != nil {
        return fmt.Errorf("failed to get content: %w", err)
    }

    // Verify source route
    if content.RouteID != fromRouteID {
        return fmt.Errorf("content not in source route")
    }

    // Update route
    err = driver.UpdateContent(ctx, contentID, ContentUpdates{
        RouteID: toRouteID,
    })
    if err != nil {
        return fmt.Errorf("failed to update content: %w", err)
    }

    // Commit transaction
    if err := driver.CommitTx(tx); err != nil {
        return fmt.Errorf("failed to commit transaction: %w", err)
    }

    utility.DefaultLogger.Info("Transferred content", "contentID", contentID, "toRoute", toRouteID)
    return nil
}
```

---

## Error Handling

### Common Error Patterns

```go
import (
    "database/sql"
    "errors"
)

// Check for "not found" error
user, err := driver.GetUserByEmail(ctx, email)
if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
        return nil, fmt.Errorf("user not found")
    }
    return nil, fmt.Errorf("database error: %w", err)
}

// Check for duplicate entry (implementation-specific)
_, err = driver.CreateUser(ctx, username, email, password, roleID)
if err != nil {
    if strings.Contains(err.Error(), "UNIQUE constraint failed") {
        return fmt.Errorf("user already exists")
    }
    return fmt.Errorf("failed to create user: %w", err)
}
```

### Logging Database Errors

```go
import "github.com/hegner123/modulacms/internal/utility"

func GetUser(driver db.DbDriver, userID int64) (*db.User, error) {
    ctx := context.Background()

    user, err := driver.GetUserByID(ctx, userID)
    if err != nil {
        utility.DefaultLogger.Error(
            "Failed to get user",
            "userID", userID,
            "error", err,
        )
        return nil, err
    }

    return user, nil
}
```

---

## Testing Database Code

### Using SQLite for Tests

```go
// Example test
func TestCreateUser(t *testing.T) {
    // Create temporary SQLite database
    driver, err := dbsqlite.NewDriver(":memory:")
    if err != nil {
        t.Fatalf("Failed to create driver: %v", err)
    }
    defer driver.Close()

    if err := driver.Connect(); err != nil {
        t.Fatalf("Failed to connect: %v", err)
    }

    // Run test
    ctx := context.Background()
    userID, err := driver.CreateUser(ctx, "testuser", "test@example.com", "hash", 1)
    if err != nil {
        t.Fatalf("Failed to create user: %v", err)
    }

    if userID == 0 {
        t.Error("Expected non-zero user ID")
    }

    // Verify user was created
    user, err := driver.GetUserByID(ctx, userID)
    if err != nil {
        t.Fatalf("Failed to get user: %v", err)
    }

    if user.Username != "testuser" {
        t.Errorf("Expected username 'testuser', got '%s'", user.Username)
    }
}
```

### Mock Driver for Testing

```go
// internal/db/mock/mock_driver.go
package mock

import "github.com/hegner123/modulacms/internal/db"

type MockDriver struct {
    Users []*db.User
}

func (m *MockDriver) GetUserByID(ctx context.Context, userID int64) (*db.User, error) {
    for _, user := range m.Users {
        if user.UserID == userID {
            return user, nil
        }
    }
    return nil, sql.ErrNoRows
}

// ... implement other methods
```

---

## Best Practices

### Interface Design

1. **Keep interface focused**
   - Methods should have single responsibility
   - Group related operations
   - Return specific error types

2. **Use context.Context**
   - Pass context as first parameter
   - Respect context cancellation
   - Set timeouts for long operations

3. **Return meaningful errors**
   - Wrap errors with context
   - Use error types for specific cases
   - Log errors before returning

### Implementation

1. **Use sqlc-generated code**
   - Don't write raw SQL in Go
   - Let sqlc handle type safety
   - Trust generated code

2. **Handle NULL correctly**
   - Use sql.Null* types
   - Check Valid field before accessing value
   - Consider default values

3. **Log appropriately**
   - Log errors with context
   - Use structured logging
   - Don't log sensitive data (passwords, tokens)

4. **Close resources**
   - Always defer Close()
   - Use defer for rollback in transactions
   - Handle connection pooling properly

### Code Organization

1. **Organize by domain**
   - Group related methods in same file
   - `users.go` for user operations
   - `content.go` for content operations

2. **Consistent naming**
   - `CreateX` for INSERT
   - `GetX` for single SELECT
   - `ListX` for multiple SELECT
   - `UpdateX` for UPDATE
   - `DeleteX` for DELETE

3. **Document complex operations**
   - Add comments for business logic
   - Explain transaction requirements
   - Note performance implications

---

## Common Tasks

### Task: Add a new database operation

1. Write SQL query (see SQL_DIRECTORY.md)
2. Run `make sqlc` to generate Go code
3. Add method to DbDriver interface
4. Implement in all three drivers
5. Write tests
6. Use in application code

### Task: Switch database drivers

```go
// Change configuration
config.DbDriver = "postgres"  // was "sqlite"
config.DbURL = "postgres://user:pass@localhost/db"

// Driver is initialized based on config
driver, err := initDatabase(config)
```

### Task: Debug database queries

```go
// Enable query logging in driver
import "github.com/charmbracelet/log"

// In driver implementation
func (d *MySQLDriver) GetUserByID(ctx context.Context, userID int64) (*db.User, error) {
    utility.DefaultLogger.Debug("Getting user", "userID", userID)

    user, err := d.queries.GetUserByID(ctx, userID)
    if err != nil {
        utility.DefaultLogger.Error("Query failed", "error", err)
        return nil, err
    }

    utility.DefaultLogger.Debug("Got user", "username", user.Username)
    return convertUser(user), nil
}
```

### Task: View interface definition

```bash
# Read the driver interface
cat /Users/home/Documents/Code/Go_dev/modulacms/internal/db/driver.go
```

---

## Troubleshooting

### Interface not satisfied

**Error:** `*MySQLDriver does not implement DbDriver (missing method X)`

**Solution:**
1. Check interface definition in `internal/db/driver.go`
2. Implement missing method in driver
3. Ensure method signature matches exactly

### Type mismatch

**Error:** `cannot use row.UserID (type int64) as type int`

**Solution:**
- Check sqlc-generated types
- Ensure db package types match
- Cast if necessary: `int(row.UserID)`

### Connection issues

**SQLite: database locked**
- Close connections properly
- Use connection pool settings
- Check for long-running transactions

**MySQL: too many connections**
- Set MaxOpenConns
- Close unused connections
- Use connection pooling

**PostgreSQL: connection refused**
- Verify host and port
- Check firewall settings
- Confirm PostgreSQL is running

---

## Related Documentation

- **SQL_DIRECTORY.md** - Guide for working with sql/ directory and .sql files
- **FILE_TREE.md** - Complete directory structure
- **CLAUDE.md** - Project-wide development guidelines

---

## Quick Reference

### File Extensions
- `.go` - Go source files
- `_test.go` - Go test files

### Key Commands
```bash
# Generate Go code from SQL
make sqlc

# Run tests
make test

# Run specific package tests
go test -v ./internal/db-sqlite/
```

### Key Paths
- DB interface: `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/`
- SQLite driver: `/Users/home/Documents/Code/Go_dev/modulacms/internal/db-sqlite/`
- MySQL driver: `/Users/home/Documents/Code/Go_dev/modulacms/internal/db-mysql/`
- PostgreSQL driver: `/Users/home/Documents/Code/Go_dev/modulacms/internal/db-psql/`

### Import Paths
```go
import (
    "github.com/hegner123/modulacms/internal/db"
    "github.com/hegner123/modulacms/internal/db-sqlite"
    "github.com/hegner123/modulacms/internal/db-mysql"
    "github.com/hegner123/modulacms/internal/db-psql"
)
```

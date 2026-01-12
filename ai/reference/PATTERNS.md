# PATTERNS.md

Common Code Patterns in ModulaCMS

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/ai/reference/PATTERNS.md`
**Purpose:** Document recurring code patterns, conventions, and best practices used throughout the ModulaCMS codebase to ensure consistency and maintainability.
**Last Updated:** 2026-01-12

---

## Overview

This document catalogs the common patterns used in ModulaCMS. These patterns have been established to ensure consistency, readability, and maintainability across the codebase. When adding new features or modifying existing code, follow these established patterns to maintain code quality.

**Key Pattern Categories:**
- Error handling and propagation
- Logging strategies
- NULL handling for optional database fields
- Context usage for cancellation and timeouts
- Resource cleanup with defer
- Naming conventions
- File organization
- Database abstraction
- Type conversions

---

## Error Handling Patterns

### Immediate Error Checking

**Pattern:** Check errors immediately after they occur and handle them appropriately.

**Location:** Used throughout the codebase
**File Example:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/db.go:316-318`

```go
// ALWAYS check errors immediately
err := d.CreateUserTable()
if err != nil {
	return err
}

err = d.CreateRouteTable()
if err != nil {
	return err
}

err = d.CreateDatatypeFieldTable()
if err != nil {
	return err
}
```

**Key Points:**
- Never ignore errors
- Check immediately after the operation
- Return errors up the call stack
- Don't accumulate errors - fail fast

### Error Wrapping with Context

**Pattern:** Provide context when returning errors to make debugging easier.

**Location:** Used in database operations
**File Example:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/user.go:256-258`

```go
row, err := queries.CreateUser(d.Context, params)
if err != nil {
	e := fmt.Errorf("Failed to CreateUser.\n %v\n", err)
	return nil, e
}
```

**Key Points:**
- Use `fmt.Errorf` to wrap errors with context
- Include operation name in error message
- Return descriptive errors, not just the raw error
- Use `%v` or `%w` for error formatting (prefer `%w` when using Go 1.13+)

### Error Logging Before Exit

**Pattern:** Log errors using `utility.DefaultLogger` before returning or exiting.

**Location:** Application entry points and critical paths
**File Example:** `/Users/home/Documents/Code/Go_dev/modulacms/cmd/main.go:52-54`

```go
code, err := run()
if err != nil || code == ERRSIG {
	utility.DefaultLogger.Fatal("Root Return: ", err)
}
```

**File Example:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/db.go:746-748`

```go
script, err := sqlFiles.ReadFile("sql/dump_sql.sh")
if err != nil {
	utility.DefaultLogger.Fatal("failed to read embedded script: %v", err)
	return err
}
```

**Key Points:**
- Use `utility.DefaultLogger.Fatal()` for unrecoverable errors
- Use `utility.DefaultLogger.Error()` for recoverable errors
- Include context in log messages
- Fatal automatically exits with code 1

---

## Logging Patterns

### Using DefaultLogger

**Pattern:** Use `utility.DefaultLogger` singleton for all logging throughout the application.

**Location:** All packages
**File Path:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/utility/log.go:43`

```go
var DefaultLogger = NewLogger(BLANK)
```

**Usage Examples:**

```go
// Info logging
utility.DefaultLogger.Info("Starting SSH server", "ssh "+config.SSH_Host+" -p "+config.SSH_Port)

// Error logging with error object
utility.DefaultLogger.Error("Could not start server", err)

// Fatal logging (exits program)
utility.DefaultLogger.Fatal("Failed to load configuration", err)

// File logging (to debug.log)
utility.DefaultLogger.Finfo("Executing query:", query)
utility.DefaultLogger.Ferror("Database error", err)
```

**Log Levels Available:**
- `BLANK` - Raw output without formatting
- `DEBUG` - Development debugging information
- `INFO` - Informational messages
- `WARN` - Warning messages
- `ERROR` - Error messages (non-fatal)
- `FATAL` - Fatal errors (exits program)

**File Logging:**
All methods have file variants (prefix with `F`):
- `Fblank()` - Log to file without formatting
- `Fdebug()` - Debug to file
- `Finfo()` - Info to file
- `Fwarn()` - Warn to file
- `Ferror()` - Error to file
- `Ffatal()` - Fatal to file (and exit)

**Location:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/utility/log.go`

### Logging Pattern in Database Operations

**Pattern:** Log query execution for debugging and auditing.

**File Example:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/db.go:894-896`

```go
func (d Database) Query(db *sql.DB, query string) (sql.Result, error) {
	utility.DefaultLogger.Finfo("Executing query:", query)
	return db.Exec(query)
}
```

**Key Points:**
- Use `Finfo()` for file logging in database operations
- Log queries before execution
- Consistent format: operation description + relevant data

---

## NULL Handling Patterns

### sql.Null* Types for Optional Fields

**Pattern:** Use `sql.NullString`, `sql.NullInt64`, etc. for database fields that can be NULL.

**Location:** Database models
**File Example:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/user.go:23-24`

```go
type Users struct {
	UserID       int64          `json:"user_id"`
	Username     string         `json:"username"`
	Name         string         `json:"name"`
	Email        string         `json:"email"`
	Hash         string         `json:"hash"`
	Role         int64          `json:"role"`
	DateCreated  sql.NullString `json:"date_created"`
	DateModified sql.NullString `json:"date_modified"`
}
```

**File Example:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/getTree.go:18-22`

```go
type GetRouteTreeByRouteIDRow struct {
	ContentDataID int64          `json:"content_data_id"`
	ParentID      sql.NullInt64  `json:"parent_id"`
	FirstChildID  sql.NullInt64  `json:"first_child_id"`
	NextSiblingID sql.NullInt64  `json:"next_sibling_id"`
	PrevSiblingID sql.NullInt64  `json:"prev_sibling_id"`
	DatatypeLabel string         `json:"datatype_label"`
	DatatypeType  string         `json:"datatype_type"`
	FieldLabel    string         `json:"field_label"`
	FieldType     string         `json:"field_type"`
	FieldValue    sql.NullString `json:"field_value"`
}
```

**Common NULL Types:**
- `sql.NullString` - For VARCHAR/TEXT columns that can be NULL
- `sql.NullInt64` - For INTEGER/BIGINT columns that can be NULL
- `sql.NullInt32` - For INT columns that can be NULL (MySQL/PostgreSQL)
- `sql.NullBool` - For BOOLEAN columns that can be NULL
- `sql.NullFloat64` - For FLOAT/DOUBLE columns that can be NULL
- `sql.NullTime` - For TIMESTAMP/DATETIME columns that can be NULL

### Type Conversion Helpers

**Pattern:** Use helper functions for converting between NULL types and regular types.

**Location:** Type conversion functions in `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/`

**Common Conversions:**
```go
// String to NullString
func StringToNullString(s string) sql.NullString

// NullString to String
func NullStringToString(ns sql.NullString) string

// Int64 to NullInt32 (for MySQL/PostgreSQL)
func Int64ToNullInt32(i int64) sql.NullInt32

// NullInt32 to NullInt64
func NullInt32ToNullInt64(ni32 sql.NullInt32) sql.NullInt64

// Time conversions
func StringToNTime(s string) sql.NullTime
func TimeToNullString(t time.Time) sql.NullString
func NullTimeToString(nt sql.NullTime) string
```

### Accessing NULL Values

**Pattern:** Always check `Valid` field before accessing `sql.Null*` value.

```go
var user Users
// ... load user from database

// Check if field is valid (not NULL)
if user.DateCreated.Valid {
	fmt.Println("Created:", user.DateCreated.String)
} else {
	fmt.Println("Created: never")
}

// Similar pattern for NullInt64
if row.ParentID.Valid {
	parentID := row.ParentID.Int64
	// ... use parentID
}
```

**Key Points:**
- Always check `.Valid` before accessing `.String`, `.Int64`, etc.
- Invalid NULL values have zero values (empty string, 0, false)
- Use helper functions for conversions to avoid repetitive code

---

## Context Usage Patterns

### Context in Database Structs

**Pattern:** Store `context.Context` in database struct for all database operations.

**Location:** Database implementations
**File Example:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/db.go:27-35`

```go
type Database struct {
	Src            string
	Status         DbStatus
	Connection     *sql.DB
	LastConnection string
	Err            error
	Config         config.Config
	Context        context.Context
}
```

**Usage:**
```go
queries := mdb.New(d.Connection)
row, err := queries.GetUser(d.Context, id)
```

### Context with Timeout

**Pattern:** Create context with timeout for operations that should not block indefinitely.

**Location:** Application entry point
**File Example:** `/Users/home/Documents/Code/Go_dev/modulacms/cmd/main.go:67-68`

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
```

**Key Points:**
- Always defer `cancel()` to prevent context leaks
- Use timeouts for operations that can hang (network, database)
- Typical timeout: 30 seconds for initialization
- Pass context through call chain to all database operations

### Context Cancellation Pattern

**Pattern:** Use context with signal handling for graceful shutdown.

**File Example:** `/Users/home/Documents/Code/Go_dev/modulacms/cmd/main.go:64-65`

```go
done := make(chan os.Signal, 1)
signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
```

**Then later:**
```go
go func() {
	// ... server running
	<-done
	utility.DefaultLogger.Info("Stopping SSH Server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	// ... shutdown operations
}()
```

**Key Points:**
- Create signal channel for OS interrupt signals
- Use context with timeout for shutdown operations
- Always defer cancel to prevent leaks

---

## Resource Cleanup Patterns

### Defer for Cleanup

**Pattern:** Use `defer` to ensure resources are cleaned up even if errors occur.

**Location:** Throughout codebase
**File Example:** `/Users/home/Documents/Code/Go_dev/modulacms/cmd/main.go:124`

```go
if !InitStatus.DbFileExists || *app.ResetFlag {
	databaseConnection, _, _ := db.ConfigDB(*configuration).GetConnection()
	defer utility.HandleConnectionCloseDeferErr(databaseConnection)
}
```

**File Example:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/db.go:757-761`

```go
tmpFile, err := os.CreateTemp("", "embedded_script_*.sh")
if err != nil {
	utility.DefaultLogger.Fatal("failed to create temporary file: %v", err)
	return err
}
// Ensure the file is removed after execution.
defer func() {
	if closeErr := os.Remove(tmpFile.Name()); closeErr != nil && err == nil {
		err = closeErr
	}
}()
```

### Defer Pattern for Named Return Values

**Pattern:** Use defer with named return values to handle cleanup errors.

```go
defer func() {
	if closeErr := os.Remove(tmpFile.Name()); closeErr != nil && err == nil {
		err = closeErr
	}
}()
```

**Key Points:**
- Use named return values (`err error`) to capture cleanup errors
- Check if original error is nil before overwriting
- Execute cleanup in anonymous function
- Resources are cleaned up in reverse order of defer statements

### Helper Functions for Deferred Cleanup

**Pattern:** Create helper functions for common cleanup operations.

**Location:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/utility/helpers.go`

```go
func HandleRowsCloseDeferErr(r *sql.Rows) {
	err := r.Close()
	if err != nil {
		return
	}
}

func HandleConnectionCloseDeferErr(r *sql.DB) {
	err := r.Close()
	if err != nil {
		return
	}
}
```

**Usage:**
```go
rows, err := db.Query(query)
defer utility.HandleRowsCloseDeferErr(rows)
```

**Key Points:**
- Create helpers for repeated cleanup patterns
- Silently handle cleanup errors (already logged elsewhere)
- Use descriptive names: `Handle<Resource>Close<Action>`

---

## Transaction Patterns

### Using sqlc WithTx

**Pattern:** Use sqlc's `WithTx()` method to execute queries within a transaction.

**Location:** Transaction operations
**File Example:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db-sqlite/queries.sql.go:19-32`

```go
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
	return &Queries{
		db: tx,
	}
}
```

**Usage Pattern:**
```go
// Begin transaction
tx, err := db.Begin()
if err != nil {
	return err
}
defer tx.Rollback() // Rollback if not committed

// Use transaction with sqlc queries
queries := dbsqlite.New(db).WithTx(tx)
err = queries.CreateUser(ctx, params)
if err != nil {
	return err // Rollback via defer
}

err = queries.CreateSession(ctx, sessionParams)
if err != nil {
	return err // Rollback via defer
}

// Commit transaction
if err := tx.Commit(); err != nil {
	return err
}
```

**Key Points:**
- Always `defer tx.Rollback()` after beginning transaction
- Rollback is safe to call even after Commit
- Use `WithTx()` to create query instance scoped to transaction
- All queries in transaction use same context
- Commit explicitly when all operations succeed

---

## Naming Conventions

### Exported vs Unexported

**Pattern:** Use PascalCase for exported names, camelCase for unexported.

**Exported (Public API):**
```go
// Types
type Database struct { ... }
type Users struct { ... }

// Functions
func NewLogger(level LogLevel) *Logger { ... }
func CreateUser(params CreateUserParams) (*Users, error) { ... }

// Methods
func (d Database) GetUser(id int64) (*Users, error) { ... }
```

**Unexported (Internal):**
```go
// Functions
func formatLogMessage(level LogLevel, message string, err error, args ...any) string { ... }
func sanitizeCertDir(dir string) (string, error) { ... }

// Variables
var levelStyleMap = map[LogLevel]LogLevelStyle{ ... }
```

### Constants

**Pattern:** Use UPPERCASE for constants, enums use iota.

**File Example:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/utility/consts.go:3-10`

```go
type StorageUnit int64

const (
	KB StorageUnit = 1 << 10
	MB StorageUnit = 1 << 20
	GB StorageUnit = 1 << 30
	TB StorageUnit = 1 << 40
)

const (
	AppJson string = "application/json"
)
```

**File Example:** `/Users/home/Documents/Code/Go_dev/modulacms/cmd/main.go:44-47`

```go
type ReturnCode int16

const (
	OKSIG ReturnCode = iota
	ERRSIG
)
```

**File Example:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/db.go:62-66`

```go
type DbStatus string

const (
	Open   DbStatus = "open"
	Closed DbStatus = "closed"
	Err    DbStatus = "error"
)
```

**Key Points:**
- Typed constants for enums (not raw strings/ints)
- Use `iota` for sequential integer constants
- Group related constants together
- String constants for status values

### Function Naming

**Pattern:** Function names should be verb-based and descriptive.

**Common Prefixes:**
- `Create*` - Create new records
- `Get*` - Retrieve single record
- `List*` - Retrieve multiple records
- `Update*` - Modify existing record
- `Delete*` - Remove record
- `Count*` - Count records
- `Build*` - Construct complex objects
- `Map*` - Convert between types
- `Handle*` - Process events or operations
- `Init*` - Initialize components

**Examples:**
```go
func CreateUser(params CreateUserParams) (*Users, error)
func GetUser(id int64) (*Users, error)
func ListUsers() (*[]Users, error)
func UpdateUser(params UpdateUserParams) (*string, error)
func DeleteUser(id int64) error
func CountUsers() (*int64, error)
func BuildTree(cd []db.ContentData, ...) Root
func MapUser(a mdb.Users) Users
func HandleRowsCloseDeferErr(r *sql.Rows)
func InitDB(v *bool) error
```

### Type Naming

**Pattern:** Descriptive names that indicate purpose, avoid abbreviations unless widely known.

**Good Examples:**
```go
type Users struct { ... }
type CreateUserParams struct { ... }
type UpdateUserFormParams struct { ... }
type DbDriver interface { ... }
type GetRouteTreeByRouteIDRow struct { ... }
```

**Avoid:**
```go
type Usr struct { ... }        // Use Users
type CUP struct { ... }         // Use CreateUserParams
type DB struct { ... }          // Use Database (DB is acceptable only for well-known abbreviation)
```

### Variable Naming

**Pattern:** Clear, descriptive variable names. Single letters only for short loops.

**Good:**
```go
var configuration *config.Config
var databaseConnection *sql.DB
var queries *mdb.Queries
var timestamp string
```

**Acceptable for loops:**
```go
for i := range n { ... }
for _, v := range rows { ... }
for k, v := range map { ... }
```

---

## File Organization Patterns

### Package Structure

**Pattern:** One package per directory, files grouped by functionality.

**Common File Naming Patterns:**
- `<entity>.go` - Main entity definitions and logic
- `<entity>_test.go` - Tests for entity
- `<feature>_<variant>.go` - Variant implementations (e.g., `queries_mysql.sql.go`)
- `types.go` / `structs.go` - Type definitions
- `consts.go` / `constants.go` - Constants
- `helpers.go` / `util.go` - Utility functions
- `errors.go` - Error definitions

**Example from `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/`:**
```
db/
├── db.go                  # Core Database types and interface
├── user.go                # User CRUD operations
├── getTree.go             # Tree query operations
├── secure_query_builders.go
├── types.go               # Type definitions
└── ...
```

**Example from `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/`:**
```
cli/
├── model.go               # Main TUI model
├── init.go                # Initialization
├── update.go              # Update handler
├── update_cms.go          # CMS-specific updates
├── update_controls.go     # Control updates
├── update_database.go     # Database updates
├── view.go                # View rendering
├── message_types.go       # Message type definitions
├── commands.go            # Command definitions
└── ...
```

### Import Organization

**Pattern:** Group imports into three sections separated by blank lines.

**File Example:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/db.go:3-15`

```go
import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	config "github.com/hegner123/modulacms/internal/config"
	utility "github.com/hegner123/modulacms/internal/utility"
)
```

**Order:**
1. Standard library imports
2. (blank line)
3. Third-party imports (if any)
4. (blank line)
5. Project imports

**Key Points:**
- Go formatter (`go fmt`) handles sorting within each group
- Use import aliases for clarity when needed
- Project imports use full path from module root

---

## Database Abstraction Patterns

### DbDriver Interface Pattern

**Pattern:** Define interface with all database operations, implement for each database driver.

**Location:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/db.go:69-262`

```go
type DbDriver interface {
	// Database Connection
	CreateAllTables() error
	InitDB(v *bool) error
	Ping() error
	GetConnection() (*sql.DB, context.Context, error)

	// Count operations
	CountUsers() (*int64, error)
	CountRoutes() (*int64, error)
	// ... more count methods

	// Create operations
	CreateUser(CreateUserParams) (*Users, error)
	CreateRoute(CreateRouteParams) Routes
	// ... more create methods

	// Get operations
	GetUser(int64) (*Users, error)
	GetRoute(int64) (*Routes, error)
	// ... more get methods

	// List operations
	ListUsers() (*[]Users, error)
	ListRoutes() (*[]Routes, error)
	// ... more list methods

	// Update operations
	UpdateUser(UpdateUserParams) (*string, error)
	UpdateRoute(UpdateRouteParams) (*string, error)
	// ... more update methods

	// Delete operations
	DeleteUser(int64) error
	DeleteRoute(int64) error
	// ... more delete methods
}
```

**Implementation Pattern:**
```go
// SQLite implementation
type Database struct {
	Connection *sql.DB
	Context    context.Context
	// ...
}

func (d Database) GetUser(id int64) (*Users, error) {
	queries := mdb.New(d.Connection)
	row, err := queries.GetUser(d.Context, id)
	if err != nil {
		return nil, err
	}
	res := d.MapUser(row)
	return &res, nil
}

// MySQL implementation
type MysqlDatabase struct {
	Connection *sql.DB
	Context    context.Context
	// ...
}

func (d MysqlDatabase) GetUser(id int64) (*Users, error) {
	queries := mdbm.New(d.Connection)
	row, err := queries.GetUser(d.Context, int32(id))
	if err != nil {
		return nil, err
	}
	res := d.MapUser(row)
	return &res, nil
}

// PostgreSQL implementation - similar pattern
```

**Key Points:**
- All three databases implement same interface
- Type conversions handled in driver implementation
- Use sqlc-generated query functions
- Map database-specific types to common types

### Mapper Pattern

**Pattern:** Create `Map*` functions to convert between database-specific types and common types.

**Location:** Each database implementation
**File Example:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/user.go:193-204`

```go
// SQLite mapper
func (d Database) MapUser(a mdb.Users) Users {
	return Users{
		UserID:       a.UserID,
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         a.Role,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MySQL mapper
func (d MysqlDatabase) MapUser(a mdbm.Users) Users {
	return Users{
		UserID:       int64(a.UserID),
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         int64(a.Role),
		DateCreated:  StringToNullString(a.DateCreated.String()),
		DateModified: StringToNullString(a.DateModified.String()),
	}
}

// PostgreSQL mapper - similar pattern with sql.NullTime conversions
```

**Parameter Mapping Pattern:**
```go
// Common parameter type
type CreateUserParams struct {
	Username     string
	Name         string
	Email        string
	Hash         string
	Role         int64
	DateCreated  sql.NullString
	DateModified sql.NullString
}

// Map to database-specific parameter type
func (d Database) MapCreateUserParams(a CreateUserParams) mdb.CreateUserParams {
	return mdb.CreateUserParams{
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         a.Role,
		DateCreated:  a.DateCreated,
		DateModified: a.DateModified,
	}
}

// MySQL uses int32 and time.Time
func (d MysqlDatabase) MapCreateUserParams(a CreateUserParams) mdbm.CreateUserParams {
	return mdbm.CreateUserParams{
		Username:     a.Username,
		Name:         a.Name,
		Email:        a.Email,
		Hash:         a.Hash,
		Role:         int32(a.Role),
		DateCreated:  StringToNTime(a.DateCreated.String).Time,
		DateModified: StringToNTime(a.DateModified.String).Time,
	}
}
```

**Key Points:**
- One mapper per entity per database driver
- Mappers handle type conversions (int64 ↔ int32, sql.NullString ↔ time.Time)
- Parameter mappers convert common params to driver-specific params
- Result mappers convert driver-specific results to common types

### Using sqlc Generated Code

**Pattern:** Use sqlc-generated query functions, never write raw SQL in Go code.

```go
// Create queries instance from connection
queries := mdb.New(d.Connection)

// Single row query
row, err := queries.GetUser(d.Context, id)
if err != nil {
	return nil, err
}

// Multiple row query
rows, err := queries.ListUser(d.Context)
if err != nil {
	return nil, fmt.Errorf("failed to get Users: %v\n", err)
}

// Exec query (no return)
err := queries.CreateUserTable(d.Context)
if err != nil {
	return err
}
```

**Key Points:**
- `New()` creates queries instance from `*sql.DB` or `*sql.Tx`
- All queries accept `context.Context` as first parameter
- Query parameters follow as typed arguments
- Returns are type-safe structs generated by sqlc

---

## Type Conversion Patterns

### String Conversions

**Pattern:** Use standard library for basic conversions, helper functions for complex conversions.

**Basic Conversions:**
```go
import "strconv"

// String to int64
id := strconv.FormatInt(user.UserID, 10)

// Int64 to string
userID, err := strconv.ParseInt(idString, 10, 64)
```

**NULL Type Conversions:**
```go
// String to NullString
func StringToNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

// NullString to String
func NullStringToString(ns sql.NullString) string {
	if !ns.Valid {
		return ""
	}
	return ns.String
}
```

### Integer Type Conversions

**Pattern:** Convert between int64 (common) and int32 (MySQL/PostgreSQL).

```go
// Int64 to Int32 (for MySQL/PostgreSQL queries)
func Int64ToInt32(i int64) int32 {
	return int32(i)
}

// Int32 to Int64 (for common types)
func Int32ToInt64(i int32) int64 {
	return int64(i)
}

// NullInt32 to NullInt64
func NullInt32ToNullInt64(ni32 sql.NullInt32) sql.NullInt64 {
	if !ni32.Valid {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: int64(ni32.Int32), Valid: true}
}
```

**Usage in Database Calls:**
```go
// Common type uses int64
var id int64 = 123

// MySQL query requires int32
row, err := queries.GetUser(d.Context, int32(id))
```

---

## Interface Patterns

### Small, Focused Interfaces

**Pattern:** Prefer small interfaces that define specific capabilities.

**Example: DBTX Interface for sqlc**

**File Example:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db-sqlite/queries.sql.go:12-17`

```go
type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}
```

**Key Points:**
- Interface satisfied by both `*sql.DB` and `*sql.Tx`
- Allows same query code to work with or without transactions
- Only includes methods actually used

### Interface Embedding

**Pattern:** Embed interfaces to create composite interfaces.

```go
// Not found in current codebase, but common Go pattern
type Reader interface {
	Read(p []byte) (n int, err error)
}

type Writer interface {
	Write(p []byte) (n int, err error)
}

type ReadWriter interface {
	Reader
	Writer
}
```

---

## Constant Definition Patterns

### Typed Constants with iota

**Pattern:** Use custom types for constants and iota for sequential values.

**File Example:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/utility/log.go:17-26`

```go
type LogLevel int

const (
	BLANK LogLevel = iota
	DEBUG
	INFO
	WARN
	ERROR
	FATAL
)
```

**File Example:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/utility/consts.go:3-10`

```go
type StorageUnit int64

const (
	KB StorageUnit = 1 << 10  // 1024
	MB StorageUnit = 1 << 20  // 1048576
	GB StorageUnit = 1 << 30
	TB StorageUnit = 1 << 40
)
```

**Key Points:**
- Create custom type for enum-like constants
- Use `iota` for sequential integers
- Use bit shifting for powers of 2
- Typed constants provide compile-time type safety

### String Constants for Status

**Pattern:** Use string constants with custom type for status values.

**File Example:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/db.go:59-66`

```go
type DbStatus string

const (
	Open   DbStatus = "open"
	Closed DbStatus = "closed"
	Err    DbStatus = "error"
)
```

**Usage:**
```go
type Database struct {
	Status DbStatus
	// ...
}

db := Database{
	Status: Open,
}

if db.Status == Closed {
	// ...
}
```

**Key Points:**
- String constants are human-readable in logs/JSON
- Custom type prevents accidental string assignments
- Clear and explicit

---

## Common Anti-Patterns to Avoid

### Don't Ignore Errors

**❌ Bad:**
```go
user, _ := db.GetUser(id)  // Ignoring error!
fmt.Println(user.Name)      // Could panic if user is nil
```

**✅ Good:**
```go
user, err := db.GetUser(id)
if err != nil {
	return fmt.Errorf("failed to get user: %w", err)
}
fmt.Println(user.Name)
```

### Don't Use Naked Returns

**❌ Bad:**
```go
func GetUser(id int64) (user *Users, err error) {
	user, err = db.Get(id)
	if err != nil {
		return  // Naked return - unclear
	}
	return  // What are we returning?
}
```

**✅ Good:**
```go
func GetUser(id int64) (*Users, error) {
	user, err := db.Get(id)
	if err != nil {
		return nil, err
	}
	return user, nil
}
```

### Don't Access sql.Null* Without Checking Valid

**❌ Bad:**
```go
fmt.Println("Created:", user.DateCreated.String)  // Could be empty if NULL
```

**✅ Good:**
```go
if user.DateCreated.Valid {
	fmt.Println("Created:", user.DateCreated.String)
} else {
	fmt.Println("Created: unknown")
}
```

### Don't Forget to Defer cancel()

**❌ Bad:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
// Forgot to defer cancel() - context leak!
result, err := doSomething(ctx)
```

**✅ Good:**
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()  // Always defer immediately
result, err := doSomething(ctx)
```

### Don't Mix Database Types

**❌ Bad:**
```go
// Mixing SQLite types with MySQL types
var user mdb.Users
mysqlResult := mysqlQueries.GetUser(ctx, id)
user = mysqlResult  // Type mismatch!
```

**✅ Good:**
```go
// Use common types and mappers
mysqlResult := mysqlQueries.GetUser(ctx, id)
user := d.MapUser(mysqlResult)  // Convert to common Users type
```

---

## Related Documentation

### Architecture
- **[DATABASE_LAYER.md](../architecture/DATABASE_LAYER.md)** - Database abstraction philosophy
- **[TREE_STRUCTURE.md](../architecture/TREE_STRUCTURE.md)** - Tree implementation patterns
- **[TUI_ARCHITECTURE.md](../architecture/TUI_ARCHITECTURE.md)** - Elm Architecture patterns

### Workflows
- **[ADDING_FEATURES.md](../workflows/ADDING_FEATURES.md)** - Feature development workflow
- **[TESTING.md](../workflows/TESTING.md)** - Testing patterns and strategies
- **[DEBUGGING.md](../workflows/DEBUGGING.md)** - Debugging techniques

### Packages
- **[DB_PACKAGE.md](../DB_PACKAGE.md)** - Database package details
- **[CLI_PACKAGE.md](../packages/CLI_PACKAGE.md)** - TUI patterns
- **[MODEL_PACKAGE.md](../packages/MODEL_PACKAGE.md)** - Model patterns

### Database
- **[SQLC.md](../SQLC.md)** - sqlc usage patterns
- **[SQL_DIRECTORY.md](../SQL_DIRECTORY.md)** - SQL file organization

### Reference
- **[DEPENDENCIES.md](DEPENDENCIES.md)** - Why each dependency exists
- **[TROUBLESHOOTING.md](TROUBLESHOOTING.md)** - Common issues and solutions
- **[GLOSSARY.md](GLOSSARY.md)** - Term definitions

---

## Quick Reference

### Error Handling
```go
// Check immediately
if err != nil {
	return err
}

// Wrap with context
if err != nil {
	return fmt.Errorf("failed to create user: %w", err)
}

// Log before returning
if err != nil {
	utility.DefaultLogger.Error("Database error", err)
	return err
}
```

### Logging
```go
utility.DefaultLogger.Info("message", "extra", "args")
utility.DefaultLogger.Error("message", err)
utility.DefaultLogger.Fatal("message", err)  // Exits
utility.DefaultLogger.Finfo("file log", "message")
```

### NULL Handling
```go
// Define with sql.Null*
DateCreated sql.NullString `json:"date_created"`

// Check before use
if row.DateCreated.Valid {
	fmt.Println(row.DateCreated.String)
}
```

### Context
```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Pass to all database operations
result, err := queries.GetUser(ctx, id)
```

### Defer Cleanup
```go
// Simple defer
defer connection.Close()

// Defer with error handling
defer func() {
	if err := cleanup(); err != nil {
		log.Error("cleanup failed", err)
	}
}()

// Helper functions
defer utility.HandleConnectionCloseDeferErr(db)
```

### Database Abstraction
```go
// Use interface
var driver db.DbDriver

// Call methods
user, err := driver.GetUser(id)

// Map types
commonUser := d.MapUser(sqliteUser)
```

### Naming
- **Exported:** PascalCase (Users, CreateUser)
- **Unexported:** camelCase (formatLog, sanitizePath)
- **Constants:** UPPERCASE (OKSIG, ERRSIG)
- **Type Constants:** PascalCase (Open, Closed)

---

**Last Updated:** 2026-01-12

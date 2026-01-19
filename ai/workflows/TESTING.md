# TESTING.md

Comprehensive testing strategies and patterns for ModulaCMS.

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/TESTING.md`
**Purpose:** Guide for writing tests, organizing test code, and testing strategies
**Last Updated:** 2026-01-12

---

## Overview

ModulaCMS uses Go's built-in testing framework with specific patterns for database testing, TUI testing, and integration testing. This guide covers:

1. **Test organization** - Where tests live and how to structure them
2. **Unit testing** - Testing individual functions and methods
3. **Database testing** - Testing with SQLite in-memory databases
4. **TUI testing** - Testing Bubbletea components
5. **Integration testing** - End-to-end testing
6. **Mocking strategies** - Creating test fixtures
7. **Test coverage** - Measuring and improving coverage
8. **CI/CD integration** - Running tests in automation

---

## Prerequisites

**Before writing tests, you should understand:**

- **Go testing basics** - `testing` package, `t.Error`, `t.Fatal`
- **Table-driven tests** - Common Go testing pattern
- **Test fixtures** - Mock data and test setup
- **Database transactions** - For isolated database tests
- **Bubbletea architecture** - For TUI component testing

**External Resources:**
- [Go Testing Documentation](https://pkg.go.dev/testing)
- [Table-Driven Tests in Go](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [Testing with Database/SQL](https://go.dev/doc/database/sql-injection)

**Related Documentation:**
- [ADDING_TABLES.md](ADDING_TABLES.md) - Database table creation (includes test examples)
- [CREATING_TUI_SCREENS.md](CREATING_TUI_SCREENS.md) - TUI screen creation (includes test examples)
- [DB_PACKAGE.md](../DB_PACKAGE.md) - Database abstraction layer
- [CLI_PACKAGE.md](../packages/CLI_PACKAGE.md) - CLI/TUI package reference

---

## Test Organization

### File Naming Convention

Test files must end with `_test.go`:

```
internal/
├── db/
│   ├── db.go
│   ├── db_test.go           # Tests for db.go
│   ├── permission.go
│   └── foreignKey_test.go   # Tests for foreign key logic
├── model/
│   ├── model.go
│   ├── model_test.go        # Tests for model.go
│   └── build_test.go        # Tests for build functions
└── cli/
    ├── model.go
    └── update_test.go       # Tests for update functions
```

### Test Package Naming

Tests can be in the same package or a separate `_test` package:

**Same package (white-box testing):**
```go
package db

import "testing"

// Can access unexported functions
func TestInternalFunction(t *testing.T) {
    result := internalHelper()
    // ...
}
```

**Separate package (black-box testing):**
```go
package db_test

import (
    "testing"
    "github.com/hegner123/modulacms/internal/db"
)

// Only access exported functions
func TestPublicAPI(t *testing.T) {
    d := db.ConfigDB(config)
    // ...
}
```

**Recommendation:** Use same package for most tests in ModulaCMS to access internal helpers.

---

## Running Tests

### Basic Commands

```bash
# Run all tests
make test

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test -v ./internal/db
go test -v ./internal/model

# Run specific test
go test -v ./internal/db -run TestPermissionCRUD

# Run tests matching pattern
go test -v ./internal/db -run TestCreate

# Run tests with race detector
go test -race ./...

# Run tests with coverage
make coverage
```

### Test Targets in Makefile

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/Makefile`

```makefile
test: ## Run all tests
	touch testdb/create_tests.db
	touch ./testdb/testing2348263.db
	rm ./testdb/*.db

	touch ./backups/tmp.zip
	rm ./backups/*.zip
	$(GOTEST) -v ./...
	rm ./testdb/*.db

coverage: ## Run tests with coverage
	$(GOTEST) -cover -covermode=count -coverprofile=profile.cov ./...
	$(GOCMD) tool cover -func profile.cov

test-development: ## Run tests for specific package
	$(GOTEST) -v ./internal/development
```

**Important:** The test target cleans up test databases before and after running.

---

## Unit Testing Basics

### Simple Unit Test

**File:** `internal/utility/timestamp_test.go`

```go
package utility

import (
    "testing"
    "time"
)

func TestTimestampS(t *testing.T) {
    timestamp := TimestampS()

    if timestamp == "" {
        t.Error("TimestampS returned empty string")
    }

    // Check format (should be YYYYMMDD_HHMMSS)
    if len(timestamp) != 15 { // YYYYMMDD_HHMMSS
        t.Errorf("Expected timestamp length 15, got %d", len(timestamp))
    }
}

func TestParseTimestamp(t *testing.T) {
    testCases := []struct {
        name     string
        input    string
        expected time.Time
        wantErr  bool
    }{
        {
            name:     "valid timestamp",
            input:    "20260112_143022",
            expected: time.Date(2026, 1, 12, 14, 30, 22, 0, time.UTC),
            wantErr:  false,
        },
        {
            name:    "invalid format",
            input:   "invalid",
            wantErr: true,
        },
        {
            name:    "empty string",
            input:   "",
            wantErr: true,
        },
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result, err := ParseTimestamp(tc.input)

            if tc.wantErr {
                if err == nil {
                    t.Error("Expected error, got nil")
                }
                return
            }

            if err != nil {
                t.Fatalf("Unexpected error: %v", err)
            }

            if !result.Equal(tc.expected) {
                t.Errorf("Expected %v, got %v", tc.expected, result)
            }
        })
    }
}
```

### Table-Driven Tests

Table-driven tests are the preferred pattern in Go:

```go
func TestStringToInt64(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected int64
        wantErr  bool
    }{
        {"valid number", "123", 123, false},
        {"zero", "0", 0, false},
        {"negative", "-456", -456, false},
        {"invalid", "abc", 0, true},
        {"empty", "", 0, true},
        {"overflow", "9223372036854775808", 0, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := StringToInt64(tt.input)

            if tt.wantErr {
                if result != 0 {
                    t.Errorf("Expected 0 for invalid input, got %d", result)
                }
                return
            }

            if result != tt.expected {
                t.Errorf("Expected %d, got %d", tt.expected, result)
            }
        })
    }
}
```

**Benefits:**
- Easy to add new test cases
- Each case runs independently
- Clear test failures with case names
- Self-documenting test expectations

---

## Database Testing

### Test Database Setup

ModulaCMS uses SQLite for testing with a dedicated `testdb/` directory.

**Directory Structure:**
```
testdb/
├── create_tests.db    # Created and deleted by test runs
└── testing2348263.db  # Test database file
```

### Basic Database Test

**File:** `internal/db/permission_test.go`

```go
package db

import (
    "testing"
    "github.com/hegner123/modulacms/internal/config"
)

func setupTestDB(t *testing.T) DbDriver {
    t.Helper()

    // Load test configuration
    p := config.NewFileProvider("")
    m := config.NewManager(p)
    c, err := m.Config()
    if err != nil {
        t.Fatalf("Failed to load config: %v", err)
    }

    // Override database to use test database
    c.Db_Driver = "sqlite"
    c.Db_URL = "./testdb/test.db"

    // Create database connection
    d := ConfigDB(*c)

    // Initialize tables
    err = d.CreateAllTables()
    if err != nil {
        t.Fatalf("Failed to create tables: %v", err)
    }

    return d
}

func cleanupTestDB(t *testing.T, d DbDriver) {
    t.Helper()

    // Close connection
    con, _, err := d.GetConnection()
    if err != nil {
        t.Logf("Warning: Failed to get connection for cleanup: %v", err)
        return
    }

    if err := con.Close(); err != nil {
        t.Logf("Warning: Failed to close connection: %v", err)
    }
}

func TestPermissionCRUD(t *testing.T) {
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    // Create
    createParams := CreatePermissionParams{
        TableID: 1,
        Mode:    1,
        Label:   "create_test",
    }
    permission := db.CreatePermission(createParams)

    if permission.PermissionID == 0 {
        t.Fatal("CreatePermission failed: permission_id is 0")
    }

    if permission.Label != "create_test" {
        t.Errorf("Expected label 'create_test', got '%s'", permission.Label)
    }

    // Read
    fetchedPermission, err := db.GetPermission(permission.PermissionID)
    if err != nil {
        t.Fatalf("GetPermission failed: %v", err)
    }

    if fetchedPermission.PermissionID != permission.PermissionID {
        t.Errorf("Expected ID %d, got %d",
            permission.PermissionID,
            fetchedPermission.PermissionID)
    }

    // Update
    updateParams := UpdatePermissionParams{
        TableID:      2,
        Mode:         2,
        Label:        "updated_label",
        PermissionID: permission.PermissionID,
    }
    err = db.UpdatePermission(updateParams)
    if err != nil {
        t.Fatalf("UpdatePermission failed: %v", err)
    }

    updatedPermission, _ := db.GetPermission(permission.PermissionID)
    if updatedPermission.Label != "updated_label" {
        t.Errorf("Label not updated. Expected 'updated_label', got '%s'",
            updatedPermission.Label)
    }

    // Delete
    err = db.DeletePermission(permission.PermissionID)
    if err != nil {
        t.Fatalf("DeletePermission failed: %v", err)
    }

    // Verify deletion
    _, err = db.GetPermission(permission.PermissionID)
    if err == nil {
        t.Error("Permission still exists after deletion")
    }
}
```

### Testing with Transactions

Use transactions for isolated tests:

```go
func TestWithTransaction(t *testing.T) {
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    con, ctx, err := db.GetConnection()
    if err != nil {
        t.Fatal(err)
    }

    // Begin transaction
    tx, err := con.BeginTx(ctx, nil)
    if err != nil {
        t.Fatal(err)
    }
    defer tx.Rollback() // Rollback if test fails

    // Perform test operations using tx
    // ...

    // Commit if test passes
    if err := tx.Commit(); err != nil {
        t.Fatal(err)
    }
}
```

### Testing Foreign Key Constraints

**File:** `internal/db/foreignKey_test.go`

```go
package db

import (
    "fmt"
    "testing"
    "github.com/hegner123/modulacms/internal/config"
)

func TestForeignKeyConstraints(t *testing.T) {
    p := config.NewFileProvider("")
    m := config.NewManager(p)
    c, err := m.Config()
    if err != nil {
        t.Fatal(err)
    }

    d := ConfigDB(*c)
    args := []string{"admin_content_data"}
    rows := d.GetForeignKeys(args)

    if rows == nil {
        t.Fatal("GetForeignKeys returned nil")
    }

    r := d.ScanForeignKeyQueryRows(rows)

    if len(r) == 0 {
        t.Error("Expected foreign keys, got none")
    }

    for _, row := range r {
        fmt.Printf("FK: id=%d, seq=%d, table=%s, from=%s, to=%s\n",
            row.id, row.seq, row.tableName, row.fromCol, row.toCol)

        // Verify referenced table exists
        data, ok := d.(Database)
        if !ok {
            t.Fatal("Failed to type assert to Database")
        }

        data.SelectColumnFromTable(row.tableName, row.toCol)
    }
}
```

---

## Mock Data and Fixtures

### Creating Mock Data

**File:** `internal/model/model_test.go`

```go
package model

import (
    "database/sql"
    "github.com/hegner123/modulacms/internal/db"
)

// Mock data creation functions
func createMockDatatype() db.DatatypeJSON {
    p := sql.NullInt64{Valid: false, Int64: 0}
    dc := sql.NullString{String: "2023-01-01", Valid: true}
    dm := sql.NullString{String: "2023-01-01", Valid: true}
    h := sql.NullString{String: "", Valid: false}

    return db.DatatypeJSON{
        DatatypeID: 1,
        ParentID: db.NullInt64{
            NullInt64: p,
        },
        Label:    "Page",
        Type:     "content",
        AuthorID: 1,
        DateCreated: db.NullString{
            NullString: dc,
        },
        DateModified: db.NullString{
            NullString: dm,
        },
        History: db.NullString{
            NullString: h,
        },
    }
}

func createMockContentData() db.ContentDataJSON {
    p := sql.NullInt64{Int64: 0, Valid: false}
    dc := sql.NullString{String: "2023-01-01", Valid: true}
    dm := sql.NullString{String: "2023-01-01", Valid: true}
    h := sql.NullString{String: "", Valid: false}

    return db.ContentDataJSON{
        ContentDataID: 1,
        ParentID: db.NullInt64{
            NullInt64: p,
        },
        RouteID:    1,
        DatatypeID: 1,
        AuthorID:   1,
        DateCreated: db.NullString{
            NullString: dc,
        },
        DateModified: db.NullString{
            NullString: dm,
        },
        History: db.NullString{
            NullString: h,
        },
    }
}

func createMockField() db.FieldsJSON {
    p := sql.NullInt64{Int64: 0, Valid: false}
    dc := sql.NullString{String: "2023-01-01", Valid: true}
    dm := sql.NullString{String: "2023-01-01", Valid: true}
    h := sql.NullString{String: "", Valid: false}

    return db.FieldsJSON{
        FieldID: 1,
        ParentID: db.NullInt64{
            NullInt64: p,
        },
        Label:    "Title",
        Data:     "string",
        Type:     "text",
        AuthorID: 1,
        DateCreated: db.NullString{
            NullString: dc,
        },
        DateModified: db.NullString{
            NullString: dm,
        },
        History: db.NullString{
            NullString: h,
        },
    }
}

// Using mock data in tests
func TestCreateNode(t *testing.T) {
    datatype := createMockDatatype()
    contentData := createMockContentData()
    field := createMockField()

    node := &Node{
        Datatype: Datatype{
            Info:    datatype,
            Content: contentData,
        },
        Fields: []Field{
            {
                Info:    field,
                Content: createMockContentField(),
            },
        },
    }

    if node.Datatype.Info.DatatypeID != 1 {
        t.Errorf("Expected DatatypeID 1, got %d",
            node.Datatype.Info.DatatypeID)
    }
}
```

### Mock Data Best Practices

**1. Use helper functions:**
```go
func createMockX() db.X {
    // Create and return mock data
}
```

**2. Create variations for different scenarios:**
```go
func createMockUserAdmin() db.User {
    return db.User{UserID: 1, Role: "admin"}
}

func createMockUserRegular() db.User {
    return db.User{UserID: 2, Role: "user"}
}
```

**3. Make mocks realistic:**
```go
// Bad - unrealistic data
user := db.User{
    Username: "a",
    Email:    "x",
}

// Good - realistic data
user := db.User{
    Username: "testuser",
    Email:    "testuser@example.com",
}
```

**4. Avoid mock data duplication:**
```go
// Create a test fixtures file
// internal/db/fixtures_test.go

package db

var TestUser1 = User{
    UserID:   1,
    Username: "alice",
    Email:    "alice@example.com",
}

var TestUser2 = User{
    UserID:   2,
    Username: "bob",
    Email:    "bob@example.com",
}
```

---

## Testing Tree Structures

### Tree Construction Tests

**File:** `internal/model/build_test.go`

```go
package model

import (
    "testing"
)

func createMockNode() Root {
    root := NewRoot()

    child := &Node{
        Datatype: Datatype{
            Info:    createMockDatatype2(),
            Content: createMockContentData2(),
        },
        Fields: []Field{
            {
                Info:    createMockField2(),
                Content: createMockContentField2(),
            },
        },
        Nodes: nil,
    }

    r := AddChild(root, child)

    // Add more children
    child2 := &Node{
        Datatype: Datatype{
            Info:    createMockDatatype2(),
            Content: createMockContentData2(),
        },
        Fields: []Field{
            {
                Info:    createMockField2(),
                Content: createMockContentField2(),
            },
        },
        Nodes: nil,
    }

    r.Node.AddChild(child2)

    return r
}

func TestTreeCreation(t *testing.T) {
    tree := createMockNode()

    if tree.Node == nil {
        t.Fatal("Tree root node is nil")
    }

    rendered := tree.Render()
    if len(rendered) < 1 {
        t.Error("Tree render returned empty result")
    }
}

func TestTreeTraversal(t *testing.T) {
    tree := createMockNode()

    // Count nodes
    count := 0
    tree.Walk(func(n *Node) {
        count++
    })

    expectedCount := 3 // root + 2 children
    if count != expectedCount {
        t.Errorf("Expected %d nodes, got %d", expectedCount, count)
    }
}

func TestTreeDepth(t *testing.T) {
    tree := createMockNode()

    depth := tree.MaxDepth()

    if depth < 1 {
        t.Errorf("Expected depth >= 1, got %d", depth)
    }
}
```

### Testing Tree Operations

```go
func TestAddChild(t *testing.T) {
    root := NewRoot()
    child := createMockNode().Node

    result := AddChild(root, child)

    if result.Node.Nodes == nil {
        t.Fatal("AddChild did not initialize Nodes")
    }

    if len(result.Node.Nodes) != 1 {
        t.Errorf("Expected 1 child, got %d", len(result.Node.Nodes))
    }
}

func TestRemoveChild(t *testing.T) {
    root := createMockNode()
    initialCount := len(root.Node.Nodes)

    if initialCount == 0 {
        t.Skip("No children to remove")
    }

    childToRemove := root.Node.Nodes[0]
    result := RemoveChild(root, childToRemove)

    if len(result.Node.Nodes) != initialCount-1 {
        t.Errorf("Expected %d children after removal, got %d",
            initialCount-1, len(result.Node.Nodes))
    }
}
```

---

## TUI Testing

### Testing Update Functions

**File:** `internal/cli/update_test.go`

```go
package cli

import (
    "testing"
    tea "github.com/charmbracelet/bubbletea"
    "github.com/hegner123/modulacms/internal/config"
    "github.com/hegner123/modulacms/internal/db"
)

func createTestModel() Model {
    c := &config.Config{
        Db_Driver: "sqlite",
        Db_URL:    ":memory:",
    }

    m, _ := InitialModel(nil, c)
    return m
}

func TestNavigationUpdate(t *testing.T) {
    m := createTestModel()

    // Test navigation message
    msg := NavigateToPage{
        Page: NewPage(DATABASEPAGE, "Database"),
    }

    newM, cmd := m.Update(msg)

    if cmd == nil {
        t.Error("Expected command from navigation, got nil")
    }

    model := newM.(Model)
    if model.Page.Index != DATABASEPAGE {
        t.Errorf("Expected page index %d, got %d",
            DATABASEPAGE, model.Page.Index)
    }
}

func TestCursorMovement(t *testing.T) {
    m := createTestModel()
    m.CursorMax = 10
    m.Cursor = 5

    // Test cursor down
    msg := tea.KeyMsg{Type: tea.KeyDown}
    newM, _ := m.Update(msg)
    model := newM.(Model)

    if model.Cursor != 6 {
        t.Errorf("Expected cursor 6, got %d", model.Cursor)
    }

    // Test cursor up
    msg = tea.KeyMsg{Type: tea.KeyUp}
    newM, _ = model.Update(msg)
    model = newM.(Model)

    if model.Cursor != 5 {
        t.Errorf("Expected cursor 5, got %d", model.Cursor)
    }
}

func TestCursorBounds(t *testing.T) {
    m := createTestModel()
    m.CursorMax = 5
    m.Cursor = 5

    // Try to move down beyond max
    msg := tea.KeyMsg{Type: tea.KeyDown}
    newM, _ := m.Update(msg)
    model := newM.(Model)

    if model.Cursor != 5 {
        t.Errorf("Cursor exceeded max: expected 5, got %d", model.Cursor)
    }

    // Try to move up beyond 0
    model.Cursor = 0
    msg = tea.KeyMsg{Type: tea.KeyUp}
    newM, _ = model.Update(msg)
    model = newM.(Model)

    if model.Cursor != 0 {
        t.Errorf("Cursor went below 0: got %d", model.Cursor)
    }
}
```

### Testing Message Handlers

```go
func TestCommentListFetchedMsg(t *testing.T) {
    m := createTestModel()
    m.Loading = true

    comments := []db.Comment{
        {CommentID: 1, CommentText: "Test 1", Status: "pending"},
        {CommentID: 2, CommentText: "Test 2", Status: "approved"},
    }

    msg := CommentListFetchedMsg{Comments: comments}
    newM, cmd := m.UpdateComment(msg)

    if cmd != nil {
        t.Error("Expected nil command after fetch, got command")
    }

    model := newM.(Model)

    if model.Loading {
        t.Error("Loading should be false after fetch")
    }

    if len(model.Rows) != 2 {
        t.Errorf("Expected 2 rows, got %d", len(model.Rows))
    }

    if model.CursorMax != 1 {
        t.Errorf("Expected CursorMax 1, got %d", model.CursorMax)
    }
}

func TestErrorHandling(t *testing.T) {
    m := createTestModel()

    msg := CommentErrorMsg{
        Error: fmt.Errorf("test error"),
    }

    newM, _ := m.UpdateComment(msg)
    model := newM.(Model)

    if model.Status != ERROR {
        t.Error("Status should be ERROR after error message")
    }

    if model.Err == nil {
        t.Error("Error should be set")
    }

    if model.Loading {
        t.Error("Loading should be false after error")
    }
}
```

### Testing View Rendering

```go
func TestViewRendering(t *testing.T) {
    m := createTestModel()
    m.Page = NewPage(HOMEPAGE, "Home")
    m.Width = 80
    m.Height = 24

    view := m.View()

    if view == "" {
        t.Error("View returned empty string")
    }

    if !strings.Contains(view, "Home") {
        t.Error("View should contain page title")
    }
}

func TestLoadingSpinner(t *testing.T) {
    m := createTestModel()
    m.Loading = true

    view := m.View()

    if !strings.Contains(view, "Loading") {
        t.Error("Loading view should contain 'Loading' text")
    }
}
```

---

## Integration Testing

### End-to-End Feature Testing

```go
func TestCommentFeatureFlow(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    // Create test user
    user := db.CreateUser(CreateUserParams{
        Username: "testuser",
        Password: "hashedpass",
    })

    // Create test content
    content := db.CreateContentData(CreateContentDataParams{
        RouteID:    1,
        DatatypeID: 1,
        AuthorID:   user.UserID,
    })

    // Test 1: Create comment
    createParams := CreateCommentParams{
        ContentDataID: content.ContentDataID,
        AuthorID:      user.UserID,
        CommentText:   "Test comment",
        Status:        "pending",
    }
    comment := db.CreateComment(createParams)

    if comment.CommentID == 0 {
        t.Fatal("Failed to create comment")
    }

    // Test 2: List comments
    comments, err := db.ListCommentsByContent(content.ContentDataID)
    if err != nil {
        t.Fatalf("Failed to list comments: %v", err)
    }

    if len(comments) != 1 {
        t.Errorf("Expected 1 comment, got %d", len(comments))
    }

    // Test 3: Approve comment
    err = db.ApproveComment(comment.CommentID)
    if err != nil {
        t.Fatalf("Failed to approve comment: %v", err)
    }

    // Test 4: Verify approval
    approved, err := db.GetComment(comment.CommentID)
    if err != nil {
        t.Fatalf("Failed to get comment: %v", err)
    }

    if approved.Status != "approved" {
        t.Errorf("Expected status 'approved', got '%s'", approved.Status)
    }

    // Test 5: Delete comment
    err = db.DeleteComment(comment.CommentID)
    if err != nil {
        t.Fatalf("Failed to delete comment: %v", err)
    }

    // Test 6: Verify deletion
    _, err = db.GetComment(comment.CommentID)
    if err == nil {
        t.Error("Comment should not exist after deletion")
    }
}
```

### Testing Database Migrations

```go
func TestDatabaseMigration(t *testing.T) {
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    // Test that all tables exist
    tables := []string{
        "permissions",
        "roles",
        "users",
        "routes",
        "datatypes",
        "fields",
        "content_data",
        "content_fields",
        "comments", // New table
    }

    for _, table := range tables {
        count, err := db.CountTable(table)
        if err != nil {
            t.Errorf("Table %s does not exist: %v", table, err)
        }

        if count == nil {
            t.Errorf("Failed to count table %s", table)
        }
    }
}
```

---

## Test Coverage

### Measuring Coverage

```bash
# Generate coverage report
make coverage

# View coverage in browser
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Check coverage for specific package
go test -cover ./internal/db
go test -cover ./internal/model
go test -cover ./internal/cli
```

### Coverage Output

```
ok      github.com/hegner123/modulacms/internal/db      2.341s  coverage: 67.8% of statements
ok      github.com/hegner123/modulacms/internal/model   1.892s  coverage: 82.3% of statements
ok      github.com/hegner123/modulacms/internal/cli     3.105s  coverage: 45.2% of statements
```

### Coverage Goals

**Targets:**
- **Critical packages** (db, model): **>80%** coverage
- **Business logic**: **>70%** coverage
- **UI/TUI code**: **>50%** coverage (harder to test)
- **Utilities**: **>90%** coverage

**Priority:**
1. Database operations (CRUD)
2. Business logic (model package)
3. Data transformations
4. Critical paths (authentication, permissions)

### Improving Coverage

**1. Identify uncovered code:**
```bash
go test -coverprofile=coverage.out ./internal/db
go tool cover -func=coverage.out | grep -E '^.*0.0%'
```

**2. Add tests for uncovered functions:**
```go
// If coverage shows this is untested:
func ValidateEmail(email string) bool {
    return strings.Contains(email, "@")
}

// Add test:
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        email string
        valid bool
    }{
        {"user@example.com", true},
        {"invalid", false},
        {"", false},
    }

    for _, tt := range tests {
        result := ValidateEmail(tt.email)
        if result != tt.valid {
            t.Errorf("ValidateEmail(%q) = %v, want %v",
                tt.email, result, tt.valid)
        }
    }
}
```

---

## Benchmarking

### Writing Benchmarks

```go
func BenchmarkTreeCreation(b *testing.B) {
    for i := 0; i < b.N; i++ {
        createMockNode()
    }
}

func BenchmarkTreeTraversal(b *testing.B) {
    tree := createMockNode()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        tree.Walk(func(n *Node) {
            // Do nothing, just traverse
        })
    }
}

func BenchmarkDatabaseInsert(b *testing.B) {
    db := setupTestDB(&testing.T{})

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        params := CreatePermissionParams{
            TableID: 1,
            Mode:    1,
            Label:   fmt.Sprintf("perm_%d", i),
        }
        db.CreatePermission(params)
    }
}
```

### Running Benchmarks

```bash
# Run benchmarks
go test -bench=. ./internal/model

# Run specific benchmark
go test -bench=BenchmarkTreeCreation ./internal/model

# With memory stats
go test -bench=. -benchmem ./internal/model

# Save results
go test -bench=. ./internal/model > old.txt
# ... make changes ...
go test -bench=. ./internal/model > new.txt

# Compare results
benchstat old.txt new.txt
```

---

## Testing Best Practices

### 1. Test Naming

**Good names describe what is being tested:**
```go
// Good
func TestCreatePermissionWithValidParams(t *testing.T) {}
func TestGetPermissionReturnsErrorWhenNotFound(t *testing.T) {}
func TestUpdatePermissionUpdatesAllFields(t *testing.T) {}

// Bad
func TestPermission1(t *testing.T) {}
func TestFunction(t *testing.T) {}
func TestDB(t *testing.T) {}
```

### 2. Test Independence

Each test should be independent:

```go
// Bad - tests depend on order
func TestCreate(t *testing.T) {
    globalUser = db.CreateUser(params)
}

func TestUpdate(t *testing.T) {
    db.UpdateUser(globalUser.UserID, newParams) // Depends on TestCreate
}

// Good - each test is independent
func TestCreate(t *testing.T) {
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    user := db.CreateUser(params)
    // Test creation
}

func TestUpdate(t *testing.T) {
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    user := db.CreateUser(params) // Create own data
    db.UpdateUser(user.UserID, newParams)
    // Test update
}
```

### 3. Use t.Helper()

Mark helper functions:

```go
func setupTestDB(t *testing.T) DbDriver {
    t.Helper() // Marks this as a helper

    // Setup code...
    return db
}

// When test fails, error points to calling line, not inside helper
```

### 4. Clear Error Messages

```go
// Bad
if result != expected {
    t.Error("wrong")
}

// Good
if result != expected {
    t.Errorf("Expected %d, got %d", expected, result)
}

// Even better with context
if result != expected {
    t.Errorf("CreateUser: expected user ID %d, got %d", expected, result)
}
```

### 5. Test Edge Cases

```go
func TestStringToInt64EdgeCases(t *testing.T) {
    tests := []struct {
        name  string
        input string
        want  int64
    }{
        {"empty string", "", 0},
        {"zero", "0", 0},
        {"negative", "-1", -1},
        {"max int64", "9223372036854775807", 9223372036854775807},
        {"overflow", "9223372036854775808", 0},
        {"letters", "abc", 0},
        {"mixed", "123abc", 0},
        {"whitespace", " 123 ", 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := StringToInt64(tt.input)
            if got != tt.want {
                t.Errorf("got %d, want %d", got, tt.want)
            }
        })
    }
}
```

### 6. Use Subtests

```go
func TestUserOperations(t *testing.T) {
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    t.Run("Create", func(t *testing.T) {
        user := db.CreateUser(params)
        // Assert creation
    })

    t.Run("Get", func(t *testing.T) {
        user, err := db.GetUser(1)
        // Assert get
    })

    t.Run("Update", func(t *testing.T) {
        err := db.UpdateUser(1, newParams)
        // Assert update
    })

    t.Run("Delete", func(t *testing.T) {
        err := db.DeleteUser(1)
        // Assert delete
    })
}
```

### 7. Parallel Tests

Run independent tests in parallel:

```go
func TestParallelOperations(t *testing.T) {
    t.Parallel() // Marks test as parallelizable

    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    // Test code...
}
```

**Note:** Only use `t.Parallel()` for tests that don't share state.

---

## Common Testing Patterns

### Pattern 1: Setup-Execute-Assert

```go
func TestSomething(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    user := createTestUser(db)

    // Execute
    result, err := db.DoSomething(user.ID)

    // Assert
    if err != nil {
        t.Fatalf("Unexpected error: %v", err)
    }
    if result == nil {
        t.Error("Expected result, got nil")
    }
}
```

### Pattern 2: Table-Driven Tests with Setup

```go
func TestMultipleScenarios(t *testing.T) {
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    tests := []struct {
        name  string
        setup func()
        input string
        want  string
    }{
        {
            name: "scenario 1",
            setup: func() {
                db.CreateUser(params1)
            },
            input: "test1",
            want:  "result1",
        },
        // More cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tt.setup()
            got := doSomething(tt.input)
            if got != tt.want {
                t.Errorf("got %s, want %s", got, tt.want)
            }
        })
    }
}
```

### Pattern 3: Testing Errors

```go
func TestErrorCases(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
        errMsg  string
    }{
        {
            name:    "empty input",
            input:   "",
            wantErr: true,
            errMsg:  "input cannot be empty",
        },
        {
            name:    "valid input",
            input:   "valid",
            wantErr: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := doSomething(tt.input)

            if tt.wantErr {
                if err == nil {
                    t.Error("Expected error, got nil")
                    return
                }
                if tt.errMsg != "" && err.Error() != tt.errMsg {
                    t.Errorf("Expected error %q, got %q", tt.errMsg, err.Error())
                }
            } else {
                if err != nil {
                    t.Errorf("Unexpected error: %v", err)
                }
            }
        })
    }
}
```

---

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Tests

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v3

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.23'

    - name: Create test directories
      run: |
        mkdir -p testdb
        mkdir -p backups

    - name: Run tests
      run: make test

    - name: Run coverage
      run: make coverage

    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        files: ./profile.cov
```

### Pre-commit Hook

**File:** `.git/hooks/pre-commit`

```bash
#!/usr/bin/env bash

# Run tests before commit
echo "Running tests..."
make test

if [ $? -ne 0 ]; then
    echo "Tests failed. Commit aborted."
    exit 1
fi

echo "Tests passed."
exit 0
```

Make executable:
```bash
chmod +x .git/hooks/pre-commit
```

---

## Troubleshooting Tests

### Common Issues

**Issue 1: Database locked**
```
Error: database is locked
```

**Solution:** Close previous connections:
```go
defer cleanupTestDB(t, db) // Always defer cleanup
```

**Issue 2: Test files not found**
```
Error: cannot find testdb/test.db
```

**Solution:** Create directories:
```bash
mkdir -p testdb backups
```

**Issue 3: Tests fail in CI but pass locally**
```
Error: connection refused
```

**Solution:** Use relative paths and ensure directories exist:
```go
c.Db_URL = "./testdb/test.db" // Relative path
```

**Issue 4: Parallel tests interfere**
```
Error: foreign key constraint failed
```

**Solution:** Don't use `t.Parallel()` for database tests, or use separate test databases:
```go
func TestParallel(t *testing.T) {
    // Don't use t.Parallel() if sharing database
    db := setupTestDB(t)
    // ...
}
```

---

## Checklist for Writing Tests

Use this checklist when adding tests:

- [ ] Test file named `*_test.go`
- [ ] Import `testing` package
- [ ] Test functions start with `Test`
- [ ] Use table-driven tests for multiple cases
- [ ] Include edge cases (empty, nil, max values)
- [ ] Test error conditions
- [ ] Use descriptive test names
- [ ] Add comments explaining complex test logic
- [ ] Clean up resources (defer cleanup)
- [ ] Use `t.Helper()` for helper functions
- [ ] Clear error messages with context
- [ ] Independent tests (no shared state)
- [ ] Run tests locally before committing
- [ ] Check test coverage
- [ ] Add benchmarks for performance-critical code
- [ ] Update CI/CD if needed

---

## Related Documentation

**Essential Reading:**
- [ADDING_TABLES.md](ADDING_TABLES.md) - Includes database test examples
- [CREATING_TUI_SCREENS.md](CREATING_TUI_SCREENS.md) - Includes TUI test examples
- [DB_PACKAGE.md](../DB_PACKAGE.md) - Database abstraction layer
- [CLI_PACKAGE.md](../packages/CLI_PACKAGE.md) - TUI implementation

**Related Workflows:**
- [DEBUGGING.md](DEBUGGING.md) - Debugging strategies (once available)
- [ADDING_FEATURES.md](ADDING_FEATURES.md) - Step 9: Write tests

**External Resources:**
- [Go Testing Package](https://pkg.go.dev/testing)
- [Table-Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [testify/assert](https://github.com/stretchr/testify) - Assertion library (optional)

---

## Quick Reference

### Test Commands

```bash
# Run all tests
make test

# Run with verbose output
go test -v ./...

# Run specific test
go test -run TestName ./path/to/package

# Run with coverage
make coverage

# Run benchmarks
go test -bench=. ./...

# Run with race detector
go test -race ./...
```

### Basic Test Structure

```go
func TestSomething(t *testing.T) {
    // Setup
    setup()
    defer cleanup()

    // Execute
    result := doSomething()

    // Assert
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

### Table-Driven Test Template

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   Type
        want    Type
        wantErr bool
    }{
        {"case 1", input1, want1, false},
        {"case 2", input2, want2, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Function(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

---

**Last Updated:** 2026-01-12
**Status:** Complete
**Part of:** Phase 2 High Priority Documentation

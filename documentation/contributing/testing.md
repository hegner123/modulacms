# Testing

ModulaCMS uses Go's built-in testing framework with specific patterns for database testing, TUI component testing, and integration testing. Tests run against SQLite in-memory or file-based databases in the `testdb/` directory.

## Running Tests

```bash
# Run all tests
just test

# Run with verbose output
go test -v ./...

# Run a specific package
go test -v ./internal/db

# Run a specific test by name
go test -v ./internal/db -run TestPermissionCRUD

# Run tests matching a pattern
go test -v ./internal/db -run TestCreate

# Run with race detector
go test -race ./...

# Run with coverage report
just coverage

# Run S3 integration tests (requires MinIO)
just test-minio
just test-integration
just test-minio-down
```

The `just test` target creates and cleans up test databases in `testdb/` and backup files in `backups/` automatically.

## Test Organization

Test files live alongside the code they test, following Go convention:

```
internal/db/
  db.go
  db_test.go
  permission.go
  permission_test.go
internal/model/
  model.go
  model_test.go
  build_test.go
```

Tests in the same package (e.g., `package db`) can access unexported functions for white-box testing. Tests in a separate `_test` package (e.g., `package db_test`) test only the public API.

ModulaCMS primarily uses same-package tests to access internal helpers and mapping functions.

## Database Testing

### Test Database Setup

Database tests use a helper that creates a SQLite database, initializes all tables, and returns a `DbDriver`:

```go
func setupTestDB(t *testing.T) DbDriver {
    t.Helper()

    p := config.NewFileProvider("")
    m := config.NewManager(p)
    c, err := m.Config()
    if err != nil {
        t.Fatalf("Failed to load config: %v", err)
    }

    c.Db_Driver = "sqlite"
    c.Db_URL = "./testdb/test.db"

    d := ConfigDB(*c)

    err = d.CreateAllTables()
    if err != nil {
        t.Fatalf("Failed to create tables: %v", err)
    }

    return d
}

func cleanupTestDB(t *testing.T, d DbDriver) {
    t.Helper()
    con, _, err := d.GetConnection()
    if err != nil {
        t.Logf("Warning: cleanup failed: %v", err)
        return
    }
    if err := con.Close(); err != nil {
        t.Logf("Warning: close failed: %v", err)
    }
}
```

Always defer cleanup to prevent "database is locked" errors from unclosed connections:

```go
func TestSomething(t *testing.T) {
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    // test code
}
```

### CRUD Test Pattern

```go
func TestPermissionCRUD(t *testing.T) {
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    // Create
    permission := db.CreatePermission(CreatePermissionParams{
        TableID: 1,
        Mode:    1,
        Label:   "test_permission",
    })
    if permission.PermissionID == 0 {
        t.Fatal("CreatePermission failed: permission_id is 0")
    }

    // Read
    fetched, err := db.GetPermission(permission.PermissionID)
    if err != nil {
        t.Fatalf("GetPermission failed: %v", err)
    }
    if fetched.Label != "test_permission" {
        t.Errorf("expected label 'test_permission', got '%s'", fetched.Label)
    }

    // Update
    err = db.UpdatePermission(UpdatePermissionParams{
        TableID:      2,
        Mode:         2,
        Label:        "updated_label",
        PermissionID: permission.PermissionID,
    })
    if err != nil {
        t.Fatalf("UpdatePermission failed: %v", err)
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

## Table-Driven Tests

Use table-driven tests for functions with varied inputs:

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
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := StringToInt64(tt.input)
            if tt.wantErr {
                if result != 0 {
                    t.Errorf("expected 0 for invalid input, got %d", result)
                }
                return
            }
            if result != tt.expected {
                t.Errorf("expected %d, got %d", tt.expected, result)
            }
        })
    }
}
```

## TUI Testing

Test Bubbletea components by creating a model, sending messages, and asserting the resulting state:

```go
func TestNavigationUpdate(t *testing.T) {
    m := createTestModel()

    msg := NavigateToPage{
        Page: NewPage(DATABASEPAGE, "Database"),
    }

    newM, cmd := m.Update(msg)
    if cmd == nil {
        t.Error("Expected command from navigation, got nil")
    }

    model := newM.(Model)
    if model.Page.Index != DATABASEPAGE {
        t.Errorf("expected page index %d, got %d", DATABASEPAGE, model.Page.Index)
    }
}

func TestCursorBounds(t *testing.T) {
    m := createTestModel()
    m.CursorMax = 5
    m.Cursor = 5

    // Attempt to move below maximum
    msg := tea.KeyMsg{Type: tea.KeyDown}
    newM, _ := m.Update(msg)
    model := newM.(Model)

    if model.Cursor != 5 {
        t.Errorf("cursor exceeded max: expected 5, got %d", model.Cursor)
    }
}
```

Test view rendering by checking that the output contains expected content:

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
}
```

## Integration Testing

Test complete feature flows that span multiple database operations:

```go
func TestFeatureFlow(t *testing.T) {
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)

    // Create prerequisites
    user := db.CreateUser(CreateUserParams{
        Username: "testuser",
        Password: "hashedpass",
    })

    content := db.CreateContentData(CreateContentDataParams{
        RouteID:    1,
        DatatypeID: 1,
        AuthorID:   user.UserID,
    })

    // Test the feature
    comment := db.CreateComment(CreateCommentParams{
        ContentDataID: content.ContentDataID,
        AuthorID:      user.UserID,
        CommentText:   "Test comment",
        Status:        "pending",
    })
    if comment.CommentID == 0 {
        t.Fatal("Failed to create comment")
    }

    // Verify the full flow
    comments, err := db.ListCommentsByContent(content.ContentDataID)
    if err != nil {
        t.Fatalf("Failed to list comments: %v", err)
    }
    if len(comments) != 1 {
        t.Errorf("expected 1 comment, got %d", len(comments))
    }
}
```

## Test Coverage

```bash
# Generate coverage report
just coverage

# View coverage in browser
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Check coverage for a specific package
go test -cover ./internal/db
```

### Coverage Targets

| Package | Target |
|---------|--------|
| Database operations (`internal/db`) | 80%+ |
| Business logic (`internal/model`) | 70%+ |
| Utilities | 90%+ |
| TUI code (`internal/tui`) | 50%+ |

## Benchmarking

Write benchmarks for performance-critical operations:

```go
func BenchmarkTreeCreation(b *testing.B) {
    for i := 0; i < b.N; i++ {
        createMockNode()
    }
}
```

Run benchmarks:

```bash
go test -bench=. ./internal/model
go test -bench=. -benchmem ./internal/model
```

## Best Practices

**Test naming.** Use descriptive names that state what is being tested: `TestCreatePermissionWithValidParams`, not `TestPermission1`.

**Test independence.** Each test should create its own data. Never depend on test execution order or shared mutable state.

**Use `t.Helper()`.** Mark setup and assertion helper functions with `t.Helper()` so failure messages point to the calling test, not the helper.

**Clear error messages.** Include expected and actual values: `t.Errorf("expected %d, got %d", expected, result)`.

**Test edge cases.** Include empty strings, nil values, zero values, maximum values, and invalid inputs in your test cases.

**Clean up resources.** Always `defer` database cleanup. Unclosed connections cause "database is locked" errors in subsequent tests.

**Avoid `t.Parallel()` for database tests.** Tests sharing a database file can interfere with each other. Use parallel execution only for tests with no shared state.

## Troubleshooting Tests

**"database is locked"** -- A previous test left a connection open. Add `defer cleanupTestDB(t, db)` to every test that creates a database connection.

**"cannot find testdb/test.db"** -- The test directories do not exist. Run `mkdir -p testdb backups` or use `just test` which creates them automatically.

**Tests pass locally but fail in CI** -- Check for absolute paths in test setup. Use relative paths like `./testdb/test.db` and ensure the CI workflow creates the required directories.

**"foreign key constraint failed"** -- The test is referencing a parent record that does not exist. Create all prerequisite records (routes, datatypes, users) before creating dependent records.

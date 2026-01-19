# Package Testing Plan for ModulaCMS

**Purpose:** Strategy for testing individual packages with dependencies
**Created:** 2026-01-19

---

## Dependency Layers (Test Bottom-Up)

```
Layer 0: utility, config          ← No internal deps, test first
Layer 1: db, bucket              ← Depend on config/utility only
Layer 2: auth, model             ← Depend on db, config
Layer 3: media, middleware       ← Depend on bucket, auth, db
Layer 4: router, cli             ← Depend on everything
```

### Dependency Graph

```
utility (no internal deps)
    ↑
    ├── config ─┐
    │           ├── auth
    │           ├── bucket
    │           ├── backup
    │           ├── cli
    │           ├── db
    │           ├── install
    │           ├── media
    │           ├── middleware
    │           ├── model
    │           ├── router
    │           └── transform
    │
    └── db ─┐
            ├── auth
            ├── backup
            ├── cli
            ├── install
            ├── media
            ├── middleware
            ├── model
            ├── router
            └── transform

auth ─┐
      ├── middleware
      └── router

bucket ─┐
        ├── install
        └── media

media ─┐
       ├── bucket
       └── router

model ─┐
       ├── cli
       ├── router
       └── transform

cli ──→ router (via SSH/TUI)
```

**No circular dependencies detected** - clean layering enables parallel testing.

---

## Testing Strategy by Package

### Layer 0: Foundation (Direct Testing)

#### internal/utility
No mocks needed - pure functions.

```bash
go test ./internal/utility/...
```

#### internal/config
Mock Provider interface or use temp files.

```go
// Test with file provider
func TestConfigLoad(t *testing.T) {
    tmpFile := createTempConfig(t, testConfigJSON)
    defer os.Remove(tmpFile)

    provider := config.NewFileProvider(tmpFile)
    cfg, err := provider.Load()
    // assertions
}
```

---

### Layer 1: Data Layer

#### internal/db
Use in-memory SQLite for integration tests.

```bash
# Run with test database
DB_DRIVER=sqlite DB_URL=":memory:" go test ./internal/db/...
```

```go
// db_test.go - use test helper
func setupTestDB(t *testing.T) db.DbDriver {
    driver := db.NewSQLiteDriver(":memory:")
    if err := driver.CreateAllTables(); err != nil {
        t.Fatal(err)
    }
    return driver
}
```

#### internal/bucket
Mock AWS SDK or use localstack.

```go
// bucket_test.go
type mockS3Client struct {
    uploadFunc func(*s3.PutObjectInput) (*s3.PutObjectOutput, error)
}

func TestObjectUpload(t *testing.T) {
    // Use localstack or mock S3 client
    mockClient := &mockS3Client{
        uploadFunc: func(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
            return &s3.PutObjectOutput{}, nil
        },
    }
    // test with mock
}
```

---

### Layer 2: Service Layer

#### internal/auth
Mock db.DbDriver.

```go
// auth_test.go
type mockDbDriver struct {
    db.DbDriver // embed interface for partial implementation
    getUserByEmail func(string) (*db.User, error)
}

func (m *mockDbDriver) GetUserByEmail(email string) (*db.User, error) {
    return m.getUserByEmail(email)
}

func TestAuthenticate(t *testing.T) {
    mock := &mockDbDriver{
        getUserByEmail: func(email string) (*db.User, error) {
            return &db.User{Email: email, PasswordHash: "..."}, nil
        },
    }
    // test auth with mock db
}
```

#### internal/model
Test JSON serialization - pure struct tests, minimal mocking.

```go
// model_test.go
func TestNodeMarshal(t *testing.T) {
    node := model.Node{
        Datatype: model.Datatype{...},
        Fields:   []model.Field{...},
    }
    data, err := json.Marshal(node)
    // assert expected JSON structure
}
```

---

### Layer 3: Feature Layer

#### internal/media
Mock bucket, config, db.

```go
// media_test.go
type testDeps struct {
    db     *mockDbDriver
    bucket *mockBucket
    config config.Config
}

func setupMediaTest(t *testing.T) *testDeps {
    return &testDeps{
        db:     &mockDbDriver{...},
        bucket: &mockBucket{...},
        config: config.Config{
            Bucket_Media:       "test-bucket",
            Bucket_Default_ACL: "private",
        },
    }
}

func TestMediaUpload(t *testing.T) {
    deps := setupMediaTest(t)
    // test upload flow with mocks
}
```

#### internal/middleware
Mock auth, config.

```go
// middleware_test.go
func TestCORSMiddleware(t *testing.T) {
    cfg := config.Config{
        Cors_Origins: []string{"https://example.com"},
    }

    handler := middleware.CORS(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))

    req := httptest.NewRequest("OPTIONS", "/", nil)
    req.Header.Set("Origin", "https://example.com")
    rec := httptest.NewRecorder()

    handler.ServeHTTP(rec, req)
    // assert CORS headers
}
```

---

### Layer 4: Application Layer

#### internal/router
Full integration with httptest.

```go
// router_test.go
func setupTestRouter(t *testing.T) (*http.ServeMux, *testDeps) {
    deps := &testDeps{
        db:     setupTestDB(t),  // real SQLite in-memory
        config: testConfig(),
    }

    mux := router.NewModulacmsMux(deps.config)
    return mux, deps
}

func TestMediaEndpoint(t *testing.T) {
    mux, _ := setupTestRouter(t)

    req := httptest.NewRequest("GET", "/api/media", nil)
    rec := httptest.NewRecorder()

    mux.ServeHTTP(rec, req)

    assert.Equal(t, http.StatusOK, rec.Code)
}
```

#### internal/cli
Test Bubbletea models in isolation.

```go
// cli_test.go
func TestTreeModel(t *testing.T) {
    mockDB := &mockDbDriver{
        getContentData: func() ([]db.ContentData, error) {
            return []db.ContentData{{ID: 1, Title: "Test"}}, nil
        },
    }

    model := cli.NewTreeModel(mockDB)
    updated, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

    // assert state changes
}
```

---

## Test Execution Commands

### Single Package Tests

```bash
go test -v ./internal/utility/...
go test -v ./internal/config/...
go test -v ./internal/db/...
go test -v ./internal/bucket/...
go test -v ./internal/media/...
go test -v ./internal/router/...
go test -v ./internal/cli/...
```

### Coverage for Specific Package

```bash
go test -v -coverprofile=coverage.out ./internal/media/...
go tool cover -html=coverage.out
```

### Layer-by-Layer Testing

```bash
# Layer 0
go test ./internal/utility/... ./internal/config/...

# Layer 1
go test ./internal/db/... ./internal/bucket/...

# Layer 2
go test ./internal/auth/... ./internal/model/...

# Layer 3
go test ./internal/media/... ./internal/middleware/...

# Layer 4
go test ./internal/router/... ./internal/cli/...
```

### Race Detection

```bash
go test -race ./internal/...
```

---

## Mock Infrastructure

### Recommended Structure

```
internal/testutil/
├── mock_db.go        # MockDbDriver with configurable methods
├── mock_bucket.go    # MockS3Client for bucket operations
├── mock_config.go    # TestConfig() helper function
├── fixtures/         # JSON test fixtures
│   ├── content.json
│   ├── media.json
│   └── user.json
└── helpers.go        # Common test utilities
```

### MockDbDriver Implementation

```go
// internal/testutil/mock_db.go
package testutil

import "modulacms/internal/db"

type MockDbDriver struct {
    // Function fields for configurable behavior
    CreateContentDataFunc func(db.CreateContentDataParams) (db.ContentData, error)
    GetContentDataFunc    func(int64) (*db.ContentData, error)
    ListMediaFunc         func() ([]db.Media, error)
    // ... add as needed
}

func (m *MockDbDriver) CreateContentData(p db.CreateContentDataParams) (db.ContentData, error) {
    if m.CreateContentDataFunc != nil {
        return m.CreateContentDataFunc(p)
    }
    return db.ContentData{}, nil
}

func (m *MockDbDriver) GetContentData(id int64) (*db.ContentData, error) {
    if m.GetContentDataFunc != nil {
        return m.GetContentDataFunc(id)
    }
    return nil, nil
}

// Default mock that returns empty/success for all methods
func NewMockDbDriver() *MockDbDriver {
    return &MockDbDriver{}
}
```

---

## Makefile Targets

```makefile
test-utility:
	go test -v ./internal/utility/...

test-config:
	go test -v ./internal/config/...

test-db:
	DB_DRIVER=sqlite DB_URL=":memory:" go test -v ./internal/db/...

test-bucket:
	go test -v ./internal/bucket/...

test-media:
	go test -v ./internal/media/...

test-router:
	go test -v ./internal/router/...

test-cli:
	go test -v ./internal/cli/...

test-layer0:
	go test -v ./internal/utility/... ./internal/config/...

test-layer1:
	go test -v ./internal/db/... ./internal/bucket/...

test-layer2:
	go test -v ./internal/auth/... ./internal/model/...

test-layer3:
	go test -v ./internal/media/... ./internal/middleware/...

test-layer4:
	go test -v ./internal/router/... ./internal/cli/...

test-all-layers: test-layer0 test-layer1 test-layer2 test-layer3 test-layer4
```

---

## Summary Table

| Package | Layer | Mock Strategy | Run Command |
|---------|-------|--------------|-------------|
| utility | 0 | None needed | `go test ./internal/utility/...` |
| config | 0 | Temp files or mock Provider | `go test ./internal/config/...` |
| db | 1 | In-memory SQLite | `go test ./internal/db/...` |
| bucket | 1 | Mock S3 client | `go test ./internal/bucket/...` |
| auth | 2 | Mock DbDriver | `go test ./internal/auth/...` |
| model | 2 | None (pure structs) | `go test ./internal/model/...` |
| media | 3 | Mock bucket + db + config | `go test ./internal/media/...` |
| middleware | 3 | Mock auth + config | `go test ./internal/middleware/...` |
| router | 4 | httptest + mock all deps | `go test ./internal/router/...` |
| cli | 4 | Mock DbDriver + tea.Model | `go test ./internal/cli/...` |

---

## Key Mockable Interfaces

1. **db.DbDriver** (100+ methods) - Most critical for all consumers
2. **config.Provider** - For configuration loading
3. **AWS SDK types** - For S3 operations
4. **OAuth2 Provider** - For auth flows
5. **Bubbletea Model** - For CLI testing

---

## Next Steps

1. Create `internal/testutil/` package with mock infrastructure
2. Restore tests from `.trash/` directory
3. Update tests to use new mock patterns
4. Add Makefile targets for layer-by-layer testing
5. Implement CI pipeline running tests in dependency order

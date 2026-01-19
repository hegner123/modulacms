# Test Organization Strategy for ModulaCMS

**Date:** 2026-01-15
**Status:** Recommendation Guide
**Context:** How to structure test files to reduce directory clutter while working with Go conventions

---

## Table of Contents

1. [Current State](#current-state)
2. [Go Testing Conventions](#go-testing-conventions)
3. [Organization Options](#organization-options)
4. [Recommended Strategy](#recommended-strategy)
5. [Implementation Examples](#implementation-examples)
6. [Migration Plan](#migration-plan)

---

## Current State

### Existing Test Files (14 found)

**Packages with tests:**
```
internal/middleware/tests/cors_test.go          ← Subdirectory (black-box)
internal/install/install_form_test.go           ← Same directory
internal/config/config_test.go                  ← Same directory
internal/bucket/object_storage_test.go          ← Same directory
internal/model/model_test.go                    ← Same directory
internal/model/build_test.go                    ← Same directory
internal/db/foreignKey_test.go                  ← Same directory
internal/backup/backup_test.go                  ← Same directory
internal/utility/log_test.go                    ← Same directory
internal/utility/timestamp_test.go              ← Same directory
internal/utility/url_test.go                    ← Same directory
internal/media/media_optomize_test.go           ← Same directory
internal/media/media_upload_test.go             ← Same directory
internal/router/adminContentData_test.go        ← Same directory
```

**Packages WITHOUT tests:**
- ❌ `internal/cli/` - 44 files, ZERO tests
- ❌ `internal/db/` - 36 entity files, only 1 test (foreignKey)
- ❌ Most other packages

**Test fixture directories:**
```
/testFiles/   - Test data files
/testdb/      - Test databases
internal/bucket/testfiles/ - Bucket test data
```

### Problem Statement

**User concern:** "listing under each file makes a messy directory hard to parse"

**Reality for ModulaCMS:**
- CLI package: 44 files, 0 tests → No clutter problem yet
- DB package: 36 files, 1 test → Minimal clutter
- **The clutter problem is FUTURE concern** as test coverage grows

**This is a GOOD problem to anticipate!**

---

## Go Testing Conventions

### Standard Approach (Most Common)

**Pattern:** `filename.go` + `filename_test.go` side-by-side

```
internal/db/
├── content_data.go
├── content_data_test.go
├── user.go
├── user_test.go
└── route.go
    route_test.go
```

**Pros:**
- ✅ Idiomatic Go - everyone understands this
- ✅ `go test ./...` discovers tests automatically
- ✅ Tests live next to code they test (easy to find)
- ✅ Can test unexported functions (white-box testing)
- ✅ Works with all Go tooling (coverage, benchmarks, examples)

**Cons:**
- ❌ Directory listing shows both code and tests (visual clutter)
- ❌ Large packages get long file lists
- ❌ Can't easily skip tests when browsing

### Black-Box Testing (package_test)

**Pattern:** Use `package_test` suffix in test files

```go
// content_data.go
package db

// content_data_test.go
package db_test  // ← Different package!

import "github.com/hegner123/modulacms/internal/db"
```

**Pros:**
- ✅ Tests use public API only (better design)
- ✅ Catches exported interface issues
- ✅ Still side-by-side with code
- ✅ Standard Go practice

**Cons:**
- ❌ Can't test unexported functions
- ❌ Still creates visual clutter in directory

**When to use:** Integration tests, API tests, package boundary tests

---

## Organization Options

### Option 1: Standard Go (Keep As-Is)

**Structure:**
```
internal/cli/
├── commands.go
├── commands_test.go
├── forms.go
├── forms_test.go
├── model.go
├── model_test.go
└── ... (44 files → 88 files with tests)
```

**Verdict:** ✅ **Idiomatic, but creates the clutter you want to avoid**

---

### Option 2: Subdirectory Per Test Type

**Structure:**
```
internal/cli/
├── commands.go
├── forms.go
├── model.go
├── ... (44 production files)
├── unit/
│   ├── commands_test.go
│   ├── forms_test.go
│   └── model_test.go
├── integration/
│   ├── cms_integration_test.go
│   └── form_flow_test.go
└── benchmark/
    └── parse_bench_test.go
```

**Test package names:**
```go
// unit/commands_test.go
package cli_test  // Black-box testing

// integration/cms_integration_test.go
package cli_test

// benchmark/parse_bench_test.go
package cli_test
```

**Pros:**
- ✅ Separates tests from production code (clean directory)
- ✅ Groups tests by type (unit, integration, benchmark)
- ✅ Easy to navigate production code without test clutter
- ✅ Can have different test configurations per directory

**Cons:**
- ❌ Less idiomatic Go (but still valid)
- ❌ Tests farther from code (need to remember structure)
- ❌ Must use black-box testing (package_test)
- ❌ Slightly more verbose imports

**Verdict:** ✅ **Good compromise - reduces clutter, maintains organization**

---

### Option 3: Nested Packages (Domain-Based)

**Structure:**
```
internal/cli/
├── commands/
│   ├── commands.go
│   ├── commands_test.go
│   ├── database.go
│   └── database_test.go
├── forms/
│   ├── forms.go
│   ├── forms_test.go
│   ├── builders.go
│   └── builders_test.go
├── rendering/
│   ├── view.go
│   ├── view_test.go
│   └── styles.go
│       styles_test.go
└── model.go
    model_test.go
```

**Package names:**
```go
// commands/commands.go
package commands

// forms/forms.go
package forms
```

**Pros:**
- ✅ Splits large packages into logical domains
- ✅ Each subdomain has manageable file count
- ✅ Standard Go convention (nested packages)
- ✅ Can control visibility with internal/exported types

**Cons:**
- ❌ Requires significant refactoring
- ❌ Changes import paths everywhere
- ❌ May over-fragment code (too many small packages)
- ❌ Breaks existing architecture

**Verdict:** ❌ **Too invasive - better for new projects or major refactors**

---

### Option 4: Hybrid Approach (Production + Tests Subdirectory)

**Structure:**
```
internal/cli/
├── commands.go
├── forms.go
├── model.go
├── ... (44 production files - UNCHANGED)
└── tests/
    ├── commands_test.go
    ├── forms_test.go
    ├── model_test.go
    ├── integration_test.go
    └── benchmark_test.go
```

**Test package:**
```go
// tests/commands_test.go
package cli_test  // Black-box testing
```

**Pros:**
- ✅ All tests in one place (easy to find)
- ✅ Production code directory stays clean
- ✅ Minimal refactoring needed
- ✅ Works with all Go tooling

**Cons:**
- ❌ All tests in single directory (could get messy)
- ❌ Loses co-location benefit
- ❌ Less idiomatic (but still valid)

**Verdict:** ✅ **Simple solution - good for transitioning**

---

### Option 5: Build Tags (Hide Tests from Listings)

**Structure:**
```
internal/cli/
├── commands.go
├── commands_test.go          ← Hidden with build tags
├── forms.go
├── forms_test.go             ← Hidden with build tags
└── model.go
    model_test.go             ← Hidden with build tags
```

**Test files:**
```go
// +build integration

package cli

import "testing"
```

**Editor config (.ignore files):**
```
# .gitignore (for editor file trees)
*_test.go
```

**Pros:**
- ✅ Standard Go approach
- ✅ Tests still co-located
- ✅ Can use editor filtering to hide tests
- ✅ Build tags control which tests run

**Cons:**
- ❌ Tests still physically present (not solving clutter)
- ❌ Relies on editor configuration
- ❌ Different devs see different structure

**Verdict:** ⚠️ **Band-aid - doesn't actually solve problem**

---

## Recommended Strategy

### For ModulaCMS: **Hybrid Subdirectory Approach (Option 2)**

**Why this works:**

1. **Matches existing pattern**: You already have `/internal/middleware/tests/` subdirectory
2. **Reduces clutter**: Production code directory stays clean
3. **Organizes by test type**: Unit, integration, benchmark in separate subdirs
4. **Maintains flexibility**: Can use white-box OR black-box testing as needed
5. **Go-compatible**: Works with all standard tooling

### Structure

```
internal/cli/
├── commands.go
├── forms.go
├── model.go
├── ... (44 production files)
└── tests/
    ├── unit/
    │   ├── commands_test.go
    │   ├── forms_test.go
    │   └── model_test.go
    ├── integration/
    │   ├── cms_flow_test.go
    │   └── database_ops_test.go
    └── helpers/
        └── test_utils.go

internal/db/
├── content_data.go
├── user.go
├── ... (36 entity files)
└── tests/
    ├── unit/
    │   ├── content_data_test.go
    │   └── user_test.go
    ├── integration/
    │   ├── tree_operations_test.go
    │   └── multi_db_test.go
    └── helpers/
        └── db_fixtures.go
```

### Test Package Convention

**Unit tests - Black-box (preferred):**
```go
// internal/cli/tests/unit/commands_test.go
package cli_test

import (
    "testing"
    "github.com/hegner123/modulacms/internal/cli"
)

func TestDatabaseInsert(t *testing.T) {
    // Test public API only
    m := cli.Model{}
    // ...
}
```

**Unit tests - White-box (when needed):**
```go
// internal/cli/tests/unit/parse_internals_test.go
package cli  // Same package - can test unexported

import "testing"

func TestInternalParseLogic(t *testing.T) {
    // Test unexported functions
    result := parseInternal("test")
    // ...
}
```

**Integration tests:**
```go
// internal/cli/tests/integration/cms_flow_test.go
package cli_test

import (
    "testing"
    "github.com/hegner123/modulacms/internal/cli"
    "github.com/hegner123/modulacms/internal/db"
)

func TestCMSContentCreationFlow(t *testing.T) {
    // Test multiple packages working together
}
```

---

## Implementation Examples

### Example 1: CLI Package Tests

**Directory structure:**
```
internal/cli/tests/
├── unit/
│   ├── commands_test.go       - Database command tests
│   ├── forms_test.go          - Form builder tests
│   ├── model_test.go          - Model state tests
│   ├── parse_test.go          - Parser tests
│   └── validation_test.go     - Validation tests
├── integration/
│   ├── cms_flow_test.go       - End-to-end CMS operations
│   ├── form_submission_test.go- Form → DB → UI flow
│   └── navigation_test.go     - Page navigation flow
├── benchmark/
│   ├── parse_bench_test.go    - Parser performance
│   └── render_bench_test.go   - View rendering performance
└── helpers/
    ├── fixtures.go            - Test data builders
    ├── mocks.go               - Mock implementations
    └── test_utils.go          - Helper functions
```

**Test file example:**
```go
// internal/cli/tests/unit/commands_test.go
package cli_test

import (
    "testing"

    "github.com/hegner123/modulacms/internal/cli"
    "github.com/hegner123/modulacms/internal/config"
)

func TestDatabaseInsert(t *testing.T) {
    // Setup
    cfg := config.Config{
        Database: config.DatabaseConfig{
            Driver: "sqlite",
            Src:    ":memory:",
        },
    }

    model := cli.Model{
        Config: cfg,
    }

    // Test
    cmd := model.DatabaseInsert("users", map[string]any{
        "username": "test",
        "email":    "test@example.com",
    })

    // Assert
    if cmd == nil {
        t.Fatal("Expected command, got nil")
    }
}
```

---

### Example 2: DB Package Tests

**Directory structure:**
```
internal/db/tests/
├── unit/
│   ├── content_data_test.go    - ContentData CRUD
│   ├── user_test.go            - User operations
│   ├── tree_test.go            - Tree structure
│   └── query_builder_test.go  - SecureQueryBuilder
├── integration/
│   ├── multi_db_test.go        - SQLite vs MySQL vs PostgreSQL
│   ├── transaction_test.go     - Transaction handling
│   └── tree_operations_test.go - Complex tree ops
├── benchmark/
│   ├── tree_traversal_bench.go - Tree performance
│   └── query_bench.go          - Query performance
└── helpers/
    ├── db_fixtures.go          - Test database setup
    ├── seed_data.go            - Seed test data
    └── assertions.go           - Custom test assertions
```

**Test file example:**
```go
// internal/db/tests/unit/content_data_test.go
package db_test

import (
    "testing"

    "github.com/hegner123/modulacms/internal/db"
    "github.com/hegner123/modulacms/internal/db/tests/helpers"
)

func TestCreateContentData(t *testing.T) {
    // Setup test database
    d, cleanup := helpers.SetupTestDB(t, "sqlite")
    defer cleanup()

    // Create content
    params := db.CreateContentDataParams{
        DatatypeID: helpers.Int64Ptr(1),
        RouteID:    helpers.Int64Ptr(1),
        // ...
    }

    result := d.CreateContentData(params)

    // Assert
    if result.ContentDataID == 0 {
        t.Error("Expected valid ID, got 0")
    }
}
```

---

## Migration Plan

### Phase 1: Add Test Infrastructure (1-2 hours)

**Create test directory structure:**

```bash
# CLI package
mkdir -p internal/cli/tests/{unit,integration,benchmark,helpers}

# DB package
mkdir -p internal/db/tests/{unit,integration,benchmark,helpers}

# Other large packages
mkdir -p internal/middleware/tests/{unit,integration}  # Already has tests/
mkdir -p internal/model/tests/{unit,integration}
```

**Create helper files:**

```bash
# CLI helpers
touch internal/cli/tests/helpers/{fixtures.go,mocks.go,test_utils.go}

# DB helpers
touch internal/db/tests/helpers/{db_fixtures.go,seed_data.go,assertions.go}
```

**Create README in each tests/ directory:**

```markdown
# Tests Directory

Tests for the [package-name] package.

## Structure

- `unit/` - Unit tests for individual functions/methods
- `integration/` - Integration tests across multiple components
- `benchmark/` - Performance benchmarks
- `helpers/` - Test utilities, fixtures, and mocks

## Running Tests

```bash
# All tests
go test ./...

# Unit tests only
go test ./unit/...

# Integration tests only
go test ./integration/...

# Benchmarks
go test -bench=. ./benchmark/...
```

## Writing Tests

- Use `package_test` suffix for black-box tests
- Use same package name for white-box tests (testing unexported)
- Place test helpers in `helpers/` subdirectory
```

---

### Phase 2: Move Existing Tests (1 hour)

**Current test locations:**
```
internal/db/foreignKey_test.go
internal/model/model_test.go
internal/model/build_test.go
internal/utility/log_test.go
internal/utility/timestamp_test.go
internal/utility/url_test.go
```

**Move to subdirectories:**

```bash
# Move DB test
mv internal/db/foreignKey_test.go internal/db/tests/unit/

# Move model tests
mv internal/model/model_test.go internal/model/tests/unit/
mv internal/model/build_test.go internal/model/tests/unit/

# Move utility tests
mkdir -p internal/utility/tests/unit
mv internal/utility/log_test.go internal/utility/tests/unit/
mv internal/utility/timestamp_test.go internal/utility/tests/unit/
mv internal/utility/url_test.go internal/utility/tests/unit/
```

**Update package declarations if needed:**

```go
// Before (if white-box)
package db

// After (change to black-box)
package db_test
```

**Run tests to verify:**
```bash
go test ./...
```

---

### Phase 3: Add New Tests (Ongoing)

**When writing new tests, follow structure:**

1. **Decide test type:**
   - Unit test → `tests/unit/`
   - Integration test → `tests/integration/`
   - Benchmark → `tests/benchmark/`

2. **Create test file:**
   ```bash
   # Example: Testing new form builder
   touch internal/cli/tests/unit/form_builder_test.go
   ```

3. **Choose package:**
   ```go
   // Black-box (preferred)
   package cli_test

   // White-box (if testing internals)
   package cli
   ```

4. **Write test:**
   ```go
   func TestFormBuilder(t *testing.T) {
       // Arrange, Act, Assert
   }
   ```

---

## Benefits of This Approach

### For Development

**Clean production directories:**
```bash
# Before (with many tests)
$ ls internal/cli/
commands.go  commands_test.go  forms.go  forms_test.go  model.go  model_test.go
parse.go  parse_test.go  render.go  render_test.go  ...
# 88 files (44 code + 44 tests)

# After (with subdirectory)
$ ls internal/cli/
commands.go  forms.go  model.go  parse.go  render.go  ...  tests/
# 45 entries (44 code + 1 directory)

$ ls internal/cli/tests/
unit/  integration/  benchmark/  helpers/
```

**Easy test navigation:**
```bash
# Find all unit tests
ls internal/*/tests/unit/*_test.go

# Find all integration tests
ls internal/*/tests/integration/*_test.go

# Find all benchmarks
ls internal/*/tests/benchmark/*_bench_test.go
```

**Run selective tests:**
```bash
# All CLI unit tests
go test ./internal/cli/tests/unit/...

# All integration tests across project
go test ./internal/.../integration/...

# All benchmarks
go test -bench=. ./internal/.../benchmark/...
```

### For CI/CD

**Separate test stages:**

```yaml
# .github/workflows/test.yml
jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - run: go test ./internal/.../unit/...

  integration-tests:
    runs-on: ubuntu-latest
    needs: unit-tests
    steps:
      - run: go test ./internal/.../integration/...

  benchmarks:
    runs-on: ubuntu-latest
    needs: unit-tests
    steps:
      - run: go test -bench=. ./internal/.../benchmark/...
```

### For New Contributors

**Clear organization:**
- Want to see production code? Look in package directory
- Want to see tests? Look in `tests/` subdirectory
- Want to add test? Know exactly where it goes

**Matches documentation:**
```markdown
## Testing

Tests are organized by type in `tests/` subdirectories:
- `tests/unit/` - Unit tests
- `tests/integration/` - Integration tests
- `tests/benchmark/` - Performance benchmarks
- `tests/helpers/` - Test utilities
```

---

## Alternatives Considered

### Why NOT Standard Go Layout

**Problem:** Creates visual clutter in large packages

**CLI package example:**
```
commands.go          forms.go              model.go
commands_test.go     forms_test.go         model_test.go
constructors.go      handlers.go           page_builders.go
constructors_test.go handlers_test.go      page_builders_test.go
...
```

**With 44 files + tests → 88 files in one directory = Hard to parse**

### Why NOT Nested Packages (Option 3)

**Too invasive:**
- Requires refactoring entire package structure
- Changes all import paths
- Breaks existing code
- Better suited for new projects

**May still be valuable for Phase 2 CLI consolidation**, but separate concern.

### Why NOT Build Tags (Option 5)

**Doesn't solve the problem:**
- Test files still physically present
- Relies on editor configuration
- Inconsistent between developers
- Just hiding the problem, not organizing it

---

## Frequently Asked Questions

### Q: Is this approach Go-idiomatic?

**A:** It's less common than side-by-side tests, but still valid and used in production codebases:
- Kubernetes uses `test/` directories in some packages
- Docker uses separate test directories
- Many large Go projects organize tests this way

The Go community prefers simplicity, but **there's no rule against subdirectories for tests**.

### Q: Will Go tooling work?

**A:** Yes, all standard Go tools work:
- `go test ./...` - Discovers tests in subdirectories
- `go test -cover ./...` - Coverage reports work
- `go test -bench=.` - Benchmarks work
- IDE test runners - Work fine

### Q: Can I mix white-box and black-box tests?

**A:** Yes!
```go
// unit/commands_test.go - Black-box
package cli_test

// unit/parse_internals_test.go - White-box
package cli
```

Go allows multiple packages in same directory for testing.

### Q: What about small packages (< 10 files)?

**A:** Keep tests side-by-side for small packages:
```
internal/utility/
├── log.go
├── log_test.go
├── timestamp.go
└── timestamp_test.go
```

Only use subdirectories for packages with 20+ files.

### Q: How do I share test helpers across packages?

**A:** Create top-level test helpers:
```
internal/testutil/
├── db.go        - Database test utilities
├── fixtures.go  - Common test fixtures
└── mocks.go     - Shared mocks
```

Import in tests:
```go
import "github.com/hegner123/modulacms/internal/testutil"
```

---

## Summary Recommendation

### For ModulaCMS

**Use subdirectory organization for large packages (20+ files):**

```
internal/cli/tests/          ← Tests separate from production
internal/db/tests/           ← Tests separate from production
internal/model/tests/        ← Tests separate from production
```

**Keep side-by-side for small packages:**

```
internal/utility/            ← Tests alongside code (small package)
internal/config/             ← Tests alongside code (small package)
```

### Action Items

1. ✅ Create `tests/` subdirectories in cli, db, model packages
2. ✅ Add README.md explaining structure
3. ✅ Move existing tests to subdirectories
4. ✅ Create helper files for test utilities
5. ✅ Write new tests using subdirectory structure
6. ✅ Update documentation to reflect organization

### Expected Outcome

**Clean production directories:**
- Easy to browse code without test clutter
- Clear separation of concerns
- Tests organized by type (unit, integration, benchmark)

**Maintainable tests:**
- Easy to find specific test types
- Shared helpers reduce duplication
- CI/CD can run selective test suites

**Go-compatible:**
- Works with all standard tooling
- Discoverable by `go test ./...`
- IDE support unchanged

---

**Status:** ✅ Recommendation Complete - Ready to Implement

**Next Step:** Execute Phase 1 (Create test directory structure) whenever ready to add tests

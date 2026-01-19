# ADDING_FEATURES.md

Comprehensive guide for adding new features to ModulaCMS from concept to deployment.

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/ADDING_FEATURES.md`
**Purpose:** Master workflow document for end-to-end feature development
**Audience:** AI agents and developers implementing features

---

## Overview

This document provides a complete workflow for adding features to ModulaCMS. It covers the decision-making process, implementation steps, testing strategies, and deployment considerations.

**Key Principle:** ModulaCMS follows a structured flow from database schema → generated code → business logic → interface → deployment.

---

## Quick Decision Trees

### Decision Tree 1: Does This Feature Need Database Changes?

```
Does the feature require storing new data?
├─ YES → Does existing table have the field?
│   ├─ YES → Use existing table, skip to Step 3
│   └─ NO → Does the data logically fit in existing table?
│       ├─ YES → Add column to existing table (Step 1)
│       └─ NO → Create new table (Step 1)
└─ NO → Skip to Step 4 (Business Logic)
```

**Examples:**
- **Add "published" status to content** → Add column to content_data (existing table)
- **Add "comments" system** → Create new comments table (new domain concept)
- **Add "export to JSON" feature** → No database changes (read-only operation)

### Decision Tree 2: Does This Feature Need TUI Changes?

```
Does the feature need user interaction in the SSH TUI?
├─ YES → Is it a new screen or modify existing?
│   ├─ NEW SCREEN → Create new Bubbletea model (Step 5)
│   └─ MODIFY EXISTING → Update existing model (Step 5)
└─ NO → API/background feature only (Step 6)
```

**Examples:**
- **Content status toggle** → Modify existing content editor screen
- **Media browser** → Create new screen with navigation
- **Automatic backup cron** → No TUI (background job)

---

## The Complete Feature Development Flow

**ModulaCMS Architecture Flow:**
```
Schema Design
    ↓
SQL Files (schema + queries)
    ↓
sqlc Code Generation
    ↓
DbDriver Interface Update
    ↓
Driver Implementations (SQLite, MySQL, PostgreSQL)
    ↓
Model/Business Logic
    ↓
TUI Interface (if needed)
    ↓
HTTP/API Endpoints (if needed)
    ↓
Testing
    ↓
Deployment
```

---

## Step 1: Schema Design & Migration

**When:** Feature requires new database tables or columns.

### 1.1 Determine Migration Number

**Rule:** Schema migrations are numbered sequentially in `sql/schema/`.

```bash
# List existing schema directories
ls -1 sql/schema/

# Output example:
# 1_permissions/
# 2_roles/
# ...
# 20_datatypes_fields/
```

**Next number:** If highest is `20_datatypes_fields`, use `21_your_feature`.

### 1.2 Create Schema Directory

```bash
# Create schema directory
mkdir -p sql/schema/21_content_status
```

### 1.3 Write Schema for All Three Databases

**Critical:** ModulaCMS supports SQLite, MySQL, and PostgreSQL. Create schema files for all three.

**File structure:**
```
sql/schema/21_content_status/
├── schema.sql          # SQLite
├── schema_mysql.sql    # MySQL
└── schema_psql.sql     # PostgreSQL
```

**Example: Adding "status" to content_data**

**SQLite (`schema.sql`):**
```sql
-- Add status column to content_data table
ALTER TABLE content_data ADD COLUMN status INTEGER DEFAULT 0 NOT NULL;

-- Create index for status queries
CREATE INDEX IF NOT EXISTS idx_content_data_status ON content_data(status);
```

**MySQL (`schema_mysql.sql`):**
```sql
-- Add status column to content_data table
ALTER TABLE content_data ADD COLUMN status INT DEFAULT 0 NOT NULL;

-- Create index for status queries
CREATE INDEX idx_content_data_status ON content_data(status);
```

**PostgreSQL (`schema_psql.sql`):**
```sql
-- Add status column to content_data table
ALTER TABLE content_data ADD COLUMN status INTEGER DEFAULT 0 NOT NULL;

-- Create index for status queries
CREATE INDEX IF NOT EXISTS idx_content_data_status ON content_data(status);
```

**Database-Specific Considerations:**

| Feature | SQLite | MySQL | PostgreSQL |
|---------|--------|-------|------------|
| Auto-increment | INTEGER PRIMARY KEY | AUTO_INCREMENT | SERIAL |
| Boolean | INTEGER (0/1) | TINYINT | BOOLEAN |
| Text | TEXT | TEXT | TEXT |
| Foreign Keys | REFERENCES | FOREIGN KEY | FOREIGN KEY |
| NULL handling | NULL / NOT NULL | NULL / NOT NULL | NULL / NOT NULL |

### 1.4 Update Combined Schema Files

**Critical:** ModulaCMS uses combined schema files for fresh installations.

**Files to update:**
- `/Users/home/Documents/Code/Go_dev/modulacms/sql/all_schema.sql` (SQLite)
- `/Users/home/Documents/Code/Go_dev/modulacms/sql/all_schema_mysql.sql` (MySQL)
- `/Users/home/Documents/Code/Go_dev/modulacms/sql/all_schema_psql.sql` (PostgreSQL)

**Add your migration SQL to the appropriate section in each file.**

**Example location in all_schema.sql:**
```sql
-- ... existing tables ...

-- 21_content_status
ALTER TABLE content_data ADD COLUMN status INTEGER DEFAULT 0 NOT NULL;
CREATE INDEX IF NOT EXISTS idx_content_data_status ON content_data(status);

-- ... continue ...
```

---

## Step 2: Write SQL Queries with sqlc

**When:** Feature needs database operations (SELECT, INSERT, UPDATE, DELETE).

### 2.1 Choose Query File Location

**Query files are organized by domain:**
- `sql/mysql/content.sql` - Content-related queries
- `sql/mysql/datatype.sql` - Datatype queries
- `sql/mysql/field.sql` - Field queries
- `sql/mysql/user.sql` - User queries
- `sql/postgres/` - PostgreSQL versions
- Create new file if needed: `sql/mysql/comments.sql`

### 2.2 Write sqlc Annotated Queries

**sqlc annotations:**
- `-- name: QueryName :one` - Returns single row
- `-- name: QueryName :many` - Returns multiple rows
- `-- name: QueryName :exec` - Executes, no return
- `-- name: QueryName :execrows` - Returns affected row count

**Example: Status queries**

**File:** `sql/mysql/content.sql` (add to existing file)

```sql
-- name: GetContentByStatus :many
SELECT cd.*
FROM content_data cd
WHERE cd.route_id = ?
  AND cd.status = ?
ORDER BY cd.date_created DESC;

-- name: UpdateContentStatus :exec
UPDATE content_data
SET status = ?,
    date_modified = CURRENT_TIMESTAMP
WHERE content_data_id = ?;

-- name: CountContentByStatus :one
SELECT COUNT(*) as count
FROM content_data
WHERE route_id = ?
  AND status = ?;
```

**PostgreSQL version** (`sql/postgres/content.sql`):
```sql
-- name: GetContentByStatus :many
SELECT cd.*
FROM content_data cd
WHERE cd.route_id = $1
  AND cd.status = $2
ORDER BY cd.date_created DESC;

-- name: UpdateContentStatus :exec
UPDATE content_data
SET status = $1,
    date_modified = CURRENT_TIMESTAMP
WHERE content_data_id = $2;

-- name: CountContentByStatus :one
SELECT COUNT(*) as count
FROM content_data
WHERE route_id = $1
  AND status = $2;
```

**Note:** MySQL uses `?` placeholders, PostgreSQL uses `$1, $2, $3`.

### 2.3 Generate Go Code with sqlc

```bash
# Navigate to sql directory
cd sql

# Generate Go code from SQL queries
make sqlc

# Or use sqlc directly
sqlc generate
```

**Output locations:**
- MySQL queries → `internal/db-mysql/`
- PostgreSQL queries → `internal/db-psql/`
- SQLite queries → `internal/db-sqlite/`

**Generated files:**
- `models.go` - Struct definitions
- `content.sql.go` - Query functions
- `db.go` - Database interface

**Example generated function:**
```go
// GetContentByStatus
func (q *Queries) GetContentByStatus(ctx context.Context, arg GetContentByStatusParams) ([]ContentData, error) {
    // Generated implementation
}
```

---

## Step 3: Update DbDriver Interface

**When:** New queries were added in Step 2.

### 3.1 Add Methods to DbDriver Interface

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/driver.go`

```go
type DbDriver interface {
    // ... existing methods ...

    // Content status operations
    GetContentByStatus(ctx context.Context, routeID int64, status int32) ([]ContentData, error)
    UpdateContentStatus(ctx context.Context, contentDataID int64, status int32) error
    CountContentByStatus(ctx context.Context, routeID int64, status int32) (int64, error)
}
```

**Naming conventions:**
- Use descriptive method names
- Context as first parameter
- Return error as last return value
- Use domain types (ContentData, not raw structs)

### 3.2 Implement in All Three Drivers

**Critical:** Must implement in SQLite, MySQL, and PostgreSQL drivers.

#### SQLite Implementation

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db-sqlite/driver.go`

```go
func (d *Driver) GetContentByStatus(ctx context.Context, routeID int64, status int32) ([]db.ContentData, error) {
    rows, err := d.queries.GetContentByStatus(ctx, GetContentByStatusParams{
        RouteID: routeID,
        Status:  status,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to get content by status: %w", err)
    }

    // Convert from sqlc-generated type to db.ContentData
    result := make([]db.ContentData, len(rows))
    for i, row := range rows {
        result[i] = db.ContentData{
            ContentDataID: row.ContentDataID,
            Status:        row.Status,
            // ... map other fields ...
        }
    }

    return result, nil
}

func (d *Driver) UpdateContentStatus(ctx context.Context, contentDataID int64, status int32) error {
    err := d.queries.UpdateContentStatus(ctx, UpdateContentStatusParams{
        ContentDataID: contentDataID,
        Status:        status,
    })
    if err != nil {
        return fmt.Errorf("failed to update content status: %w", err)
    }
    return nil
}

func (d *Driver) CountContentByStatus(ctx context.Context, routeID int64, status int32) (int64, error) {
    count, err := d.queries.CountContentByStatus(ctx, CountContentByStatusParams{
        RouteID: routeID,
        Status:  status,
    })
    if err != nil {
        return 0, fmt.Errorf("failed to count content by status: %w", err)
    }
    return count, nil
}
```

#### MySQL Implementation

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db-mysql/driver.go`

```go
// Similar implementation as SQLite
// Map MySQL-specific types to db.ContentData
func (d *Driver) GetContentByStatus(ctx context.Context, routeID int64, status int32) ([]db.ContentData, error) {
    // Implementation (same pattern as SQLite)
}

func (d *Driver) UpdateContentStatus(ctx context.Context, contentDataID int64, status int32) error {
    // Implementation
}

func (d *Driver) CountContentByStatus(ctx context.Context, routeID int64, status int32) (int64, error) {
    // Implementation
}
```

#### PostgreSQL Implementation

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/db-psql/driver.go`

```go
// Similar implementation as SQLite and MySQL
func (d *Driver) GetContentByStatus(ctx context.Context, routeID int64, status int32) ([]db.ContentData, error) {
    // Implementation (same pattern)
}

func (d *Driver) UpdateContentStatus(ctx context.Context, contentDataID int64, status int32) error {
    // Implementation
}

func (d *Driver) CountContentByStatus(ctx context.Context, routeID int64, status int32) (int64, error) {
    // Implementation
}
```

---

## Step 4: Implement Business Logic

**When:** Feature has logic beyond database operations.

### 4.1 Choose Location for Business Logic

**Guidelines:**
- **Simple CRUD:** Put in driver implementations (done in Step 3)
- **Domain logic:** Create functions in `internal/model/`
- **HTTP handlers:** Put in appropriate HTTP handler file
- **TUI logic:** Put in Bubbletea Update function

### 4.2 Define Status Constants

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/model/content_status.go` (new file)

```go
package model

// Content status constants
const (
    ContentStatusDraft     int32 = 0
    ContentStatusPublished int32 = 1
    ContentStatusArchived  int32 = 2
)

// Status labels for display
var ContentStatusLabels = map[int32]string{
    ContentStatusDraft:     "Draft",
    ContentStatusPublished: "Published",
    ContentStatusArchived:  "Archived",
}

// GetStatusLabel returns human-readable status label
func GetStatusLabel(status int32) string {
    if label, ok := ContentStatusLabels[status]; ok {
        return label
    }
    return "Unknown"
}

// ValidateStatus checks if status value is valid
func ValidateStatus(status int32) bool {
    _, ok := ContentStatusLabels[status]
    return ok
}
```

### 4.3 Implement Helper Functions

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/model/content_operations.go` (add to existing)

```go
// PublishContent marks content as published
func PublishContent(ctx context.Context, db db.DbDriver, contentDataID int64) error {
    err := db.UpdateContentStatus(ctx, contentDataID, ContentStatusPublished)
    if err != nil {
        return fmt.Errorf("failed to publish content: %w", err)
    }

    utility.DefaultLogger.Info("Content published",
        "content_data_id", contentDataID,
        "status", ContentStatusPublished)

    return nil
}

// ArchiveContent marks content as archived
func ArchiveContent(ctx context.Context, db db.DbDriver, contentDataID int64) error {
    err := db.UpdateContentStatus(ctx, contentDataID, ContentStatusArchived)
    if err != nil {
        return fmt.Errorf("failed to archive content: %w", err)
    }

    utility.DefaultLogger.Info("Content archived",
        "content_data_id", contentDataID,
        "status", ContentStatusArchived)

    return nil
}
```

---

## Step 5: Add TUI Interface (If Needed)

**When:** Feature needs user interaction in SSH TUI.

### 5.1 Decide: New Screen or Modify Existing?

**New screen:** Create new Bubbletea model (see CREATING_TUI_SCREENS.md)
**Modify existing:** Update existing model's Update and View functions

### 5.2 Add Message Type

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/messages.go` (add to existing)

```go
// StatusChangedMsg is sent when content status changes
type StatusChangedMsg struct {
    ContentDataID int64
    NewStatus     int32
    Error         error
}
```

### 5.3 Add Keyboard Command

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/model.go` (Update function)

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.String() {
        // ... existing keys ...

        case "p": // Publish content
            if m.selectedNode != nil {
                return m, m.publishContent(m.selectedNode.Instance.ContentDataID)
            }

        case "a": // Archive content
            if m.selectedNode != nil {
                return m, m.archiveContent(m.selectedNode.Instance.ContentDataID)
            }
        }

    case StatusChangedMsg:
        if msg.Error != nil {
            m.errorMessage = fmt.Sprintf("Status change failed: %v", msg.Error)
        } else {
            m.successMessage = "Status updated successfully"
            // Refresh content tree
            return m, m.loadContentTree()
        }
    }

    return m, nil
}
```

### 5.4 Implement Command Functions

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/commands.go` (add to existing)

```go
// publishContent creates command to publish content
func (m *Model) publishContent(contentDataID int64) tea.Cmd {
    return func() tea.Msg {
        ctx := context.Background()
        err := model.PublishContent(ctx, m.db, contentDataID)
        return StatusChangedMsg{
            ContentDataID: contentDataID,
            NewStatus:     model.ContentStatusPublished,
            Error:         err,
        }
    }
}

// archiveContent creates command to archive content
func (m *Model) archiveContent(contentDataID int64) tea.Cmd {
    return func() tea.Msg {
        ctx := context.Background()
        err := model.ArchiveContent(ctx, m.db, contentDataID)
        return StatusChangedMsg{
            ContentDataID: contentDataID,
            NewStatus:     model.ContentStatusArchived,
            Error:         err,
        }
    }
}
```

### 5.5 Update View Function

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/view.go` (modify existing)

```go
func (m Model) renderContentItem(node *TreeNode) string {
    // Get status indicator
    statusIndicator := ""
    switch node.Instance.Status {
    case model.ContentStatusDraft:
        statusIndicator = lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render("●") // Yellow
    case model.ContentStatusPublished:
        statusIndicator = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("●") // Green
    case model.ContentStatusArchived:
        statusIndicator = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render("●") // Gray
    }

    return fmt.Sprintf("%s %s", statusIndicator, node.Datatype.Label)
}

func (m Model) renderHelpBar() string {
    help := []string{
        "↑/↓: Navigate",
        "enter: Edit",
        "p: Publish",
        "a: Archive",
        "q: Quit",
    }
    return lipgloss.JoinHorizontal(lipgloss.Left, help...)
}
```

---

## Step 6: Add HTTP/API Endpoints (If Needed)

**When:** Feature needs HTTP API access for external frontends.

### 6.1 Create Handler Function

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/handlers/content.go` (add to existing)

```go
// HandlePublishContent publishes content via API
func HandlePublishContent(db db.DbDriver) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Parse request
        var req struct {
            ContentDataID int64 `json:"content_data_id"`
        }
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Invalid request", http.StatusBadRequest)
            return
        }

        // Publish content
        ctx := r.Context()
        err := model.PublishContent(ctx, db, req.ContentDataID)
        if err != nil {
            utility.DefaultLogger.Error("Failed to publish content", "error", err)
            http.Error(w, "Failed to publish content", http.StatusInternalServerError)
            return
        }

        // Return success
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]any{
            "success": true,
            "message": "Content published successfully",
        })
    }
}
```

### 6.2 Register Route

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/cmd/main.go` (add to HTTP server setup)

```go
// Content status routes
mux.HandleFunc("POST /api/content/publish", handlers.HandlePublishContent(dbDriver))
mux.HandleFunc("POST /api/content/archive", handlers.HandleArchiveContent(dbDriver))
mux.HandleFunc("GET /api/content/status/{status}", handlers.HandleGetContentByStatus(dbDriver))
```

---

## Step 7: Write Tests

**Critical:** Every feature must have tests.

### 7.1 Create Test File

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/model/content_status_test.go`

```go
package model

import (
    "context"
    "testing"

    "github.com/modula/modulacms/internal/db"
    "github.com/modula/modulacms/internal/db-sqlite"
)

func TestPublishContent(t *testing.T) {
    // Setup test database
    driver, err := sqlite.NewDriver(":memory:")
    if err != nil {
        t.Fatalf("Failed to create test database: %v", err)
    }
    defer driver.Close()

    // Create test content
    ctx := context.Background()
    contentID, err := driver.CreateContentData(ctx, db.CreateContentDataParams{
        RouteID:    1,
        DatatypeID: 1,
        AuthorID:   1,
        Status:     ContentStatusDraft,
    })
    if err != nil {
        t.Fatalf("Failed to create test content: %v", err)
    }

    // Test publishing
    err = PublishContent(ctx, driver, contentID)
    if err != nil {
        t.Errorf("PublishContent failed: %v", err)
    }

    // Verify status changed
    content, err := driver.GetContentData(ctx, contentID)
    if err != nil {
        t.Fatalf("Failed to get content: %v", err)
    }

    if content.Status != ContentStatusPublished {
        t.Errorf("Expected status %d, got %d", ContentStatusPublished, content.Status)
    }
}

func TestValidateStatus(t *testing.T) {
    tests := []struct {
        status int32
        valid  bool
    }{
        {ContentStatusDraft, true},
        {ContentStatusPublished, true},
        {ContentStatusArchived, true},
        {999, false},
        {-1, false},
    }

    for _, tt := range tests {
        result := ValidateStatus(tt.status)
        if result != tt.valid {
            t.Errorf("ValidateStatus(%d) = %v, want %v", tt.status, result, tt.valid)
        }
    }
}
```

### 7.2 Run Tests

```bash
# Run all tests
make test

# Run specific package tests
go test -v ./internal/model

# Run specific test
go test -v ./internal/model -run TestPublishContent

# Run with coverage
go test -cover ./internal/model
```

---

## Step 8: Update Documentation

### 8.1 Update CLAUDE.md

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/CLAUDE.md`

Add feature to appropriate section:

```markdown
## Content Status

ModulaCMS supports content status workflow:
- Draft (0) - Content being worked on
- Published (1) - Live content
- Archived (2) - Hidden from public

**TUI Commands:**
- `p` - Publish selected content
- `a` - Archive selected content

**API Endpoints:**
- `POST /api/content/publish` - Publish content
- `POST /api/content/archive` - Archive content
- `GET /api/content/status/{status}` - List content by status
```

### 8.2 Update README (If User-Facing)

**File:** `/Users/home/Documents/Code/Go_dev/modulacms/README.md`

Add to features section if user-facing.

---

## Step 9: Commit Changes

**Follow git commit guidelines from CLAUDE.md:**

```bash
# Check status
git status

# See changes
git diff

# Add files
git add sql/schema/21_content_status/
git add sql/all_schema*.sql
git add sql/mysql/content.sql
git add sql/postgres/content.sql
git add internal/db/driver.go
git add internal/db-sqlite/driver.go
git add internal/db-mysql/driver.go
git add internal/db-psql/driver.go
git add internal/model/content_status.go
git add internal/model/content_operations.go
git add internal/cli/messages.go
git add internal/cli/model.go
git add internal/cli/commands.go
git add internal/cli/view.go

# Commit with descriptive message
git commit -m "$(cat <<'EOF'
Add content status workflow (draft/published/archived)

- Add status column to content_data table
- Create queries for status filtering and updates
- Implement status validation and helper functions
- Add TUI commands for publish/archive (p/a keys)
- Add status indicators in content tree view
- Include tests for status operations

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>
EOF
)"
```

---

## Step 10: Deployment

### 10.1 Build Binary

```bash
# Build for development
make dev

# Build for production (AMD64 and x86)
make build
```

### 10.2 Test Locally

```bash
# Run with test database
./modulacms-x86 --config=config.test.json

# SSH into TUI
ssh user@localhost -p 2222

# Test new feature:
# - Navigate to content item
# - Press 'p' to publish
# - Verify status indicator changes
```

### 10.3 Deploy to Production

```bash
# Deploy using Makefile
make build  # Builds and rsyncs to remote server

# Or manually:
scp modulacms-amd64 user@server:/path/to/app/
ssh user@server
sudo systemctl restart modulacms
```

### 10.4 Verify in Production

```bash
# Check logs
ssh user@server
tail -f /path/to/logs/modulacms.log

# Test feature
ssh -p 2222 user@server
# Test new functionality
```

---

## Common Feature Patterns

### Pattern 1: Adding a Column to Existing Table

**Use when:** Extending existing data model with new property.

**Steps:**
1. Create migration: `sql/schema/N_feature_name/`
2. Add ALTER TABLE statements (all 3 databases)
3. Update combined schema files
4. Add queries if needed (usually UPDATE and SELECT)
5. Run `make sqlc`
6. Update DbDriver interface (if new queries)
7. Implement in all drivers
8. Update business logic
9. Update TUI/API (if needed)
10. Write tests
11. Commit and deploy

**Example features:**
- Content status (this doc)
- Content tags
- Author bio field
- Published date

### Pattern 2: Creating a New Table

**Use when:** Adding entirely new domain concept.

**Steps:**
1. Create migration: `sql/schema/N_table_name/`
2. Write CREATE TABLE (all 3 databases)
3. Add foreign keys to related tables
4. Update combined schema files
5. Write CRUD queries
6. Run `make sqlc`
7. Add methods to DbDriver interface
8. Implement in all drivers
9. Create model functions
10. Create TUI screen (see CREATING_TUI_SCREENS.md)
11. Add API endpoints
12. Write tests
13. Commit and deploy

**Example features:**
- Comments system
- Tags/taxonomy
- Media library
- User preferences

### Pattern 3: Read-Only Feature (No Database Changes)

**Use when:** Feature uses existing data in new way.

**Steps:**
1. Skip Step 1-3 (no database changes)
2. Implement business logic in `internal/model/`
3. Add TUI interface or HTTP endpoint
4. Write tests
5. Update documentation
6. Commit and deploy

**Example features:**
- Export content to JSON
- Content statistics dashboard
- Search functionality
- Content preview

### Pattern 4: Background Job/Cron

**Use when:** Feature runs automatically without user interaction.

**Steps:**
1. Implement business logic
2. Create cron/scheduler in `cmd/main.go`
3. Add configuration to config.json
4. Write tests (especially error handling)
5. Add logging
6. Commit and deploy

**Example features:**
- Automatic backups
- Content archiving
- Email notifications
- Cache clearing

---

## Testing Checklist

Before considering feature complete:

- [ ] Unit tests written for all business logic
- [ ] Database operations tested with all three drivers (SQLite, MySQL, PostgreSQL)
- [ ] TUI commands tested manually via SSH
- [ ] API endpoints tested with curl/Postman
- [ ] Error cases handled and tested
- [ ] Edge cases tested (empty data, invalid input, etc.)
- [ ] Performance tested with realistic data volume
- [ ] Logging added for debugging
- [ ] `make test` passes
- [ ] `make lint` passes
- [ ] Documentation updated

---

## Common Pitfalls and Solutions

### Pitfall 1: Forgetting to Update All Three Drivers

**Problem:** Feature works in SQLite but fails in MySQL/PostgreSQL.

**Solution:**
- Always implement DbDriver methods in all three drivers
- Test with all three databases before deploying
- Run `make test` which uses SQLite
- Manually test with MySQL and PostgreSQL in dev

### Pitfall 2: SQL Differences Between Databases

**Problem:** Query works in MySQL but not PostgreSQL.

**Solution:**
- MySQL uses `?` placeholders, PostgreSQL uses `$1, $2, $3`
- MySQL: `AUTO_INCREMENT`, PostgreSQL: `SERIAL`
- MySQL: `LIMIT ?, ?`, PostgreSQL: `LIMIT $1 OFFSET $2`
- Test queries in all databases

### Pitfall 3: Not Updating Combined Schema Files

**Problem:** Fresh installs fail because migration not in combined schema.

**Solution:**
- Always update `all_schema.sql`, `all_schema_mysql.sql`, `all_schema_psql.sql`
- These are used for fresh installations
- Migrations are for existing installations

### Pitfall 4: Breaking Existing Functionality

**Problem:** New feature breaks existing code.

**Solution:**
- Run full test suite before committing: `make test`
- Test TUI manually for regressions
- Check that existing API endpoints still work
- Review git diff carefully

### Pitfall 5: Poor Error Handling

**Problem:** Feature crashes on error instead of gracefully handling.

**Solution:**
- Always check errors: `if err != nil`
- Wrap errors with context: `fmt.Errorf("context: %w", err)`
- Log errors: `utility.DefaultLogger.Error("message", "error", err)`
- Return user-friendly error messages in TUI/API

### Pitfall 6: Ignoring Performance at Scale

**Problem:** Feature works with 10 items but times out with 1,000.

**Solution:**
- Think about scale from day one (see TREE_STRUCTURE.md performance notes)
- Add indexes for filtered/sorted columns
- Use lazy loading for large datasets
- Test with realistic data volumes (1,000+ rows)

---

## Debugging Tips

### Database Issues

```bash
# Check generated SQL
cat internal/db-mysql/content.sql.go

# Test query directly
sqlite3 modulacms.db
> SELECT * FROM content_data WHERE status = 1;

# Check migrations applied
ls -la sql/schema/
```

### TUI Issues

```bash
# Run in CLI mode with logs
./modulacms-x86 --cli 2>debug.log

# Check logs
tail -f debug.log

# Add debug logging
utility.DefaultLogger.Debug("Debug message", "key", value)
```

### API Issues

```bash
# Test endpoint with curl
curl -X POST http://localhost:8080/api/content/publish \
  -H "Content-Type: application/json" \
  -d '{"content_data_id": 123}'

# Check HTTP server logs
tail -f logs/http.log
```

---

## Related Documentation

**Architecture:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/TREE_STRUCTURE.md` - Understanding content trees
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/CONTENT_MODEL.md` - Domain model relationships
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/DATABASE_LAYER.md` - DbDriver abstraction

**Workflows:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/ADDING_TABLES.md` - Creating new tables
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/CREATING_TUI_SCREENS.md` - New TUI screens
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/TESTING.md` - Testing strategies

**Database:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/database/SQL_DIRECTORY.md` - SQL organization
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/database/SQLC.md` - sqlc reference
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/database/DB_PACKAGE.md` - Database packages

**Packages:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/packages/CLI_PACKAGE.md` - TUI implementation

**General:**
- `/Users/home/Documents/Code/Go_dev/modulacms/CLAUDE.md` - Development guidelines
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/FILE_TREE.md` - Project structure

---

## Quick Reference

### Complete Feature Checklist

- [ ] **Step 1:** Schema migration (if needed)
- [ ] **Step 2:** SQL queries with sqlc (if needed)
- [ ] **Step 3:** Update DbDriver interface and implement in all drivers
- [ ] **Step 4:** Business logic in model package
- [ ] **Step 5:** TUI interface (if needed)
- [ ] **Step 6:** HTTP/API endpoints (if needed)
- [ ] **Step 7:** Write comprehensive tests
- [ ] **Step 8:** Update documentation
- [ ] **Step 9:** Commit with descriptive message
- [ ] **Step 10:** Deploy and verify

### Key Commands

```bash
make sqlc          # Generate Go code from SQL
make test          # Run all tests
make lint          # Run linters
make dev           # Build local binary
make build         # Build and deploy production
./modulacms-x86    # Run locally
```

### Key Files

**Schema:**
- `sql/schema/N_feature/` - Migrations
- `sql/all_schema*.sql` - Combined schemas

**Queries:**
- `sql/mysql/*.sql` - MySQL queries
- `sql/postgres/*.sql` - PostgreSQL queries

**Database:**
- `internal/db/driver.go` - DbDriver interface
- `internal/db-sqlite/driver.go` - SQLite implementation
- `internal/db-mysql/driver.go` - MySQL implementation
- `internal/db-psql/driver.go` - PostgreSQL implementation

**Business Logic:**
- `internal/model/*.go` - Domain logic

**TUI:**
- `internal/cli/model.go` - Main model
- `internal/cli/messages.go` - Message types
- `internal/cli/commands.go` - Commands
- `internal/cli/view.go` - Rendering

**API:**
- `internal/handlers/*.go` - HTTP handlers
- `cmd/main.go` - Route registration

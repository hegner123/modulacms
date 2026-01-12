# DEBUGGING.md

Comprehensive guide to debugging ModulaCMS applications, from common scenarios to advanced profiling techniques.

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/DEBUGGING.md`
**Purpose:** Practical debugging strategies for ModulaCMS development
**Last Updated:** 2026-01-12

---

## Overview

ModulaCMS debugging involves several distinct areas:
- **Database operations** - sqlc queries, transactions, foreign keys
- **TUI state management** - Elm Architecture, message flow, rendering
- **Tree operations** - Loading, traversal, circular references, orphans
- **HTTP/SSH servers** - Request handling, middleware, sessions
- **Performance** - Query optimization, memory usage, goroutine leaks

**Key Debugging Tools:**
- `utility.DefaultLogger` - Structured logging throughout the application
- Go's built-in `delve` debugger - Breakpoints and inspection
- `go test -v` - Verbose test output
- `pprof` - CPU and memory profiling
- Database query logs - SQL execution tracing

---

## Logging with utility.DefaultLogger

ModulaCMS uses Charmbracelet Log for structured logging throughout the application.

### Logger Setup

**Location:** `/Users/home/Documents/Code/Go_dev/modulacms/internal/utility/utility.go`

```go
var DefaultLogger = log.NewWithOptions(os.Stderr, log.Options{
    ReportCaller:    true,
    ReportTimestamp: true,
    TimeFormat:      time.Kitchen,
})
```

**Configuration:**
- `ReportCaller: true` - Shows file and line number
- `ReportTimestamp: true` - Includes timestamp
- Output to stderr - Separate from application output

### Logging Levels

```go
// Info - General information
utility.DefaultLogger.Info("Content tree loaded",
    "route_id", routeID,
    "nodes", len(treeRoot.NodeIndex))

// Debug - Detailed debugging info
utility.DefaultLogger.Debug("Processing tree node",
    "content_data_id", node.Instance.ContentDataID,
    "parent_id", node.Instance.ParentID,
    "datatype", node.Datatype.Label)

// Warn - Potential issues
utility.DefaultLogger.Warn("Orphaned node detected",
    "node_id", nodeID,
    "parent_id", parentID,
    "retry_attempt", retryCount)

// Error - Actual errors
utility.DefaultLogger.Error("Failed to load content tree",
    "error", err,
    "route_id", routeID)

// Fatal - Unrecoverable errors (exits program)
utility.DefaultLogger.Fatal("Database connection failed",
    "error", err,
    "config", configPath)
```

### Structured Logging Best Practices

**Use key-value pairs:**
```go
// Good - structured with context
utility.DefaultLogger.Info("User authenticated",
    "user_id", userID,
    "provider", "github",
    "duration_ms", elapsed.Milliseconds())

// Bad - string concatenation
utility.DefaultLogger.Info(fmt.Sprintf("User %d authenticated via %s", userID, provider))
```

**Include relevant IDs:**
```go
utility.DefaultLogger.Error("Content creation failed",
    "error", err,
    "route_id", routeID,
    "datatype_id", datatypeID,
    "parent_id", parentID,
    "user_id", userID)
```

**Log at decision points:**
```go
if parent == nil {
    utility.DefaultLogger.Warn("Parent node not found",
        "node_id", nodeID,
        "parent_id", parentID)
    // Add to orphans
} else {
    utility.DefaultLogger.Debug("Attaching node to parent",
        "node_id", nodeID,
        "parent_id", parentID)
    // Attach node
}
```

### Temporary Debug Logging

**Add detailed logging for investigation:**
```go
// Temporary debugging - remove after issue resolved
utility.DefaultLogger.Debug("=== TREE LOADING DEBUG START ===")
for id, node := range treeRoot.NodeIndex {
    utility.DefaultLogger.Debug("Node in index",
        "id", id,
        "has_parent", node.Parent != nil,
        "has_first_child", node.FirstChild != nil,
        "has_next_sibling", node.NextSibling != nil,
        "datatype", node.Datatype.Label)
}
utility.DefaultLogger.Debug("=== TREE LOADING DEBUG END ===")
```

### Conditional Logging

**Enable detailed logging based on environment:**
```go
var DebugMode = os.Getenv("MODULACMS_DEBUG") == "true"

if DebugMode {
    utility.DefaultLogger.SetLevel(log.DebugLevel)
    utility.DefaultLogger.Debug("Debug mode enabled")
}
```

---

## Debugging Database Issues

### Query Execution Problems

**Problem:** Query returns unexpected results or fails.

**Debugging Steps:**

1. **Log the query parameters:**
```go
utility.DefaultLogger.Debug("Executing GetContentTreeByRoute",
    "route_id", routeID)

rows, err := db.GetContentTreeByRoute(ctx, routeID)
if err != nil {
    utility.DefaultLogger.Error("Query failed",
        "error", err,
        "route_id", routeID)
    return err
}

utility.DefaultLogger.Info("Query succeeded",
    "route_id", routeID,
    "rows_returned", len(rows))
```

2. **Test query in database directly:**
```bash
# SQLite
sqlite3 modulacms.db "SELECT * FROM content_data WHERE route_id = 1"

# MySQL
mysql -u root -p -e "SELECT * FROM content_data WHERE route_id = 1"

# PostgreSQL
psql -U postgres -c "SELECT * FROM content_data WHERE route_id = 1"
```

3. **Verify sqlc-generated code:**
```go
// Check generated file
// Location: internal/db-sqlite/content_data.sql.go (or db-mysql, db-psql)

// Ensure query matches expectation
// Verify parameter types match
```

### Foreign Key Constraint Violations

**Problem:** Insert or update fails with foreign key error.

**Symptoms:**
```
Error: FOREIGN KEY constraint failed
Error: Cannot add or update a child row: a foreign key constraint fails
```

**Debugging:**

```go
// Log all foreign key values before insert
utility.DefaultLogger.Debug("Creating content data",
    "route_id", params.RouteID,
    "datatype_id", params.DatatypeID,
    "parent_id", params.ParentID,
    "author_id", params.AuthorID)

// Verify referenced rows exist
route, err := db.GetRouteByID(ctx, params.RouteID)
if err != nil {
    utility.DefaultLogger.Error("Route does not exist",
        "route_id", params.RouteID)
    return err
}

datatype, err := db.GetDatatypeByID(ctx, params.DatatypeID)
if err != nil {
    utility.DefaultLogger.Error("Datatype does not exist",
        "datatype_id", params.DatatypeID)
    return err
}
```

**Check database constraints:**
```sql
-- SQLite: List foreign keys
PRAGMA foreign_key_list(content_data);

-- MySQL: Show create table
SHOW CREATE TABLE content_data;

-- PostgreSQL: List constraints
SELECT conname, conrelid::regclass, confrelid::regclass
FROM pg_constraint
WHERE contype = 'f' AND conrelid = 'content_data'::regclass;
```

### Transaction Issues

**Problem:** Changes not committing or partial commits.

**Debugging:**

```go
// Log transaction boundaries
utility.DefaultLogger.Debug("Starting transaction")
tx, err := db.BeginTx(ctx)
if err != nil {
    utility.DefaultLogger.Error("Failed to begin transaction", "error", err)
    return err
}
defer func() {
    if err != nil {
        utility.DefaultLogger.Warn("Rolling back transaction", "error", err)
        tx.Rollback()
    }
}()

// Log each operation
utility.DefaultLogger.Debug("Creating content data")
contentDataID, err := tx.CreateContentData(ctx, contentParams)
if err != nil {
    utility.DefaultLogger.Error("CreateContentData failed", "error", err)
    return err
}

utility.DefaultLogger.Debug("Creating content fields", "count", len(fields))
for i, field := range fields {
    err = tx.CreateContentField(ctx, field)
    if err != nil {
        utility.DefaultLogger.Error("CreateContentField failed",
            "error", err,
            "field_index", i,
            "field_id", field.FieldID)
        return err
    }
}

utility.DefaultLogger.Debug("Committing transaction")
if err = tx.Commit(); err != nil {
    utility.DefaultLogger.Error("Commit failed", "error", err)
    return err
}

utility.DefaultLogger.Info("Transaction committed successfully",
    "content_data_id", contentDataID)
```

### NULL Value Handling

**Problem:** NULL values causing unexpected behavior.

**Debugging:**

```go
// Check sql.Null* types carefully
if node.Instance.ParentID.Valid {
    utility.DefaultLogger.Debug("Node has parent",
        "node_id", node.Instance.ContentDataID,
        "parent_id", node.Instance.ParentID.Int64)
} else {
    utility.DefaultLogger.Debug("Node is root (no parent)",
        "node_id", node.Instance.ContentDataID)
}

// Common mistake: forgetting to check .Valid
// BAD:
parentID := node.Instance.ParentID.Int64  // May be 0 even if NULL!

// GOOD:
if node.Instance.ParentID.Valid {
    parentID := node.Instance.ParentID.Int64
    // Use parentID
} else {
    // Handle NULL case
}
```

### Database Driver Issues

**Problem:** Query works in one driver but not another.

**Debugging:**

```go
// Add driver-specific logging
utility.DefaultLogger.Info("Using database driver",
    "driver", "sqlite",  // or "mysql", "postgres"
    "version", version)

// Test same query across drivers
// SQLite uses ? for placeholders
// PostgreSQL uses $1, $2, $3 for placeholders
// Ensure sqlc generates correct syntax for each
```

---

## Debugging TUI State Issues

### Message Flow Problems

**Problem:** Messages not triggering expected state changes.

**Debugging:**

```go
// Add logging to Update function
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    utility.DefaultLogger.Debug("Update called",
        "msg_type", fmt.Sprintf("%T", msg),
        "current_mode", m.Mode,
        "tree_loaded", m.TreeRoot != nil)

    switch msg := msg.(type) {
    case tea.KeyMsg:
        utility.DefaultLogger.Debug("Key pressed",
            "key", msg.String(),
            "mode", m.Mode)

        switch msg.String() {
        case "j", "down":
            utility.DefaultLogger.Debug("Moving cursor down",
                "current_cursor", m.Cursor,
                "tree_size", len(m.VisibleNodes))
            // Handle movement
        }

    case ContentLoadedMsg:
        utility.DefaultLogger.Info("Content loaded",
            "route_id", msg.RouteID,
            "node_count", len(msg.TreeRoot.NodeIndex))
        m.TreeRoot = msg.TreeRoot
        return m, nil
    }

    return m, nil
}
```

### View Rendering Issues

**Problem:** Screen not updating or showing incorrect data.

**Debugging:**

```go
// Add logging to View function
func (m Model) View() string {
    utility.DefaultLogger.Debug("View rendering",
        "mode", m.Mode,
        "cursor", m.Cursor,
        "visible_nodes", len(m.VisibleNodes))

    if m.TreeRoot == nil {
        utility.DefaultLogger.Warn("Rendering with nil TreeRoot")
        return "Loading..."
    }

    // Log rendering decisions
    utility.DefaultLogger.Debug("Rendering tree",
        "total_nodes", len(m.TreeRoot.NodeIndex),
        "visible_nodes", len(m.VisibleNodes))

    // Build view
    var b strings.Builder
    for i, node := range m.VisibleNodes {
        if i == m.Cursor {
            utility.DefaultLogger.Debug("Rendering cursor node",
                "index", i,
                "node_id", node.Instance.ContentDataID)
        }
        // Render node
    }

    return b.String()
}
```

### Command Execution Issues

**Problem:** Commands not executing or returning wrong results.

**Debugging:**

```go
// Log command creation
func loadContentTreeCmd(routeID int64, db db.DbDriver) tea.Cmd {
    return func() tea.Msg {
        utility.DefaultLogger.Debug("Loading content tree",
            "route_id", routeID)

        rows, err := db.GetContentTreeByRoute(context.Background(), routeID)
        if err != nil {
            utility.DefaultLogger.Error("Failed to load tree",
                "error", err,
                "route_id", routeID)
            return ErrorMsg{Err: err}
        }

        utility.DefaultLogger.Info("Tree data loaded",
            "route_id", routeID,
            "rows", len(rows))

        // Build tree
        treeRoot := model.NewTreeRoot()
        stats, err := treeRoot.LoadFromRows(&rows)
        if err != nil {
            utility.DefaultLogger.Error("Failed to build tree",
                "error", err,
                "stats", stats)
            return ErrorMsg{Err: err}
        }

        utility.DefaultLogger.Info("Tree loaded successfully",
            "nodes", stats.NodesCount,
            "orphans_resolved", stats.OrphansResolved)

        return ContentLoadedMsg{TreeRoot: treeRoot, RouteID: routeID}
    }
}
```

### State Inconsistencies

**Problem:** Model state becomes inconsistent.

**Debugging:**

```go
// Validate state at key points
func (m *Model) validateState() error {
    errors := []string{}

    if m.Cursor < 0 {
        errors = append(errors, fmt.Sprintf("cursor negative: %d", m.Cursor))
    }

    if m.Cursor >= len(m.VisibleNodes) {
        errors = append(errors, fmt.Sprintf("cursor out of bounds: %d >= %d",
            m.Cursor, len(m.VisibleNodes)))
    }

    if m.TreeRoot != nil && m.TreeRoot.Root == nil {
        errors = append(errors, "TreeRoot has nil Root")
    }

    if len(errors) > 0 {
        utility.DefaultLogger.Error("State validation failed",
            "errors", strings.Join(errors, "; "))
        return fmt.Errorf("invalid state: %s", strings.Join(errors, "; "))
    }

    return nil
}

// Call after state changes
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // ... handle message ...

    if err := m.validateState(); err != nil {
        utility.DefaultLogger.Error("State validation failed", "error", err)
    }

    return m, cmd
}
```

---

## Debugging Tree Operations

### Tree Loading Failures

**Problem:** Tree fails to load with orphans or circular references.

**Debugging:**

```go
// Enable detailed tree loading logs
stats, err := treeRoot.LoadFromRows(&rows)

utility.DefaultLogger.Info("Tree loading stats",
    "nodes_count", stats.NodesCount,
    "orphans_resolved", stats.OrphansResolved,
    "retry_attempts", stats.RetryAttempts,
    "circular_refs", stats.CircularRefs,
    "final_orphans", stats.FinalOrphans)

if err != nil {
    utility.DefaultLogger.Error("Tree loading failed", "error", err)

    // Investigate orphans
    if len(stats.FinalOrphans) > 0 {
        utility.DefaultLogger.Error("Unresolved orphans detected",
            "count", len(stats.FinalOrphans),
            "orphan_ids", stats.FinalOrphans)

        // Check database for missing parents
        for _, orphanID := range stats.FinalOrphans {
            node := treeRoot.NodeIndex[orphanID]
            if node != nil && node.Instance.ParentID.Valid {
                parentID := node.Instance.ParentID.Int64
                utility.DefaultLogger.Error("Orphan details",
                    "orphan_id", orphanID,
                    "missing_parent_id", parentID)

                // Check if parent exists in database
                _, err := db.GetContentDataByID(ctx, parentID)
                if err != nil {
                    utility.DefaultLogger.Error("Parent does not exist in database",
                        "parent_id", parentID,
                        "error", err)
                }
            }
        }
    }

    // Investigate circular references
    if len(stats.CircularRefs) > 0 {
        utility.DefaultLogger.Error("Circular references detected",
            "count", len(stats.CircularRefs),
            "node_ids", stats.CircularRefs)

        // Trace circular chains
        for _, nodeID := range stats.CircularRefs {
            utility.DefaultLogger.Error("Circular reference chain",
                "start_node", nodeID,
                "chain", traceParentChain(treeRoot, nodeID))
        }
    }
}
```

**Helper function to trace parent chain:**
```go
func traceParentChain(tree *model.TreeRoot, startNodeID int64) []int64 {
    chain := []int64{}
    visited := make(map[int64]bool)
    currentID := startNodeID

    for {
        node := tree.NodeIndex[currentID]
        if node == nil {
            break
        }

        if visited[currentID] {
            chain = append(chain, currentID) // Show where cycle completes
            break
        }

        chain = append(chain, currentID)
        visited[currentID] = true

        if !node.Instance.ParentID.Valid {
            break // Reached root
        }

        currentID = node.Instance.ParentID.Int64
    }

    return chain
}
```

### Node Index Issues

**Problem:** NodeIndex doesn't contain expected nodes.

**Debugging:**

```go
// Verify NodeIndex contents
utility.DefaultLogger.Debug("NodeIndex contents",
    "size", len(treeRoot.NodeIndex))

for id, node := range treeRoot.NodeIndex {
    utility.DefaultLogger.Debug("Node in index",
        "content_data_id", id,
        "datatype", node.Datatype.Label,
        "has_parent", node.Parent != nil,
        "has_children", node.FirstChild != nil,
        "parent_id", node.Instance.ParentID)
}

// Check if specific node is in index
expectedNodeID := int64(123)
if node, exists := treeRoot.NodeIndex[expectedNodeID]; exists {
    utility.DefaultLogger.Info("Node found in index",
        "node_id", expectedNodeID,
        "datatype", node.Datatype.Label)
} else {
    utility.DefaultLogger.Warn("Node not found in index",
        "node_id", expectedNodeID)

    // Check if it exists in database
    dbNode, err := db.GetContentDataByID(ctx, expectedNodeID)
    if err != nil {
        utility.DefaultLogger.Error("Node not in database either",
            "node_id", expectedNodeID,
            "error", err)
    } else {
        utility.DefaultLogger.Warn("Node exists in DB but not in tree",
            "node_id", expectedNodeID,
            "route_id", dbNode.RouteID)
    }
}
```

### Tree Traversal Issues

**Problem:** Traversal missing nodes or entering infinite loops.

**Debugging:**

```go
// Safe traversal with visited tracking
func debugTraverseTree(node *model.TreeNode, visited map[int64]bool, depth int) {
    if node == nil {
        return
    }

    nodeID := node.Instance.ContentDataID

    // Check for cycles
    if visited[nodeID] {
        utility.DefaultLogger.Error("Cycle detected in tree traversal",
            "node_id", nodeID,
            "depth", depth)
        return
    }

    visited[nodeID] = true

    utility.DefaultLogger.Debug("Traversing node",
        "node_id", nodeID,
        "datatype", node.Datatype.Label,
        "depth", depth,
        "has_children", node.FirstChild != nil,
        "has_siblings", node.NextSibling != nil)

    // Traverse children
    child := node.FirstChild
    siblingCount := 0
    for child != nil {
        siblingCount++
        debugTraverseTree(child, visited, depth+1)
        child = child.NextSibling

        // Safety check for sibling loops
        if siblingCount > 1000 {
            utility.DefaultLogger.Error("Too many siblings, possible loop",
                "parent_id", nodeID,
                "sibling_count", siblingCount)
            break
        }
    }
}

// Use it:
visited := make(map[int64]bool)
debugTraverseTree(treeRoot.Root, visited, 0)

utility.DefaultLogger.Info("Traversal complete",
    "nodes_visited", len(visited),
    "nodes_in_index", len(treeRoot.NodeIndex))

if len(visited) != len(treeRoot.NodeIndex) {
    utility.DefaultLogger.Warn("Traversal didn't reach all nodes",
        "visited", len(visited),
        "total", len(treeRoot.NodeIndex))
}
```

---

## Using Go Debugger (Delve)

### Installation

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest
```

### Debugging Tests

```bash
# Debug a specific test
dlv test ./internal/model -- -test.run TestTreeLoading

# Set breakpoint before running
(dlv) break cms_struct.go:85
(dlv) continue

# Inspect variables
(dlv) print treeRoot.NodeIndex
(dlv) print node.Instance.ContentDataID
(dlv) print stats

# Step through code
(dlv) next        # next line
(dlv) step        # step into function
(dlv) stepout     # step out of function
(dlv) continue    # continue to next breakpoint
```

### Debugging Application

```bash
# Debug the main application
dlv debug ./cmd -- --cli

# Or attach to running process
dlv attach $(pgrep modulacms)
```

### Useful Delve Commands

```bash
# Breakpoints
(dlv) break internal/cli/update.go:123
(dlv) break Model.Update
(dlv) breakpoints                    # list all breakpoints
(dlv) clear 1                        # clear breakpoint #1

# Inspection
(dlv) print variableName
(dlv) print m.TreeRoot.NodeIndex[123]
(dlv) locals                         # show local variables
(dlv) args                           # show function arguments
(dlv) vars                           # show package variables

# Stack traces
(dlv) stack                          # show stack trace
(dlv) frame 2                        # switch to frame 2
(dlv) up                             # move up one frame
(dlv) down                           # move down one frame

# Goroutines
(dlv) goroutines                     # list all goroutines
(dlv) goroutine 5                    # switch to goroutine 5

# Conditional breakpoints
(dlv) break cms_struct.go:100
(dlv) condition 1 node.Instance.ContentDataID == 123

# Watch expressions
(dlv) watch treeRoot.NodeIndex
```

### Debugging TUI State

```bash
# Break in Update function
(dlv) break internal/cli/update.go:UpdateFunction

# Inspect message
(dlv) print msg
(dlv) print fmt.Sprintf("%T", msg)

# Inspect model state
(dlv) print m.Mode
(dlv) print m.Cursor
(dlv) print len(m.VisibleNodes)
(dlv) print m.TreeRoot != nil

# Inspect specific node
(dlv) print m.VisibleNodes[m.Cursor].Instance.ContentDataID
```

---

## Performance Profiling

### CPU Profiling

**Enable profiling in tests:**
```bash
# Run tests with CPU profiling
go test -cpuprofile=cpu.prof ./internal/model

# Analyze profile
go tool pprof cpu.prof

# In pprof:
(pprof) top10        # Top 10 functions by CPU time
(pprof) list LoadFromRows  # Show source with timings
(pprof) web          # Open graphical view
```

**Profile running application:**
```go
import (
    "runtime/pprof"
    "os"
)

// Add to main.go or init function
func enableProfiling() {
    f, err := os.Create("cpu.prof")
    if err != nil {
        log.Fatal(err)
    }
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()
}
```

### Memory Profiling

```bash
# Memory profile during tests
go test -memprofile=mem.prof ./internal/model

# Analyze memory usage
go tool pprof mem.prof

# In pprof:
(pprof) top10                    # Top memory allocators
(pprof) list NewTreeRoot         # Memory allocations in function
(pprof) web                      # Visual graph
```

**Check for memory leaks:**
```go
import "runtime"

// Add periodic memory stats logging
func logMemoryStats() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)

    utility.DefaultLogger.Info("Memory stats",
        "alloc_mb", m.Alloc / 1024 / 1024,
        "total_alloc_mb", m.TotalAlloc / 1024 / 1024,
        "sys_mb", m.Sys / 1024 / 1024,
        "num_gc", m.NumGC,
        "goroutines", runtime.NumGoroutine())
}

// Call periodically
ticker := time.NewTicker(30 * time.Second)
go func() {
    for range ticker.C {
        logMemoryStats()
    }
}()
```

### Goroutine Leak Detection

```bash
# Check for goroutine leaks
go test -run=TestLongRunning ./internal/cli &
PID=$!

# Check goroutines periodically
watch -n 5 "curl http://localhost:6060/debug/pprof/goroutine?debug=1"

# Or in code:
utility.DefaultLogger.Info("Goroutine count",
    "count", runtime.NumGoroutine())
```

### Benchmarking

```go
// Benchmark tree loading
func BenchmarkTreeLoading(b *testing.B) {
    // Setup
    db := setupTestDB()
    rows := loadTestRows()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        treeRoot := model.NewTreeRoot()
        _, err := treeRoot.LoadFromRows(&rows)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

```bash
# Run benchmark
go test -bench=BenchmarkTreeLoading -benchmem ./internal/model

# Output shows:
# BenchmarkTreeLoading-8   1000   1234567 ns/op   524288 B/op   1234 allocs/op
#                          runs   time/op          bytes/op     allocations/op
```

---

## Common Errors and Solutions

### Error: "no root node found"

**Cause:** Content tree has no node with NULL parent_id for the route.

**Solution:**
```sql
-- Check for root node
SELECT * FROM content_data WHERE route_id = 1 AND parent_id IS NULL;

-- Create root if missing
INSERT INTO content_data (route_id, datatype_id, parent_id, author_id)
VALUES (1, 1, NULL, 1);
```

**Prevention:**
```go
// Validate before loading
func validateRouteHasRoot(db db.DbDriver, routeID int64) error {
    rows, err := db.GetContentDataByRoute(ctx, routeID)
    if err != nil {
        return err
    }

    hasRoot := false
    for _, row := range rows {
        if !row.ParentID.Valid {
            hasRoot = true
            break
        }
    }

    if !hasRoot {
        return fmt.Errorf("route %d has no root node", routeID)
    }

    return nil
}
```

### Error: "circular references detected"

**Cause:** Database contains nodes that reference each other as parents.

**Solution:**
```sql
-- Find circular references manually
WITH RECURSIVE chain AS (
    SELECT content_data_id, parent_id, 1 as depth,
           CAST(content_data_id AS TEXT) as path
    FROM content_data
    WHERE content_data_id = 123  -- suspected node

    UNION ALL

    SELECT cd.content_data_id, cd.parent_id, c.depth + 1,
           c.path || ' -> ' || cd.content_data_id
    FROM content_data cd
    JOIN chain c ON cd.content_data_id = c.parent_id
    WHERE c.depth < 100
    AND c.path NOT LIKE '%' || cd.content_data_id || '%'
)
SELECT * FROM chain ORDER BY depth;

-- Fix by breaking the cycle
UPDATE content_data SET parent_id = NULL WHERE content_data_id = 123;
```

### Error: "FOREIGN KEY constraint failed"

**Cause:** Attempting to insert/update with invalid foreign key reference.

**Solution:**
```go
// Verify all foreign keys before insert
func validateContentDataParams(db db.DbDriver, params CreateContentDataParams) error {
    // Check route exists
    if _, err := db.GetRouteByID(ctx, params.RouteID); err != nil {
        return fmt.Errorf("route %d does not exist", params.RouteID)
    }

    // Check datatype exists
    if _, err := db.GetDatatypeByID(ctx, params.DatatypeID); err != nil {
        return fmt.Errorf("datatype %d does not exist", params.DatatypeID)
    }

    // Check parent exists (if specified)
    if params.ParentID.Valid {
        if _, err := db.GetContentDataByID(ctx, params.ParentID.Int64); err != nil {
            return fmt.Errorf("parent %d does not exist", params.ParentID.Int64)
        }
    }

    return nil
}
```

### Error: "EOF" or "connection reset"

**Cause:** Database connection lost or TUI terminal closed unexpectedly.

**Solution:**
```go
// Implement connection retry logic
func executeWithRetry(operation func() error, maxRetries int) error {
    var err error
    for i := 0; i < maxRetries; i++ {
        err = operation()
        if err == nil {
            return nil
        }

        utility.DefaultLogger.Warn("Operation failed, retrying",
            "attempt", i+1,
            "max_retries", maxRetries,
            "error", err)

        time.Sleep(time.Second * time.Duration(i+1))
    }

    return fmt.Errorf("operation failed after %d retries: %w", maxRetries, err)
}
```

### Error: "invalid memory address or nil pointer dereference"

**Cause:** Accessing nil pointer, often in tree operations.

**Solution:**
```go
// Always check for nil before dereferencing
// BAD:
child := node.FirstChild.Instance.ContentDataID  // Crash if FirstChild is nil

// GOOD:
if node.FirstChild != nil {
    child := node.FirstChild.Instance.ContentDataID
}

// Use helper functions
func safeGetContentDataID(node *model.TreeNode) (int64, bool) {
    if node == nil || node.Instance == nil {
        return 0, false
    }
    return node.Instance.ContentDataID, true
}
```

### Error: "too many SQL variables"

**Cause:** SQLite has limit of 999 variables per query.

**Solution:**
```go
// Batch large operations
func batchInsertContentFields(db db.DbDriver, fields []ContentField) error {
    batchSize := 900  // Stay under SQLite's 999 limit

    for i := 0; i < len(fields); i += batchSize {
        end := i + batchSize
        if end > len(fields) {
            end = len(fields)
        }

        batch := fields[i:end]
        utility.DefaultLogger.Debug("Inserting batch",
            "batch_start", i,
            "batch_size", len(batch))

        if err := db.BulkInsertContentFields(ctx, batch); err != nil {
            return err
        }
    }

    return nil
}
```

### Error: "context deadline exceeded"

**Cause:** Operation took longer than context timeout.

**Solution:**
```go
// Increase timeout for long operations
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Or use background context for indefinite operations
ctx := context.Background()

// Add progress logging for long operations
func loadLargeTree(db db.DbDriver, routeID int64) error {
    ctx := context.Background()

    utility.DefaultLogger.Info("Starting large tree load", "route_id", routeID)

    rows, err := db.GetContentTreeByRoute(ctx, routeID)
    if err != nil {
        return err
    }

    utility.DefaultLogger.Info("Rows fetched", "count", len(rows))

    treeRoot := model.NewTreeRoot()
    stats, err := treeRoot.LoadFromRows(&rows)

    utility.DefaultLogger.Info("Tree loaded",
        "nodes", stats.NodesCount,
        "duration", time.Since(start))

    return err
}
```

---

## Debug Mode and Development Tips

### Enable Debug Output

```bash
# Set environment variable
export MODULACMS_DEBUG=true
export MODULACMS_LOG_LEVEL=debug

# Run application
./modulacms-x86 --cli
```

**In code:**
```go
// Check debug mode
if os.Getenv("MODULACMS_DEBUG") == "true" {
    utility.DefaultLogger.SetLevel(log.DebugLevel)
}
```

### Development Database

**Use separate database for debugging:**
```bash
# Copy production database
cp modulacms.db modulacms-debug.db

# Run with debug database
./modulacms-x86 --config=./config-debug.json --cli

# config-debug.json points to modulacms-debug.db
```

### Query Logging

**Enable SQL query logging (driver-specific):**

**SQLite:**
```go
// Add to connection setup
db, err := sql.Open("sqlite3", "file:modulacms.db?cache=shared&_trace=1")
```

**MySQL:**
```sql
-- Enable general query log
SET GLOBAL general_log = 'ON';
SET GLOBAL log_output = 'TABLE';

-- View queries
SELECT * FROM mysql.general_log ORDER BY event_time DESC LIMIT 100;
```

**PostgreSQL:**
```sql
-- Enable query logging
ALTER SYSTEM SET log_statement = 'all';
SELECT pg_reload_conf();

-- View logs
tail -f /var/log/postgresql/postgresql-15-main.log
```

### Test Data Generation

```go
// Generate test content tree
func createTestTree(db db.DbDriver, depth, breadth int) error {
    routeID := int64(1)
    datatypeID := int64(1)
    authorID := int64(1)

    // Create root
    rootID, err := db.CreateContentData(ctx, CreateContentDataParams{
        RouteID:    routeID,
        DatatypeID: datatypeID,
        ParentID:   sql.NullInt64{Valid: false},
        AuthorID:   authorID,
    })
    if err != nil {
        return err
    }

    // Recursively create children
    return createChildren(db, rootID, routeID, datatypeID, authorID, depth-1, breadth)
}

func createChildren(db db.DbDriver, parentID, routeID, datatypeID, authorID int64, depth, breadth int) error {
    if depth <= 0 {
        return nil
    }

    for i := 0; i < breadth; i++ {
        childID, err := db.CreateContentData(ctx, CreateContentDataParams{
            RouteID:    routeID,
            DatatypeID: datatypeID,
            ParentID:   sql.NullInt64{Int64: parentID, Valid: true},
            AuthorID:   authorID,
        })
        if err != nil {
            return err
        }

        // Recurse
        if err := createChildren(db, childID, routeID, datatypeID, authorID, depth-1, breadth); err != nil {
            return err
        }
    }

    return nil
}
```

### Debugging Checklist

When stuck on an issue:

- [ ] Add logging at entry and exit of problematic function
- [ ] Log all input parameters and return values
- [ ] Check for nil pointers before dereferencing
- [ ] Verify database constraints and foreign keys
- [ ] Test query directly in database
- [ ] Check for race conditions (use `go test -race`)
- [ ] Review recent changes with `git diff`
- [ ] Search logs for similar error messages
- [ ] Simplify to minimal reproduction case
- [ ] Add unit test that reproduces the issue
- [ ] Check TREE_STRUCTURE.md and CONTENT_MODEL.md for pitfalls sections

---

## Related Documentation

**Architecture:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/TREE_STRUCTURE.md` - Tree implementation and pitfalls
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/CONTENT_MODEL.md` - Domain model and common issues
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/TUI_ARCHITECTURE.md` - TUI debugging patterns
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/DATABASE_LAYER.md` - Database debugging

**Workflows:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/TESTING.md` - Test strategies
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/ADDING_FEATURES.md` - Development workflow

**Packages:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/packages/CLI_PACKAGE.md` - TUI implementation details
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/packages/DB_PACKAGE.md` - Database operations

**Database:**
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/database/SQL_DIRECTORY.md` - Schema structure
- `/Users/home/Documents/Code/Go_dev/modulacms/ai/database/SQLC.md` - Query generation

**General:**
- `/Users/home/Documents/Code/Go_dev/modulacms/CLAUDE.md` - Development guidelines

---

## Quick Reference

### Essential Debug Commands

```bash
# Run tests with verbose output
go test -v ./internal/model

# Run specific test with debugging
dlv test ./internal/model -- -test.run TestTreeLoading

# Run with race detection
go test -race ./...

# Profile CPU usage
go test -cpuprofile=cpu.prof ./internal/model
go tool pprof cpu.prof

# Profile memory
go test -memprofile=mem.prof ./internal/model
go tool pprof mem.prof

# Enable debug logging
export MODULACMS_DEBUG=true
./modulacms-x86 --cli
```

### Common Log Patterns

```go
// Function entry
utility.DefaultLogger.Debug("Function called",
    "param1", value1,
    "param2", value2)

// Decision point
utility.DefaultLogger.Debug("Decision made",
    "condition", condition,
    "result", result)

// Error with context
utility.DefaultLogger.Error("Operation failed",
    "error", err,
    "context1", value1,
    "context2", value2)

// Performance measurement
start := time.Now()
// ... operation ...
utility.DefaultLogger.Info("Operation completed",
    "duration_ms", time.Since(start).Milliseconds())
```

### Delve Quick Reference

```bash
# Start debugging
dlv test ./internal/model

# Breakpoints
break file.go:123
break FunctionName
clear 1

# Navigation
next        # next line
step        # step into
stepout     # step out
continue    # continue

# Inspection
print var
locals
args
stack

# Conditional
condition 1 var == value
```

### Key Files for Debugging

- `internal/utility/utility.go` - Logger configuration
- `internal/cli/update.go` - TUI message handling
- `internal/cli/cms_struct.go` - Tree operations
- `internal/db/` - Database interface
- `cmd/main.go` - Application entry point
- `debug.log` - Default log output (if configured)

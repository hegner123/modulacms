# Debugging

This guide covers practical debugging techniques for ModulaCMS, organized by the area of the application you are investigating: database operations, TUI state, tree structures, HTTP servers, and performance.

## Logging

ModulaCMS uses structured logging throughout the application. The logger outputs to stderr with timestamps and caller information.

### Log Levels

```go
utility.DefaultLogger.Debug("Processing node", "node_id", nodeID, "parent_id", parentID)
utility.DefaultLogger.Info("Content tree loaded", "route_id", routeID, "nodes", nodeCount)
utility.DefaultLogger.Warn("Orphaned node detected", "node_id", nodeID)
utility.DefaultLogger.Error("Failed to load tree", "error", err, "route_id", routeID)
utility.DefaultLogger.Fatal("Database connection failed", "error", err)
```

Use key-value pairs instead of string formatting:

```go
// Structured (preferred)
utility.DefaultLogger.Info("User authenticated", "user_id", userID, "provider", "github")

// Unstructured (avoid)
utility.DefaultLogger.Info(fmt.Sprintf("User %d authenticated via %s", userID, provider))
```

### Enable Debug Output

Set the environment variable to enable debug-level logging:

```bash
export MODULACMS_DEBUG=true
./modulacms-x86
```

## Debugging Database Issues

### Query Failures

When a query returns unexpected results or fails, log the parameters before execution, then test the same query directly in the database:

```bash
# SQLite
sqlite3 modula.db "SELECT * FROM content_data WHERE route_id = 1"

# MySQL
mysql -u root -p -e "SELECT * FROM content_data WHERE route_id = 1"

# PostgreSQL
psql -U postgres -c "SELECT * FROM content_data WHERE route_id = 1"
```

If the query works in the database but fails in Go, check the sqlc-generated code in the relevant `internal/db-*` package to verify parameter types and return types match.

### Foreign Key Violations

```
Error: FOREIGN KEY constraint failed
```

This means you are referencing a parent record that does not exist. Verify all referenced IDs before the insert:

```sql
-- Check if the referenced route exists
SELECT route_id FROM routes WHERE route_id = 1;

-- Check if the referenced datatype exists
SELECT datatype_id FROM datatypes WHERE datatype_id = 5;
```

To see the foreign key definitions on a table:

```sql
-- SQLite
PRAGMA foreign_key_list(content_data);

-- MySQL
SHOW CREATE TABLE content_data;

-- PostgreSQL
SELECT conname, confrelid::regclass
FROM pg_constraint
WHERE contype = 'f' AND conrelid = 'content_data'::regclass;
```

### NULL Value Issues

SQLite returns `sql.NullString` and `sql.NullInt64` types. Always check `.Valid` before using the value:

```go
if node.Instance.ParentID.Valid {
    parentID := node.Instance.ParentID.Int64
    // use parentID
} else {
    // this is a root node (no parent)
}
```

A common mistake is reading `.Int64` without checking `.Valid` -- the value will be 0 for NULL, which is a valid ID in some contexts.

### Driver-Specific Behavior

A query that works in SQLite may fail in MySQL or PostgreSQL due to:
- Different placeholder syntax (`?` vs `$1, $2`)
- Different NULL handling
- Different type widths (`int64` vs `int32`)

When investigating driver-specific issues, check the sqlc-generated code for each backend to confirm the query syntax is correct.

## Debugging TUI State

### Message Flow

The TUI follows the Elm Architecture: messages trigger state changes in Update, which are reflected in View. When the TUI stops responding or shows incorrect state, add logging to the Update function:

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    utility.DefaultLogger.Debug("Update called",
        "msg_type", fmt.Sprintf("%T", msg),
        "current_mode", m.Mode)

    // handle messages
}
```

### Common TUI Problems

**Screen is blank or frozen:** Check that the SSH session has an active PTY. The TUI requires a terminal -- non-interactive SSH sessions will fail. Also verify that terminal dimensions are captured (Width and Height are non-zero).

**Commands stop executing:** If `Update()` returns `nil` for `tea.Cmd` when it should return a command, the message loop stalls. Every asynchronous operation must return a `tea.Cmd` that eventually produces a `tea.Msg`.

**State inconsistency:** Add validation after state changes to catch cursor out-of-bounds, nil TreeRoot, or other invalid states early.

## Debugging Tree Operations

### Orphaned Nodes

```
WARN Orphaned node detected node_id=789 parent_id=999
```

A node references a `parent_id` that does not exist in the query result. Find orphaned nodes in the database:

```sql
SELECT cd.content_data_id, cd.parent_id
FROM content_data cd
LEFT JOIN content_data parent ON cd.parent_id = parent.content_data_id
WHERE cd.parent_id IS NOT NULL
  AND parent.content_data_id IS NULL;
```

Fix by setting the orphaned node's parent to NULL (making it a root), or by correcting the parent_id to a valid node.

### Circular References

```
ERROR Circular reference detected node_id=123 parent_id=456
```

A node eventually references itself through its parent chain. Find the cycle with a recursive query:

```sql
WITH RECURSIVE chain AS (
    SELECT content_data_id, parent_id, 1 as depth,
           CAST(content_data_id AS TEXT) as path
    FROM content_data
    WHERE content_data_id = 123

    UNION ALL

    SELECT cd.content_data_id, cd.parent_id, c.depth + 1,
           c.path || ' -> ' || cd.content_data_id
    FROM content_data cd
    JOIN chain c ON cd.content_data_id = c.parent_id
    WHERE c.depth < 100
)
SELECT * FROM chain ORDER BY depth;
```

Break the cycle by setting one node's parent_id to NULL:

```sql
UPDATE content_data SET parent_id = NULL WHERE content_data_id = 123;
```

### Missing Root Node

```
Error: no root node found
```

The content tree for a route has no node with a NULL parent_id. Every route needs exactly one root node:

```sql
SELECT * FROM content_data WHERE route_id = 1 AND parent_id IS NULL;
```

## Using Delve

Install the Go debugger:

```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

Debug a specific test:

```bash
dlv test ./internal/model -- -test.run TestTreeLoading
```

Debug the running application:

```bash
dlv debug ./cmd -- serve
```

Useful commands inside Delve:

| Command | Action |
|---------|--------|
| `break file.go:123` | Set breakpoint at line |
| `break Model.Update` | Set breakpoint at function |
| `continue` | Run to next breakpoint |
| `next` | Step over |
| `step` | Step into |
| `stepout` | Step out of function |
| `print var` | Print variable value |
| `locals` | Show local variables |
| `args` | Show function arguments |
| `stack` | Show stack trace |
| `goroutines` | List all goroutines |
| `condition 1 var == value` | Conditional breakpoint |

## Performance Profiling

### CPU Profiling

```bash
go test -cpuprofile=cpu.prof ./internal/model
go tool pprof cpu.prof
```

Inside pprof:

```
(pprof) top10            # Top 10 functions by CPU time
(pprof) list FunctionName # Source with timing annotations
(pprof) web              # Open graphical view in browser
```

### Memory Profiling

```bash
go test -memprofile=mem.prof ./internal/model
go tool pprof mem.prof
```

Check for memory leaks at runtime:

```go
var m runtime.MemStats
runtime.ReadMemStats(&m)
utility.DefaultLogger.Info("Memory stats",
    "alloc_mb", m.Alloc / 1024 / 1024,
    "goroutines", runtime.NumGoroutine())
```

### Goroutine Leak Detection

If goroutine count grows continuously, use `runtime.NumGoroutine()` to track it, or enable the pprof HTTP endpoint:

```go
import _ "net/http/pprof"

go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

Then inspect goroutines at `http://localhost:6060/debug/pprof/goroutine?debug=2`.

## Quick Debugging Checklist

When stuck on an issue:

1. Add logging at entry and exit of the problematic function
2. Log all input parameters and return values
3. Check for nil pointers before dereferencing
4. Verify database constraints and foreign keys
5. Test the query directly in the database
6. Check for race conditions with `go test -race`
7. Review recent changes with `git diff`
8. Simplify to a minimal reproduction case
9. Add a unit test that reproduces the issue

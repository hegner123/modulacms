# TROUBLESHOOTING.md

Common errors, solutions, and workarounds for ModulaCMS development.

**Path:** `/Users/home/Documents/Code/Go_dev/modulacms/ai/reference/TROUBLESHOOTING.md`
**Purpose:** Quick reference for diagnosing and fixing common issues
**Last Updated:** 2026-01-12

---

## How to Use This Guide

This document is organized by error category. Use Ctrl+F / Cmd+F to search for specific error messages.

**Error Categories:**
1. [Build & Compilation Errors](#build--compilation-errors)
2. [Database Connection Issues](#database-connection-issues)
3. [sqlc Generation Failures](#sqlc-generation-failures)
4. [Foreign Key Constraint Violations](#foreign-key-constraint-violations)
5. [TUI Rendering Problems](#tui-rendering-problems)
6. [Tree Loading Errors](#tree-loading-errors)
7. [OAuth & Authentication Errors](#oauth--authentication-errors)
8. [Server Startup Issues](#server-startup-issues)
9. [Performance Issues](#performance-issues)
10. [Runtime Errors](#runtime-errors)

---

## Build & Compilation Errors

### Error: `undefined: sql.NullString` or similar sqlc types

**Symptoms:**
```
./internal/router/users.go:45:2: undefined: sql.NullString
```

**Cause:** sqlc-generated code hasn't been regenerated after schema changes.

**Solution:**
```bash
cd sql
sqlc generate
cd ..
go build ./cmd
```

**Prevention:** Always run `just sqlc` after modifying `.sql` files.

---

### Error: `cgo: C compiler "gcc" not found`

**Symptoms:**
```
# github.com/mattn/go-sqlite3
cgo: C compiler "gcc" not found: exec: "gcc": executable file not found in $PATH
```

**Cause:** SQLite driver requires CGO, which requires a C compiler.

**MacOS Solution:**
```bash
xcode-select --install
```

**Linux Solution:**
```bash
# Ubuntu/Debian
sudo apt-get install build-essential

# RHEL/CentOS
sudo yum groupinstall "Development Tools"
```

**Docker Solution:**
```dockerfile
FROM golang:1.24-alpine
RUN apk add --no-cache gcc musl-dev sqlite-dev
```

**Verification:**
```bash
gcc --version
# Should output version info
```

---

### Error: `cannot find package "github.com/mattn/go-sqlite3"`

**Symptoms:**
```
package github.com/mattn/go-sqlite3: cannot find package
```

**Cause:** Vendor directory is out of sync or missing.

**Solution:**
```bash
go mod vendor
go build ./cmd
```

**Alternative (if vendor is corrupted):**
```bash
rm -rf vendor/
go mod tidy
go mod vendor
go build ./cmd
```

---

### Error: Cross-compilation fails with CGO

**Symptoms:**
```
just build
# Fails on: CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build
```

**Cause:** Cross-compiling CGO code requires platform-specific C compiler.

**Solution (Linux target from Mac):**
```bash
# Install cross-compiler
brew install FiloSottile/musl-cross/musl-cross

# Update Makefile:58
CC=x86_64-linux-musl-gcc CXX=x86_64-linux-musl-g++ \
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
go build -mod vendor -o out/bin/modulacms-amd ./cmd
```

**Alternative: Build on target platform**
```bash
# On Linux server
git pull
go build -mod vendor -o modulacms ./cmd
```

---

### Error: `pattern ./...: directory prefix . does not contain main module`

**Symptoms:**
```
go test ./...
pattern ./...: directory prefix . does not contain main module
```

**Cause:** Running commands from wrong directory or `go.mod` is missing.

**Solution:**
```bash
# Ensure you're in project root
cd /Users/home/Documents/Code/Go_dev/modulacms

# Verify go.mod exists
ls go.mod

# Run tests
go test ./...
```

---

## Database Connection Issues

### Error: `failed to open SQLite database: unable to open database file`

**Symptoms:**
```
ERROR Database connection error failed to open SQLite database: unable to open database file
path=/path/to/modula.db
```

**Cause:** File permissions, directory doesn't exist, or path is invalid.

**Solution:**
```bash
# Check directory exists
mkdir -p $(dirname /path/to/modula.db)

# Check permissions
ls -la /path/to/modula.db
chmod 644 /path/to/modula.db  # If file exists
chmod 755 $(dirname /path/to/modula.db)  # Directory must be executable

# Verify path in config.json
cat config.json | grep db_url
```

**Default Path:** If `db_url` is not set in config, defaults to `./modula.db` in current directory.

---

### Error: `failed to connect to MySQL database: dial tcp: connect: connection refused`

**Symptoms:**
```
ERROR Database ping error failed to connect to MySQL database: dial tcp 127.0.0.1:3306: connect: connection refused
host=localhost:3306
```

**Cause:** MySQL server is not running or host/port is incorrect.

**Solution:**
```bash
# Check if MySQL is running
mysql -u root -p -h localhost -P 3306

# Start MySQL
# MacOS
brew services start mysql

# Linux
sudo systemctl start mysql

# Docker
docker run --name mysql -e MYSQL_ROOT_PASSWORD=password -p 3306:3306 -d mysql:8

# Verify config.json
cat config.json | jq '.db_url, .db_name, .db_username'
```

**Test Connection:**
```bash
mysql -u YOUR_USER -p -h YOUR_HOST -P YOUR_PORT YOUR_DATABASE
```

---

### Error: `Access denied for user 'root'@'localhost'`

**Symptoms:**
```
ERROR Database connection error failed to open MySQL database: Error 1045: Access denied for user 'root'@'localhost'
```

**Cause:** Incorrect username or password in config.json.

**Solution:**
```json
// config.json
{
  "db_driver": "mysql",
  "db_url": "localhost:3306",
  "db_name": "modulacms",
  "db_username": "your_user",
  "db_password": "your_password"
}
```

**Reset MySQL Password:**
```bash
# MySQL 8.0+
mysql -u root
ALTER USER 'root'@'localhost' IDENTIFIED BY 'new_password';
FLUSH PRIVILEGES;
```

---

### Error: `failed to connect to PostgreSQL database: pq: SSL is not enabled`

**Symptoms:**
```
ERROR Database ping error failed to connect to PostgreSQL database: pq: SSL is not enabled on the server
```

**Cause:** PostgreSQL requires SSL but config has `sslmode=disable`.

**Solution (Development - Disable SSL):**
```go
// internal/db/init.go:99
connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
    d.Config.Db_User, d.Config.Db_Password, d.Config.Db_URL, d.Config.Db_Name)
```

**Solution (Production - Enable SSL):**
```go
// Requires SSL certificate
connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=require&sslcert=/path/to/cert.pem&sslkey=/path/to/key.pem",
    d.Config.Db_User, d.Config.Db_Password, d.Config.Db_URL, d.Config.Db_Name)
```

---

### Error: `failed to enable foreign keys: SQL logic error`

**Symptoms:**
```
ERROR Database configuration error failed to enable foreign keys: SQL logic error
```

**Cause:** Attempting to enable foreign keys on non-SQLite database.

**Context:** This error should only occur on SQLite. Check `db_driver` config.

**Solution:**
```bash
# Verify driver in config.json
cat config.json | jq '.db_driver'
# Should be "sqlite", "mysql", or "postgres"

# If incorrect, update:
{
  "db_driver": "sqlite",
  "db_url": "./modula.db"
}
```

**File:** `internal/db/init.go:42-48`

---

## sqlc Generation Failures

### Error: `sqlc: unknown driver "sqlite"`

**Symptoms:**
```
cd sql && sqlc generate
unknown driver "sqlite"
```

**Cause:** Outdated sqlc version or incorrect `sqlc.yml` config.

**Solution:**
```bash
# Upgrade sqlc
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Verify version (requires 1.20.0+)
sqlc version

# Check sqlc.yml has correct engine
cat sqlc.yml | grep engine
# Should be: engine: "sqlite"
```

**File:** `sql/sqlc.yml`

---

### Error: `query parameter references column that is not in the query`

**Symptoms:**
```
sql/mysql/content.sql:15:1: query parameter $1 references column "content_data_id" that is not in the query
```

**Cause:** SQL query doesn't include a parameter referenced in `sqlc` annotation.

**Solution:**
```sql
-- Bad: Missing WHERE clause parameter
-- name: GetContentData :one
SELECT * FROM content_data;

-- Good: Includes parameter
-- name: GetContentData :one
SELECT * FROM content_data WHERE content_data_id = ?;
```

**Verification:**
```bash
cd sql
sqlc generate
# Should complete without errors
```

---

### Error: `column reference "table.column" is ambiguous`

**Symptoms:**
```
sqlc: sql/mysql/content.sql:23: column reference "content_data_id" is ambiguous
```

**Cause:** Multiple tables in JOIN have same column name without table prefix.

**Solution:**
```sql
-- Bad: Ambiguous column
SELECT content_data_id, title
FROM content_data
JOIN content_fields ON content_fields.content_data_id = content_data.content_data_id;

-- Good: Fully qualified columns
SELECT cd.content_data_id, cf.value as title
FROM content_data cd
JOIN content_fields cf ON cf.content_data_id = cd.content_data_id;
```

---

### Error: `sqlc.yml not found`

**Symptoms:**
```
just sqlc
cd ./sql && sqlc generate && echo "generated coded successfully"
sqlc.yml not found
```

**Cause:** Running sqlc from wrong directory.

**Solution:**
```bash
# Must run from sql/ directory
cd sql
sqlc generate
cd ..

# Or use make
just sqlc
```

**File Location:** `/Users/home/Documents/Code/Go_dev/modulacms/sql/sqlc.yml`

---

## Foreign Key Constraint Violations

### Error: `FOREIGN KEY constraint failed`

**Symptoms:**
```
ERROR Failed to create content: FOREIGN KEY constraint failed
```

**Common Causes:**
1. Referenced parent record doesn't exist
2. Parent table is empty
3. Foreign key value is NULL where NOT NULL required
4. Deleting parent with existing children

**Solution 1: Verify parent exists**
```sql
-- Check if datatype exists before creating content
SELECT datatype_id FROM datatypes WHERE datatype_id = 5;
-- If empty, parent doesn't exist
```

**Solution 2: Check foreign key definitions**
```sql
-- SQLite
PRAGMA foreign_key_list(content_data);

-- MySQL
SELECT * FROM information_schema.KEY_COLUMN_USAGE
WHERE TABLE_NAME = 'content_data'
AND REFERENCED_TABLE_NAME IS NOT NULL;

-- PostgreSQL
SELECT * FROM information_schema.table_constraints
WHERE constraint_type = 'FOREIGN KEY'
AND table_name = 'content_data';
```

**Solution 3: Insert parent first**
```go
// Create datatype first
datatype, err := dbc.CreateDatatype(db.CreateDatatypeParams{
    Label: "Page",
    RouteID: 1,
})

// Then create content referencing it
content, err := dbc.CreateContentData(db.CreateContentDataParams{
    DatatypeID: datatype.DatatypeID,
    RouteID: 1,
})
```

---

### Error: `Cannot delete or update a parent row: a foreign key constraint fails`

**Symptoms:**
```
ERROR Cannot delete or update a parent row: a foreign key constraint fails
(`modulacms`.`content_fields`, CONSTRAINT `content_fields_ibfk_1`
FOREIGN KEY (`content_data_id`) REFERENCES `content_data` (`content_data_id`))
```

**Cause:** Attempting to delete parent record with existing child records.

**Solution 1: Delete children first**
```go
// Delete all content_fields first
err := dbc.DeleteContentFieldsByContentData(contentDataID)
if err != nil {
    return err
}

// Then delete content_data
err = dbc.DeleteContentData(contentDataID)
```

**Solution 2: Use CASCADE (if defined in schema)**
```sql
-- Schema should have ON DELETE CASCADE
CREATE TABLE content_fields (
    content_field_id INTEGER PRIMARY KEY,
    content_data_id INTEGER NOT NULL REFERENCES content_data ON DELETE CASCADE
);
```

**Check Existing Constraints:**
```bash
# View schema files
cat sql/schema/17_content_fields/schema.sql
```

---

## TUI Rendering Problems

### Error: TUI screen is blank or frozen

**Symptoms:**
- SSH connection succeeds but screen is blank
- No response to keyboard input
- Terminal appears frozen

**Cause 1: PTY not active**

**Solution:**
```go
// internal/cli/middleware.go
func CliMiddleware(v *bool, c *config.Config) wish.Middleware {
    teaHandler := func(s ssh.Session) *tea.Program {
        pty, _, active := s.Pty()
        if !active {
            wish.Fatalln(s, "no active terminal")
            return nil  // Returns early if PTY not active
        }
        // ...
    }
}
```

**Cause 2: Terminal size not set**

**Solution:**
```go
// Ensure terminal dimensions are captured
m.Width = pty.Window.Width
m.Height = pty.Window.Height

if m.Width == 0 || m.Height == 0 {
    m.Width = 80  // Default fallback
    m.Height = 24
}
```

**Cause 3: Panic in View() or Update()**

**Check Logs:**
```bash
tail -f debug.log
# Look for panic messages or error traces
```

**Debug Pattern:**
```go
func (m Model) View() string {
    defer func() {
        if r := recover(); r != nil {
            utility.DefaultLogger.Error("Panic in View()", "error", r)
        }
    }()

    // Your view logic
}
```

---

### Error: `lipgloss.Render()` produces garbled output

**Symptoms:**
- Text appears corrupted or misaligned
- ANSI escape codes visible in output
- Colors don't render correctly

**Cause:** Terminal doesn't support ANSI colors or wrong TERM setting.

**Solution 1: Check TERM environment**
```bash
echo $TERM
# Should be: xterm-256color or similar

# Set if incorrect
export TERM=xterm-256color
```

**Solution 2: Disable colors in unsupported terminals**
```go
import "github.com/muesli/termenv"

// Detect color support
profile := termenv.ColorProfile()

// Conditional styling
if profile == termenv.Ascii {
    // No color support, use plain text
    return text
} else {
    // Apply lipgloss styles
    return lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(text)
}
```

---

### Error: Bubbletea message loop stops responding

**Symptoms:**
- TUI renders initially but stops updating
- Commands don't execute
- No response to input

**Cause:** Update() returns nil for tea.Cmd when it should return a command.

**Solution:**
```go
// Bad: Forgets to return command
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case FetchDataMsg:
        data, err := fetchData()
        // Oops, no command returned - message loop stalls
    }
    return m, nil
}

// Good: Always return appropriate command
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case FetchDataMsg:
        return m, func() tea.Msg {
            data, err := fetchData()
            if err != nil {
                return ErrorMsg{err}
            }
            return DataLoadedMsg{data}
        }
    }
    return m, nil
}
```

**Debugging:**
```go
// Add logging to Update()
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    utility.DefaultLogger.Debug("Update received message",
        "type", fmt.Sprintf("%T", msg))

    // Your update logic
}
```

---

## Tree Loading Errors

### Error: `Circular reference detected in tree structure`

**Symptoms:**
```
ERROR Circular reference detected in tree structure
node_id=123 parent_id=456 ancestor_chain=[123, 456, 789, 123]
```

**Cause:** Content node references itself or its own descendant as a parent.

**Solution 1: Fix database directly**
```sql
-- Find circular references
WITH RECURSIVE tree AS (
    SELECT content_data_id, parent_id, 1 as depth,
           CAST(content_data_id AS TEXT) as path
    FROM content_data WHERE route_id = 1
    UNION ALL
    SELECT cd.content_data_id, cd.parent_id, t.depth + 1,
           t.path || ',' || CAST(cd.content_data_id AS TEXT)
    FROM content_data cd
    JOIN tree t ON cd.parent_id = t.content_data_id
    WHERE t.depth < 100
      AND INSTR(t.path, CAST(cd.content_data_id AS TEXT)) = 0
)
SELECT * FROM tree WHERE depth > 20;  -- Suspiciously deep

-- Fix by setting parent_id to NULL
UPDATE content_data SET parent_id = NULL WHERE content_data_id = 123;
```

**Solution 2: Use validation before creating**
```go
// internal/cli/validation.go
func ValidateParentNotDescendant(nodeID, newParentID int64, dbc db.DbDriver) error {
    // Walk up tree from newParentID
    currentID := newParentID
    for currentID != 0 {
        if currentID == nodeID {
            return fmt.Errorf("circular reference: node %d cannot be parent of its ancestor", newParentID)
        }

        node, err := dbc.GetContentData(currentID)
        if err != nil {
            return err
        }

        if !node.ParentID.Valid {
            break
        }
        currentID = node.ParentID.Int64
    }
    return nil
}
```

---

### Error: `Orphaned nodes detected during tree load`

**Symptoms:**
```
WARN Orphaned node detected node_id=789 parent_id=999 retry_attempt=1
WARN Orphaned node detected node_id=789 parent_id=999 retry_attempt=2
```

**Cause:** Node references parent_id that doesn't exist in result set.

**Solution 1: Check data consistency**
```sql
-- Find orphaned nodes
SELECT cd.content_data_id, cd.parent_id
FROM content_data cd
LEFT JOIN content_data parent ON cd.parent_id = parent.content_data_id
WHERE cd.parent_id IS NOT NULL
  AND parent.content_data_id IS NULL;

-- Fix by setting parent to NULL
UPDATE content_data SET parent_id = NULL
WHERE content_data_id IN (/* IDs from above query */);
```

**Solution 2: Adjust orphan resolution attempts**
```go
// internal/model/model.go
const maxOrphanAttempts = 10  // Increase if needed

func (page *TreeRoot) resolveOrphans(stats *LoadStats) error {
    for attempt := 1; attempt <= maxOrphanAttempts; attempt++ {
        // Orphan resolution logic
    }
}
```

---

### Error: `Tree load timeout` or very slow loading

**Symptoms:**
- Tree takes > 30 seconds to load
- High CPU usage during load
- Memory usage spikes

**Cause:** Large tree with thousands of nodes loading eagerly.

**Solution: Enable lazy loading**
```go
// Only load root-level nodes initially
func LoadTreeLazy(routeID int64, dbc db.DbDriver) (*model.TreeRoot, error) {
    // Load only first-level children
    rows, err := dbc.GetContentTreeRootLevel(routeID)
    if err != nil {
        return nil, err
    }

    tree := model.BuildTree(rows, datatypes, fields, fieldDefs)
    return tree, nil
}

// Load children on-demand
func (node *TreeNode) LoadChildren(dbc db.DbDriver) error {
    if node.ChildrenLoaded {
        return nil
    }

    children, err := dbc.GetContentDataChildren(node.Instance.ContentDataID)
    if err != nil {
        return err
    }

    // Build child nodes
    node.ChildrenLoaded = true
    return nil
}
```

**Performance Profiling:**
```bash
# Add profiling to tests
go test -cpuprofile=cpu.prof -memprofile=mem.prof ./internal/model

# Analyze
go tool pprof cpu.prof
# Commands: top, list, web
```

---

## OAuth & Authentication Errors

### Error: `Missing code parameter` in OAuth callback

**Symptoms:**
```
HTTP 400: Missing code parameter
```

**Cause:** OAuth provider didn't return authorization code (user cancelled or provider error).

**Solution 1: Check provider configuration**
```json
// config.json
{
  "oauth_client_id": "your-client-id",
  "oauth_client_secret": "your-client-secret",
  "oauth_endpoint": {
    "oauth_auth_url": "https://provider.com/oauth/authorize",
    "oauth_token_url": "https://provider.com/oauth/token"
  }
}
```

**Solution 2: Verify callback URL registered**
- OAuth provider settings must include: `http://localhost:8080/api/v1/auth/oauth/callback`
- Must match exactly (http vs https, trailing slash, etc.)

**Solution 3: Check state parameter (CSRF protection)**
```go
// internal/router/auth.go:57-58 (currently commented out)
state := r.URL.Query().Get("state")
// Validate state against stored value to prevent CSRF
```

---

### Error: `Error exchanging token: oauth2: cannot fetch token`

**Symptoms:**
```
HTTP 500: Error exchanging token: oauth2: cannot fetch token: 400 Bad Request
```

**Cause:** Invalid client credentials or incorrect token endpoint.

**Solution 1: Verify credentials**
```bash
# Test OAuth config manually
curl -X POST "https://provider.com/oauth/token" \
  -d "grant_type=authorization_code" \
  -d "code=AUTHORIZATION_CODE" \
  -d "client_id=YOUR_CLIENT_ID" \
  -d "client_secret=YOUR_CLIENT_SECRET"
```

**Solution 2: Check token URL**
```json
// Common provider token URLs
{
  "github": "https://github.com/login/oauth/access_token",
  "google": "https://oauth2.googleapis.com/token",
  "azure": "https://login.microsoftonline.com/common/oauth2/v2.0/token"
}
```

---

### Error: `sessions don't match` or `session is expired`

**Symptoms:**
```
WARN sessions don't match
HTTP 401: Unauthorized
```

**Cause 1: Cookie tampering or corruption**

**Solution:**
```go
// Verify cookie format
// internal/middleware/cookies.go:29-47
func ReadCookie(c *http.Cookie) (*MiddlewareCookie, error) {
    // Validates cookie structure
    err := c.Valid()
    if err != nil {
        return nil, err
    }

    // Base64 decode
    b, err := base64.StdEncoding.DecodeString(c.Value)
    // ...
}
```

**Cause 2: Session expired**

**Solution: Check expiration in database**
```sql
SELECT session_id, user_id, expires_at,
       datetime('now') as current_time
FROM sessions
WHERE user_id = 123;

-- Update expiration
UPDATE sessions
SET expires_at = datetime('now', '+1 day')
WHERE session_id = 456;
```

**Configure Cookie Duration:**
```json
// config.json
{
  "cookie_duration": "24h"  // 24 hours
}
```

---

## Server Startup Issues

### Error: `bind: address already in use`

**Symptoms:**
```
ERROR Could not start server listen tcp :8080: bind: address already in use
```

**Cause:** Another process is using port 8080, 8443, or SSH port.

**Solution 1: Find and kill process**
```bash
# Find process using port
lsof -i :8080
# Output: PID, process name

# Kill process
kill -9 PID

# Or use killall
killall modulacms-x86
```

**Solution 2: Change port in config**
```json
// config.json
{
  "port": "8081",       // Changed from 8080
  "ssl_port": "8444",   // Changed from 8443
  "ssh_port": "2223"    // Changed from 22
}
```

---

### Error: `Certificate Directory path is invalid`

**Symptoms:**
```
FATAL Certificate Directory path is invalid: stat /path/to/certs: no such file or directory
```

**Cause:** SSL certificate directory doesn't exist or incorrect path.

**Solution:**
```bash
# Create cert directory
mkdir -p /path/to/certs
chmod 755 /path/to/certs

# Update config.json
{
  "cert_dir": "/path/to/certs"
}

# For Let's Encrypt autocert (creates certs automatically)
{
  "cert_dir": "./certs",
  "client_site": "example.com",
  "admin_site": "admin.example.com"
}
```

**Development (Self-Signed):**
```bash
# Generate self-signed cert
openssl req -x509 -newkey rsa:4096 -keyout certs/key.pem -out certs/cert.pem -days 365 -nodes

# Use in config
{
  "cert_dir": "./certs"
}
```

---

### Error: `Could not start server: ssh: no host key available`

**Symptoms:**
```
ERROR Could not start server ssh: no host key available
```

**Cause:** SSH host key not found at `.ssh/id_ed25519`.

**Solution:**
```bash
# Generate SSH host key
mkdir -p .ssh
ssh-keygen -t ed25519 -f .ssh/id_ed25519 -N ""

# Verify
ls -la .ssh/id_ed25519
```

**File:** `cmd/main.go:164-169`

---

## Performance Issues

### Issue: High CPU usage during tree operations

**Symptoms:**
- CPU usage > 80% during tree load
- Application becomes unresponsive
- Multiple goroutines blocked

**Diagnosis:**
```bash
# Enable CPU profiling
go test -cpuprofile=cpu.prof ./internal/model
go tool pprof -http=:8080 cpu.prof

# Or runtime profiling
import _ "net/http/pprof"

// In main.go
go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()

# Access profiling
http://localhost:6060/debug/pprof/
```

**Solution 1: Add indexing**
```sql
-- Index frequently queried columns
CREATE INDEX idx_content_data_route ON content_data(route_id);
CREATE INDEX idx_content_data_parent ON content_data(parent_id);
CREATE INDEX idx_content_data_datatype ON content_data(datatype_id);
```

**Solution 2: Batch queries**
```go
// Bad: N+1 queries
for _, node := range nodes {
    fields, _ := dbc.GetContentFields(node.ContentDataID)
}

// Good: Single query
fields, _ := dbc.ListContentFieldsByRoute(routeID)
// Then filter in memory
```

---

### Issue: Memory leak in long-running server

**Symptoms:**
- Memory usage grows continuously
- Eventually crashes with OOM
- Garbage collection takes longer over time

**Diagnosis:**
```bash
# Memory profiling
go test -memprofile=mem.prof ./...
go tool pprof -alloc_space mem.prof

# Check goroutine leaks
curl http://localhost:6060/debug/pprof/goroutine?debug=2

# Heap dump
curl http://localhost:6060/debug/pprof/heap > heap.prof
go tool pprof heap.prof
```

**Common Causes:**
1. Database connections not closed
2. Goroutines not terminating
3. Large objects retained in memory
4. Growing caches without eviction

**Solution 1: Close connections**
```go
con, _, err := dbc.GetConnection()
if err != nil {
    return err
}
defer con.Close()  // Always defer Close()
```

**Solution 2: Context cancellation**
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := performLongOperation(ctx)
```

**Solution 3: Bounded caches**
```go
// Use LRU cache with eviction
import "github.com/hashicorp/golang-lru"

cache, _ := lru.New(1000)  // Max 1000 items
cache.Add(key, value)
```

---

### Issue: Slow database queries

**Symptoms:**
- API requests take > 5 seconds
- Database CPU usage high
- Query log shows slow queries

**Diagnosis:**
```sql
-- SQLite: Enable query logging
PRAGMA query_only = ON;

-- MySQL: Slow query log
SET GLOBAL slow_query_log = 'ON';
SET GLOBAL long_query_time = 2;  -- Queries > 2 seconds

-- PostgreSQL: log slow queries
ALTER SYSTEM SET log_min_duration_statement = 2000;
SELECT pg_reload_conf();
```

**Solution 1: Add indexes (see SQL_DIRECTORY.md)**
```sql
-- Analyze query plan
EXPLAIN QUERY PLAN
SELECT * FROM content_data WHERE route_id = 1;

-- Look for "SCAN TABLE" (bad) vs "SEARCH TABLE USING INDEX" (good)
```

**Solution 2: Use prepared statements**
```go
// sqlc generates prepared statements automatically
// Verify they're being used:
stmt, err := db.Prepare("SELECT * FROM content_data WHERE route_id = ?")
defer stmt.Close()
```

**Solution 3: Reduce N+1 queries**
```go
// Use JOIN instead of multiple queries
-- name: GetContentWithFields :many
SELECT cd.*, cf.*
FROM content_data cd
LEFT JOIN content_fields cf ON cf.content_data_id = cd.content_data_id
WHERE cd.route_id = ?;
```

---

## Runtime Errors

### Error: `panic: runtime error: invalid memory address or nil pointer dereference`

**Symptoms:**
```
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x...]
```

**Cause:** Accessing field or method on nil pointer.

**Common Locations:**
```go
// Cause 1: Nil database result
user, err := dbc.GetUser(userID)
// user might be nil if not found
fmt.Println(user.Email)  // PANIC if user is nil

// Solution: Check nil
if err != nil || user == nil {
    return fmt.Errorf("user not found")
}
fmt.Println(user.Email)

// Cause 2: Nil sql.Null* types
var title sql.NullString
// If title is not Valid, String is empty but accessing causes issues
fmt.Println(title.String)  // OK but might be empty

// Better: Check Valid
if title.Valid {
    fmt.Println(title.String)
}

// Cause 3: Uninitialized map
var myMap map[string]int
myMap["key"] = 123  // PANIC

// Solution: Initialize with make
myMap := make(map[string]int)
myMap["key"] = 123  // OK
```

**Debugging:**
```bash
# Get full stack trace
GOTRACEBACK=all go run ./cmd

# Or set in code
debug.SetTraceback("all")
```

---

### Error: `panic: send on closed channel`

**Symptoms:**
```
panic: send on closed channel
```

**Cause:** Sending to a channel after it's been closed.

**Solution:**
```go
// Bad: Close then send
close(ch)
ch <- value  // PANIC

// Good: Track channel state
type SafeChannel struct {
    ch     chan Message
    closed bool
    mu     sync.Mutex
}

func (sc *SafeChannel) Send(msg Message) error {
    sc.mu.Lock()
    defer sc.mu.Unlock()

    if sc.closed {
        return fmt.Errorf("channel closed")
    }

    sc.ch <- msg
    return nil
}

func (sc *SafeChannel) Close() {
    sc.mu.Lock()
    defer sc.mu.Unlock()

    if !sc.closed {
        close(sc.ch)
        sc.closed = true
    }
}
```

---

### Error: `fatal error: all goroutines are asleep - deadlock!`

**Symptoms:**
```
fatal error: all goroutines are asleep - deadlock!
```

**Cause:** Goroutines waiting on each other or unbuffered channel with no receiver.

**Common Patterns:**
```go
// Cause 1: Unbuffered channel, no receiver
ch := make(chan int)
ch <- 1  // Blocks forever waiting for receiver
// DEADLOCK

// Solution: Use buffered channel or goroutine
ch := make(chan int, 1)
ch <- 1  // OK

// Or
go func() {
    ch <- 1
}()

// Cause 2: Mutex lock not released
mu.Lock()
// Code that panics or returns early
mu.Lock()  // DEADLOCK

// Solution: Always defer Unlock
mu.Lock()
defer mu.Unlock()
```

---

### Error: `too many open files`

**Symptoms:**
```
accept tcp [::]:8080: accept4: too many open files
```

**Cause:** File descriptor limit reached (unclosed connections, files).

**Solution 1: Increase limit**
```bash
# Check current limit
ulimit -n
# Output: 256 (too low)

# Increase temporarily
ulimit -n 4096

# Increase permanently (macOS)
sudo launchctl limit maxfiles 65536 200000

# Increase permanently (Linux)
# Edit /etc/security/limits.conf
* soft nofile 65536
* hard nofile 65536
```

**Solution 2: Close resources**
```go
// Always close database connections
con, _, err := dbc.GetConnection()
if err != nil {
    return err
}
defer con.Close()

// Close HTTP response bodies
resp, err := http.Get(url)
if err != nil {
    return err
}
defer resp.Body.Close()

// Close files
f, err := os.Open(path)
if err != nil {
    return err
}
defer f.Close()
```

---

## Related Documentation

- **[DEBUGGING.md](../workflows/DEBUGGING.md)** - Debugging strategies and profiling tools
- **[TESTING.md](../workflows/TESTING.md)** - Test-driven debugging approaches
- **[DB_PACKAGE.md](../packages/DB_PACKAGE.md)** - Database error handling patterns
- **[CLI_PACKAGE.md](../packages/CLI_PACKAGE.md)** - TUI debugging techniques
- **[TREE_STRUCTURE.md](../architecture/TREE_STRUCTURE.md)** - Tree algorithm error prevention

---

## Quick Reference

**Most Common Errors:**
1. `cgo: C compiler not found` → Install gcc/clang
2. `failed to open database` → Check path and permissions
3. `FOREIGN KEY constraint failed` → Verify parent exists
4. `sqlc: query parameter references column` → Fix SQL query
5. `bind: address already in use` → Kill existing process or change port
6. `sessions don't match` → Session expired or cookie corrupted
7. `nil pointer dereference` → Check for nil before accessing

**Emergency Commands:**
```bash
# Kill hung process
pkill -9 modulacms

# Reset database (DESTRUCTIVE)
rm modula.db
./modulacms-x86 --install

# Clear test artifacts
rm -rf testdb/*.db backups/*.zip

# Regenerate code
cd sql && sqlc generate && cd ..

# Fresh build
make clean
just dev

# View logs
tail -f debug.log
```

**When All Else Fails:**
1. Check `debug.log` for errors
2. Run with verbose logging: `./modulacms-x86 --verbose`
3. Test with SQLite first (simpler than MySQL/PostgreSQL)
4. Verify configuration with `cat config.json | jq`
5. Check GitHub issues: https://github.com/hegner123/modulacms/issues
6. Ask in discussions with error log attached

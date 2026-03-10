# Observability Instrumentation Plan

Issues: #163, #164, #165, #166, #167, #168, #169
Milestone: Observability

## Existing Infrastructure

All metric constants, types, and recording helpers already exist:

- `internal/utility/metrics.go` — `GlobalMetrics` singleton, `Metrics` struct (counters, gauges, histograms), `MeasureTime()`, `MeasureTimeCtx()`, label-keyed storage, `GetSnapshot()`
- `internal/utility/observability.go` — `ObservabilityClient`, `ObservabilityProvider` interface, `CaptureError()`, `GlobalObservability`, console + Sentry providers
- `internal/plugin/metrics.go` — reference implementation showing recording patterns (`RecordHTTPRequest`, `RecordHookExecution`, `RecordError`, etc.)
- Metric constants defined: `MetricHTTPRequests`, `MetricHTTPDuration`, `MetricHTTPErrors`, `MetricDBQueries`, `MetricDBDuration`, `MetricDBErrors`, `MetricSSHConnections`, `MetricSSHErrors`, `MetricCacheHits`, `MetricCacheMisses`, `MetricActiveConnections`, `MetricMemoryUsage`, `MetricGoroutines`

No instrumentation is wired up yet. The constants and infrastructure exist but nothing records to them.

## Dependency Order

```
#163 HTTP middleware  ─┐
#164 DB layer         ─┤
#165 SSH server       ─┤── all independent, can be done in parallel
#166 Permission cache ─┤
#167 Runtime metrics  ─┘
         │
         ▼
#168 /metrics endpoint ── depends on at least one source producing metrics
         │
         ▼
#169 CaptureError wiring ── independent but best done last (touches many files)
```

Issues #163-167 are independent and can be implemented in any order or in parallel. #168 depends on metrics being produced. #169 is a cross-cutting concern best tackled last.

---

## Issue #163: HTTP Middleware Metrics

**New file:** `internal/middleware/metrics.go`

Create a `MetricsMiddleware` that:
1. Wraps `http.ResponseWriter` to capture the status code (default 200 if `WriteHeader` was never called)
2. Records `time.Now()` before calling `next.ServeHTTP`
3. After the handler returns, records:
   - `utility.GlobalMetrics.Increment(utility.MetricHTTPRequests, Labels{"method": r.Method, "path": pattern, "status": statusStr})`
   - `utility.GlobalMetrics.Timing(utility.MetricHTTPDuration, duration, Labels{"method": r.Method, "path": pattern})`
   - If status >= 400: `utility.GlobalMetrics.Increment(utility.MetricHTTPErrors, Labels{"method": r.Method, "status": statusStr})`

**Path normalization:** Use `r.Pattern` (Go 1.22+ registered pattern) to avoid high-cardinality label explosion from path parameters. Fall back to `r.URL.Path` if `r.Pattern` is empty.

**ResponseWriter wrapper:**

```go
type statusWriter struct {
    http.ResponseWriter
    status int
    written bool
}

func (w *statusWriter) WriteHeader(code int) {
    if !w.written {
        w.status = code
        w.written = true
    }
    w.ResponseWriter.WriteHeader(code)
}
```

**Wire into chain:** Add `MetricsMiddleware()` as position 1 in `DefaultMiddlewareChain` (before RequestID, so it captures full request duration including all middleware).

```go
// internal/middleware/http_chain.go DefaultMiddlewareChain
return Chain(
    MetricsMiddleware(),              // 1. Request metrics (outermost for accurate timing)
    RequestIDMiddleware(),            // 2. Request ID generation
    ClientIPMiddleware(),             // 3. Client IP resolution
    // ... rest unchanged
)
```

**Test:** `internal/middleware/metrics_test.go` — table-driven tests with `httptest.NewRecorder`:
- 200 response increments requests counter, records duration, does not increment errors
- 500 response increments requests and errors counters
- Verify labels contain method and status

---

## Issue #164: Database Layer Metrics

**Approach:** Wrap `*sql.DB` at the `database/sql/driver` level. This intercepts every query that passes through the connection — sqlc-generated code, raw `Exec` calls, PRAGMA statements — without touching any generated code or maintaining a 150-method decorator.

**Scope:** SQLite, MySQL, and PostgreSQL only. `RemoteDriver` has no `*sql.DB` (it's HTTP calls to the Go SDK). Remote operations are measured on the server side; double-counting on the client adds noise.

### Connection Wrapping

**New file:** `internal/db/dbmetrics/dbmetrics.go`

A thin wrapper package that intercepts `QueryContext`, `ExecContext`, and `PrepareContext` on the underlying `database/sql/driver.Conn`. For each intercepted call:

1. Parse the query to extract operation + table (see Query Parsing below)
2. Record start time
3. Execute the real query
4. On return, record:
   - `utility.GlobalMetrics.Increment(utility.MetricDBQueries, Labels{"operation": op, "table": table, "driver": driverName})`
   - `utility.GlobalMetrics.Timing(utility.MetricDBDuration, duration, Labels{"operation": op, "table": table, "driver": driverName})`
   - On error: `utility.GlobalMetrics.Increment(utility.MetricDBErrors, Labels{"operation": op, "table": table, "driver": driverName})`

**Wire in:** Three lines change in `internal/db/init.go`, one per `GetDb()` method. Wrap the `*sql.DB` immediately after `sql.Open` succeeds, before assigning to the struct's `Connection` field:

```go
// In Database.GetDb(), after sql.Open:
db = dbmetrics.Wrap(db, "sqlite")

// In MysqlDatabase.GetDb(), after sql.Open:
db = dbmetrics.Wrap(db, "mysql")

// In PsqlDatabase.GetDb(), after sql.Open:
db = dbmetrics.Wrap(db, "postgres")
```

Also wrap in `OpenPool()` which creates independent pools for plugins:

```go
pool, err := sql.Open(driverName, dsn)
// ...
pool = dbmetrics.Wrap(pool, driverName)
```

### Query Parsing

**New file:** `internal/db/dbmetrics/parse.go`

Pre-destructure every query at recording time into a structured log entry. The parser extracts structured fields from the raw SQL so consumers (metrics endpoint, log aggregators, observability providers) get both the raw query and its parsed components without needing to re-parse at read time.

**Parsed fields:**

```go
type QueryInfo struct {
    Raw       string // full query text
    Operation string // "select", "insert", "update", "delete", "pragma", "create", "alter", "drop", "other"
    Table     string // primary table name, empty if unparseable
    Driver    string // "sqlite", "mysql", "postgres"
}
```

**Parsing strategy:** Token-based, not regex. Read the first non-whitespace token to determine operation, then extract the table name based on operation:

| Operation | Table token position |
|-----------|---------------------|
| `SELECT`  | first token after `FROM` |
| `INSERT`  | first token after `INTO` |
| `UPDATE`  | token immediately after `UPDATE` |
| `DELETE`  | first token after `FROM` |
| `CREATE`/`ALTER`/`DROP` | token after `TABLE`/`INDEX` keyword |
| `PRAGMA`  | the pragma name (e.g., `foreign_keys`) |
| `WITH` (CTE) | follow through to the terminal SELECT/INSERT/UPDATE/DELETE |

Implementation:

```go
func ParseQuery(raw string) QueryInfo {
    // Normalize: lowercase, collapse whitespace
    // Tokenize by splitting on whitespace (no need for a full SQL parser)
    // Switch on first token
    // Extract table name by position rules above
    // Strip schema prefix if present (e.g., "public.content" → "content")
    // Strip backticks/quotes around identifiers
}
```

**Edge cases to handle:**
- Subqueries: only extract the outermost table
- JOINs: only the primary table (FROM target), not joined tables
- CTEs (`WITH ... AS`): follow through to the main statement
- Parameterized identifiers: shouldn't exist (schema identifiers are constructed, not parameterized), but if encountered, return table as "unknown"
- Empty/malformed queries: operation = "other", table = ""

**What NOT to parse:**
- WHERE clauses, column lists, expressions — these are irrelevant for metrics
- No attempt at a full SQL AST — this is first-two-tokens extraction with table-name lookahead

### Structured Logging Integration

When recording a query metric, also emit a structured log entry at Debug level so `slog` aggregators get the parsed form:

```go
utility.DefaultLogger.Debug("db.query",
    "operation", info.Operation,
    "table", info.Table,
    "driver", info.Driver,
    "duration_ms", durationMs,
    "error", err, // nil on success
)
```

This gives log aggregators pre-parsed fields without needing log-side SQL parsing. The raw query is intentionally excluded from the log line (it can contain sensitive data in string literals). It's available in the metrics snapshot for the `/api/metrics` endpoint where access is gated by `config:read`.

### Metrics Labels

The metrics system stores labels per metric key. For DB metrics the label set is:

```go
Labels{
    "operation": info.Operation,  // "select", "insert", etc.
    "table":     info.Table,      // "content", "datatype_fields", etc.
    "driver":    info.Driver,     // "sqlite", "mysql", "postgres"
}
```

**Cardinality estimate:** ~9 operations * ~25 tables * 1 driver (per instance) = ~225 unique keys. Well within the flat-map approach in `Metrics`.

### Test Plan

**`internal/db/dbmetrics/parse_test.go`** — table-driven tests for `ParseQuery`:
- `SELECT * FROM content WHERE id = ?` → operation="select", table="content"
- `INSERT INTO datatype_fields (name) VALUES (?)` → operation="insert", table="datatype_fields"
- `UPDATE users SET name = ? WHERE id = ?` → operation="update", table="users"
- `DELETE FROM sessions WHERE expired_at < ?` → operation="delete", table="sessions"
- `PRAGMA foreign_keys=1` → operation="pragma", table="foreign_keys"
- `WITH cte AS (SELECT ...) SELECT * FROM cte` → operation="select", table="cte"
- `CREATE TABLE IF NOT EXISTS foo (...)` → operation="create", table="foo"
- Empty string → operation="other", table=""
- Backtick-quoted identifiers: `` SELECT * FROM `content` `` → table="content"
- Schema-prefixed: `SELECT * FROM public.content` → table="content"

**`internal/db/dbmetrics/dbmetrics_test.go`** — integration test:
- Open a SQLite in-memory DB wrapped with `dbmetrics.Wrap`
- Execute a CREATE TABLE, INSERT, SELECT, UPDATE, DELETE
- Verify `GlobalMetrics` has the correct counters, timings, and labels for each
- Verify error counter increments on a bad query

### Files Summary

| File | Action |
|------|--------|
| `internal/db/dbmetrics/dbmetrics.go` | New — `Wrap(*sql.DB, driver) *sql.DB`, driver-level interceptor |
| `internal/db/dbmetrics/parse.go` | New — `ParseQuery(raw) QueryInfo`, token-based SQL parser |
| `internal/db/dbmetrics/parse_test.go` | New — table-driven parser tests |
| `internal/db/dbmetrics/dbmetrics_test.go` | New — integration test with wrapped SQLite |
| `internal/db/init.go` | Modify — add `dbmetrics.Wrap()` call in each `GetDb()` + `OpenPool()` |

---

## Issue #165: SSH Connection Metrics

**Files:** `cmd/serve.go`, `internal/tui/model.go`

**Connection counter:** In `cmd/serve.go` where the SSH server middleware is configured, add a middleware or callback that increments `MetricSSHConnections` on new connections:

```go
wish.WithMiddleware(
    sshMetricsMiddleware(),               // new: increment connection counter
    middleware.SSHSessionLoggingMiddleware(cfg),
    middleware.SSHAuthenticationMiddleware(cfg),
    middleware.SSHAuthorizationMiddleware(cfg),
    // ...
)
```

Alternatively, use the existing `SSHSessionLoggingMiddleware` as the instrumentation point since it already runs on every connection.

**Active connections gauge:** Use an `atomic.Int64` or the existing `Gauge` method:
- Increment on session start: `utility.GlobalMetrics.Gauge(utility.MetricActiveConnections, float64(active), nil)`
- Decrement on session end (TUI model cleanup or Wish session close)

**SSH errors:** Instrument the SSH server error path in `cmd/serve.go`:
```go
if sshErr := sshServer.ListenAndServe(); sshErr != nil && !errors.Is(sshErr, ssh.ErrServerClosed) {
    utility.GlobalMetrics.Increment(utility.MetricSSHErrors, utility.Labels{"error_type": "listen"})
    // ... existing error handling
}
```

**New file:** `internal/middleware/ssh_metrics.go` — contains `sshMetricsMiddleware` or the instrumentation functions.

**Test:** Verify counter increments on simulated SSH session lifecycle.

---

## Issue #166: Permission Cache Hit/Miss Metrics

**File:** `internal/middleware/authorization.go`

**Instrumentation point:** `PermissionsForRole` method on `PermissionCache`.

```go
func (pc *PermissionCache) PermissionsForRole(roleID types.RoleID) PermissionSet {
    pc.mu.RLock()
    defer pc.mu.RUnlock()
    ps, ok := pc.cache[roleID]
    if ok {
        utility.GlobalMetrics.Increment(utility.MetricCacheHits, nil)
    } else {
        utility.GlobalMetrics.Increment(utility.MetricCacheMisses, nil)
    }
    return ps
}
```

This is the only change needed — two lines in one method. The `PermissionsForRole` method is called once per authenticated request via `PermissionInjector`.

**Performance note:** `GlobalMetrics.Increment` takes a lock. Since this runs on every authenticated request, verify this doesn't become a bottleneck under load. The lock scope is small (map key concatenation + map write), so it should be fine. If profiling shows contention, switch to `atomic.Int64` counters.

**Test:** Add test case to `internal/middleware/authorization_test.go` — call `PermissionsForRole` with a known role and a missing role, verify hit/miss counters.

---

## Issue #167: Runtime Metrics Collector

**New file:** `internal/utility/runtime_metrics.go`

```go
func StartRuntimeMetricsCollector(ctx context.Context, interval time.Duration) {
    go func() {
        ticker := time.NewTicker(interval)
        defer ticker.Stop()
        for {
            select {
            case <-ticker.C:
                var m runtime.MemStats
                runtime.ReadMemStats(&m)
                GlobalMetrics.Gauge(MetricMemoryUsage, float64(m.HeapAlloc), nil)
                GlobalMetrics.Gauge(MetricGoroutines, float64(runtime.NumGoroutine()), nil)
            case <-ctx.Done():
                return
            }
        }
    }()
}
```

**Wire in:** In `cmd/serve.go`, start the collector when the observability client starts (or unconditionally — these metrics are useful even without an external provider):

```go
utility.StartRuntimeMetricsCollector(ctx, 10*time.Second)
```

**Interval:** 10 seconds is reasonable. `runtime.ReadMemStats` does a stop-the-world pause, but at 10s intervals the overhead is negligible.

**Test:** `internal/utility/runtime_metrics_test.go` — start collector with short interval, wait, verify gauges are non-zero.

---

## Issue #168: /metrics HTTP Endpoint

**File:** `internal/router/mux.go`

Register a new endpoint:

```go
mux.Handle("GET /api/metrics", RequirePermission("config:read")(http.HandlerFunc(metricsHandler)))
```

**Handler** (can be inline in `mux.go` or a new file `internal/router/metrics.go`):

```go
func metricsHandler(w http.ResponseWriter, r *http.Request) {
    snapshot := utility.GlobalMetrics.GetSnapshot()
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(snapshot)
}
```

**Permission:** `config:read` (already exists in bootstrap data). The issue mentions `metrics:read` as an option — if a dedicated permission is preferred, it needs to be added to the bootstrap permissions in `sql/schema/26_role_permissions/`.

**Optional Prometheus format:** Add content negotiation or a `?format=prometheus` query parameter. Prometheus text format is straightforward:

```
# TYPE http_requests counter
http_requests{method="GET",path="/api/content",status="200"} 42
```

This can be deferred to a follow-up if JSON is sufficient initially.

**Test:** `internal/router/metrics_test.go` — hit the endpoint, verify JSON response contains expected metric structure. Verify 403 without admin auth.

---

## Issue #169: CaptureError Wiring

**Approach:** Add `utility.CaptureError(err, context)` calls at error boundaries throughout the codebase. This is a scattershot change touching many files.

**Priority order** (highest impact first):

### 1. HTTP Recovery Middleware (new)
**New file:** `internal/middleware/recovery.go`

```go
func RecoveryMiddleware() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            defer func() {
                if rec := recover(); rec != nil {
                    err := fmt.Errorf("panic: %v", rec)
                    utility.CaptureError(err, map[string]any{
                        "method": r.Method,
                        "path":   r.URL.Path,
                        "stack":  string(debug.Stack()),
                    })
                    http.Error(w, "internal server error", http.StatusInternalServerError)
                }
            }()
            next.ServeHTTP(w, r)
        })
    }
}
```

Add to `DefaultMiddlewareChain` as the outermost middleware (position 0, before MetricsMiddleware).

### 2. Database Connection Failures
In `cmd/serve.go` where database connections are established, wrap errors:

```go
if err != nil {
    utility.CaptureError(err, map[string]any{"driver": cfg.DB_Driver})
    // ... existing error handling
}
```

### 3. SSH Server Errors
Already has a log line in `cmd/serve.go` — add `CaptureError` alongside:

```go
if sshErr := sshServer.ListenAndServe(); sshErr != nil && !errors.Is(sshErr, ssh.ErrServerClosed) {
    utility.CaptureError(sshErr, map[string]any{"component": "ssh"})
    // ... existing error handling
}
```

### 4. Backup/Restore Failures
In `internal/backup/` — wrap top-level error returns with `CaptureError`.

### 5. Media Upload/Processing Failures
In `internal/media/` — wrap upload errors and image processing failures.

### 6. Plugin Errors
Already partially wired via `plugin.RecordError()`. Add `CaptureError` for critical failures (VM creation, pool exhaustion).

**Test:** For recovery middleware: trigger a panic in a test handler, verify `CaptureError` was called (mock or check `GlobalObservability`). For other paths: unit tests verifying error paths call `CaptureError`.

---

## Implementation Order (Recommended)

| Phase | Issues | Work | Parallelizable |
|-------|--------|------|----------------|
| 1 | #163, #166, #167 | HTTP metrics middleware, cache hit/miss, runtime collector | Yes (3 agents) |
| 2 | #164 | DB metrics via *sql.DB wrapping + query parser | Solo |
| 3 | #165 | SSH metrics (small, depends on understanding serve.go flow) | Solo |
| 4 | #168 | /metrics endpoint (needs metrics flowing to be testable) | Solo |
| 5 | #169 | CaptureError wiring (cross-cutting, many small changes) | Solo |

**Estimated new files:** 8-10
**Estimated modified files:** 5-8

## Files Summary

| File | Action | Issue(s) |
|------|--------|----------|
| `internal/middleware/metrics.go` | New | #163 |
| `internal/middleware/metrics_test.go` | New | #163 |
| `internal/middleware/recovery.go` | New | #169 |
| `internal/middleware/recovery_test.go` | New | #169 |
| `internal/middleware/ssh_metrics.go` | New | #165 |
| `internal/middleware/http_chain.go` | Modify — add MetricsMiddleware + RecoveryMiddleware to chain | #163, #169 |
| `internal/middleware/authorization.go` | Modify — add hit/miss counting in PermissionsForRole | #166 |
| `internal/middleware/authorization_test.go` | Modify — add cache metric tests | #166 |
| `internal/db/dbmetrics/dbmetrics.go` | New — `Wrap(*sql.DB, driver)`, driver-level interceptor | #164 |
| `internal/db/dbmetrics/dbmetrics_test.go` | New — integration test with wrapped SQLite | #164 |
| `internal/db/dbmetrics/parse.go` | New — `ParseQuery(raw) QueryInfo`, token-based SQL destructuring | #164 |
| `internal/db/dbmetrics/parse_test.go` | New — table-driven parser tests | #164 |
| `internal/db/init.go` | Modify — add `dbmetrics.Wrap()` in each `GetDb()` + `OpenPool()` | #164 |
| `internal/utility/runtime_metrics.go` | New | #167 |
| `internal/utility/runtime_metrics_test.go` | New | #167 |
| `internal/router/mux.go` | Modify — add /api/metrics route | #168 |
| `internal/router/metrics.go` | New (or inline) | #168 |
| `cmd/serve.go` | Modify — wire runtime collector, SSH metrics, CaptureError calls | #165, #167, #169 |
| `internal/backup/*.go` | Modify — add CaptureError at error boundaries | #169 |
| `internal/media/*.go` | Modify — add CaptureError at error boundaries | #169 |

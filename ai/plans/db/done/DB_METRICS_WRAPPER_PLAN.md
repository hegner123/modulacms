# DB Metrics: Connection-Level Wrapper

## Goal

Instrument all database queries — including sqlc-generated code, transactions, PRAGMAs, and DDL —
by wrapping at the `database/sql/driver` level. The existing `dbmetrics.RecordQueryMetrics` in the
Q* functions covers the query builder path (~15 functions). This plan covers the remaining ~1,471
sqlc call sites (970 non-transactional + 501 transactional) without modifying any of them.

## Current State

- `dbmetrics.RecordQueryMetrics` is called in Q* functions (query_builder.go, query_filtered.go)
- `dbmetrics.ParseQuery` extracts operation + table from raw SQL (token-based, ~720ns/query)
- sqlc-generated code calls `d.Connection` directly via `mdb.New(d.Connection)` or `mdb.New(tx)`
- These 1,471 call sites bypass metrics entirely
- Three `sql.Open()` calls in `init.go`: `"sqlite3"`, `"mysql"`, `"postgres"`
- One `sql.Open()` in `OpenPool()` for plugin isolated pools

## Approach: `database/sql/driver` Wrapper

Register instrumented driver names (`"sqlite3-metrics"`, `"mysql-metrics"`, `"postgres-metrics"`)
that wrap the real drivers. Change the 4 `sql.Open()` calls to use the instrumented names. Zero
changes to sqlc code, wrapper methods, query builder, or plugin code.

### Why driver-level, not DBTX-level

A DBTX wrapper (`mdb.New(instrumentedConn)`) would require changing all 1,471 `mdb.New()` call
sites across 56 files via `repfor`. The driver-level approach changes 4 lines in `init.go` and
catches everything transparently — including transactions, prepared statements, and raw `Exec`
calls that never touch `mdb.New()`.

## Design

### New package: `internal/db/dbmetrics/driver.go`

Four wrapper types, each delegating to the real driver's implementation:

```go
// metricsDriver wraps a database/sql/driver.Driver to record query metrics.
type metricsDriver struct {
    inner      driver.Driver
    driverName string // "sqlite", "mysql", "postgres"
}

// metricsConn wraps driver.Conn with query interception.
type metricsConn struct {
    inner      driver.Conn
    driverName string
}

// metricsTx wraps driver.Tx (no query interception needed — queries go through conn).
type metricsTx struct {
    inner driver.Tx
}

// metricsStmt wraps driver.Stmt to capture the query string at Prepare time
// and record metrics at Exec/Query time.
type metricsStmt struct {
    inner      driver.Stmt
    query      string
    driverName string
}
```

### Interfaces to implement

**`metricsDriver`:**
- `driver.Driver` — `Open(name string) (driver.Conn, error)` → wraps returned conn

**`metricsConn`:**
- `driver.Conn` — `Prepare(query) (driver.Stmt, error)` → wraps returned stmt
- `driver.Conn` — `Close() error` → delegates
- `driver.Conn` — `Begin() (driver.Tx, error)` → wraps returned tx
- `driver.ConnBeginTx` — `BeginTx(ctx, opts) (driver.Tx, error)` → wraps returned tx
- `driver.ExecerContext` — `ExecContext(ctx, query, args) (driver.Result, error)` → time + record + delegate
- `driver.QueryerContext` — `QueryContext(ctx, query, args) (driver.Rows, error)` → time + record + delegate

`ExecerContext` and `QueryerContext` are the fast path — `database/sql` calls these directly when
available (all three drivers support them). This is where most metrics are recorded.

**`metricsStmt`:**
- `driver.Stmt` — `Close()`, `NumInput()` → delegate
- `driver.Stmt` — `Exec(args) (driver.Result, error)` → time + record (using captured query)
- `driver.Stmt` — `Query(args) (driver.Rows, error)` → time + record (using captured query)
- `driver.StmtExecContext` — `ExecContext(ctx, args)` → time + record
- `driver.StmtQueryContext` — `QueryContext(ctx, args)` → time + record

The query string is captured at `Prepare` time and stored on `metricsStmt`. This is correct
because prepared statements reuse the same query.

**`metricsTx`:**
- `driver.Tx` — `Commit() error`, `Rollback() error` → delegate only (no query to record)

Transactions don't need metrics themselves — the queries within a transaction flow through the
same `metricsConn` methods (`ExecerContext`, `QueryerContext`, or via `metricsStmt`).

### Interface forwarding

The underlying driver may implement optional interfaces beyond the minimum (`driver.Pinger`,
`driver.SessionResetter`, `driver.NamedValueChecker`, etc.). The wrapper does NOT need to forward
these — `database/sql` handles the absence gracefully. If profiling later shows that a missing
interface causes performance regression (e.g., `Pinger` for health checks), add forwarding then.

### Registration

```go
// In internal/db/dbmetrics/driver.go:
func Register(driverName string, inner driver.Driver, metricsName string) {
    sql.Register(metricsName, &metricsDriver{inner: inner, driverName: driverName})
}
```

Registration happens once at init time. Each database backend registers its instrumented driver.

### Deduplication with existing Q* metrics

The Q* functions already call `dbmetrics.RecordQueryMetrics`. With the driver-level wrapper, every
query would be double-counted (once at Q* level, once at driver level).

**Resolution:** Remove `RecordQueryMetrics` calls from the Q* functions. The driver-level wrapper
is the single source of truth for all query metrics. This simplifies the Q* functions back to
their pre-instrumentation state and eliminates the `dialectDriver()` helper.

## Implementation Steps

### Step 1: Driver wrapper types (`internal/db/dbmetrics/driver.go`)

- [ ] `metricsDriver` implementing `driver.Driver`
- [ ] `metricsConn` implementing `driver.Conn`, `driver.ConnBeginTx`, `driver.ExecerContext`,
      `driver.QueryerContext`
- [ ] `metricsStmt` implementing `driver.Stmt`, `driver.StmtExecContext`, `driver.StmtQueryContext`
- [ ] `metricsTx` implementing `driver.Tx`
- [ ] `Register(driverName string, inner driver.Driver, metricsName string)` function
- [ ] Each intercepted method: `start := time.Now()`, delegate, `RecordQueryMetrics(query,
      driverName, time.Since(start), err)`

### Step 2: Register instrumented drivers (`internal/db/dbmetrics/register.go`)

- [ ] `func init()` that registers all three:
      - `"sqlite3-metrics"` wrapping the sqlite3 driver
      - `"mysql-metrics"` wrapping the mysql driver
      - `"postgres-metrics"` wrapping the postgres/pgx driver
- [ ] Import the real driver packages for their `init()` side effects and access their `Driver`
      implementations

Note: The real drivers are registered via blank imports (`_ "github.com/mattn/go-sqlite3"`).
To get the `driver.Driver` instance, use `sql.Open` then `db.Driver()`, or instantiate the
driver struct directly (e.g., `&sqlite3.SQLiteDriver{}`). Check which approach the vendored
drivers support.

### Step 3: Wire into init.go (`internal/db/init.go`)

- [ ] Change `sql.Open("sqlite3", d.Src)` → `sql.Open("sqlite3-metrics", d.Src)` (line 40)
- [ ] Change `sql.Open("mysql", dsn)` → `sql.Open("mysql-metrics", dsn)` (line 91)
- [ ] Change `sql.Open("postgres", connStr)` → `sql.Open("postgres-metrics", connStr)` (line 142)
- [ ] Change `sql.Open(driverName, dsn)` in `OpenPool()` → `sql.Open(driverName+"-metrics", dsn)`
      (line 311)
- [ ] Add import for `_ "github.com/hegner123/modulacms/internal/db/dbmetrics"` to trigger
      `init()` registration

### Step 4: Remove Q* function instrumentation (`internal/db/query_builder.go`, `query_filtered.go`)

- [ ] Remove all `dbmetrics.RecordQueryMetrics(...)` calls from Q* functions
- [ ] Remove `time.Now()` / `time.Since(start)` timing code from Q* functions
- [ ] Remove `dialectDriver()` helper function
- [ ] Remove `"time"` and `dbmetrics` imports if no longer used
- [ ] Keep `dbmetrics` package import in `query_filtered.go` only if still needed

### Step 5: Tests

- [ ] `internal/db/dbmetrics/driver_test.go` — integration test:
      - Register a test driver wrapping SQLite
      - Open a connection
      - Run CREATE TABLE, INSERT, SELECT, UPDATE, DELETE
      - Verify metrics snapshot contains entries for each operation with correct labels
      - Verify error metrics on a bad query
      - Verify transaction queries are captured
      - Verify prepared statement queries are captured
- [ ] Verify existing `internal/db/` tests still pass (they now use instrumented driver)
- [ ] Verify existing `internal/plugin/` tests still pass

### Step 6: Verify no double-counting

- [ ] After removing Q* instrumentation (Step 4), run a test that executes a query through
      `QSelect` and verifies exactly 1 `db.queries` counter increment (not 2)

## Files Modified

| File | Changes |
|------|---------|
| `internal/db/dbmetrics/driver.go` | New: driver wrapper types |
| `internal/db/dbmetrics/register.go` | New: `init()` driver registration |
| `internal/db/dbmetrics/driver_test.go` | New: integration tests |
| `internal/db/init.go` | Change 4 `sql.Open()` driver name strings |
| `internal/db/query_builder.go` | Remove `RecordQueryMetrics` calls, timing code, `dialectDriver()` |
| `internal/db/query_filtered.go` | Remove `RecordQueryMetrics` calls, timing code |

## Risks

**Driver interface compatibility:** The three database drivers (mattn/go-sqlite3, go-sql-driver/mysql,
lib/pq or jackc/pgx) may implement different optional interfaces. The wrapper only needs the
mandatory interfaces plus `ExecerContext`/`QueryerContext` for the fast path. Test with all three
backends via `just test-integration-db`.

**Prepared statement query capture:** The query is captured at `Prepare` time. If a prepared
statement is reused across multiple executions, each execution records the same query string.
This is correct behavior — the query is the same, only the args differ.

**Performance:** Each query adds ~1μs overhead (timing + ParseQuery + metric recording). This is
negligible compared to actual I/O. The existing `ParseQuery` benchmarks at ~720ns.

**Registration ordering:** `sql.Register` panics on duplicate names. The `init()` function must
run exactly once. Standard Go init ordering handles this — the `dbmetrics` package's `init()` runs
when imported.

## Not In Scope

- Prometheus text format export (tracked by separate issue if needed)
- Per-query arg logging (security risk — args may contain sensitive data)
- Slow query detection/alerting (can be built on top of the histogram data later)
- Connection pool metrics (`sql.DBStats`) — could be added to the runtime metrics collector

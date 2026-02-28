# Go Backend Review - Audit Package Design (Self-Hosted Context)

**Reviewer:** go-backend-reviewer agent
**Date:** 2026-01-23
**Document:** AUDIT_PACKAGE_REVISED_DESIGN.md
**Context:** Self-hosted CMS (like WordPress, Strapi, Directus), single-tenant per deployment

---

## Assessment

This design doc describes an audit logging system for a CMS using a command pattern with generics. The goal is atomic mutations with audit trail recording.

Given the **self-hosted context**, let me re-evaluate what actually matters.

---

## What Actually Matters for Self-Hosted CMS

### Correctness (Solid)
The transaction wrapping is sound. Mutation + audit in same tx is the right call. No partial writes.

### What Self-Hosters Actually Need

**Strapi/Directus comparison**: Their audit systems are simple append-only logs queried through admin UI. This design is already more sophisticated than both with HLC timestamps and structured change capture.

**Competitive advantages this design provides**:
1. Full before/after snapshots (Strapi only logs "what changed", not old values)
2. HLC for ordering even with clock skew
3. Request/IP tracking for debugging issues

---

## Critical Issues

### 1. NodeID is Over-Engineering for Self-Hosted

```go
type AuditContext struct {
    NodeID    types.NodeID  // Why?
    UserID    types.UserID
    RequestID string
    IP        string
}
```

Single-tenant self-hosted CMS does not need distributed node identification. The deployer runs one instance. If they scale horizontally, they put a load balancer in front and share a database. NodeID adds complexity with zero benefit.

**Recommendation**: Remove NodeID. If future multi-node support is needed, add it then. YAGNI.

### 2. HLC is Premature

Hybrid Logical Clocks solve clock skew in distributed systems. A self-hosted CMS hitting a single database doesn't have this problem. The database's `wall_timestamp` column with `datetime('now')` / `NOW()` is sufficient.

**Recommendation**: Use wall clock only. Remove HLC field from change_events table. If you ship multi-master replication later, add it then.

### 3. Driver-Specific SQL in Go Code

```go
func recordChangeEventTx(ctx context.Context, tx *sql.Tx, p RecordChangeEventParams) error {
    query := `
        INSERT INTO change_events (...)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))
    `
```

This raw SQL in Go code contradicts the project's "use sqlc for all database operations" rule. The comment below acknowledges MySQL/PostgreSQL need different syntax, but the solution (sqlc queries) should be the implementation, not an afterthought.

**Recommendation**: Define `RecordChangeEvent` in sqlc query files from the start. The audited package should accept a `ChangeEventRecorder` interface that each driver implements.

### 4. Interface Has Too Many Methods

```go
type CreateCommand[T any] interface {
    Context() context.Context
    AuditContext() AuditContext
    Connection() *sql.DB
    TableName() string
    Execute(context.Context, DBTX) (T, error)
    GetID(T) string
    Params() any
}
```

Seven methods to implement per command struct. For 23 entities x 3 drivers x 3 operations = **207 command structs**, each with 6-7 methods. The doc admits this: "~5,175 lines" for boilerplate.

This is not idiomatic Go. Idiomatic Go has small interfaces. The stdlib's `io.Reader` has one method. `http.Handler` has one method.

**Alternative**: Reduce to essential interface:

```go
type CreateCommand[T any] interface {
    Execute(ctx context.Context, tx DBTX) (T, error)
    AuditInfo() AuditInfo
}

type AuditInfo struct {
    TableName string
    RecordID  string
    UserID    types.UserID
    RequestID string
    IP        string
    Params    any
}
```

The generic `Create` function can extract everything it needs from `AuditInfo`. Context and connection can be passed directly to `Create`.

---

## Improvements

### 1. Consider Opt-In Auditing Per Entity

Not every table needs audit logging. `sessions` table? Probably not. `content_data`? Definitely. Let deployers configure which tables are audited, either via config or compile-time.

### 2. Retention Should Be Trivial

Self-hosters need a simple way to prune old audit records:

```sql
DELETE FROM change_events WHERE wall_timestamp < datetime('now', '-90 days');
```

Document this. Maybe add a CLI command. Don't build a sophisticated retention policy system.

### 3. Query Interface for Admin UI

The doc focuses on write path but doesn't address read path. Self-hosters will want:
- "Show me changes to this content item"
- "Show me what user X did today"
- "Show me all deletes in the last week"

Add sqlc queries:
```sql
-- name: ListChangeEventsByRecordID :many
SELECT * FROM change_events WHERE record_id = ? ORDER BY wall_timestamp DESC;

-- name: ListChangeEventsByUserID :many
SELECT * FROM change_events WHERE user_id = ? ORDER BY wall_timestamp DESC LIMIT ?;
```

### 4. Sensitive Field Handling (Open Question #1)

The doc asks about `Redact()`. For self-hosted, this is less critical (owner has DB access anyway), but still good hygiene. Don't store password hashes in audit logs.

**Recommendation**: Use `json:"-"` tags consistently on sensitive fields. The `PasswordHash` in `CreateUserParams` already does this. Enforce this pattern.

### 5. Batch Operations (Open Question #2)

Skip for now. Single-record operations cover 95% of use cases. Batch creates complexity (partial failures, per-record events vs single event). Add when there's a concrete need.

### 6. Soft Deletes (Open Question #3)

Use `audited.Update` with status change. Creating `SoftDeleteUserCmd` is overkill. The audit log captures the old status, new status is `archived`. That's enough forensic information.

---

## Testing Notes (Flag Only)

Areas requiring strict test coverage:

1. **Rollback verification**: Confirm both mutation and audit record roll back on any failure
2. **Context cancellation mid-transaction**: Ensure no partial writes
3. **Concurrent updates to same record**: Verify `GetBefore` captures correct state
4. **Empty/nil params handling**: Edge cases in JSON marshaling
5. **Cross-driver consistency**: Same operations produce equivalent audit records across SQLite/MySQL/PostgreSQL

---

## Verdict

The core design (atomic mutation + audit) is correct. The implementation is over-engineered for self-hosted use:

- Remove NodeID and HLC until there's a concrete multi-node use case
- Reduce interface surface area
- Use sqlc for change event recording instead of raw SQL
- Add read-path queries for admin UI

The 5,000+ lines of boilerplate for command structs is a maintenance burden. Consider whether a simpler approach (callback-based or decorator pattern) could achieve the same atomicity with less code.

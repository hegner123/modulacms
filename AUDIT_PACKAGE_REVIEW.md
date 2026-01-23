# Audit Package Implementation Plan Review

**Reviewed**: 2026-01-23
**Reviewer**: go-backend-reviewer agent
**Status**: Critical issues identified

---

## Quick Assessment

The plan aims to wrap every database mutation with atomic audit logging using a command pattern and Go generics. The design is reasonable in its goals but has several issues that will cause production problems if not addressed.

---

## Strengths

**1. Clear separation of concerns.** The `audited` package handles transaction orchestration while entity files define the command structs. This keeps audit logic centralized and testable.

**2. Good use of Go generics.** The `CreateCommand[T any]`, `UpdateCommand[T any, ID any]` interfaces avoid code duplication while maintaining type safety. The interface design is small and focused, which is idiomatic.

**3. Proper error wrapping.** Each error is wrapped with context (`fmt.Errorf("audited.Create[%s]: %w", cmd.TableName(), err)`), which will help operators diagnose issues.

**4. Leveraging existing infrastructure.** The plan correctly identifies that validation happens during JSON unmarshal via the `types.*` package, avoiding duplicate validation.

**5. Clear implementation order.** Starting with User, writing tests, then expanding to other entities is a sensible rollout strategy.

---

## Critical Issues

### Issue 1: Transaction not passed to command execution

**Severity**: Critical - Breaks atomicity guarantee

This is the most serious flaw. The `withTx` function creates a transaction, but the `cmd.Execute()` call does not use that transaction:

```go
err = withTx(ctx, conn, func(tx *sql.Tx) error {
    // 1. Execute the create operation
    created, err := cmd.Execute()  // <-- Uses d.Connection, NOT tx
    // ...
    _, err = d.RecordChangeEvent(...)  // <-- Also uses d.Connection, NOT tx
```

Both `cmd.Execute()` and `d.RecordChangeEvent()` will use the database driver's stored connection (`d.Connection`), not the transaction. This means:

- The mutation and audit event are **not atomic**
- A failure in `RecordChangeEvent` will not roll back the mutation
- The entire atomicity guarantee is broken

**Fix Options:**

**Option A: Pass transaction to Execute**

```go
type CreateCommand[T any] interface {
    TableName() string
    ExecuteTx(tx *sql.Tx) (T, error)  // Transaction-aware
    GetID(T) string
    Params() any
    // Remove Driver() - not needed if using tx directly
    Context() context.Context
    AuditContext() AuditContext
}
```

Then modify the sqlc-generated queries to accept a transaction. sqlc supports this with `DBTX` interface.

**Option B: Use sqlc's DBTX interface**

Your sqlc-generated code likely has a `New(db DBTX)` constructor where `DBTX` is `*sql.DB | *sql.Tx`. Restructure so commands use this:

```go
func Create[T any](cmd CreateCommand[T]) (T, error) {
    // ...
    err = withTx(ctx, conn, func(tx *sql.Tx) error {
        // Create sqlc Queries with transaction
        queries := mdb.New(tx)  // Use tx, not conn

        created, err := cmd.ExecuteWithQueries(queries)
        // ...
        _, err = queries.RecordChangeEvent(...)  // Uses same tx
```

**Option C: Inject the connection dynamically**

Modify the driver structs to support transaction injection, then swap the connection temporarily within the transaction scope. This is messier but requires fewer interface changes.

---

### Issue 2: Before-state capture race condition

**Severity**: Medium

For `Update` and `Delete`, `GetBefore()` is called inside the transaction, which is correct. However, the current implementation has a read/write race:

```go
// Inside withTx:
before, err := cmd.GetBefore()  // Read current state
// ...
if err := cmd.Execute(); err != nil {  // Modify state
```

If the transaction isolation level allows (e.g., READ COMMITTED, which is default for MySQL/PostgreSQL), another transaction could modify the record between `GetBefore()` and `Execute()`. The audit log would then capture stale before-state.

**Fix Options:**

1. Use `SELECT ... FOR UPDATE` in `GetBefore()` to lock the row
2. Set transaction isolation to `SERIALIZABLE` (performance cost)
3. Accept the race condition with documentation (may be acceptable for audit purposes since the mutation will either succeed with current state or fail due to conflict)

For a CMS, option 3 is probably acceptable with proper documentation. But you should be aware of it.

---

### Issue 3: Duplicate transaction helper

**Severity**: High - Code duplication

The plan defines a new `withTx` function in `audited.go`:

```go
func withTx(ctx context.Context, conn *sql.DB, fn func(*sql.Tx) error) error {
    tx, err := conn.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }

    if err := fn(tx); err != nil {
        if rbErr := tx.Rollback(); rbErr != nil {
            return fmt.Errorf("rollback failed: %v (original error: %w)", rbErr, err)
        }
        return err
    }

    if err := tx.Commit(); err != nil {
        return fmt.Errorf("commit transaction: %w", err)
    }
    return nil
}
```

But you already have `internal/db/types/transaction.go` with `WithTransaction` that does the same thing, except it uses `defer tx.Rollback()` which is safer:

```go
func WithTransaction(ctx context.Context, db *sql.DB, fn TxFunc) error {
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }
    defer tx.Rollback() // no-op if already committed

    if err := fn(tx); err != nil {
        return err
    }
    if err := tx.Commit(); err != nil {
        return fmt.Errorf("commit transaction: %w", err)
    }
    return nil
}
```

**Fix**: Use the existing `types.WithTransaction` or `types.WithTransactionResult` instead of duplicating. The `defer tx.Rollback()` pattern is more robust because it handles panics.

---

### Issue 4: Update captures params, not actual new state

**Severity**: Low - Inconsistency

In the `Update` function:

```go
// 3. Serialize new params for audit record
newValues, err := json.Marshal(cmd.Params())
```

This serializes the **update parameters**, not the **resulting state**. If your update only modifies some fields (e.g., just `name`), the audit `new_values` will only contain `{"name": "newvalue"}`, not the full record.

This may be intentional (you can reconstruct full state from old_values + new_values), but it is inconsistent with `Create` which stores the full new record.

**Fix Options:**

1. Document this difference clearly
2. Add a `GetAfter()` call after `Execute()` to capture full new state (extra query cost)
3. Accept the asymmetry if partial new_values is sufficient

---

## Improvements

### Improvement 1: Command boilerplate is excessive

Each entity requires 3 command structs with 3 database-specific factory methods each. For 25 entities, that is 225 factory methods plus 75 command structs. This is significant boilerplate.

**Alternative**: Use a single generic command struct:

```go
type GenericCreateCmd[T any, P any] struct {
    ctx       context.Context
    auditCtx  AuditContext
    db        DbDriver
    tableName string
    params    P
    executor  func(P) (T, error)
    getID     func(T) string
}

func NewGenericCreateCmd[T any, P any](
    ctx context.Context,
    auditCtx AuditContext,
    db DbDriver,
    tableName string,
    params P,
    executor func(P) (T, error),
    getID func(T) string,
) GenericCreateCmd[T, P] {
    return GenericCreateCmd[T, P]{...}
}

// Usage:
cmd := audited.NewGenericCreateCmd(
    ctx, auditCtx, db, "users", params,
    db.CreateUser,
    func(u Users) string { return string(u.UserID) },
)
```

This reduces 225 factory methods to 0 and 75 command structs to 3 generic ones.

---

### Improvement 2: Missing context timeout handling

The plan does not address what happens when the context is cancelled mid-transaction. The current `withTx` will respect cancellation because `BeginTx` takes the context, but you should document expected behavior.

Consider adding a transaction timeout if the caller does not provide one:

```go
func Create[T any](cmd CreateCommand[T]) (T, error) {
    ctx := cmd.Context()
    if _, ok := ctx.Deadline(); !ok {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
        defer cancel()
    }
    // ...
}
```

---

### Improvement 3: Audit context should include request metadata

Your `AuditContext` has `NodeID` and `UserID`. Consider adding:

```go
type AuditContext struct {
    NodeID    types.NodeID
    UserID    types.UserID
    RequestID string  // Trace correlation
    IP        string  // Client IP for security audits
}
```

This helps with debugging and security forensics.

---

## Security Concerns

### 1. JSON serialization of sensitive data

The plan serializes entire entities to JSON for audit logging:

```go
newValues, err := json.Marshal(created)
```

If `Users` contains password hashes, API keys, or other secrets, these will be written to the audit log.

**Fix Options:**

- Strip sensitive fields before serialization
- Use a separate "auditable" struct that excludes secrets
- Encrypt the JSON blob at rest
- Use `json:"-"` tags on sensitive fields

### 2. No rate limiting on audit writes

A malicious or buggy client could flood the system with mutations, filling the change_events table.

**Mitigations:**

- Index on `wall_timestamp` for efficient pruning
- Background job to archive/delete old events
- Rate limiting at the HTTP layer

---

## Testing Strategy Assessment

The testing strategy is adequate but needs expansion.

### Missing test cases

1. **Context cancellation mid-transaction** - What happens if context is cancelled between mutation and audit write?

2. **Concurrent modification** - Two transactions try to update the same record simultaneously. Does the before-state capture handle this correctly?

3. **Large payloads** - What happens when `old_values` or `new_values` exceeds database column limits?

4. **Database driver differences** - SQLite handles transactions differently than MySQL/PostgreSQL. Test all three drivers.

5. **Recovery scenarios** - Application crashes between mutation and audit write. What state is the database in when it restarts?

### Suggested test additions

```go
func TestAuditedCreate_ContextCancelled(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    // Cancel immediately after starting
    go func() {
        time.Sleep(10 * time.Millisecond)
        cancel()
    }()
    // Execute and verify clean rollback
}

func TestAuditedUpdate_ConcurrentModification(t *testing.T) {
    // Create record
    // Start two goroutines that update simultaneously
    // Verify at least one fails with appropriate error
    // Verify audit log is consistent with final state
}

func TestAuditedCreate_LargePayload(t *testing.T) {
    // Create params with very large JSON field
    // Verify graceful failure or successful storage
}

func TestAuditedCreate_AllDrivers(t *testing.T) {
    drivers := []db.DbDriver{sqliteDB, mysqlDB, psqlDB}
    for _, d := range drivers {
        t.Run(d.Name(), func(t *testing.T) {
            // Run standard create test
        })
    }
}
```

---

## Priority of Fixes

| Priority | Issue | Effort |
|----------|-------|--------|
| **Critical** | Fix transaction not being passed to command execution | High |
| **High** | Use existing `WithTransaction` helper instead of duplicating | Low |
| **Medium** | Document or fix before-state race condition | Low |
| **Medium** | Address sensitive data in JSON serialization | Medium |
| **Low** | Reduce boilerplate with generic command structs | Medium |
| **Low** | Add request ID and IP to audit context | Low |

---

## Recommended Approach for Transaction Fix

The cleanest fix is Option B (sqlc's DBTX interface). Here's how it would work:

### Modified Command Interface

```go
type CreateCommand[T any] interface {
    TableName() string
    ExecuteTx(ctx context.Context, tx *sql.Tx) (T, error)
    GetID(T) string
    Params() any
    Connection() *sql.DB
    Context() context.Context
    AuditContext() AuditContext
}
```

### Modified Create Function

```go
func Create[T any](cmd CreateCommand[T]) (T, error) {
    var result T
    ctx := cmd.Context()
    auditCtx := cmd.AuditContext()

    err := types.WithTransaction(ctx, cmd.Connection(), func(tx *sql.Tx) error {
        // Execute within transaction
        created, err := cmd.ExecuteTx(ctx, tx)
        if err != nil {
            return fmt.Errorf("execute: %w", err)
        }
        result = created

        newValues, _ := json.Marshal(created)

        // Record event within same transaction
        queries := mdb.New(tx)
        _, err = queries.RecordChangeEvent(ctx, mdb.RecordChangeEventParams{
            EventID:      types.NewEventID(),
            HlcTimestamp: types.HLCNow(),
            NodeID:       auditCtx.NodeID,
            TableName:    cmd.TableName(),
            RecordID:     cmd.GetID(created),
            Operation:    types.OperationINSERT,
            UserID:       types.NullableUserID{ID: auditCtx.UserID, Valid: auditCtx.UserID != ""},
            NewValues:    types.JSONData(newValues),
        })
        return err
    })

    return result, err
}
```

### Modified Command Implementation

```go
type NewUserCmd struct {
    ctx      context.Context
    auditCtx audited.AuditContext
    conn     *sql.DB
    params   CreateUserParams
}

func (d Database) NewUser(ctx context.Context, auditCtx audited.AuditContext, params CreateUserParams) NewUserCmd {
    return NewUserCmd{ctx: ctx, auditCtx: auditCtx, conn: d.Connection, params: params}
}

func (c NewUserCmd) TableName() string                  { return "users" }
func (c NewUserCmd) Connection() *sql.DB               { return c.conn }
func (c NewUserCmd) Context() context.Context           { return c.ctx }
func (c NewUserCmd) AuditContext() audited.AuditContext { return c.auditCtx }
func (c NewUserCmd) Params() any                        { return c.params }
func (c NewUserCmd) GetID(u Users) string               { return string(u.UserID) }

func (c NewUserCmd) ExecuteTx(ctx context.Context, tx *sql.Tx) (Users, error) {
    queries := mdb.New(tx)
    return queries.CreateUser(ctx, mdb.CreateUserParams{
        // Map params...
    })
}
```

This ensures both the mutation and audit event use the same transaction.

---

## Summary

The plan has a solid foundation but the critical transaction issue must be fixed or the atomicity guarantee is broken. The existing `WithTransaction` helper should be reused. The before-state race condition should be documented even if not fixed. The testing strategy needs concurrent and failure-mode tests.

The command pattern is reasonable but the boilerplate could be reduced with generic command structs. Security considerations around sensitive data serialization should be addressed before production deployment.

# Skeptical Architect Review - Audit Package Design

**Reviewer:** skeptical-architect agent
**Date:** 2026-01-23
**Document:** AUDIT_PACKAGE_REVISED_DESIGN.md

---

## Initial Assessment

**Overall Risk Level: HIGH**

The revised design addresses several issues from the initial review, but introduces new problems while leaving critical gaps unresolved. The design shows good intentions but insufficient rigor for a system that claims to guarantee atomicity. This is the kind of design that works in happy-path testing and fails unpredictably in production.

---

## Critical Concerns

### 1. The `request_id` and `ip` fields do not exist in the schema

**Severity: Critical - Will fail at runtime**

The revised design adds `request_id` and `ip` to `RecordChangeEventParams` and the INSERT statement (lines 141, 148-171), but looking at `sql/schema/0_change_events/schema.sql`, these columns do not exist in the table:

```sql
CREATE TABLE IF NOT EXISTS change_events (
    event_id TEXT PRIMARY KEY CHECK (length(event_id) = 26),
    hlc_timestamp INTEGER NOT NULL,
    wall_timestamp TEXT NOT NULL DEFAULT ...,
    node_id TEXT NOT NULL CHECK (length(node_id) = 26),
    table_name TEXT NOT NULL,
    record_id TEXT NOT NULL CHECK (length(record_id) = 26),
    operation TEXT NOT NULL CHECK (operation IN ('INSERT', 'UPDATE', 'DELETE')),
    action TEXT,
    user_id TEXT CHECK ...
    old_values TEXT,
    new_values TEXT,
    metadata TEXT,
    synced_at TEXT,
    consumed_at TEXT
);
```

No `request_id`, no `ip`. The INSERT will fail with "table change_events has no column named request_id" on the first audited operation. This is a design-implementation mismatch that should have been caught in review.

---

### 2. Raw SQL in change_event.go bypasses sqlc type safety

**Severity: High - Defeats the purpose of sqlc**

The design shows `recordChangeEventTx` using raw SQL (lines 146-173):

```go
query := `
    INSERT INTO change_events (
        event_id, hlc_timestamp, node_id, table_name, record_id,
        operation, user_id, old_values, new_values, request_id, ip,
        wall_timestamp
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))
`
```

Meanwhile, you already have a sqlc-generated `RecordChangeEvent` query (queries.sql line 22-38) that is type-safe. The design acknowledges this later in "Driver-Specific Change Event Recording" but then proposes adding more sqlc queries rather than using the existing ones.

The raw SQL approach:
- Duplicates query maintenance across files
- Loses sqlc's type checking
- Must be manually synchronized if schema changes
- Has already failed (wrong columns)

---

### 3. Driver-specific command structs create maintenance explosion

**Severity: High - Long-term maintenance burden**

The design shows separate command structs for each driver (lines 447-483):

```go
type NewUserCmdMysql struct { ... }
func (c NewUserCmdMysql) Execute(ctx context.Context, tx audited.DBTX) (mdbm.Users, error) { ... }
```

For 25 entities x 3 operations x 3 drivers = 225 command struct variants, plus factory methods. The design estimates ~5,175 lines (line 639).

When you need to change the command interface, you change it 225 times. When you add a new database driver (say, CockroachDB), you add 75 new structs. When you add a new entity, you add 9 structs across 3 files.

This is not "idiomatic Go" - this is code generation territory. The design should either:
1. Accept this is a code generation problem and write a generator
2. Use a single generic implementation with driver abstraction at the Execute level
3. Use the existing `DbDriver` interface more effectively

---

### 4. `SELECT ... FOR UPDATE` is mentioned but not implemented

**Severity: Medium - Known race condition documented but not solved**

The comment at line 255-256:

```go
// Note: For strict consistency, use SELECT ... FOR UPDATE in GetBefore
// (MySQL/PostgreSQL only - SQLite uses database-level locking)
```

But the example `GetBefore` implementations (lines 394-396, 431-432) just call the standard `GetUser`:

```go
func (c UpdateUserCmd) GetBefore(ctx context.Context, tx audited.DBTX) (mdb.Users, error) {
    queries := mdb.New(tx)
    return queries.GetUser(ctx, string(c.params.UserID))
}
```

There is no `GetUserForUpdate` query defined, and the design does not specify how to add one. This means the documented race condition (another transaction modifies the record between GetBefore and Execute) will actually happen.

For a CMS under concurrent editing, this is a real scenario. User A opens editor, User B opens editor, User A saves, User B saves. The audit log for User B's save will show User A's changes as the "before" state... except when it does not, because the GetBefore read happened before User A's commit completed.

---

### 5. `json.Marshal` errors silently ignored

**Severity: Medium - Silent data loss in audit log**

In the Update and Delete functions (lines 261, 269, 306):

```go
oldValues, _ := json.Marshal(before)
// ...
newValues, _ := json.Marshal(cmd.Params())
```

Errors from `json.Marshal` are discarded. If marshaling fails (circular references, types that cannot be marshaled, very large objects), the audit record will contain empty/partial JSON, and the operation will succeed anyway.

This defeats the purpose of audit logging. If we cannot record what changed, should the operation still proceed? The design does not address this question.

---

### 6. No handling for records that do not exist before delete

**Severity: Medium - Partial failure mode**

In `Delete`, if `GetBefore` fails because the record does not exist:

```go
before, err := cmd.GetBefore(ctx, tx)
if err != nil {
    return fmt.Errorf("get before: %w", err)
}
```

The delete is aborted. But what if the record was already deleted by another transaction? The audit log shows nothing, and the caller gets an error even though the desired end state (record gone) is achieved.

Should this be an error? Should it be a no-op? Should it log a "delete of non-existent record attempted"? The design is silent.

---

### 7. 30-second default timeout is arbitrary and potentially dangerous

**Severity: Medium - Operational concern**

Lines 203-207, 247-250, 295-298:

```go
if _, ok := ctx.Deadline(); !ok {
    var cancel context.CancelFunc
    ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
}
```

30 seconds is a long time to hold a transaction lock. If the operation hangs (network issue to S3 for media, slow disk I/O), you are holding database locks for 30 seconds. In SQLite, this can block all other writers. In MySQL/PostgreSQL with default isolation, you are holding row locks.

This timeout should be configurable, shorter by default (5-10 seconds for CRUD operations), and documented as to what happens when it fires mid-transaction.

---

## Questions That Need Answers

1. **How will the schema migration add `request_id` and `ip` columns to existing deployments?** The design assumes these columns exist but does not specify the migration.

2. **What happens when `old_values` or `new_values` exceed TEXT column limits?** SQLite's TEXT can hold ~1GB, but MySQL's TEXT is 64KB, and large content records could exceed this.

3. **How do you test atomicity across all three database drivers?** The testing section shows generic tests but no driver-specific atomicity verification.

4. **What is the performance impact of double-reading on Update/Delete?** Every update requires a GetBefore read. For high-frequency updates, this doubles query load.

5. **How do you handle cascading deletes?** If deleting a Datatype cascades to DatatypeFields and ContentFields, do you log each deletion separately, or one event for the parent with nested data?

6. **What happens when the change_events table becomes large?** I see a `DeleteChangeEventsOlderThan` query, but no automated pruning strategy. Who calls this? When?

7. **How do you handle auditing of operations that span multiple entities?** Creating content with fields is typically one logical operation but multiple database inserts.

---

## What's Actually Good

1. **Using existing `types.WithTransaction` is the right call.** The revised design correctly identifies that the original plan duplicated transaction handling. Using the existing helper is simpler and already handles panics via defer.

2. **The DBTX interface approach.** Using sqlc's generated interface to accept both `*sql.DB` and `*sql.Tx` is the correct pattern. This enables the command to execute within the caller's transaction.

3. **Typed ID fields.** The `types.UserID`, `types.EventID`, etc. provide runtime validation and prevent ID type confusion. This is good defensive programming.

4. **The command pattern itself is reasonable.** Bundling context, params, and audit metadata into a single struct that implements an interface is a clean API design. The call sites are readable.

5. **Context timeout injection.** Adding a default timeout when the caller provides an unbounded context is defensive. The specific value is questionable, but the pattern is correct.

6. **Testing strategy mentions rollback verification.** The tests include checking that failed operations leave no trace. This is critical for atomicity claims.

---

## Verdict

**Can I be convinced this is viable?** Conditionally, yes. But not in its current state.

**What would it take:**

1. **Fix the schema mismatch.** Either add `request_id` and `ip` columns to change_events, or remove them from the code. This is blocking - the code will not run.

2. **Use sqlc for change event recording.** Drop the raw SQL in `recordChangeEventTx`. Add a transaction-compatible query to sqlc or inject the driver's queries into the audited package.

3. **Address the maintenance burden.** Either accept you are writing a code generator, or redesign to use a single generic command struct that takes function arguments for the driver-specific parts.

4. **Implement or explicitly defer SELECT FOR UPDATE.** Either add `GetBeforeForUpdate` queries to sqlc for MySQL/PostgreSQL, or document that the race condition is accepted and what its consequences are.

5. **Handle json.Marshal failures.** Either treat them as errors that abort the transaction, or log them separately and continue with empty JSON.

6. **Make timeout configurable.** Move the 30-second default to a constant or config option, and reduce it to 10 seconds.

7. **Add integration tests for all three drivers.** The current test strategy is abstract. Show concrete tests that verify atomicity on SQLite, MySQL, and PostgreSQL using actual database instances.

The design shows evidence of iterative improvement (the review documents exist, issues were identified and addressed). But this iteration introduced new problems while solving old ones. One more pass with focus on the implementation details, not just the architecture, is needed before this is production-ready.

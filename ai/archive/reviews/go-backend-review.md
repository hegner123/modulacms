# Go Backend Review - Audit Package Design

**Reviewer:** go-backend-reviewer agent
**Date:** 2026-01-23
**Document:** AUDIT_PACKAGE_REVISED_DESIGN.md

---

## Assessment

This is a well-considered design for adding audit logging to entity mutations. The command pattern with generics is idiomatic and the atomic transaction guarantee addresses the core requirement. The call sites shown in the HTTP handler examples are clean and would be straightforward to implement across the codebase.

---

## Critical Issues

### 1. Ignored marshal errors in Update and Delete

Lines 261, 269, 306: `oldValues, _ := json.Marshal(before)` discards errors. While `json.Marshal` rarely fails on struct types, ignoring it creates a silent failure mode. If a custom `MarshalJSON` on a type returns an error (as your typed fields do), you will record empty JSON to the audit log with no indication of why.

```go
// Fix: propagate the error
oldValues, err := json.Marshal(before)
if err != nil {
    return fmt.Errorf("marshal before state: %w", err)
}
```

### 2. HLC generation inside transaction creates ordering issues

`types.HLCNow()` is called inside the transaction (lines 227, 275, 316) but the actual commit happens later. Under concurrent load, event A could get HLC timestamp T1, event B gets T2 (where T2 > T1), but B commits before A. The HLC ordering will not match the actual database commit order.

**Options:**
- Generate HLC after successful mutation execution but before audit insert (acceptable)
- Generate HLC at commit time via database trigger (most accurate but adds complexity)
- Document this as a known limitation and use wall_timestamp for strict ordering queries

For most audit use cases this is acceptable, but document it.

### 3. Raw SQL in change_event.go bypasses sqlc type safety

Line 147-172: The hardcoded INSERT bypasses the entire sqlc abstraction. The comment on lines 176-178 acknowledges this but the proposed solution (driver-specific queries) is incomplete. The current code only works for SQLite.

**Concrete fix:** The `DBTX` interface is passed to `Execute`, but `recordChangeEventTx` receives `*sql.Tx` directly. Either:
- Add `RecordChangeEvent` to each command interface and let implementations call the correct sqlc-generated method
- Pass a `ChangeEventRecorder` interface to the generic functions

---

## Improvements

### 4. Context timeout as default behavior

Lines 202-207, 247-250, 296-299: Adding a 30-second timeout when the context has no deadline is reasonable, but 30 seconds is long for a database operation. Consider 5-10 seconds. Also, this should probably be configurable per-deployment, not hardcoded.

### 5. Command struct duplication across drivers

The design acknowledges ~5,175 lines for 23 entities across 3 drivers. Most of this is mechanical. Consider code generation (go generate) using templates. The pattern is repetitive enough that a simple text/template would eliminate the maintenance burden without introducing runtime complexity.

### 6. GetBefore race condition note

Line 255-256 comments on `SELECT ... FOR UPDATE`. This is correct but incomplete. For SQLite, the comment says "database-level locking" but that depends on the journal mode. In WAL mode, readers don't block writers. If strict before-state consistency matters, document the behavior differences per driver.

### 7. AuditContext validation

`AuditContext` has no validation. An empty `NodeID` will be written to the database. Consider:
```go
func (a AuditContext) Validate() error {
    if a.NodeID.IsZero() {
        return errors.New("audit context: NodeID required")
    }
    return nil
}
```
Call this at the start of Create/Update/Delete.

### 8. Connection exposure

The `Connection() *sql.DB` method on command interfaces exposes the raw connection. This works, but now every command struct stores and surfaces the connection. If you later want to support connection pooling or read replicas, you will need to change every command struct. Consider passing the connection to the generic functions instead:

```go
func Create[T any](conn *sql.DB, cmd CreateCommand[T]) (T, error)
```

This removes `Connection()` from the interface and gives you flexibility.

### 9. Error wrapping could be more specific

`fmt.Errorf("execute: %w", err)` is minimal. For debugging at 3 AM, include the table name and operation:

```go
return fmt.Errorf("audited create %s: execute: %w", cmd.TableName(), err)
```

---

## Testing Notes

**Areas requiring strict test coverage:**

1. **Transaction rollback paths** - Verify both entity and change_event are rolled back when either fails. The `TestAuditedCreate_Rollback` test is a good start; also test failure in `recordChangeEventTx` specifically.

2. **Concurrent mutations** - Test two goroutines updating the same entity. Verify change events have correct before/after states even under contention. This will surface any races in HLC generation or transaction isolation issues.

3. **Context cancellation mid-transaction** - Cancel the context after `GetBefore` but before `Execute`. Verify clean rollback.

4. **Driver parity tests** - The table-driven `TestAuditedOperations_AllDrivers` approach is good. Ensure all three drivers behave identically, especially around NULL handling in `user_id` and JSON serialization in `old_values`/`new_values`.

5. **Type validation during unmarshal** - Test that invalid `types.Email` in JSON request body fails early with a useful error, not deep in the transaction.

---

## Strengths

- The generic interface design is genuinely idiomatic. This will look familiar to any Go developer who has worked with the stdlib or popular libraries.
- Keeping command structs in entity files alongside existing wrapper code maintains cohesion.
- The `audited.Ctx()` shorthand is a nice ergonomic touch that will encourage adoption.
- Separating the audited operations as an additive layer (design goal 5) reduces migration risk.

---

## Answers to Open Questions

1. **Sensitive field stripping**: Yes. Add `AuditParams() any` to the interface that returns a redacted copy. For most entities this returns `c.params`; for User it zeros `PasswordHash`. This is safer than relying on `json:"-"` which only works for JSON encode, not if someone logs the struct.

2. **Batch operations**: Defer this. The current design can be extended with `BatchCreateCommand[T]` later. Get single-record operations working first.

3. **Soft deletes**: Use `audited.Update` with the existing update command. A soft delete is semantically an update. Creating `SoftDelete` adds interface surface area without adding capability.

4. **SELECT FOR UPDATE**: Add it as an optional method on `UpdateCommand`/`DeleteCommand`:
   ```go
   type LockingCommand interface {
       UseLocking() bool
   }
   ```
   Check for this interface in `Update`/`Delete` and use the appropriate query variant. SQLite implementations return `false`.

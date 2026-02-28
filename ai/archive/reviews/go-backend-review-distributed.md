# Go Backend Review - Audit Package Design (Distributed-Ready Context)

**Reviewer:** go-backend-reviewer agent
**Date:** 2026-01-23
**Document:** AUDIT_PACKAGE_REVISED_DESIGN.md
**Context:** Distributed-ready from day one, learning enterprise patterns, not an MVP

---

## Assessment

The design implements an audited operations package using a command pattern with generics. The goal is atomic mutation+audit recording with HLC timestamps and NodeID for distributed systems readiness. The architecture is sound for the stated goals.

---

## Critical Issues

### 1. HLC Implementation Has Race Condition on Counter Overflow (types_hlc.go:36-45)

```go
if hlcCounter == 0 {
    time.Sleep(time.Millisecond)
    hlcMu.Unlock()        // releases lock
    result := HLCNow()    // recursive call acquires lock
    hlcMu.Lock()          // re-acquires lock
    return result
}
```

The unlock/relock sequence creates a window where another goroutine can issue an HLC with the same physical time. The counter overflow path needs to sleep while holding the lock, or use a different approach entirely (e.g., increment physical time by 1ms artificially rather than sleeping).

### 2. Driver-Specific SQL in Generic Package (change_event.go:146-173)

`recordChangeEventTx` uses SQLite-specific SQL (`datetime('now')` and `?` placeholders) but the audited package is meant to be driver-agnostic. This breaks MySQL (`NOW()`) and PostgreSQL (`NOW()` with `$1, $2...` placeholders).

The design doc mentions this (lines 595-628) but the proposed solution of using sqlc-generated code requires passing the recorder as part of the command or making the `audited` package aware of which driver is in use. Neither is shown concretely.

**Fix options:**
- Add a `ChangeEventRecorder` interface to the command interface, letting each driver provide its implementation
- Pass the recorder function to `Create/Update/Delete` as a parameter
- Make `recordChangeEventTx` a method on the Database types, passed via command

### 3. `HLCUpdate` Has Semantic Issues for Distributed Use (types_hlc.go:52-75)

The function is intended for receiving events from other nodes, but:
- It updates the global `hlcLast` state, which means receiving a single event with a future timestamp permanently advances your local clock
- No protection against malicious/buggy nodes sending far-future timestamps (clock drift amplification)
- Should probably cap the physical part to `max(local_time, received_time + reasonable_drift)`

For a production distributed system, you need drift bounds. Without them, one node with a bad clock poisons all nodes.

---

## Improvements

### 1. AuditContext Should Include Session/Correlation Context

```go
type AuditContext struct {
    NodeID    types.NodeID
    UserID    types.UserID
    RequestID string
    IP        string
}
```

Consider adding:
- `SessionID` - useful for audit trails
- `TenantID` - if multi-tenancy is planned
- `TraceID`/`SpanID` - for OpenTelemetry integration

These are easy to add later, but the design should acknowledge them as future extension points.

### 2. Command Interface Passes `*sql.DB` for Transaction, but Gets `*sql.Tx` in Execute

The design correctly has `Execute(context.Context, DBTX)` accepting the transaction interface. However, `Connection() *sql.DB` returning the raw connection pool is needed only to start the transaction in `audited.Create/Update/Delete`.

This is fine, but consider whether the command should instead receive a `TxStarter` interface:

```go
type TxStarter interface {
    BeginTx(context.Context, *sql.TxOptions) (*sql.Tx, error)
}
```

This would allow testing without a real `*sql.DB`.

### 3. Update Stores Params, Not Final State in NewValues

```go
// 3. Serialize update params (captures "what changed", not full new state)
newValues, _ := json.Marshal(cmd.Params())
```

This is a deliberate design choice per the comment. For auditing this is fine, but for replication you typically want the full new state so the replica can apply it directly. If you plan to use change_events for replication, consider storing both the "what changed" params AND the full post-update state.

### 4. Missing: Idempotency Keys for Distributed Replay

For true distributed readiness, events should be deduplicated during replay. The current design uses `EventID` (ULID) which is unique per generation, but if a node crashes and replays, it would generate new EventIDs.

Consider adding an `IdempotencyKey` to `ChangeEvent` computed from `(NodeID, TableName, RecordID, Operation, HLC)` or passed from the caller.

### 5. Missing: Version/Sequence Number on Entities

For conflict detection in distributed scenarios (optimistic concurrency), entities typically need a version column that's checked during updates:

```sql
UPDATE users SET ... WHERE user_id = ? AND version = ?
```

The design doesn't address this. HLC provides ordering but not conflict detection. If two nodes update the same record, HLC tells you which happened "later" but not whether the update was based on stale data.

### 6. No Mechanism for Replaying Events

`GetEventsSince` and `GetUnsyncedEvents` exist in the `EventLogger` interface, but there's no corresponding `ApplyEvent` or replay mechanism shown. For distributed replication you'll need:

- Event applier that can reconstruct state from events
- Conflict resolution when applying events out-of-order
- Tombstone handling for deletes

---

## What's Missing for True Distributed Readiness

1. **Node Registration/Discovery**: How do nodes discover each other? How is NodeID assigned and persisted?

2. **Consensus on Operations**: If two nodes create the same username, who wins? The design has `ConflictPolicy` enum (`lww`, `manual`) but no implementation.

3. **Vector Clocks vs HLC**: HLC is a good choice for single-writer scenarios. For multi-writer with conflict detection, you may need vector clocks or version vectors. HLC alone can't detect concurrent modifications (it only orders them).

4. **Event Compaction**: Change events will grow forever. Need a compaction strategy (snapshot + truncate events before snapshot).

5. **Partition Tolerance**: What happens when a node is offline for extended periods? Events accumulate. Replication catches up. But what if the event log is pruned before the node returns?

6. **Schema Evolution**: How do you handle replaying events if the schema has changed? Old events may have fields that no longer exist.

---

## Is the Command Pattern Right?

Yes. The command pattern is appropriate here because:

1. It bundles all inputs needed for an operation into a single object
2. It enables the generic `Create[T]`, `Update[T]`, `Delete[T]` functions to work across all entity types
3. It keeps the audit logic centralized rather than scattered across handlers
4. It's testable - commands can be mocked or inspected

The boilerplate is real (~75 lines per entity across 3 drivers = ~5k lines for 23 entities), but it's straightforward boilerplate. A code generator could produce it.

---

## Testing Notes

Areas requiring strict test coverage:

1. **HLC counter overflow path** - race condition testing with `-race` flag
2. **Transaction rollback** - verify neither mutation nor event exists after failure
3. **Cross-driver parity** - same inputs should produce equivalent change_events across SQLite/MySQL/PostgreSQL
4. **Event ordering** - verify HLC ordering matches insert order under concurrent load
5. **`HLCUpdate` bounds** - test behavior with extreme timestamp values

---

## Strengths

- Atomic mutation+audit using transactions is correctly implemented
- HLC format (48-bit physical + 16-bit logical) is standard and well-chosen
- Command pattern provides clean call sites (`audited.Create(cmd)`)
- Typed params with validation (`types.Email`, `types.Slug`) reduce validation code at call sites
- ULID-based IDs provide time-ordering while remaining globally unique
- Separation of concerns: sqlc handles queries, wrapper types handle abstraction, audit package handles atomicity
- The existing `types` package already has all the primitives needed

---

## Summary

The design is technically sound for single-node operation with future distribution aspirations. The HLC implementation needs the race condition fixed. The driver-specific SQL in `recordChangeEventTx` needs resolution before implementation.

For true distributed operation, you'll need to add: conflict resolution implementation, node discovery, event replay/apply, and version/sequence numbers on entities. These can be added incrementally without changing the core audit pattern.

The command pattern is the right abstraction. It will scale from mom-and-pop to enterprise without architectural changes - you'll be adding implementation details, not redesigning the foundation.

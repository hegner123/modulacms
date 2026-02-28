# Go Backend Review - Audit Package Design (Distributed-Ready) - Round 2

**Reviewer:** go-backend-reviewer agent
**Date:** 2026-01-23
**Document:** AUDIT_PACKAGE_REVISED_DESIGN.md
**Context:** Distributed-ready from day one, learning enterprise patterns, not an MVP

---

## Assessment

This design implements a command pattern for audited database operations with distributed-system primitives (NodeID, HLC) baked in from day one. The goal is atomic mutation+audit logging that can later support multi-node replication.

---

## Critical Issues

### 1. HLC Counter Overflow Creates Deadlock Risk

`/Users/home/Documents/Code/Go_dev/modulacms/internal/db/types/types_hlc.go` lines 36-45:

```go
if hlcCounter == 0 {
    // Counter overflow, wait for wall clock to advance
    time.Sleep(time.Millisecond)
    hlcMu.Unlock()
    result := HLCNow()
    hlcMu.Lock()
    return result
}
```

The unlock/relock dance here is a race condition. Another goroutine can grab the lock between `Unlock()` and `Lock()`, corrupt state, and the function returns with `hlcMu` locked but having done nothing useful after the recursive call. The deferred `Unlock()` at the function start will double-unlock.

Fix: Remove the recursive call. Spin-wait or use a condition variable. Better: track the last physical timestamp and increment counter correctly without recursion.

### 2. HLCUpdate Physical Extraction Bug

`/Users/home/Documents/Code/Go_dev/modulacms/internal/db/types/types_hlc.go` lines 60-74:

```go
if received > maxPhysical {
    maxPhysical = received & ^HLC(0xFFFF) // Extract physical part
}
```

This compares `received` (full HLC) against `maxPhysical` (physical part only after first iteration). The comparison should extract physical from `received` first:

```go
receivedPhysical := received & ^HLC(0xFFFF)
if receivedPhysical > maxPhysical {
    maxPhysical = receivedPhysical
}
```

Without this fix, HLCUpdate will incorrectly advance clocks when receiving events with lower physical time but high counters.

### 3. Raw SQL in Transaction Bypasses sqlc

`/Users/home/Documents/Code/Go_dev/modulacms/AUDIT_PACKAGE_REVISED_DESIGN.md` lines 146-174 shows `recordChangeEventTx` using raw SQL. The design doc acknowledges this but the "use sqlc-generated code" section (lines 597-627) should be the primary approach, not an afterthought.

Raw SQL here means:
- No compile-time type checking
- Manual column/parameter sync required
- Divergence risk between drivers

Since sqlc queries can accept `DBTX` (which includes `*sql.Tx`), generate `RecordChangeEventTx` via sqlc and call it within the transaction.

### 4. Context Leak in Create/Update/Delete

`/Users/home/Documents/Code/Go_dev/modulacms/AUDIT_PACKAGE_REVISED_DESIGN.md` lines 198-239:

```go
func Create[T any](cmd CreateCommand[T]) (T, error) {
    var result T
    ctx := cmd.Context()

    if _, ok := ctx.Deadline(); !ok {
        var cancel context.CancelFunc
        ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
        defer cancel()
    }
```

The derived `ctx` shadows the original but is only used inside this function. The `cancel()` is correctly deferred. However, if the caller's context is already cancelled, you check for deadline but not cancellation. Add:

```go
if ctx.Err() != nil {
    return result, ctx.Err()
}
```

---

## Improvements

### 1. NodeID Lifecycle Missing

The design assumes `h.nodeID` exists in handlers but never specifies:
- When NodeID is generated (application startup vs first request)
- Where it's persisted (config file, database, environment)
- Whether it survives restarts

For distributed semantics, NodeID must be stable across restarts. Generate once at install time, store in config/database.

### 2. HLC Needs Persistence for Crash Recovery

`HLCNow()` uses in-memory state. If the process crashes and restarts within the same millisecond, you could generate duplicate HLC values. Options:

1. Persist `hlcLast` periodically (every N events or every M seconds)
2. On startup, read last persisted HLC and add a safety margin
3. Accept millisecond-level uniqueness is sufficient (likely fine for CMS)

### 3. Change Event Schema Missing Version Field

`/Users/home/Documents/Code/Go_dev/modulacms/internal/db/types/types_change_events.go`:

For future schema evolution (adding fields to change events), include a `version` field. Replication consumers need to know how to parse old vs new event formats.

### 4. Update Captures Params, Not Full After-State

`/Users/home/Documents/Code/Go_dev/modulacms/AUDIT_PACKAGE_REVISED_DESIGN.md` lines 267-269:

```go
// 3. Serialize update params (captures "what changed", not full new state)
newValues, _ := json.Marshal(cmd.Params())
```

The comment acknowledges this. For replication, you typically want full before/after states to replay events. The params-only approach loses computed fields, triggers, and default values. Consider fetching the row again after Execute to capture actual new state.

### 5. Driver Abstraction for recordChangeEventTx

The design mentions driver-specific SQL but doesn't show how `audited.Create` knows which driver to use. Options:

1. Pass a `ChangeEventRecorder` interface to the audited functions
2. Have command structs implement a `RecordEvent(ctx, tx, params)` method
3. Use sqlc-generated code that accepts DBTX

Option 3 is cleanest and already suggested in the doc.

### 6. Batch Operations Design Gap

Open question 2 asks about batch operations. For distributed systems, batch operations need:
- Single HLC for ordering (all events in batch have same timestamp)
- OR sequential HLCs with same RequestID for correlation
- Atomicity guarantee (all or nothing)

The current design handles single-record operations. Batch support should use the same transaction with multiple change event inserts.

---

## What's Missing for True Distribution

### 1. Conflict Resolution

The schema has `ConflictPolicy` enum but no implementation in the audit package. When two nodes modify the same record, you need:
- Detection (compare HLC + NodeID on merge)
- Resolution (LWW is simple, manual requires UI)
- Storage for conflict state

### 2. Replication Protocol

The `EventLogger` interface exists but no:
- Push mechanism (how do events get to other nodes?)
- Pull mechanism (how does a node catch up?)
- Ordering guarantees (causal consistency via HLC)
- Deduplication (events may arrive multiple times)

### 3. Node Registry

For multi-node, you need to know which nodes exist. Currently NodeID is just an identifier with no registration or discovery.

### 4. Vector Clocks vs HLC

HLC gives total ordering but doesn't capture causality between nodes. If Node A and Node B both modify different records, HLC can't tell you they're concurrent (vs one-before-other). For a CMS, this is likely fine since most conflicts are on the same record.

### 5. Compaction / Retention

Change events will grow unbounded. Need:
- Retention policy (keep last N days)
- Compaction (merge old events into snapshots)
- Archival (move to cold storage)

---

## Testing Notes

Areas requiring strict test coverage:

1. **HLC monotonicity under concurrent load** - multiple goroutines calling HLCNow() simultaneously must never get duplicate or out-of-order values
2. **Transaction rollback** - verify both entity and change_event are rolled back on failure at any step
3. **Context cancellation mid-transaction** - ensure proper cleanup
4. **Cross-driver consistency** - same operations on SQLite/MySQL/PostgreSQL produce semantically identical change events
5. **HLCUpdate with various received values** - edge cases around counter overflow, physical time in future, etc.

---

## Strengths

1. **Interface-based command pattern** is idiomatic Go and matches stdlib patterns like `http.Handler`
2. **Atomic audit logging** via single transaction is the correct approach
3. **HLC choice** is appropriate for eventual consistency across nodes
4. **ULID for EventID** gives sortable, unique identifiers with embedded timestamps
5. **Separation of concerns** - audited package handles orchestration, entity files handle entity-specific logic
6. **Typed validation on unmarshal** catches malformed input early

---

## Summary

The foundation is solid for single-node audit logging with distribution-ready primitives. The HLC implementation has two bugs that need fixing before production use. The raw SQL for change events should be replaced with sqlc-generated code. NodeID persistence and HLC crash recovery are the main gaps before this can safely run multi-node.

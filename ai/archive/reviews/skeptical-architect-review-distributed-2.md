# Skeptical Architect Review - Audit Package Design (Distributed-Ready) - Round 2

**Reviewer:** skeptical-architect agent
**Date:** 2026-01-23
**Document:** AUDIT_PACKAGE_REVISED_DESIGN.md
**Context:** Distributed-ready from day one, learning enterprise patterns, not an MVP

---

## Initial Assessment

**Risk Level: MEDIUM-HIGH for distributed operation**

The design is solid for single-node auditing. However, there are significant gaps that will require non-trivial work when moving to true multi-node distributed operation. Some of these gaps would benefit from being addressed now while there is no migration burden.

---

## Critical Concerns

### 1. HLC Implementation Has a Fatal Flaw for Multi-Node

The HLC implementation in `types_hlc.go` (lines 18-75) uses process-global state:

```go
var (
    hlcMu      sync.Mutex
    hlcLast    HLC
    hlcCounter uint16
)
```

This works for single-node. But for multi-node:

- `HLCUpdate()` is intended to merge received HLCs from other nodes, but there is no mechanism to call it. The audit package design never invokes `HLCUpdate()`. Every call is `HLCNow()`.
- When Node A receives a change event from Node B with HLC 1000, Node A should update its local HLC to at least 1001 before recording any new events. This is the whole point of HLCs.
- Without this receive-side update, your HLCs are just fancy timestamps. You lose the causal ordering guarantee that HLCs exist to provide.

### 2. No NodeID Persistence or Registration

`NewNodeID()` generates a ULID every time it is called. This is wrong for distributed systems.

- Where does a node discover its NodeID? From config? At startup? Generated once and persisted?
- What happens on node restart? Does it get a new NodeID? If so, you lose the ability to track which events came from which physical node over time.
- There is no node registry. In a multi-node deployment, nodes need to know about each other for:
  - Replication targeting
  - Quorum decisions
  - Leader election (if applicable)
  - Health checking

The `AuditContext` receives `nodeID` from the handler, but how does the handler get it? This is undefined.

### 3. Missing Conflict Detection in Change Events

The schema has `ConflictPolicy` enum (`lww`, `manual`) in `types_enums.go`, but:

- The change_events table has no `version` or `vector_clock` column
- There is no mechanism to detect that two nodes modified the same record concurrently
- LWW (Last Write Wins) requires a comparison basis. HLC comparison alone is not conflict detection; it is ordering after the fact.

When Node A and Node B both update User 123 at nearly the same time:
- Who wins?
- How do you know there was a conflict?
- Where is the conflict recorded?

### 4. Replication Architecture is Undefined

The `EventLogger` interface has `GetUnsyncedEvents()` and `MarkSynced()`, which implies a replication model, but:

- Push or pull? Does the origin push events to replicas, or do replicas pull from origin?
- Sync target tracking: `synced_at` is a single timestamp. What if there are 5 replica nodes and event E is synced to 3 of them? You cannot track partial sync.
- No replication cursor per destination. Each node needs to track "last HLC I sent to Node B" separately from "last HLC I sent to Node C".

### 5. No Idempotency Keys for Event Application

When applying replicated events, what prevents double-application?

- Node B receives event E from Node A
- Network hiccup, Node B does not acknowledge
- Node A retries sending event E
- Node B receives E again

The `event_id` being a ULID is globally unique, but there is no index or constraint preventing duplicate application of the same event's effect. You need:
- A `replicated_events` or `applied_events` tracking table
- Or idempotent apply logic (which requires knowing the current state)

### 6. SELECT FOR UPDATE Comment is Necessary, Not Optional

Line 255-256 in the design doc:
```go
// Note: For strict consistency, use SELECT ... FOR UPDATE in GetBefore
// (MySQL/PostgreSQL only - SQLite uses database-level locking)
```

This is not optional for distributed operation. Without row-level locking on MySQL/PostgreSQL:
1. Two concurrent updates read the same "before" state
2. Both apply their changes
3. One overwrites the other
4. The audit log shows the wrong "before" for the second update

This is a data integrity bug waiting to happen.

### 7. No Partition Key Strategy

At scale, change_events will grow unbounded. The current indexes are:
- `idx_events_record` (table_name, record_id)
- `idx_events_hlc` (hlc_timestamp)
- `idx_events_node` (node_id)
- `idx_events_user` (user_id)

Missing:
- No time-based partitioning strategy
- No archival strategy beyond "delete old synced+consumed events"
- No consideration for how to query across partitions

When you have 100M events, these indexes will still work but get slow. Planning for sharding/partitioning now is easier than retrofitting.

---

## Questions That Need Answers

1. **NodeID Lifecycle**: When is NodeID assigned? On first boot? From config file? From a coordination service? What happens if two nodes accidentally get the same NodeID (misconfiguration)?

2. **Replication Topology**: Star? Mesh? Single leader? Multi-leader? Each has different failure modes and conflict scenarios.

3. **Ordering Guarantees**: Are you aiming for causal consistency? Eventual consistency? Strong consistency within a datacenter?

4. **Schema Changes**: When Node A has schema v2 and Node B has schema v1, what happens to replicated events that reference v2 columns?

5. **Recovery from Divergence**: If two nodes process writes in isolation for an hour (network partition), then reconnect, what is the merge strategy?

6. **Tombstones**: The DELETE operation stores `old_values`. How long do tombstones persist? Can you "resurrect" a deleted record from another node's out-of-order INSERT?

---

## What is Actually Good

1. **HLC in the change_events table is the right call**. Many systems skip this and regret it. The implementation just needs the receive-side update wiring.

2. **Separating synced_at from consumed_at**. This correctly distinguishes replication (synced) from downstream effects (consumed/webhooks).

3. **ULID for EventID**. Time-ordered, globally unique, no coordination required. Exactly right.

4. **The command pattern abstraction**. Clean, testable, and keeps the transaction boundary explicit.

5. **Driver-specific SQL for timestamps**. Acknowledging the datetime function differences across SQLite/MySQL/PostgreSQL upfront avoids runtime surprises.

6. **Recording old_values for UPDATE and DELETE**. Essential for audit replay and conflict resolution.

---

## What Would Break First Under Distributed Load

1. **Clock skew between nodes**. If Node A's clock is 5 seconds ahead of Node B, HLCs will be consistently higher from A. Combined with LWW, A always wins conflicts even if B was "actually" first.

2. **Replication lag under write storms**. GetUnsyncedEvents with a simple LIMIT will fall behind if writes exceed sync throughput. No backpressure mechanism.

3. **The global HLC mutex** (`hlcMu`). Under high concurrency, this becomes a bottleneck. Consider per-goroutine HLC generation with periodic merge, or sharded locks.

4. **JSON serialization of old_values/new_values**. At scale, this becomes expensive. Consider storing diffs rather than full states for UPDATE.

---

## Enterprise Patterns Missing That Should Be Added Now

### 1. VectorClock or Version column on change_events
- Enables true conflict detection
- Formula: `{nodeID: HLC, nodeID2: HLC, ...}`
- Allows detecting concurrent writes (neither happened-before the other)

### 2. Replication Cursor Table
```sql
CREATE TABLE replication_cursors (
    source_node_id TEXT NOT NULL,
    destination_node_id TEXT NOT NULL,
    last_synced_hlc INTEGER NOT NULL,
    last_synced_at TEXT NOT NULL,
    PRIMARY KEY (source_node_id, destination_node_id)
);
```

### 3. Node Registry Table
```sql
CREATE TABLE nodes (
    node_id TEXT PRIMARY KEY,
    hostname TEXT NOT NULL,
    region TEXT,
    role TEXT CHECK (role IN ('leader', 'follower', 'read-replica')),
    registered_at TEXT NOT NULL,
    last_heartbeat TEXT,
    status TEXT CHECK (status IN ('active', 'draining', 'offline'))
);
```

### 4. Conflict Log Table
```sql
CREATE TABLE conflicts (
    conflict_id TEXT PRIMARY KEY,
    table_name TEXT NOT NULL,
    record_id TEXT NOT NULL,
    event_a_id TEXT NOT NULL,
    event_b_id TEXT NOT NULL,
    resolution TEXT CHECK (resolution IN ('lww', 'manual', 'merged', 'pending')),
    resolved_at TEXT,
    resolved_by TEXT
);
```

### 5. Sequence Numbers per Record
- Add `version INTEGER NOT NULL DEFAULT 1` to each entity table
- Increment on update
- Enables optimistic concurrency control
- Enables detecting "I updated version 5, but current is 7, what happened?"

---

## Verdict

Can I be convinced this is viable for distributed operation? **Yes, with work.**

What would it take:

1. **Add HLCUpdate() call path** when receiving events from other nodes. This is architectural, not just code. Where does event reception happen? gRPC? HTTP? Message queue?

2. **Define NodeID lifecycle** in config or persistence layer. Add node registration table.

3. **Add version column** to entities or vector clock to change_events. Pick one conflict detection strategy and implement it.

4. **Add replication_cursors table** to track per-destination sync progress.

5. **Implement SELECT FOR UPDATE** in GetBefore for MySQL/PostgreSQL. This is mandatory, not optional.

6. **Add partition/archive strategy** for change_events. Even just documenting "we will partition by month" is better than nothing.

The current design is a solid foundation for auditing. It is 60% of the way to distributed replication. The missing 40% is not boilerplate; it is the hard parts that differentiate "works in the demo" from "works at scale."

# Skeptical Architect Review - Audit Package Design (Distributed-Ready Context)

**Reviewer:** skeptical-architect agent
**Date:** 2026-01-23
**Document:** AUDIT_PACKAGE_REVISED_DESIGN.md
**Context:** Distributed-ready from day one, learning enterprise patterns, not an MVP

---

## Initial Assessment

**Overall Risk Level: Medium-High**

The design document shows solid understanding of single-node transactional audit logging. However, for the stated goal of "distributed-ready architecture from day one," there are significant gaps that will require rewrites when scaling to multi-node deployments. The HLC implementation has a fundamental flaw, the NodeID scheme lacks essential infrastructure, and several enterprise patterns required for distributed operation are entirely absent.

---

## Critical Concerns

### 1. HLC Implementation Bug: Deadlock/Recursive Lock on Counter Overflow

**Severity: Critical**

In `/Users/home/Documents/Code/Go_dev/modulacms/internal/db/types/types_hlc.go`, lines 36-44:

```go
hlcCounter++
if hlcCounter == 0 {
    // Counter overflow, wait for wall clock to advance
    time.Sleep(time.Millisecond)
    hlcMu.Unlock()        // DANGER: manual unlock
    result := HLCNow()    // Recursive call
    hlcMu.Lock()          // Re-acquire
    return result
}
```

This code has two problems:

1. **Lock ordering violation**: The mutex is manually unlocked, then re-acquired after recursive call. If another goroutine grabs the lock during that window, the re-lock will succeed but the code continues with stale local state assumptions.

2. **Counter overflow is unlikely but catastrophic**: 65,536 operations in a single millisecond would overflow. Under extreme load (batch imports, bulk operations), this is reachable. The recursive unlock/relock pattern is fragile.

**What breaks at scale**: High-throughput batch operations hitting counter overflow will exhibit race conditions, potential HLC ordering violations, and data corruption in the change_events log.

### 2. HLC Has No Node Identification - Cannot Guarantee Global Ordering

**Severity: Critical**

The HLC format `(wall_time_ms << 16) | logical_counter` contains 48 bits of wall time and 16 bits of counter. There is **no node identifier embedded in the HLC itself**.

In a distributed system, two nodes can generate identical HLC values:

- Node A at timestamp 1706100000000ms, counter 5: `HLC = 0x0001_8D66_B2D4_0005`
- Node B at timestamp 1706100000000ms, counter 5: `HLC = 0x0001_8D66_B2D4_0005`

These are identical. When events from both nodes are merged, the ordering is ambiguous.

**What the design document assumes**: HLC alone provides total ordering. It does not. You need `(HLC, NodeID)` pairs for global ordering, or embed NodeID into HLC (common approaches: 64-bit HLC with 48-bit time, 8-bit node, 8-bit counter; or 96-bit/128-bit HLC).

**What breaks at scale**: Event replication will have ordering ambiguities. Conflict resolution cannot reliably determine "last write" across nodes. The entire distributed consistency model is compromised.

### 3. NodeID Is Generated Per-Process, Not Persisted or Registered

**Severity: High**

`NodeID` is a ULID generated via `NewNodeID()`. Looking at the config structure, there is **no NodeID field**. The audit design shows:

```go
auditCtx := audited.Ctx(h.nodeID, getUserID(r), getRequestID(r), getIP(r))
```

But `h.nodeID` is not defined anywhere in configuration. This means either:
- NodeID is generated at startup (changes every restart, breaking event correlation)
- NodeID is intended to be added but is missing

**What breaks at scale**:
- Node restarts generate new NodeIDs. Events cannot be grouped by originating node across restarts.
- No node registry exists. When doing distributed sync, how does Node B know about Node A? How do you enumerate nodes for replication?
- Stale nodes cannot be detected (no heartbeat, no node table).

### 4. No Vector Clock or Causal Dependency Tracking

**Severity: High**

The design uses HLC timestamps but lacks causal dependency information. In a distributed system, you need to know not just "when" but "what did this node see when it made this write."

For CMS content with relations (e.g., content references media, content has parent/child tree structure), writes can depend on other writes. Current design cannot express:

- "This update saw events up to HLC X from Node A and HLC Y from Node B"
- "This delete depends on the existence of child records being deleted first"

**What breaks at scale**:
- Content tree operations (your sibling-pointer tree) will have ordering issues during concurrent distributed edits
- Cannot implement causal consistency guarantees
- Conflict resolution lacks context about what state the writer observed

### 5. No Idempotency Keys for Distributed Replay

**Severity: High**

The change_events table uses `event_id` (ULID) as primary key. Events are recorded in transactions. But for distributed replication, you need idempotency - applying the same event twice must be safe.

Current design has no mechanism for:
- Detecting duplicate event delivery
- Marking events as "replayed from Node X"
- Tracking which events each node has seen (high-water marks)

**What breaks at scale**: Event replay (for recovery, new node bootstrap, or replication retry) will create duplicate events or fail on primary key conflict.

### 6. ConflictPolicy Exists But No Conflict Detection Infrastructure

**Severity: High**

The types package defines `ConflictPolicy` (lww, manual) but:
- No `conflict_records` table for flagged conflicts
- No mechanism to detect concurrent writes to the same record from different nodes
- No UI or API for manual conflict resolution
- LWW is implemented implicitly by HLC ordering, but HLC ordering across nodes is broken (see concern #2)

**What breaks at scale**: Concurrent edits from different nodes will silently overwrite each other (if LWW) or silently corrupt data (if no detection). Manual conflict policy cannot work without detection and storage.

### 7. RecordID Constraint Breaks Cross-Node ID Generation

**Severity: Medium**

The schema constraint `CHECK (length(record_id) = 26)` assumes ULID format. This is fine. However, all record IDs are generated client-side with `NewXxxID()` functions using process-local entropy.

While ULIDs are designed to be globally unique, the entropy source `ulid.Monotonic(rand.Reader, 0)` is per-process. In extremely high-throughput scenarios across many nodes, collision probability increases slightly.

More importantly, for content created on disconnected nodes (offline editing scenario), ULID collision risk is non-zero over long time horizons.

**What breaks at scale**: Rare ID collisions during offline/reconnect scenarios or extreme throughput will cause primary key violations during sync.

### 8. No Event Compaction or Snapshotting Strategy

**Severity: Medium**

The `DeleteChangeEventsOlderThan` query exists but:
- No snapshot mechanism to rebuild state from a known point
- No way to compact multiple updates to the same record into a single "current state" snapshot
- No way to bootstrap a new node without replaying full event history

**What breaks at scale**: Event log grows unbounded. New node bootstrap becomes slower over time. Recovery time increases with system age.

---

## Questions That Need Answers

1. **How is NodeID provisioned?** Is it configuration, auto-generated on first run and persisted, or assigned by a coordinator?

2. **What is the replication topology?** Leader-follower? Multi-leader? Leaderless? The design is silent on this.

3. **How does a new node discover existing nodes?** Static configuration? Service discovery? Gossip protocol?

4. **What happens during network partition?** Can nodes continue accepting writes? How are conflicting writes reconciled?

5. **What is the consistency model?** Eventual consistency? Causal consistency? Strong consistency? This drives every distributed design decision.

6. **How is the "before state" captured for replication?** Currently it is the local node's view. In distributed systems, this might be stale relative to other nodes' views.

7. **What is the sync protocol?** Push? Pull? Both? Event streaming? Batch polling?

---

## What Is Actually Good

1. **The change_events schema is reasonable**. Table structure, indexes on HLC and node_id, and the synced_at/consumed_at tracking are correct patterns for event-sourcing foundations.

2. **Typed ID system is solid**. ULID-based IDs with proper validation, JSON marshaling, and SQL scanning are production-grade.

3. **Transaction helper with defer pattern** in `types.WithTransaction` is correct and handles panics properly.

4. **EventLogger interface** is well-designed with the right operations for a single-node event store that could be extended for replication.

5. **The command pattern with generics** is idiomatic Go and provides clean abstractions for audited operations.

6. **Separation of Operation/Action enums** allows distinguishing database-level operations from business-level actions, which is useful for different consumers (replication vs webhooks vs audit UI).

---

## Verdict

**Can I be convinced this is viable? Yes, but only for single-node deployment.**

The current design is solid for a single-node CMS with audit logging. For the stated goal of distributed-ready architecture, it is **not viable without addressing concerns 1-6**.

---

## Must Fix Now (Before Any Production Use)

1. **Fix HLC overflow bug** - Replace recursive unlock pattern with proper handling
2. **Add NodeID to configuration** - Persist across restarts, document provisioning
3. **Define composite ordering** - Document that global ordering is `(HLC, NodeID)`, update comparisons

## Must Fix Before Multi-Node (Zero Migration Cost Now)

4. **Add origin_node_id and sequence_num columns** to change_events for proper event ordering
5. **Add nodes table** - Track registered nodes, last seen timestamp, HLC high-water marks
6. **Add conflicts table** - Store detected conflicts for manual resolution
7. **Add version/HLC column to content tables** - Enable optimistic locking for concurrent updates

## Should Add for Enterprise Scale

8. **Event snapshots/compaction** - Periodic state snapshots, ability to prune old events
9. **Idempotency tokens** - For safe event replay during sync
10. **Causal metadata** - Track what state was observed when making writes

---

## Enterprise Patterns Missing That Should Be Added Now

While complexity is intentional, these patterns are significantly cheaper to add without migration burden:

### 1. Optimistic Locking Column
Add `version` (or use HLC) to content tables. Check version on update, reject if stale. This is a single column addition per table.

### 2. Node Registration Table
Even if unused initially, having `nodes(node_id, hostname, last_heartbeat, hlc_high_water_mark)` ready means sync logic has somewhere to store state.

### 3. Conflict Queue Table
`conflicts(conflict_id, table_name, record_id, local_event_id, remote_event_id, detected_at, resolved_at, resolution)` - empty until needed, but schema is in place.

### 4. Causality Tracking Fields
Add `depends_on_event_ids TEXT[]` to change_events (JSON array of event IDs this event depends on). Can be null initially but schema supports it.

### 5. Event Sequence Per Node
Add `node_sequence INTEGER` to change_events - a monotonically increasing counter per node, separate from HLC. Enables gap detection in replication ("I have events 1-500 from Node A, but not 501").

---

## Final Assessment

The design shows good instincts but conflates "has distributed primitives" with "ready for distributed operation." The primitives (HLC, NodeID, change_events) are present but not wired together into a coherent distributed model. The hardest part is not the boilerplate - it is the protocol design that the boilerplate would implement. That protocol is not specified.

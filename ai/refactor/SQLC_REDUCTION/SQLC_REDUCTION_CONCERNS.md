# SQLC Type Unification - Critical Review and Concerns

**Reviewed:** 2026-01-22
**Context:** Greenfield project, zero active users, no production data
**Overall Risk Level:** MEDIUM

---

## Executive Summary

The SQLC Type Unification plan is a legitimate refactor with real benefits. However, the plan conflates two objectives (code reduction and distributed system foundation) and overestimates parallelization feasibility. Core type unification work is sound but underestimated by ~30%.

**Key recommendation:** Split scope. Do type unification first with int64 IDs. Add distributed features (ULID, HLC, change_events) as a separate project IF and WHEN needed.

---

## Critical Concerns (By Severity)

### 1. Scope Creep Disguised as Opportunity

**Severity:** HIGH

The plan started as "reduce 34,589 lines of boilerplate" but has expanded to include:
- ULID primary keys (distribution-ready)
- HLC timestamps (distributed ordering)
- `change_events` table (audit + replication + webhooks)
- `node_id` on every table
- Backup coordination tables
- Conflict resolution policies

This is no longer a refactor. This is a distributed systems foundation being built alongside a code reduction effort.

**Question:** Is distributed deployment actually needed? If there is no immediate need for multi-node deployment, building this infrastructure now is premature optimization.

**Mitigation:** Split into two phases:
- **Phase A:** Complete type unification with int64 IDs
- **Phase B:** Add ULID/distributed features IF and WHEN needed

---

### 2. 21-Agent Parallelization is Unrealistic

**Severity:** HIGH

The plan proposes 21 agents working in parallel on Steps 18-38, each simplifying a different table wrapper.

**Why it will fail:**
- All 21 files import from `internal/db/` (shared types)
- All 21 files are touched by Step 17a/17b/17c
- Git merge conflicts will be constant
- If any agent modifies shared utilities (e.g., `convert.go`), all others break

The dependency graph claims isolation but the code shows coupling. Every wrapper file references `StringToNullInt64`, `StringToInt64`, `StringToNullString` from shared utilities.

**Mitigation:**
1. Run wrapper simplification serially (recommended)
2. Or cap at 3-5 parallel agents
3. Or create a shared "types migration" step that completes before any wrapper work

---

### 3. sqlc Wildcard Overrides Are Untested

**Severity:** MEDIUM-HIGH

The proposed `sqlc.yml` uses patterns like:
```yaml
- column: "*.datatype_id"
  nullable: true
  go_type:
    import: "github.com/hegner123/modulacms/internal/db"
    type: "NullableDatatypeID"
```

**Concerns:**
- Wildcard `*` matching may not work as documented
- May match columns unintentionally (e.g., `legacy_datatype_id`)
- Behavior may differ between sqlc versions
- Later overrides may override earlier explicit ones

**Mitigation:** Before any implementation:
1. Create minimal test: one table, one wildcard override
2. Verify sqlc generates expected code
3. Document exact sqlc version this works with

---

### 4. DbDriver Interface Changes Underestimated

**Severity:** MEDIUM

The plan does not clearly state that `DbDriver` interface will have breaking signature changes:

```go
// Before
GetDatatype(int64) (*Datatypes, error)

// After
GetDatatype(DatatypeID) (*Datatypes, error)
```

This means **every call site** using `DbDriver` must be updated. Step 40 (API handlers) and Step 41 (CLI operations) mention this but vastly underestimate the effort.

**Mitigation:** Before starting:
```bash
grep -r "GetDatatype\|DeleteDatatype\|CreateDatatype" --include="*.go" | wc -l
```
Multiply result by number of tables (21). That is actual scope for Phase 3/4.

---

### 5. HLC Implementation Has Bugs

**Severity:** MEDIUM

The HLC implementation has problems:

```go
func HLCNow() HLC {
    // ...
    if hlcCounter == 0 {
        // Counter overflow, wait for wall clock to advance
        time.Sleep(time.Millisecond)
        return HLCNow()  // RECURSIVE CALL WHILE HOLDING MUTEX
    }
}
```

**Issues:**
- Recursive call can stack overflow under extreme load (65,536 events in one millisecond)
- `time.Sleep(time.Millisecond)` doesn't guarantee wall clock advances on all OSes
- `HLCUpdate` function is defined but never used in the plan

**Mitigation:**
1. Replace recursion with iteration
2. Use `time.Now().UnixMilli()` in a loop until it advances
3. If `HLCUpdate` not needed for this refactor, remove it (YAGNI)

---

### 6. Code Reduction Estimate is Optimistic

**Severity:** MEDIUM

Claimed: 16,795 lines -> 7,000-8,000 lines (50-55% reduction)

**Reality check for one table (datatype.go):**
- Before: ~565 lines (structs + mapping + wrappers)
- After: ~230 lines
- Reduction: ~60%

**But:** Plan adds ~200+ lines of new type definitions:
- `types_ids.go` (~80 lines per type × 25+ types)
- `types_timestamp.go`
- `types_hlc.go`
- `types_enums.go`
- `types_validation.go`
- `types_change_events.go`

**Net reduction is likely 40-45%, not 50-55%.** Still valuable, but set expectations correctly.

---

### 7. Change Events Implementation Undefined

**Severity:** MEDIUM

The plan creates the `change_events` table and Go types, but does not define:
- Who writes the code to log events?
- Application-level logging or database triggers?
- Every CREATE/UPDATE/DELETE needs to emit an event - that's significant new code

**Mitigation:** Define explicit step for change event logging implementation, or defer to Phase B.

---

### 8. Change Events Scalability

**Severity:** LOW-MEDIUM

The `change_events` table will grow unboundedly:
- Every INSERT, UPDATE, DELETE creates a row
- `old_values JSONB` and `new_values JSONB` can be megabytes per change
- No retention policy defined
- `idx_events_unsynced` partial index grows until sync is implemented

**Estimate:** 100 edits/minute = 52 million rows/year = ~25GB/year for audit logs alone.

**Mitigation:** Define retention policy before implementation, or defer to Phase B.

---

## Questions Requiring Answers

| # | Question | Impact |
|---|----------|--------|
| 1 | How many total call sites exist for DbDriver methods? | Determines Phase 3/4 scope |
| 2 | Has anyone tested sqlc wildcard overrides (`*.column`)? | Blocks Step 13 |
| 3 | What is actual timeline need for distributed deployment? | Determines if ULID/HLC needed now |
| 4 | Who maintains the 49-step HQ project if agent fails? | Recovery complexity |
| 5 | What happens if sqlc generates non-compiling code? | Rollback procedure |
| 6 | What test coverage exists for database layer? | Refactor safety |
| 7 | Who writes change_events logging code? | Scope clarity |

---

## What's Good About the Plan

1. **Problem statement is accurate** - Wrapper layer IS bloated; 9 struct variants per table IS excessive
2. **Custom ID types prevent real bugs** - Passing `UserID` where `DatatypeID` expected is a real bug class
3. **Unified Timestamp handling is sensible** - Current `sql.NullString` for dates is painful
4. **Document organization is excellent** - 8 well-structured documents with clear dependencies
5. **Rollback plan exists** - Most plans don't have one
6. **Architecture diagram is sound** - Custom types in `internal/db/` shared by all sqlc packages is correct
7. **FK indexes are overdue** - Adding these is legitimately free and prevents future performance issues

---

## Recommendations

### Must Do Before Implementation

1. **Validate sqlc wildcards** - Test `*.column` syntax works before Step 13
2. **Count call sites** - `grep -r` for DbDriver method calls to scope Phase 3/4
3. **Fix HLC deadlock** - Replace recursion with iteration

### Strongly Recommended

4. **Split scope** - Type unification (int64) first, distributed features second
5. **Serialize Phase 3** - Run wrapper simplification sequentially, not 21 parallel agents
6. **Define change_events implementation** - Who writes the logging code?

### Consider

7. **Accept 40-45% reduction** - Still valuable, but realistic
8. **Defer backup tables** - Features, not prerequisites
9. **Cap parallel agents at 3-5** - If parallelization is desired

---

## Revised Scope Proposal

### Phase A: Type Unification (Current Priority)

**Goal:** Reduce wrapper boilerplate through custom types with int64 IDs

**Include:**
- Steps 0, 5-16 (baseline, types, config, codegen)
- Steps 17a-17c, 18-38 (wrapper simplification) - SERIALIZE
- Steps 39-45 (cleanup, integration, testing, docs)

**Exclude (defer to Phase B):**
- Steps 1-4 (schema improvements: change_events, backups, node_id)
- ULID primary keys
- HLC timestamps
- Distributed system foundation

**Expected outcome:** 40-45% code reduction, compile-time type safety

### Phase B: Distributed Foundation (Future)

**Trigger:** When multi-node deployment is actually needed

**Include:**
- ULID migration
- HLC implementation (fixed)
- change_events table + logging
- node_id columns
- Backup coordination tables

---

## Document History

| Date | Change |
|------|--------|
| 2026-01-22 | Initial review by skeptical-architect agent |

# Go Backend Review - Audit Package Design (Enterprise Context)

**Reviewer:** go-backend-reviewer agent
**Date:** 2026-01-23
**Document:** AUDIT_PACKAGE_REVISED_DESIGN.md
**Context:** Greenfield project, zero users, competing with Contentful/Strapi/Sanity

---

## Assessment

This design document specifies an audit trail system for a CMS using Go generics with a command pattern to wrap database mutations in atomic transactions that record change events. The approach bundles context, audit metadata, and operation parameters into per-entity command structs that implement generic interfaces.

---

## Critical Issues

### 1. Raw SQL in a Multi-Driver System

`/Users/home/Documents/Code/Go_dev/modulacms/AUDIT_PACKAGE_REVISED_DESIGN.md` lines 146-178

```go
func recordChangeEventTx(ctx context.Context, tx *sql.Tx, p RecordChangeEventParams) error {
    query := `
        INSERT INTO change_events (
            event_id, hlc_timestamp, node_id, table_name, record_id,
            operation, user_id, old_values, new_values, request_id, ip,
            wall_timestamp
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))
    `
```

This function only works for SQLite. The document acknowledges MySQL/PostgreSQL need different SQL but the core `audited.go` calls this directly. You cannot ship this to production without solving the driver dispatch problem.

**Fix**: The command interfaces must include a method that provides the change event recorder, or the `audited` package needs a `ChangeEventRecorder` interface injected at initialization. The sqlc approach mentioned in lines 599-627 is correct but not wired into the actual code.

### 2. Swallowed Errors in Marshal Operations

Lines 261, 269, 306:
```go
oldValues, _ := json.Marshal(before)
newValues, _ := json.Marshal(cmd.Params())
```

Ignoring marshal errors means silent audit log corruption. If `before` contains an unmarshallable type (channel, func, cycle), the audit record gets `null` instead of data. For a compliance-critical audit system, this is unacceptable.

**Fix**: Return errors or use a logging fallback:
```go
oldValues, err := json.Marshal(before)
if err != nil {
    return fmt.Errorf("marshal before state for %s: %w", cmd.TableName(), err)
}
```

### 3. No Connection Pooling Awareness

The commands take `*sql.DB` and immediately start transactions. Under high load, this design will exhaust connection pools because:
- Each audited operation holds a connection for the entire transaction
- No backpressure mechanism exists
- The 30-second default timeout is too long for high-traffic scenarios

For enterprise-grade multi-tenant usage, you need connection pool exhaustion handling and shorter default timeouts (5-10 seconds for typical CMS operations).

### 4. Missing Error Classification

All errors bubble up as wrapped errors, but HTTP handlers need to distinguish:
- Not found (404)
- Constraint violation (409)
- Validation failure (400)
- Internal error (500)

Current design forces handlers to string-match errors or use sentinel comparisons scattered throughout the codebase. Enterprise competitors like Contentful return structured error responses with codes.

---

## Improvements

### Interface Design is Overly Verbose

Three separate interfaces with 6-7 methods each when you could use a single interface with a discriminated union approach or functional options. The current design requires ~75 lines per entity across 23 entities times 3 drivers. That is 5,175 lines of near-identical boilerplate.

Consider a generator or embedding a base command struct to reduce this. The pattern is sound but the repetition invites copy-paste bugs during maintenance.

### `GetBefore` Without Row Locking

Lines 255-256 note "For strict consistency, use SELECT ... FOR UPDATE" but leave implementation as a future concern. For a CMS competing with enterprise products, concurrent edits to the same content item are common. Without row locking on read-before-update:
1. User A reads content (version 1)
2. User B reads content (version 1)
3. User A updates, audit shows version 1 -> 2
4. User B updates, audit shows version 1 -> 3 (lost update, misleading audit)

This is a data integrity issue for compliance-focused customers.

### Audit Context Lacks Tenant/Organization

For multi-tenant SaaS (standard for enterprise CMS), `AuditContext` should include `TenantID` or `OrganizationID`. Adding this later requires migrating all change_events records and updating every command struct factory call.

### HLC Generation Inside Transaction

`types.HLCNow()` is called inside the transaction but HLC correctness depends on monotonic time across the distributed system. If two nodes generate events simultaneously, the HLC ordering may not match transaction commit order. This may or may not matter for your use case, but it is worth understanding the semantics you are getting.

---

## Testing Notes

Areas requiring strict test coverage:

1. **Rollback atomicity**: Verify that partial failures (mutation succeeds, audit insert fails) properly roll back the mutation
2. **Connection exhaustion**: Simulate pool exhaustion and verify graceful degradation
3. **Concurrent updates to same record**: Verify audit log accuracy under race conditions
4. **Driver parity**: All three drivers must produce identical audit semantics
5. **Context cancellation mid-transaction**: Verify cleanup behavior
6. **JSON marshal edge cases**: Structs with nil pointers, empty slices, zero-value typed fields

---

## Architectural Feedback for Enterprise Competition

**What is Solid:**
- Atomic transaction approach is correct for audit integrity
- Generic interfaces are idiomatic Go 1.18+
- Command pattern provides clean call sites
- Using sqlc for type-safe queries is the right call
- HLC for distributed ordering shows forward thinking

**Gaps vs. Enterprise Competitors:**

1. **No rate limiting or quota enforcement in the audit path.** Contentful/Sanity limit API calls per organization. Your audit system has no hooks for this.

2. **No async audit option.** High-volume systems often need fire-and-forget audit with eventual consistency. Requiring synchronous audit on every mutation limits throughput ceiling.

3. **No audit log retention policy hooks.** Enterprise customers need configurable retention (GDPR, SOC2). The design has no mechanism for audit log rotation or archival.

4. **No webhook/event emission.** Modern CMSes emit events on mutations. The audit system captures data but does not trigger external notifications.

5. **Single-region assumption.** The HLC and transaction model assume single-database deployment. Multi-region active-active replication (a Sanity differentiator) would require CRDT-style conflict resolution the current design cannot accommodate.

---

## Verdict

This is a reasonable starting point for a single-region, moderate-traffic CMS. The core transaction-audit atomicity is correctly designed. The command pattern, while verbose, is maintainable and idiomatic.

For enterprise competition, you need to solve the driver dispatch for change event recording before implementation, add error classification for proper HTTP response codes, and plan for the multi-tenancy metadata now rather than retrofitting it later. The connection pool pressure under load is a real concern that should be addressed with circuit breakers or adaptive timeouts.

The 5,000+ lines of command struct boilerplate is acceptable if you generate it or treat it as a one-time cost. Manual maintenance of that much repetitive code across three drivers will eventually introduce inconsistencies.

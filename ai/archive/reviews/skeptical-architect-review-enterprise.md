# Skeptical Architect Review - Audit Package Design (Enterprise Context)

**Reviewer:** skeptical-architect agent
**Date:** 2026-01-23
**Document:** AUDIT_PACKAGE_REVISED_DESIGN.md
**Context:** Greenfield project, zero users, competing with Contentful/Strapi/Sanity

---

## Initial Assessment

**Risk Level: MEDIUM-HIGH**

The design shows solid Go idioms and addresses the immediate atomic transaction problem correctly. However, for a system that aims to compete with Contentful and Sanity at enterprise scale, this design has several architectural blind spots that will become painful at 10K+ operations/second and when enterprise compliance teams start asking questions.

The good news: this is greenfield, and these issues can be addressed before they become technical debt. The bad news: some of these concerns require design changes, not just implementation tweaks.

---

## Critical Concerns

### 1. Missing Audit Log Integrity Guarantees (SOC2/HIPAA Blocker)

**Severity: HIGH**

The design stores audit events in the same database as application data. Enterprise compliance (SOC2, HIPAA, GDPR Article 30) requires demonstrable tamper-evidence for audit logs.

**The problem:**
```go
// Your current approach
return recordChangeEventTx(ctx, tx, RecordChangeEventParams{...})
// ^ This writes to change_events table in same DB
```

Anyone with database admin access can:
- Modify audit records after the fact
- Delete inconvenient entries
- Alter timestamps

**What Contentful/Sanity do:**
- Separate, append-only audit storage
- Cryptographic chaining (hash of previous record embedded in current)
- Write-once storage (S3 Object Lock, WORM storage)
- Audit log integrity checksums

**What you need:**
```go
type RecordChangeEventParams struct {
    // ... existing fields ...
    PreviousHash string // Hash of previous event (chain integrity)
    EventHash    string // Hash of this event's contents
}
```

And a periodic integrity verification process that can detect tampering.

---

### 2. No Retention Policy or Archival Strategy

**Severity: HIGH**

At enterprise scale, a busy CMS generates millions of change events per month. Your design shows no consideration for:

- **Retention periods** (GDPR right to erasure vs. audit requirements)
- **Archival to cold storage** (S3 Glacier, cold blob storage)
- **Partition strategy** (time-based partitioning for query performance)
- **Size-based rollover** (change_events table will grow unbounded)

**The math:**
- 100K content updates/day = 100K audit records/day
- Average record size: ~2KB (with JSON old/new values)
- Monthly: 3M records, ~6GB
- Yearly: 36M records, ~72GB
- In 3 years: 100M+ records

Your `ListChangeEvents` with offset/limit pagination will fall over spectacularly around 10M records.

**What you need:**
- Time-partitioned tables (PostgreSQL native partitioning)
- Archival jobs that move old events to cold storage
- Cursor-based pagination (not offset-based)
- Configurable retention policies per customer/table

---

### 3. Change Event Table Will Become a Write Bottleneck

**Severity: MEDIUM-HIGH**

At 10K writes/second, every mutation writes to `change_events`. This single table becomes your global write bottleneck because:

1. Every insert requires acquiring a lock (briefly)
2. Indexes on `change_events` must be updated
3. If you add the cryptographic chain I mentioned above, it gets worse
4. Replication lag on the change_events table affects all operations

**What happens at scale:**
```
10K content writes/sec → 10K change_event inserts/sec
                       → Index pressure
                       → Write amplification
                       → Increased replication lag
                       → Degraded read consistency
```

**Solutions to consider:**
- Buffered async writes (write to local queue, batch insert to DB)
- Sharded change_events tables (by tenant, by table_name, by time bucket)
- Separate audit database (isolate audit write load from content reads)

The trade-off with async writes: you lose atomicity. If the app crashes between content write and audit write, you have unaudited mutations. Enterprise customers will ask about this gap.

---

### 4. The AuditContext is Too Thin for Enterprise Needs

**Severity: MEDIUM**

```go
type AuditContext struct {
    NodeID    types.NodeID
    UserID    types.UserID
    RequestID string
    IP        string
}
```

Enterprise audit requirements typically need:
- **Session ID** (link multiple operations to a session)
- **Tenant/Organization ID** (multi-tenant isolation)
- **User Agent** (detect API vs. browser vs. mobile)
- **Geographic location** (GDPR Article 44 - data transfer tracking)
- **Authentication method** (OAuth, SSO, API key, service account)
- **Permission context** (what role/permissions enabled this action)
- **Correlation ID** (trace across distributed systems)
- **Client application ID** (which app/integration made this call)

Contentful and Sanity track all of this for their enterprise customers.

---

### 5. No Consideration for Multi-Tenant Audit Isolation

**Severity: MEDIUM-HIGH**

If ModulaCMS serves multiple organizations (SaaS model), audit data must be strictly isolated. Your design shows:
- No tenant_id in AuditContext
- No tenant_id in change_events table
- No query scoping by tenant

**The risk:**
- Customer A's admin could potentially query Customer B's audit logs
- Data breach disclosure would affect all customers, not just the compromised one
- SOC2 Type II auditors will flag this immediately

---

### 6. The "Action" Field is Semantically Confusing

**Severity: LOW-MEDIUM**

You have both `Operation` (INSERT, UPDATE, DELETE) and `Action` (create, update, delete, publish, archive). The distinction is unclear in the design doc.

```go
Operation types.Operation // INSERT, UPDATE, DELETE
Action    types.Action    // create, update, delete, publish, archive
```

Questions that need answers:
- When would Operation=UPDATE but Action=publish?
- Is "archive" an Action that maps to Operation=UPDATE?
- What about "restore from archive" - is that an Action?
- "publish" is domain-specific - does this belong in a generic audit layer?

This kind of semantic overlap leads to data inconsistency as different developers interpret it differently.

---

### 7. Driver-Specific SQL in the Audited Package is a Maintenance Burden

**Severity: MEDIUM**

```go
// Note: For MySQL, use NOW() instead of datetime('now')
// Note: For PostgreSQL, use NOW() and $1, $2, etc. placeholders
```

You're proposing to have driver-specific SQL inside the `audited` package. This:
- Duplicates the driver abstraction pattern you already have
- Creates a second place where SQL dialect differences must be managed
- Will be forgotten when adding new fields or queries

Better approach: Use the existing sqlc infrastructure. Add a `RecordChangeEventTx` query to each driver's queries.sql and have the audited package call through the existing wrapper layer.

---

### 8. No Query Capabilities for Audit Consumers

**Severity: MEDIUM**

Your design focuses on writing audit events but barely addresses reading them. Enterprise needs include:

- **"Show me all changes to this content item"** - You have this (GetChangeEventsByRecord)
- **"Show me all changes by this user"** - Partially (ListChangeEventsByUserParams exists but not in new design)
- **"Show me all changes in the last 24 hours"** - Missing
- **"Show me all changes to 'published' status"** - Missing
- **"Export audit log for compliance review"** - Missing
- **"Show me changes across multiple tables for a workflow"** - Missing
- **"Search audit logs by content in old/new values"** - Missing (and hard with JSON)

These aren't nice-to-haves. Your enterprise sales team will lose deals without them.

---

### 9. The 30-Second Default Timeout is Arbitrary

**Severity: LOW**

```go
if _, ok := ctx.Deadline(); !ok {
    var cancel context.CancelFunc
    ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
}
```

Why 30 seconds? What if the operation is a batch import of 10,000 records? What if it's a quick single-field update that should timeout in 5 seconds?

This should be configurable, not hardcoded. At minimum, make it a package-level variable. Better: pass it through configuration.

---

### 10. No Mention of Audit Event Versioning

**Severity: MEDIUM**

Your change_event schema will evolve. What happens when you add a new field? What about when the structure of OldValues/NewValues changes because an entity schema changed?

Enterprise systems need:
- Schema version field in each event
- Migration path for querying old events with old schema
- Forward-compatible design (new fields should be optional)

---

## Questions That Need Answers

1. **Tenant isolation**: Is this single-tenant or multi-tenant? If multi-tenant, where is tenant scoping in the audit design?

2. **Audit log access control**: Who can read audit logs? Is there a separate permission model for audit access?

3. **PII handling**: GDPR Article 17 (right to erasure) vs. audit retention - how do you anonymize audit records for deleted users?

4. **Bulk operations**: What happens when someone imports 50,000 content items? 50,000 audit events atomically?

5. **Audit of audit access**: SOC2 requires auditing who accessed the audit logs. Is that covered?

6. **Disaster recovery**: Audit logs are often required for post-incident analysis. What's the backup/restore story specifically for audit data?

7. **Real-time streaming**: Many enterprises want real-time audit feeds (SIEM integration). Is there a plan for CDC/streaming?

8. **Performance SLAs**: What's the acceptable overhead of auditing? 1ms? 10ms? 100ms per operation?

9. **Selective auditing**: Can customers opt out of auditing certain tables (e.g., high-frequency low-value data)?

10. **Geographic compliance**: For GDPR, can audit data be stored in a specific region? EU customers often require EU data residency.

---

## What is Actually Good

1. **Atomic transactions**: The core insight that mutations and audit records must be in the same transaction is correct and critical. Many audit systems get this wrong.

2. **Command pattern**: The interface-based command pattern is idiomatic Go and will be familiar to anyone maintaining this code.

3. **Typed fields with validation**: Using types.Email, types.Slug, etc. for automatic validation during JSON unmarshal is smart defensive programming.

4. **HLC timestamps**: Having hybrid logical clocks shows you've thought about distributed ordering, which is more than most CMS systems do.

5. **Existing change_event infrastructure**: You're not starting from zero. The existing change_event.go and types_change_events.go provide a solid foundation.

6. **Call site ergonomics**: The `audited.Ctx()` constructor and single-argument generic functions make the API genuinely pleasant to use.

7. **Testing strategy**: The integration test structure that runs against all three database drivers is exactly right.

---

## Verdict

**Can I be convinced this is viable?** Yes, but not as currently specified.

This design solves the immediate problem (atomic audit logging) well. For a v1 with early customers, it would work. But you stated the ambition is to compete with Contentful and Sanity at enterprise scale. That requires addressing:

**Must-have before enterprise deals:**
1. Audit log integrity guarantees (cryptographic chaining or separate immutable storage)
2. Multi-tenant isolation (if SaaS model)
3. Retention and archival strategy
4. Richer audit context (session, tenant, auth method, etc.)

**Should-have for competitive parity:**
1. Cursor-based pagination for audit queries
2. Time-partitioned storage
3. Comprehensive query capabilities
4. Schema versioning

**Nice-to-have but differentiating:**
1. Real-time audit streaming (SIEM integration)
2. Audit log export for compliance
3. Configurable audit scoping (which tables, which operations)

**What it would take to convince me:**
- A separate document addressing audit log integrity for compliance
- A capacity planning section showing this scales to 100K ops/sec
- Multi-tenant isolation strategy (even if "not yet implemented")
- Retention policy configuration design
- Acknowledgment of the Action/Operation semantic overlap with a resolution

The architecture is not wrong - it's incomplete for the stated ambition. Better to address these gaps now while you have zero users and zero technical debt than to discover them when an enterprise customer's security team rejects your SOC2 questionnaire.

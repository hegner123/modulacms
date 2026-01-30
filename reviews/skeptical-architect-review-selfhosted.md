# Skeptical Architect Review - Audit Package Design (Self-Hosted Context)

**Reviewer:** skeptical-architect agent
**Date:** 2026-01-23
**Document:** AUDIT_PACKAGE_REVISED_DESIGN.md
**Context:** Self-hosted CMS (like WordPress, Strapi, Directus), single-tenant per deployment

---

## Initial Assessment

**Overall Risk Level: Medium** - The design is technically sound but architecturally misaligned with the deployment model.

This is a well-engineered solution to a problem that may not need solving at this level of sophistication. The design shows enterprise patterns being applied to a self-hosted tool where the threat model, scale requirements, and operational context are fundamentally different.

---

## Critical Concerns

### 1. This is overengineered for the deployment model

**Severity: High**

A self-hosted CMS has one user/team with root access to the server. They own the database. They can read the logs. They can modify anything. The atomicity guarantees and tamper-resistance implied by this design provide security theater, not actual security.

What does a self-hosted WordPress user get for audit trails? A simple activity log plugin that writes to the same database they already control. No HLC timestamps. No node IDs. No atomic transactions. And it is plenty.

**Competitor comparison:**
- **Strapi (comparable self-hosted headless CMS)**: Has an audit log feature in Enterprise only. The open-source version has none.
- **Directus**: Activity tracking that writes who-what-when. No before/after state capture by default.
- **Payload CMS**: Versions collection for content, simple activity logging.

### 2. 5,175 lines of boilerplate for 23 entities

**Severity: High**

The code statistics section admits this design will produce over 5,000 lines of repetitive command struct code across three database drivers. For a self-hosted CMS that might have 1-5 active users, this is a maintenance burden that far exceeds the value delivered.

Every new entity requires ~75 lines across the drivers. Every schema change propagates to command structs. This is the kind of technical debt that makes projects hard to maintain after the original authors leave.

### 3. Hybrid Logical Clocks are wrong tool for this job

**Severity: Medium**

HLCs solve distributed system ordering problems. Your deployment model is single-node self-hosted. You are building infrastructure for a problem you do not have.

A simple auto-incrementing integer or timestamp would suffice. The wall clock on a single-node deployment is authoritative. You do not need causality tracking across distributed nodes because there are no distributed nodes in a typical self-hosted deployment.

### 4. NodeID complexity for single-tenant

**Severity: Medium**

Every audit context requires a NodeID. For what purpose in a self-hosted system? A self-hosted CMS is not running multiple nodes that need to be distinguished. This is infrastructure for a distributed system being imposed on a standalone application.

### 5. Driver-specific SQL for change event recording

**Severity: Low-Medium**

The design acknowledges needing `datetime('now')` vs `NOW()` vs PostgreSQL syntax, then adds more complexity with sqlc overrides. For a self-hosted CMS, you could reasonably recommend SQLite for most deployments and simplify significantly.

---

## Questions That Need Answers

1. **Who is the audit trail for?** In a self-hosted model, the admin IS the database owner. Who are you protecting against? Rogue editors? They should use proper user permissions. Compliance auditors? They can query the database directly.

2. **What is the actual query pattern for this data?** Will anyone ever query "show me all changes to content_data in the last week"? Or is this write-only data that consumes disk space indefinitely?

3. **What is the retention policy?** A VPS with 20GB disk running for 3 years with active content editing - how large does `change_events` grow? Is there a cleanup mechanism? The design explicitly punts on "archival" as a separate concern.

4. **What happens on disaster recovery?** If someone restores from backup, the change_events table and actual data can become inconsistent. Has this been considered?

5. **Why are all 25 entities getting audited?** Do you really need audit trails for `sessions`, `tokens`, `media_dimensions`, `tables`? These are system bookkeeping, not business data. Audit fatigue is real.

---

## What is Actually Good

1. **The transaction wrapper integration with existing `types.WithTransaction`** is the right approach. At least the revised design reuses existing infrastructure.

2. **Using sqlc's DBTX interface** for transaction-aware queries is idiomatic and clean. This is how Go database code should work.

3. **The command pattern provides clean call sites.** The HTTP handler examples are readable and the intent is clear.

4. **Capturing before-state for updates and deletes** is genuinely useful for content recovery, which self-hosters would actually want.

5. **Context timeout handling** with the 30-second default is a reasonable defensive measure.

---

## Verdict: This Needs Rescoping

I cannot recommend this design as-is for a self-hosted CMS. It applies enterprise SaaS patterns where simpler solutions would serve better.

**What would convince me this is appropriate:**

1. Evidence that your target market actually needs compliance-grade audit trails (regulated industries, large agency clients with security requirements)

2. A concrete plan for the multi-node distributed deployment that justifies HLC and NodeID

3. Reduction of audited entities to just business-critical content (content_data, content_fields, users, roles, permissions - maybe 8 entities, not 25)

4. A retention/archival strategy that prevents unbounded growth

5. A migration path for existing deployments to enable auditing gradually

---

## Simpler Alternatives Worth Considering

### Option A: Activity Log Table (WordPress-style)

```sql
CREATE TABLE activity_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id TEXT,
    action TEXT,  -- 'create_content', 'update_user', 'delete_media'
    object_type TEXT,
    object_id TEXT,
    summary TEXT,  -- human-readable: "Updated post 'Hello World'"
    created_at TEXT
);
```

No before/after capture. No atomicity with mutations. Just "who did what when". This covers 80% of what self-hosters actually want with 5% of the complexity.

### Option B: Content Versions Table

Focus audit logging only on content (the actual business value), and make it a proper versioning system that enables undo/rollback.

```sql
CREATE TABLE content_versions (
    version_id TEXT PRIMARY KEY,
    content_id TEXT,
    version_number INTEGER,
    data JSON,
    created_by TEXT,
    created_at TEXT
);
```

This gives self-hosters something they will actually use (content history, rollback) instead of a compliance artifact they will never query.

### Option C: Optional Audit Plugin

Make auditing a Lua plugin that site owners can enable if they want. Default deployment has no auditing. This matches how WordPress handles it (activity log is a plugin, not core).

---

## The Real Question

Before building any of this, ask: What are your competitors doing?

- **Strapi open-source**: No audit logging
- **Directus**: Simple activity tracking (no before/after)
- **Payload**: Content versioning only
- **WordPress**: Core has no audit logging; plugins add simple activity logs

If the comparable products in your space do not have this feature, either you are targeting a different market segment, or you are over-building. Which is it?

---

## Summary Table

| Aspect | Assessment |
|--------|------------|
| Technical correctness | Good - the revised design fixes the atomicity issues |
| Alignment with deployment model | Poor - enterprise patterns for self-hosted tool |
| Maintenance burden | High - 5000+ lines of boilerplate |
| Operational simplicity | Poor - HLC, NodeID add complexity without benefit |
| Value to self-hosters | Questionable - most would prefer simpler versioning |
| Competitive positioning | Unclear - exceeds what comparable tools provide |

**Recommendation**: Scope down dramatically. Audit content changes only. Use simple timestamps. Make it optional. Ship a simpler activity log that serves 90% of users, and leave the enterprise audit infrastructure for a future version when you have enterprise customers asking for it.

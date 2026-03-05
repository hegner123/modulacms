# Service Layer Extraction Roadmap

**Goal:** Extract shared business logic from admin handlers, API handlers, and MCP tools into a unified `internal/service/` package. Each consumer becomes a thin adapter: parse input, call service, format output.

**Current state:** 3 consumers (admin panel, REST API, MCP server) each independently implement validation, orchestration, and transformation against the DbDriver interface. ~328 DbDriver methods, ~200 API routes, ~170 MCP tools, ~22 admin handler files.

---

## Service Inventory

### Tier 1 — Largest / Highest Value (do first)

These have the most duplication, the most active development, and the most complex business logic beyond simple CRUD.

| # | Service | Consumers | DbDriver Methods | Complexity Notes |
|---|---------|-----------|-----------------|------------------|
| 1 | **ContentService** | Admin, API, MCP | ~60 | Tree ops (reorder, move, save), heal, batch update, publish/unpublish/schedule, versions, restore. Two parallel sets (admin + public). Most complex domain. |
| 2 | **SchemaService** (Datatypes + Fields + FieldTypes) | Admin, API, MCP | ~50 | Datatype-field linking/unlinking, sort order management, "full" joins. Two parallel sets. |
| 3 | **MediaService** | Admin, API, MCP | ~18 | Upload with image optimization, health check, orphan cleanup, dimension presets. Side effects (S3/filesystem). |

### Tier 2 — Medium Size / Moderate Value

Straightforward CRUD with some business rules.

| # | Service | Consumers | DbDriver Methods | Complexity Notes |
|---|---------|-----------|-----------------|------------------|
| 4 | **RouteService** | Admin, API, MCP | ~10 | Slug uniqueness, rename propagation. |
| 5 | **UserService** | Admin, API, MCP | ~10 | Password hashing, role assignment (viewer default), full-user aggregation. |
| 6 | **RBACService** (Roles + Permissions) | Admin, API, MCP | ~28 | System-protected guards, permission cache refresh trigger, role-permission associations. |
| 7 | **PluginService** | Admin, API, MCP | ~10 | Lifecycle (enable/disable/reload), route & hook approval, orphan table cleanup. |
| 8 | **WebhookService** | Admin, API | ~20 | CRUD, test dispatch, delivery tracking, retry logic, pruning. |
| 9 | **LocaleService** | Admin, API | ~12 | Default locale management, enabled/disabled filtering, translation creation. |

### Tier 3 — Small / Thin CRUD (do last)

Little business logic beyond the DbDriver call. Still worth extracting for consistency, but low urgency.

| # | Service | Consumers | DbDriver Methods | Complexity Notes |
|---|---------|-----------|-----------------|------------------|
| 10 | **SessionService** | Admin, API, MCP | ~8 | Mostly passthrough CRUD. |
| 11 | **TokenService** | Admin, API, MCP | ~10 | Create with expiry, revocation. |
| 12 | **SSHKeyService** | API, MCP | ~8 | Fingerprint extraction on create. |
| 13 | **OAuthService** | API, MCP | ~10 | Token refresh, provider linkage. |
| 14 | **TableService** | API, MCP | ~8 | Dynamic table metadata CRUD. |
| 15 | **ConfigService** | Admin, API, MCP | ~3 | Wraps config.Manager, redacts sensitive values, category filtering. |
| 16 | **ImportService** | Admin, API, MCP | ~6 | Multi-format import orchestration (contentful, sanity, strapi, wordpress). |
| 17 | **DeployService** | API, MCP | ~4 | Export/import sync payloads, dry-run. |
| 18 | **AuditService** | Admin | ~8 | Read-only change event queries. |
| 19 | **BackupService** | (cmd only) | ~10 | Backup/restore orchestration. |

**Total: 19 services covering ~328 DbDriver methods across 3 consumer layers.**

---

## Phases

### Phase 0 — Foundation
**Scope:** Create `internal/service/` package structure and establish the pattern.

- Define base conventions: how services accept dependencies (DbDriver, config.Manager, PermissionCache, etc.), error types, context propagation, audit context threading
- Create `service.Registry` or similar to hold all service instances (injected at startup, passed to handlers)
- Decide on per-domain files vs per-domain sub-packages (recommend files: `content.go`, `schema.go`, etc.)
- No handler changes yet — just the skeleton

**Size:** Small. 1 session.

### Phase 1 — Prove the Pattern (1 Tier-1 service)
**Scope:** Extract **SchemaService** (datatypes + fields + field types) end-to-end.

Why schema first (not content):
- Complex enough to validate the pattern (linking, sort order, "full" joins)
- Smaller than content (~50 methods vs ~60, no tree ops)
- Content depends on schema (datatypes, fields), so schema service exists first
- Changes here immediately benefit admin panel iteration workflow

Work:
- Implement SchemaService with all datatype, field, field type, and datatype-field operations
- Rewire admin handlers in `datatypes.go`, `fields.go`, `field_types.go` to call service
- Rewire API handlers in `datatypes.go`, `adminDatatypes.go`, `fields.go`, `adminFields.go`, `field_types.go`, `admin_field_types.go`
- Rewire MCP tools in `tools_schema.go`, `tools_admin_schema.go`
- Write service-level tests (table-driven, SQLite)

**Size:** Large. 2-3 sessions with parallel agents (admin, API, MCP rewiring can parallelize).

### Phase 2 — Content (the big one)
**Scope:** Extract **ContentService** — the largest and most complex domain.

Work:
- Content CRUD (admin + public variants)
- Content fields CRUD
- Tree operations: reorder, move, save tree structure
- Publishing: publish, unpublish, schedule
- Versioning: create, list, restore, snapshot
- Batch updates (atomic content + fields)
- Heal (dry-run + apply)
- Content relations

Depends on SchemaService (Phase 1) for datatype/field lookups.

**Size:** Very large. 3-5 sessions. Can split into sub-phases:
- 2a: Basic CRUD + fields
- 2b: Tree operations + reorder/move
- 2c: Versioning + publishing + scheduling
- 2d: Batch, heal, relations

### Phase 3 — Media + Routes
**Scope:** Extract **MediaService** and **RouteService**.

- Media: upload orchestration (optimization + S3), health, cleanup, dimensions
- Routes: CRUD with slug uniqueness and rename

**Size:** Medium. 1-2 sessions (can parallelize the two).

### Phase 4 — Users + RBAC
**Scope:** Extract **UserService**, **RBACService**.

- Users: password hashing, role assignment, full aggregation
- RBAC: system-protected guards, cache refresh trigger, association management

**Size:** Medium. 1-2 sessions.

### Phase 5 — Plugins + Webhooks + Locales
**Scope:** Extract **PluginService**, **WebhookService**, **LocaleService**.

**Size:** Medium. 1-2 sessions (three independent domains, fully parallelizable).

### Phase 6 — Thin CRUD Sweep
**Scope:** Extract remaining Tier 3 services: Sessions, Tokens, SSH Keys, OAuth, Tables, Config, Import, Deploy, Audit, Backup.

10 small services, all straightforward. Agents can do 3-4 per session in parallel.

**Size:** Medium total, small individually. 2-3 sessions.

### Phase 7 — MCP Server Alignment
**Scope:** Refactor the MCP server (`mcp/`) to call services directly instead of going through the Go SDK over HTTP.

Currently: `MCP tool -> Go SDK -> HTTP API -> handler -> DbDriver`
After:     `MCP tool -> Service -> DbDriver`

This eliminates the HTTP round-trip for local MCP usage and means MCP tools automatically get any service-layer improvements. Requires the MCP binary to import `internal/service/` directly (moves from separate binary to either embedded or shared library approach).

**Size:** Large (architectural decision + rewiring ~170 tools). 2-3 sessions. Could also be deferred if the HTTP indirection is acceptable.

---

## Estimated Timeline

| Phase | Scope | Est. Sessions | Parallelizable |
|-------|-------|:---:|:---:|
| 0 | Foundation | 1 | No |
| 1 | Schema (prove pattern) | 2-3 | Yes (admin/API/MCP) |
| 2 | Content | 3-5 | Yes (sub-phases) |
| 3 | Media + Routes | 1-2 | Yes (two domains) |
| 4 | Users + RBAC | 1-2 | Yes (two domains) |
| 5 | Plugins + Webhooks + Locales | 1-2 | Yes (three domains) |
| 6 | Thin CRUD sweep | 2-3 | Yes (10 domains) |
| 7 | MCP direct integration | 2-3 | Defer optional |

**Total: ~13-21 sessions** (with parallelization on the lower end).

---

## Admin/Public Consolidation Question

The codebase has parallel admin and public variants for content, schema, and routes (separate DbDriver method sets, separate handler files, separate MCP tool files). Two approaches:

**A. Unified service methods with access-level parameter:**
```
ContentService.List(ctx, params, AccessLevel)  // AccessLevel = Admin | Public
```
Pro: Single implementation. Con: Every method needs branching logic.

**B. Separate service interfaces per access level:**
```
AdminContentService / PublicContentService
```
Pro: Clean separation, no branching. Con: Some duplication between admin/public.

**Recommendation:** Start with **B** (separate), merge later if the overlap proves >80%. The admin and public query sets are different enough (admin has versioning, scheduling, tree ops; public has slug delivery, format selection) that forcing them together may create more complexity than it saves.

---

## Risk & Dependencies

- **DbDriver interface is stable** — services wrap it, don't replace it. No schema migration needed.
- **Audited command pattern** must thread through services (services accept `audited.AuditContext`, not handlers).
- **Permission cache refresh** — RBACService must trigger `pc.Load()` after mutations. Services need access to PermissionCache.
- **Config hot-reload** — ConfigService wraps Manager, which already handles this. No new complexity.
- **Backward compatibility** — handlers can be migrated incrementally. No big-bang cutover required.
- **Testing** — service tests run against SQLite directly. Existing handler tests continue to work during migration (handlers just delegate).

---

## How to Use This Plan

Each phase is scoped for a detailed implementation plan. When starting a phase:

1. Reference this document for scope boundaries
2. Have an agent create a detailed plan for that specific phase
3. Use `hq` for phases with parallelizable sub-work
4. After each phase, verify all three consumers (admin, API, MCP) still pass tests

The phases are ordered by dependency and value. Phases 3-6 can be reordered based on which domains you're actively developing.

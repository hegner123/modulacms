# SDK Sync Plan

Date: 2026-03-07
Status: Not started
Source: Automated audit of Go, TypeScript, and Swift SDKs against `internal/router/mux.go` and `internal/db/types/`

## Overview

All three SDKs (Go, TypeScript, Swift) have endpoint and type gaps relative to the HTTP router. The TypeScript SDK has the best coverage. The Go SDK has the most gaps. This plan is organized by priority, with cross-SDK work first.

Run SDK tests after each phase:
```bash
just sdk go test
just sdk ts test
just sdk swift test
```

---

## Phase 1: Content Tree Operations (All SDKs)

**Goal:** Enable content reordering and movement from all SDKs.

### Endpoints to add

| Endpoint | Method | Permission |
|----------|--------|------------|
| `/api/v1/contentdata/reorder` | POST | `content:update` |
| `/api/v1/contentdata/move` | POST | `content:update` |
| `/api/v1/admincontentdatas/reorder` | POST | `admincontent:update` |
| `/api/v1/admincontentdatas/move` | POST | `admincontent:update` |

### Per-SDK work

**Go SDK:**
- Add `ReorderContent(contentID, direction)` and `MoveContent(sourceID, targetID, routeID)` to content resource
- Add admin variants

**TypeScript Admin SDK:**
- Add `contentData.reorder(id, direction)` and `contentData.move(sourceID, targetID, routeID)`
- Add `adminContentData.reorder(...)` and `adminContentData.move(...)`

**Swift SDK:**
- Add `reorder(contentID:direction:)` and `move(sourceID:targetID:routeID:)` methods
- Add admin variants

---

## Phase 2: Sort Order Endpoints (All SDKs)

**Goal:** Enable datatype and field sort order management.

### Endpoints to add

| Endpoint | Method | Permission |
|----------|--------|------------|
| `/api/v1/datatype/max-sort-order` | GET | `datatypes:read` |
| `/api/v1/datatype/{id}/sort-order` | PUT | `datatypes:update` |
| `/api/v1/fields/max-sort-order` | GET | `fields:read` |
| `/api/v1/fields/{id}/sort-order` | PUT | `fields:update` |
| `/api/v1/admindatatypes/max-sort-order` | GET | `admindatatypes:read` |
| `/api/v1/admindatatypes/{id}/sort-order` | PUT | `admindatatypes:update` |

### Per-SDK work

All three SDKs: add `getMaxSortOrder()` and `updateSortOrder(id, sortOrder)` methods to datatype and field resources (both regular and admin).

---

## Phase 3: Go SDK — Missing Core Endpoints

**Goal:** Bring Go SDK to feature parity with TypeScript.

### 3A. Publishing & Versioning

Add to Go SDK:

```
POST /api/v1/content/publish
POST /api/v1/content/unpublish
POST /api/v1/content/schedule
GET  /api/v1/content/versions
GET  /api/v1/content/versions/{id}
POST /api/v1/content/versions
DELETE /api/v1/content/versions/{id}
POST /api/v1/content/restore
```

Plus admin variants (`/api/v1/admin/content/...`).

Requires: `PublishingResource` and `AdminPublishingResource` structs.

### 3B. Configuration

Add to Go SDK:

```
GET   /api/v1/admin/config
PATCH /api/v1/admin/config
GET   /api/v1/admin/config/meta
```

Requires: `ConfigResource` struct with `Get()`, `Update()`, `Meta()`.

### 3C. Globals & Query

Add to Go SDK:

```
GET /api/v1/globals
GET /api/v1/query/{datatype}
```

- `GetGlobals()` — public endpoint, returns global field values
- `QueryContent(datatype, params)` — public endpoint with filters, sort, pagination

### 3D. Content Healing

Add to Go SDK:

```
POST /api/v1/admin/content/heal
```

Requires: `ContentHealResource` with `Heal(dryRun bool)`.

---

## Phase 4: TypeScript SDK — Minor Gaps

**Goal:** Close remaining TypeScript gaps.

### 4A. Delivery SDK: Add getGlobals()

Add `getGlobals()` method to the delivery SDK (`@modulacms/sdk`) for `/api/v1/globals`.

### 4B. Types: Add missing branded IDs

Add to `@modulacms/types`:
- `TokenID` — used by token CRUD (high priority)
- `UserSshKeyID` — used by SSH key operations (medium priority)

### 4C. Types: Add RouteType enum

Add `RouteType` enum with values: `static`, `dynamic`, `api`, `redirect`.

### 4D. Standardize AdminRouteID usage

Audit admin SDK for inconsistent `AdminRouteID` usage (some methods use string, some use branded type). Standardize to branded type.

---

## Phase 5: Swift SDK — Endpoint Gaps

**Goal:** Close Swift-specific endpoint gaps beyond Phase 1-2.

### 5A. Content Relations CRUD

Add `ContentRelationsResource` with full CRUD for `/api/v1/contentrelations`.

### 5B. Admin Content Translations

Add method for `POST /api/v1/admin/admincontentdata/{id}/translations` (regular content translations already covered).

### 5C. Complete Plugin Stubs

`PluginRoutesResource` — add `list()`, `approve()`, `revoke()` methods.
`PluginHooksResource` — implement or remove stub (depends on router-side completeness).

---

## Phase 6: Missing Enums (All SDKs)

**Goal:** Add enums that exist in `internal/db/types/types_enums.go` but not in SDKs.

### 6A. Operational enums (medium priority)

| Enum | Values | Used by |
|------|--------|---------|
| `PluginStatus` | installed, enabled | Plugin management |
| `RouteType` | static, dynamic, api, redirect | Route classification |

### 6B. Audit enums (low priority)

| Enum | Values | Used by |
|------|--------|---------|
| `Operation` | INSERT, UPDATE, DELETE | Change events |
| `Action` | create, update, delete, publish | Audit trail |
| `ConflictPolicy` | lww, manual | Distributed sync |

### 6C. Backup enums (low priority)

| Enum | Values | Used by |
|------|--------|---------|
| `BackupType` | full, incremental, differential | Backup system |
| `BackupStatus` | pending, in_progress, completed, failed | Backup ops |
| `VerificationStatus` | pending, verified, failed | Backup verification |
| `BackupSetStatus` | pending, complete, partial | Backup collections |

---

## Phase 7: Missing ID Types (Go + Swift)

**Goal:** Add ID types that exist in `internal/db/types/types_ids.go` but not in SDKs.

### Go SDK — add to `ids.go`:
- `BackupSetID`
- `NodeID`
- `PipelineID`
- `PluginID`
- `VerificationID`

### Swift SDK — already complete (30 types).

### TypeScript — add `TokenID`, `UserSshKeyID` (covered in Phase 4B).

---

## Phase 8: Media Management (All SDKs)

**Goal:** Add media health, cleanup, and references endpoints.

| Endpoint | Method | Permission |
|----------|--------|------------|
| `GET /api/v1/media/health` | GET | `media:admin` |
| `DELETE /api/v1/media/cleanup` | DELETE | `media:admin` |
| `GET /api/v1/media/references` | GET | `media:read` |

All three SDKs: add methods to existing media resource or create `MediaAdminResource`.

---

## Phase 9: Plugin Route/Hook Management (All SDKs)

**Goal:** Add plugin route approval and hook management.

**Prerequisite:** Verify router-side endpoints are complete. Plugin hooks approval may be an incomplete feature.

| Endpoint | Method | Permission |
|----------|--------|------------|
| `GET /api/v1/admin/plugins/routes` | GET | `plugins:read` |
| `POST /api/v1/admin/plugins/routes/approve` | POST | `plugins:update` |
| `POST /api/v1/admin/plugins/routes/revoke` | POST | `plugins:update` |

---

## Current Coverage Summary

| SDK | Endpoint Coverage | ID Types | Enums | Notes |
|-----|-------------------|----------|-------|-------|
| **Go** | ~70% | Missing 5 | Missing 8+ | Largest gaps: publishing, versioning, config, query |
| **TypeScript** | ~95% | Missing 2 | Missing RouteType | Best coverage; delivery SDK missing globals |
| **Swift** | ~80% | Complete | Missing 8 | Good ID coverage; missing sort-order, tree ops, relations |

## Verification

After each phase:
1. `just sdk go test` / `just sdk go vet`
2. `just sdk ts test` / `just sdk ts typecheck`
3. `just sdk swift test`
4. Manual smoke test against running CMS instance

# Plan: Add `sort_order` to Datatypes and AdminDatatypes Tables

## Context

When creating content, the datatype picker list has no predictable ordering — datatypes appear in ULID order (effectively random). Adding a `sort_order` column (matching the existing fields pattern) gives users control over display order across the TUI, admin panel, and API.

Greenfield project with zero deployments — no migration concerns.

## Key Difference from Fields

Fields always have a non-null `parent_id` (a DatatypeID). Datatypes can have `parent_id = NULL` (root-level). SQL `WHERE parent_id = ?` with a NULL param returns 0 rows. We need **two** max-sort-order queries: one for root datatypes (`WHERE parent_id IS NULL`) and one for children (`WHERE parent_id = ?`). The wrapper method branches on `NullableDatatypeID.Valid`.

AdminDatatypes follow the same pattern with `NullableAdminDatatypeID`.

---

# Part A: Public Datatypes

## Phase 1: Schema (6 files in `sql/schema/7_datatypes/`) ✅

Add `sort_order` column after `parent_id` in all schema and query files.

### 1a. DDL schemas (3 files)

| File | Column definition |
|------|-------------------|
| `schema.sql` (SQLite) | `sort_order INTEGER NOT NULL DEFAULT 0,` after parent_id FK |
| `schema_mysql.sql` | `sort_order INT NOT NULL DEFAULT 0,` after `parent_id` line |
| `schema_psql.sql` | `sort_order INTEGER NOT NULL DEFAULT 0,` after parent_id FK |

### 1b. Query files (3 files)

Changes per file:

1. **CreateDatatypeTable** inline DDL — add `sort_order` column
2. **All LIST queries** — change `ORDER BY datatype_id` / `ORDER BY label` to `ORDER BY sort_order, datatype_id`
   - `ListDatatype`, `ListDatatypeGlobal`, `ListDatatypeRoot`, `ListDatatypePaginated` — currently `ORDER BY datatype_id`
   - `ListDatatypeChildren`, `ListDatatypeChildrenPaginated` — currently `ORDER BY label`
3. **CreateDatatype** — add `sort_order` to column list and values
4. **UpdateDatatype** — add `sort_order = ?` to SET clause
5. **Three new queries:**

```sql
-- name: UpdateDatatypeSortOrder :exec
UPDATE datatypes SET sort_order = ? WHERE datatype_id = ?;

-- name: GetMaxDatatypeRootSortOrder :one
SELECT COALESCE(MAX(sort_order), -1) FROM datatypes WHERE parent_id IS NULL;

-- name: GetMaxDatatypeSortOrderByParentID :one
SELECT COALESCE(MAX(sort_order), -1) FROM datatypes WHERE parent_id = ?;
```

PostgreSQL uses `$1`/`$2` placeholders; SQLite and MySQL use `?`.

## Phase 2: Combined Schema Files ✅

Add `sort_order` column to the datatypes block in:
- `sql/all_schema.sql`
- `sql/all_schema_mysql.sql`
- `sql/all_schema_psql.sql`

## Phase 3: Regenerate sqlc ✅

```bash
just sqlc
```

## Phase 4: Update dbgen Definition + Regenerate ✅

**File: `tools/dbgen/definitions.go`** — SortOrder added, `just dbgen` run.

This updates `internal/db/datatype_gen.go` automatically — adding `SortOrder` to:
- `Datatypes` struct
- `CreateDatatypeParams` / `UpdateDatatypeParams`
- `MapStringDatatype`
- All `MapDatatype`, `MapCreateDatatypeParams`, `MapUpdateDatatypeParams` (3 backends each)
- All `NewDatatypeCmd.Execute`, `UpdateDatatypeCmd.Execute` (3 backends each)

## Phase 5: Manual Wrapper Updates ✅

### 5a. `internal/db/imports.go` ✅
Add `SortOrder string` to `StringDatatypes` after `ParentID`.

### 5b. `internal/db/datatype_custom.go`
Add `SortOrder string` to `DatatypeJSON` struct and `SortOrder: fmt.Sprintf("%d", a.SortOrder)` to `MapDatatypeJSON`.

### 5c. New file: `internal/db/datatype_sort_order.go`

Following the exact pattern in `internal/db/field_custom.go`:

- `UpdateDatatypeSortOrderParams` struct (`DatatypeID types.DatatypeID`, `SortOrder int64`)
- 3 audited command structs: `UpdateDatatypeSortOrderCmd` (SQLite), `UpdateDatatypeSortOrderCmdMysql`, `UpdateDatatypeSortOrderCmdPsql`
  - Each implements: `Context()`, `AuditContext()`, `Connection()`, `Recorder()`, `TableName() → "datatypes"`, `Params()`, `GetID()`, `GetBefore()`, `Execute()`
  - MySQL/PostgreSQL cast: `int32(c.params.SortOrder)`
- 3 factory methods + 3 public `UpdateDatatypeSortOrder` methods
- `GetMaxDatatypeSortOrder(parentID types.NullableDatatypeID) (int64, error)` — branches:
  - `parentID.Valid == false` → calls `GetMaxDatatypeRootSortOrder`
  - `parentID.Valid == true` → calls `GetMaxDatatypeSortOrderByParentID`
- Reuse existing `coalesceToInt64` from `field_custom.go` (same package)

## Phase 6: DbDriver Interface ✅

**File: `internal/db/db.go`**

Add after `UpdateDatatype`:
```go
UpdateDatatypeSortOrder(context.Context, audited.AuditContext, UpdateDatatypeSortOrderParams) error
GetMaxDatatypeSortOrder(types.NullableDatatypeID) (int64, error)
```

## Phase 7: Bootstrap Data ✅

**File: `internal/db/db.go`** — 6 `CreateDatatypeParams` literals (2 per backend, 3 backends):
- Page datatype: add `SortOrder: 0`
- Reference datatype: add `SortOrder: 1`

## Phase 8: Preserve SortOrder in UpdateDatatype Call Sites ✅

Adding `SortOrder` to `UpdateDatatypeParams` means every caller must set it or it defaults to zero, silently resetting the sort order. All 4 call sites need fixing:

### 8a. `internal/tui/update_cms.go:83` (TUI form submit)
Already fetches `existing` via `d.GetDatatype(datatypeID)`. Add:
```go
SortOrder: existing.SortOrder,
```
to the `UpdateDatatypeParams` literal.

### 8b. `internal/tui/update_dialog.go:2195` (TUI dialog submit)
Already fetches `existing`. Add:
```go
SortOrder: existing.SortOrder,
```
to the `UpdateDatatypeParams` literal.

### 8c. `internal/admin/handlers/datatypes.go:315` (Admin panel PATCH)
Already fetches `existing`. Add:
```go
SortOrder: existing.SortOrder,
```
to the `UpdateDatatypeParams` literal.

### 8d. `internal/router/datatypes.go:184` (API PUT)
Decodes JSON directly into `UpdateDatatypeParams`. Since `SortOrder` is now a field on the struct with `json:"sort_order"`, the API caller can include it in the JSON body. If they omit it, JSON unmarshaling defaults to `0`. This is acceptable — the dedicated `PUT /api/v1/datatype/{id}/sort-order` endpoint is the intended way to change sort order. The general update endpoint is for metadata (name, label, type).

**No change needed** — the `json:"sort_order"` tag on the struct handles it. API callers who want to preserve sort_order should include it in the request body.

## Phase 9: Router — Sort Order Endpoints ✅

### 9a. New file: `internal/router/datatypeSortOrder.go`

Modeled on `internal/router/fieldSortOrder.go`:
- `DatatypeSortOrderHandler` — PUT, validates DatatypeID, decodes `{"sort_order": int64}`, calls `UpdateDatatypeSortOrder`
- `DatatypeMaxSortOrderHandler` — GET, reads `?parent_id=` (optional — omit for root), calls `GetMaxDatatypeSortOrder`. When `parent_id` is absent, construct `NullableDatatypeID{Valid: false}`. When present, validate as ULID then construct `NullableDatatypeID{ID: ..., Valid: true}`.

### 9b. Route registration in `internal/router/mux.go`

Insert between the `GET /api/v1/datatype/full` route and the `/api/v1/datatype/` catch-all:

```go
mux.Handle("PUT /api/v1/datatype/{id}/sort-order", ...)  // datatypes:update
mux.Handle("GET /api/v1/datatype/max-sort-order", ...)    // datatypes:read
```

## Phase 10: TUI — Creation with Sort Order ✅

**File: `internal/tui/update_dialog.go` — `HandleCreateDatatypeFromDialog`**

Before `d.CreateDatatype`, add:
```go
maxSort, sortErr := d.GetMaxDatatypeSortOrder(parentID)
if sortErr != nil {
    maxSort = -1
}
```
Add `SortOrder: maxSort + 1` to the `CreateDatatypeParams`.

## Phase 11: TUI — Reorder Handler ✅

**File: `internal/tui/commands.go`**

Add after `HandleReorderField` block:
- `DatatypeReorderedMsg` struct
- `HandleReorderDatatype` method — swaps sort_order between two datatypes (two `UpdateDatatypeSortOrder` calls)

## Phase 12: TUI — Reorder Controls ✅

**File: `internal/tui/update_controls.go` — `DatatypesControls`**

Add `ActionReorderUp`/`ActionReorderDown` key handlers in the `TreePanel` focus block. Two helper methods: `datatypesControlsReorderUp`, `datatypesControlsReorderDown` — get adjacent items from `m.AllDatatypes`, call `HandleReorderDatatype`.

**File: `internal/tui/update_cms.go`**

Handle `DatatypeReorderedMsg`: adjust cursor, call `AllDatatypesFetchCmd()` to reload.

## Phase 13: Admin Panel — Creation with Sort Order

**File: `internal/admin/handlers/datatypes.go` — `DatatypeCreateHandler`**

Admin panel does not support creating child datatypes (no parent_id in form). Before `driver.CreateDatatype`, add max sort order lookup with null parent:
```go
maxSort, sortErr := driver.GetMaxDatatypeSortOrder(types.NullableDatatypeID{})
if sortErr != nil {
    maxSort = -1
}
```
Add `SortOrder: maxSort + 1` to `CreateDatatypeParams`.

---

# Part B: AdminDatatypes (mirror of Part A)

## Phase 14: AdminDatatypes Schema (6 files in `sql/schema/9_admin_datatypes/`) ✅

Identical pattern to Phase 1 but for `admin_datatypes` table.

### 14a. DDL schemas (3 files)

| File | Column definition |
|------|-------------------|
| `schema.sql` (SQLite) | `sort_order INTEGER NOT NULL DEFAULT 0,` after parent_id FK |
| `schema_mysql.sql` | `sort_order INT NOT NULL DEFAULT 0,` after `parent_id` line |
| `schema_psql.sql` | `sort_order INTEGER NOT NULL DEFAULT 0,` after parent_id FK |

### 14b. Query files (3 files)

Same changes as Phase 1b but for `admin_datatypes`:
1. **CreateAdminDatatypeTable** inline DDL — add `sort_order` column
2. **All LIST queries** — `ORDER BY sort_order, admin_datatype_id`
3. **CreateAdminDatatype** — add `sort_order` to column list and values
4. **UpdateAdminDatatype** — add `sort_order = ?` to SET clause
5. **Three new queries:**

```sql
-- name: UpdateAdminDatatypeSortOrder :exec
UPDATE admin_datatypes SET sort_order = ? WHERE admin_datatype_id = ?;

-- name: GetMaxAdminDatatypeRootSortOrder :one
SELECT COALESCE(MAX(sort_order), -1) FROM admin_datatypes WHERE parent_id IS NULL;

-- name: GetMaxAdminDatatypeSortOrderByParentID :one
SELECT COALESCE(MAX(sort_order), -1) FROM admin_datatypes WHERE parent_id = ?;
```

## Phase 15: AdminDatatypes Combined Schema Files ✅

Add `sort_order` column to the `admin_datatypes` block in:
- `sql/all_schema.sql`
- `sql/all_schema_mysql.sql`
- `sql/all_schema_psql.sql`

## Phase 16: Regenerate sqlc (covers both tables) ✅

```bash
just sqlc
```

## Phase 17: Update dbgen Definition for AdminDatatypes ✅

**File: `tools/dbgen/definitions.go`** — SortOrder added, `just dbgen` run.

## Phase 18: AdminDatatypes Manual Wrapper Updates ✅

### 18a. `internal/db/imports.go` ✅
Add `SortOrder string` to `StringAdminDatatypes` after `ParentID`.

### 18b. `internal/db/admin_datatype_custom.go`
Add `SortOrder string` to the admin variant of `DatatypeJSON` (in `MapAdminDatatypeJSON`) and `SortOrder: fmt.Sprintf("%d", a.SortOrder)`.

### 18c. New file: `internal/db/admin_datatype_sort_order.go`

Same pattern as `datatype_sort_order.go` but with `AdminDatatypeID`, `NullableAdminDatatypeID`, table name `"admin_datatypes"`:

- `UpdateAdminDatatypeSortOrderParams` struct
- 3 audited command structs + 3 factory methods + 3 public methods
- `GetMaxAdminDatatypeSortOrder(parentID types.NullableAdminDatatypeID) (int64, error)` — branches on `Valid`

## Phase 19: AdminDatatypes DbDriver Interface ✅

**File: `internal/db/db.go`**

Add after `UpdateAdminDatatype`:
```go
UpdateAdminDatatypeSortOrder(context.Context, audited.AuditContext, UpdateAdminDatatypeSortOrderParams) error
GetMaxAdminDatatypeSortOrder(types.NullableAdminDatatypeID) (int64, error)
```

## Phase 20: AdminDatatypes Bootstrap Data ✅

**File: `internal/db/db.go`** — 3 `CreateAdminDatatypeParams` literals (1 per backend, 3 backends):
- Admin Page datatype: add `SortOrder: 0`

## Phase 21: Preserve SortOrder in UpdateAdminDatatype Call Sites ✅

### 21a. `internal/tui/admin_update_dialog.go:278` (TUI dialog submit)
Already fetches `existing`. Add:
```go
SortOrder: existing.SortOrder,
```
to the `UpdateAdminDatatypeParams` literal.

### 21b. `internal/router/adminDatatypes.go:114` (API PUT)
Decodes JSON directly into `UpdateAdminDatatypeParams`. Same treatment as Phase 8d — the `json:"sort_order"` tag handles it. No change needed.

## Phase 22: AdminDatatypes Router — Sort Order Endpoints ✅

### 22a. New file: `internal/router/adminDatatypeSortOrder.go`

Same pattern as `datatypeSortOrder.go` but with `AdminDatatypeID`:
- `AdminDatatypeSortOrderHandler` — PUT
- `AdminDatatypeMaxSortOrderHandler` — GET

### 22b. Route registration in `internal/router/mux.go`

```go
mux.Handle("PUT /api/v1/admindatatypes/{id}/sort-order", ...)  // admin_datatypes:update
mux.Handle("GET /api/v1/admindatatypes/max-sort-order", ...)    // admin_datatypes:read
```

## Phase 23: AdminDatatypes TUI — Creation with Sort Order ✅

**File: `internal/tui/admin_update_dialog.go`** — find `HandleCreateAdminDatatypeFromDialog`

Before `d.CreateAdminDatatype`, add max sort order lookup and `SortOrder: maxSort + 1`.

## Phase 24: AdminDatatypes TUI — Reorder Handler ✅

**File: `internal/tui/commands.go`** (or `admin_constructors.go` if admin commands are separate)

- `AdminDatatypeReorderedMsg` struct
- `HandleReorderAdminDatatype` method

## Phase 25: AdminDatatypes TUI — Reorder Controls ✅

**File: `internal/tui/admin_controls.go` — `AdminDatatypesControls`**

Add `ActionReorderUp`/`ActionReorderDown` key handlers. Two helper methods.

**File: `internal/tui/admin_update_cms.go`**

Handle `AdminDatatypeReorderedMsg`: adjust cursor, reload.

---

# Part C: Remote Driver

## Phase 26: Remote Driver Conversion Functions ✅

### 26a. `internal/remote/convert_core.go`

Update `datatypeToDb` (line 65) — add `SortOrder` mapping:
```go
SortOrder: s.SortOrder,
```

Update `datatypeUpdateFromDb` (line 110) — add `SortOrder` mapping:
```go
SortOrder: d.SortOrder,
```

### 26b. `internal/remote/convert_admin.go`

Update `adminDatatypeCreateFromDb` (line 93) — add `SortOrder` mapping.

Update `adminDatatypeUpdateFromDb` (line 104) — add `SortOrder` mapping.

Update `adminDatatypeToDb` — add `SortOrder` mapping.

## Phase 27: Remote Driver Methods

**File: `internal/remote/driver.go`**

Add 4 new methods that call the SDK sort order endpoints:
- `UpdateDatatypeSortOrder` — calls SDK `DatatypesExtra.UpdateSortOrder`
- `GetMaxDatatypeSortOrder` — calls SDK `DatatypesExtra.MaxSortOrder`
- `UpdateAdminDatatypeSortOrder` — calls SDK `AdminDatatypesExtra.UpdateSortOrder`
- `GetMaxAdminDatatypeSortOrder` — calls SDK `AdminDatatypesExtra.MaxSortOrder`

---

# Part D: Go SDK

## Phase 28: Go SDK Entity Types ✅

**File: `sdks/go/types.go`**

### 28a. Datatype (line 122)
Add after `ParentID`:
```go
SortOrder int64 `json:"sort_order"`
```

### 28b. CreateDatatypeParams
Add:
```go
SortOrder int64 `json:"sort_order"`
```

### 28c. UpdateDatatypeParams
Add:
```go
SortOrder int64 `json:"sort_order"`
```

### 28d. AdminDatatype (line 675)
Add after `ParentID`:
```go
SortOrder int64 `json:"sort_order"`
```

### 28e. CreateAdminDatatypeParams (line 687)
Add:
```go
SortOrder int64 `json:"sort_order"`
```

### 28f. UpdateAdminDatatypeParams (line 696)
Add:
```go
SortOrder int64 `json:"sort_order"`
```

## Phase 29: Go SDK Extra Resource — Datatypes

**New file: `sdks/go/datatypes_extra.go`**

Following `sdks/go/fields_extra.go` pattern:

```go
type DatatypesExtraResource struct {
    http *httpClient
}

func (r *DatatypesExtraResource) UpdateSortOrder(ctx context.Context, datatypeID DatatypeID, sortOrder int64) error
func (r *DatatypesExtraResource) MaxSortOrder(ctx context.Context, parentID *DatatypeID) (int64, error)
```

Note: `MaxSortOrder` takes `*DatatypeID` — `nil` means root, non-nil means children of that parent.

Endpoints:
- `PUT /api/v1/datatype/{id}/sort-order`
- `GET /api/v1/datatype/max-sort-order` (optional `?parent_id=`)

## Phase 30: Go SDK Extra Resource — AdminDatatypes

**New file: `sdks/go/admin_datatypes_extra.go`**

Same pattern as Phase 29 but with `AdminDatatypeID`:

```go
type AdminDatatypesExtraResource struct {
    http *httpClient
}

func (r *AdminDatatypesExtraResource) UpdateSortOrder(ctx context.Context, id AdminDatatypeID, sortOrder int64) error
func (r *AdminDatatypesExtraResource) MaxSortOrder(ctx context.Context, parentID *AdminDatatypeID) (int64, error)
```

Endpoints:
- `PUT /api/v1/admindatatypes/{id}/sort-order`
- `GET /api/v1/admindatatypes/max-sort-order`

## Phase 31: Go SDK Client Registration

**File: `sdks/go/modula.go`**

Add fields to `ModulaCMSClient`:
```go
DatatypesExtra *DatatypesExtraResource
AdminDatatypesExtra *AdminDatatypesExtraResource
```

Initialize in `NewClient`:
```go
DatatypesExtra: &DatatypesExtraResource{http: h},
AdminDatatypesExtra: &AdminDatatypesExtraResource{http: h},
```

---

# Part E: TypeScript SDK

## Phase 32: TypeScript Shared Types

**File: `sdks/typescript/types/src/entities/schema.ts`**

### 32a. Datatype type
Add after `parent_id`:
```typescript
/** Display ordering position. */
sort_order: number
```

### 32b. CreateDatatypeParams
Add `sort_order: number`.

### 32c. UpdateDatatypeParams
Add `sort_order: number`.

## Phase 33: TypeScript Admin SDK Types

**File: `sdks/typescript/modulacms-admin-sdk/src/types/admin.ts`**

### 33a. AdminDatatype type (line 103)
Add after `parent_id`:
```typescript
/** Display ordering position. */
sort_order: number
```

### 33b. CreateAdminDatatypeParams (line 238)
Add `sort_order: number`.

### 33c. UpdateAdminDatatypeParams (line 356)
Add `sort_order: number`.

## Phase 34: TypeScript Admin SDK — Sort Order Methods ✅

**File: `sdks/typescript/modulacms-admin-sdk/src/index.ts`**

### 34a. Datatypes resource (around line 531)
Extend the datatypes resource with sort order methods:
```typescript
datatypes: {
  ...createResource<...>(http, 'datatype'),
  getFull(...) {...},
  updateSortOrder(id: DatatypeID, sortOrder: number, opts?: RequestOptions): Promise<void> {
    return http.put(`/datatype/${id}/sort-order`, { sort_order: sortOrder }, opts)
  },
  maxSortOrder(parentId?: DatatypeID, opts?: RequestOptions): Promise<{ max_sort_order: number }> {
    const params: Record<string, string> = {}
    if (parentId) params.parent_id = String(parentId)
    return http.get('/datatype/max-sort-order', params, opts)
  },
},
```

### 34b. AdminDatatypes resource (around line 775)
Add same methods for `admindatatypes`:
```typescript
adminDatatypes: {
  ...createResource<...>(http, 'admindatatypes'),
  updateSortOrder(id: AdminDatatypeID, sortOrder: number, opts?: RequestOptions): Promise<void> {
    return http.put(`/admindatatypes/${id}/sort-order`, { sort_order: sortOrder }, opts)
  },
  maxSortOrder(parentId?: AdminDatatypeID, opts?: RequestOptions): Promise<{ max_sort_order: number }> {
    const params: Record<string, string> = {}
    if (parentId) params.parent_id = String(parentId)
    return http.get('/admindatatypes/max-sort-order', params, opts)
  },
},
```

### 34c. Update TypeScript type exports
Update the `datatypes` and `adminDatatypes` type declarations in the client interface to include the new methods.

---

# Part F: Swift SDK

## Phase 35: Swift SDK Entity Types ✅

**File: `sdks/swift/Sources/Modula/Types.swift`**

### 35a. Datatype struct (line 293)
Add after `parentID`:
```swift
public let sortOrder: Int64
```
Add to CodingKeys:
```swift
case sortOrder = "sort_order"
```

### 35b. CreateDatatypeParams (line 315)
Add `public let sortOrder: Int64` property, init parameter with default `sortOrder: Int64 = 0`, CodingKey.

### 35c. UpdateDatatypeParams (line 349)
Add `public let sortOrder: Int64` property, init parameter with default `sortOrder: Int64 = 0`, CodingKey.

### 35d. AdminDatatype struct (line 1564)
Add `public let sortOrder: Int64` + CodingKey.

### 35e. CreateAdminDatatypeParams (line 1586)
Add `public let sortOrder: Int64` + init parameter + CodingKey.

### 35f. UpdateAdminDatatypeParams (line 1616)
Add `public let sortOrder: Int64` + init parameter + CodingKey.

---

# Part G: Verification

## Phase 36: Build and Test ✅

1. `just sqlc` — regenerate sqlc (covers both datatypes and admin_datatypes)
2. `just dbgen` — regenerate db wrappers (covers both entities)
3. `just check` — compile check all packages
4. `just test` — run full test suite
5. `just sdk go test` — run Go SDK tests
6. `just sdk go vet` — vet Go SDK
7. `just sdk ts build` — build TypeScript SDKs
8. `just sdk ts typecheck` — typecheck TypeScript SDKs
9. `just sdk swift build` — build Swift SDK
10. Manual: start TUI, navigate to datatypes page, verify list is ordered by sort_order. Create new datatypes, verify they appear at the end. Reorder with keybindings. Repeat for admin datatypes.

---

## Files Modified (complete summary)

### SQL Schema (18 files)
| File | Action |
|------|--------|
| `sql/schema/7_datatypes/schema.sql` | Add column |
| `sql/schema/7_datatypes/schema_mysql.sql` | Add column |
| `sql/schema/7_datatypes/schema_psql.sql` | Add column |
| `sql/schema/7_datatypes/queries.sql` | Add column, ORDER BY, new queries |
| `sql/schema/7_datatypes/queries_mysql.sql` | Add column, ORDER BY, new queries |
| `sql/schema/7_datatypes/queries_psql.sql` | Add column, ORDER BY, new queries |
| `sql/schema/9_admin_datatypes/schema.sql` | Add column |
| `sql/schema/9_admin_datatypes/schema_mysql.sql` | Add column |
| `sql/schema/9_admin_datatypes/schema_psql.sql` | Add column |
| `sql/schema/9_admin_datatypes/queries.sql` | Add column, ORDER BY, new queries |
| `sql/schema/9_admin_datatypes/queries_mysql.sql` | Add column, ORDER BY, new queries |
| `sql/schema/9_admin_datatypes/queries_psql.sql` | Add column, ORDER BY, new queries |
| `sql/all_schema.sql` | Add column to both tables |
| `sql/all_schema_mysql.sql` | Add column to both tables |
| `sql/all_schema_psql.sql` | Add column to both tables |

### Code Generation (3 files modified, 2 regenerated)
| File | Action |
|------|--------|
| `tools/dbgen/definitions.go` | Add SortOrder to Datatypes + AdminDatatypes entities |
| `internal/db/datatype_gen.go` | **Regenerated by dbgen** |
| `internal/db/admin_datatype_gen.go` | **Regenerated by dbgen** |

### Database Layer (8 files)
| File | Action |
|------|--------|
| `internal/db/db.go` | Add 4 interface methods + update 9 bootstrap params |
| `internal/db/imports.go` | Add SortOrder to StringDatatypes + StringAdminDatatypes |
| `internal/db/datatype_custom.go` | Add SortOrder to DatatypeJSON |
| `internal/db/admin_datatype_custom.go` | Add SortOrder to admin DatatypeJSON |
| `internal/db/datatype_sort_order.go` | **New** — sort order commands + queries |
| `internal/db/admin_datatype_sort_order.go` | **New** — admin sort order commands + queries |

### Router (3 files)
| File | Action |
|------|--------|
| `internal/router/mux.go` | Register 4 new routes |
| `internal/router/datatypeSortOrder.go` | **New** — PUT + GET handlers |
| `internal/router/adminDatatypeSortOrder.go` | **New** — PUT + GET handlers |

### TUI (8 files)
| File | Action |
|------|--------|
| `internal/tui/update_dialog.go` | Sort order on create + preserve on update |
| `internal/tui/update_cms.go` | Preserve on update + handle DatatypeReorderedMsg |
| `internal/tui/commands.go` | DatatypeReorderedMsg + HandleReorderDatatype |
| `internal/tui/update_controls.go` | Reorder key handlers |
| `internal/tui/admin_update_dialog.go` | Sort order on create + preserve on update |
| `internal/tui/admin_update_cms.go` | Handle AdminDatatypeReorderedMsg |
| `internal/tui/admin_constructors.go` | AdminDatatypeReorderedMsg + HandleReorderAdminDatatype |
| `internal/tui/admin_controls.go` | Admin reorder key handlers |

### Admin Panel (1 file)
| File | Action |
|------|--------|
| `internal/admin/handlers/datatypes.go` | Sort order on create + preserve on update |

### Remote Driver (2 files)
| File | Action |
|------|--------|
| `internal/remote/convert_core.go` | Add SortOrder to datatype conversions |
| `internal/remote/convert_admin.go` | Add SortOrder to admin datatype conversions |
| `internal/remote/driver.go` | Add 4 sort order proxy methods |

### Go SDK (4 files)
| File | Action |
|------|--------|
| `sdks/go/types.go` | Add SortOrder to 6 structs (Datatype/Admin + Create/Update params) |
| `sdks/go/modula.go` | Register DatatypesExtra + AdminDatatypesExtra |
| `sdks/go/datatypes_extra.go` | **New** — UpdateSortOrder + MaxSortOrder |
| `sdks/go/admin_datatypes_extra.go` | **New** — UpdateSortOrder + MaxSortOrder |

### TypeScript SDK (3 files)
| File | Action |
|------|--------|
| `sdks/typescript/types/src/entities/schema.ts` | Add sort_order to Datatype + params |
| `sdks/typescript/modulacms-admin-sdk/src/types/admin.ts` | Add sort_order to AdminDatatype + params |
| `sdks/typescript/modulacms-admin-sdk/src/index.ts` | Add sort order methods to datatypes + adminDatatypes |

### Swift SDK (1 file)
| File | Action |
|------|--------|
| `sdks/swift/Sources/Modula/Types.swift` | Add sortOrder to 6 structs |

**Total: ~50 files** (manual edits + regenerated)

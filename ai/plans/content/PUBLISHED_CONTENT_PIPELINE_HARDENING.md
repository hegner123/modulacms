# Published Content Pipeline Hardening

## Context

Production bug: two `content_versions` rows for the same `content_data_id` + `locale` both had `published=true`. The `GetPublishedSnapshot` query lacks `ORDER BY`, so the API non-deterministically served stale content. The publish flow in `PublishContent` (`publishing.go:231-258`) calls `GetMaxVersionNumber`, `ClearPublishedFlag`, then `CreateContentVersion`, but they run in separate transactions -- `ClearPublishedFlag` is a bare `:exec`, while `CreateContentVersion` wraps its INSERT+audit in its own transaction via `audited.Create`. Two concurrent publishes can interleave in the gap between these operations.

Three layers of defense: deterministic query, atomic publish transaction, periodic heal detection.

## Phase 1: Defensive SQL (GetPublishedSnapshot ORDER BY)

In all 6 query files, find the `GetPublishedSnapshot` (or `GetAdminPublishedSnapshot`) query and add `ORDER BY version_number DESC` before the existing `LIMIT 1`. Ensures the newest published version always wins, even if the single-published invariant is violated.

**Status: Partially complete.** All 3 public files have the ORDER BY fix. All 3 admin files still need it.

Already done:
- [x] `sql/schema/31_content_versions/queries.sql` -- ORDER BY added (line 48)
- [x] `sql/schema/31_content_versions/queries_mysql.sql` -- ORDER BY added (line 51)
- [x] `sql/schema/31_content_versions/queries_psql.sql` -- ORDER BY added (line 48)

Still needed:
- [ ] `sql/schema/32_admin_content_versions/queries.sql` -- MISSING ORDER BY on `GetAdminPublishedSnapshot`
- [ ] `sql/schema/32_admin_content_versions/queries_mysql.sql` -- MISSING ORDER BY on `GetAdminPublishedSnapshot`
- [ ] `sql/schema/32_admin_content_versions/queries_psql.sql` -- MISSING ORDER BY on `GetAdminPublishedSnapshot`

## Phase 2: New SQL Queries for Heal System

Add two queries to each of the 6 files (12 queries total). Also add a `GetMaxVersionNumberForUpdate` query to support in-transaction version number reads (see Phase 4).

### GetMaxVersionNumberForUpdate (new query, all 6 files)

Identical to `GetMaxVersionNumber` but intended for use inside a transaction to prevent concurrent publishes from reading the same max version number. SQLite serializes writes via WAL, so this is effectively the same query. MySQL and PostgreSQL variants should use `FOR UPDATE` to lock the rows.

SQLite (`queries.sql`):
```sql
-- name: GetMaxVersionNumberForUpdate :one
SELECT COALESCE(MAX(version_number), 0) FROM content_versions
WHERE content_data_id = ? AND locale = ?;
```

MySQL (`queries_mysql.sql`):
```sql
-- name: GetMaxVersionNumberForUpdate :one
SELECT COALESCE(MAX(version_number), 0) FROM content_versions
WHERE content_data_id = ? AND locale = ?
FOR UPDATE;
```

PostgreSQL (`queries_psql.sql`):
```sql
-- name: GetMaxVersionNumberForUpdate :one
SELECT COALESCE(MAX(version_number), 0) FROM content_versions
WHERE content_data_id = $1 AND locale = $2
FOR UPDATE;
```

Admin variants: replace `content_versions` with `admin_content_versions`, `content_data_id` with `admin_content_data_id`.

### ListDuplicatePublished (new query, all 6 files)

Finds `content_data_id`+`locale` groups with >1 published version.

SQLite (`queries.sql`):
```sql
-- name: ListDuplicatePublished :many
SELECT content_data_id, locale, COUNT(*) as pub_count
FROM content_versions WHERE published = 1
GROUP BY content_data_id, locale HAVING COUNT(*) > 1;
```

MySQL (`queries_mysql.sql`):
```sql
-- name: ListDuplicatePublished :many
SELECT content_data_id, locale, COUNT(*) as pub_count
FROM content_versions WHERE published = 1
GROUP BY content_data_id, locale HAVING COUNT(*) > 1;
```

PostgreSQL (`queries_psql.sql`):
```sql
-- name: ListDuplicatePublished :many
SELECT content_data_id, locale, COUNT(*) as pub_count
FROM content_versions WHERE published = TRUE
GROUP BY content_data_id, locale HAVING COUNT(*) > 1;
```

Admin variants: replace `content_versions` with `admin_content_versions`, `content_data_id` with `admin_content_data_id`. PostgreSQL admin uses `published = TRUE`.

### ClearPublishedFlagExcept (new query, all 6 files)

Clears published on all versions except one.

SQLite (`queries.sql`):
```sql
-- name: ClearPublishedFlagExcept :exec
UPDATE content_versions SET published = 0
WHERE content_data_id = ? AND locale = ? AND published = 1 AND content_version_id != ?;
```

MySQL (`queries_mysql.sql`):
```sql
-- name: ClearPublishedFlagExcept :exec
UPDATE content_versions SET published = 0
WHERE content_data_id = ? AND locale = ? AND published = 1 AND content_version_id != ?;
```

PostgreSQL (`queries_psql.sql`):
```sql
-- name: ClearPublishedFlagExcept :exec
UPDATE content_versions SET published = FALSE
WHERE content_data_id = $1 AND locale = $2 AND published = TRUE AND content_version_id != $3;
```

Admin variants: replace table with `admin_content_versions`, columns with `admin_content_data_id` and `admin_content_version_id`. PostgreSQL admin uses `published = FALSE` / `published = TRUE`.

**Note:** These heal queries scan the full table with no index on `(published, content_data_id, locale)`. This is acceptable for infrequent heal runs. If heal performance degrades at scale, add a partial index on `published = true` rows.

**Status:** `just sqlc` was already run after the Phase 1 public file changes (generated `internal/db-sqlite/`, `internal/db-mysql/`, `internal/db-psql/` `.go` files are modified). After completing the remaining admin ORDER BY fixes and adding all new queries, run `just sqlc` again.

## Phase 3: DbDriver Interface + Wrapper Methods

`internal/db/repositories.go` -- add to ContentVersions section (after `PruneOldVersions` at line 98, before `ContentFieldRepository` at line 101):

```go
ListDuplicatePublished() (*[]DuplicatePublishedRow, error)
ClearPublishedFlagExcept(types.ContentID, string, types.ContentVersionID) error
```

AdminContentVersions section (after `PruneAdminOldVersions` at line 178) -- exact admin signatures:

```go
ListAdminDuplicatePublished() (*[]AdminDuplicatePublishedRow, error)
ClearAdminPublishedFlagExcept(types.AdminContentID, string, types.AdminContentVersionID) error
```

**`GetMaxVersionNumberForUpdate` is NOT added to the `DbDriver` interface.** No existing `DbDriver` method takes `*sql.Tx` -- transaction-aware operations are always standalone package-level functions (e.g., `GetContentDataInTx` in `content_data_tx.go`). The sqlc query `GetMaxVersionNumberForUpdate` exists for use via `mdb.New(tx)` inside the standalone `GetMaxVersionNumberInTx` function from Phase 4b.

`internal/db/content_version.go` -- add new result type + SQLite wrapper methods for `ListDuplicatePublished` and `ClearPublishedFlagExcept`:

```go
type DuplicatePublishedRow struct {
    ContentDataID types.ContentID
    Locale        string
    PubCount      int64
}
```

`internal/db/admin_content_version.go` -- parallel admin type + methods.

`internal/remote/driver.go` -- add stubs for all 4 new interface methods (2 public + 2 admin). `RemoteDriver` implements `DbDriver` with 505+ methods and must implement every new method or compilation fails. Follow the existing stub pattern:

```go
func (r *RemoteDriver) ListDuplicatePublished() (*[]db.DuplicatePublishedRow, error) {
    return nil, ErrNotSupported{Method: "ListDuplicatePublished"}
}
func (r *RemoteDriver) ClearPublishedFlagExcept(_ types.ContentID, _ string, _ types.ContentVersionID) error {
    return ErrNotSupported{Method: "ClearPublishedFlagExcept"}
}
// Same pattern for ListAdminDuplicatePublished, ClearAdminPublishedFlagExcept
```

After: Run `just drivergen` to generate MySQL/PostgreSQL variants.

## Phase 4: Atomic Publish Transaction (ClearPublishedFlag + CreateContentVersion)

The clear and create MUST be in the same transaction. If clearing fails, no new version should be created. This requires adding `CreateInTx` to the audited package (mirroring the existing `UpdateInTx` pattern) and transaction-aware helpers for content version operations.

### Step 4a: Add CreateInTx to audited package

`internal/db/audited/audited.go` -- new function after `UpdateInTx` (which spans lines 199-269), inserting at line 270 (before `Delete` at line 276):

`CreateInTx` must be a direct copy of the `Create` function body (`audited.go:29-86`) with these exact changes:

1. **Signature:** `func CreateInTx[T any](cmd CreateCommand[T], tx *sql.Tx) (T, error)` -- takes a `*sql.Tx` instead of calling `cmd.Connection()`.
2. **Remove the `types.WithTransaction` wrapper** (lines 43-77) -- the function body runs directly against the passed-in `tx`, not a new transaction.
3. **Remove the after-hooks block** (lines 81-83) -- after-hooks fire post-commit, and `CreateInTx` does not commit. The caller fires after-hooks post-commit.
4. **Keep everything else verbatim** from `Create`: timeout setup via `context.WithTimeout(ctx, 30*time.Second)` (lines 33-37), `before_create` hooks check (lines 56-60), `cmd.Execute(ctx, tx)` call, `marshalToJSONData(created)` for newValues, and `cmd.Recorder().Record(ctx, tx, ChangeEventParams{...})` with all 11 fields (including `types.NewEventID()`, `types.HLCNow()`, `types.OpInsert`, `types.ActionCreate` constants). Copy the `ChangeEventParams` struct literal from `Create` exactly -- do not abbreviate or guess field names.

Follow the same pattern as `UpdateInTx` (lines 199-269) which also takes `*sql.Tx`, omits after-hooks, and copies the `Update` function body with the transaction wrapper removed.

### Step 4b: Transaction helpers for content version operations

New file: `internal/db/content_version_tx.go` -- following the exact pattern of `content_data_tx.go`:

```go
package db

// GetMaxVersionNumberInTx reads the highest version number within an existing transaction.
// MySQL/PostgreSQL variants use FOR UPDATE to prevent concurrent reads.
func GetMaxVersionNumberInTx(d DbDriver, ctx context.Context, tx *sql.Tx,
    contentDataID types.ContentID, locale string) (int64, error) {
    switch d.(type) {
    case Database:
        queries := mdb.New(tx)
        return queries.GetMaxVersionNumberForUpdate(ctx, mdb.GetMaxVersionNumberForUpdateParams{
            ContentDataID: contentDataID,
            Locale:        locale,
        })
    case MysqlDatabase:
        return getMaxVersionNumberInTxMySQL(ctx, tx, contentDataID, locale)
    case PsqlDatabase:
        return getMaxVersionNumberInTxPsql(ctx, tx, contentDataID, locale)
    default:
        return 0, fmt.Errorf("tx get max version number: unsupported driver type %T", d)
    }
}

// ClearPublishedFlagInTx clears published flags within an existing transaction.
func ClearPublishedFlagInTx(d DbDriver, ctx context.Context, tx *sql.Tx,
    contentDataID types.ContentID, locale string) error {
    switch d.(type) {
    case Database:
        queries := mdb.New(tx)
        return queries.ClearPublishedFlag(ctx, mdb.ClearPublishedFlagParams{
            ContentDataID: contentDataID,
            Locale:        locale,
        })
    case MysqlDatabase:
        return clearPublishedFlagInTxMySQL(ctx, tx, contentDataID, locale)
    case PsqlDatabase:
        return clearPublishedFlagInTxPsql(ctx, tx, contentDataID, locale)
    default:
        return fmt.Errorf("tx clear published flag: unsupported driver type %T", d)
    }
}

// CreateContentVersionInTx creates a content version with audit trail in an existing transaction.
// For MySQL: uses :exec INSERT followed by GetContentVersion SELECT within the same tx,
// matching the pattern in NewContentVersionCmdMysql.Execute (content_version.go:763-782).
func CreateContentVersionInTx(d DbDriver, ctx context.Context, tx *sql.Tx,
    ac audited.AuditContext, params CreateContentVersionParams) (*ContentVersion, error) {
    switch drv := d.(type) {
    case Database:
        cmd := Database{}.NewContentVersionCmd(ctx, ac, params)
        result, err := audited.CreateInTx(cmd, tx)
        if err != nil {
            return nil, fmt.Errorf("tx create content version: %w", err)
        }
        r := drv.MapContentVersion(result)
        return &r, nil
    case MysqlDatabase:
        return createContentVersionInTxMySQL(drv, ctx, tx, ac, params)
    case PsqlDatabase:
        return createContentVersionInTxPsql(drv, ctx, tx, ac, params)
    default:
        return nil, fmt.Errorf("tx create content version: unsupported driver type %T", d)
    }
}
```

New file: `internal/db/content_version_tx_mysql.go` -- MySQL variants (same pattern as `content_data_tx_mysql.go`). For `createContentVersionInTxMySQL`: follow the same pattern as `updateContentDataInTxMySQL` (`content_data_tx_mysql.go:23-25`) -- construct the MySQL command struct via `MysqlDatabase{}.NewContentVersionCmd(ctx, ac, params)`, then call `audited.CreateInTx(cmd, tx)`, then map the result via `MysqlDatabase{}.MapContentVersion()`. The `*sql.Tx` satisfies the `audited.DBTX` interface, so the MySQL command's `Execute` method (which uses `mdbm.New(tx)` for both `:exec` INSERT and follow-up `GetContentVersion` SELECT) works correctly within the caller's transaction. Note: `GetMaxVersionNumber` returns `int32` on MySQL (sqlc generates `INT` as `int32`); the `getMaxVersionNumberInTxMySQL` must widen to `int64` via `int64(result)`.

New file: `internal/db/content_version_tx_psql.go` -- PostgreSQL variants. Same pattern as MySQL: use `PsqlDatabase{}.NewContentVersionCmd(ctx, ac, params)` + `audited.CreateInTx(cmd, tx)` + `PsqlDatabase{}.MapContentVersion()`. Same `int64()` widening for `GetMaxVersionNumber` result.

Admin parallels:
- `GetAdminMaxVersionNumberInTx` in same file
- `ClearAdminPublishedFlagInTx` in same file
- `CreateAdminContentVersionInTx` in same file
- MySQL/PostgreSQL admin variants in the same `_tx_mysql.go` / `_tx_psql.go` files

All functions must include a `default` case returning an error (matching `content_data_tx.go` pattern).

### Step 4c: Rewrite publish to use atomic transaction

`internal/publishing/publishing.go` -- replace lines 231-258 (steps 5-7) with a single transaction that reads the max version number, clears published flags, and creates the new version atomically. Use the request-scoped `ctx` (not `d.Context` from `GetConnection`) so the transaction respects request cancellation. Use `db.WithTransactionResult` from `internal/db/transaction.go` (not `types.WithTransactionResult`). `publishing.go` already imports the `db` package.

```go
// 5+6+7. Get next version, clear published flag, and create new version atomically.
// All three operations MUST be in the same transaction. If any step fails, the
// transaction rolls back and no partial state is left.
conn, _, connErr := d.GetConnection()
if connErr != nil {
    return nil, fmt.Errorf("get connection for publish tx: %w", connErr)
}
now := types.TimestampNow()
version, err := db.WithTransactionResult(ctx, conn, func(tx *sql.Tx) (*db.ContentVersion, error) {
    // Read max version number inside the transaction.
    // MySQL/PostgreSQL use FOR UPDATE to prevent concurrent reads.
    maxVersion, maxErr := db.GetMaxVersionNumberInTx(d, ctx, tx, rootID, locale)
    if maxErr != nil {
        return nil, fmt.Errorf("get max version number: %w", maxErr)
    }
    nextVersion := maxVersion + 1

    // Clear published flag within the transaction.
    if clearErr := db.ClearPublishedFlagInTx(d, ctx, tx, rootID, locale); clearErr != nil {
        return nil, fmt.Errorf("clear published flag: %w", clearErr)
    }

    // Create new content version within the same transaction.
    ver, createErr := db.CreateContentVersionInTx(d, ctx, tx, ac, db.CreateContentVersionParams{
        ContentDataID: rootID,
        VersionNumber: nextVersion,
        Locale:        locale,
        Snapshot:      string(snapshotBytes),
        Trigger:       "publish",
        Label:         "",
        Published:     true,
        PublishedBy:   types.NullableUserID{ID: userID, Valid: true},
        DateCreated:   now,
    })
    if createErr != nil {
        return nil, fmt.Errorf("create content version: %w", createErr)
    }
    return ver, nil
})
if err != nil {
    return nil, err
}
```

Remove the standalone step 5 (`GetMaxVersionNumber` call at lines 231-236) and step 6 (`ClearPublishedFlag` at lines 238-241) and step 7 (`CreateContentVersion` at lines 243-258). Also move `now := types.TimestampNow()` before the transaction (it was previously between steps 6 and 7 at line 244).

**Step 8 (`UpdateContentDataPublishMeta`) remains outside the transaction.** This is an acceptable gap: if the transaction commits but step 8 fails, the published version is served correctly to clients but admin metadata (status, published_at) is stale. The next publish or manual status update will correct it. Including step 8 in the transaction would widen the lock scope across two tables (`content_versions` + `content_data`) and increase contention for a low-risk metadata field.

`internal/publishing/publishing_admin.go` -- same transformation for `PublishAdminContent` (lines 215-242), using `GetAdminMaxVersionNumberInTx`, `ClearAdminPublishedFlagInTx`, and `CreateAdminContentVersionInTx`. Note: `PublishAdminContent` returns `error` (not `*ContentVersion`). Use `db.WithTransaction` (not `db.WithTransactionResult`) since the admin publish does not return the created version to the caller, or use `WithTransactionResult` and discard the result.

## Phase 5: Content Heal -- New Pass for Duplicate Published Versions

`internal/service/content_heal.go`:

New report type:

```go
type DuplicatePublishedReport struct {
    ContentDataID string `json:"content_data_id"`
    Locale        string `json:"locale"`
    Count         int    `json:"count"`
    KeptVersionID string `json:"kept_version_id"`
    Repaired      bool   `json:"repaired"`
}
```

Add to `HealReport`:

```go
VersionsScanned    int                        `json:"versions_scanned"`
DuplicatePublished []DuplicatePublishedReport `json:"duplicate_published"`
```

New Pass 11 in `Heal()` after existing Pass 10:
- Call `s.driver.ListDuplicatePublished()` to find violations
- For each violation, call `ListContentVersionsByContentLocale` to get all versions sorted by `version_number DESC`. `ListContentVersionsByContentLocale` returns ALL versions (published and unpublished). Filter the result slice in Go to keep only rows where `Published == true`, then keep the first element (highest version_number) as the version to preserve.
- Call `ClearPublishedFlagExcept` with the kept version's ID to clear all others (unless `dry_run`)
- Append to `report.DuplicatePublished`

**`ClearPublishedFlagExcept` is intentionally unaudited.** This is a data repair operation fixing invariant violations, not a user-initiated content change. The heal report itself serves as the audit record. This is consistent with other heal passes that use bare `:exec` queries for structural repairs (e.g., nulling orphaned route refs).

`internal/service/content_heal_admin.go`:
1. Add `VersionsScanned int` (`json:"versions_scanned"`) and `DuplicatePublished []DuplicatePublishedReport` (`json:"duplicate_published"`) fields to `AdminHealReport`.
2. Reuse the `DuplicatePublishedReport` type from `content_heal.go` (it uses plain `string` fields, not admin-typed IDs, matching how the JSON response works).
3. Add Pass 10 after the existing Pass 9 (`healAdminRootlessContent`), following the same logic as public Pass 11: call `ListAdminDuplicatePublished`, then for each violation call `ListAdminContentVersionsByContentLocale` and filter for `Published == true` rows in Go, keep the highest `version_number`, then call `ClearAdminPublishedFlagExcept` (unless `dry_run`).
4. Initialize the new slice fields in the `AdminHealReport` constructor.

## Phase 6: SDK Type Updates

### Go SDK (`sdks/go/content_heal.go`)

The Go SDK's `HealReport` currently has 5 fields. The server's `HealReport` (`internal/service/content_heal.go:14-28`) has 13 fields. This phase brings the Go SDK fully in sync with the server, not just adding the 2 new fields.

Add the following missing report types (matching the server structs in `internal/service/content_heal.go`):

```go
type MissingFieldReport struct { ... }
type DuplicateFieldReport struct { ... }
type OrphanedFieldReport struct { ... }
type DanglingPointerReport struct { ... }
type OrphanedRouteReport struct { ... }
type UnroutedRootReport struct { ... }
type RootlessContentReport struct { ... }
type InvalidUserRefReport struct { ... }
type DuplicatePublishedReport struct { ... }
```

Update `HealReport` to match all server fields:

```go
type HealReport struct {
    DryRun              bool                       `json:"dry_run"`
    ContentDataScanned  int                        `json:"content_data_scanned"`
    ContentDataRepairs  []HealRepair               `json:"content_data_repairs"`
    ContentFieldScanned int                        `json:"content_field_scanned"`
    ContentFieldRepairs []HealRepair               `json:"content_field_repairs"`
    MissingFields       []MissingFieldReport       `json:"missing_fields"`
    DuplicateFields     []DuplicateFieldReport     `json:"duplicate_fields"`
    OrphanedFields      []OrphanedFieldReport      `json:"orphaned_fields"`
    DanglingPointers    []DanglingPointerReport    `json:"dangling_pointers"`
    OrphanedRouteRefs   []OrphanedRouteReport      `json:"orphaned_route_refs"`
    UnroutedRoots       []UnroutedRootReport       `json:"unrouted_roots"`
    RootlessContent     []RootlessContentReport    `json:"rootless_content"`
    InvalidUserRefs     []InvalidUserRefReport     `json:"invalid_user_refs"`
    VersionsScanned     int                        `json:"versions_scanned"`
    DuplicatePublished  []DuplicatePublishedReport `json:"duplicate_published"`
}
```

Copy exact field names and JSON tags from the server structs. Each report type struct fields must match the server's JSON output.

### TypeScript Admin SDK (`sdks/typescript/modulacms-admin-sdk/src/types/content.ts`)

The TypeScript SDK's `HealReport` currently has 7 fields (missing `orphaned_fields`, `dangling_pointers`, `orphaned_route_refs`, `unrouted_roots`, `rootless_content`, `invalid_user_refs`). Bring it fully in sync with the server.

Add the following NEW report types. `MissingFieldReport` and `DuplicateFieldReport` already exist in this file (lines 186-205) -- do not duplicate them.

Fields copied exactly from server structs in `internal/service/content_heal.go`:

```typescript
export type OrphanedFieldReport = { content_field_id: string; content_data_id: string; field_id: string; deleted: boolean }
export type DanglingPointerReport = { content_data_id: string; column: string; target_id: string; nulled: boolean }
export type OrphanedRouteReport = { content_data_id: string; route_id: string; nulled: boolean }
export type UnroutedRootReport = { content_data_id: string; datatype_id: string; datatype_name: string }
export type RootlessContentReport = { content_data_id: string; route_id: string; route_slug: string; datatype_name: string; deleted: boolean }
export type InvalidUserRefReport = { table: string; row_id: string; column: string; old_value: string; new_value: string; repaired: boolean }
export type DuplicatePublishedReport = { content_data_id: string; locale: string; count: number; kept_version_id: string; repaired: boolean }
```

Add all missing fields to `HealReport`:

```typescript
orphaned_fields: OrphanedFieldReport[]
dangling_pointers: DanglingPointerReport[]
orphaned_route_refs: OrphanedRouteReport[]
unrouted_roots: UnroutedRootReport[]
rootless_content: RootlessContentReport[]
invalid_user_refs: InvalidUserRefReport[]
versions_scanned: number
duplicate_published: DuplicatePublishedReport[]
```

Copy exact field names from the server's JSON tags. Verify each report type's fields against the server struct before writing.

## Implementation Order

| Step | Action | Depends On | Command After |
|------|--------|------------|---------------|
| 1 | Edit 6 SQL files (Phase 1+2: ORDER BY, GetMaxVersionNumberForUpdate, ListDuplicatePublished, ClearPublishedFlagExcept) | -- | `just sqlc` |
| 2 | DbDriver interface (4 new methods) + wrapper methods + RemoteDriver stubs (Phase 3) | Step 1 | `just drivergen` |
| 3 | `audited.CreateInTx` (Phase 4a) | -- | `just check` |
| 4 | Transaction helpers + publish rewrite (Phase 4b-c) | Steps 2+3 | `just check` |
| 5 | Heal passes (Phase 5) | Step 2 | `just check` |
| 6 | SDK types (Phase 6: full HealReport sync, not just new fields) | Step 5 | `just sdk ts build && just sdk go vet` |
| 7 | Full test run + concurrency test | Steps 4-6 | `just test` |

Steps 3 and 1-2 are independent. Steps 4 and 5 are independent after their deps.

## Key Files

| File | Change |
|------|--------|
| `sql/schema/31_content_versions/queries*.sql` (3) | ORDER BY + 3 new queries (GetMaxVersionNumberForUpdate, ListDuplicatePublished, ClearPublishedFlagExcept) |
| `sql/schema/32_admin_content_versions/queries*.sql` (3) | ORDER BY + 3 new queries (admin variants) |
| `internal/db/repositories.go` | 4 new interface methods (2 public + 2 admin; `GetMaxVersionNumberForUpdate` is NOT an interface method) |
| `internal/db/content_version.go` | `DuplicatePublishedRow` type + 2 wrapper methods + `GetMaxVersionNumberForUpdate` sqlc wrapper |
| `internal/db/admin_content_version.go` | `AdminDuplicatePublishedRow` type + 2 wrapper methods + admin `GetMaxVersionNumberForUpdate` |
| `internal/remote/driver.go` | 4 new `ErrNotSupported` stubs for `RemoteDriver` (must implement all `DbDriver` methods) |
| `internal/db/audited/audited.go` | `CreateInTx` function |
| `internal/db/content_version_tx.go` (new) | `GetMaxVersionNumberInTx` + `ClearPublishedFlagInTx` + `CreateContentVersionInTx` + admin variants |
| `internal/db/content_version_tx_mysql.go` (new) | MySQL transaction variants |
| `internal/db/content_version_tx_psql.go` (new) | PostgreSQL transaction variants |
| `internal/publishing/publishing.go` | Atomic get-version+clear+create transaction (steps 5-7 merged) |
| `internal/publishing/publishing_admin.go` | Atomic get-version+clear+create transaction |
| `internal/service/content_heal.go` | `DuplicatePublishedReport` + Pass 11 |
| `internal/service/content_heal_admin.go` | Same for admin + Pass 10 |
| `sdks/go/content_heal.go` | Full `HealReport` sync: 8 missing report types + 9 missing fields + `DuplicatePublishedReport` |
| `sdks/typescript/modulacms-admin-sdk/src/types/content.ts` | Full `HealReport` sync: 7 missing report types + 8 missing fields + `DuplicatePublishedReport` |

## Verification

1. `just sqlc` succeeds (SQL is valid across all dialects). After running, verify the generated `ListDuplicatePublishedRow` struct in each `db-sqlite`/`db-mysql`/`db-psql` package. The `pub_count` column alias should generate a `PubCount` field. If sqlc names it differently, update the wrapper mapping in `content_version.go` accordingly.
2. `just drivergen` succeeds (wrappers replicate cleanly)
3. `just check` compiles (interface satisfied by all three drivers including `RemoteDriver`)
4. `just test` passes (existing tests unbroken)
5. Write concurrency test in `internal/publishing/publishing_test.go`: spawn 10 goroutines each calling `PublishContent` for the same content+locale concurrently. After all complete, verify exactly one `content_versions` row has `published=true` for that content+locale, and all version numbers are unique (no duplicates). Use `sync.WaitGroup` and a file-backed SQLite database (not `:memory:`) with `PRAGMA journal_mode=WAL` and `PRAGMA busy_timeout=5000` to allow concurrent access. Check existing test files in `internal/publishing/` and `internal/service/` for test setup patterns (driver initialization, mock dispatcher, mock indexer). If no test file exists, follow the patterns in `internal/service/content_heal_test.go` for driver setup.
6. Write heal test in `internal/service/content_heal_test.go`: directly INSERT two `content_versions` rows with `published=1` for the same content+locale (bypassing the publish flow), run `Heal(dryRun=false)`, verify only one remains published and the heal report contains the violation.

## Post-Deploy Verification (manual, not for implementing agent)

7. Manually verify deployed fix: unpublish v11 on production, then deploy new binary

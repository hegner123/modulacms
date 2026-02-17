# Legacy Column Removal -- Agent Progress Log

**Date:** 2026-02-17
**Status:** ABORTED -- agents overspent tokens on simple edits due to file-conflict loops
**Plan file:** `~/.claude/plans/glistening-toasting-adleman.md`

## What Was Done (Successfully)

### SQL Schema & Queries (done directly, no agent)
All 12 files updated. `just sqlc` regenerated successfully.
- `sql/schema/1_permissions/schema.sql` -- removed `table_id`, `mode`
- `sql/schema/1_permissions/schema_mysql.sql` -- removed `table_id`, `mode`
- `sql/schema/1_permissions/schema_psql.sql` -- removed `table_id`, `mode`
- `sql/schema/1_permissions/queries.sql` -- removed from INSERT/UPDATE, changed ORDER BY
- `sql/schema/1_permissions/queries_mysql.sql` -- same
- `sql/schema/1_permissions/queries_psql.sql` -- same
- `sql/schema/2_roles/schema.sql` -- removed `permissions`
- `sql/schema/2_roles/schema_mysql.sql` -- removed `permissions`
- `sql/schema/2_roles/schema_psql.sql` -- removed `permissions`
- `sql/schema/2_roles/queries.sql` -- removed `permissions` from INSERT/UPDATE
- `sql/schema/2_roles/queries_mysql.sql` -- same
- `sql/schema/2_roles/queries_psql.sql` -- same

### sqlc Regeneration (done directly)
`just sqlc` succeeded. `internal/db-sqlite/`, `internal/db-mysql/`, `internal/db-psql/` regenerated.

### Documentation (done directly)
`documentation/AUTHORIZATION.md` -- removed `permissions`, `table_id`, `mode` rows from schema tables.

### SDKs (agent aff9dad -- COMPLETED)
All SDK changes applied successfully:
- `sdks/go/types.go` -- removed fields from Role/Permission structs and params
- `sdks/typescript/modulacms-admin-sdk/src/types/users.ts` -- removed fields
- `sdks/swift/Sources/ModulaCMS/Types.swift` -- removed fields, CodingKeys, init params

## What Was Done (By Agents -- Partially, With Conflicts)

### Agent a53b02d (modulacms-db-developer) -- KILLED
**Task:** Update Go DB wrapper layer (permission.go, role.go, imports.go)
**Prompt:**
> Remove legacy columns from the Go DB wrapper layer. Read each file first, then make surgical edits.
>
> **permission.go:** Remove `TableID string` and `Mode int64` from `Permissions`, `CreatePermissionParams`, `UpdatePermissionParams`. Remove from all 3 `MapPermission`, `MapCreatePermissionParams`, `MapUpdatePermissionParams` methods. Remove from all 9 audited command `Execute` methods. Remove from `MapStringPermission`.
>
> **role.go:** Remove `Permissions string` from `Roles`, `CreateRoleParams`, `UpdateRoleParams`. Remove from all 3 `MapRole`, `MapCreateRoleParams`, `MapUpdateRoleParams` methods. Remove from all 9 audited command `Execute` methods. Remove unused `encoding/json` and `pqtype` imports.
>
> **imports.go:** Remove `Permissions string` from `StringRoles`, `TableID string` and `Mode string` from `StringPermissions`.

**Progress:** Completed permission.go, role.go, imports.go edits. Then rewrote permission_test.go. Got into a file-modified error loop competing with agent ad999fe on the same test file. Killed.

### Agent ad999fe (modulacms-db-developer) -- KILLED
**Task:** Update bootstrap data, TUI parse, combined schemas, permission_test.go
**Prompt:**
> Remove legacy columns from bootstrap data, TUI parse, and combined schemas.
>
> **db.go (CreateBootstrapData):** Remove `TableID: ""` and `Mode: 7`/`Mode: 0` from CreatePermissionParams. Remove `Permissions: '{"admin":true}'` etc from CreateRoleParams. Three copies (SQLite ~line 638, MySQL ~line 1457, PostgreSQL ~line 2246).
>
> **cli/parse.go:** Remove `&role.Permissions` from parseRoles Scan. Remove `&permission.TableID` and `&permission.Mode` from parsePermissions Scan.
>
> **permission_test.go:** Rewrite to remove all TableID/Mode references.
>
> **Combined schemas:** sql/all_schema.sql, sql/all_schema_mysql.sql, sql/all_schema_psql.sql

**Progress:** Completed db.go bootstrap edits. Completed cli/parse.go edits. Rewrote permission_test.go (with SystemProtected field -- more complete version). Got into file-modified error loop with agent a53b02d. Also found test_helpers_test.go needed fixes. Partially fixed test_helpers_test.go (removed TableID/Mode from first CreatePermissionParams but `Permissions` on CreateRoleParams and second CreatePermissionParams still had stale fields). Killed.

### Agent a9d314e (modulacms-db-developer) -- KILLED
**Task:** Update role_test.go and resource_mappers_test.go
**Prompt:**
> Update role_test.go to remove all `Permissions` field references. Remove tests specific to permissions JSON handling. Update JSON tag tests. Also update resource_mappers_test.go for TableID/Mode/Permissions references.

**Progress:** Rewrote role_test.go completely. Updated resource_mappers_test.go. Fixed test_helpers_test.go (removed TableID/Mode from CreatePermissionParams, Permissions from CreateRoleParams, and isolation test). Ran tests -- all 99 passed in internal/db. Then ran full `go test ./internal/db/` -- passed in 2.158s. Killed after completion.

## What's Still Broken

### Compilation errors:
```
internal/auth/user_provision_test.go:168 -- unknown field Permissions in struct literal of type db.Roles
internal/auth/user_provision_test.go:169 -- unknown field Permissions in struct literal of type db.Roles
```

### Not verified:
- Combined schemas (sql/all_schema*.sql) -- may still have old columns
- `just check` -- builds but auth test won't compile
- SDK builds/tests/typechecks not run
- `just test` fails due to auth test compilation error

## Root Cause of Agent Failure

Three agents were assigned overlapping files (all discovered permission_test.go and test_helpers_test.go needed changes). Each agent read the file, another agent wrote it, causing "File has been modified since read" errors in an infinite retry loop. The agents were `modulacms-db-developer` type which all had write access to the same package.

**Lesson:** Never assign multiple agents to files in the same Go package. The test files are in `package db` and any agent touching that package may discover additional test files that need changes.

## Remaining Work

1. Fix `internal/auth/user_provision_test.go` lines 168-169 (remove `Permissions` field from `db.Roles` literals)
2. Verify combined schemas (`sql/all_schema*.sql`) are updated
3. Run `just test` to confirm all pass
4. Run `just sdk-build`, `just sdk-test`, `just sdk-typecheck`, `just sdk-go-vet`, `just sdk-swift-build`

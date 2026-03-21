# Validation Configuration Table Plan

## Overview

Extract inline validation JSON from `fields.validation` into a standalone `validations` table (and `admin_validations`). Fields reference validation configs by ID. Multiple fields can share the same reusable config.

## Design Decision

**Replace `fields.validation` TEXT column with `fields.validation_id` FK column.**

Greenfield project, no users, no migration needed. The inline JSON column is removed entirely. Fields reference a validation config by ID or have no validation (NULL).

## New Table: `validations`

| Column | Type | Notes |
|--------|------|-------|
| `validation_id` | TEXT PK (ULID) | 26-char typed ID |
| `name` | TEXT NOT NULL | Human-readable ("Email rules", "Password strength") |
| `description` | TEXT NOT NULL | Optional longer description |
| `config` | TEXT NOT NULL | JSON validation config (same schema as current inline) |
| `author_id` | TEXT FK nullable | References users |
| `date_created` | TEXT | Timestamp |
| `date_modified` | TEXT | Timestamp |

Parallel `admin_validations` table with `admin_validation_id` PK.

## Phases

### Phase 1: SQL Schema and Code Generation ✅

1. **Typed IDs** — Add `ValidationID`, `AdminValidationID`, plus nullable wrappers to `internal/db/types/`
2. **Schema directories** — Create `sql/schema/38_validations/` and `sql/schema/39_admin_validations/` (6 files each)
3. **Alter fields tables** — Remove `validation TEXT` column, add `validation_id TEXT DEFAULT NULL REFERENCES validations(validation_id) ON DELETE SET NULL` in `sql/schema/8_fields/` and `sql/schema/10_admin_fields/` (all 3 dialects). Update all queries that read/write the `validation` column.
4. **sqlc overrides** — Add type mappings in `tools/sqlcgen/definitions.go`
5. **Run `just sqlc`**
6. **dbgen entity definitions** — Add Validations/AdminValidations entities, update Fields/AdminFields to replace `Validation string` with `ValidationID NullableValidationID` in `tools/dbgen/definitions.go`
7. **Run `just dbgen`**

### Phase 2: DbDriver Interface and Registration ✅

1. **ValidationRepository interface** — Standard CRUD plus `ListValidationsByName` (search)
2. **Embed in DbDriver** — Add to interface in `internal/db/db.go`
3. **consts.go** — Add `ValidationT` / `Admin_validationT` DBTable constants, allTables, TableStructMap, CastToTypedSlice
4. **CreateAllTables** — Validations BEFORE Fields (FK dependency). DropAllTables in reverse.
5. **Regenerate combined schemas** — `sql/schema/read_sql.sh`
6. **GenericHeaders/GenericList** — Add cases + test entries

### Phase 3: Service Layer ✅

1. **ValidationService** (new file `internal/service/validations.go`)
   - `CreateValidation` — validates name non-empty, validates config JSON via `types.ValidateValidationConfig`
   - `UpdateValidation` — same validation
   - `DeleteValidation` — ON DELETE SET NULL handles FK; UI should warn
   - `GetValidation` / `ListValidations` / `ListValidationsPaginated`
2. **Update SchemaService** — `CreateField`/`UpdateField` accept `validation_id`, validate it exists in validations table
3. **Update content field validation** — `content_fields.go` resolves config by fetching the `Validations` row via `validation_id`. NULL `validation_id` means no validation.
4. **Register in service.Registry**

### Phase 4: REST API ✅

1. **CRUD handlers** — `internal/router/validations.go`
   - `GET /api/v1/validations` (list + pagination)
   - `GET /api/v1/validations/{id}`
   - `POST /api/v1/validations`
   - `PUT /api/v1/validations/{id}`
   - `DELETE /api/v1/validations/{id}`
   - `GET /api/v1/validations/search?name=...`
   - Same for admin at `/api/v1/admin/validations/...`
2. **Routes** — Register in `mux.go` with permission middleware
3. **RBAC** — Add `validations:read/create/update/delete` to bootstrap permissions

### Phase 5: Admin Panel ✅

1. **Validation CRUD pages**
   - `validations_list.templ` — data table of all configs
   - `validation_create.templ` — name, description, `<mcms-validation-wizard>` for config JSON
   - `validation_detail.templ` — edit form, shows list of fields referencing this config
2. **Handlers** — `internal/admin/handlers/validations.go` (list, create page, create, detail, update, delete)
3. **Routes** — `/admin/schema/validations`, `/admin/schema/validations/new`, `/admin/schema/validations/{id}`
4. **Update field pages** — Replace inline wizard with dropdown select of existing validation configs + "Create new" link
5. **Update field handlers** — Read `validation_id` from form, pass validations list to templates
6. **Sidebar nav** — Add "Validations" under Schema section

### Phase 6: Cleanup ✅ (completed during Phase 1 rename sweep)

1. **Remove `Validation string` field** from Fields/AdminFields structs, CreateFieldParams, UpdateFieldParams
2. **Remove inline validation parsing** — delete `ParseValidationConfig` calls from field handlers and service methods that operated on the inline string
3. **Remove `validation` column references** from all SQL queries
4. **Update existing tests** that set `Validation` on field params to use `ValidationID` instead

### Phase 7: TUI ✅

- `screen_validations.go` — list/detail/create/edit
- Update field screens to show validation picker instead of raw JSON

### Phase 8: SDK Updates ✅

- **TypeScript:** Add `Validation` type to `@modulacms/types`, CRUD to admin SDK, replace `validation: string` with `validation_id?: string` on Fields type
- **Go:** Add `Validation` struct, `ValidationID` branded type, CRUD resource, update Fields struct
- **Swift:** Add `Validation` struct, `ValidationID`, CRUD resource, update Fields struct

## Challenges

1. **Table creation order** — Validations must be created before Fields (FK dependency). `CreateAllTables()` must respect order.
2. **Cascade on delete** — ON DELETE SET NULL means fields lose their validation reference. UI should warn before deleting a validation that has referencing fields.
3. **Runtime performance** — Extra DB read to resolve `validation_id` at content save time. Acceptable; cache in service layer later if needed.

## File Impact Summary

| Area | Files |
|------|-------|
| Typed IDs | `internal/db/types/types_ids.go`, `types_nullable_ids.go` |
| SQL schema | `sql/schema/38_validations/` (6 new), `sql/schema/39_admin_validations/` (6 new), modify `8_fields/` (6), `10_admin_fields/` (6) |
| Code gen config | `tools/sqlcgen/definitions.go`, `tools/dbgen/definitions.go` |
| Generated code | `internal/db/validation_gen.go`, `admin_validation_gen.go`, updated `field_gen.go`, `admin_field_gen.go` |
| DbDriver | `internal/db/db.go`, `internal/db/consts.go`, `internal/db/generic_list.go` |
| Service | `internal/service/validations.go` (new), `schema.go`, `content_fields.go`, `registry.go` |
| API | `internal/router/validations.go` (new), `mux.go` |
| Admin handlers | `internal/admin/handlers/validations.go` (new), `fields.go` |
| Admin pages | `pages/validations_list.templ`, `validation_create.templ`, `validation_detail.templ` (new), update `field_create.templ`, `field_detail.templ`, `admin_field_detail.templ` |
| Nav | `internal/admin/components/nav.go` |
| SDKs | TypeScript types + admin SDK, Go SDK, Swift SDK |

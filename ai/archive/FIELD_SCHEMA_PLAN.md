# Plan: Field Config Separation & Content Relations Tables

> **Status: COMPLETED** — The NullableContentID → NullableDatatypeID/NullableAdminDatatypeID migrations described here have been applied. This document is retained for historical reference.

## Context

The `fields.data` column currently serves as a catch-all JSON blob for type-specific config, validation rules, UI hints, and relation metadata. This overloads a single column with multiple concerns. Additionally, relation fields store target content IDs as serialized strings in `content_fields.field_value`, preventing indexed reverse lookups and referential integrity.

This plan separates field configuration into dedicated columns (`data` for type-specific config, `validation` for rules, `ui_config` for TUI/UI hints) and adds proper junction tables (`content_relations`, `admin_content_relations`) for relational content references with FK constraints and indexed lookups.

**Scope:** Both public (`fields`, `content_relations`) and admin (`admin_fields`, `admin_content_relations`) systems. Public and admin are fully independent — no cross-references between them.

---

## Design Decisions

### Typed config structs and their consumer

Step 3 introduces typed Go structs (`RelationConfig`, `ValidationConfig`, `UIConfig`, `Cardinality`) in `internal/db/types/types_field_config.go`. These are **not** used by the db wrapper layer — the wrapper stores `validation` and `ui_config` as raw `string` (matching the existing `data string` pattern). The typed structs exist as canonical parse targets for the **transform package** (`internal/transform/`), which uses them during relational tree construction to interpret `fields.data` (for `RelationConfig`) and the new `validation`/`ui_config` columns (for `ValidationConfig`/`UIConfig`). Parse functions in the types package convert the raw strings into these structs at the point of use.

### Column storage: raw string, not custom type

The wrapper structs store `Validation string` and `UIConfig string` (not custom types) to match the existing pattern for `Data string`. JSON validity is enforced at the types package layer via `ParseValidationConfig()` and `ParseUIConfig()`, which return parse errors for malformed input. The db wrapper layer is a passthrough — validation belongs at the boundary where data enters the system (router handlers, TUI dialogs, import handlers).

### EmptyJSON constant

All call sites use `types.EmptyJSON` (value `"{}"`) instead of string literals. This prevents typos and makes the default semantically explicit.

### No DEFAULT on new columns

New columns use `TEXT NOT NULL` with no SQL DEFAULT, matching the existing `data TEXT NOT NULL` pattern. The application always provides values explicitly via `types.EmptyJSON` for creates.

### ON DELETE RESTRICT error handling

`content_relations.field_id` and `admin_content_relations.admin_field_id` use `ON DELETE RESTRICT`. When a user attempts to delete a field that has active relations, the database returns a foreign key constraint error. The TUI layer and router handlers must catch this error and present a clear message: "Cannot delete field: active content relations exist. Remove all relations for this field first." This is handled at the call site where `DeleteField` / `DeleteAdminField` is invoked — the existing error propagation pattern (`fmt.Errorf("failed to delete field: %w", err)`) carries the constraint error up to the TUI/router where it can be displayed.

### Self-reference prevention

Both `content_relations` and `admin_content_relations` include `CHECK (source_content_id != target_content_id)` to prevent content from relating to itself. Cycle detection (A->B->A) is not enforced at the schema level — it is handled by the transform package during tree construction, which already has cycle detection in `BuildNodes`.

### Pattern field and ReDoS

`ValidationConfig.Pattern` stores a regex string for end-user field validation. The transform package should compile patterns with a timeout or complexity limit before applying them. This is outside the scope of this plan but noted as a requirement for the validation execution layer.

### Interface consistency notes

The existing `DbDriver` interface returns `*[]T` (pointer to slice) and omits `context.Context` on read operations. Both are non-idiomatic Go but consistent across the entire interface. New methods follow the existing convention to avoid interface inconsistency. A future refactor should address both codebase-wide.

---

## Implementation Steps

Steps are ordered by dependency.

### Step 1: Add ID types and EmptyJSON constant

**File:** `internal/db/types/types_ids.go`

Add two new ID types following the existing pattern (see `ContentFieldID` as reference):

- `ContentRelationID` — for `content_relations.content_relation_id`
- `AdminContentRelationID` — for `admin_content_relations.admin_content_relation_id`

Each needs: `NewXxxID()`, `String()`, `IsZero()`, `Validate()`, `ULID()`, `Time()`, `Value()`, `Scan()`, `MarshalJSON()`, `UnmarshalJSON()`, `ParseXxxID()`.

Nullable variants are not needed — these are primary keys, never used as foreign keys elsewhere.

**File:** `internal/db/types/types_field_config.go` (add constant at top)

```go
// EmptyJSON is the default value for JSON columns that have no configuration.
// Use this instead of raw "{}" string literals at call sites.
const EmptyJSON = "{}"
```

### Step 2: Add FieldConfig Go types

**File (new):** `internal/db/types/types_field_config.go`

These typed structs are parse targets for the transform package, which uses them during relational tree construction. The db wrapper layer stores these as raw strings; parsing happens at the point of use.

```go
// Cardinality constrains how many targets a relation field allows.
// Enforced by the transform package during tree construction, not at the schema level.
type Cardinality string

const (
    CardinalityOne  Cardinality = "one"
    CardinalityMany Cardinality = "many"
)

func (c Cardinality) Validate() error {
    switch c {
    case CardinalityOne, CardinalityMany:
        return nil
    default:
        return fmt.Errorf("Cardinality: invalid value %q (must be %q or %q)", c, CardinalityOne, CardinalityMany)
    }
}

// RelationConfig lives in fields.data when field type is "relation".
// Parsed by the transform package to resolve relation fields during tree construction.
type RelationConfig struct {
    TargetDatatypeID DatatypeID  `json:"target_datatype_id"`
    Cardinality      Cardinality `json:"cardinality"`
    MaxDepth         *int        `json:"max_depth,omitempty"`
}

// ValidationConfig lives in fields.validation.
// All numeric constraint fields use *int so zero is distinguishable from "not set".
type ValidationConfig struct {
    Required  bool   `json:"required,omitempty"`
    MinLength *int   `json:"min_length,omitempty"`
    MaxLength *int   `json:"max_length,omitempty"`
    Min       *int   `json:"min,omitempty"`
    Max       *int   `json:"max,omitempty"`
    Pattern   string `json:"pattern,omitempty"`
    MaxItems  *int   `json:"max_items,omitempty"`
}

// UIConfig lives in fields.ui_config.
type UIConfig struct {
    Widget      string `json:"widget,omitempty"`
    Placeholder string `json:"placeholder,omitempty"`
    HelpText    string `json:"help_text,omitempty"`
    Hidden      bool   `json:"hidden,omitempty"`
}
```

Add parse functions that return typed structs from raw JSON strings:

```go
// ParseValidationConfig parses a JSON string into a ValidationConfig.
// Returns an empty ValidationConfig (not an error) for EmptyJSON input.
func ParseValidationConfig(s string) (ValidationConfig, error) {
    if s == "" || s == EmptyJSON {
        return ValidationConfig{}, nil
    }
    var vc ValidationConfig
    if err := json.Unmarshal([]byte(s), &vc); err != nil {
        return ValidationConfig{}, fmt.Errorf("ParseValidationConfig: %w", err)
    }
    return vc, nil
}

// ParseUIConfig parses a JSON string into a UIConfig.
// Returns an empty UIConfig (not an error) for EmptyJSON input.
func ParseUIConfig(s string) (UIConfig, error) {
    if s == "" || s == EmptyJSON {
        return UIConfig{}, nil
    }
    var uc UIConfig
    if err := json.Unmarshal([]byte(s), &uc); err != nil {
        return UIConfig{}, fmt.Errorf("ParseUIConfig: %w", err)
    }
    return uc, nil
}

// ParseRelationConfig parses a JSON string into a RelationConfig.
// Returns an error if the JSON is invalid or required fields are missing.
func ParseRelationConfig(s string) (RelationConfig, error) {
    if s == "" || s == EmptyJSON {
        return RelationConfig{}, fmt.Errorf("ParseRelationConfig: relation fields require a non-empty data config")
    }
    var rc RelationConfig
    if err := json.Unmarshal([]byte(s), &rc); err != nil {
        return RelationConfig{}, fmt.Errorf("ParseRelationConfig: %w", err)
    }
    if rc.TargetDatatypeID.IsZero() {
        return RelationConfig{}, fmt.Errorf("ParseRelationConfig: target_datatype_id is required")
    }
    if err := rc.Cardinality.Validate(); err != nil {
        return RelationConfig{}, fmt.Errorf("ParseRelationConfig: %w", err)
    }
    return rc, nil
}
```

### Step 3: Modify fields schema — add columns

**Files to modify (6 total):**

| File | Column type for `validation` | Column type for `ui_config` |
|------|-------|---------|
| `sql/schema/8_fields/schema.sql` | `TEXT NOT NULL` | `TEXT NOT NULL` |
| `sql/schema/8_fields/schema_mysql.sql` | `TEXT NOT NULL` | `TEXT NOT NULL` |
| `sql/schema/8_fields/schema_psql.sql` | `TEXT NOT NULL` | `TEXT NOT NULL` |
| `sql/schema/8_fields/queries.sql` | Add to INSERT, UPDATE | Add to INSERT, UPDATE |
| `sql/schema/8_fields/queries_mysql.sql` | Add to INSERT, UPDATE | Add to INSERT, UPDATE |
| `sql/schema/8_fields/queries_psql.sql` | Add to INSERT, UPDATE | Add to INSERT, UPDATE |

No DEFAULT — application always provides values via `types.EmptyJSON`.

**Query changes required:**
- `CreateFieldTable` — add columns to DDL
- `CreateField` — add `validation, ui_config` to INSERT column list and VALUES
- `UpdateField` — add `validation = ?, ui_config = ?` to SET clause
- `GetField`, `ListField`, `ListFieldByDatatypeID` — use `SELECT *` so no query text change needed (sqlc picks up the new columns from the schema)

**Parameter count changes:**
- SQLite: `?` placeholders increase from 8 to 10 for INSERT, 8 to 10 for UPDATE
- MySQL: same (`?` placeholders)
- PostgreSQL: `$N` placeholders increase from `$8` to `$10` for INSERT, `$8` to `$10` for UPDATE

### Step 4: Modify admin_fields schema — add columns

**Files to modify (6 total):** Same pattern as Step 3 but in `sql/schema/10_admin_fields/`

- `schema.sql`, `schema_mysql.sql`, `schema_psql.sql` — add `validation TEXT NOT NULL` and `ui_config TEXT NOT NULL` columns
- `queries.sql`, `queries_mysql.sql`, `queries_psql.sql` — add to INSERT and UPDATE column lists and parameter placeholders

### Step 5: Create content_relations schema

**Directory (new):** `sql/schema/24_content_relations/`

Create 7 files:
- `.sqlsrc.json` (copy from existing schema dir)
- `schema.sql` (SQLite)
- `schema_mysql.sql`
- `schema_psql.sql`
- `queries.sql`
- `queries_mysql.sql`
- `queries_psql.sql`

**SQLite schema:**
```sql
CREATE TABLE IF NOT EXISTS content_relations (
    content_relation_id TEXT PRIMARY KEY NOT NULL CHECK (length(content_relation_id) = 26),
    source_content_id TEXT NOT NULL
        REFERENCES content_data(content_data_id)
            ON DELETE CASCADE,
    target_content_id TEXT NOT NULL
        REFERENCES content_data(content_data_id)
            ON DELETE CASCADE,
    field_id TEXT NOT NULL
        REFERENCES fields(field_id)
            ON DELETE RESTRICT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    date_created TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (source_content_id != target_content_id)
);

-- Unique constraint ordered to also serve ListBySourceAndField (prefix: source, field)
-- and ListBySource (prefix: source) queries
CREATE UNIQUE INDEX IF NOT EXISTS idx_content_relations_unique
    ON content_relations(source_content_id, field_id, target_content_id);
-- Composite index for ListByTarget ORDER BY date_created
CREATE INDEX IF NOT EXISTS idx_content_relations_target
    ON content_relations(target_content_id, date_created);
-- Supports ON DELETE RESTRICT FK checks when deleting a field
CREATE INDEX IF NOT EXISTS idx_content_relations_field
    ON content_relations(field_id);
```

**Index rationale:**
- `(source_content_id, field_id, target_content_id)` — unique constraint that also serves `ListContentRelationsBySourceAndField` (prefix `source_content_id, field_id`) and `ListContentRelationsBySource` (prefix `source_content_id`). No separate source index needed.
- `(target_content_id, date_created)` — composite index for `ListContentRelationsByTarget` with sort coverage.
- `(field_id)` — supports `ON DELETE RESTRICT` FK checks when deleting a field.

**FK rationale:** `field_id` uses `ON DELETE RESTRICT` (not CASCADE) because deleting a field definition is a schema-level admin operation. RESTRICT forces explicit cleanup of relations before removing the field, preventing silent deletion of potentially thousands of relation rows. Content FK columns use CASCADE since deleting content should clean up its relations.

**MySQL schema:** Use `VARCHAR(26)`, `TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP`, named constraints, `INT` for sort_order, `CHECK (source_content_id != target_content_id)`.

**PostgreSQL schema:** Use `TEXT`, `TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP`, named constraints, `CHECK (source_content_id != target_content_id)`.

**Queries (all 3 dialects):**
- `CreateContentRelationTable :exec`
- `DropContentRelationTable :exec`
- `CountContentRelation :one`
- `GetContentRelation :one` — by content_relation_id
- `CreateContentRelation :one` — RETURNING * (SQLite/Postgres), `:exec` + GetXxx for MySQL
- `DeleteContentRelation :exec` — by content_relation_id
- `UpdateContentRelationSortOrder :exec` — `UPDATE content_relations SET sort_order = ? WHERE content_relation_id = ?`
- `ListContentRelationsBySource :many` — `WHERE source_content_id = ? ORDER BY sort_order`
- `ListContentRelationsByTarget :many` — `WHERE target_content_id = ? ORDER BY date_created`
- `ListContentRelationsBySourceAndField :many` — `WHERE source_content_id = ? AND field_id = ? ORDER BY sort_order`

### Step 6: Create admin_content_relations schema

**Directory (new):** `sql/schema/25_admin_content_relations/`

Mirror of Step 5 but for the admin system. Same index strategy, `NOT NULL` on `date_created`, `ON DELETE RESTRICT` on field FK, same CHECK constraint, and `UpdateAdminContentRelationSortOrder` query.

Key differences from content_relations:

- Table: `admin_content_relations`, PK: `admin_content_relation_id`
- `source_content_id` REFERENCES `admin_content_data(admin_content_data_id)` — column is named `source_content_id` for code symmetry with the public table, but holds `AdminContentID` values. The sqlc overrides in Step 8 map this to `AdminContentID`. Add a SQL comment in the DDL: `-- holds admin_content_data_id, named for code symmetry with content_relations`
- `target_content_id` REFERENCES `admin_content_data(admin_content_data_id)` — same naming note
- `admin_field_id` (not `field_id`) REFERENCES `admin_fields(admin_field_id)` with `ON DELETE RESTRICT` — using `admin_field_id` ensures the `*.admin_field_id` wildcard override maps to `AdminFieldID` automatically
- `CHECK (source_content_id != target_content_id)` — same self-reference prevention
- Indexes prefixed with `idx_admin_content_relations_*`, same composite patterns as Step 5

### Step 7: Update wipe queries

**Files to modify (3 total):**
- `sql/schema/0_wipe/queries.sql`
- `sql/schema/0_wipe/queries_mysql.sql`
- `sql/schema/0_wipe/queries_psql.sql`

Add `DROP TABLE IF EXISTS` for the new tables. They must be dropped **before** their FK targets (`content_data`, `admin_content_data`, `fields`, `admin_fields`). Insert these lines at the top of the DropAllTables query (before `admin_datatypes_fields`):

```sql
DROP TABLE IF EXISTS admin_content_relations;
DROP TABLE IF EXISTS content_relations;
```

All three dialect wipe files are currently identical — keep them in sync.

### Step 8: Update sqlc.yml overrides

**Add to ALL THREE engine sections** (sqlite, mysql, postgresql):

**Renames (in each `rename:` block):**
```yaml
content_relation_id: "ContentRelationID"
admin_content_relation_id: "AdminContentRelationID"
source_content_id: "SourceContentID"
target_content_id: "TargetContentID"
content_relation: ContentRelations
admin_content_relation: AdminContentRelations
```

**Overrides (in each `overrides:` list):**
```yaml
# CONTENT RELATIONS — PRIMARY IDs
- column: "content_relations.content_relation_id"
  go_type: {import: "github.com/hegner123/modulacms/internal/db/types", type: "ContentRelationID"}
- column: "admin_content_relations.admin_content_relation_id"
  go_type: {import: "github.com/hegner123/modulacms/internal/db/types", type: "AdminContentRelationID"}
# CONTENT RELATIONS — FK columns (no wildcard matches source_content_id/target_content_id)
- column: "content_relations.source_content_id"
  go_type: {import: "github.com/hegner123/modulacms/internal/db/types", type: "ContentID"}
- column: "content_relations.target_content_id"
  go_type: {import: "github.com/hegner123/modulacms/internal/db/types", type: "ContentID"}
- column: "admin_content_relations.source_content_id"
  go_type: {import: "github.com/hegner123/modulacms/internal/db/types", type: "AdminContentID"}
- column: "admin_content_relations.target_content_id"
  go_type: {import: "github.com/hegner123/modulacms/internal/db/types", type: "AdminContentID"}
# field_id handled by existing *.field_id wildcard -> FieldID
# admin_field_id handled by existing *.admin_field_id wildcard -> AdminFieldID
# date_created handled by existing *.date_created wildcard -> Timestamp
```

**Fix pre-existing `*.parent_id` wildcard bug** (in each `overrides:` list):

The existing `*.parent_id` wildcard maps all `parent_id` columns to `NullableContentID`. This is wrong for `fields.parent_id` (references `datatypes.datatype_id`) and `admin_fields.parent_id` (references `admin_datatypes.admin_datatype_id`). Add explicit overrides that take precedence over the wildcard:

```yaml
# FIX: fields.parent_id references datatypes, not content_data
- column: "fields.parent_id"
  go_type: {import: "github.com/hegner123/modulacms/internal/db/types", type: "NullableDatatypeID"}
# FIX: admin_fields.parent_id references admin_datatypes, not content_data
- column: "admin_fields.parent_id"
  go_type: {import: "github.com/hegner123/modulacms/internal/db/types", type: "NullableAdminDatatypeID"}
```

### Step 9: Run sqlc

**Gate:** Steps 3, 4, 5, 6, 7, and 8 must ALL be complete before running this step. Running sqlc with partial schema/query/override changes produces inconsistent generated code.

```bash
cd sql && just sqlc
```

This regenerates `internal/db-sqlite/`, `internal/db-mysql/`, `internal/db-psql/`.

### Step 10: Update field.go — add Validation and UIConfig

**File:** `internal/db/field.go`

**Changes:**
1. Add `Validation string` and `UIConfig string` to `Fields` struct (after `Data`)
2. Add same fields to `CreateFieldParams` and `UpdateFieldParams`
3. Update `ParentID` type from `types.NullableContentID` to `types.NullableDatatypeID` in `Fields`, `CreateFieldParams`, `UpdateFieldParams` (matches the sqlc override fix from Step 8)
4. Add `Validation string` and `UIConfig string` to `FieldsJSON` struct (with json tags `"validation"`, `"ui_config"`)
5. Update `MapFieldJSON()` to include new fields
6. Update `MapStringField()` to include new fields
7. Update all mapper functions (6 total: Map, MapCreate, MapUpdate x 3 drivers) to pass through the new fields
8. Update audited command Execute methods that construct sqlc params (6 total: New x 3 + Update x 3). Delete commands only take an ID — they do not need changes.
9. Update `ListFieldsByDatatypeID` parameter type from `types.NullableContentID` to `types.NullableDatatypeID` (matches the sqlc override fix)

**Pattern reference:** See how `Data string` is currently mapped — `Validation` and `UIConfig` follow the exact same pattern.

### Step 11: Update admin_field.go — add Validation and UIConfig

**File:** `internal/db/admin_field.go`

Same changes as Step 10 but for admin types:
1. Add `Validation string` and `UIConfig string` to `AdminFields`, `CreateAdminFieldParams`, `UpdateAdminFieldParams`
2. Update `ParentID` type from `types.NullableContentID` to `types.NullableAdminDatatypeID` in all structs (matches the sqlc override fix from Step 8)
3. Update `MapAdminFieldJSON()` — add Validation and UIConfig to the FieldsJSON output
4. Update `MapStringAdminField()` — add new fields
5. Update all mapper functions (6 total: Map, MapCreate, MapUpdate x 3 drivers)
6. Update audited command Execute methods that construct sqlc params (6 total: New x 3 + Update x 3). Delete commands only take an ID — they do not need changes.
7. Update `ListAdminFieldByRouteIdRow` and `ListAdminFieldsByDatatypeIDRow` — update `ParentID` type and add `Validation`/`UIConfig` if these structs include `Data`

### Step 12: Create content_relation.go

**File (new):** `internal/db/content_relation.go`

Follow the entity file pattern from `internal/db/route.go`:

```go
// STRUCTS
type ContentRelations struct {
    ContentRelationID types.ContentRelationID `json:"content_relation_id"`
    SourceContentID   types.ContentID         `json:"source_content_id"`
    TargetContentID   types.ContentID         `json:"target_content_id"`
    FieldID           types.FieldID           `json:"field_id"`
    SortOrder         int64                   `json:"sort_order"`
    DateCreated       types.Timestamp         `json:"date_created"`
}

type CreateContentRelationParams struct {
    SourceContentID types.ContentID  `json:"source_content_id"`
    TargetContentID types.ContentID  `json:"target_content_id"`
    FieldID         types.FieldID    `json:"field_id"`
    SortOrder       int64            `json:"sort_order"`
    DateCreated     types.Timestamp  `json:"date_created"`
}

type UpdateContentRelationSortOrderParams struct {
    ContentRelationID types.ContentRelationID `json:"content_relation_id"`
    SortOrder         int64                   `json:"sort_order"`
}
```

No general-purpose UpdateParams needed — only sort_order is mutable. A single-row `UPDATE sort_order WHERE content_relation_id = ?` avoids the cost, audit noise, and race conditions of delete+recreate for reordering.

**All struct fields use types from `internal/db/types/`** — no `sql.NullString`, `database/sql`, or other stock SQL types:

| Column | Go type on wrapper struct |
|--------|--------------------------|
| `content_relation_id` | `types.ContentRelationID` |
| `source_content_id` | `types.ContentID` |
| `target_content_id` | `types.ContentID` |
| `field_id` | `types.FieldID` |
| `sort_order` | `int64` |
| `date_created` | `types.Timestamp` |

Implement:
- MapStringContentRelation for TUI tables
- SQLite: Map, MapCreate, MapUpdateSortOrder, Count, CreateTable, Create, Delete, UpdateSortOrder, Get, ListBySource, ListByTarget, ListBySourceAndField
- MySQL: same (with int32->int64 for sort_order)
- PostgreSQL: same (with int32->int64 for sort_order)
- Audited commands: NewContentRelationCmd, DeleteContentRelationCmd, UpdateContentRelationSortOrderCmd x 3 drivers (9 total)

### Step 13: Create admin_content_relation.go

**File (new):** `internal/db/admin_content_relation.go`

Mirror of Step 12 using:
- `AdminContentRelations` struct with `AdminContentRelationID`
- `UpdateAdminContentRelationSortOrderParams` with `AdminContentRelationID` + `SortOrder`
- Same query methods and audited commands (including UpdateSortOrder)

**Column-to-type mapping (all from `internal/db/types/`):**

| Column | Go type on wrapper struct |
|--------|--------------------------|
| `admin_content_relation_id` | `types.AdminContentRelationID` |
| `source_content_id` | `types.AdminContentID` |
| `target_content_id` | `types.AdminContentID` |
| `admin_field_id` | `types.AdminFieldID` |
| `sort_order` | `int64` |
| `date_created` | `types.Timestamp` |

The admin table uses `admin_field_id` (not `field_id`) so the existing `*.admin_field_id` wildcard override maps it to `AdminFieldID` automatically.

### Step 14: Update db.go DbDriver interface

**File:** `internal/db/db.go`

Add two new sections to the interface:

```go
// ContentRelations
CountContentRelations() (*int64, error)
CreateContentRelation(context.Context, audited.AuditContext, CreateContentRelationParams) (*ContentRelations, error)
CreateContentRelationTable() error
DeleteContentRelation(context.Context, audited.AuditContext, types.ContentRelationID) error
GetContentRelation(types.ContentRelationID) (*ContentRelations, error)
ListContentRelationsBySource(types.ContentID) (*[]ContentRelations, error)
ListContentRelationsByTarget(types.ContentID) (*[]ContentRelations, error)
ListContentRelationsBySourceAndField(types.ContentID, types.FieldID) (*[]ContentRelations, error)
UpdateContentRelationSortOrder(context.Context, audited.AuditContext, UpdateContentRelationSortOrderParams) error

// AdminContentRelations
CountAdminContentRelations() (*int64, error)
CreateAdminContentRelation(context.Context, audited.AuditContext, CreateAdminContentRelationParams) (*AdminContentRelations, error)
CreateAdminContentRelationTable() error
DeleteAdminContentRelation(context.Context, audited.AuditContext, types.AdminContentRelationID) error
GetAdminContentRelation(types.AdminContentRelationID) (*AdminContentRelations, error)
ListAdminContentRelationsBySource(types.AdminContentID) (*[]AdminContentRelations, error)
ListAdminContentRelationsByTarget(types.AdminContentID) (*[]AdminContentRelations, error)
ListAdminContentRelationsBySourceAndField(types.AdminContentID, types.AdminFieldID) (*[]AdminContentRelations, error)
UpdateAdminContentRelationSortOrder(context.Context, audited.AuditContext, UpdateAdminContentRelationSortOrderParams) error
```

Also update `ListFieldsByDatatypeID` parameter type from `types.NullableContentID` to `types.NullableDatatypeID` to match the override fix.

### Step 15: Update combined schema files

**Files:**
- `sql/all_schema.sql` — append content_relations and admin_content_relations CREATE TABLE + indexes
- `sql/all_schema_mysql.sql` — same, MySQL dialect
- `sql/all_schema_psql.sql` — same, PostgreSQL dialect

Also update the fields and admin_fields CREATE TABLE statements in these files to include the new `validation` and `ui_config` columns.

### Step 16: Update CreateAllTables and DropAllTables

**File:** `internal/db/db.go`

- Add `CreateContentRelationTable()` after `CreateContentFieldTable()` (Tier 5.5 — depends on content_data and fields)
- Add `CreateAdminContentRelationTable()` after `CreateAdminContentFieldTable()`
- Ensure `DropAllTables` order matches the wipe queries (drop relation tables before their FK targets)

### Step 17: Update all existing call sites

Every place that constructs `CreateFieldParams`, `UpdateFieldParams`, `CreateAdminFieldParams`, or `UpdateAdminFieldParams` must use `types.EmptyJSON` for creates, or pass through existing values for updates.

Additionally, every place that references `types.NullableContentID` for field `ParentID` must be updated to `types.NullableDatatypeID` (for public fields) or `types.NullableAdminDatatypeID` (for admin fields).

**CreateFieldParams call sites:**

| File | Line | Context |
|------|------|---------|
| `internal/router/import.go` | ~320 | Import handler — add `Validation: types.EmptyJSON, UIConfig: types.EmptyJSON` |
| `internal/cli/update_dialog.go` | ~1486 | TUI create field dialog — add `Validation: types.EmptyJSON, UIConfig: types.EmptyJSON` |
| `internal/definitions/install.go` | ~44 | Install definitions — add `Validation: types.EmptyJSON, UIConfig: types.EmptyJSON` |
| `internal/db/db.go` | ~690 | Bootstrap data (SQLite) — add `Validation: types.EmptyJSON, UIConfig: types.EmptyJSON` |
| `internal/db/db.go` | ~1385 | Bootstrap data (MySQL) — add `Validation: types.EmptyJSON, UIConfig: types.EmptyJSON` |
| `internal/db/db.go` | ~2053 | Bootstrap data (PostgreSQL) — add `Validation: types.EmptyJSON, UIConfig: types.EmptyJSON` |

**UpdateFieldParams call sites:**

| File | Line | Context |
|------|------|---------|
| `internal/cli/update_dialog.go` | ~1832 | TUI edit field — add `Validation: existing.Validation, UIConfig: existing.UIConfig` |

**CreateAdminFieldParams call sites:**

| File | Line | Context |
|------|------|---------|
| `internal/cli/admin_update_dialog.go` | ~450 | TUI create admin field — add `Validation: types.EmptyJSON, UIConfig: types.EmptyJSON` |
| `internal/db/db.go` | ~673 | Bootstrap data (SQLite) — add `Validation: types.EmptyJSON, UIConfig: types.EmptyJSON` |
| `internal/db/db.go` | ~1368 | Bootstrap data (MySQL) — add `Validation: types.EmptyJSON, UIConfig: types.EmptyJSON` |
| `internal/db/db.go` | ~2036 | Bootstrap data (PostgreSQL) — add `Validation: types.EmptyJSON, UIConfig: types.EmptyJSON` |

**UpdateAdminFieldParams call sites:**

| File | Line | Context |
|------|------|---------|
| `internal/cli/admin_update_dialog.go` | ~542 | TUI edit admin field — add `Validation: existing.Validation, UIConfig: existing.UIConfig` |

**HTTP router handlers (parse request bodies into params):**

| File | Context |
|------|---------|
| `internal/router/fields.go` | Field CRUD endpoints — update request parsing to include Validation, UIConfig |
| `internal/router/adminFields.go` | Admin field CRUD endpoints — update request parsing |

**FieldsJSON struct** (used by model/tree building via `BuildTree` and `BuildAdminTree`):

| File | Context |
|------|---------|
| `internal/db/field.go` | `FieldsJSON` struct — add `Validation string` and `UIConfig string` fields |
| `internal/db/field.go` | `MapFieldJSON()` — pass through new fields |
| `internal/db/admin_field.go` | `MapAdminFieldJSON()` — pass through new fields into FieldsJSON |

---

## Files Modified (Summary)

| File | Change |
|------|--------|
| `internal/db/types/types_ids.go` | Add ContentRelationID, AdminContentRelationID |
| `internal/db/types/types_field_config.go` **(new)** | EmptyJSON, Cardinality, RelationConfig, ValidationConfig, UIConfig, Parse functions |
| `sql/schema/8_fields/*.sql` (6 files) | Add validation, ui_config columns |
| `sql/schema/10_admin_fields/*.sql` (6 files) | Add validation, ui_config columns |
| `sql/schema/24_content_relations/` **(new dir, 7 files)** | New table + queries + CHECK constraint |
| `sql/schema/25_admin_content_relations/` **(new dir, 7 files)** | New table + queries + CHECK constraint |
| `sql/schema/0_wipe/queries*.sql` (3 files) | Add DROP TABLE for new tables |
| `sql/sqlc.yml` | Add renames + overrides (x3 engines) + parent_id fix |
| `internal/db/field.go` | Add Validation, UIConfig to structs + mappers + audit cmds + FieldsJSON; fix ParentID type |
| `internal/db/admin_field.go` | Add Validation, UIConfig to structs + mappers + audit cmds; fix ParentID type |
| `internal/db/content_relation.go` **(new)** | Full entity file (including UpdateSortOrder) |
| `internal/db/admin_content_relation.go` **(new)** | Full entity file (including UpdateSortOrder) |
| `internal/db/db.go` | Add interface methods + CreateAllTables + bootstrap data; fix ListFieldsByDatatypeID param type |
| `internal/cli/update_dialog.go` | Add Validation, UIConfig to field create/update params |
| `internal/cli/admin_update_dialog.go` | Add Validation, UIConfig to admin field create/update params |
| `internal/router/import.go` | Add Validation, UIConfig to import field create |
| `internal/router/fields.go` | Update request parsing for new fields |
| `internal/router/adminFields.go` | Update request parsing for new fields |
| `internal/definitions/install.go` | Add Validation, UIConfig to install field create |
| `sql/all_schema.sql` | Add new tables + update fields tables |
| `sql/all_schema_mysql.sql` | Add new tables + update fields tables |
| `sql/all_schema_psql.sql` | Add new tables + update fields tables |

---

## Verification

1. **`just sqlc`** — must complete without errors after schema + query + override changes
2. **`go build ./...`** — must compile after all Go changes (db wrapper, types, interface)
3. **`just test`** — run full test suite
4. **Interface check** — all three drivers (Database, MysqlDatabase, PsqlDatabase) must implement all new DbDriver methods or build will fail
5. **EmptyJSON check** — search for `"{}"` in `internal/` — should only appear in `types.EmptyJSON` definition, not scattered at call sites
6. **ParentID type check** — `field.go` and `admin_field.go` ParentID fields should no longer reference `NullableContentID`

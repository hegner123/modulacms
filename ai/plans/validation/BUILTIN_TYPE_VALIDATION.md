# Built-in Type Validation Plan

## Context

The validation engine (`internal/validation/`) already runs type-specific validators via `validateType()` in `type_validators.go`. This works for the 16 built-in field types using a `switch` on hardcoded `types.FieldType` constants. Custom field types hit the `default` case and skip type validation entirely.

This means a user who registers a custom field type like `"color_picker"` gets no automatic shape validation. They can store `"banana"` in a color picker field and the system accepts it silently.

### What Already Exists

`validateType()` handles:

| Field Type | Validation | Implementation |
|------------|-----------|----------------|
| `number` | Parses as float | `strconv.ParseFloat` |
| `boolean` | `true`/`false`/`1`/`0` | Exact match |
| `email` | Valid email format | `types.Email.Validate()` |
| `url` | Valid URL | `types.URL.Validate()` |
| `date` | `YYYY-MM-DD` | `time.Parse` |
| `datetime` | RFC3339 or common variants | Multiple `time.Parse` attempts |
| `json` | Valid JSON | `json.Valid()` |
| `slug` | Slug format | `types.Slug.Validate()` |
| `select` | Value in options list | Set membership against `data.options` |
| `_id` | Valid ULID | `types.ContentID.Validate()` |
| `media` | Valid ULID | `types.MediaID.Validate()` |
| `text`, `textarea`, `richtext` | None (any string) | Returns `""` |

The validation pipeline in `ValidateField()` runs in order:
1. Parse validation config
2. Check for required rule
3. Empty gate (skip if not required and empty)
4. Required check (fail if required and empty)
5. **Type-specific validator** (`validateType()`)
6. Composable rules (`EvaluateRules()`)

### The Gap

The `switch` in `validateType()` is closed. Custom field types cannot register type validators without modifying the binary. The `field_types` table stores only `type` (name) and `label` — no validation behavior.

## Goal

Allow field types to carry a built-in type validator that runs automatically on every content write, without requiring users to attach composable validation rules to every field. The system should work for both built-in and custom field types.

## Design

### Option A: Validator Registry (recommended)

Replace the `switch` in `validateType()` with a registry that maps field type names to validator functions. Built-in validators register at init. Custom validators register via:
1. The plugin system (Lua functions that validate values)
2. A new `validator` column on the `field_types` table that names a built-in validation strategy

```go
// TypeValidator is a function that validates a field value for a specific type.
// Returns an error message string, or empty string if valid.
// data is the raw JSON from the fields.data column (for select options, etc.)
type TypeValidator func(value string, data string) string

// Registry maps field type names to their validators.
var typeValidators = map[string]TypeValidator{}

func RegisterTypeValidator(fieldType string, fn TypeValidator) {
    typeValidators[fieldType] = fn
}

func init() {
    RegisterTypeValidator("number", validateNumber)
    RegisterTypeValidator("boolean", validateBoolean)
    RegisterTypeValidator("email", validateEmail)
    RegisterTypeValidator("url", validateURL)
    RegisterTypeValidator("date", validateDate)
    RegisterTypeValidator("datetime", validateDatetime)
    RegisterTypeValidator("json", validateJSON)
    RegisterTypeValidator("slug", validateSlug)
    RegisterTypeValidator("select", validateSelect)
    RegisterTypeValidator("_id", validateIDRef)
    RegisterTypeValidator("media", validateMediaRef)
    // text, textarea, richtext: no registration = no type validation
}
```

Then `validateType()` becomes:

```go
func validateType(ft types.FieldType, value string, data string) string {
    fn, ok := typeValidators[string(ft)]
    if !ok {
        return "" // no type validator registered
    }
    return fn(value, data)
}
```

### Option B: Validator Column on field_types

Add a `validator` column to the `field_types` table that names a validation strategy:

```sql
ALTER TABLE field_types ADD COLUMN validator TEXT NOT NULL DEFAULT '';
```

Values would be predefined strategy names that the Go registry understands:

| validator | Behavior |
|-----------|----------|
| `""` (empty) | No type validation |
| `"number"` | `strconv.ParseFloat` |
| `"boolean"` | `true`/`false`/`1`/`0` |
| `"email"` | Email format check |
| `"url"` | URL parse check |
| `"date"` | ISO date parse |
| `"datetime"` | ISO datetime parse |
| `"json"` | `json.Valid` |
| `"slug"` | Slug format check |
| `"select"` | Options membership |
| `"ulid"` | ULID format check (covers `_id` and `media`) |

A user registering `color_picker` could set `validator: ""` (no built-in check) or later, when plugin validators exist, reference a plugin-provided validator.

### Recommendation

**Start with Option A** (registry in Go). It's a refactor of the existing `switch`, not a schema change. The `switch` cases become registered functions, the `default` case becomes a map miss. Zero behavior change for built-in types.

**Then add Option B** when the schema-from-filesystem feature lands. The `validator` column on `field_types` would drive which registry entry to use for custom types. The schema file could declare it:

```typescript
// In field-types.ts, generated with validator info
f.color_picker("color", "Background Color") // no built-in validator
f.hex_color("color", "Background Color")    // uses "hex" validator
```

### Plugin Validators (future)

Lua plugins could register type validators via the plugin API:

```lua
modula.register_type_validator("color_picker", function(value, data)
    -- validate hex color format
    if not value:match("^#%x%x%x%x%x%x$") then
        return "must be a hex color (e.g. #FF0000)"
    end
    return ""
end)
```

This would register into the same Go-side `typeValidators` map via a bridge function. The plugin system already has hook infrastructure that could support this.

## Implementation

### Phase 1: Registry Refactor

**Files changed:**
- `internal/validation/type_validators.go` — replace `switch` with registry, extract each case into a named function

**Scope:** Pure refactor. No new behavior, no schema changes. The existing tests should pass unchanged.

### Phase 2: Validator Column

**Files changed:**
- `sql/schema/27_field_types/` — add `validator TEXT NOT NULL DEFAULT ''` column
- `sql/schema/28_admin_field_types/` — same
- `internal/db/` — update wrappers after `just sqlc`
- `internal/validation/type_validators.go` — `LoadValidatorsFromDB()` that reads `field_types` and registers validators by strategy name
- `cmd/serve.go` — call `LoadValidatorsFromDB()` at startup

**Bootstrap data:** Built-in field types get their `validator` column populated during `CreateBootstrapData`. E.g., `number` gets `validator: "number"`, `text` gets `validator: ""`.

### Phase 3: Schema-from-Files Integration

When `modula schema generate-types` runs, it reads the `validator` column alongside `type` and `label`. The generated `field-types.ts` can include the validator name as a comment or metadata for documentation:

```typescript
// validator: "number"
number: (name: string, label: string, opts?: FieldOpts): Field => ...
// validator: "" (no built-in validation)
color_picker: (name: string, label: string, opts?: FieldOpts): Field => ...
```

### Phase 4: Plugin Validators

Extend the plugin API to allow `modula.register_type_validator(type_name, fn)`. The bridge calls `RegisterTypeValidator()` in Go. Requires the plugin system's hook infrastructure.

## Relationship to Schema-from-Files

This feature changes the schema-from-files sync validation. Currently the sync validates that a field's `type` value exists in `field_types`. With the `validator` column, the sync could also warn when a custom field type has no validator — not an error, but a diagnostic that the field will accept any string.

The `modula schema generate-types` command should surface this information so users know which of their field types have built-in validation and which don't.

## Key Files

| File | Purpose |
|------|---------|
| `internal/validation/type_validators.go` | Current `switch`-based type validation (refactor target) |
| `internal/validation/validate.go` | `ValidateField()` pipeline that calls `validateType()` |
| `internal/validation/rules.go` | Composable rule evaluation (`EvaluateRules()`) |
| `internal/db/types/types_field_config.go` | `ValidationConfig`, `RuleOp`, validation rule types |
| `sql/schema/27_field_types/schema.sql` | `field_types` table (currently `type` + `label` only) |
| `internal/plugin/` | Plugin system for future Lua validators |

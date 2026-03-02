# Validation Package

**Package:** `internal/validation`
**Purpose:** Field-level content validation with composable rules and type-specific checks

---

## Package Structure

```
internal/validation/
    validate.go         # Public API: ValidateField, ValidateBatch
    errors.go           # FieldError, ValidationErrors
    rules.go            # Composable rule evaluation engine
    type_validators.go  # Per-FieldType validators
    validation_test.go  # 1120+ lines of tests
```

---

## Validation Flow

```
ValidateField(FieldInput)
  1. Parse ValidationConfig from JSON
  2. Scan rules for "required" op
  3. Empty gate: not required + empty value -> pass (skip all checks)
  4. Required check: required + empty value -> fail
  5. Type validation: validateType(fieldType, value, data)
  6. Composable rules: EvaluateRules(value, rules) (skips required)
  7. Return *FieldError or nil
```

Type validation runs before composable rules. If type validation fails, the error is collected alongside any rule failures.

---

## Public API

### Types

```go
// FieldInput bundles everything needed to validate a single field.
type FieldInput struct {
    FieldID    types.FieldID
    Label      string
    FieldType  types.FieldType
    Value      string   // submitted value
    Validation string   // raw JSON from fields.validation column
    Data       string   // raw JSON from fields.data column (used by select type)
}
```

```go
// FieldError represents validation failures for one field.
type FieldError struct {
    FieldID  types.FieldID `json:"field_id"`
    Label    string        `json:"label"`
    Messages []string      `json:"messages"`
}
```

```go
// ValidationErrors accumulates errors across multiple fields.
type ValidationErrors struct {
    Fields []FieldError `json:"fields"`
}
```

### Functions

| Function | Signature | Returns |
|----------|-----------|---------|
| `ValidateField` | `(input FieldInput) *FieldError` | `nil` if valid, `*FieldError` with messages if invalid |
| `ValidateBatch` | `(inputs []FieldInput) ValidationErrors` | Collects all field errors; check `.HasErrors()` |
| `EvaluateRules` | `(value string, entries []types.RuleEntry) []string` | Slice of error message strings (empty = pass) |

### ValidationErrors Methods

| Method | Purpose |
|--------|---------|
| `Error() string` | Joins all field errors with `"; "` |
| `HasErrors() bool` | True if at least one FieldError exists |
| `ForField(id types.FieldID) *FieldError` | Lookup by field ID, returns nil if not found |
| `ClearField(id types.FieldID)` | Remove errors for a specific field |

---

## Type Validators

Each `types.FieldType` has a dedicated check in `validateType()`. Returns an error message string or empty string.

| FieldType | Validation |
|-----------|------------|
| `text`, `textarea`, `richtext` | No type validation (always pass) |
| `number` | Parseable as `float64` |
| `date` | Format `YYYY-MM-DD` via `time.Parse("2006-01-02", ...)` |
| `datetime` | RFC3339, `2006-01-02T15:04:05`, or `2006-01-02 15:04:05` |
| `boolean` | Exactly `"true"`, `"false"`, `"1"`, or `"0"` |
| `select` | Value exists in options parsed from `data` JSON (`{"options": [...]}`) |
| `email` | Delegates to `types.Email` validation |
| `url` | Delegates to `types.URL` validation |
| `slug` | Delegates to `types.Slug` validation (must start with `/`) |
| `media` | Valid ULID |
| `content_tree_ref` | Valid ULID |
| `relation` | Valid ULID |
| `json` | Valid JSON via `json.Valid()` |
| Unknown types | Skipped (forward compatible) |

---

## Composable Rule System

Rules are defined in `types.ValidationConfig` and stored as JSON in the `fields.validation` column.

### Rule Configuration Types (from `internal/db/types/`)

```go
type ValidationConfig struct {
    Rules []RuleEntry `json:"rules,omitempty"`
}

// Exactly one of Rule or Group must be set.
type RuleEntry struct {
    Rule  *ValidationRule `json:"rule,omitempty"`
    Group *RuleGroup      `json:"group,omitempty"`
}

// Logical operators for combining rules.
type RuleGroup struct {
    AllOf []RuleEntry `json:"all_of,omitempty"`  // AND: all must pass
    AnyOf []RuleEntry `json:"any_of,omitempty"`  // OR:  at least one must pass
}

type ValidationRule struct {
    Op      RuleOp    `json:"op"`
    Value   string    `json:"value,omitempty"`    // substring for contains/starts_with/ends_with/equals
    Values  []string  `json:"values,omitempty"`   // set for one_of
    Class   CharClass `json:"class,omitempty"`    // character class (alternative to Value)
    Cmp     Cmp       `json:"cmp,omitempty"`      // comparison operator
    N       *float64  `json:"n,omitempty"`        // numeric operand
    Negate  bool      `json:"negate,omitempty"`   // invert result
    Message string    `json:"message,omitempty"`  // custom error (overrides default)
}
```

### Rule Operations

| RuleOp | Description | Uses |
|--------|-------------|------|
| `required` | Value must be non-empty | Handled before rule evaluation |
| `contains` | Substring or character class present | `Value` or `Class` |
| `starts_with` | Begins with substring or character class | `Value` or `Class` |
| `ends_with` | Ends with substring or character class | `Value` or `Class` |
| `equals` | Exact string match | `Value` |
| `length` | Rune count comparison | `Cmp` + `N` |
| `count` | Occurrences of substring or class members | `Value` or `Class` + `Cmp` + `N` |
| `range` | Numeric value comparison (parsed as float64) | `Cmp` + `N` |
| `item_count` | Count items in comma-separated or JSON array | `Cmp` + `N` |
| `one_of` | Value is in a fixed set | `Values` |

### Comparison Operators (`Cmp`)

`eq` (==), `neq` (!=), `gt` (>), `gte` (>=), `lt` (<), `lte` (<=)

### Character Classes (`CharClass`)

| Class | Matches |
|-------|---------|
| `uppercase` | A-Z |
| `lowercase` | a-z |
| `digits` | 0-9 |
| `symbols` | Non-alphanumeric, non-whitespace |
| `spaces` | Space, tab, newline, CR |

### Negate

Any rule can be inverted with `"negate": true`. A `contains` with `negate` becomes "must not contain."

### Custom Messages

If `Message` is set on a rule, it replaces the default generated message. Default messages are human-readable descriptions generated from rule parameters (e.g., "must contain uppercase" or "length must be >= 8").

---

## Item Count Semantics

The `item_count` rule counts items in a value:

- Empty string: 0 items
- Starts with `[`: Try JSON array parse, fall back to comma-separated
- Otherwise: Comma-separated with whitespace trimming, empty segments skipped

---

## Config Limits

| Limit | Value |
|-------|-------|
| Maximum rule nesting depth | 10 levels |
| Maximum integer N | 1,000,000 |
| Range N bounds | -1e15 to 1e15 |

These are enforced during `ValidationConfig.Validate()` in the types package.

---

## JSON Examples

### Required email with length constraint

```json
{
  "rules": [
    { "rule": { "op": "required" } },
    { "rule": { "op": "length", "cmp": "lte", "n": 254 } }
  ]
}
```

### Password rules using AllOf group

```json
{
  "rules": [
    { "rule": { "op": "required" } },
    { "rule": { "op": "length", "cmp": "gte", "n": 8 } },
    {
      "group": {
        "all_of": [
          { "rule": { "op": "contains", "class": "uppercase", "message": "must include an uppercase letter" } },
          { "rule": { "op": "contains", "class": "lowercase", "message": "must include a lowercase letter" } },
          { "rule": { "op": "contains", "class": "digits", "message": "must include a number" } }
        ]
      }
    }
  ]
}
```

### At least one of two formats (AnyOf)

```json
{
  "rules": [
    {
      "group": {
        "any_of": [
          { "rule": { "op": "starts_with", "value": "https://" } },
          { "rule": { "op": "starts_with", "value": "http://" } }
        ]
      }
    }
  ]
}
```

### Must not contain (negation)

```json
{
  "rules": [
    { "rule": { "op": "contains", "value": "<script", "negate": true, "message": "must not contain script tags" } }
  ]
}
```

### Numeric range

```json
{
  "rules": [
    { "rule": { "op": "range", "cmp": "gte", "n": 0 } },
    { "rule": { "op": "range", "cmp": "lte", "n": 100 } }
  ]
}
```

### Select from fixed set

```json
{
  "rules": [
    { "rule": { "op": "one_of", "values": ["small", "medium", "large"] } }
  ]
}
```

---

## Usage in Codebase

### HTTP API (single field)

`internal/router/contentFields.go` — `apiCreateContentField()`, `apiUpdateContentField()`

```go
input := validation.FieldInput{
    FieldID:    field.FieldID,
    Label:      field.Label,
    FieldType:  field.FieldType,
    Value:      submittedValue,
    Validation: field.Validation,
    Data:       field.Data,
}

if fieldErr := validation.ValidateField(input); fieldErr != nil {
    // Write 422 with fieldErr
}
```

### HTTP API (batch)

`internal/router/contentBatch.go` — `ContentBatchHandler()`

```go
var inputs []validation.FieldInput
for _, item := range batch.Items {
    inputs = append(inputs, validation.FieldInput{
        FieldID:    item.FieldID,
        Label:      fieldDef.Label,
        FieldType:  fieldDef.FieldType,
        Value:      item.Value,
        Validation: fieldDef.Validation,
        Data:       fieldDef.Data,
    })
}

errs := validation.ValidateBatch(inputs)
if errs.HasErrors() {
    // Write 422 with errs
}
```

### TUI (form pre-submit)

`internal/cli/form_dialog.go` — validates all fields before submitting the content form dialog.

### Admin Panel

`internal/admin/handlers/content.go` — server-side validation for admin UI content editing.

### HTTP Error Response Format

Validation failures return HTTP 422 with JSON:

```json
{
  "fields": [
    {
      "field_id": "01ARYZ6S410000000000000000",
      "label": "Email",
      "messages": ["must be a valid email address"]
    },
    {
      "field_id": "01ARYZ6S410000000000000001",
      "label": "Title",
      "messages": ["required", "length must be <= 200"]
    }
  ]
}
```

---

## Design Decisions

1. **Fail-closed**: Missing or unparseable validation config is treated as "no rules" (empty gate applies). Missing field data is treated as empty string.

2. **Forward compatible**: Unknown field types skip type validation. Unknown rule ops are silently skipped. This allows new types and rules to be added without breaking existing validation configs.

3. **Required is special**: The `required` rule is extracted and checked before type validation and composable rules. This enables the empty gate — optional fields with no value skip all validation.

4. **Type validation before rules**: Type-specific checks (is this a valid email? valid JSON?) run before composable rules. Both produce separate error messages that are collected together.

5. **No regex**: All string matching uses exact substrings or character classes. No user-supplied patterns are compiled as regex.

6. **Human-readable defaults**: Every rule operation generates a readable default error message from its parameters. Custom messages override these per-rule.

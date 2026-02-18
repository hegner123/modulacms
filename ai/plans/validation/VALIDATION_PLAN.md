# Validation Engine Plan

## Context

Content field values currently flow from user input (TUI or API) directly into the database with zero validation. The existing ValidationConfig struct (Required, MinLength, MaxLength, Min, Max, Pattern, MaxItems) and typed validators (types.Email, types.URL, types.Slug) exist but have no callers on the write path. This plan replaces ValidationConfig with a composable rule system and adds a validation engine that enforces field rules at both the HTTP API and TUI layers.

No regex: all validation uses exact string matching, character classification, and numeric comparisons.

**Migration note:** Existing rows in the `fields.validation` column may contain JSON in the old flat format (`{"required":true,"min_length":5}`). After the schema change, `ParseValidationConfig` will parse these into `ValidationConfig{Rules: nil}` because the old field names no longer exist on the struct. This means existing validation rules silently disappear. This is intentional — the old validation was never enforced (no callers on the write path), so no runtime behavior changes. No data migration step is needed.

---

## Phase 1: Composable Validation Schema

Replace `ValidationConfig` in `internal/db/types/types_field_config.go` with a composable rule system. The existing flat fields (`Required`, `MinLength`, `MaxLength`, `Min`, `Max`, `Pattern`, `MaxItems`) are all removed. Every constraint — including required, length bounds, and numeric range — is expressed as a composable rule.

### Rule Types

```go
// RuleOp identifies what a rule checks
type RuleOp string

const (
    RuleRequired   RuleOp = "required"     // value must be non-empty
    RuleContains   RuleOp = "contains"     // value contains a substring or character class
    RuleStartsWith RuleOp = "starts_with"  // value starts with a substring or character class
    RuleEndsWith   RuleOp = "ends_with"    // value ends with a substring or character class
    RuleEquals     RuleOp = "equals"       // value equals exactly
    RuleLength     RuleOp = "length"       // rune count of value
    RuleCount      RuleOp = "count"        // count occurrences of substring or class members
    RuleRange      RuleOp = "range"        // numeric value comparison (parse as float)
    RuleItemCount  RuleOp = "item_count"   // count items in comma-separated list or JSON array
    RuleOneOf      RuleOp = "one_of"       // value is one of a fixed set
)
```

### Comparison Operators

```go
// Cmp is used for numeric comparisons in length/count rules
type Cmp string

const (
    CmpEq  Cmp = "eq"   // ==
    CmpNeq Cmp = "neq"  // !=
    CmpGt  Cmp = "gt"   // >
    CmpGte Cmp = "gte"  // >=
    CmpLt  Cmp = "lt"   // <
    CmpLte Cmp = "lte"  // <=
)
```

### Character Classes

Character classes are referenced by name. These are identified by the `class` field on a rule rather than a literal `value`. A rule can target either a literal string OR a character class, never both.

```go
// CharClass identifies a predefined set of characters
type CharClass string

const (
    ClassUppercase CharClass = "uppercase"  // A-Z
    ClassLowercase CharClass = "lowercase"  // a-z
    ClassDigits    CharClass = "digits"     // 0-9
    ClassSymbols   CharClass = "symbols"    // any char NOT in a-z, A-Z, 0-9, or whitespace
    ClassSpaces    CharClass = "spaces"     // whitespace (space, tab, newline)
)
```

Character classification uses a lookup function, not regex:

```go
func classifyChar(c rune, class CharClass) bool {
    switch class {
    case ClassUppercase:
        return c >= 'A' && c <= 'Z'
    case ClassLowercase:
        return c >= 'a' && c <= 'z'
    case ClassDigits:
        return c >= '0' && c <= '9'
    case ClassSymbols:
        return !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
                 (c >= '0' && c <= '9') || c == ' ' || c == '\t' || c == '\n' || c == '\r')
    case ClassSpaces:
        return c == ' ' || c == '\t' || c == '\n' || c == '\r'
    default:
        return false
    }
}
```

### Rule Struct

A single validation predicate:

```go
type ValidationRule struct {
    Op      RuleOp    `json:"op"`                // operation to perform
    Value   string    `json:"value,omitempty"`    // literal string (for contains, starts_with, ends_with, equals)
    Values  []string  `json:"values,omitempty"`   // set of values (for one_of)
    Class   CharClass `json:"class,omitempty"`    // character class (alternative to Value for contains/count)
    Cmp     Cmp       `json:"cmp,omitempty"`      // comparison operator (required for length, count, range, item_count)
    N       *float64  `json:"n,omitempty"`        // numeric operand (for length, count, range, item_count)
    Negate  bool      `json:"negate,omitempty"`   // invert the result (contains → not_contains)
    Message string    `json:"message,omitempty"`  // custom error message (overrides default)
}
```

### Rule Groups (Composition)

Rules compose via groups. A group applies a logical operator across its children. Children can be rules or nested groups.

```go
type RuleEntry struct {
    Rule  *ValidationRule `json:"rule,omitempty"`
    Group *RuleGroup      `json:"group,omitempty"`
}

type RuleGroup struct {
    AllOf []RuleEntry `json:"all_of,omitempty"` // AND — all must pass
    AnyOf []RuleEntry `json:"any_of,omitempty"` // OR  — at least one must pass
}
```

Invariant: exactly one of `AllOf` or `AnyOf` is populated. Exactly one of `Rule` or `Group` is populated per `RuleEntry`.

### Replaced ValidationConfig

```go
type ValidationConfig struct {
    Rules []RuleEntry `json:"rules,omitempty"`
}
```

All previous flat fields (`Required`, `MinLength`, `MaxLength`, `Min`, `Max`, `Pattern`, `MaxItems`) are removed. Their equivalents as composable rules:

| Old Field       | Equivalent Rule                                            |
|-----------------|------------------------------------------------------------|
| `required: true` | `{"rule": {"op": "required"}}`                            |
| `min_length: 5` | `{"rule": {"op": "length", "cmp": "gte", "n": 5}}`       |
| `max_length: 100`| `{"rule": {"op": "length", "cmp": "lte", "n": 100}}`     |
| `min: 0`        | `{"rule": {"op": "range", "cmp": "gte", "n": 0}}`        |
| `max: 100`      | `{"rule": {"op": "range", "cmp": "lte", "n": 100}}`      |
| `max_items: 5`  | `{"rule": {"op": "item_count", "cmp": "lte", "n": 5}}`   |
| `pattern: "..."` | Was never enforced — no equivalent needed                 |

`N` is `*float64` to support decimal thresholds (e.g., price >= 0.01, temperature <= 99.9). For integer-semantic ops (`length`, `count`, `item_count`), the value is truncated to int at evaluation time. `ValidateRuleDefinition` rejects fractional `N` for these ops.

### JSON Examples

**Contains exact substring:**
```json
{"rules": [{"rule": {"op": "contains", "value": "match"}}]}
```
Equivalent to user syntax: `(contains=match)`

**Does NOT contain substring:**
```json
{"rules": [{"rule": {"op": "contains", "value": "matches", "negate": true}}]}
```
Equivalent to user syntax: `(contains!=matches)`

**Length greater than 10:**
```json
{"rules": [{"rule": {"op": "length", "cmp": "gt", "n": 10}}]}
```
Equivalent to user syntax: `(length>10)`

**Contains symbols AND symbol count > 5 (compound AND):**
```json
{
  "rules": [{
    "group": {
      "all_of": [
        {"rule": {"op": "contains", "class": "symbols"}},
        {"rule": {"op": "count", "class": "symbols", "cmp": "gt", "n": 5}}
      ]
    }
  }]
}
```
Equivalent to user syntax: `(contains=#symbols&&count>5)`

**Required field:**
```json
{"rules": [{"rule": {"op": "required"}}]}
```

**Numeric range (0.01 to 99.99):**
```json
{
  "rules": [
    {"rule": {"op": "range", "cmp": "gte", "n": 0.01}},
    {"rule": {"op": "range", "cmp": "lte", "n": 99.99}}
  ]
}
```

**Max items in a list:**
```json
{"rules": [{"rule": {"op": "item_count", "cmp": "lte", "n": 5}}]}
```

**Password strength:**
```json
{
  "rules": [
    {"rule": {"op": "required"}},
    {"rule": {"op": "length", "cmp": "gte", "n": 8}},
    {"rule": {"op": "contains", "class": "uppercase"}},
    {"rule": {"op": "contains", "class": "lowercase"}},
    {"rule": {"op": "contains", "class": "digits"}},
    {
      "group": {
        "any_of": [
          {"rule": {"op": "count", "class": "symbols", "cmp": "gte", "n": 1}},
          {"rule": {"op": "length", "cmp": "gte", "n": 16}}
        ]
      }
    }
  ]
}
```
This means: required, min 8 chars, must have uppercase + lowercase + digits, and EITHER at least 1 symbol OR be at least 16 chars long.

**Top-level rules are implicitly AND'd.** The top-level `Rules` array is treated as an implicit `all_of` — every entry must pass.

### Validation of Rule Definitions

Three entry points for validating rule definitions:

```go
func ValidateRuleDefinition(r ValidationRule) error           // validates a single flat rule
func ValidateRuleEntries(entries []RuleEntry, depth int) error // validates entries + groups recursively
func ValidateValidationConfig(vc ValidationConfig) error      // top-level entry point (calls ValidateRuleEntries with depth 0)
```

`ValidateRuleDefinition(r ValidationRule) error` checks that a single rule is well-formed:

- `Op` must be a valid `RuleOp`
- `required`: all other fields must be empty (no Value, Class, Cmp, N, Values). The `required` op is only valid at the top level — `ValidateRuleEntries` rejects `required` inside a `RuleGroup` (nested in `AllOf`/`AnyOf`)
- `contains`, `starts_with`, `ends_with`: exactly one of `Value` or `Class` must be set
- `equals`: `Value` must be set, `Class` must be empty
- `length`: `Cmp` and `N` must be set, `Value`/`Class` must be empty
- `count`: `Cmp` and `N` must be set, exactly one of `Value` or `Class` must be set
- `range`: `Cmp` and `N` must be set, `Value`/`Class` must be empty (operates on numeric field value)
- `item_count`: `Cmp` and `N` must be set, `Value`/`Class` must be empty (counts comma-separated or JSON array items)
- `one_of`: `Values` must be non-empty, `Value`/`Class` must be empty
- `Negate` is only valid on `contains`, `starts_with`, `ends_with`, `equals`, `one_of`
- `Class` must be a valid `CharClass` when set
- `Cmp` must be a valid `Cmp` when set
- `N` must be non-negative when set
- For `length`, `count`, `item_count`: `N` must be a whole number (no fractional part)

`ValidateRuleEntries(entries []RuleEntry, depth int) error` validates a slice of entries recursively:

- Each `RuleEntry` must have exactly one of `Rule` or `Group` populated
- If `Rule` is set → call `ValidateRuleDefinition`
- If `Group` is set → the group must have exactly one of `AllOf` or `AnyOf` populated, with at least one entry. Recurse into group children with `depth + 1`
- Reject `required` op inside groups (any depth > 0)
- Maximum nesting depth of 10 levels — reject deeper trees

---

## Phase 2: Core Validation Package

Create `internal/validation/` with four files.

### errors.go — Structured error types

```go
type FieldError struct {
    FieldID  types.FieldID `json:"field_id"`
    Label    string        `json:"label"`
    Messages []string      `json:"messages"`
}

type ValidationErrors struct {
    Fields []FieldError `json:"fields"`
}
```

- `FieldError` implements `error` (joins messages with `; `)
- `ValidationErrors` implements `error`, plus:
  - `HasErrors() bool`
  - `ForField(types.FieldID) *FieldError`
  - `ClearField(types.FieldID)` — removes the FieldError for a specific field (used by TUI per-field error clearing)
- JSON-marshals cleanly for HTTP 422 responses

### validate.go — Public API

```go
type FieldInput struct {
    FieldID    types.FieldID
    Label      string
    FieldType  types.FieldType
    Value      string   // submitted value
    Validation string   // raw JSON from fields.validation column
    Data       string   // raw JSON from fields.data column
}

func ValidateField(input FieldInput) *FieldError
func ValidateBatch(inputs []FieldInput) ValidationErrors
```

Flow per field:
1. Parse `ValidationConfig` via `types.ParseValidationConfig`. If parsing fails, return a `FieldError` with message `invalid validation configuration: <parse error>`. Do not silently pass or panic
2. Scan rules for a `required` rule
3. If no `required` rule and value is empty → return nil (skip all checks)
4. If `required` rule and value is empty → return "required" error immediately
5. Run type-specific validator
6. Run composable rules via `EvaluateRules` (skips `required` op since already handled)
7. Collect messages, return `*FieldError` or nil

### type_validators.go — Per-type validation

| Field Type               | Validation                                                               | Reuses                                           |
|--------------------------|--------------------------------------------------------------------------|--------------------------------------------------|
| text, textarea, richtext | none (length via config rules)                                           | —                                                |
| number                   | strconv.ParseFloat                                                       | —                                                |
| date                     | time.Parse("2006-01-02")                                                 | —                                                |
| datetime                 | time.Parse (RFC3339, then 2006-01-02T15:04:05, then 2006-01-02 15:04:05) | —                                                |
| boolean                  | must be "true", "false", "1", or "0"                                     | —                                                |
| select                   | membership check against options from fields.data JSON                   | independent parser (see note below)              |
| email                    | delegate to types.Email.Validate()                                       | types_validation.go                              |
| url                      | delegate to types.URL.Validate()                                         | types_validation.go                              |
| slug                     | delegate to types.Slug.Validate()                                        | types_validation.go                              |
| media                    | validate ULID format via types.MediaID.Validate()                        | types_ids.go                                     |
| relation                 | validate ULID format via types.ContentID.Validate()                      | types_ids.go                                     |
| json                     | json.Valid()                                                             | stdlib                                           |

**Select validator note:** The select validator parses `fields.data` JSON independently using `json.Unmarshal` into `[]struct{Label string; Value string}`, matching the same format as `SelectBubble.ParseOptionsFromData` in `internal/cli/`. Do not import from `internal/cli/` — the validation package must have no dependency on the TUI package.

### rules.go — Composable rule evaluator

```go
func EvaluateRules(value string, entries []types.RuleEntry) []string
```

Returns a list of error messages (empty if all rules pass). Evaluates the rule tree:

1. For each `RuleEntry` in the top-level slice (implicit AND):
   - If `Rule` is set → evaluate the single rule
   - If `Group` is set → evaluate the group recursively

2. Single rule evaluation by `Op`:

| Op           | Behavior                                                                                                 |
|--------------|----------------------------------------------------------------------------------------------------------|
| `required`   | Handled before rule evaluation (see validate.go flow). Skipped by EvaluateRules.                         |
| `contains`   | If `Value` set: `strings.Contains(value, rule.Value)`. If `Class` set: iterate runes, return true if any char matches class. |
| `starts_with`| If `Value` set: `strings.HasPrefix(value, rule.Value)`. If `Class` set: check first rune matches class.  |
| `ends_with`  | If `Value` set: `strings.HasSuffix(value, rule.Value)`. If `Class` set: check last rune matches class.   |
| `equals`     | `value == rule.Value`. Class not supported.                                                              |
| `length`     | Compare `utf8.RuneCountInString(value)` against `*rule.N` (truncated to int) using `rule.Cmp`.          |
| `count`      | If `Value` set: `strings.Count(value, rule.Value)`. If `Class` set: count runes matching class. Compare count against `*rule.N` using `rule.Cmp`. |
| `range`      | Parse value as float64 via `strconv.ParseFloat`. Compare against `*rule.N` using `rule.Cmp`. If parse fails, error: "must be a number". |
| `item_count` | Count items in value (see item_count semantics below). Compare count against `*rule.N` (truncated to int) using `rule.Cmp`. |
| `one_of`     | Check if `value` is in `rule.Values` slice (linear scan, case-sensitive). Case-sensitive by design — CMS field values are stored exactly as entered, and rule authors control the values list. |

**`item_count` semantics:**
- Empty string → count is 0
- Value starts with `[`: attempt `json.Unmarshal` into `[]json.RawMessage`. If unmarshal succeeds, count is the array length. If unmarshal fails, fall through to comma-separated counting.
- Otherwise: `strings.Split(value, ",")`, trimming whitespace from each segment, skipping empty segments. Count is the number of non-empty segments.

3. After evaluation, if `rule.Negate` is true, invert the boolean result.

4. Generate default error messages per operation:

| Op           | Default message                                             |
|--------------|-------------------------------------------------------------|
| `required`   | `is required`                                               |
| `contains`   | `must contain "X"` / `must contain X characters`            |
| `contains` (negated) | `must not contain "X"` / `must not contain X characters` |
| `starts_with`| `must start with "X"`                                       |
| `ends_with`  | `must end with "X"`                                         |
| `equals`     | `must equal "X"`                                            |
| `length`     | `must be {cmp} {n} characters` (e.g. "must be at least 10 characters") |
| `count`      | `must have {cmp} {n} occurrences of "X"` / `must have {cmp} {n} X characters` |
| `range`      | `value must be {cmp} {n}` (e.g. "value must be at least 0") |
| `item_count` | `must have {cmp} {n} items` (e.g. "must have at most 5 items") |
| `one_of`     | `must be one of: X, Y, Z`                                  |

If `rule.Message` is set, use it instead of the default.

5. Group evaluation:
   - `AllOf`: evaluate all entries, collect ALL error messages from any failing entry
   - `AnyOf`: evaluate all entries; if ALL fail, return the first failing entry's messages; if any passes, return no errors

### validation_test.go — Table-driven tests

No DB dependency. Test categories:

- Per-type validation (valid, invalid, edge cases)
- `required` rule (empty value, non-empty value, absent rule + empty value)
- Required + empty interaction (no required rule skips all checks)
- Batch validation (all valid, mixed, all invalid)
- Error type methods (HasErrors, ForField, Error string)
- `contains` with literal value (match, no match)
- `contains` with character class (uppercase, lowercase, digits, symbols, spaces)
- `contains` with `negate: true`
- `starts_with` / `ends_with` with literal value (match, no match, negated)
- `starts_with` / `ends_with` with character class (first/last rune classification)
- `equals` (exact match, case mismatch)
- `length` with each Cmp (eq, neq, gt, gte, lt, lte)
- `count` with literal value (count substrings)
- `count` with character class (count digits, symbols, etc.)
- `range` with each Cmp (valid number, invalid number, boundary)
- `item_count` with JSON array, comma-separated list, empty value (count 0), malformed JSON fallback to comma-separated, whitespace-only segments skipped
- `one_of` (in set, not in set, negated)
- `AllOf` group (all pass, one fails, all fail)
- `AnyOf` group (all pass, one passes, all fail)
- Nested groups (group inside group)
- Custom `message` override
- Rule definition validation (`ValidateRuleDefinition`)
- Rule definition rejects nesting depth > 10
- Rule definition rejects fractional N on integer-semantic ops (length, count, item_count)
- Rule definition rejects `required` op inside groups (`ValidateRuleEntries` at depth > 0)
- `ValidateValidationConfig` → `ValidateRuleEntries` → `ValidateRuleDefinition` full chain
- `ValidateField` with malformed Validation JSON returns `FieldError` with parse error message
- `ClearField` removes a single field's errors without affecting others
- Empty rules slice (no-op, passes)

---

## Phase 3: HTTP Integration

### internal/router/contentBatch.go

Insert validation between the existing field fetch and the upsert loop — after `routeID = contentData.RouteID` and the `authorID` derivation block, before `for fieldID, value := range req.Fields`.

1. Collect all fieldIDs from req.Fields into a slice
2. Fetch all field definitions in a single query via `d.GetFieldsByIDs(fieldIDs)` (add to DbDriver if missing — takes `[]types.FieldID`, returns `[]db.Fields`)
3. Build `[]validation.FieldInput` by matching definitions to submitted values
4. If any submitted fieldID has no matching definition → HTTP 400 with the unknown field IDs, return
5. Call `validation.ValidateBatch(inputs)`
6. If HasErrors() → respond HTTP 422 with JSON ValidationErrors body, return early
7. If no errors → proceed to existing upsert loop unchanged

### internal/router/contentFields.go

apiCreateContentField (decodes into `db.CreateContentFieldParams`):
- After JSON decode, check `newContentField.FieldID.Valid` is true and `newContentField.FieldID.ID` is not zero. The params struct uses `types.NullableFieldID`, not a bare `FieldID`
- If invalid → HTTP 400 "field_id is required", return
- Call `d.GetField(newContentField.FieldID.ID)` to get field definition
- If GetField returns no rows → HTTP 404 "field not found", return
- Build FieldInput, call ValidateField
- If error → wrap as `ValidationErrors{Fields: []FieldError{*err}}`, respond HTTP 422 with JSON, return
- Otherwise proceed to d.CreateContentField

apiUpdateContentField (decodes into `db.UpdateContentFieldParams`):
- Same pattern using `updateContentField.FieldID.Valid` / `.ID`: decode → guard FieldID → lookup field → validate → wrap single FieldError → proceed or 422/400/404

### New helper: writeValidationError

Add a small helper in contentBatch.go (or a shared file) to write 422 responses consistently:

```go
func writeValidationError(w http.ResponseWriter, ve validation.ValidationErrors, logger *slog.Logger) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusUnprocessableEntity)
    if err := json.NewEncoder(w).Encode(ve); err != nil {
        logger.Error("failed to encode validation error response", "error", err)
    }
}
```

### GetField on DbDriver

Verify GetField(types.FieldID) exists on DbDriver. If not, it needs to be added. (Exploration confirmed field CRUD exists — GetField should already be present.)

---

## Phase 4: TUI Integration

### Expand ContentFieldInput struct

In internal/cli/form_dialog.go, add validation/data JSON to ContentFieldInput:

```go
type ContentFieldInput struct {
    FieldID        types.FieldID
    Label          string
    Type           string // field type (text, textarea, number, etc.)
    Widget         string // UI widget override from UIConfig (e.g. "markdown", "code-editor")
    Bubble         FieldBubble
    ValidationJSON string  // raw JSON from fields.validation
    DataJSON       string  // raw JSON from fields.data
}
```

### Expand ExistingContentField struct

```go
type ExistingContentField struct {
    ContentFieldID types.ContentFieldID
    FieldID        types.FieldID
    Label          string
    Type           string
    Widget         string // UI widget override from UIConfig
    Value          string
    ValidationJSON string
    DataJSON       string
}
```

### Populate new fields at dialog creation

NewContentFormDialog (the `func NewContentFormDialog(title string, action FormDialogAction, ...)` constructor): Already iterates []db.Fields. Add:
`ValidationJSON: f.Validation,`
`DataJSON:       f.Data,`

NewEditContentFormDialog (the `func NewEditContentFormDialog(title string, contentID types.ContentID, ...)` constructor): Already iterates []ExistingContentField. The caller (`HandleFetchContentForEdit` in `commands.go`) must populate the new fields from the db.Fields lookup.

SelectBubble data passing: Already handled — NewEditContentFormDialog needs DataJSON for select options (currently parsed inline, will also be stored for validation).

### Add validation state to ContentFormDialogModel

```go
type ContentFormDialogModel struct {
    // ... existing fields ...
    ValidationErrors *validation.ValidationErrors
}
```

### Pre-submit validation in Update

At the confirm button press (inside the `if d.focusIndex == d.ButtonConfirmIndex()` branch in `Update`, before emitting `ContentFormDialogAcceptMsg`):

1. Build []validation.FieldInput from d.Fields
2. Call validation.ValidateBatch
3. If errors → store on d.ValidationErrors, return (no msg emitted)
4. If valid → clear d.ValidationErrors, proceed to emit accept msg

### Error display in Render

In the field rendering loop (`for i, f := range d.Fields` inside `Render`), after each field bubble:
- Check d.ValidationErrors.ForField(f.FieldID)
- If present, render error messages in red below the field input
- On field value change, clear only that field's errors via `d.ValidationErrors.ClearField(f.FieldID)` (add method to ValidationErrors — removes the matching FieldError entry from the Fields slice)

---

## Files to Create

| File                                       | Purpose                                     |
|--------------------------------------------|---------------------------------------------|
| internal/validation/errors.go              | FieldError, ValidationErrors types           |
| internal/validation/validate.go            | ValidateField, ValidateBatch, FieldInput     |
| internal/validation/type_validators.go     | Per-FieldType validators                     |
| internal/validation/rules.go               | Composable rule evaluator (EvaluateRules)    |
| internal/validation/validation_test.go     | Table-driven tests (types + rules)           |

## Files to Modify

| File                                         | Change                                                                                    |
|----------------------------------------------|-------------------------------------------------------------------------------------------|
| internal/db/types/types_field_config.go      | Replace ValidationConfig: remove all flat fields; add RuleOp, Cmp, CharClass, ValidationRule, RuleEntry, RuleGroup types |
| internal/db/types/types_field_config_test.go | Delete existing TestParseValidationConfig tests for old flat fields; write new tests for rules-based ValidationConfig |
| internal/db/db.go                            | Add `GetFieldsByIDs(ctx context.Context, ids []types.FieldID) ([]Fields, error)` to DbDriver interface |
| internal/db/field.go                         | Implement GetFieldsByIDs on all three wrapper structs                                     |
| sql/schema/8_fields/queries.sql              | Add `GetFieldsByIDs` query (SQLite). After adding, run `just sqlc` to regenerate          |
| sql/schema/8_fields/queries_mysql.sql        | Add `GetFieldsByIDs` query (MySQL)                                                        |
| sql/schema/8_fields/queries_psql.sql         | Add `GetFieldsByIDs` query (PostgreSQL)                                                   |
| internal/router/contentBatch.go              | Add validation with single multi-ID field fetch before upsert loop                        |
| internal/router/contentFields.go             | Add validation in create/update handlers with FieldID null/zero guard                     |
| internal/cli/form_dialog.go                 | Expand structs, add validation state, pre-submit check, per-field error rendering/clearing |
| internal/cli/commands.go                    | In `HandleFetchContentForEdit`, populate `ValidationJSON` and `DataJSON` on each `ExistingContentField` from the corresponding `db.Fields` lookup |

## Files to Read (reference, existing validators to reuse)

| File                                    | What                                              |
|-----------------------------------------|---------------------------------------------------|
| internal/db/types/types_field_config.go | ValidationConfig, ParseValidationConfig            |
| internal/db/types/types_validation.go   | Email.Validate(), URL.Validate(), Slug.Validate()  |
| internal/db/types/types_enums.go        | FieldType enum constants                           |
| internal/db/types/types_ids.go          | ID .Validate() methods for media/relation          |
| internal/db/field.go                    | Fields struct (Validation, Data, Type fields)      |

---

## Evaluation Order

Per field, validation runs in this order:

```
1. Required scan        → scan rules for a "required" op
2. Empty gate           → if no required rule and empty value, pass immediately (skip all)
3. Required check       → if required rule and empty value, fail immediately
4. Type-specific validator → email format, URL format, number parsing, etc.
5. Composable rules     → EvaluateRules(value, config.Rules) — skips "required" op
```

All errors from stages 4-5 are collected (not short-circuited within a stage) and returned together, so the user sees all problems at once.

---

## Verification

1. `go build ./internal/validation/` — compiles
2. `go test -v ./internal/validation/` — all table-driven tests pass
3. `go build ./...` — full project compiles with handler + TUI changes
4. `just test` — existing test suite still passes
5. Manual test via API: POST invalid field value → 422 with structured JSON errors
6. Manual test via API: POST field with composable rules → 422 on rule violation
7. Manual test via TUI: submit content form with invalid values → inline error messages shown, form not submitted

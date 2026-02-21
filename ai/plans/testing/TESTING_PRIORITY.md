# Untested Package Testing Priority

Packages with no test files, ordered by priority.

## High Priority

### `internal/tree` (1 file, high complexity)

The most algorithmically complex package. Implements a four-phase tree loader with orphan resolution, cycle detection, and sibling-pointer reordering.

- `LoadFromRows` -- happy path, random arrival order, orphans resolved by retry, circular reference detection, unresolvable orphans
- `DeleteNodeByIndex` / `DeleteFirstChild` / `DeleteNestedChild` -- four delete sub-cases (first-child vs nested x leaf vs has-children). Possible latent bug: `deleteNestedChildHasChildren` returns `false` when `target.PrevSibling == nil`
- `CountVisible`, `NodeAtIndex`, `FlattenVisible`, `FindVisibleIndex` -- boundary conditions with collapsed/expanded nodes
- `IsDescendantOf` -- self, ancestor, deep child

### `internal/transform` (8 files, high complexity)

Five transformer implementations with bugs in private helpers.

- `parseFloat` -- broken decimal-tracking logic produces wrong results
- `parseInt` -- ignores negative signs
- `fieldLabelToKey` -- hand-rolled camelCase converter with edge cases
- `pluralize` -- fails on "bus", "key", etc.
- `CleanTransformer.Parse` / `ContentfulTransformer.Parse` -- round-trip fidelity
- `TransformConfig.GetTransformer()` -- factory returns correct type per format
- `ParseRequest` -- silently truncates when `ContentLength` is wrong or -1
- Stub `Parse` methods for Sanity/Strapi/WordPress -- document they return errors

### `internal/utility` (11 files, low-medium complexity but wide surface)

Many small utilities used everywhere.

- `TrimStringEnd` -- **panics** if `l > len(str)` (no bounds check)
- `NullToString[T]` / `IsNull[T]` -- all 8 `sql.Null*` variants
- `ParseDBTimestamp` -- SQLite RFC3339, Unix timestamp, MySQL/Psql format, invalid input
- `FormatTimestampForDB` -- all three driver types
- `IsValidEmail` -- edge cases (missing @, no TLD, special chars)
- `Metrics` -- concurrent increments (race detector), `keyWithLabels` has nondeterministic map iteration (correctness bug)
- `SizeInBytes`, `IsInt`, `MakeRandomString`, `FileExists`/`DirExists`

### `internal/update` (2 files, medium complexity)

- `CompareVersions` -- pure semver comparison: `v` prefix stripping, pre-release suffixes, `"dev"`/`"unknown"`, unequal segment counts
- `GetDownloadURL` -- nil release, no matching asset, exact name match
- `VerifyBinary` -- size bounds (<1MB / >500MB), executable bit
- `DownloadUpdate` / `ApplyUpdate` / `RollbackUpdate` -- filesystem operations (testable with temp dirs and `httptest`)

## Medium Priority

### `internal/tui` (9 files, low-medium complexity)

Skeleton TUI, but has testable pure functions.

- `Composite` / `spliceLine` -- ANSI-aware string overlay, the most testable part
- `padContent` -- padding/truncation by dimensions
- `FocusPanel.String()` -- all values including default
- `Update()` -- `q`/`ctrl+c` -> quit, `tab`/`shift+tab` -> focus cycling, `WindowSizeMsg` -> dimension update

### `tools/dbgen` (4 files, medium complexity)

Code generation tool with pure, testable methods.

- `Entity` methods: `StructFields()`, `CreateFields()`, `UpdateFields()`, `IDIsTyped()`, `IDToString()`, `Needs*Import()`
- Template functions: `lower`, `stringExpr`, `wrapParam`, `paginationCast`, `driverLabel`
- End-to-end: `--dry-run` / `--verify` modes

### `tools/transform_cud` (1 file, medium complexity)

- `findFuncEnd` -- brace-balanced parser handling string literals and raw strings. Most testable piece
- `transformCreate`/`transformUpdate`/`transformDelete` -- old-to-new signature transformation
- `getWrapperType` -- entity name mapping

## Low Priority

### `tools/transform_bootstrap` (1 file, low complexity)

One-shot migration tool that has already been applied. Only worth testing if you plan to run it again. The only testable function is verifying that given old-pattern input, the 23 string replacements produce expected output.

## Known Bugs Found During Exploration

1. `internal/transform`: `parseFloat` has broken decimal logic, `parseInt` ignores negatives
2. `internal/utility`: `TrimStringEnd` panics on out-of-bounds input
3. `internal/utility`: `Metrics.keyWithLabels` uses nondeterministic map iteration for cache keys
4. `internal/tree`: possible silent failure in `deleteNestedChildHasChildren` when `PrevSibling == nil`

# Testing Priority

Updated 2026-03-09. Previous version covered packages that had zero tests; most now have test files. This revision focuses on **unfixed bugs**, **coverage gaps**, and **new untested code**.

## Unfixed Bugs (fix first)

### 1. `internal/utility` — `TrimStringEnd` panics on out-of-bounds input

**File:** `string.go:14-21`

`TrimStringEnd("hi", 5)` panics with `slice bounds out of range [-3:]`. The function checks `len(str) > 0` but not `l <= len(str)`. Existing tests cover valid inputs but not this edge case.

**Fix:** Add bounds check before slicing. Add test case for `l > len(str)`.

### 2. `internal/utility` — `Metrics.keyWithLabels` nondeterministic cache keys

**File:** `metrics.go:156-166`

Iterates over a `map[string]string` to build cache keys. Map iteration order is random in Go, so `keyWithLabels("foo", {"a":"1","b":"2"})` can return `"foo,a=1,b=2"` or `"foo,b=2,a=1"`. This means the same metric+labels combination can create duplicate entries.

Comment at line 160 acknowledges the issue: `"Simple key generation - in production, use stable sorting"`.

**Fix:** Sort label keys before building the key string. Add test case with multiple labels verifying deterministic output.

## Coverage Gaps

### `internal/tui` (89 files, 6 test files)

The TUI has grown from a 9-file skeleton to a fully-featured application. Current test files cover specific subsystems (bubble fields, content/datatype/media trees, screen_media), but the majority of screens and the core update/view logic are untested.

**High-value test targets (pure functions, no Bubbletea dependency):**
- `Composite` / `spliceLine` (layer.go) — ANSI-aware string overlay
- `padContent` (panel.go) — padding/truncation by dimensions
- Screen-specific data formatting and validation helpers
- Dialog field validation logic
- Admin command builders

**Note:** The TUI QA milestone (#133-151) covers functional QA per-screen. This section covers unit test coverage for pure functions.

### `internal/tree/ops` (new package, 7 files)

Added 2026-03-08. Implements tree mutation operations (insert, move, reorder, save, unlink) extracted from the main tree package. Has `tree_test.go` but coverage depth is unknown — verify edge cases for:
- `Insert` — duplicate IDs, nil parent, circular insertion
- `Move` — move to descendant (should fail), move to same position (no-op)
- `Reorder` — empty list, single item, item not found
- `Unlink` — root node, node with children, node not in tree

### `internal/tree` — delete edge cases

The `deleteNestedChildHasChildren` function (tree.go:333-347) returns false when `PrevSibling == nil && NextSibling != nil`. The existing test at tree_test.go:974 documents this as expected behavior. In practice, `DeleteNodeByIndex` dispatches first-child cases before reaching this path, so `PrevSibling == nil` indicates inconsistent tree state.

**Verify:** The defensive return-false is correct and the caller (`DeleteNodeByIndex`) handles it properly. The test should assert the caller's behavior, not just the internal function's.

## Fully Tested (no action needed)

These packages were listed in the previous version as untested. They now have comprehensive test coverage:

| Package | Test Files | Status |
|---------|-----------|--------|
| `internal/tree` | `tree_test.go`, `tree_admin_test.go` | Covered |
| `internal/transform` | `cms_formats_test.go`, `raw_clean_test.go`, `transformer_test.go` | Covered |
| `internal/utility` | 12 test files (string, null, timestamp, metrics, consts, version, observability, certs, fuzzy, status) | Covered (except bugs above) |
| `internal/update` | `update_test.go` | Covered |
| `tools/dbgen` | 4 test files (definitions, entity, funcmap, main) | Covered |
| `tools/transform_cud` | `main_test.go` | Covered |
| `tools/transform_bootstrap` | `main_test.go` | Covered |

## Removed from Plan

These items appeared in the previous version but no longer apply:

- `internal/transform` bugs (`parseFloat`, `parseInt`, `pluralize`, `fieldLabelToKey`, `ParseRequest`, `TransformConfig.GetTransformer()`) — these functions were removed during a package refactor
- `internal/utility` functions (`FileExists`, `DirExists`, `SizeInBytes`) — removed or moved
- `internal/tui` `FocusPanel.String()` — type no longer exists

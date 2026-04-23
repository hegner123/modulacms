# Test Quality Audit

Package-by-package analysis of test quality, depth, and adversarial coverage.

**Started:** 2026-04-21
**Status:** In progress

## Scoring

Each package is rated on five dimensions (1-5 scale):

| Dimension | 1 (Poor) | 3 (Adequate) | 5 (Strong) |
|-----------|----------|--------------|------------|
| **Coverage** | Only happy path | Main paths + some edges | All paths including error/edge |
| **Adversarial** | No assumption testing | Some boundary checks | Systematically attacks abstractions |
| **Isolation** | Tests depend on each other or global state | Mostly independent | Fully independent, parallel-safe |
| **Assertions** | Weak or missing checks | Checks return values | Validates state, side effects, error messages |
| **Maintainability** | Copy-paste, magic values, fragile | Helpers exist, some structure | Table-driven, clear setup/teardown, readable |

**Overall grade:** average of five dimensions, rounded.

## Package inventory

41 packages, 3,534 test functions, ~140k lines of test code.

### Tier 1: Core data layer (highest risk)

| Package | Files | Tests | Lines | Status | Grade |
|---------|-------|-------|-------|--------|-------|
| `internal/db` | 71 | 1613 | 59,462 | pending | - |
| `internal/db/types` | 12 | 182 | 5,074 | pending | - |
| `internal/db/audited` | 1 | 22 | 1,223 | pending | - |
| `internal/db/dbmetrics` | 3 | 8 | 578 | pending | - |

### Tier 2: Business logic (high risk)

| Package | Files | Tests | Lines | Status | Grade |
|---------|-------|-------|-------|--------|-------|
| `internal/auth` | 3 | 48 | 1,614 | pending | - |
| `internal/middleware` | 5 | 39 | 1,205 | pending | - |
| `internal/publishing` | 1 | 33 | 1,584 | pending | - |
| `internal/service` | 5 | 97 | 3,531 | pending | - |
| `internal/tree` | 2 | 59 | 1,851 | pending | - |
| `internal/tree/core` | 2 | 33 | 1,276 | pending | - |
| `internal/tree/ops` | 3 | 52 | 1,728 | pending | - |
| `internal/query` | 3 | 14 | 358 | pending | - |
| `internal/validation` | 1 | 28 | 1,133 | pending | - |
| `internal/search` | 7 | 44 | 1,724 | pending | - |

### Tier 3: Infrastructure (medium risk)

| Package | Files | Tests | Lines | Status | Grade |
|---------|-------|-------|-------|--------|-------|
| `internal/config` | 11 | 99 | 3,115 | pending | - |
| `internal/router` | 2 | 24 | 937 | pending | - |
| `internal/mcp` | 5 | 116 | 3,048 | done | 3.4 |
| `internal/media` | 6 | 53 | 2,698 | pending | - |
| `internal/backup` | 2 | 47 | 2,004 | pending | - |
| `internal/bucket` | 2 | 15 | 570 | pending | - |
| `internal/deploy` | 4 | 83 | 3,040 | pending | - |
| `internal/remote` | 4 | 72 | 1,354 | pending | - |
| `internal/email` | 5 | 40 | 1,378 | pending | - |
| `internal/webhooks` | 1 | 36 | 955 | pending | - |
| `internal/plugin` | 30 | 687 | 23,748 | pending | - |
| `internal/plugin/testing` | 1 | 18 | 387 | pending | - |

### Tier 4: Presentation (lower risk)

| Package | Files | Tests | Lines | Status | Grade |
|---------|-------|-------|-------|--------|-------|
| `internal/tui` | 11 | 86 | 2,538 | pending | - |
| `internal/admin` | 2 | 23 | 469 | pending | - |
| `internal/admin/handlers` | 1 | 23 | 435 | pending | - |
| `internal/transform` | 3 | 93 | 3,048 | pending | - |

### Tier 5: Supporting packages

| Package | Files | Tests | Lines | Status | Grade |
|---------|-------|-------|-------|--------|-------|
| `internal/utility` | 11 | 125 | 3,507 | pending | - |
| `internal/model` | 1 | 3 | 262 | pending | - |
| `internal/registry` | 1 | 14 | 794 | pending | - |
| `internal/install` | 5 | 33 | 986 | pending | - |
| `internal/update` | 2 | 33 | 1,177 | pending | - |
| `internal/definitions` | 3 | 19 | 761 | pending | - |
| `cmd` | 3 | 63 | 2,213 | pending | - |

### Tier 6: SDKs and tools

| Package | Files | Tests | Lines | Status | Grade |
|---------|-------|-------|-------|--------|-------|
| `sdks/go` | 6 | 30 | 1,298 | pending | - |
| `tools/dbgen` | 4 | 63 | 1,804 | pending | - |
| `tools/transform_bootstrap` | 1 | 8 | 455 | pending | - |
| `tools/transform_cud` | 1 | 27 | 888 | pending | - |

## Completed analyses

### `internal/mcp` (Grade: 3.4)

- **Coverage: 3** -- all middleware paths tested, permission map completeness verified, but no tool handler tests
- **Adversarial: 4** -- bootstrap cross-ref, case sensitivity, concurrency, malformed input, nil permission set, wrong-resource checks (added 2026-04-20)
- **Isolation: 4** -- tests are independent, no shared mutable state
- **Assertions: 3** -- checks error text and IsError flag, but no deep state inspection
- **Maintainability: 3** -- uses table-driven for some tests, helpers exist (resultText, buildReq, passthrough), but older tests are one-off

**Gaps:** No tests for individual tool handlers (all 246 tools). Tool registration correctness is tested via completeness check, but actual tool execution paths are untested in this package (they exercise the db layer via integration).

## Red flags to investigate

Packages with suspicious ratios suggesting shallow or generated tests:

- `internal/db`: 71 files, 1613 tests, 59k lines. Likely generated per-entity test files. Check if these are meaningful or boilerplate.
- `internal/plugin`: 30 files, 687 tests, 24k lines. Large test surface. Check for depth vs breadth.
- `internal/model`: 1 file, 3 tests, 262 lines. Tiny. Model package is a domain struct package, so 3 tests is potentially fine, but verify.
- `internal/query`: 3 files, 14 tests, 358 lines. Query builder is complex logic, 14 tests seems low.
- `internal/router`: 2 files, 24 tests, 937 lines. Router registers 200+ endpoints. 24 tests may only cover structure.

## Analysis order

Priority by risk-to-coverage ratio. Packages where bugs would be most impactful and where test coverage seems thinnest relative to complexity:

1. `internal/query` -- complex logic, only 14 tests
2. `internal/model` -- domain structs, only 3 tests
3. `internal/router` -- 200+ endpoints, 24 tests
4. `internal/middleware` -- auth/RBAC, 39 tests
5. `internal/auth` -- authentication, 48 tests
6. `internal/tree` + `tree/core` + `tree/ops` -- tree operations, critical data structure
7. `internal/publishing` -- content publishing, 33 tests
8. `internal/db` -- massive, needs sampling strategy
9. `internal/service` -- business orchestration, 97 tests
10. `internal/config` -- 99 tests but 11 files, check depth
11. Remaining by tier order

# Refactor Plan: "ROOT" → "_root"

Rename the datatype classification value `ROOT` to `_root` across the entire codebase.
The leading underscore signals a system-reserved value distinct from user-created types.

## Scope

**In scope:** Every reference to `ROOT` as a datatype type classification.
**Out of scope:** Unrelated uses of "ROOT" (MYSQL_ROOT_PASSWORD, MINIO_ROOT_USER, Xcode SDKROOT, CLANG_WARN_OBJC_ROOT_CLASS). These are infrastructure/tooling constants and must not be touched.

## Execution Order

Phases must run in order. Within each phase, steps can run in parallel.

---

## Phase 1: SQL Schema Queries (source of truth)

These files drive `sqlc generate`. Change the WHERE clauses first.

| # | File | Replacement |
|---|------|-------------|
| 1 | `sql/schema/7_datatypes/queries.sql` | `'ROOT'` → `'_root'` |
| 2 | `sql/schema/7_datatypes/queries_mysql.sql` | `'ROOT'` → `'_root'` |
| 3 | `sql/schema/7_datatypes/queries_psql.sql` | `'ROOT'` → `'_root'` |
| 4 | `sql/schema/9_admin_datatypes/queries.sql` | `'ROOT'` → `'_root'` |
| 5 | `sql/schema/9_admin_datatypes/queries_mysql.sql` | `'ROOT'` → `'_root'` |
| 6 | `sql/schema/9_admin_datatypes/queries_psql.sql` | `'ROOT'` → `'_root'` |
| 7 | `sql/schema/22_joins/queries.sql` | `'ROOT'` → `'_root'` |
| 8 | `sql/schema/22_joins/queries_mysql.sql` | `'ROOT'` → `'_root'` |
| 9 | `sql/schema/22_joins/queries_psql.sql` | `'ROOT'` → `'_root'` |

**repfor command (all 9 files share the same literal):**
```
repfor: search="'ROOT'" replace="'_root'" across all sql/schema/**/queries*.sql files
```

**Verify:** `checkfor 'ROOT'` in `sql/schema/` returns 0 matches.

---

## Phase 2: Regenerate sqlc

```bash
just sqlc
```

This overwrites all files in:
- `internal/db-sqlite/queries.sql.go`
- `internal/db-mysql/queries_mysql.sql.go`
- `internal/db-psql/queries_psql.sql.go`

**Do NOT manually edit these files.** The `'ROOT'` → `'_root'` change propagates automatically from Phase 1.

**Verify:** `checkfor 'ROOT'` in `internal/db-sqlite/`, `internal/db-mysql/`, `internal/db-psql/` returns 0 matches (except Go function names like `ListDatatypeRoot` and `ListAdminDatatypeRoot` which are query identifiers, not the value).

---

## Phase 3: SQL Seed Data

| # | File | Replacement |
|---|------|-------------|
| 1 | `sql/seed_admin_panel.sql` | `'ROOT'` → `'_root'` (line 83, INSERT value) |

**Verify:** `checkfor 'ROOT'` in `sql/seed_admin_panel.sql` returns 0 matches.

---

## Phase 4: Go Application Code — String Literals

All files where `"ROOT"` appears as a Go string literal value.

### 4a. Bootstrap/seed data in db.go (6 occurrences)

| File | Lines | Context |
|------|-------|---------|
| `internal/db/db.go` | 852, 884, 1721, 1753, 2561, 2593 | `Type: "ROOT"` in CreateBootstrapData for all 3 DB drivers |

**repfor:** `search='"ROOT"' replace='"_root"'` in `internal/db/db.go`

### 4b. TUI code (internal/cli/)

| File | Lines | What |
|------|-------|------|
| `internal/cli/forms.go` | 40 | `datatype = "ROOT"` default |
| `internal/cli/forms.go` | 43 | `"Optional - ROOT is reserved..."` description string |
| `internal/cli/forms.go` | 56 | `Key: "ROOT"` select option |
| `internal/cli/forms.go` | 114 | `"Optional - ROOT is reserved..."` description string |
| `internal/cli/forms.go` | 127 | `Key: "ROOT"` select option |
| `internal/cli/menus.go` | 74 | `item.Type == "ROOT"` filter check |
| `internal/cli/form_dialog.go` | 69 | Comment: `// DatatypeID or empty for ROOT` |
| `internal/cli/form_dialog.go` | 181 | `typeInput.Placeholder = "ROOT"` |
| `internal/cli/form_dialog.go` | 187 | `Label: "ROOT (no parent)"` |
| `internal/cli/form_dialog.go` | 750 | `typeInput.Placeholder = "ROOT"` |
| `internal/cli/form_dialog.go` | 757 | `Label: "ROOT (no parent)"` |
| `internal/cli/admin_update_dialog.go` | 251 | `dtype = "ROOT"` default |
| `internal/cli/update_dialog.go` | 948 | `Label: "ROOT (no parent)"` |
| `internal/cli/update_dialog.go` | 977 | `Label: "ROOT (no parent)"` |
| `internal/cli/update_dialog.go` | 1742 | `dtype = "ROOT"` default |

**repfor replacements for cli/ (run sequentially, unique strings):**

1. `"ROOT"` → `"_root"` (value comparisons and defaults) — affects forms.go:40, menus.go:74, admin_update_dialog.go:251, update_dialog.go:1742
2. `Key:   "ROOT"` → `Key:   "_root"` — affects forms.go:56, forms.go:127
3. `Placeholder = "ROOT"` → `Placeholder = "_root"` — affects form_dialog.go:181, form_dialog.go:750
4. `"ROOT (no parent)"` → `"_root (no parent)"` — affects form_dialog.go:187, form_dialog.go:757, update_dialog.go:948, update_dialog.go:977
5. `"Optional - ROOT is reserved for root content types.\n"` → `"Optional - _root is reserved for root content types.\n"` — affects forms.go:43, forms.go:114
6. Comment `for ROOT` → `for _root` — affects form_dialog.go:69

### 4c. Content definitions (internal/definitions/)

| File | Occurrences |
|------|-------------|
| `def_modulacms.go` | 5× `NewNullableString("ROOT")` |
| `def_contentful.go` | 3× `NewNullableString("ROOT")` |
| `def_wordpress.go` | 2× `NewNullableString("ROOT")` |
| `def_sanity.go` | 3× `NewNullableString("ROOT")` |
| `def_strapi.go` | 3× `NewNullableString("ROOT")` |
| `definition.go` | 1× comment `// Category: "page", "post", "ROOT", "GLOBAL"` |

**repfor:** `search='NewNullableString("ROOT")' replace='NewNullableString("_root")'` across all `internal/definitions/def_*.go` files. Handles 16 occurrences in one pass.

Then fix the comment in `definition.go` separately.

### 4d. Model/tree builder (internal/model/)

| File | Lines | What |
|------|-------|------|
| `build.go` | 106 | Comment: `Type == "ROOT"` |
| `build.go` | 110 | Comment: `Type "ROOT"` |
| `build.go` | 139 | `node.Datatype.Info.Type == "ROOT"` — **the core tree root detection** |

**repfor:** `search='"ROOT"' replace='"_root"'` in `internal/model/build.go` for line 139. Update comments on 106, 110 separately.

### 4e. Tests

| File | Lines | What |
|------|-------|------|
| `internal/model/build_test.go` | 47, 60, 73, 86, 95 | `"ROOT"` in makeDatatype calls |
| `internal/model/build_test.go` | 130 | `Info.Type != "ROOT"` assertion |
| `internal/db/datatype_crud_test.go` | 147, 157, 166, 169, 174, 180 | ROOT in comments and `Type: "ROOT"` |

**repfor:** `search='"ROOT"' replace='"_root"'` in both test files. Fix comments manually.

### 4f. Custom DB code

| File | Lines | What |
|------|-------|------|
| `internal/db/content_data_custom.go` | 55 | Comment: `// ROOT CONTENT SUMMARY` |

Comment-only change.

---

## Phase 5: Go Comments (non-literal references)

These are comments/docs in Go files that mention ROOT as a concept, not as a string value. Update for consistency.

| File | Line | Current | New |
|------|------|---------|-----|
| `internal/cli/menus.go` | 70 | `// BuildDatatypeMenu builds a menu of ROOT datatype pages` | `// BuildDatatypeMenu builds a menu of _root datatype pages` |
| `internal/cli/update_controls.go` | 576 | `// Left panel: content data instances with slug and ROOT datatype label` | `// Left panel: content data instances with slug and _root datatype label` |
| `internal/cli/panel_view.go` | 416 | `// renderRootDatatypesList renders ROOT datatypes` | `// renderRootDatatypesList renders _root datatypes` |
| `internal/cli/panel_view.go` | 684 | `// renderRootContentSummaryList renders content data instances with slug and ROOT datatype label` | `// renderRootContentSummaryList renders content data instances with slug and _root datatype label` |
| `internal/db/content_data_custom.go` | 55 | `// ROOT CONTENT SUMMARY` | `// _root CONTENT SUMMARY` |
| `internal/model/build.go` | 106 | `// 3. Identifying the root node (Type == "ROOT").` | `// 3. Identifying the root node (Type == "_root").` |
| `internal/model/build.go` | 110 | `// Returns the root Node pointer (nil if no node has Type "ROOT")` | `// Returns the root Node pointer (nil if no node has Type "_root")` |
| `internal/model/build.go` | 139 (nearby comment) | `// the one with Type "ROOT" becomes the tree root` | `// the one with Type "_root" becomes the tree root` |

---

## Phase 6: Documentation

### 6a. Internal package docs (.md files inside internal/)

| File | Occurrences |
|------|-------------|
| `internal/model/model.md` | 3× references to ROOT in BuildNodes description |
| `internal/cli/cli.md` | 1× "filtering for ROOT type" |

### 6b. AI reference docs (ai/)

| File | Occurrences |
|------|-------------|
| `ai/CONTENT_MODEL.md` | ~15× (heaviest file — section headers, code examples, explanations) |
| `ai/domain/DATATYPES_AND_FIELDS.md` | ~20× (type classification docs, SQL examples, Go examples, tables) |
| `ai/docs/MODEL_PACKAGE.md` | ~6× (code examples, test examples) |
| `ai/reference/QUICKSTART.md` | 1× `Type: "ROOT"` |
| `ai/reference/TROUBLESHOOTING.md` | 0× (only has MYSQL_ROOT_PASSWORD — skip) |

### 6c. SQL docs

| File | Occurrences |
|------|-------------|
| `sql/SCHEMA.md` | 1× "type categorizes datatype as GLOBAL or ROOT" |

**repfor approach for docs:** Run targeted passes:
1. `"ROOT"` → `"_root"` (quoted string in code examples)
2. `'ROOT'` → `'_root'` (SQL examples)
3. `= ROOT` context-specific replacements
4. Manual review for section headers and prose (e.g., "### 1. ROOT Datatypes" → "### 1. _root Datatypes")

---

## Phase 7: Test Fixtures

| File | Replacement |
|------|-------------|
| `test_artifacts/home_page_formatted.json` | `"type": "ROOT"` → `"type": "_root"` |

---

## Phase 8: CLAUDE.md Update

Update `CLAUDE.md` at project root if it references ROOT as a datatype type value (currently referenced in the Tri-Database Pattern section). Grep and update.

---

## Phase 9: Verification

1. **checkfor sweep:** `checkfor ROOT` across entire repo (excluding vendor/, deploy/docker/, ios/) — expect 0 matches
2. **Build:** `just check` — must compile cleanly
3. **sqlc parity:** `just sqlc` should produce no diff (already regenerated in Phase 2)
4. **Tests:** `just test` — all tests pass
5. **Manual spot-check:** Run TUI, create a datatype, verify `_root` appears in type dropdown and is stored correctly

---

## repfor Execution Summary

| Pass | search | replace | Scope |
|------|--------|---------|-------|
| 1 | `'ROOT'` | `'_root'` | `sql/schema/**/queries*.sql`, `sql/seed_admin_panel.sql` |
| 2 | `"ROOT"` | `"_root"` | `internal/db/db.go`, `internal/model/build.go`, `internal/model/build_test.go`, `internal/db/datatype_crud_test.go`, `internal/cli/menus.go`, `internal/cli/admin_update_dialog.go`, `internal/cli/update_dialog.go` |
| 3 | `NewNullableString("ROOT")` | `NewNullableString("_root")` | `internal/definitions/def_*.go` |
| 4 | `"ROOT (no parent)"` | `"_root (no parent)"` | `internal/cli/form_dialog.go`, `internal/cli/update_dialog.go` |
| 5 | `Placeholder = "ROOT"` | `Placeholder = "_root"` | `internal/cli/form_dialog.go` |
| 6 | `Key:   "ROOT"` | `Key:   "_root"` | `internal/cli/forms.go` |
| 7 | `datatype = "ROOT"` | `datatype = "_root"` | `internal/cli/forms.go` |
| 8 | `ROOT is reserved` | `_root is reserved` | `internal/cli/forms.go` |

Passes 2-8 can run in parallel. Pass 1 must complete before `just sqlc` (Phase 2).
Comments and documentation (Phases 5-6) are manual/targeted repfor passes after the code passes.

---

## Data Migration

Existing databases with `type = 'ROOT'` rows need a migration:

```sql
UPDATE datatypes SET type = '_root' WHERE type = 'ROOT';
UPDATE admin_datatypes SET type = '_root' WHERE type = 'ROOT';
```

Add this as a numbered schema migration in `sql/schema/` following the existing convention.

---

## Risk Notes

- **No SDK changes needed.** The TypeScript, Go, and Swift SDKs pass type as a generic string — they never hardcode `"ROOT"`. Confirmed via grep: 0 matches in `sdks/`.
- **No admin panel (templ/HTMX) changes needed.** Confirmed via grep: 0 matches in `internal/admin/`.
- **No router/middleware changes needed.** The type value is never checked in routing or authorization logic.
- **Docker/deploy files are untouched.** All ROOT references there are MYSQL_ROOT_PASSWORD / MINIO_ROOT_USER — completely unrelated.
- **Xcode project file untouched.** SDKROOT / CLANG_WARN_OBJC_ROOT_CLASS are Apple toolchain settings.

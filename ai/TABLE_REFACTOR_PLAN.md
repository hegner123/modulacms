# Phase 2: TableModel Extraction Plan

**Status:** ✅ COMPLETED - 2026-01-13

**Goal:** Extract TableModel from the Model struct (Phase 2 of 4-phase refactoring)
**Approach:** Follow FormModel pattern established in Phase 1
**Risk Level:** Medium - 78 references across 16 files
**Dependencies:** Phase 1 (FormModel) completed ✅

## Completion Summary

**Completed:** 2026-01-13

**Results:**
- ✅ Model struct reduced from 46 → 39 fields (15% reduction)
- ✅ Created `table_model.go` with TableModel struct and NewTableModel() constructor
- ✅ All 7 table fields extracted: Table, Headers, Rows, Columns, ColumnTypes, Selected, Row
- ✅ 66 references updated to use `m.TableState` across 15 files
- ✅ 0 old field references remaining (verified with checkfor)
- ✅ Clean build with no errors
- ✅ Dev binary built successfully

**Files Modified:** 16 total
- table_model.go (NEW, +30 lines)
- model.go (-7 fields, +1 TableState field)
- 14 files updated with TableState references

**Verification:**
- Build: Clean compilation
- Old references: 0 matches for all 7 fields
- New references: 66 `m.TableState` references
- Binary: `./modulacms-x86` ready for manual testing

**Next Phase:** Phase 3 - NavigationModel extraction

---

## Problem Statement

After Phase 1 successfully extracted FormModel (8 fields → 51 references), the Model struct still has 7 table-related fields scattered across 16 files:

**Current Pain Points:**
- Table data fields mixed with navigation, display, and form state
- 78 references across 16 files make table operations hard to track
- Database viewing/editing requires touching multiple unrelated fields
- Adding table features (sorting, filtering) would bloat Model further

**Why Extract These Fields:**
- Always used together for database table operations
- Clear semantic boundary (table data lifecycle)
- Completes CMS UI refactor (forms + tables)
- Model reduced from 46 → 39 fields

---

## What We're Extracting

**7 Table Fields → TableModel struct:**

```go
type TableModel struct {
    Table       string                    // Current table name
    Headers     []string                  // Table column headers
    Rows        [][]string                // Table data rows
    Columns     *[]string                 // Column names from DB
    ColumnTypes *[]*sql.ColumnType        // Column type metadata
    Selected    map[int]struct{}          // Selected rows (for multi-select)
    Row         *[]string                 // Currently selected single row
}
```

**Field Inventory (from checkfor):**
- `Table`: 32 matches across 10 files
- `Headers`: 9 matches across 5 files
- `Rows`: 15 matches across 6 files
- `Selected`: 1 match (debug.go only)
- `Columns`: 12 matches across 3 files
- `ColumnTypes`: 5 matches across 2 files
- `Row`: 4 matches across 3 files

**Total:** ~78 references across 16 files

---

## Implementation Steps

### Step 1: Create TableModel Struct
**File:** `internal/cli/table_model.go` (NEW, ~40 lines)

```go
package cli

import "database/sql"

type TableModel struct {
    Table       string
    Headers     []string
    Rows        [][]string
    Columns     *[]string
    ColumnTypes *[]*sql.ColumnType
    Selected    map[int]struct{}
    Row         *[]string
}

func NewTableModel() *TableModel {
    return &TableModel{
        Table:    "",
        Headers:  make([]string, 0),
        Rows:     make([][]string, 0),
        Selected: make(map[int]struct{}),
    }
}
```

### Step 2: Update Model Struct
**File:** `internal/cli/model.go`

**Remove 7 fields (lines 71, 78-83):**
```go
// DELETE:
// Table        string
// Columns      *[]string
// ColumnTypes  *[]*sql.ColumnType
// Selected     map[int]struct{}
// Headers      []string
// Rows         [][]string
// Row          *[]string
```

**Add 1 field (after line 84 - after FormState):**
```go
// ADD:
TableState   *TableModel
```

**Update InitialModel (line ~155):**
```go
m := Model{
    // ... existing fields ...
    FormState:  NewFormModel(),
    TableState: NewTableModel(),  // ADD
    // ... rest ...
}
```

**Verification after Step 2:**
```
checkfor tool with:
- dir: "/Users/home/Documents/Code/Go_dev/modulacms/internal/cli"
- search: "m.Table"
- ext: ".go"
- context: 0

# Expected: 32 references need updating
```

### Step 3: Update High-Usage Files (Most References)

#### 3a. Update update_controls.go (~17 changes)
**File:** `internal/cli/update_controls.go`

Pattern: `m.Table` → `m.TableState.Table`, `m.Rows` → `m.TableState.Rows`, `m.Row` → `m.TableState.Row`

Key sections:
- Lines 38-44: TableNavigationControls calls (4 references to m.Table indirectly)
- Line 245: `len(m.Tables)` (no change - different field!)
- Line 253: `m.Tables[m.Cursor]` (no change)
- Line 305: `if m.Cursor < len(m.Rows)-1` → `len(m.TableState.Rows)`
- Line 317: `if m.PageMod < (len(m.Rows)-1)/m.MaxRows` → `len(m.TableState.Rows)`
- Line 324: `if recordIndex < len(m.Rows)` → `len(m.TableState.Rows)`
- Line 365: `if m.PageMod < len(m.Rows)/m.MaxRows` → `len(m.TableState.Rows)`
- Line 371: `rows = m.Rows` → `m.TableState.Rows`
- Line 374: `if recordIndex < len(m.Rows)` → `len(m.TableState.Rows)`
- Line 377: `m.Row = &rows[recordIndex]` → `m.TableState.Row`
- Line 420: `db.DBTable(m.Table)` → `db.DBTable(m.TableState.Table)`
- Line 439: `db.StringDBTable(m.Table)` → `db.StringDBTable(m.TableState.Table)`
- Line 446: `FetchTableHeadersRowsCmd(*m.Config, m.Table, nil)` → `m.TableState.Table`

**Verification after Step 3a:**
```
checkfor tool with:
- dir: "/Users/home/Documents/Code/Go_dev/modulacms/internal/cli"
- search: "m.Row"
- ext: ".go"
- whole_word: true
- context: 1
```

#### 3b. Update form.go (~14 changes)
**File:** `internal/cli/form.go`

Pattern: `m.Columns` → `m.TableState.Columns`, `m.ColumnTypes` → `m.TableState.ColumnTypes`, `m.Row` → `m.TableState.Row`, `m.Headers` → `m.TableState.Headers`

Key sections:
- Line 21: `for i, c := range *m.Columns` → `*m.TableState.Columns`
- Line 28: `t := *m.ColumnTypes` → `*m.TableState.ColumnTypes`
- Line 49: `LogMessageCmd(fmt.Sprintf("Headers  %v", m.Columns))` → `m.TableState.Columns`
- Line 50: `FormActionCmd(INSERT, string(table), *m.Columns, values)` → `*m.TableState.Columns`
- Line 56: `FieldsCount: len(*m.Columns)` → `len(*m.TableState.Columns)`
- Line 63: `row := *m.Row` → `*m.TableState.Row`
- Line 65: `for i, c := range *m.Columns` → `*m.TableState.Columns`
- Line 73: `t := *m.ColumnTypes` → `*m.TableState.ColumnTypes`
- Line 101: `FormActionCmd(UPDATE, string(table), m.Headers, m.FormState.FormValues)` → `m.TableState.Headers`
- Line 107: `FieldsCount: len(*m.Columns)` → `len(*m.TableState.Columns)`
- Line 114: `for i, c := range *m.Columns` → `*m.TableState.Columns`
- Line 119: `t := *m.ColumnTypes` → `*m.TableState.ColumnTypes`
- Line 146: `FieldsCount: len(*m.Columns)` → `len(*m.TableState.Columns)`

**Verification after Step 3b:**
```
checkfor tool with:
- dir: "/Users/home/Documents/Code/Go_dev/modulacms/internal/cli"
- search: "m.Columns"
- ext: ".go"
- whole_word: true
- context: 1
```

#### 3c. Update debug.go (~12 changes)
**File:** `internal/cli/debug.go`

All 7 fields referenced here for debug output.

- Line 99: `table := fmt.Sprintf("Table: %s", m.Table)` → `m.TableState.Table`
- Line 122-123: Columns check and length → `m.TableState.Columns`
- Line 132-133: ColumnTypes check and length → `m.TableState.ColumnTypes`
- Line 141: `selected := fmt.Sprintf("Selected(length): %d", len(m.Selected))` → `len(m.TableState.Selected)`
- Line 144: `headers := fmt.Sprintf("Headers(length): %d", len(m.Headers))` → `len(m.TableState.Headers)`
- Line 147: `rows := fmt.Sprintf("Rows(length): %d", len(m.Rows))` → `len(m.TableState.Rows)`
- Line 151-152: Row check and length → `m.TableState.Row`

### Step 4: Update Navigation/Routing Files

#### 4a. Update update_navigation.go (~6 changes)
**File:** `internal/cli/update_navigation.go`

- Line 42: `GetColumnsCmd(*m.Config, m.Table)` → `m.TableState.Table`
- Line 57: `FetchTableHeadersRowsCmd(*m.Config, m.Table, &page)` → `m.TableState.Table`
- Line 69: `FetchTableHeadersRowsCmd(*m.Config, m.Table, &page)` → `m.TableState.Table`
- Line 75: `FetchTableHeadersRowsCmd(*m.Config, m.Table, &page)` → `m.TableState.Table`
- Line 82: `FetchTableHeadersRowsCmd(*m.Config, m.Table, &page)` → `m.TableState.Table`
- Line 161: No change (`m.Tables` is different!)

#### 4b. Update update_fetch.go (~3 changes)
**File:** `internal/cli/update_fetch.go`

- Line 47: `LogMessageCmd(fmt.Sprintf("Table %s headers fetched: %s", m.Table, strings.Join(columns, ", ")))` → `m.TableState.Table`
- Line 51: `for _, v := range m.Headers` → `m.TableState.Headers`
- Line 119: `LogMessageCmd(fmt.Sprintf("Database fetch error for table %s: %s", m.Table, msg.Error.Error()))` → `m.TableState.Table`

#### 4c. Update update_forms.go (~4 changes)
**File:** `internal/cli/update_forms.go`

- Line 26: `if m.Columns == nil` → `if m.TableState.Columns == nil`
- Line 28: `LogMessageCmd(fmt.Sprintf("Form creation failed: no columns available for table %s", m.Table))` → `m.TableState.Table`
- Line 32: `m.NewInsertForm(db.DBTable(m.Table))` → `db.DBTable(m.TableState.Table)`
- Line 35: `LogMessageCmd(fmt.Sprintf("Database create form initialized for table %s with %d fields", m.Table, len(*m.Columns)-1))` → `m.TableState.Table` and `len(*m.TableState.Columns)`

### Step 5: Update View/Display Files

#### 5a. Update view.go (~10 changes)
**File:** `internal/cli/view.go`

- Line 119: `p.AddHeader(fmt.Sprintf("Read %s", m.Table))` → `m.TableState.Table`
- Line 121: `p.AddHeaders(m.Headers)` → `m.TableState.Headers`
- Line 122: `p.AddRows(m.Rows)` → `m.TableState.Rows`
- Line 128: `columns := make([]ReadSingleRow, 0, len(m.Headers))` → `len(m.TableState.Headers)`
- Line 129: `for i, v := range m.Headers` → `m.TableState.Headers`
- Line 138: `columns[i].Value = m.Rows[m.Cursor][i]` → `m.TableState.Rows`
- Line 148: `p.AddHeader(fmt.Sprintf("Read %s Row %d", m.Table, m.Cursor))` → `m.TableState.Table`
- Line 155: `p.AddHeaders(m.Headers)` → `m.TableState.Headers`
- Line 156: `p.AddRows(m.Rows)` → `m.TableState.Rows`
- Line 171: `p.AddHeaders(m.Headers)` → `m.TableState.Headers`
- Line 172: `p.AddRows(m.Rows)` → `m.TableState.Rows`

#### 5b. Update style.go (~2 changes)
**File:** `internal/cli/style.go`

- Line 168-169: Table name display → `m.TableState.Table`

### Step 6: Update Utility/Helper Files

#### 6a. Update commands.go (~2 changes)
**File:** `internal/cli/commands.go`

- Line 103: `valuesMap[m.Headers[i]] = *v` → `m.TableState.Headers`
- Line 230: `row := m.Rows[m.Cursor]` → `m.TableState.Rows`

#### 6b. Update constructors.go (~2 changes)
**File:** `internal/cli/constructors.go`

- Line 347: `start, end := m.Paginator.GetSliceBounds(len(m.Rows))` → `len(m.TableState.Rows)`
- Line 348: `currentView := m.Rows[start:end]` → `m.TableState.Rows`

#### 6c. Update handles.go (~1 change)
**File:** `internal/cli/handles.go`

- Line 10: `rows := m.Rows` → `m.TableState.Rows`

#### 6d. Update fields.go (~1 change)
**File:** `internal/cli/fields.go`

- Line 115: `r, err := db.GetColumnRowsString(con, ctx, m.Table, column)` → `m.TableState.Table`

#### 6e. Update status.go (~1 change)
**File:** `internal/cli/status.go`

- Line 31: `table := fmt.Sprintf("Table\n%s\n", m.Table)` → `m.TableState.Table`

#### 6f. Update update_dialog.go (~1 change)
**File:** `internal/cli/update_dialog.go`

- Line 38: `DatabaseDeleteEntryCmd(int(id), m.Table)` → `m.TableState.Table`

---

## Files Changed Summary

| File | Lines Changed | Type | Priority | Field References |
|------|---------------|------|----------|------------------|
| table_model.go | +40 | NEW | 1 | - |
| model.go | ~20 | MODIFY | 1 | All 7 |
| update_controls.go | ~17 | MODIFY | 2 | Table, Rows, Row |
| form.go | ~14 | MODIFY | 2 | Columns, ColumnTypes, Row, Headers |
| debug.go | ~12 | MODIFY | 2 | All 7 |
| view.go | ~10 | MODIFY | 3 | Table, Headers, Rows |
| update_navigation.go | ~6 | MODIFY | 3 | Table |
| update_forms.go | ~4 | MODIFY | 3 | Table, Columns |
| update_fetch.go | ~3 | MODIFY | 3 | Table, Headers |
| commands.go | ~2 | MODIFY | 4 | Headers, Rows |
| constructors.go | ~2 | MODIFY | 4 | Rows |
| style.go | ~2 | MODIFY | 4 | Table |
| handles.go | ~1 | MODIFY | 4 | Rows |
| fields.go | ~1 | MODIFY | 4 | Table |
| status.go | ~1 | MODIFY | 4 | Table |
| update_dialog.go | ~1 | MODIFY | 4 | Table |

**Total:** 16 files, ~137 line changes + 40 new lines = **~177 lines**

---

## Quick Reference: checkfor Commands for Each Field

| Field | Search String | Expected Matches | Priority Files |
|-------|--------------|------------------|----------------|
| Table | `"m.Table"` | 32 | update_controls, form, update_navigation |
| Headers | `"m.Headers"` | 9 | view, form, commands |
| Rows | `"m.Rows"` | 15 | view, update_controls, constructors |
| Columns | `"m.Columns"` | 12 | form, update_forms |
| ColumnTypes | `"m.ColumnTypes"` | 5 | form |
| Selected | `"m.Selected"` | 1 | debug |
| Row | `"m.Row"` | 4 | form, update_controls |
| TableState | `"m.TableState"` | ~78 (after) | All files |

**Standard checkfor parameters:**
```
dir: "/Users/home/Documents/Code/Go_dev/modulacms/internal/cli"
ext: ".go"
whole_word: true
context: 1
```

---

## Migration Strategy: Two-Pass Approach

### Pass 1: Add TableState with Old Fields
1. Create `table_model.go` with TableModel struct
2. Add `TableState *TableModel` to Model (keep old fields temporarily)
3. Update InitialModel to initialize TableState
4. Update 3 highest-usage files (update_controls.go, form.go, debug.go)
5. Build and verify compilation
6. **Validation:** Verify 3 files use TableState with checkfor

### Pass 2: Complete Migration and Remove Old Fields
1. Update remaining 13 files
2. Remove old Table* field declarations from Model
3. Build and verify all field references resolved
4. Run checkfor validation for all 7 fields
5. **Validation:** Complete manual testing checklist

---

## Validation Approach

### Compiler Checks
```bash
# After each file change
go build ./internal/cli/

# Should succeed or show clear missing reference errors
```

### checkfor Verification Phases

**Phase 1: Pre-Refactor Inventory**
```
# Run for each of 7 fields to establish baseline
checkfor tool with:
- dir: "/Users/home/Documents/Code/Go_dev/modulacms/internal/cli"
- search: "m.Table"
- ext: ".go"
- whole_word: false
- context: 0
```

**Phase 2: After Each Major File**
```
# After update_controls.go
checkfor: search "m.Table" - should show fewer matches

# After form.go
checkfor: search "m.Columns" - should show fewer matches
```

**Phase 3: Final Validation (After Pass 2)**
```
# Verify no old direct references remain (run for each field)
checkfor tool with:
- search: "m.Table"  # Expected: 0 matches
- search: "m.Headers"  # Expected: 0 matches
- search: "m.Rows"  # Expected: 0 matches
- search: "m.Columns"  # Expected: 0 matches
- search: "m.ColumnTypes"  # Expected: 0 matches
- search: "m.Selected"  # Expected: 0 matches
- search: "m.Row"  # Expected: 0 matches
```

**Phase 4: Positive Verification**
```
checkfor tool with:
- search: "m.TableState"
- expected: ~78 references across 16 files
```

### Manual Testing Checklist

1. **Start CLI:** `./modulacms-x86 --cli`
2. **Test Database Table Selection:**
   - Navigate to Database menu
   - Select a table
   - Verify table name displays correctly
3. **Test Table Viewing:**
   - View table data (Read operation)
   - Verify headers display
   - Verify rows display
   - Verify pagination works
4. **Test Row Selection:**
   - Navigate through rows with arrow keys
   - Verify cursor moves correctly
   - Select a row (press Enter or 'u')
   - Verify row data populates correctly
5. **Test Form Creation from Table:**
   - Press 'c' to create new entry
   - Verify form fields match table columns
   - Verify column types determine input types
6. **Test Form Update from Table:**
   - Select row and press 'u' for update
   - Verify form pre-fills with row data
   - Verify headers used for field labels
7. **Test Table Navigation:**
   - Navigate between different tables
   - Verify state clears properly
   - Verify new table loads correctly
8. **Check debug.log:**
   - `tail -f debug.log` during testing
   - Look for panics or nil pointer errors
   - Verify table state logging works

---

## Risks and Mitigations

### Risk 1: Nil Pointer Dereference
**Probability:** Low-Medium
**Impact:** Runtime panic
**Mitigation:**
- Initialize TableState in InitialModel (like FormState)
- Add nil checks for pointer fields (Columns, ColumnTypes, Row)
- Follow established FormModel pattern

### Risk 2: Missed Field Reference
**Probability:** Medium (78 references across 16 files!)
**Impact:** Compile error (good - caught early)
**Mitigation:**
- Use checkfor to track each field systematically
- Build after each file change
- Work through files by priority (highest usage first)
- Go compiler will catch all missed references

### Risk 3: Tables vs Table Confusion
**Probability:** Medium
**Impact:** Wrong field updated
**Mitigation:**
- `m.Tables` (plural) is a DIFFERENT field - list of all tables
- `m.Table` (singular) is what we're extracting - current table name
- Use whole_word matching in checkfor
- Careful code review in update_navigation.go

### Risk 4: Pointer Field Nil Checks
**Probability:** Medium
**Impact:** Panic when accessing *Columns, *ColumnTypes, *Row
**Mitigation:**
- Preserve existing nil checks in form.go, debug.go
- Test form creation path thoroughly
- Verify update operations handle nil properly

---

## Success Criteria

### Immediate (Phase 2 Complete)
- [ ] Model struct reduced from 46 to 39 fields
- [ ] All table operations working (view, select, navigate, form creation)
- [ ] No runtime panics or errors in debug.log
- [ ] Code compiles without warnings
- [ ] Manual testing checklist passes 100%
- [ ] checkfor shows 0 old references, ~78 TableState references

### Medium-Term (Enables Future Work)
- [ ] CMS UI fully refactored (forms + tables)
- [ ] Table operations more maintainable
- [ ] Pattern reinforced for Phase 3 (NavigationModel)
- [ ] Adding table features (sort, filter) now easier

---

## Rollback Plan

If something goes wrong:

1. **Git checkpoint before starting:** `git commit -m "Checkpoint before TableModel extraction"`
2. **After Pass 1:** `git commit -m "Pass 1: TableState added alongside old fields"`
3. **If Pass 2 fails:** `git reset --hard` to Pass 1 checkpoint
4. **Complete rollback:** `git reset --hard` to initial checkpoint

Keep changes in small, atomic commits for each file or small group of files.

---

## Pattern Consistency with Phase 1

Following FormModel extraction pattern:
- Pointer field in Model: `TableState *TableModel`
- Constructor function: `NewTableModel()`
- Initialize in InitialModel
- Two-pass migration (add alongside, then remove old)
- checkfor verification at each step
- Manual testing checklist

**Differences from Phase 1:**
- More fields (7 vs 8) but lower total complexity
- More files (16 vs 10) but simpler changes per file
- Some pointer fields need nil checks (Columns, ColumnTypes, Row)
- Need to watch for Tables vs Table confusion

---

## Key Insight: Completing CMS UI Refactor

Phase 1 (FormModel) + Phase 2 (TableModel) = **Complete CMS UI Refactoring**

These two phases extract all form and table state from Model, completing the primary pain point:
- Forms: 8 fields → FormModel (51 references)
- Tables: 7 fields → TableModel (78 references)
- **Total:** 15 fields extracted, 129 references refactored

After Phase 2:
- Model reduced from 54 → 39 fields (28% reduction)
- Clear separation: FormState handles forms, TableState handles tables
- Future table features (sorting, filtering, multi-select) isolated to TableModel
- Pattern proven for Phases 3-4 (NavigationModel, DisplayModel)

**This is the high-value work. After Phase 2, the hard part is done.**

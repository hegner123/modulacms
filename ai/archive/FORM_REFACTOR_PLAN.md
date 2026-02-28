# CLI Model Refactoring Plan

**Goal:** Extract FormModel from the 54-field Model struct to improve CMS UI development workflow
**Approach:** Conservative, incremental extraction following existing DialogModel pattern
**Risk Level:** Low - 160 lines changed across 8 files, following proven patterns
**Timeline:** 2 days (1 day per pass)

---

## Problem Statement

The Model struct in `internal/cli/model.go` has 54 fields mixing concerns:
- Configuration, UI dimensions, navigation, pagination, tables, forms, dialogs, CMS-specific state
- `update_state.go` handles 25+ disparate field updates (188 lines)
- CMS UI flow with forms, tables, and dialogs is difficult to manage
- Adding new UI elements means bloating Model further

**Pain Point:** Working with forms in CMS requires touching 8 scattered fields across multiple files.

---

## Phase 1: Extract FormModel (This Plan)

### What We're Extracting

**8 Form Fields → FormModel struct:**
```go
type FormModel struct {
    Form        *huh.Form
    FormLen     int
    FormMap     []string
    FormValues  []*string
    FormSubmit  bool
    FormGroups  []huh.Group
    FormFields  []huh.Field
    FormOptions *FormOptionsMap
}
```

**Why These Fields:**
- Always used together in form operations
- Isolated to 5 files (form.go, update_forms.go, fields.go, update_controls.go, constructors.go)
- Clear semantic boundary (form lifecycle)
- Follows DialogModel pattern already in codebase

### Implementation Steps

#### Step 1: Create FormModel Struct
**File:** `internal/cli/form_model.go` (NEW, ~50 lines)

```go
package cli

import "github.com/charmbracelet/huh"

type FormModel struct {
    Form        *huh.Form
    FormLen     int
    FormMap     []string
    FormValues  []*string
    FormSubmit  bool
    FormGroups  []huh.Group
    FormFields  []huh.Field
    FormOptions *FormOptionsMap
}

func NewFormModel() *FormModel {
    return &FormModel{
        FormMap:    make([]string, 0),
        FormSubmit: false,
    }
}
```

#### Step 2: Update Model Struct
**File:** `internal/cli/model.go`

**Remove 8 fields (lines 84-91):**
```go
// DELETE:
// Form         *huh.Form
// FormLen      int
// FormMap      []string
// FormValues   []*string
// FormSubmit   bool
// FormGroups   []huh.Group
// FormFields   []huh.Field
// FormOptions  *FormOptionsMap
```

**Add 1 field (after line 83):**
```go
// ADD:
FormState    *FormModel
```

**Update InitialModel (line ~162):**
```go
m := Model{
    // ... existing fields ...
    FormState: NewFormModel(),  // ADD
    // ... rest ...
}
```

**Verification after Step 2:**
```
# Use checkfor to see how many references need updating in next steps
checkfor tool with:
- dir: "/Users/home/Documents/Code/Go_dev/modulacms/internal/cli"
- search: "m.Form"
- ext: ".go"
- case_insensitive: false
- whole_word: false
- context: 0

# Count total occurrences to track progress
# Expected: ~60 references across 6 files
```

#### Step 3: Update form.go (Core Form Logic)
**File:** `internal/cli/form.go` (~20 changes)

**Pattern:** Replace `m.Form*` with `m.FormState.Form*`

**Example changes:**
```go
// Line ~28: BEFORE
m.FormValues = append(m.FormValues, &value)

// Line ~28: AFTER
m.FormState.FormValues = append(m.FormState.FormValues, &value)

// Line ~73: BEFORE
m.Form = &form

// Line ~73: AFTER
m.FormState.Form = &form

// Line ~88: BEFORE
m.FormGroups = append(m.FormGroups, *group)

// Line ~88: AFTER
m.FormState.FormGroups = append(m.FormState.FormGroups, *group)
```

**Search pattern:** `m\.Form([^S])` → `m.FormState.Form$1`

**Verification after Step 3:**
```
# Verify form.go no longer has direct Form field access
checkfor tool with:
- dir: "/Users/home/Documents/Code/Go_dev/modulacms/internal/cli"
- search: "m.Form"
- ext: ".go"
- context: 1

# Check form.go results - should only show "m.FormState.Form" patterns
```

#### Step 4: Update update_state.go (State Handlers)
**File:** `internal/cli/update_state.go` (~8 changes)

Update message handlers for Form* messages:

```go
// FormLenSet handler
case FormLenSet:
    newModel := m
    newModel.FormState.FormLen = msg.FormLen  // CHANGE
    return newModel, NewStateUpdate()

// FormSet handler
case FormSet:
    newModel := m
    newModel.FormState.Form = &msg.Form       // CHANGE
    newModel.FormState.FormValues = msg.Values // CHANGE
    return newModel, NewStateUpdate()

// FormValuesSet handler
case FormValuesSet:
    newModel := m
    newModel.FormState.FormValues = msg.Values // CHANGE
    return newModel, NewStateUpdate()

// FormOptionsSet handler
case FormOptionsSet:
    newModel := m
    if newModel.FormState.FormOptions == nil {  // CHANGE
        newModel.FormState.FormOptions = &FormOptionsMap{} // CHANGE
    }
    (*newModel.FormState.FormOptions)[msg.FieldName] = msg.Options // CHANGE
    return newModel, NewStateUpdate()
```

#### Step 5: Update update_forms.go (Form Handling)
**File:** `internal/cli/update_forms.go` (~8 changes)

```go
// Line ~40: BEFORE
if m.FormOptions == nil {

// Line ~40: AFTER
if m.FormState.FormOptions == nil {

// Line ~78: BEFORE
m.FormSubmit = true

// Line ~78: AFTER
m.FormState.FormSubmit = true

// Line ~85: BEFORE
opts := (*m.FormOptions)[key]

// Line ~85: AFTER
opts := (*m.FormState.FormOptions)[key]
```

#### Step 6: Update fields.go (Field Generation)
**File:** `internal/cli/fields.go` (~6 changes)

```go
// NewFieldFromType function: BEFORE
m.FormValues = append(m.FormValues, &value)

// NewFieldFromType function: AFTER
m.FormState.FormValues = append(m.FormState.FormValues, &value)

// Similar changes in NewUpdateFieldFromType
```

#### Step 7: Update update_controls.go (Form Controls)
**File:** `internal/cli/update_controls.go` (~15 changes)

Most changes in `FormControls` function (lines 265-289):

```go
// Line ~268: BEFORE
form, cmd := m.Form.Update(msg)

// Line ~268: AFTER
form, cmd := m.FormState.Form.Update(msg)

// Line ~269: BEFORE
m.Form = form.(*huh.Form)

// Line ~269: AFTER
m.FormState.Form = form.(*huh.Form)

// Line ~272: BEFORE
if m.Form.State == huh.StateCompleted {

// Line ~272: AFTER
if m.FormState.Form.State == huh.StateCompleted {
```

**Verification after Step 7:**
```
# Check that all major files now use FormState
# Run checkfor for each of the 8 fields to find any stragglers
checkfor tool with:
- dir: "/Users/home/Documents/Code/Go_dev/modulacms/internal/cli"
- search: "m.FormOptions"
- ext: ".go"
- whole_word: true
- context: 1

# If results show "m.FormOptions" (not "m.FormState.FormOptions"), those need fixing
# Repeat for: FormLen, FormMap, FormValues, FormSubmit, FormGroups, FormFields
```

#### Step 8: Update constructors.go (Optional)
**File:** `internal/cli/constructors.go` (~5 changes, optional)

Constructors like `SetFormDataCmd` can remain unchanged - they return messages that update_state.go handles. No changes strictly required unless you want type safety.

---

## Quick Reference: checkfor Commands for Each Field

Use these `checkfor` commands to verify each field during migration:

| Field | Search String | Use Case |
|-------|--------------|----------|
| Form | `"m.Form"` | Most common, check after each step |
| FormLen | `"m.FormLen"` | Check after update_state.go changes |
| FormMap | `"m.FormMap"` | Check after form.go changes |
| FormValues | `"m.FormValues"` | Check after form.go and fields.go |
| FormSubmit | `"m.FormSubmit"` | Check after update_forms.go |
| FormGroups | `"m.FormGroups"` | Check after form.go changes |
| FormFields | `"m.FormFields"` | Check after form.go changes |
| FormOptions | `"m.FormOptions"` | Check after update_forms.go |
| FormState | `"m.FormState"` | Final verification - should have ~60+ hits |

**Standard checkfor parameters:**
```
dir: "/Users/home/Documents/Code/Go_dev/modulacms/internal/cli"
ext: ".go"
whole_word: true
context: 1
```

---

## Migration Strategy: Two-Pass Approach

### Pass 1: Add FormState with Old Fields (Day 1)
1. Create `form_model.go` with FormModel struct
2. Add `FormState *FormModel` to Model (keep old fields temporarily)
3. Update InitialModel to initialize FormState
4. Update 2 files to use FormState (form.go, update_state.go)
5. Build and test manually
6. **Validation:** Run CLI, create a form, submit form, verify no crashes

### Pass 2: Remove Old Fields (Day 2)
1. Update remaining files (update_forms.go, fields.go, update_controls.go)
2. Remove old Form* field declarations from Model
3. Build and verify all field references resolved
4. Full manual testing
5. **Validation:** Complete testing checklist (below)

---

## Validation Approach (No Tests)

### Compiler Checks
```bash
# After each file change
go build ./internal/cli/
```

### Verification with checkfor Tool

Use the MCP `checkfor` tool for token-efficient, single-directory verification:

**Before starting - inventory all references:**
```
checkfor tool with:
- dir: "/Users/home/Documents/Code/Go_dev/modulacms/internal/cli"
- search: "m.Form"
- ext: ".go"
- context: 1
```

**After updating each file - verify specific field:**
```
checkfor tool with:
- dir: "/Users/home/Documents/Code/Go_dev/modulacms/internal/cli"
- search: "m.FormLen"
- ext: ".go"
- whole_word: true
- context: 0
```

**Final validation - confirm no old references remain:**
```
# Check each field individually for precise results
1. search: "m.Form" (then filter for FormState in results)
2. search: "m.FormLen"
3. search: "m.FormValues"
4. search: "m.FormSubmit"
5. search: "m.FormMap"
6. search: "m.FormGroups"
7. search: "m.FormFields"
8. search: "m.FormOptions"
```

### Manual Testing Checklist
1. **Start CLI:** `./modulacms-x86 --cli`
2. **Test Database Form Creation:**
   - Navigate to Database → select table → press 'c' (create)
   - Verify form appears with correct fields
3. **Test Form Submission:**
   - Fill form fields
   - Submit with Enter
   - Verify record created in table
4. **Test Form Pre-population:**
   - Navigate to Database → select row → press 'u' (update)
   - Verify form pre-filled with existing values
5. **Test CMS Forms:**
   - Navigate to CMS → Datatypes → press 'c' (create datatype)
   - Fill form, submit
   - Verify datatype created
6. **Test Form Abort:**
   - Open any form
   - Press ESC
   - Verify returns to previous page without error
7. **Test Form Options (Select Fields):**
   - Create form with select field
   - Verify dropdown shows options
8. **Check debug.log:**
   - `tail -f debug.log` during testing
   - Look for panics or errors

---

## Files Changed Summary

| File | Lines Changed | Type | Priority |
|------|---------------|------|----------|
| form_model.go | +50 | NEW | 1 |
| model.go | ~20 | MODIFY | 1 |
| form.go | ~20 | MODIFY | 2 |
| update_state.go | ~8 | MODIFY | 2 |
| update_forms.go | ~8 | MODIFY | 3 |
| fields.go | ~6 | MODIFY | 3 |
| update_controls.go | ~15 | MODIFY | 3 |
| constructors.go | ~5 | MODIFY (optional) | 4 |

**Total:** 8 files, ~132 line changes + 50 new lines = **~180 lines**

---

## Risks and Mitigations

### Risk 1: Nil Pointer Dereference
**Probability:** Low
**Impact:** Runtime panic
**Mitigation:**
- Initialize FormState in InitialModel (like Dialog)
- Add nil checks where needed: `if m.FormState != nil`
- DialogModel already handles this pattern successfully

### Risk 2: Missed Field Reference
**Probability:** Medium
**Impact:** Compile error (good - caught early)
**Mitigation:**
- Use grep to find all references before starting
- Build after each file change
- Go compiler will catch all missed references

### Risk 3: Form State Reset Issues
**Probability:** Low
**Impact:** Form state leaks between pages
**Mitigation:**
- Review form reset logic in constructors.go
- Test form creation multiple times in session
- Verify FormState is recreated/cleared properly

### Risk 4: FormOptions Nil Map
**Probability:** Medium
**Impact:** Panic on map access
**Mitigation:**
- FormOptions already has nil checks in update_forms.go
- Preserve existing nil-check pattern
- Test select fields specifically

---

## Success Criteria

### Immediate (Phase 1 Complete)
- [ ] Model struct reduced from 54 to 46 fields
- [ ] All form operations working (create, update, submit, abort)
- [ ] No runtime panics or errors in debug.log
- [ ] Code compiles without warnings
- [ ] Manual testing checklist passes 100%

### Medium-Term (Enables Future Work)
- [ ] CMS UI development is easier (less field hunting)
- [ ] FormModel can be extended with methods (validation, reset)
- [ ] Pattern established for TableModel extraction (Phase 2)
- [ ] Confidence built for remaining refactoring phases

---

## Future Phases (Brief Outline)

### Phase 2: Extract TableModel
- **Fields:** Headers, Rows, Columns, ColumnTypes, Selected, Row, Table (7 fields)
- **Effort:** ~200 lines across 10 files
- **Risk:** Medium (more widely used than forms)
- **Value:** Completes CMS UI pain point (tables + forms)

### Phase 3: Extract NavigationModel
- **Fields:** Cursor, CursorMax, PageMenu, History, Page, Focus, FocusIndex (8 fields)
- **Effort:** ~250 lines across 12+ files
- **Risk:** Higher (navigation is central)
- **Value:** Cleaner page routing and menu logic

### Phase 4: Extract DisplayModel
- **Fields:** Width, Height, TitleFont, Titles, Spinner, Viewport, Loading (7 fields)
- **Effort:** ~150 lines across 6 files
- **Risk:** Low (rendering only)
- **Value:** Nice-to-have, not critical

**End State After All Phases:**
- Model: ~20 fields (Config, Status, Err, + 4-5 sub-models)
- 4 cohesive sub-models: FormModel, TableModel, NavigationModel, DisplayModel
- update_state.go reduced from 188 lines to ~80 lines
- Clear separation of concerns

---

## Alternative Approach: More Aggressive

If Phase 1 goes smoothly and you want faster progress:

**Extract FormModel + TableModel together:**
- 12 fields total (Form 8 + Table 4)
- ~220 line change
- Higher risk but bigger immediate payoff
- Only recommended if you have time for thorough testing

**Recommendation:** Stick with FormModel-only first. Get a win, build confidence, then tackle tables.

---

## Implementation Notes

### Pattern to Follow: DialogModel
DialogModel is already successfully extracted in this codebase. Study these patterns:

**dialog.go:**
- Standalone struct with constructor
- Implements Update() and Render() methods
- Managed as pointer field in Model

**Usage in update.go (lines 40-48):**
```go
if m.DialogActive && m.Dialog != nil {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        dialog, cmd := m.Dialog.Update(msg)
        m.Dialog = &dialog
        // ...
    }
}
```

**FormModel should follow this pattern:**
- Pointer field in Model
- Nil checks where needed
- Could have Update() method in future (not Phase 1)

### checkfor Tool Commands for Verification

Use the MCP `checkfor` tool instead of grep for token-efficient verification:

**Phase 1: Pre-Refactor Inventory (Before Step 1)**
```
# Get complete inventory of all 8 fields before starting
checkfor tool with:
- dir: "/Users/home/Documents/Code/Go_dev/modulacms/internal/cli"
- search: "m.Form"
- ext: ".go"
- context: 2

# Repeat for each field: FormLen, FormValues, FormSubmit, FormMap, FormGroups, FormFields, FormOptions
```

**Phase 2: During Refactor (After Each File Update)**
```
# Verify specific file changes (example: after updating form.go)
checkfor tool with:
- dir: "/Users/home/Documents/Code/Go_dev/modulacms/internal/cli"
- search: "m.FormValues"
- ext: ".go"
- whole_word: true
- context: 1

# Should show only "m.FormState.FormValues" references, not "m.FormValues"
```

**Phase 3: Post-Refactor Validation (After Pass 2)**
```
# Verify no old direct references remain (run for each field)
checkfor tool with:
- dir: "/Users/home/Documents/Code/Go_dev/modulacms/internal/cli"
- search: "m.FormLen"
- ext: ".go"
- whole_word: true
- context: 0

# Expected result: 0 matches (or only in comments)
# Repeat for: FormValues, FormSubmit, FormMap, FormGroups, FormFields, FormOptions
```

**Phase 4: Verify FormState Usage**
```
# Confirm all fields now accessed through FormState
checkfor tool with:
- dir: "/Users/home/Documents/Code/Go_dev/modulacms/internal/cli"
- search: "m.FormState"
- ext: ".go"
- context: 0

# Should show ~60+ references across files
```

**Advantages over grep:**
- Single directory search (perfect for internal/cli/)
- Extension filtering (.go only)
- Whole-word matching (avoid false positives)
- Context lines for understanding
- Token-efficient JSON output
- No need for complex regex patterns

---

## Rollback Plan

If something goes wrong:

1. **Git checkpoint before starting:** `git commit -m "Checkpoint before FormModel extraction"`
2. **After Pass 1:** `git commit -m "Pass 1: FormState added alongside old fields"`
3. **If Pass 2 fails:** `git reset --hard` to Pass 1 checkpoint
4. **Complete rollback:** `git reset --hard` to initial checkpoint

Keep changes in small, atomic commits so you can rollback to any safe point.

---

## Key Insight: 80/20 Refactoring

This plan extracts **15% of Model fields (8 of 54)** but addresses **80% of your CMS UI pain**:
- Forms are the main UI element you're working with
- Form state is well-isolated and safe to extract
- Success here enables TableModel next (forms + tables = full CMS UI)
- Proves the pattern for remaining 3 phases

**Conservative approach, maximum value.**

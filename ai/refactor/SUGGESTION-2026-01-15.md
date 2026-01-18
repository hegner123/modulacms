# SUGGESTION: Hybrid Command Approach for Coupled Operations

**Date:** 2026-01-15
**Author:** AI Analysis
**Status:** Proposed - Not Yet Implemented
**Complexity:** Medium
**Estimated Effort:** 2-3 days

---

## Problem Being Solved

See [PROBLEM.md](PROBLEM.md) for full context.

**TL;DR:** CMS content creation requires chaining database operations (ContentData + ContentFields), but current architecture can't pass results between operations.

---

## Core Insight

**The CLI Commands layer bypasses DbDriver's typed methods:**

```go
// Current pattern (commands.go)
func (m Model) DatabaseInsert(...) tea.Cmd {
    con, _, err := d.GetConnection()         // ← Bypasses typed methods
    sqb := db.NewSecureQueryBuilder(con)     // ← Uses generic builder
    res, err := sqb.SecureExecuteModifyQuery(...)
    return DbResultCmd(res, string(table))   // ← Loses type info
}
```

**DbDriver HAS typed methods that return structs with IDs:**

```go
// db.go:114 (Existing but unused by CLI)
CreateContentData(CreateContentDataParams) ContentData  // ← Returns struct with ID!
CreateContentField(CreateContentFieldParams) ContentFields
```

**Root cause:** CLI layer chose flexibility (generic query builder) over type safety (DbDriver methods).

---

## Solution: Hybrid Approach

**Use BOTH patterns:**

1. **Generic `DatabaseInsert()`** - For simple, single-record operations
2. **Specialized commands** - For coupled operations that need type safety

This is actually **good architecture** - different tools for different jobs.

---

## Implementation Design

### Step 1: New Command for Coupled Operation

**File:** `internal/cli/commands.go`

**Add new method:**

```go
// CreateContentWithFields performs atomic content creation
// Uses DbDriver typed methods instead of generic query builder
func (m Model) CreateContentWithFields(
    c *config.Config,
    datatypeID int64,
    routeID int64,
    authorID int64,
    fieldValues map[int64]string, // map[field_id]field_value
) tea.Cmd {
    return func() tea.Msg {
        d := db.ConfigDB(*c)

        // Step 1: Create ContentData using typed DbDriver method
        contentData := d.CreateContentData(db.CreateContentDataParams{
            DatatypeID:   &datatypeID,
            RouteID:      &routeID,
            AuthorID:     &authorID,
            DateCreated:  time.Now().Unix(),
            DateModified: time.Now().Unix(),
            // Tree pointers null for now (can update later)
            ParentID:       nil,
            FirstChildID:   nil,
            NextSiblingID:  nil,
            PrevSiblingID:  nil,
        })

        // Check if creation succeeded
        if contentData.ContentDataID == 0 {
            return DbErrMsg{
                Error: fmt.Errorf("failed to create content data"),
            }
        }

        // Step 2: Create ContentFields (we have the ID now!)
        var failedFields []int64
        createdFields := 0

        for fieldID, value := range fieldValues {
            // Skip empty values
            if value == "" {
                continue
            }

            fieldResult := d.CreateContentField(db.CreateContentFieldParams{
                ContentDataID: &contentData.ContentDataID,
                FieldID:       &fieldID,
                FieldValue:    &value,
                RouteID:       &routeID,
                AuthorID:      &authorID,
                DateCreated:   time.Now().Unix(),
                DateModified:  time.Now().Unix(),
            })

            // Track failures
            if fieldResult.ContentFieldID == 0 {
                failedFields = append(failedFields, fieldID)
            } else {
                createdFields++
            }
        }

        // Step 3: Return appropriate message based on results
        if len(failedFields) > 0 {
            return ContentCreatedWithErrorsMsg{
                ContentDataID: contentData.ContentDataID,
                RouteID:       routeID,
                CreatedFields: createdFields,
                FailedFields:  failedFields,
            }
        }

        return ContentCreatedMsg{
            ContentDataID: contentData.ContentDataID,
            RouteID:       routeID,
            FieldCount:    createdFields,
        }
    }
}
```

**Why this works:**
- ✅ Uses existing DbDriver typed methods (no new infrastructure)
- ✅ Synchronous within the Cmd (operations execute in order)
- ✅ Returns structured result (ContentCreatedMsg with all data)
- ✅ Fits message-driven architecture (returns tea.Msg)
- ✅ Handles partial failures gracefully
- ✅ Works across all 3 database drivers (polymorphism via DbDriver interface)

### Step 2: New Message Types

**File:** `internal/cli/message_types.go`

**Add new messages:**

```go
// CMS content creation messages

type CreateContentWithFieldsMsg struct {
    DatatypeID  int64
    RouteID     int64
    FieldValues map[int64]string  // map[field_id]field_value
}

type ContentCreatedMsg struct {
    ContentDataID int64
    RouteID       int64
    FieldCount    int
}

type ContentCreatedWithErrorsMsg struct {
    ContentDataID int64
    RouteID       int64
    CreatedFields int
    FailedFields  []int64
}
```

### Step 3: Message Constructor

**File:** `internal/cli/constructors.go`

**Add constructor:**

```go
func CreateContentWithFieldsCmd(
    config *config.Config,
    datatypeID int64,
    routeID int64,
    authorID int64,
    fieldValues map[int64]string,
) tea.Cmd {
    return func() tea.Msg {
        m := Model{Config: config}
        return m.CreateContentWithFields(config, datatypeID, routeID, authorID, fieldValues)()
    }
}
```

### Step 4: Update Handler

**File:** `internal/cli/update_cms.go`

**Modify existing handler:**

```go
func (m Model) UpdateCms(msg tea.Msg) (Model, tea.Cmd) {
    cmds := make([]tea.Cmd, 0)

    switch msg := msg.(type) {

    case CmsAddNewContentDataMsg:
        // Collect field values from form state
        fieldValues := m.CollectFieldValuesFromForm()

        // Dispatch specialized command
        return m, CreateContentWithFieldsCmd(
            m.Config,
            msg.Datatype,
            m.PageRouteId,
            m.Config.User.ID,
            fieldValues,
        )

    case ContentCreatedMsg:
        // Success path
        return m, tea.Batch(
            ShowSuccessDialogCmd(
                fmt.Sprintf("✓ Created content with %d fields", msg.FieldCount),
            ),
            NavigateToContentListCmd(msg.RouteID),
            RefreshContentListCmd(msg.RouteID),
            CursorResetCmd(),
        )

    case ContentCreatedWithErrorsMsg:
        // Partial success path
        return m, tea.Batch(
            ShowWarningDialogCmd(
                fmt.Sprintf("⚠ Content created but %d/%d fields failed",
                    len(msg.FailedFields),
                    msg.CreatedFields + len(msg.FailedFields),
                ),
            ),
            LogMessageCmd(fmt.Sprintf("Failed field IDs: %v", msg.FailedFields)),
            NavigateToContentListCmd(msg.RouteID),
            RefreshContentListCmd(msg.RouteID),
        )

    // ... existing cases
    }

    return m, tea.Batch(cmds...)
}
```

### Step 5: Helper Method

**File:** `internal/cli/model.go` or new file `internal/cli/cms_helpers.go`

**Add helper:**

```go
// CollectFieldValuesFromForm extracts field values from form state
// Returns map[field_id]field_value
func (m Model) CollectFieldValuesFromForm() map[int64]string {
    fieldValues := make(map[int64]string)

    // Assuming FormState.FormValues is []*string and FormState.FormMap is []string
    // where FormMap contains field_id as string
    for i, value := range m.FormState.FormValues {
        if value == nil || *value == "" {
            continue
        }

        // Parse field ID from FormMap
        if i < len(m.FormState.FormMap) {
            fieldIDStr := m.FormState.FormMap[i]
            fieldID, err := strconv.ParseInt(fieldIDStr, 10, 64)
            if err != nil {
                utility.DefaultLogger.Ferror("Failed to parse field ID", err)
                continue
            }

            fieldValues[fieldID] = *value
        }
    }

    return fieldValues
}
```

---

## Why This Approach Works

### 1. Respects Message-Driven Architecture

```
User submits form
  ↓
CmsAddNewContentDataMsg dispatched
  ↓
UpdateCms handler collects field values
  ↓
CreateContentWithFieldsCmd sent (async)
  ↓
Command executes (synchronous WITHIN the Cmd)
  ├─ Insert ContentData → get ID
  ├─ Insert ContentField 1
  ├─ Insert ContentField 2
  └─ Insert ContentField N
  ↓
Returns ContentCreatedMsg
  ↓
UpdateCms handler receives success
  ↓
Dispatch UI update commands
```

**Key insight:** Async at message boundaries, synchronous within the command.

### 2. Uses Existing Infrastructure

- ✅ DbDriver interface methods (already implemented)
- ✅ Message-driven patterns (already established)
- ✅ No new frameworks or abstractions
- ✅ Fits existing codebase patterns

### 3. Type-Safe Operation Chaining

```go
// Generic approach (current) - loses types
res, err := sqb.SecureExecuteModifyQuery(query, args)
// res is sql.Result - can't easily access ID

// Typed approach (proposed) - preserves types
contentData := d.CreateContentData(params)
// contentData.ContentDataID is int64 - directly usable!
```

### 4. Graceful Error Handling

```go
// Tracks partial failures
if fieldResult.ContentFieldID == 0 {
    failedFields = append(failedFields, fieldID)
}

// Returns different messages for different outcomes
return ContentCreatedMsg{...}           // Full success
return ContentCreatedWithErrorsMsg{...} // Partial success
return DbErrMsg{...}                    // Complete failure
```

### 5. Database Abstraction Maintained

```go
d := db.ConfigDB(*c)  // Returns DbDriver interface

// Polymorphic - works with SQLite, MySQL, PostgreSQL
contentData := d.CreateContentData(params)
```

The interface handles driver-specific implementation:
- SQLite: Uses `LastInsertId()` directly
- MySQL: Uses `LAST_INSERT_ID()`
- PostgreSQL: Uses `RETURNING` clause

---

## Comparison to Other Approaches

| Approach | Lines of Code | Uses Existing Code | Complexity | Portfolio Value |
|----------|--------------|-------------------|------------|-----------------|
| Raw Transaction | ~50 | ❌ Bypasses abstraction | Low | Poor (ignores architecture) |
| State Machine | ~100 | ✅ Uses messages | Medium | Moderate (doesn't solve ID problem) |
| Saga Pattern | ~500+ | ✅ Uses messages | Very High | Poor (over-engineered) |
| **Hybrid (This)** | **~150** | **✅ Uses DbDriver** | **Medium** | **Good (right tool for job)** |

---

## Implementation Checklist

### Phase 1: Core Implementation (Day 1)

- [ ] Add `CreateContentWithFields()` method to `commands.go`
- [ ] Add message types to `message_types.go`
- [ ] Add constructor to `constructors.go`
- [ ] Add `CollectFieldValuesFromForm()` helper

### Phase 2: Message Handlers (Day 1-2)

- [ ] Update `UpdateCms()` handler in `update_cms.go`
- [ ] Handle `CmsAddNewContentDataMsg`
- [ ] Handle `ContentCreatedMsg`
- [ ] Handle `ContentCreatedWithErrorsMsg`

### Phase 3: UI Integration (Day 2)

- [ ] Implement `ShowSuccessDialogCmd()`
- [ ] Implement `ShowWarningDialogCmd()`
- [ ] Implement `NavigateToContentListCmd()`
- [ ] Implement `RefreshContentListCmd()`

### Phase 4: Testing (Day 2-3)

- [ ] Test content creation with single field
- [ ] Test content creation with multiple fields
- [ ] Test partial failure scenario
- [ ] Test complete failure scenario
- [ ] Test UI refresh after success
- [ ] Test across SQLite, MySQL, PostgreSQL

### Phase 5: Tree Pointer Updates (Day 3 - Optional)

- [ ] Add tree pointer update logic if needed
- [ ] Handle parent/sibling relationships
- [ ] Test tree navigation after content creation

---

## Potential Issues and Mitigations

### Issue 1: FormState Structure Unknown

**Problem:** Not sure how form values map to field IDs

**Mitigation:**
- Read `form.go` and `form_model.go` to understand structure
- May need to modify form building to include field IDs
- Could use FormMap to store field metadata

### Issue 2: Dialog Commands Missing

**Problem:** `ShowSuccessDialogCmd()`, `ShowWarningDialogCmd()` may not exist

**Mitigation:**
- Check `dialog.go` for existing dialog implementation
- May already have `ShowDialogMsg` with title/message
- Can reuse existing dialog system

### Issue 3: Refresh Logic Unclear

**Problem:** How to refresh content list after creation

**Mitigation:**
- Check existing `DatabaseListMsg` pattern
- May need to add `RefreshSource` tracking
- Could use FetchSource pattern for this

### Issue 4: Tree Pointers Complexity

**Problem:** Sibling pointer updates are complex (parent, first_child, next_sibling, prev_sibling)

**Mitigation:**
- Start without tree pointer updates (leave null)
- Add as Phase 5 after basic creation works
- May need separate command for tree operations

---

## Alternative Designs Considered

### Alternative 1: Keep Generic Pattern, Add Context

```go
type DbResMsg struct {
    Result  sql.Result
    Table   string
    Context OperationContext  // NEW: carry context
}

type OperationContext struct {
    OperationType string
    NextStep      func(id int64) tea.Cmd
}
```

**Pros:**
- Less code change
- Keeps generic pattern

**Cons:**
- Still loses type safety
- Function pointers in messages (harder to debug)
- Doesn't solve partial failure tracking

**Verdict:** More complex than hybrid approach, less type-safe

### Alternative 2: Batch Command

```go
func (m Model) DatabaseBatch(operations []Operation) tea.Cmd {
    // Execute multiple operations in sequence
    // Track results and rollback on failure
}
```

**Pros:**
- Reusable for other coupled operations
- Transaction-like semantics

**Cons:**
- Requires defining Operation abstraction
- More generic = less type-safe
- Harder to track partial failures

**Verdict:** Over-abstraction for single use case

### Alternative 3: Async Coordination via Model State

```go
type Model struct {
    // ...
    PendingOperation *PendingContentCreation
}

type PendingContentCreation struct {
    Step          int
    ContentDataID *int64
    FieldValues   map[int64]string
}
```

**Pros:**
- Explicit state machine
- Can see operation in progress

**Cons:**
- Pollutes Model with operation state
- Complex state management
- Doesn't solve ID passing in generic commands

**Verdict:** Adds complexity without solving core issue

---

## Why Hybrid Approach Is Best

### Technical Reasons

1. **Uses existing infrastructure** - DbDriver methods already exist
2. **Type-safe** - Structs with IDs, not generic sql.Result
3. **Simple** - ~150 lines of code, not 500+
4. **Maintainable** - Easy to understand in 6 months
5. **Testable** - Can mock DbDriver interface

### Portfolio Reasons

1. **Shows good judgment** - Right tool for the job
2. **Demonstrates pragmatism** - Not over-engineering
3. **Respects constraints** - Works within existing architecture
4. **Problem-solving** - Identifies real issue (bypassing typed methods)
5. **Clean code** - Readable, maintainable solution

### Practical Reasons

1. **Fast to implement** - 2-3 days
2. **Low risk** - Doesn't change existing code
3. **Incrementally testable** - Can test each phase
4. **Extensible** - Pattern works for other coupled operations

---

## Future Extensions

Once this pattern works for ContentData + ContentFields:

### 1. Tree Pointer Updates

```go
func (m Model) UpdateTreePointers(contentDataID, parentID int64) tea.Cmd {
    // Use DbDriver.UpdateContentData() with typed params
}
```

### 2. Content Editing

```go
func (m Model) UpdateContentWithFields(contentDataID int64, fieldUpdates map[int64]string) tea.Cmd {
    // Update existing ContentFields using typed methods
}
```

### 3. Datatype Creation with Fields

```go
func (m Model) CreateDatatypeWithFields(datatypeParams, fieldParams) tea.Cmd {
    // Similar pattern: Datatype + DatatypeFields join
}
```

### 4. Batch Content Creation

```go
func (m Model) CreateMultipleContent(items []ContentCreationParams) tea.Cmd {
    // Reuse pattern for bulk operations
}
```

---

## Questions for Implementation

### Before Starting

1. **Form structure:** How are field IDs currently stored in FormState?
2. **Dialog system:** What commands exist for user feedback?
3. **Navigation:** How does `NavigateToContentListCmd()` work?
4. **Refresh:** How to trigger content list refresh after creation?

### During Implementation

1. **Error details:** Should we log which fields failed and why?
2. **Rollback:** Should we delete ContentData if ALL fields fail?
3. **Validation:** Should we validate field values before insert?
4. **Tree:** Should we support parent_id during creation, or always null?

### After Implementation

1. **Performance:** Is creating fields one-by-one fast enough? (probably yes for TUI)
2. **Batch insert:** Should we add batch insert for fields? (optimization, not MVP)
3. **Pattern reuse:** What other coupled operations need this pattern?

---

## Success Metrics

### Functional Success

- ✅ Can create ContentData with 1 field
- ✅ Can create ContentData with 10 fields
- ✅ Partial failures handled gracefully
- ✅ UI updates after successful creation
- ✅ Error messages show which fields failed
- ✅ Can navigate back to content list

### Code Quality Success

- ✅ Code is readable and maintainable
- ✅ No duplication (DRY principle followed)
- ✅ Type-safe where possible
- ✅ Error handling is comprehensive
- ✅ Follows existing code patterns

### Portfolio Success

- ✅ Demonstrates understanding of trade-offs
- ✅ Shows pragmatic decision-making
- ✅ Respects architectural constraints
- ✅ Solves real problem elegantly
- ✅ Would pass senior engineer code review

---

## Related Documents

- [PROBLEM.md](PROBLEM.md) - Full problem statement
- [UPDATE_SECTION_REVIEW.md](../packages/UPDATE_SECTION_REVIEW.md) - Update handler strengths/weaknesses
- [DATABASE_MESSAGE_FLOW_GUIDE.md](../packages/DATABASE_MESSAGE_FLOW_GUIDE.md) - Message flow patterns
- [MODEL_STRUCT_GUIDE.md](../packages/MODEL_STRUCT_GUIDE.md) - Model field reference

---

**Status:** Ready for implementation
**Next Step:** Begin Phase 1 - Core Implementation
**Estimated Completion:** 2026-01-18 (3 days from now)

# PROBLEM: CMS Content Creation in Message-Driven Architecture

**Date Identified:** 2026-01-15
**Status:** Blocking launch - core feature incomplete
**Complexity:** High - requires understanding 4 architectural layers

---

## Executive Summary

The TUI cannot complete CMS content creation workflows because coupled database operations (ContentData + ContentFields) cannot be expressed in the current message-driven architecture. The system can insert individual records but lacks a way to:

1. Pass results (inserted IDs) between operations
2. Chain dependent operations atomically
3. Handle partial failures in multi-step workflows
4. Coordinate UI updates after complex operations complete

This is blocking launch because **creating content is the core feature** of the terminal CMS.

---

## Real Goal (Not Theory)

**Primary Use Case:** Personal terminal CMS for hosting own website + portfolio piece

**Target Audience (Reality):**
- Self (developer using terminal for content management)
- Developers who want easy-to-deploy CMS with SQLite
- Portfolio reviewers evaluating enterprise architecture skills

**Attention Grabber:** TUI interface is the differentiator (TUI apps are trending)

**Not Rushing:** Want it complete and architecturally sound, not just "working"

---

## The Data Model

### CMS Content Structure

```
Datatype (schema definition)
  ├── parent_id → Datatypes (self-referential hierarchy)
  └── Fields (via datatypes_fields join table)
       ├── field_id
       ├── label
       ├── type (text, textarea, select, etc.)
       └── data (configuration)

ContentData (content instance)
  ├── content_data_id (PK)
  ├── datatype_id → Datatypes (RESTRICT - cannot delete datatype if content exists)
  ├── route_id → Routes (which page this content belongs to)
  ├── parent_id → ContentData (tree structure, self-referential)
  ├── first_child_id → ContentData (sibling-pointer tree optimization)
  ├── next_sibling_id → ContentData
  ├── prev_sibling_id → ContentData
  └── ContentFields (CASCADE delete - fields deleted when content deleted)
       ├── content_field_id (PK)
       ├── content_data_id → ContentData
       ├── field_id → Fields
       └── field_value (actual content)
```

### Required Operation Sequence

Creating new content requires atomic execution of:

```
1. INSERT ContentData → get content_data_id
2. For each Field in Datatype:
   → INSERT ContentField(content_data_id, field_id, field_value)
3. UPDATE tree pointers (parent_id, sibling pointers)
4. Navigate back to content list
5. Refresh UI with new content
```

**Current Status:** Step 1 works. Steps 2-5 are not implemented.

---

## Current Architecture (4 Layers)

### Layer 1: TUI Message Handlers

**Location:** `internal/cli/update_*.go`

**Pattern:** Elm Architecture (Model-Update-View)

**Constraints:**
- All state changes via messages
- Cannot call functions, only send messages
- Cannot block or wait for results
- Must return `(Model, tea.Cmd)`

**Example:**
```go
// update_forms.go:49-73
case FormActionMsg:
    switch msg.Action {
    case INSERT:
        // Filters and prepares values
        return m, tea.Batch(
            DatabaseInsertCmd(db.DBTable(msg.Table), filteredColumns, filteredValues),
            LogMessageCmd(fmt.Sprintln(filteredColumns)),
        )
    }
```

**Problem:** Once DatabaseInsertCmd is dispatched, no way to wait for result and chain next operation.

### Layer 2: CLI Commands (Async Wrappers)

**Location:** `internal/cli/commands.go`

**Pattern:** Async commands that return `tea.Cmd`

**Signature:**
```go
func (m Model) DatabaseInsert(c *config.Config, table db.DBTable, columns []string, values []*string) tea.Cmd
```

**Implementation:**
```go
// commands.go:58-89
func (m Model) DatabaseInsert(...) tea.Cmd {
    d := db.ConfigDB(*c)
    con, _, err := d.GetConnection()  // ← Bypasses typed methods

    sqb := db.NewSecureQueryBuilder(con)  // ← Uses generic query builder
    query, args, err := sqb.SecureBuildInsertQuery(string(table), valuesMap)
    res, err := sqb.SecureExecuteModifyQuery(query, args)

    return tea.Batch(
        DbResultCmd(res, string(table)),  // ← Returns sql.Result (no ID!)
    )
}
```

**Problem:** Returns `sql.Result` which doesn't expose `LastInsertId()` at message handler level.

### Layer 3: SecureQueryBuilder

**Location:** `internal/db/` (query builder)

**Pattern:** Parameterized SQL generation

**Purpose:** Prevent SQL injection with prepared statements

**Methods:**
- `SecureBuildInsertQuery(table, valuesMap) → (query, args, error)`
- `SecureBuildSelectQuery(table, id) → (query, args, error)`
- `SecureExecuteModifyQuery(query, args) → (sql.Result, error)`

**Problem:** Generic interface loses type information. Works for single operations, can't chain.

### Layer 4: DbDriver Interface

**Location:** `internal/db/db.go`

**Pattern:** Database abstraction with 3 implementations (SQLite, MySQL, PostgreSQL)

**Interface Size:** 200+ methods for typed operations

**Relevant Methods:**
```go
// Typed, synchronous methods that RETURN structs with IDs
CreateContentData(CreateContentDataParams) ContentData
CreateContentField(CreateContentFieldParams) ContentFields
CreateDatatype(CreateDatatypeParams) Datatypes
// ... 200+ more
```

**Problem:** CLI layer never uses these typed methods! It bypasses them for generic query builder.

---

## Why Coupled Operations Are Hard

### The ID Problem

**ContentFields table schema:**
```sql
CREATE TABLE content_fields (
    content_field_id INTEGER PRIMARY KEY,
    content_data_id INTEGER NOT NULL,  -- ← NEEDS ID from ContentData insert
    field_id INTEGER NOT NULL,
    field_value TEXT,
    FOREIGN KEY (content_data_id) REFERENCES content_data(content_data_id) ON DELETE CASCADE
);
```

**Current flow:**
```
User submits form
  ↓
FormActionMsg(INSERT) dispatched
  ↓
DatabaseInsertCmd sent (async)
  ↓
DatabaseInsert executes in goroutine
  ↓
Returns DbResMsg(sql.Result, table)
  ↓
Handler receives DbResMsg
  ↓
❌ STUCK: No way to get LastInsertId() from sql.Result
  ↓
❌ STUCK: No way to know which operation this result belongs to
  ↓
❌ STUCK: No way to trigger next operation with the ID
```

### The Message-Driven Constraint

**Cannot do this (synchronous):**
```go
// ❌ This doesn't work in message-driven architecture
contentDataID := DatabaseInsert(...)  // Can't block for result
for field := range fields {
    DatabaseInsert(contentFieldsTable, contentDataID, field)
}
```

**Must do this (message-driven):**
```go
// ✅ Can only send messages
return m, DatabaseInsertCmd(...)  // Async, returns immediately
// ❌ But how do we send the NEXT message with the result?
```

### The Abstraction Layer Constraint

**Cannot do this (bypasses abstraction):**
```go
// ❌ Breaks multi-database support
tx, _ := db.Begin()  // Which db? SQLite? MySQL? Postgres?
tx.Exec("INSERT INTO content_data ...")
id, _ := result.LastInsertId()
tx.Exec("INSERT INTO content_fields ...")
tx.Commit()
```

**Must use this (respects abstraction):**
```go
// ✅ Works across SQLite, MySQL, PostgreSQL
d := db.ConfigDB(config)  // Returns DbDriver interface
d.CreateContentData(params)  // Polymorphic - works with any driver
```

### The Result Handling Problem

**Current DbResMsg handler:**
```go
// update_forms.go:90-93
case DbResMsg:
    return m, tea.Batch(
        LogMessageCmd(fmt.Sprintf("Database operation completed for table %s", msg.Table)),
    )
    // ❌ INCOMPLETE:
    // - No UI refresh
    // - No navigation back
    // - No table reload
    // - No way to trigger dependent operations
```

---

## What's Been Tried

### Attempt 1: Use FetchSource Pattern

**Location:** `update_database.go:10-16`

```go
type FetchSource string
const (
    DATATYPEMENU    FetchSource = "datatype_menu"
    BUILDTREE       FetchSource = "build_tree"
    PICKCONTENTDATA FetchSource = "fetch_source"
)
```

**Status:** Exists for READ operations, not used consistently for WRITE operations

**Problem:** Doesn't solve ID passing or operation chaining

### Attempt 2: Empty Handler for CMS Operations

**Location:** `update_cms.go:26-29`

```go
case CmsAddNewContentDataMsg:
    // Database commands
    // New ContentData also requires new ContentFields
    return m, tea.Batch()  // ❌ EMPTY - NO IMPLEMENTATION
```

**Status:** Stub exists, never implemented

**Reason:** Developer didn't know how to chain operations

---

## Constraints That Must Be Respected

### 1. Message-Driven Architecture (Non-Negotiable)

- ✅ Elm Architecture pattern (Bubbletea framework)
- ✅ All state changes via messages
- ✅ Commands return `tea.Cmd`, not results
- ✅ Update handlers return `(Model, tea.Cmd)`

### 2. Database Abstraction Layer (Non-Negotiable)

- ✅ DbDriver interface with 3 implementations
- ✅ Must work with SQLite, MySQL, PostgreSQL
- ✅ Type-safe operations via interface methods
- ✅ Cannot bypass abstraction for raw SQL

### 3. Security (Non-Negotiable)

- ✅ Parameterized queries (no SQL injection)
- ✅ SecureQueryBuilder already implemented
- ✅ All user input sanitized

### 4. No Framework Changes (Preference)

- ✅ Don't want to refactor entire Bubbletea architecture
- ✅ Don't want to replace database abstraction
- ✅ Want to add capabilities, not replace existing patterns

---

## Failed Approaches (External Suggestions)

### Suggestion 1: "Just Use Database Transactions"

**Why it failed:**
- Ignores message-driven constraint
- Ignores database abstraction layer
- Would require accessing raw `*sql.DB` directly
- Would duplicate transaction logic across 3 database drivers

### Suggestion 2: "Use Simple State Machine"

**Why it's insufficient:**
- Doesn't solve the ID passing problem
- Adds state tracking but not operation coordination
- Still stuck on `sql.Result` not exposing `LastInsertId()` in message handlers

### Suggestion 3: "Implement Saga Pattern"

**Why it's over-engineered:**
- Sagas are for distributed systems (microservices)
- Would add 500+ lines of code for 2 INSERT statements
- Reinvents database transaction features in application layer
- Too complex for portfolio piece (looks like poor judgment)

---

## The Core Confusion

**Developer quote:**
> "I feel stuck in a loop of where messages are sent then how those messages turn into actions within the application, and how to return data from database operations and distribute that info into ui elements."

**Translation:**
1. How do I pass operation results between message handlers?
2. How do I know which result belongs to which operation?
3. How do I chain dependent operations?
4. How do I update the UI after complex operations complete?

**Root Cause:** CLI Commands layer uses generic query builder instead of typed DbDriver methods.

---

## What Needs To Work For Launch

### Minimum Viable CMS Operations

**Create Content:**
1. User fills form for new content
2. Submit → Insert ContentData
3. Get inserted ContentData ID
4. For each Field in Datatype:
   - Insert ContentField with ContentData ID
5. Update tree pointers (if parent exists)
6. Navigate back to content list
7. Show success message
8. Refresh content list

**Additional Operations (Nice to Have):**
- Edit existing content (update ContentFields)
- Delete content (CASCADE handles ContentFields)
- View content tree (already works)

---

## Success Criteria

### Functional Requirements

- ✅ Can create ContentData with multiple ContentFields atomically
- ✅ Foreign key relationships maintained (content_data_id)
- ✅ Tree pointers updated correctly (parent, siblings)
- ✅ UI refreshes after operation completes
- ✅ Error handling for partial failures
- ✅ Works across all 3 database drivers

### Architectural Requirements

- ✅ Fits message-driven architecture (no breaking changes)
- ✅ Respects database abstraction layer
- ✅ Type-safe where possible
- ✅ Maintainable and understandable
- ✅ Portfolio-worthy (demonstrates good judgment)

### Non-Requirements

- ❌ Doesn't need to be "enterprise-grade" with Saga pattern
- ❌ Doesn't need to handle distributed transactions
- ❌ Doesn't need compensation/rollback (database handles atomicity)
- ❌ Doesn't need to be the most abstract/flexible solution

---

## Key Files to Understand

### Message Handlers
- `internal/cli/update_forms.go` - Form submission handling (90% complete)
- `internal/cli/update_cms.go` - CMS operations (0% complete)
- `internal/cli/update_database.go` - Database message routing

### Commands Layer
- `internal/cli/commands.go` - Async database operations
- `internal/cli/constructors.go` - Message and command constructors

### Messages
- `internal/cli/message_types.go` - All message type definitions

### Database Layer
- `internal/db/db.go` - DbDriver interface (200+ methods)
- `internal/db/content_data.go` - ContentData CRUD (typed methods)
- `internal/db/content_field.go` - ContentField CRUD (typed methods)

### Schema
- `sql/schema/16_content_data/schema.sql` - ContentData table
- `sql/schema/17_content_fields/schema.sql` - ContentFields table
- `sql/schema/22_joins/queries.sql` - Complex join queries

---

## Next Steps

This document should be used as context for:

1. **Evaluating proposed solutions** - Does it address all constraints?
2. **Planning implementation** - What needs to change in which files?
3. **Reviewing similar problems** - Other coupled operations in codebase?
4. **Portfolio presentation** - Understanding complexity demonstrates architecture skills

---

**Last Updated:** 2026-01-15
**Related Documents:**
- [SUGGESTION-2026-01-15.md](SUGGESTION-2026-01-15.md) - Hybrid approach proposal
- [UPDATE_SECTION_REVIEW.md](../packages/UPDATE_SECTION_REVIEW.md) - Update handler analysis
- [DATABASE_MESSAGE_FLOW_GUIDE.md](../packages/DATABASE_MESSAGE_FLOW_GUIDE.md) - Message flow patterns
- [MODEL_STRUCT_GUIDE.md](../packages/MODEL_STRUCT_GUIDE.md) - Model field reference

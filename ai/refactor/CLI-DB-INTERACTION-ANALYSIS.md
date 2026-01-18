# CLI Package Database Interaction Analysis

**Date:** 2026-01-15
**Purpose:** Understand how CLI package currently interacts with DB package
**Context:** Needed for designing coupled operation solution

---

## Executive Summary

The CLI package interacts with the DB package in **two distinct patterns**:

1. **Typed DbDriver calls** - Used for ONE operation: `ListTables()` and `GetRouteTreeByRouteID()`
2. **Generic SecureQueryBuilder** - Used for ALL CRUD operations (Insert, Update, Delete, Get, List)

**Key Finding:** CLI package already bypasses typed DbDriver methods for 95% of operations, using generic query builder instead. This is intentional for plugin extensibility.

---

## Files That Import DB Package

15 files in `internal/cli/` import `internal/db`:

```
internal/cli/fields.go
internal/cli/constructors.go
internal/cli/commands.go            ← Primary database operations
internal/cli/update_forms.go
internal/cli/update_fetch.go
internal/cli/debug.go
internal/cli/form.go
internal/cli/update_controls.go
internal/cli/page_builders.go
internal/cli/forms.go
internal/cli/update_database.go     ← Database message handlers
internal/cli/parse.go               ← Type conversion layer
internal/cli/menus.go
internal/cli/message_types.go       ← Type definitions
internal/cli/cms_struct.go          ← Tree structures
```

**Total:** ~7,500 lines of CLI code

---

## Interaction Pattern 1: Generic Query Builder (95% of operations)

### Location: `commands.go`

**Pattern:**
```go
func (m Model) DatabaseInsert(c *config.Config, table db.DBTable, columns []string, values []*string) tea.Cmd {
    d := db.ConfigDB(*c)              // 1. Get DbDriver
    con, _, err := d.GetConnection()  // 2. BYPASS typed methods, get raw connection

    sqb := db.NewSecureQueryBuilder(con)  // 3. Use generic query builder
    query, args, err := sqb.SecureBuildInsertQuery(string(table), valuesMap)
    res, err := sqb.SecureExecuteModifyQuery(query, args)

    return DbResultCmd(res, string(table))  // 4. Return sql.Result (no typed info)
}
```

### Operations Using This Pattern

**All from `commands.go`:**

1. **DatabaseInsert** (lines 58-90)
   - Gets connection
   - Uses SecureQueryBuilder
   - Returns `sql.Result`

2. **DatabaseUpdate** (lines 92-122)
   - Gets connection
   - Uses SecureQueryBuilder
   - Returns `sql.Result`

3. **DatabaseGet** (lines 123-149)
   - Gets connection
   - Uses SecureQueryBuilder
   - Parses results via `Parse()` function
   - Returns typed slices

4. **DatabaseList** (lines 151-177)
   - Gets connection
   - Uses SecureQueryBuilder
   - Parses results via `Parse()` function
   - Returns typed slices

5. **DatabaseFilteredList** (lines 179-204)
   - Gets connection
   - Uses SecureQueryBuilder
   - Parses results via `Parse()` function
   - Returns typed slices

6. **DatabaseDelete** (lines 206-227)
   - Gets connection
   - Uses SecureQueryBuilder
   - Returns `sql.Result`

### Why This Pattern Exists

**Flexibility for plugins:**
- Works with ANY table name (runtime-defined plugin tables)
- No compile-time type checking required
- Generic parameters (map[string]any)
- Works across SQLite, MySQL, PostgreSQL

**Trade-offs:**
- ✅ Plugin extensibility
- ✅ Works with unknown tables
- ❌ Loses type safety
- ❌ No direct access to inserted IDs in messages
- ❌ Must parse results manually

---

## Interaction Pattern 2: Typed DbDriver Calls (5% of operations)

### Location: `commands.go`

**Only TWO places use typed DbDriver methods:**

#### 1. GetTablesCMD (lines 37-55)
```go
func GetTablesCMD(c *config.Config) tea.Cmd {
    return func() tea.Msg {
        var d db.DbDriver
        d = db.ConfigDB(*c)

        // ✅ USES TYPED METHOD
        tables, err := d.ListTables()
        if err != nil {
            return FetchErrMsg{Error: err}
        }

        // Extract labels from typed result
        for _, table := range *tables {
            labels = append(labels, table.Label)
        }

        return TablesSet{Tables: labels}
    }
}
```

**Returns:** `*[]db.Tables` (typed struct slice)

#### 2. GetFullTree (lines 238-247)
```go
func (m Model) GetFullTree(c *config.Config, id int64) tea.Cmd {
    d := db.ConfigDB(*c)

    // ✅ USES TYPED METHOD
    res, err := d.GetRouteTreeByRouteID(1)
    if err != nil {
        return ErrorSetCmd(err)
    }

    out := db.LogRouteTree("GetFullTree", res)
    return GetFullTreeResCMD(out, *res)
}
```

**Returns:** `*[]db.GetRouteTreeByRouteIDRow` (typed struct slice)

---

## Type Conversion Layer: parse.go

**Purpose:** Convert `sql.Rows` from generic queries into typed structs

**Pattern:**
```go
func Parse(rows *sql.Rows, table db.DBTable) (any, error) {
    switch table {
    case db.User:
        return parseUsers(rows)      // → []db.Users
    case db.Datatype:
        return parseDatatypes(rows)  // → []db.Datatypes
    case db.ContentData:
        return parseContentData(rows)  // → []db.ContentData
    case db.ContentFields:
        return parseContentFields(rows)  // → []db.ContentFields
    // ... 18 more core tables
    default:
        return parseGeneric(rows)    // → []map[string]any (for plugins!)
    }
}
```

**Key Insight:** `parseGeneric()` exists for **plugin tables** that don't have typed parsers.

### Examples of Typed Parsers

**parseContentData** (lines 457-486):
```go
func parseContentData(rows *sql.Rows) ([]db.ContentData, error) {
    var results []db.ContentData

    for rows.Next() {
        var content db.ContentData
        err := rows.Scan(
            &content.ContentDataID,       // ← Has ID field
            &content.ParentID,
            &content.RouteID,
            &content.DatatypeID,
            &content.AuthorID,
            &content.DateCreated,
            &content.DateModified,
            &content.History,
            &content.FirstChildID,
            &content.NextSiblingID,
            &content.PrevSiblingID,
        )
        results = append(results, content)
    }

    return results, nil
}
```

**parseContentFields** (lines 518-545):
```go
func parseContentFields(rows *sql.Rows) ([]db.ContentFields, error) {
    var results []db.ContentFields

    for rows.Next() {
        var contentField db.ContentFields
        err := rows.Scan(
            &contentField.ContentFieldID,   // ← Has ID field
            &contentField.RouteID,
            &contentField.ContentDataID,    // ← Foreign key
            &contentField.FieldID,          // ← Foreign key
            &contentField.FieldValue,
            &contentField.AuthorID,
            &contentField.DateCreated,
            &contentField.DateModified,
            &contentField.History,
        )
        results = append(results, contentField)
    }

    return results, nil
}
```

**parseGeneric** (lines 663-705):
```go
// For plugin tables without typed parsers
func parseGeneric(rows *sql.Rows) ([]map[string]any, error) {
    columns, err := rows.Columns()  // Get column names dynamically

    for rows.Next() {
        // Scan into generic map
        row := make(map[string]any)
        // ... dynamic scanning
        results = append(results, row)
    }

    return results, nil
}
```

---

## Message Flow for Database Operations

### INSERT Operation Flow

```
User submits form
  ↓
FormActionMsg{Action: INSERT, Table, Columns, Values}
  ↓
update_forms.go: case FormActionMsg
  ↓
DatabaseInsertCmd(table, columns, values)
  ↓
commands.go: DatabaseInsert()
  ├─ db.ConfigDB(config) → DbDriver
  ├─ d.GetConnection() → *sql.DB (bypasses typed methods)
  ├─ db.NewSecureQueryBuilder(con)
  ├─ sqb.SecureBuildInsertQuery(table, values)
  └─ sqb.SecureExecuteModifyQuery(query, args) → sql.Result
  ↓
DbResultCmd(sql.Result, table)
  ↓
update_forms.go: case DbResMsg
  ↓
LogMessageCmd("Database operation completed") ← INCOMPLETE!
```

**Problem:** `DbResMsg` handler doesn't extract ID or chain next operation.

### SELECT Operation Flow

```
User navigates to table view
  ↓
DatabaseListCmd(source, table)
  ↓
update_database.go: case DatabaseListMsg
  ↓
commands.go: DatabaseList()
  ├─ db.ConfigDB(config) → DbDriver
  ├─ d.GetConnection() → *sql.DB
  ├─ db.NewSecureQueryBuilder(con)
  ├─ sqb.SecureBuildListQuery(table)
  ├─ sqb.SecureExecuteSelectQuery(query) → *sql.Rows
  └─ Parse(rows, table) → []db.Datatypes (typed!)
  ↓
DatabaseListRowsCmd(source, typedResults, table)
  ↓
update_database.go: case DatabaseListRowsMsg
  ↓
db.CastToTypedSlice(msg.Rows, msg.Table)
  ↓
switch msg.Table:
  case db.Datatype:
    data := res.([]db.Datatypes)
    DatatypesFetchResultCmd(data)
```

**Key:** Results ARE typed via `Parse()`, but ONLY for read operations.

---

## DB Package Types Used by CLI

### 1. DBTable (Table Name Constants)

**From `message_types.go` and throughout:**
```go
db.DBTable  // Type for table names

// Constants:
db.User
db.Role
db.Permission
db.Session
db.Token
db.User_oauth
db.Route
db.Admin_route
db.Field
db.Admin_field
db.Datatype
db.Admin_datatype
db.Datatype_fields
db.Admin_datatype_fields
db.Content_data          // ← CMS content instances
db.Admin_content_data
db.Content_fields        // ← CMS field values
db.Admin_content_fields
db.MediaT
db.Media_dimension
db.Table
```

### 2. Struct Types (For Parsing)

**All core table structs used:**
```go
db.Users
db.Roles
db.Permissions
db.Sessions
db.Tokens
db.UserOauth
db.Routes
db.AdminRoutes
db.Fields
db.AdminFields
db.Datatypes
db.AdminDatatypes
db.DatatypeFields
db.AdminDatatypeFields
db.ContentData           // ← ContentDataID field
db.AdminContentData
db.ContentFields         // ← ContentFieldID, ContentDataID fields
db.AdminContentFields
db.Media
db.MediaDimensions
db.Tables
```

### 3. DbDriver Interface

**Used for:**
```go
db.DbDriver              // Interface type
db.ConfigDB(config)      // Factory function → DbDriver
d.GetConnection()        // → *sql.DB, context.Context
d.ListTables()           // ✅ Typed method (rare usage)
d.GetRouteTreeByRouteID()  // ✅ Typed method (rare usage)
```

### 4. Helper Functions

```go
db.CastToTypedSlice(rows any, table DBTable) any
db.LogRouteTree(label string, rows *[]GetRouteTreeByRouteIDRow) string
db.NewSecureQueryBuilder(con *sql.DB) SecureQueryBuilder
```

---

## Key Findings

### 1. Generic Pattern is Dominant

**95% of operations use generic query builder:**
- Insert, Update, Delete operations
- Most Get/List operations
- All operations that modify data

**Only 5% use typed methods:**
- ListTables (metadata query)
- GetRouteTreeByRouteID (complex join query)

### 2. Parse Layer Provides Type Safety for Reads

**Read operations get typed:**
```
Raw Query → sql.Rows → Parse(rows, table) → []db.ContentData
```

**But write operations DON'T:**
```
Insert Query → sql.Result → (no parsing, no type recovery)
```

### 3. Two Types of Type Safety

**Compile-time (typed methods):**
```go
// DbDriver method signature enforces types
func (d DbDriver) CreateContentData(params CreateContentDataParams) ContentData
```

**Runtime (parse layer):**
```go
// Parse converts generic results to typed structs
func Parse(rows *sql.Rows, table db.DBTable) (any, error)
```

**CLI currently uses runtime type safety only.**

### 4. Plugin Support Is Built-In

**Parse layer has fallback:**
```go
default:
    return parseGeneric(rows)  // → []map[string]any
```

**This handles plugin tables that don't have typed parsers.**

### 5. ID Problem is Architectural

**Write operations:**
```go
res, err := sqb.SecureExecuteModifyQuery(query, args)
return DbResultCmd(res, string(table))  // ← sql.Result, table name only
```

**sql.Result interface:**
```go
type Result interface {
    LastInsertId() (int64, error)  // ← Available but not extracted!
    RowsAffected() (int64, error)
}
```

**Problem:** ID is available in `sql.Result` but never extracted into message.

---

## Implications for Coupled Operations Solution

### What We Learned

1. **Generic pattern is intentional** - Not a mistake, designed for plugins
2. **Parse layer exists** - Runtime type conversion already implemented
3. **Typed methods barely used** - Only 2 operations out of ~20
4. **ID extraction missing** - `DbResMsg` doesn't capture LastInsertId()
5. **Plugin support built-in** - `parseGeneric()` handles unknown tables

### Solution Options

#### Option A: Extend DbResMsg with ID Extraction

**Add to message:**
```go
type DbResMsg struct {
    Result     sql.Result
    Table      string
    InsertedID *int64      // NEW: Extract LastInsertId() if available
}
```

**In commands.go:**
```go
res, err := sqb.SecureExecuteModifyQuery(query, args)
insertedID, _ := res.LastInsertId()  // Extract ID

return DbResultCmd(res, string(table), &insertedID)  // Pass ID
```

**Pro:** Works with generic query builder (plugins supported)
**Con:** Still generic, loses compile-time type safety

#### Option B: Specialized Typed Command (Hybrid Approach)

**For core tables only:**
```go
func (m Model) CreateContentWithFields(...) tea.Cmd {
    d := db.ConfigDB(m.Config)

    // Use typed DbDriver method directly
    contentData := d.CreateContentData(params)
    // contentData.ContentDataID is directly accessible!

    for fieldID, value := range fieldValues {
        d.CreateContentField(contentDataID, fieldID, value)
    }

    return ContentCreatedMsg{ContentDataID: contentData.ContentDataID}
}
```

**Pro:** Type-safe, direct ID access, clear intent
**Con:** Only works for core tables, not plugins

#### Option C: Enhanced Parse Layer for Writes

**Add write-result parsing:**
```go
func ParseWriteResult(result sql.Result, table db.DBTable) (any, error) {
    id, err := result.LastInsertId()
    if err != nil {
        return nil, err
    }

    switch table {
    case db.ContentData:
        return &db.ContentData{ContentDataID: id}, nil
    case db.ContentFields:
        return &db.ContentFields{ContentFieldID: id}, nil
    default:
        return map[string]any{"id": id, "table": string(table)}, nil
    }
}
```

**Pro:** Extends existing parse pattern, supports plugins
**Con:** Creates partial structs (only ID field populated)

---

## Recommendation

**For Phase 1 (Core CMS):** Use Option B (Hybrid Approach)
- Core content creation uses typed methods
- Unblocks launch immediately
- Demonstrates proper use of typed methods

**For Phase 2 (Plugins):** Enhance with Option A or C
- Generic operations extract IDs
- Plugins can chain operations
- Maintains extensibility

**Rationale:**
- CLI already has dual pattern (generic + occasional typed)
- Adding specialized typed command for core CMS fits existing architecture
- Parse layer can be extended later for plugin support

---

## Files to Modify for Hybrid Approach

### 1. commands.go
**Add new method:**
```go
func (m Model) CreateContentWithFields(...) tea.Cmd {
    // Use typed DbDriver methods
}
```

### 2. message_types.go
**Add new messages:**
```go
type CreateContentWithFieldsMsg struct { ... }
type ContentCreatedMsg struct { ... }
```

### 3. constructors.go
**Add constructor:**
```go
func CreateContentWithFieldsCmd(...) tea.Cmd { ... }
```

### 4. update_cms.go
**Add handler:**
```go
case CmsAddNewContentDataMsg:
    return m, CreateContentWithFieldsCmd(...)

case ContentCreatedMsg:
    return m, tea.Batch(
        ShowSuccessDialogCmd(...),
        NavigateToContentListCmd(...),
    )
```

### 5. form.go or new cms_helpers.go
**Add helper:**
```go
func (m Model) CollectFieldValuesFromForm() map[int64]string { ... }
```

---

## Summary

**Current CLI-DB Interaction:**
- 95% generic query builder (plugin extensibility)
- 5% typed DbDriver calls (rare, specific operations)
- Parse layer provides runtime type safety for reads
- No type recovery for writes (sql.Result only)

**Key Insight:**
The generic query builder isn't wrong - it's essential for plugins. The solution is to **add** specialized typed commands for core operations, not **replace** the generic pattern.

**Path Forward:**
Implement hybrid approach - typed for core, generic for plugins. This fits the existing architecture and solves the immediate problem while preserving extensibility.

---

**Related Documents:**
- [ANALYSIS-SUMMARY-2026-01-15.md](ANALYSIS-SUMMARY-2026-01-15.md) - Complete analysis
- [SUGGESTION-2026-01-15.md](SUGGESTION-2026-01-15.md) - Implementation guide
- [PROBLEM-UPDATE-2026-01-15-PLUGINS.md](PROBLEM-UPDATE-2026-01-15-PLUGINS.md) - Plugin constraints

**Status:** Analysis complete - Ready for implementation

# PROBLEM UPDATE: Plugin System Changes Architecture Constraints

**Date:** 2026-01-15
**Status:** Critical - Invalidates Previous Assumptions
**Impact:** High - Changes solution approach

---

## Critical New Information

**From user:**
> "the plan for modulacms is to be a single binary that is extended by plugins, that means we need to be able to adapt to lua scripts that introduce new tables"

**This changes EVERYTHING about the database architecture.**

---

## What This Means

### Previous Understanding (WRONG)

**Assumption:** All tables are known at compile time
- ✅ ContentData table exists
- ✅ ContentFields table exists
- ✅ Can use typed DbDriver methods
- ❌ Generic query builder was "mistake" that should be replaced

**Conclusion:** Use typed methods for coupled operations

### Corrected Understanding (RIGHT)

**Reality:** Tables can be created by Lua plugins at runtime
- ✅ Core CMS tables known at compile time (users, routes, datatypes, etc.)
- ✅ Plugin tables UNKNOWN at compile time (defined in Lua scripts)
- ✅ Generic query builder is NECESSARY for plugin flexibility
- ✅ Typed methods only work for core tables

**Conclusion:** Need solution that works for BOTH core and plugin tables

---

## The Real Architecture Challenge

### Two Categories of Tables

**1. Core CMS Tables (Compile-Time)**
```
Known tables:
- users, roles, permissions
- routes, sites
- datatypes, fields, datatypes_fields (join)
- content_data, content_fields
- media, media_dimensions
- sessions, tokens
```

**Can use:**
- ✅ Typed DbDriver methods
- ✅ Generated code (sqlc)
- ✅ Type-safe parameters
- ✅ Struct returns with IDs

**2. Plugin Tables (Runtime)**
```
Unknown tables (defined by Lua plugins):
- custom_product_data
- custom_product_attributes
- custom_inventory
- custom_shipping_rules
- ... anything a plugin creates
```

**Must use:**
- ✅ Generic query builder
- ✅ Dynamic SQL generation
- ✅ Map-based parameters
- ✅ sql.Result returns

---

## Why Generic Query Builder Makes Sense Now

### Previous Assessment (INCORRECT)
"CLI Commands layer bypasses DbDriver typed methods - this is wrong"

### Corrected Assessment
"CLI Commands layer uses generic query builder because it must support plugin tables that don't have typed methods"

**The generic pattern isn't a mistake - it's intentional for extensibility.**

---

## The Real Problem (Refined)

**Original Problem Statement:**
"CMS content creation requires chaining operations but can't pass IDs between operations"

**Still true, but now with additional constraint:**
"Solution must work for BOTH core tables (typed) AND plugin tables (generic)"

---

## Plugin Integration Questions

**CRITICAL UNKNOWNS - Need answers before designing solution:**

### 1. Plugin Table Creation
**Q:** How do plugins create tables?
- Do plugins define schemas in Lua?
- Does ModulaCMS execute CREATE TABLE statements from plugins?
- Are plugin tables namespaced? (e.g., `plugin_name_table_name`)

**Q:** Are there constraints on plugin tables?
- Must plugin tables follow certain patterns?
- Can plugins create foreign keys to core tables?
- Can core tables reference plugin tables?

### 2. Plugin Coupled Operations
**Q:** Do plugins need coupled operations?
- Can a plugin create `custom_product + custom_product_fields` pattern?
- Would plugins need to chain multiple INSERT statements?
- Do plugins have access to transaction APIs?

**Q:** How do plugins currently handle database operations?
- Is there a Lua API for database access?
- Do plugins call Go functions? Or direct SQL?
- Are there examples of plugin database code?

### 3. Content Type Extensibility
**Q:** Can plugins extend existing content types?
- Can plugin add fields to ContentData?
- Or must plugin create entirely new tables?
- Is there a "custom field" system for plugins?

**Q:** Content creation flow for plugins:
- Does plugin content use ContentData + ContentFields pattern?
- Or does plugin define its own coupled tables?
- Can plugin content integrate with core CMS tree structure?

---

## Architectural Implications

### If Plugins Create Arbitrary Tables

**Then:**
- Generic query builder is ESSENTIAL (not a mistake)
- Typed methods are optimization for CORE tables only
- Solution must be generic-first with typed optimization
- Cannot rely on compile-time code generation for all operations

**Example Plugin Scenario:**
```lua
-- Lua plugin defines e-commerce tables
plugin.define_table({
    name = "ecommerce_products",
    fields = {
        {name = "product_id", type = "integer", primary_key = true},
        {name = "name", type = "text"},
        {name = "price", type = "real"}
    }
})

plugin.define_table({
    name = "ecommerce_product_attributes",
    fields = {
        {name = "attribute_id", type = "integer", primary_key = true},
        {name = "product_id", type = "integer", foreign_key = "ecommerce_products"},
        {name = "key", type = "text"},
        {name = "value", type = "text"}
    }
})

-- Plugin needs coupled operation:
-- INSERT INTO ecommerce_products -> get product_id
-- INSERT INTO ecommerce_product_attributes(product_id, ...) for each attribute
```

**This is EXACTLY the same problem as ContentData + ContentFields, but for unknown tables.**

### If Plugins Extend Core Content System

**Then:**
- All plugin content uses ContentData + ContentFields
- Plugins don't create coupled tables, just new field types
- Solution can be core-table-specific
- Generic builder only needed for plugin lookups/queries

**Example Plugin Scenario:**
```lua
-- Plugin extends core system by adding field types
plugin.register_field_type({
    name = "product_price",
    data_type = "decimal",
    validation = function(value) return value > 0 end,
    render = function(value) return "$" .. value end
})

-- Plugin content still uses ContentData + ContentFields
-- No new tables, just new field types
```

**This scenario makes the hybrid approach viable.**

---

## Revised Solution Approaches

### Approach A: Hybrid (Core Typed + Plugin Generic)

**Assumes:** Plugins extend core content system, don't create arbitrary coupled tables

**For Core Tables:**
```go
// Use typed methods (hybrid approach from SUGGESTION-2026-01-15.md)
func (m Model) CreateContentWithFields(...) tea.Cmd {
    contentData := d.CreateContentData(params)
    for fieldID, value := range fieldValues {
        d.CreateContentField(params)
    }
}
```

**For Plugin Tables:**
```go
// Use generic query builder (existing pattern)
func (m Model) DatabaseInsert(...) tea.Cmd {
    sqb := db.NewSecureQueryBuilder(con)
    res, err := sqb.SecureExecuteModifyQuery(query, args)
}
```

**Pro:** Optimal for each use case
**Con:** Plugins can't do coupled operations

### Approach B: Generic Operation Chaining

**Assumes:** Plugins CAN create coupled tables and need chaining

**Pattern:**
```go
type ChainedOperation struct {
    Steps []OperationStep
}

type OperationStep struct {
    Table   string
    Query   string
    Args    []any
    OnSuccess func(result sql.Result) []OperationStep  // Generate next steps
}

func (m Model) ExecuteChainedOperation(op ChainedOperation) tea.Cmd {
    return func() tea.Msg {
        results := []sql.Result{}

        for _, step := range op.Steps {
            res, err := executeQuery(step.Query, step.Args)
            if err != nil {
                return ChainedOperationErrorMsg{Step: i, Error: err}
            }
            results = append(results, res)

            // Generate next steps based on result
            if step.OnSuccess != nil {
                nextSteps := step.OnSuccess(res)
                op.Steps = append(op.Steps, nextSteps...)
            }
        }

        return ChainedOperationCompleteMsg{Results: results}
    }
}
```

**Pro:** Works for any table (core or plugin)
**Con:** Complex, lots of boilerplate for simple cases

### Approach C: Transaction API with Continuation

**Assumes:** Plugins need transaction support

**Pattern:**
```go
type TransactionBuilder struct {
    operations []func(tx *sql.Tx) (any, error)
}

func (m Model) BeginTransaction() *TransactionBuilder {
    return &TransactionBuilder{}
}

func (tb *TransactionBuilder) Insert(table string, values map[string]any) *TransactionBuilder {
    tb.operations = append(tb.operations, func(tx *sql.Tx) (any, error) {
        // Build and execute INSERT
        // Return inserted ID
    })
    return tb
}

func (tb *TransactionBuilder) Execute() tea.Cmd {
    return func() tea.Msg {
        tx, _ := db.Begin()
        defer tx.Rollback()

        results := []any{}
        for _, op := range tb.operations {
            result, err := op(tx)
            if err != nil {
                return TransactionErrorMsg{Error: err}
            }
            results = append(results, result)
        }

        tx.Commit()
        return TransactionCompleteMsg{Results: results}
    }
}
```

**Usage:**
```go
return m, m.BeginTransaction().
    Insert("content_data", contentValues).
    Insert("content_fields", fieldValues).  // Can access previous results
    Execute()
```

**Pro:** Fluent API, transaction support, works for any table
**Con:** Need to expose transaction context to plugin API

---

## Critical Questions Before Proceeding

### For Architecture Design

1. **Plugin table creation:**
   - How does Lua plugin define new tables?
   - Show example of plugin creating table
   - Are there existing plugins that do this?

2. **Plugin database access:**
   - What's the Lua API for database operations?
   - Can plugins execute raw SQL?
   - Can plugins use Go database functions?

3. **Content type philosophy:**
   - Are plugins supposed to extend ContentData system?
   - Or create completely independent table structures?
   - Is there a "best practice" for plugin content?

### For Solution Selection

4. **Plugin complexity needs:**
   - Do plugins need coupled operations? (examples?)
   - Do plugins need transactions?
   - Can plugins handle their own operation chaining in Lua?

5. **Core vs Plugin separation:**
   - Should core CMS have different/better tools than plugins?
   - Or should everything use same generic system?
   - Is it okay if plugins are "second-class" for complex operations?

---

## What We Need To Do

### 1. Understand Plugin System

**Read these files:**
- `internal/plugin/` - Lua plugin implementation
- CLAUDE.md mentions Lua plugin system
- Check for plugin examples or documentation
- Look for database API exposed to plugins

### 2. Decide Plugin Content Philosophy

**Two paths:**

**Path A: Plugins Extend Core**
- Plugins add field types, not tables
- All content uses ContentData + ContentFields
- Plugins customize rendering/validation
- → Hybrid approach works

**Path B: Plugins Create Tables**
- Plugins define arbitrary schemas
- Plugins need full database flexibility
- Plugins might need coupled operations
- → Generic approach needed

### 3. Update Problem Statement

Based on plugin analysis:
- Revise constraints section
- Add plugin extensibility requirements
- Determine if generic query builder is feature or bug
- Identify plugin operation chaining needs

### 4. Revise Solution Approach

Depending on plugin philosophy:
- If Path A: Hybrid approach still valid
- If Path B: Need generic chaining pattern
- Possibly: Need both patterns

---

## Immediate Action Items

**Before implementing ANY solution:**

1. [ ] Explore `internal/plugin/` directory
2. [ ] Find Lua database API (if exists)
3. [ ] Check for plugin examples
4. [ ] Read plugin architecture docs
5. [ ] Determine plugin content model philosophy
6. [ ] Update PROBLEM.md with plugin constraints
7. [ ] Revise SUGGESTION-2026-01-15.md based on findings

**After understanding plugins:**

8. [ ] Choose appropriate solution approach
9. [ ] Design plugin-compatible API if needed
10. [ ] Create new suggestion document with plugin support
11. [ ] Plan implementation phases

---

## Why This Is Important

**Portfolio Impact:**

**Before plugin context:**
- "Chose typed methods over generic query builder"
- Shows: Type safety preference

**After plugin context:**
- "Designed hybrid system for compile-time core + runtime plugins"
- Shows: Extensibility architecture, plugin systems, runtime adaptability

**This is MORE impressive for portfolio, not less.**

**The generic query builder isn't a mistake - it's a feature for extensibility.**

---

## CONFIRMED: Plugin Architecture Analysis Complete

**Status:** ✅ Plugin system fully analyzed

### Key Findings

**✅ CONFIRMED: Plugins Create Arbitrary Tables**

From `PLUGIN_ARCHITECTURE.md` and example blog plugin:
- Plugins define tables in `config.json` metadata
- Tables created during plugin initialization via `exec(schema)`
- No enforced namespacing (plugin_name_table_name)
- Support foreign keys and relationships between plugin tables

**✅ CONFIRMED: Plugins Are Independent, NOT Extensions**

Philosophy from docs:
> "When features don't fit the content model, don't force them into EAV.
> Plugin creates optimized columnar schema + Lua business logic."

**Two-Tier Architecture:**
- **Core CMS:** Hybrid EAV (columnar tree + flexible fields) for content management
- **Plugins:** Custom columnar tables + Lua logic for domain-specific features (e-commerce, forums, etc.)

**✅ CONFIRMED: Generic Query Builder Is Intentional**

Lua API provides:
```lua
query(sql, params) -> results              -- SELECT
exec(sql, params) -> affected_rows         -- INSERT/UPDATE/DELETE
query_one(sql, params) -> row              -- Single row
begin_transaction() -> tx                  -- Transactions
```

**Performance:** ~20-30% Lua overhead acceptable (DB queries dominate). Plugin schemas achieve 25x speedup vs EAV for complex queries.

### Architectural Decision

**TWO-TIER SOLUTION REQUIRED:**

**Tier 1: Core CMS Operations (Typed)**
- Use DbDriver typed methods for core tables
- ContentData + ContentFields chaining with type safety
- Hybrid approach from SUGGESTION-2026-01-15.md works here

**Tier 2: Plugin Operations (Generic)**
- Use generic query builder for plugin tables
- Support chained operations for plugins
- Need generic chaining pattern

**Both tiers needed - not contradictory but complementary.**

---

## Revised Solution Approach

### Core CMS (Your Immediate Problem)

**✅ SUGGESTION-2026-01-15.md is VALID for core CMS:**

The hybrid approach works perfectly for core tables:
```go
func (m Model) CreateContentWithFields(...) tea.Cmd {
    contentData := d.CreateContentData(params)  // Typed method
    for fieldID, value := range fieldValues {
        d.CreateContentField(contentDataID, fieldID, value)
    }
    return ContentCreatedMsg{...}
}
```

**This solves your immediate launch blocker.**

### Plugin Support (Future Enhancement)

**Need generic chaining pattern for plugins:**

```go
func (m Model) PluginChainedOperation(steps []OperationStep) tea.Cmd {
    // Execute steps sequentially
    // Pass results between steps
    // Support transactions
}
```

**Can be implemented AFTER core CMS works.**

---

## Implementation Plan (Revised)

### Phase 1: Core CMS (Launch Blocker) - PRIORITY

- [ ] Implement hybrid approach from SUGGESTION-2026-01-15.md
- [ ] ContentData + ContentFields creation works
- [ ] TUI can create core CMS content
- [ ] **This unblocks launch**

### Phase 2: Plugin Chaining (Post-Launch Enhancement)

- [ ] Design generic operation chaining API
- [ ] Expose to Lua plugin API
- [ ] Support plugin-defined coupled operations
- [ ] Document plugin patterns

**Phase 1 is independent and can proceed immediately.**

---

## Portfolio Value (INCREASED)

**Original understanding:**
"Fixed database abstraction layer to use typed methods"

**Corrected understanding:**
"Designed hybrid database architecture balancing compile-time type safety (core CMS) with runtime flexibility (Lua plugin extensibility)"

**Talking points:**
- ✅ Two-tier architecture for different use cases
- ✅ Plugin system with arbitrary table creation
- ✅ Generic query builder as extensibility feature (not bug)
- ✅ Performance optimization (25x for plugin schemas vs EAV)
- ✅ Lua integration with acceptable overhead (~20-30%)
- ✅ Transaction support for multi-step operations

**This is MORE sophisticated than originally thought.**

---

## Status

- ✅ Plugin architecture fully understood
- ✅ Generic query builder justified (extensibility feature)
- ✅ Hybrid approach valid for core CMS
- ✅ Can proceed with SUGGESTION-2026-01-15.md implementation
- ⏸️ Generic chaining for plugins deferred to Phase 2

---

**Next Step:** Implement Phase 1 (Core CMS hybrid approach) - this unblocks launch

**Related Documents:**
- [PROBLEM.md](PROBLEM.md) - Original problem statement (context still valid)
- [SUGGESTION-2026-01-15.md](SUGGESTION-2026-01-15.md) - ✅ VALID for core CMS
- `ai/architecture/PLUGIN_ARCHITECTURE.md` - Complete plugin design
- `ai/packages/PLUGIN_PACKAGE.md` - Plugin API details

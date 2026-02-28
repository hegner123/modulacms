# What Is Bad

Honest problems. Not nitpicks - these are things that will cause real friction as the project grows.

## The 315-Method Interface

`DbDriver` in `internal/db/db.go` has 315 methods. This is the single biggest structural problem. It means:

- **Untestable.** You cannot write a mock that implements 315 methods. Every test that needs any database access must satisfy the entire interface or use a real database.
- **High coupling.** Every new entity adds ~10 methods to the interface and requires implementation on all three database wrappers. The cost of adding a table is not writing the SQL - it's implementing 30+ lines of interface methods across 4 files.
- **Cognitive load.** A 3,199-line file defines the contract for the entire data layer. No developer can hold this in their head.

This should have been decomposed years ago into focused interfaces (`ContentRepository`, `UserRepository`, `MediaRepository`, etc.) that compose into a single driver struct.

## Hand-Written Wrapper Duplication

The sqlc-generated code is correct and maintained by tooling. The wrapper layer between sqlc types and application types is not. Every entity has three mapper functions (SQLite, MySQL, PostgreSQL) that are nearly identical, differing only in int32 vs int64 field widths and occasional NULL handling. This accounts for ~5,000+ lines of mechanical, hand-written code that must be manually updated when schemas change.

A code generator for the wrapper layer would save significant maintenance effort and eliminate a class of bugs where one backend's mapper gets updated but the others don't.

## TUI Global State Variables

15+ package-level variables hold dialog context:
```go
var deleteContentContext *DeleteContentContext
var deleteFieldContext *DeleteFieldContext
// ... 13 more
```

This directly contradicts the Elm architecture the TUI claims to follow. All state should live in the Model. These globals:
- Cannot be tested
- Could be overwritten if two operations overlap
- Create invisible coupling between update handlers
- Make it impossible to reason about state by looking at the Model

## Admin Panel Has Zero Tests

12,900 lines of TypeScript with complex state management (block editor, auth proxy, media uploads) and zero test coverage. The block editor alone manages local edits, dirty tracking, per-block saves, and drag-drop reordering - all untested.

The Go server has ~99,000 lines of tests. The TypeScript SDKs have 200+ tests. The admin panel has none.

## Missing Database Indexes

`content_fields.content_data_id` and `content_fields.field_id` have no indexes despite being frequently queried. On a CMS with thousands of content entries, these queries will degrade to full table scans. This is a performance time bomb.

## Inconsistent Error Handling in TUI

Three different patterns observed in the same codebase:
1. `return FetchErrMsg{Error: err}` - structured error message
2. `return LogMessageCmd(err.Error())` - log and continue
3. `return ResultMsg{Data: []db.Routes{}}` - silently return empty data on error

There's no standard for how errors are displayed to the TUI user, whether operations are retried, or how failures are logged. A user deleting content might get a silent failure with no feedback.

## No Query Filtering in SDKs

All three SDKs support `List()` (get everything) and `Get(id)` (get one), but none support filtering, sorting, or searching. Frontend developers who need "list published content sorted by date" must either:
- Use `RawList()` with manual query parameter construction
- Fetch all items and filter client-side
- Bypass the SDK and make raw HTTP calls

This is the most impactful missing feature for SDK consumers.

## serve.go Does Too Much

530 lines that orchestrate HTTP server, HTTPS server with autocert, SSH server with Bubbletea, plugin initialization, email service, permission cache, config watching, and database setup. Adding any new server-level feature means modifying this increasingly complex function.

## NULL Author IDs

`content_data.author_id` and `datatypes.author_id` are nullable in the schema despite all content logically requiring a creator. This forces defensive NULL checks throughout the application code and SDK type definitions for a field that should never be NULL in practice.

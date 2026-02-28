# TUI (Terminal User Interface) Review

## Overview

62 Go files, ~22,700 lines. Built with Bubbletea (Elm Architecture). Provides SSH-accessible content management with 30+ page types, a three-panel CMS layout, dialog system, form system, and plugin management.

## What Solves a Real Problem

**SSH-based administration** is genuinely valuable. You can manage your CMS from any terminal, over SSH, with no browser needed. For server administrators, DevOps workflows, and headless environments, this is a differentiator that no other CMS offers. The Charmbracelet ecosystem (Bubbletea, Wish, Lipgloss) is the right choice for this.

**The three-panel CMS layout** (content tree | datatype/field list | details) is well-designed for terminal navigation. Content tree browsing with keyboard shortcuts, in-place reordering, and inline actions makes common operations fast.

**200+ message types** provide compile-time type safety for the entire state machine. Every state transition is explicit and traceable. This is the Elm architecture done right.

## What Is Good

**Message system** is excellent. Request messages trigger async work, result messages carry data back, state mutation messages update the model. Clear naming conventions: `*Msg` suffix, `*Cmd` suffix for command constructors.

**FieldBubble interface** enables extensible input types: text, email, URL, slug, textarea, number, boolean, select. The interface is clean (Update, View, Value, SetValue, Focus, Blur, Focused, SetWidth) and new field types can be added without touching core TUI code.

**Navigation with history** preserves cursor position and page context when navigating back. Each page has its own cursor state.

**Database integration** transparently supports all three backends through the DbDriver interface. Audit context is created from CLI session metadata.

**Plugin management** from the TUI: enable/disable/reload plugins, approve routes and hooks, view circuit breaker status.

## What Is Bad

### Global Context Variables (Critical)
15+ package-level variables store dialog context:
```go
var deleteContentContext *DeleteContentContext
var deleteFieldContext *DeleteFieldContext
var deleteDatatypeContext *DeleteDatatypeContext
// ... 12 more
```
These violate the Elm architecture's central principle: all state should live in the Model. They create race conditions if two dialogs could theoretically overlap, make testing impossible, and create invisible coupling between update handlers.

**Fix:** Move all context into the Model struct or pass through message fields.

### Model Struct Bloat
361 lines, 50+ fields mixing:
- UI state (Cursor, FocusIndex, Height, Width)
- Page data (Routes, Datatypes, MediaList, UsersList)
- Dialog state (Dialog, FormDialog, ContentFormDialog)
- Tree state (Root, PendingCursorContentID)
- Plugin state (PluginsList, SelectedPlugin)
- Config state (ConfigCategory, ConfigFieldCursor)
- Provisioning state (NeedsProvisioning, SSHFingerprint)

All of these in one flat struct with no grouping. It's hard to know which fields go together or which page uses which fields.

**Fix:** Break into sub-models: `CMSState`, `DialogState`, `PluginState`, `ConfigState`, etc.

### File Size Problems
- `update_dialog.go`: 2,893 lines - mixes 30+ dialog types and their context management
- `commands.go`: 2,378 lines - database operation commands
- `form_dialog.go`: 2,302 lines - FormDialogModel with deeply nested switch statements
- `update_controls.go`: 1,800 lines - 22 control handler functions in one file

Each of these should be split into smaller files by page or dialog type.

### Inconsistent Error Handling
Some handlers return `FetchErrMsg{Error: err}`, others return `LogMessageCmd(err.Error())`, others silently return empty data instead of errors. No consistent strategy for displaying errors to the user or retrying failed operations.

### Dead Code
- `DEVELOPMENT` page exists but has minimal functionality
- `DYNAMICPAGE` has a template but no real implementation
- Several TODO comments for unimplemented features

## What Is Extra

**Raw database CRUD pages** (CREATEPAGE, READPAGE, UPDATEPAGE, DELETEPAGE) allow direct table manipulation. These are developer tools, not user-facing features. With the admin panel now providing a proper UI for all these operations, these pages serve little purpose for end users. They could be useful for debugging but should be clearly marked as development-only.

## The Bubble Duplication Problem
`bubble_text.go`, `bubble_email.go`, `bubble_url.go`, `bubble_slug.go` are each 57 lines following the exact same pattern. A shared base struct with composition would eliminate this repetition.

## Recommendations

1. **Eliminate global context variables** - move to Model or message fields
2. **Split Model into sub-models** by concern area
3. **Break large files** - max 500 lines per file
4. **Standardize error handling** across all handlers
5. **Extract shared FieldBubble base** to reduce duplication
6. **Remove or gate dead code** (DEVELOPMENT, DYNAMICPAGE)
7. **Add context.Context with timeout** to all async database operations

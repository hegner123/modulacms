# Block Picker Spec

## Overview

Inverted fuzzy-finder block picker for the content editor. Telescope/fzf style — results on top, search input on bottom. Launches on Enter key with the currently active/selected tree node as the insert point. New block inserts as a child of that node.

## Trigger

- **Key:** Enter (when a tree node is focused/active in the drag-and-drop structure panel)
- **Also:** `+` button at each nesting point in the tree (click opens the same picker, insert point = that node)
- **Dismiss:** Escape, click-outside, or successful insert

## Layout

```
┌─────────────────────────────┐
│ {Root Type Name}            │  ← category header (current root datatype)
│   Hero Section              │  ← selectable block
│     Hero Image              │  ← child block, indented
│     Hero CTA                │
│   Content Block             │
│   Two Column Layout         │
│ Collections                 │  ← category header
│   Blog Post Card            │
│     Post Thumbnail          │
│ Global                      │  ← category header
│   Header                    │
│   Footer                    │
├─────────────────────────────┤
│ > _                         │  ← search input, auto-focused
└─────────────────────────────┘
```

- Search input at bottom, always focused on open
- Block list above, scrollable
- Category headers are non-selectable labels
- Arrow keys navigate the selectable items (skip headers)
- Enter on a highlighted block inserts it
- List anchored to bottom of panel, grows upward

## Search Behavior

- Substring match on block name (case-insensitive)
- Search spans all sections — results are flat when filtered, no category headers shown for empty categories
- Category headers disappear entirely when they have zero matching children
- Empty search = full categorized list

## Category Order and Content

1. **Current root datatype name** (e.g., "Page", "Product", "Landing Page")
   - Direct children of the current root datatype (blocks with `parent_id` = current root datatype ID)
   - Their children nested/indented beneath them
   - Full depth shown

2. **Collections**
   - Children of `_collection` type datatypes
   - Their children nested beneath

3. **Global**
   - `_global` datatypes and their children
   - Their children nested beneath

## Insert Behavior

- Selected block is inserted as a child of the active tree node (the node that was focused when picker launched)
- No validation on insert — any block can go anywhere
- Parent_id on datatypes is organizational (picker grouping only), not enforced at content level
- After insert, picker closes, new node is selected in the tree, field editor loads on the right

## Keyboard

| Key | Action |
|-----|--------|
| Enter (tree focused) | Open picker |
| Arrow Up/Down | Navigate selectable blocks (skip headers) |
| Enter (picker open) | Insert highlighted block |
| Escape | Close picker without inserting |
| Any printable char | Appends to search, filters list |
| Backspace | Removes from search, widens list |

## Visual

- No animation needed
- Highlight current selection with background color
- Indentation for nested blocks (padding-left per depth level)
- Category headers styled as de-emphasized labels (smaller, lighter, uppercase or semibold)
- Search input has a `>` prompt character (non-editable, visual only)
- Picker width matches the tree panel width
- Picker appears as an overlay/popover anchored to the bottom of the tree panel, or inline at the bottom

## Data Source

The block list comes from datatypes in the database. The content editor already loads in the context of a specific root datatype, so the current root type ID is known. Query needs:

- All datatypes where `parent_id` = current root datatype ID (section 1)
- All datatypes where parent is a `_collection` type (section 2)
- All `_global` datatypes and their children (section 3)
- Recursive children for nesting display

This can be a single HTMX endpoint that returns the picker HTML, or pre-loaded with the editor and filtered client-side.

## Non-Goals

- No drag-from-picker (click/enter to insert only)
- No block previews or descriptions (name is enough for developer audience)
- No insert validation or type enforcement (future opt-in feature)
- No recents/favorites (can add later if needed)
- No fancy transitions

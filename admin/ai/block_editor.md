───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
 Ready to code?

 Here is Claude's plan:
╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌╌
 Block Editor: Gutenberg-Style Content Editor

 Context

 The current content editor (/content/$title) uses a two-panel layout: a tree sidebar on the left and a form panel on the right. The user wants a Gutenberg-style block editor where content
 nodes render as visual blocks in a centered canvas, with inline editing, block insertion between blocks, and a content-first layout.

 Data Model Recap

 The tree API returns ContentTree { root: ContentNode }. Each ContentNode has:
 - datatype.info — schema (label, type)
 - datatype.content — instance (content_data_id, route_id, sibling pointers, status)
 - fields[] — NodeField with info (Field def) and content (ContentField value)
 - nodes[] — child blocks

 The root node = the page. Root's children = the page's content blocks.

 Files to Create

 src/components/block-editor/
   build-merged-fields.ts       -- Extract shared field merge logic
   use-block-editor-state.ts    -- Central state: selection, local edits, save
   block-editor.tsx             -- Top-level layout orchestrator
   block-editor-header.tsx      -- Back button, title, save all, settings toggle
   page-fields.tsx              -- Root node fields (large title input, etc.)
   block-list.tsx               -- Renders ordered child blocks + inserters
   block-card.tsx               -- Single block: card + toolbar + inline fields
   block-toolbar.tsx            -- Per-block toolbar (type badge, move, delete, save)
   block-inserter.tsx           -- "+" button between blocks, popover to pick datatype
   block-field-editor.tsx       -- Renders FieldRenderer for each field in a block
   document-settings.tsx        -- Optional right sidebar (status, dates, metadata)

 Files to Modify

 - src/routes/_admin/content/$title.tsx — Replace two-panel layout with <BlockEditor>

 Existing Code to Reuse

 - src/components/fields/field-renderer.tsx — FieldRenderer + FieldComponentProps type. 14 field type components already exist. Use this instead of the duplicate renderInput in
 node-editor.tsx.
 - src/queries/content.ts — useTree, useCreateContentData, useCreateContentField, useUpdateContentField, useDeleteContentData, useDeleteContentField
 - src/queries/datatypes.ts — useDatatypes, useDatatypeFieldsByDatatype
 - src/queries/fields.ts — useFields
 - src/components/shared/confirm-dialog.tsx — ConfirmDialog for block deletion
 - src/components/shared/empty-state.tsx — EmptyState
 - src/lib/auth.tsx — useAuthContext

 Component Design

 Layout (block-editor.tsx)

 ┌──────────────────────────────────────────────────────┐
 │  ← Back   Page Title   slug    [3 unsaved] [Save] ⚙ │  ← header
 ├───────────────────────────────────┬───────────────────┤
 │                                   │ Document Settings │
 │   ┌─────────────────────────┐     │ Status: draft     │
 │   │ [Page Title___________] │     │ Created: ...      │
 │   │ [Description__________] │     │ Modified: ...     │
 │   ├─────────────────────────┤     │ Blocks: 3         │
 │   │         ── + ──         │     │                   │
 │   ├─────────────────────────┤     │                   │
 │   │ Hero          ▲ ▼ 🗑    │     │
 │   │ Heading: [___________]  │     │                   │
 │   │ Subtitle: [__________]  │     │                   │
 │   ├─────────────────────────┤     │                   │
 │   │         ── + ──         │     │                   │
 │   ├─────────────────────────┤     │                   │
 │   │ Text Section   ▲ ▼ 🗑   │     │
 │   │ "Lorem ipsum dolor..." │     │                   │
 │   ├─────────────────────────┤     │                   │
 │   │         ── + ──         │     │                   │
 │   └─────────────────────────┘     │                   │
 │                                   │                   │
 └───────────────────────────────────┴───────────────────┘

 - Main content area: max-w-3xl mx-auto, centered canvas
 - Settings sidebar: w-80 border-l, toggled by gear icon
 - Blocks: stacked vertically, inserter lines between them

 State (use-block-editor-state.ts)

 Uses the local-edits-over-server-values pattern (no useEffect setState — avoids infinite loop from unstable useDatatypeFieldsByDatatype array refs):

 // State shape
 selectedBlockId: string | null
 localEdits: Record<string, Record<string, string>>  // contentDataId -> fieldId -> value
 saving: boolean

 // Key functions
 getFieldValue(contentDataId, fieldId, serverValue) → localEdits override or serverValue
 setFieldValue(contentDataId, fieldId, value) → updates localEdits
 isBlockDirty(contentDataId) → checks if any localEdits for that block
 isDirty → any localEdits at all
 saveBlock(contentDataId, mergedFields) → create/update changed fields via mutations, then clear only that block's localEdits on success
 saveAll() → save all dirty blocks (each clears its own localEdits on success)
 clearBlockEdits(contentDataId) → removes localEdits[contentDataId]

 Edit preservation: localEdits are cleared per-block in saveBlock's onSuccess, NOT on tree reference change.
 This ensures saving block A does not wipe unsaved edits on blocks B and C when the tree refetch occurs.

 Block Card (block-card.tsx)

 Unselected state: Collapsed preview — shows first text field value as a summary line. Transparent border. Toolbar appears on hover.

 Selected state: Full inline editing — all fields rendered via FieldRenderer. Blue border with ring. Toolbar always visible.

 Click outside any block → deselects.

 Block Toolbar (block-toolbar.tsx)

 Sits inside the card as a header bar: bg-muted/50 border-b

 Contains: datatype label badge | spacer | [Save] (if dirty) | ▲ ▼ | 🗑

 Block Inserter (block-inserter.tsx)

 Between every pair of blocks and at the end:
 - Horizontal line + centered "+" circle button
 - Both hidden by default, fade in on hover (group-hover:opacity-100)
 - Click opens Popover listing non-ROOT datatypes
 - Selecting a datatype calls useCreateContentData with the route's root as parent

 Page Fields (page-fields.tsx)

 Root node's fields rendered at the top:
 - "Title"-like text fields: large borderless input (text-3xl font-bold border-0)
 - Other fields: standard FieldRenderer with labels

 Field Merge (build-merged-fields.ts)

 Extract from node-editor.tsx into a pure function:

 function buildMergedFields(
   node: ContentNode,
   datatypeFields: DatatypeField[] | undefined,
   allFields: Field[] | undefined,
 ): MergedField[]

 Same logic: walk schema-assigned fields in sort order, merge with content field values, include orphaned fields. Extended to carry validation, ui_config, data from Field definitions (needed
  by FieldRenderer).

 Block Operations

 Insert: useCreateContentData with parent_id = root node ID, datatype_id from popover selection. Tree invalidates and re-renders.

 Delete: ConfirmDialog → delete all content fields for the block → delete the content data node. Clear localEdits for that block.

 Reorder (move up/down): Requires updating sibling pointers on adjacent nodes. Needs a useUpdateContentData mutation (not yet in queries — add it mirroring useUpdateAdminContentData
 pattern). Swap prev/next sibling IDs on the moved node and its neighbors.

 Implementation Order

 1. build-merged-fields.ts — Extract shared utility
 2. use-block-editor-state.ts — Central state hook
 3. block-editor-header.tsx — Header bar
 4. page-fields.tsx — Root node field editing
 5. block-toolbar.tsx — Per-block toolbar
 6. block-field-editor.tsx — Field editing using FieldRenderer
 7. block-card.tsx — Block wrapper (card + toolbar + fields)
 8. block-inserter.tsx — "+" inserter with popover
 9. block-list.tsx — Renders blocks + inserters
 10. document-settings.tsx — Optional sidebar
 11. block-editor.tsx — Top-level orchestrator
 12. Modify $title.tsx — Wire in BlockEditor, remove old tree/panel layout
 13. Add useUpdateContentData to src/queries/content.ts for reordering

 Verification

 1. npm run build — TypeScript + Vite build passes
 2. npm run dev — Navigate to /content, open a route
 3. Verify: root fields render at top, child blocks render as cards
 4. Verify: clicking a block selects it, shows inline editing fields
 5. Verify: clicking outside deselects
 6. Verify: edit a field value, "Save" button enables, click save, value persists after refresh
 7. Verify: "+" inserter appears between blocks on hover, clicking opens datatype popover
 8. Verify: inserting a new block adds it to the list
 9. Verify: deleting a block shows confirmation, then removes it
 10. Verify: move up/down reorders blocks
 11. Verify: "Save All" saves all dirty blocks at once
 12. Verify: settings sidebar toggles open/closed

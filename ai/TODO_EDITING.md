# CMS Editing Tasks

Common editing tasks that need to be implemented for the TUI content editor.

## Content Operations

### CRUD Operations
- [x] **Create content** - Dialog-based content creation with dynamic fields
- [x] **Read/View content** - Tree view with content preview panel
- [x] **Update content fields** - Edit field values for existing content
- [x] **Delete content** - Remove content node (with child handling options)

### Content Navigation
- [x] **Expand/Collapse tree nodes** - Toggle node expansion with +/- keys
- [x] **Navigate to parent** - Jump to parent node (press `g`)
- [x] **Navigate to children** - Jump to first child (press `G`)

### Content Reorganization
- [x] **Move content** - Reparent a node to different parent (press `m`)
- [x] **Reorder siblings** - Move up/down within same parent (Shift+Up/K, Shift+Down/J) with cursor tracking
- [x] **Copy content** - Duplicate a node with fields
- [ ] **Cut/Paste content** - Move via clipboard

### Content Status
- [x] **Publish/Unpublish** - Toggle content status (draft/published)
- [x] **Draft/Archive** - Toggle archive status (press `a`), `~` tree indicator
- [ ] **Schedule publishing** - Time-based status changes

## Datatypes Page

### Implemented
- [x] **Create datatype** - Dialog-based creation with parent selector
- [x] **Edit datatype** - Edit label, type, parent via modal dialog
- [x] **Create field** - Dialog-based field creation (content panel)
- [x] **Edit field** - Edit label and type via modal dialog (content panel)

### Implemented
- [x] **Delete datatype** - Remove datatype from tree panel (with child-check protection, junction cleanup)
- [x] **Delete field** - Remove field from content panel (with junction record deletion)

## Routes Page

### Implemented
- [x] **Create route** - Dialog-based route creation
- [x] **Edit route** - Edit title and slug via modal dialog

### Implemented
- [x] **Delete route** - Remove route via confirmation dialog

## Media Page

### Implemented
- [x] **Browse media** - List and navigate media items

### Implemented
- [x] **Delete media** - Remove media item via confirmation dialog

### Implemented
- [x] **Upload media** - File picker overlay with async upload pipeline

## Users Admin Page

### Implemented
- [x] **Create user** - `UserFormDialogModel` with username, name, email, role fields
- [x] **Edit user** - Edit via modal dialog (preserves password hash)
- [x] **Delete user** - Delete via confirmation dialog

## Field Operations
- [x] **Edit single field** - Inline field editing (press `e` on right panel)
- [x] **Add field to content** - Add new field value (press `n` on right panel)
- [x] **Remove field from content** - Delete field value (press `d` on right panel)
- [x] **Reorder fields** - Change field display order (Shift+Up/Down on right panel)

## Bulk Operations
- [ ] **Bulk delete** - Delete multiple selected items
- [ ] **Bulk status change** - Publish/unpublish multiple items
- [ ] **Bulk move** - Move multiple items to new parent

## Media Integration
- [ ] **Attach media to content** - Link media to content fields
- [ ] **Upload media inline** - Upload while editing content

## Version Control
- [ ] **View revision history** - See previous versions
- [ ] **Restore previous version** - Rollback changes
- [ ] **Compare versions** - Diff between revisions

## Search & Filter
- [ ] **Search content** - Find content by text
- [ ] **Filter by datatype** - Show only specific types
- [ ] **Filter by status** - Show draft/published/archived
- [ ] **Filter by date** - Date range filtering

## UI/UX Improvements
- [x] **Controls guide** - Context-sensitive keybinding hints in status bar
- [x] **Quit confirmation** - Confirm before quitting on content pages

---

## Implementation Priority

### Phase 1 - Core Editing (complete)
1. ~~Edit content fields (update existing)~~ done
2. ~~Delete content~~ done
3. ~~Expand/Collapse tree nodes~~ done (+/= expand, -/_ collapse)

### Phase 1.5 - Missing Action Handlers (complete)
4. ~~Delete datatype (tree panel)~~ done
5. ~~Delete field (content panel)~~ done
6. ~~Delete route~~ done
7. ~~Upload media~~ done (file picker overlay + async pipeline)
8. ~~Delete media~~ done
9. ~~Users admin CRUD~~ done (full create/edit/delete with UserFormDialogModel)

### Phase 2 - Reorganization (complete)
10. ~~Move content (reparent)~~ done (press `m`, target selector)
11. ~~Reorder siblings~~ done (Shift+Up/K, Shift+Down/J, tree reorder by pointers)
12. ~~Copy content~~ done (handler + field duplication)

### Phase 3 - Status Management (complete)
13. ~~Publish/Unpublish~~ done (toggle draft/published)
14. ~~Draft/Archive~~ done (toggle archived/draft with `a` key, `~` tree indicator)

### Phase 4 - Advanced Features
15. Version history (HQ step 10)
16. Search & filter (HQ step 11)
17. Bulk operations (HQ step 12, depends on 5+8+11)

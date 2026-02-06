# CMS Editing Tasks

Common editing tasks that need to be implemented for the TUI content editor.

## Content Operations

### CRUD Operations
- [x] **Create content** - Dialog-based content creation with dynamic fields
- [x] **Read/View content** - Tree view with content preview panel
- [x] **Update content fields** - Edit field values for existing content
- [x] **Delete content** - Remove content node (with child handling options)

### Content Navigation
- [ ] **Expand/Collapse tree nodes** - Toggle node expansion with +/- keys (status bar hints +/- but only enter works)
- [ ] **Navigate to parent** - Jump to parent node
- [ ] **Navigate to children** - Jump to first child

### Content Reorganization
- [ ] **Move content** - Reparent a node to different parent
- [ ] **Reorder siblings** - Change sibling order (move up/down)
- [ ] **Copy content** - Duplicate a node (with/without children)
- [ ] **Cut/Paste content** - Move via clipboard

### Content Status
- [ ] **Publish/Unpublish** - Change content visibility status
- [ ] **Draft/Archive** - Lifecycle state transitions
- [ ] **Schedule publishing** - Time-based status changes

## Datatypes Page

### Implemented
- [x] **Create datatype** - Dialog-based creation with parent selector
- [x] **Edit datatype** - Edit label, type, parent via modal dialog
- [x] **Create field** - Dialog-based field creation (content panel)
- [x] **Edit field** - Edit label and type via modal dialog (content panel)

### Missing (hinted in status bar but not implemented)
- [ ] **Delete datatype** - Remove datatype from tree panel (`ActionDelete` handler missing in `DatatypesControls`)
- [ ] **Delete field** - Remove field from content panel (`ActionDelete` handler missing in `DatatypesControls`)

## Routes Page

### Implemented
- [x] **Create route** - Dialog-based route creation
- [x] **Edit route** - Edit title and slug via modal dialog

### Missing (hinted in status bar but not implemented)
- [ ] **Delete route** - Remove route (`ActionDelete` handler missing in `RoutesControls`)

## Media Page

### Implemented
- [x] **Browse media** - List and navigate media items

### Missing (hinted in status bar but not implemented)
- [ ] **Upload media** - Upload new media (`ActionNew` handler missing in `MediaControls`)
- [ ] **Delete media** - Remove media item (`ActionDelete` handler missing in `MediaControls`)

## Users Admin Page

### Missing (all CRUD hinted in status bar but uses generic `BasicCMSControls`)
- [ ] **Create user** - No `UsersAdminControls` function exists
- [ ] **Edit user** - No `UsersAdminControls` function exists
- [ ] **Delete user** - No `UsersAdminControls` function exists

## Field Operations
- [ ] **Edit single field** - Inline field editing
- [ ] **Add field to content** - Add new field value
- [ ] **Remove field from content** - Delete field value
- [ ] **Reorder fields** - Change field display order

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

### Phase 1 - Core Editing (current)
1. ~~Edit content fields (update existing)~~ done
2. ~~Delete content~~ done
3. Expand/Collapse tree nodes (add +/- key handlers)

### Phase 1.5 - Missing Action Handlers
Hinted in status bar but not wired up:
4. Delete datatype (tree panel)
5. Delete field (content panel)
6. Delete route
7. Upload media
8. Delete media
9. Users admin CRUD (needs dedicated controls function)

### Phase 2 - Reorganization
10. Move content (reparent)
11. Reorder siblings
12. Copy content

### Phase 3 - Status Management
13. Publish/Unpublish
14. Draft/Archive states

### Phase 4 - Advanced Features
15. Version history
16. Search & filter
17. Bulk operations

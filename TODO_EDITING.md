# CMS Editing Tasks

Common editing tasks that need to be implemented for the TUI content editor.

## Content Operations

### CRUD Operations
- [x] **Create content** - Dialog-based content creation with dynamic fields
- [x] **Read/View content** - Tree view with content preview panel
- [x] **Update content fields** - Edit field values for existing content
- [x] **Delete content** - Remove content node (with child handling options)

### Content Navigation
- [ ] **Expand/Collapse tree nodes** - Toggle node expansion with +/- keys
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

### Phase 1 - Core Editing
1. Edit content fields (update existing)
2. Delete content
3. Expand/Collapse tree nodes

### Phase 2 - Reorganization
4. Move content (reparent)
5. Reorder siblings
6. Copy content

### Phase 3 - Status Management
7. Publish/Unpublish
8. Draft/Archive states

### Phase 4 - Advanced Features
9. Version history
10. Search & filter
11. Bulk operations

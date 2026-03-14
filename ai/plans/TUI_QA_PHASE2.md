# TUI QA Phase 2: Content System

Covers: Content screen (Select phase, Tree phase, Version phase), all content operations (create, edit, delete, publish, unpublish, copy, move, reorder, versions), Admin content parity.

## Prerequisites

- DB Wipe & Redeploy completed
- Modula Default schema installed
- At least one route with content exists (bootstrap Home page at `/`)

## 2.1 Content Screen — Select Phase (Route List)

### 2.1.1 Content list renders
- Navigate to Content, wait for `Content  [postgres]`
- Verify left panel shows "Content" header with route-grouped items
- Verify `-- Pages --` section header
- Verify at least `/ ○ [Page]` (Home page from bootstrap)
- Verify Details panel shows: Route, Title, Datatype, Status, Author, ID, Created, Modified
- Verify Stats panel shows: Total, Pages, Standalone counts
- Verify key hints bar: up/down, enter, +/-, e, n, d, p, h, q
- Snapshot: `goldens/2.1.1_content_list.txt`

### 2.1.2 Create new page via route list
- Press `n` → verify "New Content" form dialog
- Verify fields: Title (text input), Slug (slug input), Datatype (selector showing "Page")
- Type title "About Us", tab to slug, type "about-us"
- Navigate to Confirm, press enter
- Verify creation (may need exit/re-enter Content to see refresh)
- Re-enter Content → verify `/about-us ○ [Page]` appears in list
- Verify Stats total incremented

### 2.1.3 Edit page from route list
- Select a page, press `e`
- Verify edit form dialog with pre-populated fields
- Modify title, confirm
- Verify Details panel shows updated title

### 2.1.4 Delete page from route list
- Create a throwaway page first
- Select it, press `d`
- Verify confirmation dialog with delete warning
- Confirm → verify page removed from list
- Also test Cancel path: verify page preserved

### 2.1.5 Publish from route list
- Select draft page (○ indicator), press `p`
- Verify publish confirmation dialog
- Confirm → verify status changes to published (● indicator)
- Press `p` again → verify unpublish dialog
- Confirm unpublish → verify back to draft (○)

### 2.1.6 Expand/Collapse route groups
- Press `+` or `=` → verify route group expands (if applicable)
- Press `-` or `_` → verify collapse

## 2.2 Content Screen — Tree Phase

### 2.2.1 Enter tree view
- Select Home page (`/ ○ [Page]`), press `enter`
- Verify transition to Tree phase
- Verify Tree panel shows: `├─ ○ Page` (root node)
- Verify Preview panel shows: `Page ════════` with field values
- Verify key hints bar changes: includes n, e, d, c, m, p, shift+up/shift+down

### 2.2.2 Child type selector shows all types
- On Page root node, press `n`
- Verify "Select Child Type" dialog
- Verify ALL expected types present (scroll through list):
  - Layout: Row, Columns, Grid, Area, Settings
  - Content: CTA, Rich Text, Text, Image, Button, Card, Animation
  - Menu: Menu, Menu Link, Menu Icon Link, Menu List, Menu Nested List, etc.
- Press `escape` to cancel

### 2.2.3 Create Row
- On Page root, press `n` → select Row (first in list)
- Verify form: "Full Width" boolean field
- Submit → verify "Content created with 1 fields"
- Dismiss → verify Row appears in tree: `▼ ○ Page` → `├─ ○ Row`
- Verify Preview shows Row with "Full Width: false"

### 2.2.4 Create CTA under Row
- Navigate to Row node, press `n`
- Select CTA from child type list
- Verify form fields: Heading, Subheading (textarea), Button Text, Button URL
- Fill all fields:
  - Heading: "Welcome to ModulaCMS"
  - Subheading: "The flexible headless CMS"
  - Button Text: "Get Started"
  - Button URL: "https://example.com"
- Submit → verify "Content created with 4 fields"
- Verify tree: Row → CTA
- Verify Preview shows all field values

### 2.2.5 Create Rich Text block
- Navigate to appropriate parent, press `n` → select Rich Text
- Verify form: single "Content" textarea field
- Fill content, submit
- Verify in tree and preview

### 2.2.6 Create Text block
- Press `n` → select Text
- Verify form: single "Content" textarea field
- Submit, verify

### 2.2.7 Create Image block
- Press `n` → select Image
- Verify form: Image (media), Alt Text, Caption (textarea)
- Fill alt text and caption (skip Image media field), submit
- Verify in tree and preview

### 2.2.8 Create Button block with select field
- Press `n` → select Button
- Verify form: Label, URL, Variant (select field)
- Navigate to Variant field → verify options: primary, secondary, outline, ghost
- Select an option, fill other fields, submit
- Verify Variant value shows in preview

### 2.2.9 Create Card block
- Press `n` → select Card
- Verify form: Title, Description (textarea), Image (media), Link URL
- Fill all text fields, submit
- Verify in tree and preview

### 2.2.10 Create Grid + Area
- On Page root, press `n` → select Grid
- Verify form: Columns, Rows, Gap (all text fields)
- Fill: "3", "2", "10px", submit
- Navigate to Grid node, press `n` → select Area
- Verify form: Column Start, Column End, Row Start, Row End (all number)
- Fill: 1, 2, 1, 1, submit
- Verify nested structure in tree

### 2.2.11 Create Settings block
- On Page root, press `n` → select Settings
- Verify form: Margin, Padding (text fields)
- Fill, submit, verify

### 2.2.12 Create Animation block
- Press `n` → select Animation
- Verify form: Type (select: fade/slide/scale/rotate), Duration, Delay, Easing (select), Direction (select), Iterations
- Navigate select fields, choose options
- Submit, verify all values in preview

## 2.3 Tree Operations

### 2.3.1 Edit content fields
- Select a node with fields (e.g., CTA), press `e`
- Verify edit form pre-populated with existing values
- Modify one field (e.g., change heading text)
- Submit → verify updated value in Preview panel

### 2.3.2 Edit single field (if supported)
- Select node, verify if single-field edit mode exists
- Document behavior

### 2.3.3 Delete leaf node
- Select a leaf node (no children), press `d`
- Verify confirmation dialog
- Confirm → verify node removed from tree
- Verify sibling pointers updated (adjacent nodes still connected)

### 2.3.4 Delete node with children
- Select a node that has children (e.g., Row with CTA), press `d`
- Verify warning about children in confirmation dialog
- Confirm → verify node AND children removed
- Cancel → verify preserved

### 2.3.5 Reorder siblings — shift+down
- Create 2+ sibling nodes under same parent
- Select first sibling, press `shift+down`
- Verify order swapped in tree
- Verify Preview reflects new order

### 2.3.6 Reorder siblings — shift+up
- Select second sibling, press `shift+up`
- Verify order restored

### 2.3.7 Reorder boundary — top
- Select first sibling, press `shift+up`
- Verify no change (already at top), no crash

### 2.3.8 Reorder boundary — bottom
- Select last sibling, press `shift+down`
- Verify no change, no crash

### 2.3.9 Move content to new parent
- Select a content node, press `m`
- Verify move dialog with list of valid parent targets
- Verify current parent excluded (or self excluded)
- Select new parent, confirm
- Verify node moved in tree under new parent
- Verify old parent's tree updated (node gone)

### 2.3.10 Copy content
- Select a content node with field values, press `c`
- Verify success message
- Verify copy appears as sibling after original
- Verify copy has same field values but different content ID
- Verify tree structure intact

### 2.3.11 Go to parent (g key)
- Navigate to a child node, press `g`
- Verify cursor jumps to parent node

### 2.3.12 Go to child (G key)
- Navigate to a parent node, press `G`
- Verify cursor jumps to first child

### 2.3.13 Expand/Collapse tree nodes
- Select a parent with children, press `-` → verify children hidden
- Press `+` → verify children shown again

### 2.3.14 Back from tree to select phase
- In tree view, press `h`
- Verify return to route list (Select phase), NOT quit dialog

## 2.4 Content Versions

### 2.4.1 Version list
- Publish a content item first (creates a version/snapshot)
- Select the published item, press `v`
- Verify version list overlay appears
- Verify at least one version entry with timestamp

### 2.4.2 Restore version
- In version list, select a version, press `enter`
- Verify confirmation dialog for restore
- Confirm → verify content restored
- Cancel → verify no change

### 2.4.3 Back from version list
- In version list, press `h` or `escape`
- Verify return to normal tree view

## 2.5 Publish/Unpublish in Tree

### 2.5.1 Publish from tree
- Select draft node in tree, press `p`
- Verify publish confirmation dialog
- Confirm → verify status indicator changes (○ → ●)

### 2.5.2 Unpublish from tree
- Select published node, press `p`
- Verify unpublish confirmation dialog
- Confirm → verify status reverts (● → ○)

## 2.6 Admin Content (ctrl+a)

### 2.6.1 Toggle to admin mode
- On Content screen, press `ctrl+a`
- Verify status bar changes from `[Client]` to `[Admin]`
- Verify content list changes (shows admin content, not public)

### 2.6.2 Admin CRUD parity
- In Admin mode: create content, edit, delete, publish
- Verify all operations work identically to Client mode
- Verify data isolation (admin changes don't affect public content)

### 2.6.3 Toggle back to client mode
- Press `ctrl+a` again
- Verify `[Client]` restored, public content shown

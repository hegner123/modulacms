Plan 1: Tree Package — Admin Content TUI Support

Context

The tree package has two layers:
- core (internal/tree/core/) — shared tree-building algorithms, type-agnostic via ID conversion
- tree (internal/tree/) — TUI wrapper with expand/collapse, indent, and visible-node traversal

core.BuildAdminTree() already exists (build.go:207-379). It converts admin types
(AdminContentData, AdminDatatypes, AdminContentFields, AdminFields) into regular types
(ContentData, Datatypes, ContentFields, Fields) via string-based ID casting, then builds a
core.Root with NodeIndex keyed by types.ContentID. This is used by the router/model layer
for JSON API responses.

The TUI tree layer (tree.Root) only has LoadFromRows(*[]db.GetContentTreeByRouteRow) which
calls core.BuildFromRows(). There is no equivalent that calls core.BuildAdminTree().

Goal

Add a LoadFromAdminData method to tree.Root that accepts the split admin slices, delegates to
core.BuildAdminTree, and wraps the result in tree.Node with TUI state — exactly as LoadFromRows
does for regular content.

Prerequisite: None (this is the foundation for all subsequent plans)

---

Implementation

File: internal/tree/tree.go

Add method:

func (page *Root) LoadFromAdminData(
    cd []db.AdminContentData,
    dt []db.AdminDatatypes,
    cf []db.AdminContentFields,
    df []db.AdminFields,
) (*LoadStats, error)

Logic (mirrors LoadFromRows lines 120-178):
1. Call core.BuildAdminTree(cd, dt, cf, df) → coreRoot, coreStats, coreErr
2. Set page.Core = coreRoot
3. Phase 1: Create tree.Node wrapper for each entry in coreRoot.NodeIndex
   - Copy Instance (ContentData), InstanceFields (ContentFields), Datatype, Fields
   - Set Expand=true, CoreNode=reference
   - Add to page.NodeIndex keyed by the same types.ContentID (already converted by core)
4. Phase 2: Re-link pointer references from core space to tree space
   - For each core.Node → tree.Node mapping, rewire Parent, FirstChild, NextSibling, PrevSibling
5. Set page.Root from coreRoot.Node
6. Collect root-level nodes into page.rootNodes
7. Convert core.LoadStats to tree.LoadStats
8. Return stats, coreErr

Note: Since core.BuildAdminTree converts all admin IDs to types.ContentID via .String(),
the resulting NodeIndex uses ContentID keys. This means tree.Root.NodeIndex
(map[types.ContentID]*Node), FlattenVisible, FindVisibleIndex, CountVisible, NodeAtIndex,
InsertNodeByIndex, and DeleteNodeByIndex all work without modification. The admin content
IDs and regular content IDs are the same ULID format — just different branded types in Go.

---

Steps

1. Add LoadFromAdminData method to tree.Root in internal/tree/tree.go
2. Verify: just check
3. Add test in internal/tree/tree_test.go (or tree_admin_test.go):
   - Build admin content slices with known parent/child/sibling relationships
   - Call LoadFromAdminData
   - Assert tree structure, NodeIndex population, FlattenVisible ordering
4. Verify: go test ./internal/tree/...

---

Size estimate: ~60 lines of new code (method body mirrors LoadFromRows).

---

Risk assessment

Low risk. The method is a thin wrapper over existing, tested core.BuildAdminTree. The same
wrapping pattern in LoadFromRows has been stable since the core/tree split. No existing code
is modified — this is purely additive.

The one subtlety: tree.Node.Instance is *db.ContentData, not *db.AdminContentData. After
conversion, the Instance field contains a ContentData struct with IDs that were originally
AdminContentIDs cast to ContentID. The TUI rendering code (DecideNodeName, FormatTreeRow)
accesses Instance.ContentDataID, Instance.Status, and Datatype.Label — all of which are
preserved through the conversion. No admin-specific ID access is needed in the TUI tree
rendering path.

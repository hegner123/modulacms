Plan 5: Legacy Cleanup — Remove Migrated Code

Context

After Plan 4 is complete and manually verified, the legacy content handling code becomes dead
code. This plan removes it.

Prerequisites: Plan 4 complete and manually tested in both CONTENT and ADMINCONTENT modes.

---

Pre-execution Gate

Before starting any removal, verify all three conditions:

1. `screen_content.go` and `screen_content_view.go` exist in `internal/tui/`
2. `screenForPage()` in `screen.go` has cases for both CONTENT and ADMINCONTENT
3. Content operations work end-to-end in both modes via the new Screen path

If any condition fails, stop. This plan cannot proceed without Plan 4.

---

Removals from update_controls.go

- case CONTENT: and case ADMINCONTENT: from PageSpecificMsgHandlers switch (~4 lines)
- ContentBrowserControls method (lines 512-886, ~375 lines)
- contentBrowserCursorUpCmd helper (lines 889-896, ~8 lines)
- contentBrowserCursorDownCmd helper (lines 899-906, ~8 lines)
- contentFieldEdit method (lines 909-922, ~14 lines)
- contentFieldAdd method (~45 lines)
- contentFieldDelete method (~16 lines)
- contentFieldReorderUp method (~20 lines)
- contentFieldReorderDown method (~20 lines)

Total: ~510 lines removed from update_controls.go

---

Removals from admin_controls.go

- AdminContentBrowserControls method (lines 448-501, ~54 lines)

Total: ~54 lines removed from admin_controls.go

---

Removals from panel_view.go

- case CONTENT: and case ADMINCONTENT: from cmsPanelContent (~22 lines)
- case CONTENT: from cmsPanelTitles (~8 lines)
- case CONTENT: from getContextControls (~29 lines, lines 595-623)
- case ADMINCONTENT: from getContextControls (~2 lines, lines 676-677)
- renderRootContentSummaryList function
- renderContentSummaryDetail function
- renderContentActions function
- renderVersionList function

Total: ~250 lines removed from panel_view.go

---

Removals from admin_panel_view.go

- renderAdminContentList function (lines 268-289, ~22 lines)
- renderAdminContentDetail function (lines 291-316, ~26 lines)

Total: ~48 lines removed from admin_panel_view.go

---

Removals from page_builders.go — Full Function Disposition

Every function in page_builders.go and its disposition:

| Function | Type | Disposition | Reason |
|---|---|---|---|
| CMSPage (struct) | type | Remove | Replaced by ContentScreen |
| CMSPage.ProcessTreeDatatypes | method | Remove | Replaced by ContentScreen rendering |
| CMSPage.traverseTree | method | Remove | Internal to CMSPage |
| CMSPage.traverseTreeWithDepth | method | Remove | Internal to CMSPage |
| CMSPage.FormatTreeRow | method | Remove | Internal to CMSPage |
| CMSPage.ProcessContentPreview | method | Remove | Replaced by ContentScreen |
| CMSPage.ProcessFields | method | Remove | Replaced by ContentScreen |
| FormatRow | package-level | Remove | Exported but never called anywhere in codebase (dead code) |
| DecideNodeName | package-level | Move or keep | grep to determine: if only screen_content_view.go uses it, move there; if used elsewhere, keep package-level |
| FieldMatchesLabel | package-level | Follows DecideNodeName | Used only by DecideNodeName — must move/stay with it |
| resolveAuthorName | package-level | Move or remove | Used only by CMSPage.ProcessContentPreview. If Plan 4's ContentScreen reimplements this logic, remove. If ContentScreen calls this function, move with DecideNodeName |

After applying the above, if page_builders.go contains only DecideNodeName, FieldMatchesLabel,
and optionally resolveAuthorName, either:
- Move them to screen_content_view.go and delete page_builders.go, OR
- Rename page_builders.go to something like tree_helpers.go if they serve multiple screens

If page_builders.go becomes empty, delete it.

Total: ~300 lines removed from page_builders.go

---

Orphaned Model Fields

After Plan 4's ContentScreen owns its own state, the following Model fields become dead state
that is no longer read by any rendering or control code. These fields should be identified
during cleanup but removal is deferred to a separate follow-up commit to limit blast radius:

- RootContentSummary
- AdminRootContentSummary
- SelectedContentFields
- VersionCursor
- ShowVersionList
- Versions
- VersionContentID
- VersionRouteID
- PageRouteId (if fully owned by ContentScreen)
- RootDatatypes (if fully owned by ContentScreen)

After code removal in this plan, grep each field name. If no remaining code reads the field
(outside of update_state.go setters), note it for removal. The actual field removal and
corresponding update_state.go message handler cleanup is a separate step to avoid combining
structural deletion with state refactoring.

---

Post-cleanup Verification

After removal, PageSpecificMsgHandlers should only contain:
  case HOMEPAGE:
      return m.BasicControls(msg)
  case CMSPAGE:
      return m.BasicCMSControls(msg)
  case ADMINCMSPAGE:
      return m.BasicCMSControls(msg)

cmsPanelContent should have no CONTENT or ADMINCONTENT cases.
getContextControls should have no CONTENT or ADMINCONTENT cases.

---

Steps

1. Pre-execution gate: verify screen_content.go, screen_content_view.go exist and
   screenForPage has CONTENT/ADMINCONTENT cases
2. Verify DecideNodeName, FieldMatchesLabel, resolveAuthorName usage:
   grep for each across the TUI package to determine disposition per table above
3. Remove legacy control handlers from update_controls.go
4. Remove AdminContentBrowserControls from admin_controls.go
5. Remove legacy rendering functions from panel_view.go
6. Remove admin rendering functions from admin_panel_view.go
7. Remove CMSPage, all methods, and FormatRow from page_builders.go
8. Move or delete DecideNodeName + FieldMatchesLabel + resolveAuthorName per step 2 findings
9. Delete page_builders.go if empty
10. Clean up unused imports in all modified files (go will refuse to compile with unused imports)
11. Verify: just check
12. Verify: just test
13. Manual testing checklist (both CONTENT and ADMINCONTENT modes):
    - Navigate content tree (cursor up/down, expand/collapse nodes)
    - Select content node and view preview/detail panel
    - Create new content node
    - Edit content node fields
    - Delete content node
    - Move/copy content node
    - Reorder content nodes (up/down)
    - Publish/unpublish content
    - View version list and restore a version
    - Switch locale (if i18n enabled)
    - Add/edit/delete/reorder content fields
    - Navigate between panels (tree, content, fields)
    - ADMINCONTENT: verify admin-specific content list and detail views
14. Verify PageSpecificMsgHandlers residual (only HOMEPAGE/CMSPAGE/ADMINCMSPAGE)
15. Verify cmsPanelContent has no CONTENT/ADMINCONTENT cases
16. Grep all removed function names one final time to confirm zero remaining references

---

Size estimate: ~1200 lines removed, 0-20 lines added (moved helpers only).

---

Risk assessment

Low risk if Plan 4 is thoroughly tested. Primary risks:

1. Removing something still referenced by non-obvious code paths.
   Mitigation: just check catches all compile errors. Final grep for each removed function name.

2. Silent behavior loss — a removed message handler causes an operation to silently stop working
   without a compile error.
   Mitigation: The manual testing checklist in step 13 covers all content operations explicitly.

3. Unused imports after large deletions causing compile failures.
   Mitigation: Explicit import cleanup step (step 10) before verification.

---

Phase 8.1 Note

The UNIFIED_TUI_PLAN specifies Phase 8.1 as creating PanelScreen base struct with
RenderPanels and HandlePanelNav. This does not exist yet. ContentScreen (Plan 4) inlines
its panel rendering, consistent with all other migrated screens. If PanelScreen is later
implemented, ContentScreen can be refactored to embed it. This is not a prerequisite and
does not block any of these plans.

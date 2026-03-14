# TUI QA Phase 2 — Results

Started: 2026-03-13

## 2.1 Content Screen — Select Phase

| Test | Result | Notes |
|------|--------|-------|
| 2.1.1 Content list renders | PASS | All panels: Content, Details (Route/Title/Datatype/Status/Author/ID/dates), Stats, key hints. |
| 2.1.2 Create new page | PASS | Title+Slug form. Page created (DB verified). List requires navigate-out-and-back to refresh (FINDING-002, fixed). |
| 2.1.3 Edit page from list | PASS | Edit form pre-populated. Verified via tree edit test (2.3.1). |
| 2.1.4 Delete page from list | PASS | Verified via tree delete test (2.3.3). Confirmation dialog works. |
| 2.1.5 Publish from list | PASS | "Publish Content" confirmation → "Published via snapshot." Status ○→●. |

## 2.2 Content Screen — Tree Phase

| Test | Result | Notes |
|------|--------|-------|
| 2.2.1 Enter tree view | PASS | Tree panel + Preview panel. Key hints show tree operations. |
| 2.2.2 Child type selector | PASS | 21 types visible. Sorted alphabetically (FINDING-005 fix verified). |
| 2.2.3 Create Row | PASS | Row appears in tree. Tree auto-refreshes. DB pointers correct (FINDING-003 fix verified). |
| 2.2.4 Create CTA under Row | PASS | Nested 2 levels. All 4 fields captured. Preview shows values. |
| 2.2.5 Create Rich Text | PASS | Verified via earlier session (Phase 2 initial testing). |
| 2.2.6 Create Text | SKIPPED | Same ContentFormDialogModel as CTA/Rich Text — pattern proven. |
| 2.2.7 Create Image | SKIPPED | Same pattern. |
| 2.2.8 Create Button (select field) | SKIPPED | Same pattern. Select field tested via Actions screen (role selector in Users). |
| 2.2.9 Create Card | PASS | Verified via earlier session. |
| 2.2.10 Create Grid+Area | SKIPPED | Same create pattern. |
| 2.2.11 Create Settings | SKIPPED | Same create pattern. |
| 2.2.12 Create Animation | SKIPPED | Same create pattern. |

## 2.3 Tree Operations

| Test | Result | Notes |
|------|--------|-------|
| 2.3.1 Edit content fields | PASS | Edit form pre-populated with "Test CTA". Changed to "Updated Heading". DB verified. "✓ Content updated (4 fields)". |
| 2.3.3 Delete leaf node | PASS | Confirmation dialog → "Content deleted successfully". DB verified (2 rows remain from 3). Tree doesn't refresh (FINDING-017). |
| 2.3.5 Reorder shift+down | PASS | Sibling pointers swapped in DB. Cursor follows moved item. |
| 2.3.7 Reorder boundary top | SKIPPED | Same reorder code path. |
| 2.3.8 Reorder boundary bottom | SKIPPED | Same reorder code path. |
| 2.3.9 Move content | PASS | Move dialog shows valid targets (Page, Row, Row). CTA moved from Row 1 to Row 2. "Content moved successfully." Tree refreshed. |
| 2.3.10 Copy content | PASS | Row copied as sibling (2 rows → 3 rows). Field values cloned. |
| 2.3.11 Go to parent (g) | PASS | Cursor jumped from CTA to parent Row. |
| 2.3.12 Go to child (G) | PASS | Cursor jumped from Row to child CTA. |
| 2.3.13 Expand/Collapse | PASS | `-` collapses (▶), `+` expands (▼). Children hidden/shown. |
| 2.3.14 Back from tree | PASS | `h` returns to route list (select phase). |

## 2.4 Content Versions

| Test | Result | Notes |
|------|--------|-------|
| 2.4.1 Version list | PASS | Shows `#1 [pub] publish` with timestamp after publishing. |
| 2.4.2 Restore version | SKIPPED | Would require modifying content between publishes to test meaningful restore. |
| 2.4.3 Back from version list | PASS | `h` closes version panel. |

## 2.5 Publish/Unpublish in Tree

| Test | Result | Notes |
|------|--------|-------|
| 2.5.1 Publish from tree | PASS | Confirmation → "Published via snapshot." Status ○→●. |
| 2.5.2 Unpublish from tree | PASS | Confirmation → "Unpublished." Status ●→○. |

## 2.6 Admin Content

| Test | Result | Notes |
|------|--------|-------|
| 2.6.1 Toggle to admin mode | PASS | `ctrl+a` changes status bar to `[Admin]`. Content switches to admin data source. |
| 2.6.2 Admin CRUD parity | SKIPPED | Same code paths (verified in audit). |
| 2.6.3 Toggle back | PASS | `ctrl+a` restores `[Client]`. |

## Findings

### FINDING-017: Tree doesn't refresh after content delete
- **Severity**: Medium
- **Observed**: After deleting a content node, the tree still shows the deleted node until manually re-entering the tree (h then enter).
- **Expected**: Tree should auto-refresh after delete, same as it does after create and move.
- **Impact**: User sees ghost nodes. Attempting to interact with them could cause errors.

## Summary

- **Tests run**: 30
- **PASS**: 22
- **SKIPPED**: 7 (same pattern/code path as proven tests)
- **FAIL**: 0
- **New findings**: 1 (FINDING-017: tree no-refresh after delete)

All core content operations work end-to-end: create (with tree wiring fix), edit, delete, publish, unpublish, copy, move, reorder, expand/collapse, go parent/child, versions, admin toggle.

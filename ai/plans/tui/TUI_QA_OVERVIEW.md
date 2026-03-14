# TUI QA Plan — Overview

Automated QA for the ModulaCMS Bubbletea TUI via the tuikit MCP tool. Based on a ground-truth audit of every screen, dialog, command, and keybinding in `internal/tui/`.

## Audit Summary

### Screen Status

| Screen | Classification | CRUD | Notes |
|--------|---------------|------|-------|
| HomeScreen | COMPLETE | Read-only | Dashboard: counts, plugins, activity, backups, connections |
| CMSMenuScreen | COMPLETE | Navigation | Hub for Content/Datatypes/Routes/etc. sub-screens |
| ContentScreen | COMPLETE | Full CRUD + publish/copy/move/reorder/versions | Most complex screen. Select + Tree + Version phases |
| DatatypesScreen | COMPLETE | Full CRUD + reorder + search | Browse + Fields dual phase |
| MediaScreen | COMPLETE | Create/Read/Delete | No metadata edit in TUI |
| RoutesScreen | COMPLETE | Full CRUD | Regular + Admin modes |
| UsersScreen | COMPLETE | Full CRUD | With permission display per role |
| ActionsScreen | COMPLETE | 12 actions, all real | Destructive actions have confirmation dialogs |
| ConfigScreen | COMPLETE | Read + Edit | Category browsing + field edit + raw JSON view |
| DatabaseScreen | COMPLETE | Generic CRUD | Any table, paginated, insert/update/delete |
| FieldTypesScreen | COMPLETE | Full CRUD | Regular + Admin modes |
| PluginsScreen | COMPLETE | Read + navigate | List with drift/circuit breaker detection |
| PluginDetailScreen | COMPLETE | Enable/Disable/Reload/Approve | 6 actions, all implemented |
| QuickstartScreen | COMPLETE | Install | Schema selection + installation |
| DeployScreen | COMPLETE | Test/Pull/Push/Dry-run | Environment-based sync |
| PipelinesScreen | COMPLETE | Read-only (by design) | Chain inspection with stats |
| PipelineDetailScreen | COMPLETE | Read-only (by design) | Entry inspection |
| WebhooksScreen | FUNCTIONAL | Read-only (GAP) | **Missing Create/Edit/Delete** |

### Known Gaps (to document, not block QA)

1. **WebhooksScreen** — read-only, no CUD operations wired
2. **MediaScreen** — no metadata edit (alt text, caption, etc.)
3. **10 dead keybindings** — defined but never matched: TitlePrev, TitleNext, PagePrev, PageNext, ScreenNext, ScreenToggle, ScreenReset, Accordion, TabPrev, TabNext
4. **DeployScreen hardcoded keys** — t/p/s/P/S bypass KeyMap (not remappable)
5. **1 dead FormDialogAction** — `FORMDIALOGINITIALIZEROUTECONTENT` (line 25, never used)

### Form/Dialog Status

- **36 FormDialogActions**: 35 COMPLETE, 1 dead code
- **26 DialogActions**: all 26 COMPLETE
- **9 field bubble types**: text, textarea, select, boolean, slug, url, email, number, date — all COMPLETE
- **6 fallback field types**: richtext, idref, datetime, media, json, time — render as TextInput, validation works
- **Validation**: COMPLETE — 15 type validators, composable rules, per-field error display/clear

## Phase Structure

| Phase | File | Focus | Screens |
|-------|------|-------|---------|
| 1 | `TUI_QA_PHASE1.md` | Foundation | Home, Navigation, Actions, Quickstart |
| 2 | `TUI_QA_PHASE2.md` | Content System | Content (Select + Tree + Versions), all content operations |
| 3 | `TUI_QA_PHASE3.md` | Schema Management | Datatypes, Fields, FieldTypes, Routes |
| 4 | `TUI_QA_PHASE4.md` | Administration | Users, Config, Database, Media |
| 5 | `TUI_QA_PHASE5.md` | Extensions | Plugins, Deploy, Pipelines, Webhooks |
| 6 | `TUI_QA_PHASE6.md` | Edge Cases | Resize, rapid input, empty states, error states, admin mode parity |

## Execution

- Terminal: 120x40 (consistent snapshot dimensions)
- Each phase is independent but assumes Phase 1 has run (DB Wipe & Redeploy + Quickstart install)
- Use `tui_wait` with `text` match for deterministic sync — avoid `stable_ms` alone
- Golden snapshots in `testdata/tui_goldens/` for regression
- Tests are idempotent: each phase can re-run after a fresh wipe+install

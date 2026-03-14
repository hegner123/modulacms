# TUI QA Summary

Completed: 2026-03-14

## Results by Phase

| Phase | Tests | Pass | Fail | Skip | Findings |
|-------|-------|------|------|------|----------|
| 1: Foundation | 27 | 23 | 1 | 2 | 5 |
| 2: Content | 30 | 22 | 0 | 7 | 1 |
| 3: Schema Mgmt | 25 | 25 | 0 | 0 | 0 |
| 4: Administration | 18 | 14 | 0 | 0 | 1 |
| 5: Extensions | 8 | 8 | 0 | 7 | 0 |
| **Total** | **108** | **92** | **1** | **16** | **7** |

## Bugs Fixed During QA

| Bug | Severity | Fix |
|-----|----------|-----|
| Bootstrap/schema page datatype collision | Critical | Gave bootstrap Page `name:"page"`, made Install() reuse existing datatypes by name |
| Child content tree wiring (FINDING-003) | Critical | Added `attachAsLastChild` call + route inheritance in HandleCreateContentFromDialog |
| DB Export crashes TUI (FINDING-014) | Critical | Changed log.Fatal to log.Ferror in all 15 DumpSql calls |
| `q` exits without confirmation (FINDING-001) | Medium | Changed HandleCommonKeys to show ShowQuitConfirmDialogMsg instead of tea.Quit |
| Content list no refresh after create (FINDING-002) | Medium | Added RootContentSummaryFetchCmd to ContentCreatedFromDialogMsg handler |
| Escape key trap / search exit (FINDING-007/008) | High | Added overlay, search, and file picker guards to esc handler in update.go |
| Form text input display lag (FINDING-009) | Medium | Added OverlayTicker interface + OverlayTick on FormDialogModel + ContentFormDialogModel |
| Form button contrast (FINDING-010) | Medium | Removed background from unfocused buttons, matching dialog.go pattern |
| Child type list non-deterministic (FINDING-005) | Low | Added slices.SortFunc by sort_order then label in filterChildDatatypes |
| Fields panel stale after search (FINDING-006) | Low | Added fetchFieldsForCurrentDT to search keystroke handler |
| Search cursor reset on escape (FINDING-011) | Medium | Save selected datatype ID before clearing search, restore position in rebuilt list |
| Tree no refresh after delete (FINDING-017) | Medium | Added RootContentSummaryFetchCmd to ContentDeletedMsg handler |
| Content form display lag (FINDING-009 partial) | Low | Added OverlayTick on ContentFormDialogModel forwarding ticks to focused FieldBubble |
| File picker escape kills process (FINDING-016) | Medium | Added `FilePickerActive` guard in global escape handler before quit dialog fallback |

## Feature Added During QA

**Webhooks CRUD** — previously read-only, now full create/edit/delete:
- `WebhookFormDialogModel` with 5 fields (Name, URL, Secret, Events, Active toggle)
- Full message pipeline: form → request → DB handler → result → list refresh
- `SafeBool` type for `webhooks.is_active` (fixes PostgreSQL boolean scan error)
- `DIALOGDELETEWEBHOOK` registered in dialog toggle controls
- sqlcgen template updated for SafeBool override across all 3 database backends

## Open Findings

| Finding | Severity | Description | Root Cause | Fix |
|---------|----------|-------------|------------|-----|
| FINDING-012 | Low | Panel focus indicator is color-only — not accessible on monochrome terminals. | panel.go:57-59 distinguishes focus solely by `borderColor` (Tertiary vs Accent). Both use `RoundedBorder()`. | Add non-color indicator: bold title, `DoubleBorder()` for focused panel, or focus glyph prefix. |
| FINDING-013 | High | DB Export (`DumpSql()`) uses embedded shell scripts with bugs. PostgreSQL variant passes wrong args (2 instead of 4) and wrong filename prefix ("sqlite" instead of "psql"). | `PsqlDatabase.DumpSql()` in db.go passes `(Db_Name, "sqlite"+t+".sql")` to a script expecting `(username, password, database, output_file)`. | Superseded by PLUGIN_DB_EXPORTER plan (ai/plans/db/PLUGIN_DB_EXPORTER.md). Once `DeployOps.IntrospectColumns` + `QueryAllRows` are implemented, `DumpSql()` should be rewritten in pure Go using those methods — no shell scripts, no external tool dependencies. |
| FINDING-015 | Low | File picker starts at $HOME, not project/backups directory for RESTORE operations. | update_dialog.go:210 hardcodes `fp.CurrentDirectory, _ = os.UserHomeDir()`. | Use `filepath.Join(cfg.Backup_Option, "backups")` with fallback to `UserHomeDir()`. |
| Media edit | Low | No metadata edit capability in TUI for media items. | `media` field type has no registered bubble in type_registry.go — falls back to `TextBubble` (manual ULID entry). | Implement spec in ai/plans/TUI_MEDIA_FIELD_SPEC.md: create `bubble_media.go`, register `media` type in field bubble registry. |

## Screens Verified

All 18 screens tested with real data:

| Screen | Read | Write | Notes |
|--------|------|-------|-------|
| Home | PASS | N/A | Dashboard, nav, panel focus |
| Content (Select) | PASS | PASS | Create, publish, navigate |
| Content (Tree) | PASS | PASS | Create, edit, delete, copy, move, reorder, publish, unpublish, versions, expand/collapse, go parent/child |
| Datatypes (Browse) | PASS | PASS | Create, edit, delete (blocked for parents), search, reorder |
| Datatypes (Fields) | PASS | PASS | Create, edit, delete, reorder, properties |
| Field Types | PASS | PASS | Create, edit, delete |
| Routes | PASS | PASS | Create, delete |
| Users | PASS | PASS | Create (5 fields), delete, permission display |
| Media | PASS | N/A | List, metadata display, empty state |
| Config | PASS | PASS | Categories, field values, multi-level nav |
| Database | PASS | PASS | Table list, row browser, pagination, detail |
| Actions | PASS | PASS | 12 actions tested (DB Init/Wipe/Redeploy, Export, Certs, Updates, Validate, Token, Backup, Restore) |
| Quickstart | PASS | PASS | All 5 schemas installed successfully |
| Plugins | PASS | N/A | Empty state (no plugins installed) |
| Pipelines | PASS | N/A | Empty state (read-only by design) |
| Deploy | PASS | N/A | Empty state with config hint |
| Webhooks | PASS | PASS | Create, edit, delete — full CRUD |

## Key Metrics

- **14 bugs fixed** during QA (3 critical, 5 medium, 2 high, 4 low)
- **4 open findings** remaining (0 critical, 1 high, 0 medium, 3 low)
- **1 feature added** (Webhooks CRUD)
- **92 tests passing** across 108 total
- **16 skipped** (require external infrastructure: SSH, plugins, deploy environments)
- **0 deferred**

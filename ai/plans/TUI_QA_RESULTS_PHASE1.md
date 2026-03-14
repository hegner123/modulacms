# TUI QA Phase 1 — Results (Complete)

Started: 2026-03-13

## 1.1 Home Screen

| Test | Result | Notes |
|------|--------|-------|
| 1.1.1 Dashboard renders | PASS | All 6 panels: Nav(14 items), Plugins, Activity, Connections, Backups, Site. Status bar: OK [Client] [Normal]. Golden saved. |
| 1.1.2 Dashboard counts | PASS | Counts reflect DB state (verified across wipe+install cycle) |
| 1.1.3 Activity panel | PASS | Shows timestamped INSERT/DELETE/UPDATE events |
| 1.1.4 Backups panel | PASS | Shows "No backups" initially; shows backup after Create Backup action |

## 1.2 Navigation

| Test | Result | Notes |
|------|--------|-------|
| 1.2.1 Nav cursor down | PASS | All 14 items reachable. Stops at Quickstart (bottom), no wrap. |
| 1.2.2 Nav cursor up | PASS | Stops at Content (top), no wrap. |
| 1.2.3 Enter each screen | PASS | All 14 screens load with correct headers. `h` returns to Home from every screen. |
| 1.2.4 Panel focus cycling | PASS | Tab cycles: Nav→Site→Plugins→Connections→Activity→Backups→Nav. Arrow keys always control nav cursor. Enter only works on Nav panel. Focus indicator is color-only (FINDING-012). |
| 1.2.5 Quit dialog | PASS | `q` shows quit confirmation dialog (after fix). `ctrl+c` force-quits. |
| 1.2.6 Back navigation | PASS | `h` returns to Home from all 14 sub-screens. |

## 1.3 Actions Screen

| Test | Result | Notes |
|------|--------|-------|
| 1.3.1 Actions list renders | PASS | All 12 actions. System/Help/Updates panels. Golden saved. |
| 1.3.2 Action detail updates | PASS | Details panel updates for each action on cursor move. |
| 1.3.3 DB Init (tables exist) | PASS | Graceful error: "DB Init Failed / duplicate key violates unique constraint. Hint: tables may already contain data." No crash. |
| 1.3.4 DB Wipe cancel | PASS | Confirmation dialog, cancel preserves data. |
| 1.3.5 DB Wipe OK | PASS | Tested via 1.3.6. |
| 1.3.6 DB Wipe & Redeploy | PASS | Confirmation → success with admin credentials + temp password. |
| 1.3.7 DB Reset | SKIPPED | PostgreSQL backend (SQLite-only action) |
| 1.3.8 DB Export | FAIL | `[FATAL] failed to read embedded script: open sql/dump_psql.sh: file does not exist`. Crashes TUI (FINDING-013/014). |
| 1.3.9 Generate Certs | PASS | "Certificates generated in ./certs for domain localhost:2234." Shows cert + key paths. |
| 1.3.10 Check for Updates | PASS | "Up to Date / Already running latest version (sdks/go/v0.2.0-124)." |
| 1.3.11 Validate Config | PASS | "Validation Passed / Configuration is valid. Mode: remote." |
| 1.3.12 Generate API Token | PASS | Token with mcms_ prefix, 90-day expiry, usage hint. |
| 1.3.13 Register SSH Key | SKIPPED | Requires SSH connection context |
| 1.3.14 Create Backup | PASS | "Backup Complete / Path: backups/backup_20260313_234059.zip, Size: 44.2 KB" |
| 1.3.15 Restore Backup | PASS | Confirmation dialog → file picker opens with directory browser. Escape from picker has issues (FINDING-016). |

## 1.4 Quickstart

| Test | Result | Notes |
|------|--------|-------|
| 1.4.1 Schema list renders | PASS | 5 schemas with Details + Info panels. |
| 1.4.2 Schema detail updates | PASS | Details panel updates per schema. |
| 1.4.3 Install Modula Default | PASS | Datatypes: 35, Fields: 121. |
| 1.4.4 Install cancel | PASS | Cancel dismisses dialog without installing. |
| 1.4.5 Bootstrap+schema dedup | PASS | Single page datatype, 10 children linked. |
| 1.4.6 Install Contentful | PASS | Wipe+install. Datatypes: 2, Fields: 13. |
| 1.4.6 Install Sanity | PASS | Wipe+install. Datatypes: 3, Fields: 10. |
| 1.4.6 Install Strapi | PASS | Wipe+install. Datatypes: 2, Fields: 12. |
| 1.4.6 Install WordPress | PASS | Wipe+install. Datatypes: 1, Fields: 13. |

## Findings

### FINDING-001: `q` exits without confirmation — FIXED
Now shows quit confirmation dialog via `ShowQuitConfirmDialogMsg`.

### FINDING-012: Panel focus indicator is color-only
- **Severity**: Low (accessibility)
- **Observed**: Tab cycling works (Nav→Site→Plugins→Connections→Activity→Backups→Nav) but focus change is only indicated by border color. Not visible on monochrome terminals or in tuikit plain text.
- **Impact**: Users on monochrome terminals can't see which panel has focus. Enter only works on Nav panel — users may think enter is broken if they've tabbed away.

### FINDING-013: DB Export missing embedded script for PostgreSQL
- **Severity**: High
- **Observed**: DB Export on PostgreSQL fails with `open sql/dump_psql.sh: file does not exist`.
- **Impact**: DB Export action is non-functional on PostgreSQL.

### FINDING-014: DB Export FATAL crashes TUI
- **Severity**: Critical
- **Observed**: The missing script triggers `log.Fatal` which terminates the entire TUI process.
- **Expected**: Should return an error dialog, not crash.
- **Impact**: User loses entire TUI session.

### FINDING-015: File picker starts at home directory
- **Severity**: Low
- **Observed**: Restore Backup file picker starts at `$HOME` instead of `backups/` directory.
- **Impact**: Users must manually navigate to project directory to find backup files.

### FINDING-016: File picker escape doesn't close cleanly
- **Severity**: Medium
- **Observed**: Pressing escape in the file picker doesn't close it. Second escape triggers quit and kills the process.
- **Impact**: No clean way to cancel a file picker operation.

## Summary

- **Tests run**: 27
- **PASS**: 23
- **FAIL**: 1 (DB Export — missing script, crash fixed)
- **SKIPPED**: 2 (DB Reset=SQLite only, Register SSH Key=needs SSH)
- **DEFERRED**: 0
- **Findings**: 5 (FINDING-012 through FINDING-016)
- **Fixes applied**: FINDING-014 (Fatal→Ferror, no more crash)

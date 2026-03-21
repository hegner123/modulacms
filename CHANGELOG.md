# Changelog

All notable changes to ModulaCMS are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added

- **Windows platform support** — ModulaCMS now compiles and runs on Windows (amd64 and arm64)
  - Replaced `process.Signal(syscall.SIGTERM)` self-signaling with `os.Exit(1)` for cross-platform shutdown
  - Rewrote `DumpSql` methods as direct CLI calls (`sqlite3`, `mysqldump`, `pg_dump`) instead of bash scripts
  - Added Windows editor fallback (`notepad`) in TUI when `$EDITOR`/`$VISUAL` are unset
  - Added `.exe` suffix handling in self-updater asset matching
  - Skipped Unix file permission checks on Windows in `VerifyBinary`
  - Added `#cgo windows LDFLAGS` to vendored go-webp encoder for Windows linking
  - Cross-platform `just check` recipe (`NUL` on Windows, `/dev/null` on Unix)
- **Windows CI workflow** — separate `windows.yml` builds for amd64 (MINGW64) and arm64 (llvm-mingw cross-compilation with libwebp built from source)
- **CI badges** — CI/CD and Windows build status badges in README
- **Config layering** — `--overlay` CLI flag merges a per-environment overlay file on top of the base config. Overlay files contain only the fields that differ, replacing full config duplication across environments.
  - `LayeredFileProvider` reads base + overlay, shallow-merges at load time
  - `MergeMaps` shared helper extracted from `Manager.Update()`
  - `config show --raw` prints the overlay file contents only
  - `config set --base` writes to the base config when layered
  - `config overlay --env <name>` scaffolds a minimal overlay file
- **Validation tables** — `validations` and `admin_validations` schema directories (`sql/schema/38_validations/`, `sql/schema/39_admin_validations/`)

### Changed

- `Manager.Update()` refactored to use `MergeMaps` instead of inline merge loop

### Fixed

- `DumpSql` methods were broken on all platforms due to incorrect embed paths for bash scripts — rewritten in pure Go

## Earlier Changes

Changes prior to this changelog are documented in git history. Key milestones:

- **Admin panel** — server-rendered HTMX + templ admin UI with full CRUD for all resources, Tailwind CSS, Light DOM web components, block editor
- **Plugin system** — Lua-based plugins with sandboxed VMs, hook engine, HTTP bridge, circuit breakers, hot reload, file watcher, DB state coordinator
- **Observability** — Sentry/Datadog/New Relic integration with HTTP, SSH, DB, and runtime metrics
- **Query builder** — filter, sort, paginate, aggregate, upsert support with Lua API
- **Search** — full-text content search with SDK support (TypeScript, Go, Swift)
- **Media folders** — hierarchical folder CRUD with tree UI and SDK support
- **Block editor** — rich text editing with undo/redo
- **Deploy sync** — export/import/push/pull content between environments including plugin tables
- **SDKs** — TypeScript (6 packages), Go, and Swift SDKs for content delivery and admin CRUD
- **Tri-database support** — SQLite, MySQL, and PostgreSQL via `DbDriver` interface
- **RBAC** — role-based access control with 72 granular permissions
- **Webhook system** — event dispatch with signing and delivery tracking
- **SSH TUI** — Bubbletea terminal UI via Charmbracelet Wish

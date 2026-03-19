# Changelog

All notable changes to ModulaCMS are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added

- **Config layering** — `--overlay` CLI flag merges a per-environment overlay file on top of the base config. Overlay files contain only the fields that differ, replacing full config duplication across environments.
  - `LayeredFileProvider` reads base + overlay, shallow-merges at load time
  - `MergeMaps` shared helper extracted from `Manager.Update()`
  - `config show --raw` prints the overlay file contents only
  - `config set --base` writes to the base config when layered
  - `config overlay --env <name>` scaffolds a minimal overlay file
- **Validation tables** — `validations` and `admin_validations` schema directories (`sql/schema/38_validations/`, `sql/schema/39_admin_validations/`)

### Changed

- `Manager.Update()` refactored to use `MergeMaps` instead of inline merge loop

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

# What Is Extra

Things that work but may not be pulling their weight. Not bugs, not bad code - just effort that could be redirected.

## TUI Raw Database Pages

The TUI has CREATEPAGE, READPAGE, UPDATEPAGE, DELETEPAGE screens that let you perform raw CRUD on any database table. These are direct SQL-level operations wrapped in a terminal form.

With the admin panel now providing a proper UI for all entity management, these pages serve a narrow audience. They're developer debugging tools masquerading as user features. If they stay, they should be gated behind a flag or hidden menu, not presented alongside the CMS content management screens.

## Six Docker Compose Files

Six separate compose files for different database combinations:
- `docker-compose.yml` (dev)
- `docker-compose.full.yml` (all databases)
- `docker-compose.sqlite.yml`
- `docker-compose.mysql.yml`
- `docker-compose.postgres.yml`
- `docker-compose.prod.yml`

Docker Compose profiles could reduce this to two files (dev and prod) with `--profile sqlite`, `--profile postgres`, etc. The Justfile already provides the abstraction layer (`just docker-sqlite-up`, `just docker-postgres-up`), so users rarely interact with compose files directly.

## Dual Management Interfaces

Both the TUI and admin panel can manage: content, datatypes, fields, routes, users, media, roles, permissions, tokens, and plugins. Maintaining feature parity across both is a significant ongoing cost.

The TUI is valuable for SSH-based server administration. The admin panel is better for content editing. But every new feature must be built twice. Consider whether the TUI should focus on operations (backup, config, plugins, monitoring) and leave content management to the admin panel.

## 200+ Message Types in TUI

The TUI defines 200+ message types across `message_types.go` (780 lines) and `admin_message_types.go` (266 lines). Many are one-line structs used in exactly one place. While type safety is good, the sheer volume suggests the message system could use intermediate abstraction - for example, a generic `FetchResultMsg[T any]` instead of separate types for each entity.

## Full Tri-Database Testing in CI

Every test runs against SQLite in CI. The tri-database support means there are three codepaths for every operation, but CI only tests one. The MySQL and PostgreSQL paths are validated locally via Docker but not in the CI pipeline. If the tri-database support is a core feature, it should have CI coverage. If it's rarely used, consider whether maintaining three backends is worth the ongoing effort.

## Plugin System Maturity vs Usage

The plugin system is arguably the most polished subsystem: VM pooling, circuit breakers, hot reload, approval workflows, metrics, comprehensive CLI. It represents significant engineering investment. But if no production plugins exist yet, much of this infrastructure is ahead of demand.

This isn't necessarily bad - it's a strategic investment. But it means the plugin system is a bet on future adoption rather than a response to current need.

## 70+ AI Documentation Files

The `ai/` directory contains detailed documentation about nearly every aspect of the system. Some files overlap (architecture + domain + package docs covering the same concepts from different angles). A few files in `ai/refactor/` describe plans from months ago that may or may not still be relevant.

Consolidation and pruning would make the documentation more maintainable. A single comprehensive guide per subsystem would be more useful than 3-4 partial docs covering overlapping ground.

## Dead TUI Pages

- `DEVELOPMENT` page exists with minimal functionality
- `DYNAMICPAGE` has a template but no implementation
- `QUICKSTARTPAGE` may be partially implemented

These should either be completed or removed to avoid confusion when navigating the TUI.

# START.md

ModulaCMS AI Agent Documentation Index

**Purpose:** Navigation index for ModulaCMS documentation
**Last Updated:** 2026-01-13


## Onboarding

Memorize the following absolute paths, not their contents.
Memorize the key directories, not their contents.

CURRENT IMPLEMENTATION PROJECT: None - Phase 2 complete, awaiting Phase 3 planning

COMPLETED PROJECTS:
- **[TABLE_REFACTOR_PLAN.md](TABLE_REFACTOR_PLAN.md)** - CLI Model TableModel Extraction (Phase 2 of 4) ✅
- **[FORM_REFACTOR_PLAN.md](FORM_REFACTOR_PLAN.md)** - CLI Model FormModel Extraction (Phase 1 of 4) ✅

---

## Documentation Index

### Architecture
- **[TREE_STRUCTURE.md](architecture/TREE_STRUCTURE.md)** - Sibling-pointer tree implementation
- **[CONTENT_MODEL.md](architecture/CONTENT_MODEL.md)** - Dynamic schema system
- **[TUI_ARCHITECTURE.md](architecture/TUI_ARCHITECTURE.md)** - Elm Architecture in practice
- **[DATABASE_LAYER.md](architecture/DATABASE_LAYER.md)** - Database abstraction philosophy
- **[MULTI_DATABASE.md](architecture/MULTI_DATABASE.md)** - Multi-database support
- **[HTTP_SSH_SERVERS.md](architecture/HTTP_SSH_SERVERS.md)** - Triple-server architecture
- **[PLUGIN_ARCHITECTURE.md](architecture/PLUGIN_ARCHITECTURE.md)** - Lua plugin system design

### Workflows
- **[ADDING_FEATURES.md](workflows/ADDING_FEATURES.md)** - Feature development process
- **[ADDING_TABLES.md](workflows/ADDING_TABLES.md)** - Database table creation
- **[CREATING_TUI_SCREENS.md](workflows/CREATING_TUI_SCREENS.md)** - TUI screen development
- **[TESTING.md](workflows/TESTING.md)** - Testing strategies
- **[DEBUGGING.md](workflows/DEBUGGING.md)** - Debugging guide

### Packages
- **[CLI_PACKAGE.md](packages/CLI_PACKAGE.md)** - TUI implementation
- **[MODEL_PACKAGE.md](packages/MODEL_PACKAGE.md)** - Business logic and data structures
- **[MIDDLEWARE_PACKAGE.md](packages/MIDDLEWARE_PACKAGE.md)** - HTTP middleware (CORS, auth, sessions)
- **[PLUGIN_PACKAGE.md](packages/PLUGIN_PACKAGE.md)** - Lua plugin system
- **[AUTH_PACKAGE.md](packages/AUTH_PACKAGE.md)** - OAuth and authentication
- **[MEDIA_PACKAGE.md](packages/MEDIA_PACKAGE.md)** - Media processing and S3

### Database
- **[SQL_DIRECTORY.md](database/SQL_DIRECTORY.md)** - SQL file organization
- **[DB_PACKAGE.md](database/DB_PACKAGE.md)** - Database abstraction layer
- **[SQLC.md](database/SQLC.md)** - Type-safe query generation

### Domain
- **[ROUTES_AND_SITES.md](domain/ROUTES_AND_SITES.md)** - Multi-site architecture
- **[DATATYPES_AND_FIELDS.md](domain/DATATYPES_AND_FIELDS.md)** - Dynamic content schema
- **[CONTENT_TREES.md](domain/CONTENT_TREES.md)** - Tree operations and navigation
- **[MEDIA_SYSTEM.md](domain/MEDIA_SYSTEM.md)** - S3 integration and optimization
- **[AUTH_AND_OAUTH.md](domain/AUTH_AND_OAUTH.md)** - Authentication flows

### Reference
- **[FILE_TREE.md](FILE_TREE.md)** - Project structure
- **[CLAUDE.md](../CLAUDE.md)** - Development guidelines and build commands
- **[PATTERNS.md](reference/PATTERNS.md)** - Common code patterns
- **[DEPENDENCIES.md](reference/DEPENDENCIES.md)** - Why each dependency exists
- **[TROUBLESHOOTING.md](reference/TROUBLESHOOTING.md)** - Common issues and solutions
- **[QUICKSTART.md](reference/QUICKSTART.md)** - Get started fast
- **[GLOSSARY.md](reference/GLOSSARY.md)** - Term definitions

---

## Key Directories

```
cmd/main.go              - Application entry point
internal/cli/            - TUI implementation
internal/db/             - Database interface
internal/db-sqlite/      - SQLite driver
internal/db-mysql/       - MySQL driver
internal/db-psql/        - PostgreSQL driver
internal/model/          - Business logic
sql/schema/              - Database migrations
sql/mysql/               - MySQL queries
sql/postgres/            - PostgreSQL queries
```

---

**Last Updated:** 2026-01-13

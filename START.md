# START.md

ModulaCMS AI Agent Onboarding

**Location:** /Users/home/Documents/Code/Go_dev/modulacms/START.md
**Last Updated:** 2026-01-22

---

## Onboarding Instructions

When the onboarding skill triggers, execute these steps in order:

### Step 1: Check Memory for Cached Paths

Search mem MCP for previously stored documentation paths:

```
mcp__mem__recall
  key: "modulacms-paths"
```

If results exist, use cached paths but verify files still exist. If files are missing, proceed to Step 2.

**Freshness rule:** Memory is a cache, not truth. Filesystem state is authoritative.

### Step 2: Store Documentation Paths in Memory

Store paths using the mem MCP tool for faster onboarding in future sessions.

```
mcp__mem__store
  key: "modulacms-paths"
  content: "Primary: /Users/home/Documents/Code/Go_dev/modulacms/CLAUDE.md (dev guidelines), /Users/home/Documents/Code/Go_dev/modulacms/ai/ (documentation root). Active project: See START.md Project Status section."
  tags: ["paths", "documentation"]
```

### Step 3: Read Required Context

Read these files to establish working context:

1. **Required:** `/Users/home/Documents/Code/Go_dev/modulacms/CLAUDE.md`
   - Build commands, code style, development guidelines
   - Read fully before any code changes

2. **If working on active project:** Read the file listed under ACTIVE in Project Status below

### Step 4: Acknowledge Onboarding Complete

Confirm to user:
- Which files were read
- Current active project (from Project Status)
- Any stale team-memory entries that were refreshed

---

## Project Status

This section tracks what the user is working on. It is for **user awareness only**.

**ACTIVE — Work on this when asked:**
| Project | Path | Status |
|---------|------|--------|
| SQLC Refactor | `/Users/home/Documents/Code/Go_dev/modulacms/ai/refactor/SQLC_REDUCTION/` | In progress |

**PAUSED — Do NOT resume unless user explicitly requests:**
| Project | Path | Status | Notes |
|---------|------|--------|-------|
| Core CMS Content Creation | `/Users/home/Documents/Code/Go_dev/modulacms/ai/refactor/CORE-CMS-CONTENT-CREATION-PLAN.md` | 90% complete | Testing remaining |

**COMPLETED — Historical reference only:**
| Project | Path |
|---------|------|
| CLI TableModel Extraction | `/Users/home/Documents/Code/Go_dev/modulacms/ai/TABLE_REFACTOR_PLAN.md` |
| CLI FormModel Extraction | `/Users/home/Documents/Code/Go_dev/modulacms/ai/FORM_REFACTOR_PLAN.md` |

---

## Documentation Index

All paths are absolute from project root: `/Users/home/Documents/Code/Go_dev/modulacms/`

### Legend
- ⭐ High-value reference — consult when working in related area
- ⚠️ Critical constraint — read before making changes in this domain

### When to Read What

| Task Type | Read First |
|-----------|------------|
| Any code changes | `/Users/home/Documents/Code/Go_dev/modulacms/CLAUDE.md` |
| Database/SQL work | SQLC_REDUCTION/00-OVERVIEW.md, DATABASE_LAYER.md, SQLC.md |
| SQLC Refactor work | SQLC_REDUCTION/00-OVERVIEW.md (start here, links to all docs) |
| TUI/CLI screens | TUI_ARCHITECTURE.md, CLI_PACKAGE.md, MODEL_STRUCT_GUIDE.md |
| Content tree operations | TREE_STRUCTURE.md, CONTENT_TREES.md |
| API changes | API_CONTRACT.md |
| Debugging | TROUBLESHOOTING.md, then relevant architecture doc |

### Architecture
| Document | Path | Description |
|----------|------|-------------|
| Tree Structure | `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/TREE_STRUCTURE.md` | Sibling-pointer tree implementation |
| Content Model | `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/CONTENT_MODEL.md` | Dynamic schema system |
| TUI Architecture | `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/TUI_ARCHITECTURE.md` | Elm Architecture in practice |
| Database Layer | `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/DATABASE_LAYER.md` | Database abstraction philosophy |
| Multi-Database | `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/MULTI_DATABASE.md` | Multi-database support |
| HTTP/SSH Servers | `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/HTTP_SSH_SERVERS.md` | Triple-server architecture |
| Plugin Architecture | `/Users/home/Documents/Code/Go_dev/modulacms/ai/architecture/PLUGIN_ARCHITECTURE.md` | Lua plugin system design |

### API
| Document | Path | Description |
|----------|------|-------------|
| API Contract ⭐ | `/Users/home/Documents/Code/Go_dev/modulacms/ai/api/API_CONTRACT.md` | Complete REST API reference (v1) |
| API Index | `/Users/home/Documents/Code/Go_dev/modulacms/ai/api/README.md` | API documentation index |

### Refactoring Analysis
| Document | Path | Description |
|----------|------|-------------|
| Analysis Summary ⭐ | `/Users/home/Documents/Code/Go_dev/modulacms/ai/refactor/ANALYSIS-SUMMARY-2026-01-15.md` | CMS content creation architecture |
| CLI Consolidation ⭐ | `/Users/home/Documents/Code/Go_dev/modulacms/ai/refactor/CLI-PACKAGE-CONSOLIDATION.md` | File consolidation (44 to 28 files) |
| DB Consolidation ⭐ | `/Users/home/Documents/Code/Go_dev/modulacms/ai/refactor/DB-PACKAGE-CONSOLIDATION.md` | DB package analysis (36 files) |
| CLI-DB Interaction | `/Users/home/Documents/Code/Go_dev/modulacms/ai/refactor/CLI-DB-INTERACTION-ANALYSIS.md` | CLI and DB package interaction |
| Problem Statement | `/Users/home/Documents/Code/Go_dev/modulacms/ai/refactor/PROBLEM.md` | Coupled operations architecture |
| Implementation Guide | `/Users/home/Documents/Code/Go_dev/modulacms/ai/refactor/SUGGESTION-2026-01-15.md` | Hybrid approach (Phase 1) |
| Plugin Constraints | `/Users/home/Documents/Code/Go_dev/modulacms/ai/refactor/PROBLEM-UPDATE-2026-01-15-PLUGINS.md` | Plugin extensibility constraints |

### SQLC Type Unification (Active Project)
| Document | Path | Description |
|----------|------|-------------|
| Overview ⭐ | `.../ai/refactor/SQLC_REDUCTION/00-OVERVIEW.md` | Status, goals, architecture, document index |
| Schema Improvements | `.../ai/refactor/SQLC_REDUCTION/01-SCHEMA-IMPROVEMENTS.md` | ULID PKs, indexes, constraints, change_events |
| Type System ⭐ | `.../ai/refactor/SQLC_REDUCTION/02-TYPE-SYSTEM.md` | Custom Go type definitions |
| SQLC Config | `.../ai/refactor/SQLC_REDUCTION/03-SQLC-CONFIG.md` | Complete sqlc.yml configuration |
| Implementation Phases ⚠️ | `.../ai/refactor/SQLC_REDUCTION/04-IMPLEMENTATION-PHASES.md` | All phases and steps with details |
| HQ Project | `.../ai/refactor/SQLC_REDUCTION/05-HQ-PROJECT.md` | Multi-agent coordination JSON |
| Operations | `.../ai/refactor/SQLC_REDUCTION/06-OPERATIONS.md` | Agent instructions, verification, rollback |
| Summary | `.../ai/refactor/SQLC_REDUCTION/07-SUMMARY.md` | Tables, expected results, references |
| Concerns | `.../ai/refactor/SQLC_REDUCTION/SQLC_REDUCTION_CONCERNS.md` | Critical review, risks, recommendations |

### Workflows
| Document | Path | Description |
|----------|------|-------------|
| Adding Features | `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/ADDING_FEATURES.md` | Feature development process |
| Adding Tables | `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/ADDING_TABLES.md` | Database table creation |
| Creating TUI Screens | `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/CREATING_TUI_SCREENS.md` | TUI screen development |
| Test Organization ⭐ | `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/TEST_ORGANIZATION.md` | Test file organization strategy |
| Testing | `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/TESTING.md` | Testing strategies |
| Debugging | `/Users/home/Documents/Code/Go_dev/modulacms/ai/workflows/DEBUGGING.md` | Debugging guide |

### Packages
| Document | Path | Description |
|----------|------|-------------|
| CLI Package | `/Users/home/Documents/Code/Go_dev/modulacms/ai/packages/CLI_PACKAGE.md` | TUI implementation |
| Model Struct Guide ⭐ | `/Users/home/Documents/Code/Go_dev/modulacms/ai/packages/MODEL_STRUCT_GUIDE.md` | Model struct field reference |
| Update Section Review | `/Users/home/Documents/Code/Go_dev/modulacms/ai/packages/UPDATE_SECTION_REVIEW.md` | Update handler analysis |
| Message Flow Guide ⭐ | `/Users/home/Documents/Code/Go_dev/modulacms/ai/packages/DATABASE_MESSAGE_FLOW_GUIDE.md` | Database and CMS message patterns |
| Model Package | `/Users/home/Documents/Code/Go_dev/modulacms/ai/packages/MODEL_PACKAGE.md` | Business logic and data structures |
| Middleware Package | `/Users/home/Documents/Code/Go_dev/modulacms/ai/packages/MIDDLEWARE_PACKAGE.md` | HTTP middleware |
| Plugin Package | `/Users/home/Documents/Code/Go_dev/modulacms/ai/packages/PLUGIN_PACKAGE.md` | Lua plugin system |
| Auth Package | `/Users/home/Documents/Code/Go_dev/modulacms/ai/packages/AUTH_PACKAGE.md` | OAuth and authentication |
| Media Package | `/Users/home/Documents/Code/Go_dev/modulacms/ai/packages/MEDIA_PACKAGE.md` | Media processing and S3 |

### Database
| Document | Path | Description |
|----------|------|-------------|
| SQL Directory | `/Users/home/Documents/Code/Go_dev/modulacms/ai/database/SQL_DIRECTORY.md` | SQL file organization |
| DB Package | `/Users/home/Documents/Code/Go_dev/modulacms/ai/database/DB_PACKAGE.md` | Database abstraction layer |
| SQLC | `/Users/home/Documents/Code/Go_dev/modulacms/ai/database/SQLC.md` | Type-safe query generation |
| Table Creation Order ⚠️ | `/Users/home/Documents/Code/Go_dev/modulacms/ai/reference/TABLE_CREATION_ORDER.md` | FK constraint ordering |

### Domain
| Document | Path | Description |
|----------|------|-------------|
| Routes and Sites | `/Users/home/Documents/Code/Go_dev/modulacms/ai/domain/ROUTES_AND_SITES.md` | Multi-site architecture |
| Datatypes and Fields | `/Users/home/Documents/Code/Go_dev/modulacms/ai/domain/DATATYPES_AND_FIELDS.md` | Dynamic content schema |
| Content Trees | `/Users/home/Documents/Code/Go_dev/modulacms/ai/domain/CONTENT_TREES.md` | Tree operations and navigation |
| Media System | `/Users/home/Documents/Code/Go_dev/modulacms/ai/domain/MEDIA_SYSTEM.md` | S3 integration and optimization |
| Auth and OAuth | `/Users/home/Documents/Code/Go_dev/modulacms/ai/domain/AUTH_AND_OAUTH.md` | Authentication flows |

### Reference
| Document | Path | Description |
|----------|------|-------------|
| File Tree | `/Users/home/Documents/Code/Go_dev/modulacms/ai/FILE_TREE.md` | Project structure |
| Install System ⚠️ | `/Users/home/Documents/Code/Go_dev/modulacms/ai/reference/INSTALL_SYSTEM.md` | Installation status and issues |
| Non-Null Fields ⭐ | `/Users/home/Documents/Code/Go_dev/modulacms/ai/reference/NON_NULL_FIELDS_REFERENCE.md` | Database non-nullable fields |
| Patterns | `/Users/home/Documents/Code/Go_dev/modulacms/ai/reference/PATTERNS.md` | Common code patterns |
| Dependencies | `/Users/home/Documents/Code/Go_dev/modulacms/ai/reference/DEPENDENCIES.md` | Dependency rationale |
| Troubleshooting | `/Users/home/Documents/Code/Go_dev/modulacms/ai/reference/TROUBLESHOOTING.md` | Common issues and solutions |
| Quickstart | `/Users/home/Documents/Code/Go_dev/modulacms/ai/reference/QUICKSTART.md` | Get started fast |
| Glossary | `/Users/home/Documents/Code/Go_dev/modulacms/ai/reference/GLOSSARY.md` | Term definitions |

---

## Key Directories

Directories modified most frequently during development:

```
/Users/home/Documents/Code/Go_dev/modulacms/cmd/main.go           - Application entry point
/Users/home/Documents/Code/Go_dev/modulacms/internal/cli/         - TUI implementation
/Users/home/Documents/Code/Go_dev/modulacms/internal/router/      - REST API handlers
/Users/home/Documents/Code/Go_dev/modulacms/internal/db/          - Database interface
/Users/home/Documents/Code/Go_dev/modulacms/internal/db-sqlite/   - SQLite driver
/Users/home/Documents/Code/Go_dev/modulacms/internal/db-mysql/    - MySQL driver
/Users/home/Documents/Code/Go_dev/modulacms/internal/db-psql/     - PostgreSQL driver
/Users/home/Documents/Code/Go_dev/modulacms/internal/model/       - Business logic
/Users/home/Documents/Code/Go_dev/modulacms/sql/schema/           - Database migrations
/Users/home/Documents/Code/Go_dev/modulacms/sql/mysql/            - MySQL queries
/Users/home/Documents/Code/Go_dev/modulacms/sql/postgres/         - PostgreSQL queries
```

---

## Notes

- Documentation may lag implementation. When code and docs conflict, code is authoritative.
- Check `git log --oneline -10` for recent changes if behavior seems inconsistent with documentation.
- This file is the single source for onboarding. The `ai/` directory contains detailed documentation but no longer has its own START.md.

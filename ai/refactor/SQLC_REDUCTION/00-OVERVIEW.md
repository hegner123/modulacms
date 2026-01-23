# SQLC Configuration Refactor: Type Unification

## Status: COMPLETED (2026-01-23)

Build passes with zero errors. Type unification complete across all packages.

## Problem Statement

The current database layer has **34,589 lines of code**:
- `internal/db/`: 16,795 lines (manual wrapper/mapping layer)
- `internal/db-sqlite/`: 6,168 lines (sqlc generated)
- `internal/db-mysql/`: 5,859 lines (sqlc generated)
- `internal/db-psql/`: 5,767 lines (sqlc generated)

**Root causes:**
1. sqlc generates different Go types per engine (int32 vs int64, sql.NullTime vs sql.NullString)
2. Duplicate struct variants: `*JSON`, `*FormParams` for each model
3. Complex type conversion functions throughout mapping code
4. No compile-time type safety for IDs (can pass UserID where DatatypeID expected)
5. No centralized validation for API/CLI inputs
6. Inconsistent null handling and JSON serialization

## Goals

1. **Simplify** - Reduce wrapper boilerplate by 40-45% (see [SQLC_REDUCTION_CONCERNS.md](SQLC_REDUCTION_CONCERNS.md) for estimate rationale)
2. **Type Safety** - Compile-time prevention of ID mixups
3. **Validation** - Centralized input validation for API and CLI
4. **Reliability** - Enterprise-grade error handling and logging
5. **Consistency** - Unified behavior across all three database engines
6. **Defense in Depth** - Database constraints mirror Go validation
7. **Distribution-Ready** - ULIDs, HLC timestamps, change events for future multi-node support
8. **Audit Trail** - change_events table serves as audit log, replication log, and webhook source

## Architecture Constraint

Go uses **nominal typing**. Even if sqlc generates byte-identical structs across engines, they remain different types. **The wrapper layer is architecturally necessary** for runtime database switching.

However, by using **custom types defined in `internal/db/`**, all three sqlc packages will use the SAME Go types for fields, enabling:
- Compile-time type safety
- Centralized validation
- Consistent JSON serialization
- Specific error messages

```
Application Code (API, CLI, etc.)
         │
         ▼ uses db.DbDriver, db.Datatypes, db.DatatypeID
┌─────────────────────────────────────────────────────────┐
│                    internal/db/                          │
│                                                          │
│  Custom Types (shared by ALL packages):                  │
│  ├── DatatypeID, UserID, ContentID, etc.                │
│  ├── NullableDatatypeID, NullableUserID, etc.           │
│  ├── Timestamp (unified datetime)                        │
│  ├── ContentStatus, FieldType (enums)                   │
│  └── Slug, Email, URL, History (validation types)       │
│                                                          │
│  Wrapper Layer:                                          │
│  ├── db.Datatypes (unified struct)                      │
│  ├── db.DbDriver (interface)                            │
│  └── Database, MysqlDatabase, PsqlDatabase              │
└─────────────────────────────────────────────────────────┘
         │                │                │
         ▼                ▼                ▼
    mdb.Queries     mdbm.Queries     mdbp.Queries
      (sqlc)          (sqlc)           (sqlc)
         │                │                │
         └────── All use db.DatatypeID, db.Timestamp, etc.
```

---

## Document Index

This plan is split into multiple documents for easier navigation:

| Document | Description |
|----------|-------------|
| [00-OVERVIEW.md](00-OVERVIEW.md) | This file - status, problem, goals, architecture |
| [01-SCHEMA-IMPROVEMENTS.md](01-SCHEMA-IMPROVEMENTS.md) | Database schema changes (ULID, indexes, constraints, change_events, backups) |
| [02-TYPE-SYSTEM.md](02-TYPE-SYSTEM.md) | Go type definitions (IDs, timestamps, enums, validation) |
| [03-SQLC-CONFIG.md](03-SQLC-CONFIG.md) | New sqlc.yml configuration |
| [04-IMPLEMENTATION-PHASES.md](04-IMPLEMENTATION-PHASES.md) | All implementation phases and steps |
| [05-HQ-PROJECT.md](05-HQ-PROJECT.md) | HQ multi-agent project configuration |
| [06-OPERATIONS.md](06-OPERATIONS.md) | Agent instructions, verification checklist, rollback plan |
| [07-SUMMARY.md](07-SUMMARY.md) | Expected results, type summary tables, references |
| [SQLC_REDUCTION_CONCERNS.md](SQLC_REDUCTION_CONCERNS.md) | Critical review, risks, recommendations |

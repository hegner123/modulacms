# Table Creation Order Reference

This document defines the correct order for creating ModulaCMS database tables to satisfy foreign key constraints.

> **⚠️ CRITICAL**: The existing `sql/create_order.md` has an error - `content_fields` is listed before `fields`, but `content_fields` has a foreign key dependency on `fields`.

---

## Quick Reference: Correct Creation Order

```
1.  permissions          (no dependencies)
2.  roles                (no dependencies)
3.  media_dimensions     (no dependencies)
4.  users                (→ roles)
5.  tokens               (→ users)
6.  user_oauth           (→ users)
7.  sessions             (→ users)
8.  tables               (→ users)
9.  media                (→ users)
10. admin_routes         (→ users)
11. routes               (→ users)
12. datatypes            (→ users, self-ref parent_id)
13. admin_datatypes      (→ users, self-ref parent_id)
14. fields               (→ datatypes, users)
15. admin_fields         (→ admin_datatypes, users)
16. content_data         (→ routes, datatypes, users, self-ref)
17. admin_content_data   (→ admin_routes, admin_datatypes, users, self-ref)
18. content_fields       (→ content_data, fields, routes, users)
19. admin_content_fields (→ admin_content_data, admin_fields, admin_routes, users)
20. datatypes_fields     (→ datatypes, fields)
21. admin_datatypes_fields (→ admin_datatypes, admin_fields)
```

---

## Dependency Tiers

### Tier 0: Foundation Tables (No Dependencies)

These tables have no foreign key dependencies and can be created first in any order.

```sql
1. permissions
2. roles
3. media_dimensions
```

**Why first**: No foreign keys to other tables.

---

### Tier 1: User Management (Depends on roles)

```sql
4. users → roles.role_id
```

**Why here**: Users depend on roles for the `role` field (FK with DEFAULT 4).

**Dependency chain**: `roles` → `users`

---

### Tier 2: User-Related Tables & Core Content Tables

All tables in this tier depend only on `users` (Tier 1).

```sql
5.  tokens            → users.user_id
6.  user_oauth        → users.user_id
7.  sessions          → users.user_id
8.  tables            → users.user_id (author_id)
9.  media             → users.user_id (author_id)
10. admin_routes      → users.user_id (author_id)
11. routes            → users.user_id (author_id)
12. datatypes         → users.user_id (author_id), datatypes.datatype_id (parent_id, self-ref)
13. admin_datatypes   → users.user_id (author_id), admin_datatypes.admin_datatype_id (parent_id, self-ref)
```

**Why here**: All depend only on users. Self-referential FKs (datatypes.parent_id, admin_datatypes.parent_id) are nullable, so table can be created without existing rows.

**Dependency chain**: `roles` → `users` → [these tables]

---

### Tier 3: Field Definition Tables

```sql
14. fields       → datatypes.datatype_id (parent_id), users.user_id (author_id)
15. admin_fields → admin_datatypes.admin_datatype_id (parent_id), users.user_id (author_id)
```

**Why here**: Depends on datatype tables (Tier 2) and users (Tier 1).

**Dependency chain**:
- `roles` → `users` → `datatypes` → `fields`
- `roles` → `users` → `admin_datatypes` → `admin_fields`

---

### Tier 4: Content Data Tables (Self-Referential Tree Structures)

```sql
16. content_data       → routes.route_id, datatypes.datatype_id, users.user_id (author_id),
                          content_data.content_data_id (parent_id, first_child_id, etc., self-ref)

17. admin_content_data → admin_routes.admin_route_id, admin_datatypes.admin_datatype_id,
                          users.user_id (author_id), admin_content_data.admin_content_data_id (self-ref)
```

**Why here**: Requires routes, datatypes, and users. Self-referential tree pointer FKs are nullable.

**Dependency chain**:
- `roles` → `users` → `routes` + `datatypes` → `content_data`
- `roles` → `users` → `admin_routes` + `admin_datatypes` → `admin_content_data`

---

### Tier 5: Content Field Values

```sql
18. content_fields       → content_data.content_data_id, fields.field_id,
                            routes.route_id, users.user_id (author_id)

19. admin_content_fields → admin_content_data.admin_content_data_id,
                            admin_fields.admin_field_id, admin_routes.admin_route_id,
                            users.user_id (author_id)
```

**Why here**: Requires content data (Tier 4), field definitions (Tier 3), routes (Tier 2), and users (Tier 1).

**Dependency chain**:
- `content_fields`: roles → users → routes + datatypes → fields + content_data → content_fields
- `admin_content_fields`: roles → users → admin_routes + admin_datatypes → admin_fields + admin_content_data → admin_content_fields

---

### Tier 6: Junction/Association Tables

```sql
20. datatypes_fields       → datatypes.datatype_id, fields.field_id
21. admin_datatypes_fields → admin_datatypes.admin_datatype_id, admin_fields.admin_field_id
```

**Why last**: Junction tables that link datatypes to fields. Must wait for both sides to exist.

**Dependency chain**:
- `datatypes_fields`: roles → users → datatypes + fields → datatypes_fields
- `admin_datatypes_fields`: roles → users → admin_datatypes + admin_fields → admin_datatypes_fields

---

## Self-Referential Tables (Special Handling)

Some tables have foreign keys that reference themselves:

### 1. **datatypes** (parent_id → datatypes.datatype_id)

```sql
CREATE TABLE datatypes (
    datatype_id INTEGER PRIMARY KEY,
    parent_id INTEGER DEFAULT NULL REFERENCES datatypes ON DELETE SET DEFAULT,
    ...
);
```

**Strategy**: parent_id is NULLABLE. Create table first, then insert rows:
1. Insert root datatypes (parent_id = NULL)
2. Insert child datatypes (parent_id = existing datatype_id)

### 2. **admin_datatypes** (parent_id → admin_datatypes.admin_datatype_id)

Same strategy as datatypes.

### 3. **content_data** (parent_id, first_child_id, next_sibling_id, prev_sibling_id)

```sql
CREATE TABLE content_data (
    content_data_id INTEGER PRIMARY KEY,
    parent_id INTEGER REFERENCES content_data ON DELETE SET NULL,
    first_child_id INTEGER REFERENCES content_data ON DELETE SET NULL,
    next_sibling_id INTEGER REFERENCES content_data ON DELETE SET NULL,
    prev_sibling_id INTEGER REFERENCES content_data ON DELETE SET NULL,
    ...
);
```

**Strategy**: All tree pointers are NULLABLE. Create table first, then:
1. Insert root content (all tree pointers = NULL)
2. Build tree structure (update tree pointers to link nodes)

### 4. **admin_content_data** (same self-referential tree structure)

Same strategy as content_data.

---

## Foreign Key Constraint Dependency Graph

```
Tier 0
├── permissions
├── roles
└── media_dimensions

Tier 1
└── users ← roles

Tier 2
├── tokens ← users
├── user_oauth ← users
├── sessions ← users
├── tables ← users
├── media ← users
├── admin_routes ← users
├── routes ← users
├── datatypes ← users (+ self)
└── admin_datatypes ← users (+ self)

Tier 3
├── fields ← datatypes, users
└── admin_fields ← admin_datatypes, users

Tier 4
├── content_data ← routes, datatypes, users (+ self)
└── admin_content_data ← admin_routes, admin_datatypes, users (+ self)

Tier 5
├── content_fields ← content_data, fields, routes, users
└── admin_content_fields ← admin_content_data, admin_fields, admin_routes, users

Tier 6
├── datatypes_fields ← datatypes, fields
└── admin_datatypes_fields ← admin_datatypes, admin_fields
```

---

## Error in sql/create_order.md

The existing `sql/create_order.md` has this order:

```
...
14. content_data
15. content_fields    ← ERROR: created before fields
16. fields            ← Should be BEFORE content_fields
17. admin_content_data
...
```

**Problem**: `content_fields` has foreign key `field_id → fields.field_id`, but `fields` table is created **after** `content_fields`. This violates FK constraints.

**Correct order**:
```
14. fields            ← Create first
15. content_data
16. content_fields    ← Create after fields
17. admin_content_data
```

### Why This Might Not Have Caused Issues

1. **CREATE TABLE IF NOT EXISTS**: Schema may already exist
2. **Deferred constraints**: Some databases defer FK checks
3. **Manual schema creation**: Schema created manually in correct order
4. **No strict enforcement**: SQLite doesn't always enforce FK constraints unless `PRAGMA foreign_keys = ON`

### How to Fix

**Option 1: Update sql/create_order.md**
```diff
  content_data
- content_fields
  fields
+ content_fields
  admin_content_data
```

**Option 2: Use correct order from this document**

---

## Migration Strategy

### For New Databases (Fresh Install)

Use the order from "Quick Reference" section above.

### For Existing Databases (Migrations)

When adding new tables:

1. **Identify dependencies**: What tables does the new table reference?
2. **Check if dependencies exist**: Query `sqlite_master` or equivalent
3. **Create dependencies first**: If missing, create in dependency order
4. **Create new table**: After all dependencies exist
5. **Create dependent tables**: Any tables that reference the new table

### Dropping Tables

**Reverse the creation order** to drop tables without FK violations:

```
21. admin_datatypes_fields  (drop first - no dependents)
20. datatypes_fields
19. admin_content_fields
18. content_fields
17. admin_content_data
16. content_data
15. admin_fields
14. fields
13. admin_datatypes
12. datatypes
11. routes
10. admin_routes
9.  media
8.  tables
7.  sessions
6.  user_oauth
5.  tokens
4.  users
3.  media_dimensions
2.  roles
1.  permissions            (drop last - most depended upon)
```

---

## Circular Dependencies

ModulaCMS schema has **no circular dependencies** (A → B → A). All dependencies are acyclic.

Self-referential tables (datatypes, content_data) are handled with nullable FKs.

---

## Database-Specific Considerations

### SQLite

- Use `PRAGMA foreign_keys = ON` to enforce FK constraints
- FK checks can be temporarily disabled: `PRAGMA foreign_keys = OFF`
- Self-referential FKs work with nullable fields

### MySQL

- InnoDB engine required for FK support
- Self-referential FKs work with nullable fields
- Use `SET FOREIGN_KEY_CHECKS = 0` to temporarily disable (dangerous!)

### PostgreSQL

- FK constraints always enforced
- Self-referential FKs work with nullable fields
- Use `ALTER TABLE ... DISABLE TRIGGER ALL` to temporarily disable (requires superuser)

---

## Verification Query

Check if all tables exist in correct order:

```sql
-- SQLite
SELECT name FROM sqlite_master
WHERE type='table'
ORDER BY name;

-- PostgreSQL
SELECT tablename FROM pg_tables
WHERE schemaname = 'public'
ORDER BY tablename;

-- MySQL
SHOW TABLES;
```

Check FK constraint violations:

```sql
-- SQLite
PRAGMA foreign_key_check;

-- PostgreSQL
-- Run constraint validation per table

-- MySQL
-- Check information_schema.table_constraints
```

---

## Summary

**Critical Order Rules**:

1. ✅ `roles` before `users`
2. ✅ `users` before any table with `author_id`
3. ✅ `datatypes` before `fields`
4. ✅ `admin_datatypes` before `admin_fields`
5. ✅ `routes` before `content_data`
6. ✅ `datatypes` before `content_data`
7. ✅ **`fields` before `content_fields`** ← Currently violated in sql/create_order.md
8. ✅ `content_data` before `content_fields`
9. ✅ Junction tables last (datatypes_fields, admin_datatypes_fields)

**Self-Referential Tables** (nullable FKs):
- datatypes.parent_id
- admin_datatypes.parent_id
- content_data (4 tree pointers)
- admin_content_data (4 tree pointers)

---

---

## Bootstrap Data Requirements

**⚠️ CRITICAL**: After creating foundation tables, you MUST insert bootstrap data for the system to function.

See **[SQL_DIRECTORY.md](../database/SQL_DIRECTORY.md#bootstrap-data-requirements)** for complete bootstrap data requirements including:

1. **permissions** - System admin permission (permission_id = 1)
2. **roles** - System admin role (role_id = 1) and viewer role (role_id = 4)
3. **users** - System admin user (user_id = 1)
4. **routes** - Default home route (route_id = 1) - Recommended
5. **datatypes** - Default page datatype (datatype_id = 1) - Recommended

Without bootstrap data, you'll encounter FK constraint failures when:
- Creating users (requires role_id)
- Creating content (requires author_id, route_id, datatype_id)
- Creating any record with author_id foreign key

**Bootstrap must be inserted sequentially** after each foundation table is created.

---

**Last Updated**: 2026-01-16
**Schema Version**: SQLite (all_schema.sql)

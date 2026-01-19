# Non-Nullable Fields Reference

This document lists all fields that **cannot be NULL** in ModulaCMS database tables. Critical for understanding data integrity constraints and debugging foreign key violations.

> **Note**: All PRIMARY KEY fields are implicitly NOT NULL and are listed for completeness.

---

## permissions

| Field | Type | Default | References | Notes |
|-------|------|---------|------------|-------|
| `permission_id` | INTEGER | AUTO | - | Primary Key |
| `table_id` | INTEGER | - | - | |
| `mode` | INTEGER | - | - | |
| `label` | TEXT | - | - | |

---

## roles

| Field | Type | Default | References | Notes |
|-------|------|---------|------------|-------|
| `role_id` | INTEGER | AUTO | - | Primary Key |
| `label` | TEXT | - | - | UNIQUE |
| `permissions` | TEXT | - | - | UNIQUE |

---

## media_dimensions

| Field | Type | Default | References | Notes |
|-------|------|---------|------------|-------|
| `md_id` | INTEGER | AUTO | - | Primary Key |

**All other fields are nullable.**

---

## users

| Field | Type | Default | References | Notes |
|-------|------|---------|------------|-------|
| `user_id` | INTEGER | AUTO | - | Primary Key |
| `username` | TEXT | - | - | UNIQUE |
| `name` | TEXT | - | - | |
| `email` | TEXT | - | - | |
| `hash` | TEXT | - | - | Password hash |
| `role` | INTEGER | 4 | roles → role_id | Default role: 4 |

---

## admin_routes

| Field | Type | Default | References | Notes |
|-------|------|---------|------------|-------|
| `admin_route_id` | INTEGER | AUTO | - | Primary Key |
| `slug` | TEXT | - | - | UNIQUE |
| `title` | TEXT | - | - | |
| `status` | INTEGER | - | - | |
| `author_id` | INTEGER | 1 | users → user_id | ⚠️ FK: Must exist |

---

## routes

| Field | Type | Default | References | Notes |
|-------|------|---------|------------|-------|
| `route_id` | INTEGER | AUTO | - | Primary Key |
| `slug` | TEXT | - | - | UNIQUE |
| `title` | TEXT | - | - | |
| `status` | INTEGER | - | - | |
| `author_id` | INTEGER | 1 | users → user_id | ⚠️ FK: Must exist |

---

## datatypes

| Field | Type | Default | References | Notes |
|-------|------|---------|------------|-------|
| `datatype_id` | INTEGER | AUTO | - | Primary Key |
| `label` | TEXT | - | - | |
| `type` | TEXT | - | - | |
| `author_id` | INTEGER | 1 | users → user_id | ⚠️ FK: Must exist |

---

## admin_datatypes

| Field | Type | Default | References | Notes |
|-------|------|---------|------------|-------|
| `admin_datatype_id` | INTEGER | AUTO | - | Primary Key |
| `label` | TEXT | - | - | |
| `type` | TEXT | - | - | |
| `author_id` | INTEGER | - | users → user_id | ⚠️ FK: Must exist, NO DEFAULT |

---

## admin_fields

| Field | Type | Default | References | Notes |
|-------|------|---------|------------|-------|
| `admin_field_id` | INTEGER | AUTO | - | Primary Key |
| `label` | TEXT | 'unlabeled' | - | |
| `data` | TEXT | '' | - | Empty string default |
| `type` | TEXT | 'text' | - | |
| `author_id` | INTEGER | 1 | users → user_id | ⚠️ FK: Must exist |

---

## tokens

| Field | Type | Default | References | Notes |
|-------|------|---------|------------|-------|
| `id` | INTEGER | AUTO | - | Primary Key |
| `user_id` | INTEGER | - | users → user_id | ⚠️ FK: Must exist |
| `token_type` | TEXT | - | - | |
| `token` | TEXT | - | - | UNIQUE |
| `issued_at` | TEXT | - | - | |
| `expires_at` | TEXT | - | - | |
| `revoked` | BOOLEAN | 0 | - | Default: false |

---

## user_oauth

| Field | Type | Default | References | Notes |
|-------|------|---------|------------|-------|
| `user_oauth_id` | INTEGER | AUTO | - | Primary Key |
| `user_id` | INTEGER | - | users → user_id | ⚠️ FK: Must exist |
| `oauth_provider` | TEXT | - | - | |
| `oauth_provider_user_id` | TEXT | - | - | |
| `access_token` | TEXT | - | - | |
| `refresh_token` | TEXT | - | - | |
| `token_expires_at` | TEXT | - | - | |
| `date_created` | TEXT | CURRENT_TIMESTAMP | - | |

---

## tables

| Field | Type | Default | References | Notes |
|-------|------|---------|------------|-------|
| `id` | INTEGER | AUTO | - | Primary Key |
| `label` | TEXT | - | - | UNIQUE |
| `author_id` | INTEGER | 1 | users → user_id | ⚠️ FK: Must exist |

---

## media

| Field | Type | Default | References | Notes |
|-------|------|---------|------------|-------|
| `media_id` | INTEGER | AUTO | - | Primary Key |
| `author_id` | INTEGER | 1 | users → user_id | ⚠️ FK: Must exist |

**All other fields are nullable.**

---

## sessions

| Field | Type | Default | References | Notes |
|-------|------|---------|------------|-------|
| `session_id` | INTEGER | AUTO | - | Primary Key |
| `user_id` | INTEGER | - | users → user_id | ⚠️ FK: Must exist |

**All other fields are nullable (with defaults).**

---

## content_data ⚠️ CRITICAL

| Field | Type | Default | References | Notes |
|-------|------|---------|------------|-------|
| `content_data_id` | INTEGER | AUTO | - | Primary Key |
| `route_id` | INTEGER | - | routes → route_id | ⚠️ FK: Must exist, ON DELETE RESTRICT |
| `datatype_id` | INTEGER | - | datatypes → datatype_id | ⚠️ FK: Must exist, ON DELETE RESTRICT |
| `author_id` | INTEGER | - | users → user_id | ⚠️ FK: Must exist, ON DELETE SET DEFAULT |

**Tree pointer fields (parent_id, first_child_id, next_sibling_id, prev_sibling_id) are NULLABLE.**

---

## content_fields ⚠️ CRITICAL

| Field | Type | Default | References | Notes |
|-------|------|---------|------------|-------|
| `content_field_id` | INTEGER | AUTO | - | Primary Key |
| `content_data_id` | INTEGER | - | content_data → content_data_id | ⚠️ FK: Must exist |
| `field_id` | INTEGER | - | fields → field_id | ⚠️ FK: Must exist |
| `field_value` | TEXT | - | - | Cannot be empty |
| `author_id` | INTEGER | - | users → user_id | ⚠️ FK: Must exist |

---

## fields

| Field | Type | Default | References | Notes |
|-------|------|---------|------------|-------|
| `field_id` | INTEGER | AUTO | - | Primary Key |
| `label` | TEXT | 'unlabeled' | - | |
| `data` | TEXT | - | - | |
| `type` | TEXT | - | - | |
| `author_id` | INTEGER | 1 | users → user_id | ⚠️ FK: Must exist |

---

## admin_content_data

| Field | Type | Default | References | Notes |
|-------|------|---------|------------|-------|
| `admin_content_data_id` | INTEGER | AUTO | - | Primary Key |
| `admin_route_id` | INTEGER | - | admin_routes → admin_route_id | ⚠️ FK: Must exist |
| `admin_datatype_id` | INTEGER | - | admin_datatypes → admin_datatype_id | ⚠️ FK: Must exist |
| `author_id` | INTEGER | 1 | users → user_id | ⚠️ FK: Must exist |

**Tree pointer fields (parent_id, first_child_id, next_sibling_id, prev_sibling_id) are NULLABLE.**

---

## admin_content_fields

| Field | Type | Default | References | Notes |
|-------|------|---------|------------|-------|
| `admin_content_field_id` | INTEGER | AUTO | - | Primary Key |
| `admin_content_data_id` | INTEGER | - | admin_content_data → admin_content_data_id | ⚠️ FK: Must exist |
| `admin_field_id` | INTEGER | - | admin_fields → admin_field_id | ⚠️ FK: Must exist |
| `admin_field_value` | TEXT | - | - | Cannot be empty |
| `author_id` | INTEGER | 0 | users → user_id | ⚠️ FK: Must exist, **DEFAULT 0 IS INVALID** |

---

## datatypes_fields

| Field | Type | Default | References | Notes |
|-------|------|---------|------------|-------|
| `id` | INTEGER | AUTO | - | Primary Key |
| `datatype_id` | INTEGER | - | datatypes → datatype_id | ⚠️ FK: Must exist |
| `field_id` | INTEGER | - | fields → field_id | ⚠️ FK: Must exist |

---

## admin_datatypes_fields

| Field | Type | Default | References | Notes |
|-------|------|---------|------------|-------|
| `id` | INTEGER | AUTO | - | Primary Key |
| `admin_datatype_id` | INTEGER | - | admin_datatypes → admin_datatype_id | ⚠️ FK: Must exist |
| `admin_field_id` | INTEGER | - | admin_fields → admin_field_id | ⚠️ FK: Must exist |

---

## Common Patterns & Important Notes

### ⚠️ Foreign Key Constraints

**CRITICAL**: All foreign key fields marked with ⚠️ **MUST** reference an existing record. Attempting to insert a non-existent ID will cause a `FOREIGN KEY constraint failed` error.

### Author ID Requirements

Most tables require `author_id` (references `users.user_id`):

- **Default 1**: `admin_routes`, `routes`, `datatypes`, `admin_fields`, `tables`, `media`, `fields`, `admin_content_data`
- **Default 0** ⚠️ **BUG**: `admin_content_fields` (should be 1, user with ID=0 doesn't exist)
- **NO DEFAULT** ⚠️: `admin_datatypes`, `tokens`, `user_oauth`, `sessions`, `content_data`, `content_fields`

### Tree Structure Fields (NULLABLE)

These fields in `content_data` and `admin_content_data` **ARE NULLABLE**:
- `parent_id`
- `first_child_id`
- `next_sibling_id`
- `prev_sibling_id`

⚠️ **CRITICAL**: When creating new content, these must be set to `sql.NullInt64{}` (NULL), NOT `db.Int64ToNullInt64(0)` (valid 0).

### Content Creation Requirements

To successfully create `content_data`:

1. **Required non-null fields**:
   - `route_id` → Must exist in `routes` table
   - `datatype_id` → Must exist in `datatypes` table
   - `author_id` → Must exist in `users` table

2. **Nullable tree pointers** (set to NULL initially):
   - `parent_id`
   - `first_child_id`
   - `next_sibling_id`
   - `prev_sibling_id`

3. **Default fields** (auto-populated):
   - `date_created` → CURRENT_TIMESTAMP
   - `date_modified` → CURRENT_TIMESTAMP
   - `history` → NULL

### Field Value Creation Requirements

To successfully create `content_fields`:

1. **Required non-null fields**:
   - `content_data_id` → Must exist in `content_data` table
   - `field_id` → Must exist in `fields` table
   - `field_value` → Cannot be empty string (NOT NULL)
   - `author_id` → Must exist in `users` table

2. **Nullable fields**:
   - `route_id` → Optional
   - `history` → NULL

### Known Schema Issues

1. **`admin_content_fields.author_id`** has `DEFAULT 0` but user with ID=0 doesn't exist
   - This will cause FK constraint failures
   - Should be changed to `DEFAULT 1` to match other tables

---

## Debugging Foreign Key Errors

When you encounter `FOREIGN KEY constraint failed`:

1. **Check all NOT NULL fields** are provided
2. **Verify referenced IDs exist** in parent tables:
   ```sql
   SELECT user_id FROM users WHERE user_id = ?;
   SELECT route_id FROM routes WHERE route_id = ?;
   SELECT datatype_id FROM datatypes WHERE datatype_id = ?;
   ```
3. **Check mapping functions** include ALL required fields:
   - SQLite: `MapCreateContentDataParams`
   - MySQL: `MapCreateContentDataParams` (mysql version)
   - PostgreSQL: `MapCreateContentDataParams` (psql version)

4. **Verify NULL handling** for optional FK fields:
   - Use `sql.NullInt64{}` for NULL
   - NOT `db.Int64ToNullInt64(0)` which creates valid 0

---

## Database Setup Requirements

**⚠️ CRITICAL**: Bootstrap data is REQUIRED for system operation.

See **[SQL_DIRECTORY.md](../database/SQL_DIRECTORY.md#bootstrap-data-requirements)** for complete bootstrap requirements.

**Required bootstrap records** (in order):

```sql
-- 1. Required: System admin permission
INSERT INTO permissions (permission_id, table_id, mode, label)
VALUES (1, 0, 7, 'system_admin');

-- 2. Required: System admin role and default viewer role
INSERT INTO roles (role_id, label, permissions)
VALUES (1, 'system_admin', '{"system_admin": true}'),
       (4, 'viewer', '{"read": true}');

-- 3. Required: System admin user (user_id = 1)
INSERT INTO users (user_id, username, name, email, hash, role)
VALUES (1, 'system', 'System Administrator', 'system@modulacms.local', '', 1);

-- 4. Recommended: Default home route
INSERT INTO routes (route_id, slug, title, status, author_id)
VALUES (1, '/', 'Home', 1, 1);

-- 5. Recommended: Default page datatype
INSERT INTO datatypes (datatype_id, label, type, author_id)
VALUES (1, 'Page', 'page', 1);
```

**Why required**:
- Most tables have `author_id` FK to `users.user_id`
- Many tables DEFAULT `author_id = 1` (system user)
- Users table requires valid `role_id` FK
- Content creation requires valid `route_id` and `datatype_id`

Without bootstrap records, you'll encounter FK constraint failures during:
- User creation
- Content creation
- Any operation requiring author_id

---

**Last Updated**: 2026-01-16
**Schema Version**: SQLite (all_schema.sql)

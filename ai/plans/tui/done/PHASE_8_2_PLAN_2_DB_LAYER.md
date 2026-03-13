Plan 2: DB Layer — Missing Admin Content Queries

Context

The admin content DB layer already has most queries needed for tree building and basic CRUD.
This plan adds the missing queries needed for the TUI content editor's field-editing,
tree-loading, and tree-navigation workflows.

The parent plan (phase_8_2.md Sub-Plan A) lists 4 missing queries. This plan covers all 4.

Existing (no work needed):
- ListAdminContentDataWithDatatypeByRoute (schema/22_joins) — tree data with datatype join
- ListAdminContentFieldsWithFieldByRoute (schema/22_joins) — all fields for a route with definitions
- ListAdminContentFieldsByContentDataAndLocale (schema/19) — fields for one content item + locale
  (covers the single-content-item field fetch when locale is known; see Query 2 note below)
- ListAdminFieldsByDatatypeID (admin_field_custom.go) — field definitions by datatype
- All AdminContentVersion CRUD (schema/32) — versions exist with full lifecycle
- CreateAdminContentData, UpdateAdminContentData, DeleteAdminContentData — content CRUD
- CreateAdminContentField, UpdateAdminContentField, DeleteAdminContentField — field CRUD
- UpdateAdminContentDataPublishMeta — publish metadata

Prerequisite: None (independent of Plan 1)

---

Missing Query 1: GetAdminContentDataDescendants

Purpose: Recursive CTE to load all descendants of a content node for tree navigation.
Regular equivalent: GetContentDataDescendants (sql/schema/16_content_data/queries.sql:120-128)

This is needed for subtree operations: expanding a branch, deleting a subtree, or
moving a node with its children.

SQL (all 3 dialects):
  Copy from GetContentDataDescendants, adapt table/column names:
  - content_data -> admin_content_data
  - content_data_id -> admin_content_data_id
  - parent_id stays parent_id (same column name in admin table)
  Placeholder style: SQLite/MySQL use ?, PostgreSQL uses $1

Return type: AdminContentData (existing generated type)

Files to modify:
- sql/schema/18_admin_content_data/queries.sql — add CTE query
- sql/schema/18_admin_content_data/queries_mysql.sql — add CTE query
- sql/schema/18_admin_content_data/queries_psql.sql — add CTE query
- After just sqlc: sqlc generates GetAdminContentDataDescendants in all 3 db packages
- internal/db/db.go — add to DbDriver interface:
    GetAdminContentDataDescendants(context.Context, types.AdminContentID) (*[]AdminContentData, error)
- internal/db/admin_content_data.go — add wrapper methods on Database, MysqlDatabase, PsqlDatabase

---

Missing Query 2: ListAdminContentFieldsByContentData

Purpose: Fetch all admin content fields for a single content item (all locales).
Regular equivalent: ListContentFieldsByContentData (sql/schema/17_content_fields/queries.sql:42-45)

Note: ListAdminContentFieldsByContentDataAndLocale already exists and covers the use case
when locale is known. This non-locale variant is needed when the caller wants all fields
regardless of locale (e.g., tree building, content duplication, bulk operations).

SQL (all 3 dialects):
  SELECT * FROM admin_content_fields
  WHERE admin_content_data_id = ?
  ORDER BY admin_content_field_id;

Return type: AdminContentFields (existing generated type)

Files to modify:
- sql/schema/19_admin_content_fields/queries.sql — add query
- sql/schema/19_admin_content_fields/queries_mysql.sql — add query
- sql/schema/19_admin_content_fields/queries_psql.sql — add query
- After just sqlc: sqlc generates the query function in all 3 db packages
- internal/db/db.go — add to DbDriver interface:
    ListAdminContentFieldsByContentData(types.NullableAdminContentID) (*[]AdminContentFields, error)
- internal/db/admin_content_field.go — add wrapper methods on Database, MysqlDatabase, PsqlDatabase

---

Missing Query 3: ListAdminContentFieldsWithFieldByContentData

Purpose: Fetch admin content fields joined with field definitions for a single content item.
Regular equivalent: ListContentFieldsWithFieldByContentData (schema/22_joins, db.go:273)

This is needed by the edit dialog to display field values alongside their type, label, and
validation config.

SQL (all 3 dialects):
  Identical to ListAdminContentFieldsWithFieldByRoute but filtered by
  acf.admin_content_data_id instead of acf.admin_route_id.

Return type: AdminContentFieldsWithFieldRow (already defined in getTree.go:125-144)

Note on sqlc-generated types: sqlc will generate a new Go type per query name
(ListAdminContentFieldsWithFieldByContentDataRow) that is structurally identical to
ListAdminContentFieldsWithFieldByRouteRow but is a distinct type. Each database struct
needs both a mapper (converting the sqlc type to the wrapper type) and a wrapper method
(calling sqlc and mapping). That is 3 mapper methods + 3 wrapper methods = 6 new methods.
Follow the existing MapAdminContentFieldsWithFieldRow pattern at getTree.go:289-337 (SQLite),
569-618 (MySQL), 850-900 (PostgreSQL).

Files to modify:
- sql/schema/22_joins/queries.sql — add query
- sql/schema/22_joins/queries_mysql.sql — add query
- sql/schema/22_joins/queries_psql.sql — add query
- After just sqlc: sqlc generates the query function in all 3 db packages
- internal/db/db.go — add to DbDriver interface:
    ListAdminContentFieldsWithFieldByContentData(types.NullableAdminContentID) (*[]AdminContentFieldsWithFieldRow, error)
- internal/db/getTree.go — add 3 mapper methods + 3 wrapper methods on Database,
  MysqlDatabase, PsqlDatabase (6 methods total, following the ByRoute pattern)

---

Missing Query 4: ListAdminContentFieldsByContentDataIDs (batch)

Purpose: Batch-load admin content fields for multiple content IDs at once.
Regular equivalent: ListContentFieldsByContentDataIDs (db.go:275, content_field_batch.go)

This is used during tree building to attach fields to nodes. The regular version uses
raw SQL with IN clause construction (not sqlc-generated).

Note: This query uses raw SQL (not sqlc) because sqlc does not support dynamic IN clauses.
Follow the exact pattern in content_field_batch.go:
  - listContentFieldsByContentDataIDs() shared helper with dialect parameter
  - Wrapper methods on Database, MysqlDatabase, PsqlDatabase

Row mapper requirement: The batch query needs a rowToAdminContentFields function to convert
raw QSelect Row (map[string]any) into AdminContentFields structs. This does not exist today.
It must be written in admin_content_field_batch.go alongside the batch query. The admin
columns differ from regular columns:
  - admin_content_field_id (AdminContentFieldID)
  - admin_route_id (NullableAdminRouteID)
  - admin_content_data_id (NullableAdminContentID)
  - admin_field_id (NullableAdminFieldID)
  - admin_field_value (string)
  - locale (string)
  - author_id (UserID — shared, not admin-specific)
  - date_created, date_modified (Timestamp — shared)

Helper functions needed: The existing rowNullable* helpers in content_field_batch.go cover
ContentID, RouteID, and FieldID. The admin batch needs:
  - rowNullableAdminContentID(row Row, key string) types.NullableAdminContentID
  - rowNullableAdminRouteID(row Row, key string) types.NullableAdminRouteID
  - rowNullableAdminFieldID(row Row, key string) types.NullableAdminFieldID
These follow the identical pattern (extract string via rowString, return empty if "", else
return valid NullableX). The shared helpers (rowString, rowTimestamp) are already in
content_field_batch.go and can be called from the new file since they are package-level.

Files to modify:
- internal/db/admin_content_field_batch.go (new file) containing:
    - rowNullableAdminContentID, rowNullableAdminRouteID, rowNullableAdminFieldID (~18 lines)
    - rowToAdminContentFields (~15 lines)
    - listAdminContentFieldsByContentDataIDs shared helper (~35 lines)
    - 3 wrapper methods on Database, MysqlDatabase, PsqlDatabase (~9 lines)
- internal/db/db.go — add to DbDriver interface:
    ListAdminContentFieldsByContentDataIDs(context.Context, []types.AdminContentID, string) (*[]AdminContentFields, error)

---

Steps

1.  Write GetAdminContentDataDescendants CTE for all 3 SQL dialects (schema/18)
2.  Write ListAdminContentFieldsByContentData for all 3 SQL dialects (schema/19)
3.  Write ListAdminContentFieldsWithFieldByContentData for all 3 SQL dialects (schema/22_joins)
4.  Run just sqlc to regenerate
5.  Add all 4 methods to DbDriver interface in db.go
6.  Implement GetAdminContentDataDescendants wrappers on all 3 database structs in admin_content_data.go
7.  Implement ListAdminContentFieldsByContentData wrappers on all 3 database structs in admin_content_field.go
8.  Implement ListAdminContentFieldsWithFieldByContentData mappers + wrappers (6 methods) in getTree.go
9.  Create admin_content_field_batch.go: row helpers, rowToAdminContentFields, shared batch
    helper, 3 wrapper methods
10. Verify: just check
11. Verify: go test ./internal/db/...

---

Size estimate: ~250 lines new code.
- 9 SQL query files (3 queries x 3 dialects): ~50 lines
- GetAdminContentDataDescendants wrappers: ~30 lines
- ListAdminContentFieldsByContentData wrappers: ~30 lines
- ListAdminContentFieldsWithFieldByContentData mappers + wrappers (6 methods): ~60 lines
- admin_content_field_batch.go (helpers + mapper + batch + wrappers): ~80 lines

---

Risk assessment

Low risk. All 4 queries are mechanical ports of existing regular content equivalents:
- The CTE query is a direct table/column name substitution.
- The simple field query is a one-line WHERE clause change.
- The join query reuses the existing AdminContentFieldsWithFieldRow wrapper type.
- The batch query follows the proven raw-SQL-with-dialect pattern from content_field_batch.go.

The one area requiring attention is the rowToAdminContentFields mapper and its 3 nullable
ID helpers. These are straightforward (identical pattern to existing helpers, different types)
but must correctly map the admin column names. A typo in a column name key would silently
return zero-value fields. The test step (go test ./internal/db/...) will not catch this
since the batch query requires a live database. Manual verification against the schema column
names is the primary safeguard — compare admin_content_fields CREATE TABLE columns against
the rowToAdminContentFields key strings.

Error handling note: The batch query helper returns the tree even when orphan fields exist
(fields referencing content IDs not in the IN list). Callers should treat a non-nil error
as informational when the result slice is also non-nil, consistent with the regular batch
query's behavior.

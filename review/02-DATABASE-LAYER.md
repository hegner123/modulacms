# Database Layer Review

## Schema Design: Strong

27 numbered schema directories, each containing 6 files (DDL + queries for SQLite, MySQL, PostgreSQL). Well-normalized. Proper foreign keys with appropriate ON DELETE policies (CASCADE, SET NULL, RESTRICT). ULID primary keys with length constraints. Composite indexes on junction tables. Content tree uses sibling pointers for O(1) navigation and reordering.

### What Is Good

**Schema organization** is excellent. Numbered directories create a clear creation order. Each directory is self-contained with all three SQL dialects. The `sql/schema/` directory is the definitive source of truth, and sqlc generates Go code from it.

**Content tree design** using `parent_id`, `first_child_id`, `next_sibling_id`, `prev_sibling_id` is a smart choice. It enables O(1) insertion, deletion, and reordering without touching other rows. Most CMS systems use `sort_order` integers that require renumbering.

**Change events table** has proper indexes: `(table_name, record_id)` for entity history, `hlc_timestamp` for distributed ordering, `node_id` for multi-node replication, `user_id` for user activity. The schema is ready for distributed deployment even though the app currently runs single-node.

**The audited command pattern** wraps mutations and audit records in a single transaction. If audit recording fails, the mutation rolls back. Before-hooks can abort operations. After-hooks fire asynchronously. This is a solid foundation for compliance.

### What Needs Attention

**Missing indexes on frequently-queried foreign keys:**
- `content_fields.content_data_id` - no index, but filtered in most content queries
- `content_fields.field_id` - no index
- These will cause full table scans on large datasets

**Nullable author_id** on `content_data` and `datatypes` tables. All content should have a creator. This forces defensive NULL checks throughout the application code for a field that should logically never be NULL.

**Inconsistent timestamp column naming**: Some tables use `date_created`/`date_modified`, others use `created_at`/`updated_at`, the change_events table uses `wall_timestamp`/`synced_at`. Pick one convention.

**N+1 query risks**: `GetContentFieldsByRoute` returns content fields without loading field definitions. No batch fetch methods like `GetFieldsByIDs`. Complex tree queries join 5 tables without EXPLAIN analysis.

### The Type Conversion Layer

`convert.go` has 150+ conversion functions. This is necessary given SQLite's int64 vs MySQL's int32 width differences and the custom Nullable types. It's verbose but correct. The sqlc.yml type overrides (250+ entries) ensure the generated code uses the right branded types from the start.

### Test Coverage

79 test files, ~40,000 lines. Comprehensive coverage of CRUD operations, pagination, transactions, and audit trails. Table-driven tests throughout. The db package has the highest test coverage in the entire project.

## Recommendations

1. Add indexes on `content_fields(content_data_id)` and `content_fields(field_id)`
2. Make `author_id` NOT NULL on content_data and datatypes (with migration)
3. Standardize timestamp column names across all tables
4. Add batch fetch methods (e.g., `GetFieldsByIDs([]FieldID)`) to reduce N+1 queries
5. Consider generating the wrapper layer code instead of hand-writing it

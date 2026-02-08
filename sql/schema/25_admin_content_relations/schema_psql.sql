CREATE TABLE IF NOT EXISTS admin_content_relations (
    admin_content_relation_id TEXT PRIMARY KEY NOT NULL CHECK (length(admin_content_relation_id) = 26),
    -- holds admin_content_data_id, named for code symmetry with content_relations
    source_content_id TEXT NOT NULL
        REFERENCES admin_content_data(admin_content_data_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    -- holds admin_content_data_id, named for code symmetry with content_relations
    target_content_id TEXT NOT NULL
        REFERENCES admin_content_data(admin_content_data_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    admin_field_id TEXT NOT NULL
        REFERENCES admin_fields(admin_field_id)
            ON UPDATE CASCADE ON DELETE RESTRICT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (source_content_id != target_content_id)
);

-- Unique constraint ordered to also serve ListBySourceAndField (prefix: source, field)
-- and ListBySource (prefix: source) queries
CREATE UNIQUE INDEX IF NOT EXISTS idx_admin_content_relations_unique
    ON admin_content_relations(source_content_id, admin_field_id, target_content_id);
-- Composite index for ListByTarget ORDER BY date_created
CREATE INDEX IF NOT EXISTS idx_admin_content_relations_target
    ON admin_content_relations(target_content_id, date_created);
-- Supports ON DELETE RESTRICT FK checks when deleting a field
CREATE INDEX IF NOT EXISTS idx_admin_content_relations_field
    ON admin_content_relations(admin_field_id);

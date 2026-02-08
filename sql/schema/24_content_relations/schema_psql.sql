CREATE TABLE IF NOT EXISTS content_relations (
    content_relation_id TEXT PRIMARY KEY NOT NULL CHECK (length(content_relation_id) = 26),
    source_content_id TEXT NOT NULL
        REFERENCES content_data(content_data_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    target_content_id TEXT NOT NULL
        REFERENCES content_data(content_data_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    field_id TEXT NOT NULL
        REFERENCES fields(field_id)
            ON UPDATE CASCADE ON DELETE RESTRICT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (source_content_id != target_content_id)
);

-- Unique constraint ordered to also serve ListBySourceAndField (prefix: source, field)
-- and ListBySource (prefix: source) queries
CREATE UNIQUE INDEX IF NOT EXISTS idx_content_relations_unique
    ON content_relations(source_content_id, field_id, target_content_id);
-- Composite index for ListByTarget ORDER BY date_created
CREATE INDEX IF NOT EXISTS idx_content_relations_target
    ON content_relations(target_content_id, date_created);
-- Supports ON DELETE RESTRICT FK checks when deleting a field
CREATE INDEX IF NOT EXISTS idx_content_relations_field
    ON content_relations(field_id);

-- name: CreateContentRelationTable :exec
CREATE TABLE IF NOT EXISTS content_relations (
    content_relation_id TEXT PRIMARY KEY NOT NULL CHECK (length(content_relation_id) = 26),
    source_content_id TEXT NOT NULL
        REFERENCES content_data(content_data_id)
            ON DELETE CASCADE,
    target_content_id TEXT NOT NULL
        REFERENCES content_data(content_data_id)
            ON DELETE CASCADE,
    field_id TEXT NOT NULL
        REFERENCES fields(field_id)
            ON DELETE RESTRICT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    date_created TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (source_content_id != target_content_id)
);

-- name: DropContentRelationTable :exec
DROP TABLE IF EXISTS content_relations;

-- name: CountContentRelation :one
SELECT COUNT(*) FROM content_relations;

-- name: GetContentRelation :one
SELECT * FROM content_relations
WHERE content_relation_id = ? LIMIT 1;

-- name: CreateContentRelation :one
INSERT INTO content_relations (
    content_relation_id,
    source_content_id,
    target_content_id,
    field_id,
    sort_order,
    date_created
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
) RETURNING *;

-- name: DeleteContentRelation :exec
DELETE FROM content_relations
WHERE content_relation_id = ?;

-- name: UpdateContentRelationSortOrder :exec
UPDATE content_relations
SET sort_order = ?
WHERE content_relation_id = ?;

-- name: ListContentRelationsBySource :many
SELECT * FROM content_relations
WHERE source_content_id = ?
ORDER BY sort_order;

-- name: ListContentRelationsByTarget :many
SELECT * FROM content_relations
WHERE target_content_id = ?
ORDER BY date_created;

-- name: ListContentRelationsBySourceAndField :many
SELECT * FROM content_relations
WHERE source_content_id = ? AND field_id = ?
ORDER BY sort_order;

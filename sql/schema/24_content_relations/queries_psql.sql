-- name: CreateContentRelationTable :exec
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

-- name: DropContentRelationTable :exec
DROP TABLE IF EXISTS content_relations;

-- name: CountContentRelation :one
SELECT COUNT(*) FROM content_relations;

-- name: GetContentRelation :one
SELECT * FROM content_relations
WHERE content_relation_id = $1 LIMIT 1;

-- name: CreateContentRelation :one
INSERT INTO content_relations (
    content_relation_id,
    source_content_id,
    target_content_id,
    field_id,
    sort_order,
    date_created
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
) RETURNING *;

-- name: DeleteContentRelation :exec
DELETE FROM content_relations
WHERE content_relation_id = $1;

-- name: UpdateContentRelationSortOrder :exec
UPDATE content_relations
SET sort_order = $1
WHERE content_relation_id = $2;

-- name: ListContentRelationsBySource :many
SELECT * FROM content_relations
WHERE source_content_id = $1
ORDER BY sort_order;

-- name: ListContentRelationsByTarget :many
SELECT * FROM content_relations
WHERE target_content_id = $1
ORDER BY date_created;

-- name: ListContentRelationsBySourceAndField :many
SELECT * FROM content_relations
WHERE source_content_id = $1 AND field_id = $2
ORDER BY sort_order;

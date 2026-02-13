-- name: CreateContentRelationTable :exec
CREATE TABLE IF NOT EXISTS content_relations (
    content_relation_id VARCHAR(26) NOT NULL,
    source_content_id VARCHAR(26) NOT NULL,
    target_content_id VARCHAR(26) NOT NULL,
    field_id VARCHAR(26) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (content_relation_id),
    CONSTRAINT fk_content_relations_source FOREIGN KEY (source_content_id)
        REFERENCES content_data(content_data_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_content_relations_target FOREIGN KEY (target_content_id)
        REFERENCES content_data(content_data_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_content_relations_field FOREIGN KEY (field_id)
        REFERENCES fields(field_id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT uq_content_relations_unique UNIQUE (source_content_id, field_id, target_content_id)
);

-- name: DropContentRelationTable :exec
DROP TABLE IF EXISTS content_relations;

-- name: CountContentRelation :one
SELECT COUNT(*) FROM content_relations;

-- name: GetContentRelation :one
SELECT * FROM content_relations
WHERE content_relation_id = ? LIMIT 1;

-- name: CreateContentRelation :exec
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
);

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

-- name: CreateAdminContentRelationTable :exec
CREATE TABLE IF NOT EXISTS admin_content_relations (
    admin_content_relation_id TEXT PRIMARY KEY NOT NULL CHECK (length(admin_content_relation_id) = 26),
    -- holds admin_content_data_id, named for code symmetry with content_relations
    source_content_id TEXT NOT NULL
        REFERENCES admin_content_data(admin_content_data_id)
            ON DELETE CASCADE,
    -- holds admin_content_data_id, named for code symmetry with content_relations
    target_content_id TEXT NOT NULL
        REFERENCES admin_content_data(admin_content_data_id)
            ON DELETE CASCADE,
    admin_field_id TEXT NOT NULL
        REFERENCES admin_fields(admin_field_id)
            ON DELETE RESTRICT,
    sort_order INTEGER NOT NULL DEFAULT 0,
    date_created TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CHECK (source_content_id != target_content_id)
);

-- name: DropAdminContentRelationTable :exec
DROP TABLE IF EXISTS admin_content_relations;

-- name: CountAdminContentRelation :one
SELECT COUNT(*) FROM admin_content_relations;

-- name: GetAdminContentRelation :one
SELECT * FROM admin_content_relations
WHERE admin_content_relation_id = ? LIMIT 1;

-- name: CreateAdminContentRelation :one
INSERT INTO admin_content_relations (
    admin_content_relation_id,
    source_content_id,
    target_content_id,
    admin_field_id,
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

-- name: DeleteAdminContentRelation :exec
DELETE FROM admin_content_relations
WHERE admin_content_relation_id = ?;

-- name: UpdateAdminContentRelationSortOrder :exec
UPDATE admin_content_relations
SET sort_order = ?
WHERE admin_content_relation_id = ?;

-- name: ListAdminContentRelationsBySource :many
SELECT * FROM admin_content_relations
WHERE source_content_id = ?
ORDER BY sort_order;

-- name: ListAdminContentRelationsByTarget :many
SELECT * FROM admin_content_relations
WHERE target_content_id = ?
ORDER BY date_created;

-- name: ListAdminContentRelationsBySourceAndField :many
SELECT * FROM admin_content_relations
WHERE source_content_id = ? AND admin_field_id = ?
ORDER BY sort_order;

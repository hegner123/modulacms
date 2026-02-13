-- name: CreateAdminContentRelationTable :exec
CREATE TABLE IF NOT EXISTS admin_content_relations (
    admin_content_relation_id VARCHAR(26) NOT NULL,
    -- holds admin_content_data_id, named for code symmetry with content_relations
    source_content_id VARCHAR(26) NOT NULL,
    -- holds admin_content_data_id, named for code symmetry with content_relations
    target_content_id VARCHAR(26) NOT NULL,
    admin_field_id VARCHAR(26) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    date_created TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (admin_content_relation_id),
    CONSTRAINT fk_admin_content_relations_source FOREIGN KEY (source_content_id)
        REFERENCES admin_content_data(admin_content_data_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_relations_target FOREIGN KEY (target_content_id)
        REFERENCES admin_content_data(admin_content_data_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_admin_content_relations_field FOREIGN KEY (admin_field_id)
        REFERENCES admin_fields(admin_field_id)
        ON UPDATE CASCADE ON DELETE RESTRICT,
    CONSTRAINT uq_admin_content_relations_unique UNIQUE (source_content_id, admin_field_id, target_content_id)
);

-- name: DropAdminContentRelationTable :exec
DROP TABLE IF EXISTS admin_content_relations;

-- name: CountAdminContentRelation :one
SELECT COUNT(*) FROM admin_content_relations;

-- name: GetAdminContentRelation :one
SELECT * FROM admin_content_relations
WHERE admin_content_relation_id = ? LIMIT 1;

-- name: CreateAdminContentRelation :exec
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
);

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

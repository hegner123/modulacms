-- name: DropAdminContentDataTable :exec
DROP TABLE admin_content_data;

-- name: CreateAdminContentDataTable :exec
CREATE TABLE IF NOT EXISTS admin_content_data (
    admin_content_data_id TEXT PRIMARY KEY NOT NULL,
    parent_id TEXT
        CONSTRAINT fk_parent_id
            REFERENCES admin_content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    first_child_id TEXT
        CONSTRAINT fk_first_child_id
            REFERENCES admin_content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    next_sibling_id TEXT
        CONSTRAINT fk_next_sibling_id
            REFERENCES admin_content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    prev_sibling_id TEXT
        CONSTRAINT fk_prev_sibling_id
            REFERENCES admin_content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    admin_route_id TEXT NOT NULL
        CONSTRAINT fk_admin_routes
            REFERENCES admin_routes
            ON UPDATE CASCADE,
    admin_datatype_id TEXT NOT NULL
        CONSTRAINT fk_admin_datatypes
            REFERENCES admin_datatypes
            ON UPDATE CASCADE,
    author_id TEXT NOT NULL
        CONSTRAINT fk_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'draft',
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMP,
    published_by TEXT
        CONSTRAINT fk_admin_published_by
            REFERENCES users
            ON UPDATE CASCADE ON DELETE SET NULL,
    publish_at TIMESTAMP,
    revision INTEGER NOT NULL DEFAULT 0
);


-- name: CountAdminContentData :one
SELECT COUNT(*)
FROM admin_content_data;

-- name: GetAdminContentData :one
SELECT * FROM admin_content_data
WHERE admin_content_data_id = $1 LIMIT 1;

-- name: ListAdminContentData :many
SELECT * FROM admin_content_data
ORDER BY admin_content_data_id;

-- name: ListAdminContentDataByRoute :many
SELECT * FROM admin_content_data
WHERE admin_route_id = $1
ORDER BY admin_content_data_id;

-- name: CreateAdminContentData :one
INSERT INTO admin_content_data (
    admin_content_data_id,
    parent_id,
    first_child_id,
    next_sibling_id,
    prev_sibling_id,
    admin_route_id,
    admin_datatype_id,
    author_id,
    status,
    date_created,
    date_modified
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    $8,
    $9,
    $10,
    $11
) RETURNING *;

-- name: UpdateAdminContentData :exec
UPDATE admin_content_data
SET parent_id = $1,
    first_child_id = $2,
    next_sibling_id = $3,
    prev_sibling_id = $4,
    admin_route_id = $5,
    admin_datatype_id = $6,
    author_id = $7,
    status = $8,
    date_created = $9,
    date_modified = $10
WHERE admin_content_data_id = $11;

-- name: DeleteAdminContentData :exec
DELETE FROM admin_content_data
WHERE admin_content_data_id = $1;

-- name: ListAdminContentDataPaginated :many
SELECT * FROM admin_content_data
ORDER BY admin_content_data_id
LIMIT $1 OFFSET $2;

-- name: ListAdminContentDataByRoutePaginated :many
SELECT * FROM admin_content_data
WHERE admin_route_id = $1
ORDER BY admin_content_data_id
LIMIT $2 OFFSET $3;

-- name: ListAdminContentDataTopLevelPaginated :many
SELECT acd.*, u.name AS author_name, COALESCE(ar.slug, '') AS route_slug, COALESCE(ar.title, '') AS route_title, COALESCE(adt.label, '') AS datatype_label FROM admin_content_data acd
LEFT JOIN admin_datatypes adt ON acd.admin_datatype_id = adt.admin_datatype_id
LEFT JOIN users u ON acd.author_id = u.user_id
LEFT JOIN admin_routes ar ON acd.admin_route_id = ar.admin_route_id
WHERE acd.admin_route_id IS NOT NULL OR adt.type = '_root'
ORDER BY acd.admin_content_data_id
LIMIT $1 OFFSET $2;

-- name: CountAdminContentDataTopLevel :one
SELECT COUNT(*) FROM admin_content_data acd
LEFT JOIN admin_datatypes adt ON acd.admin_datatype_id = adt.admin_datatype_id
WHERE acd.admin_route_id IS NOT NULL OR adt.type = '_root';

-- name: UpdateAdminContentDataPublishMeta :exec
UPDATE admin_content_data
SET status = $1,
    published_at = $2,
    published_by = $3,
    revision = revision + 1,
    date_modified = $4
WHERE admin_content_data_id = $5;

-- name: UpdateAdminContentDataWithRevision :exec
UPDATE admin_content_data
SET admin_route_id = $1,
    parent_id = $2,
    first_child_id = $3,
    next_sibling_id = $4,
    prev_sibling_id = $5,
    admin_datatype_id = $6,
    author_id = $7,
    status = $8,
    date_created = $9,
    date_modified = $10
WHERE admin_content_data_id = $11 AND revision = $12;

-- name: UpdateAdminContentDataSchedule :exec
UPDATE admin_content_data
SET publish_at = $1,
    date_modified = $2
WHERE admin_content_data_id = $3;

-- name: ClearAdminContentDataSchedule :exec
UPDATE admin_content_data
SET publish_at = NULL,
    date_modified = $1
WHERE admin_content_data_id = $2;

-- name: ListAdminContentDataDueForPublish :many
SELECT * FROM admin_content_data
WHERE publish_at IS NOT NULL AND publish_at <= $1 AND status = 'draft';

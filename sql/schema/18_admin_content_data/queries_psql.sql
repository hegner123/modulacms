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
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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

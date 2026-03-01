-- name: DropAdminContentFieldTable :exec
DROP TABLE admin_content_fields;

-- name: CreateAdminContentFieldTable :exec
CREATE TABLE IF NOT EXISTS admin_content_fields (
    admin_content_field_id TEXT PRIMARY KEY NOT NULL,
    admin_route_id TEXT
        CONSTRAINT fk_admin_route_id
            REFERENCES admin_routes
            ON UPDATE CASCADE ON DELETE SET NULL,
    admin_content_data_id TEXT NOT NULL
        CONSTRAINT fk_admin_content_data
            REFERENCES admin_content_data
            ON UPDATE CASCADE ON DELETE CASCADE,
    admin_field_id TEXT NOT NULL
        CONSTRAINT fk_admin_fields
            REFERENCES admin_fields
            ON UPDATE CASCADE ON DELETE CASCADE,
    admin_field_value TEXT NOT NULL,
    locale TEXT NOT NULL DEFAULT '',
    author_id TEXT NOT NULL
        CONSTRAINT fk_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE CASCADE,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- name: CountAdminContentField :one
SELECT COUNT(*)
FROM admin_content_fields;

-- name: GetAdminContentField :one
SELECT * FROM admin_content_fields
WHERE admin_content_field_id = $1 LIMIT 1;

-- name: ListAdminContentFields :many
SELECT * FROM admin_content_fields
ORDER BY admin_content_field_id;

-- name: ListAdminContentFieldsByRoute :many
SELECT * FROM admin_content_fields
WHERE admin_route_id = $1
ORDER BY admin_content_field_id;

-- name: CreateAdminContentField :one
INSERT INTO admin_content_fields (
    admin_content_field_id,
    admin_route_id,
    admin_content_data_id,
    admin_field_id,
    admin_field_value,
    locale,
    author_id,
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
    $9
) RETURNING *;

-- name: UpdateAdminContentField :exec
UPDATE admin_content_fields
SET admin_route_id=$1,
    admin_content_data_id=$2,
    admin_field_id=$3,
    admin_field_value=$4,
    locale=$5,
    author_id=$6,
    date_created=$7,
    date_modified=$8
WHERE admin_content_field_id = $9;

-- name: DeleteAdminContentField :exec
DELETE FROM admin_content_fields
WHERE admin_content_field_id = $1;

-- name: ListAdminContentFieldsPaginated :many
SELECT * FROM admin_content_fields
ORDER BY admin_content_field_id
LIMIT $1 OFFSET $2;

-- name: ListAdminContentFieldsByRoutePaginated :many
SELECT * FROM admin_content_fields
WHERE admin_route_id = $1
ORDER BY admin_content_field_id
LIMIT $2 OFFSET $3;

-- name: ListAdminContentFieldsByContentDataAndLocale :many
SELECT * FROM admin_content_fields
WHERE admin_content_data_id = $1 AND locale IN ($2, '')
ORDER BY admin_content_field_id;

-- name: ListAdminContentFieldsByRouteAndLocale :many
SELECT * FROM admin_content_fields
WHERE admin_route_id = $1 AND locale IN ($2, '')
ORDER BY admin_content_data_id, admin_field_id;

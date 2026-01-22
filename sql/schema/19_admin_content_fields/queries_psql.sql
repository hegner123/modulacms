-- name: DropAdminContentFieldTable :exec
DROP TABLE admin_content_fields;

-- name: CreateAdminContentFieldTable :exec
CREATE TABLE IF NOT EXISTS admin_content_fields (
    admin_content_field_id SERIAL
        PRIMARY KEY,
    admin_route_id INTEGER
        CONSTRAINT fk_admin_route_id
            REFERENCES admin_routes
            ON UPDATE CASCADE ON DELETE SET NULL,
    admin_content_data_id INTEGER NOT NULL
        CONSTRAINT fk_admin_content_data
            REFERENCES admin_content_data
            ON UPDATE CASCADE ON DELETE CASCADE,
    admin_field_id INTEGER NOT NULL
        CONSTRAINT fk_admin_fields
            REFERENCES admin_fields
            ON UPDATE CASCADE ON DELETE CASCADE,
    admin_field_value TEXT NOT NULL,
    author_id INTEGER NOT NULL
        CONSTRAINT fk_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
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
    admin_route_id,
    admin_content_data_id,
    admin_field_id,
    admin_field_value,
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
    $7
) RETURNING *;

-- name: UpdateAdminContentField :exec
UPDATE admin_content_fields
SET admin_route_id=$1,
    admin_content_data_id=$2,
    admin_field_id=$3,
    admin_field_value=$4,
    author_id=$5,
    date_created=$6,
    date_modified=$7
WHERE admin_content_field_id = $8;

-- name: DeleteAdminContentField :exec
DELETE FROM admin_content_fields
WHERE admin_content_field_id = $1;

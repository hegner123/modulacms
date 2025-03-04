-- name: CreateAdminFieldTable :exec
CREATE TABLE IF NOT EXISTS admin_fields (
    admin_field_id SERIAL PRIMARY KEY,
    admin_route_id INTEGER DEFAULT 1
        REFERENCES admin_routes(admin_route_id)
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    parent_id INTEGER DEFAULT NULL
        REFERENCES admin_datatypes
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    label TEXT NOT NULL DEFAULT 'unlabeled',
    data TEXT NOT NULL DEFAULT '',
    type TEXT NOT NULL DEFAULT 'text',
    author TEXT NOT NULL DEFAULT 'system'
        REFERENCES users(username)
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    author_id INTEGER NOT NULL DEFAULT 1
        REFERENCES users(user_id)
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);

-- name: GetAdminField :one
SELECT *
FROM admin_fields
WHERE admin_field_id = $1
LIMIT 1;

-- name: CountAdminField :one
SELECT COUNT(*)
FROM admin_fields;

-- name: GetAdminFieldId :one
SELECT admin_field_id
FROM admin_fields
WHERE admin_field_id = $1
LIMIT 1;

-- name: ListAdminField :many
SELECT *
FROM admin_fields
ORDER BY admin_field_id;

-- name: CreateAdminField :one
INSERT INTO admin_fields (
    admin_route_id,
    parent_id,
    label,
    data,
    type,
    author,
    author_id,
    date_created,
    date_modified,
    history
)
VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: UpdateAdminField :exec
UPDATE admin_fields
SET admin_route_id = $1,
    parent_id    = $2,
    label        = $3,
    data         = $4,
    type         = $5,
    author       = $6,
    author_id    = $7,
    date_created = $8,
    date_modified= $9,
    history      = $10
WHERE admin_field_id = $11;

-- name: DeleteAdminField :exec
DELETE FROM admin_fields
WHERE admin_field_id = $1;

-- name: ListAdminFieldByRouteId :many
SELECT admin_field_id,
       admin_route_id,
       parent_id,
       label,
       data,
       type,
       history
FROM admin_fields
WHERE admin_route_id = $1;

-- name: ListAdminFieldsByDatatypeID :many
SELECT admin_field_id,
       admin_route_id,
       parent_id,
       label,
       data,
       type,
       history
FROM admin_fields
WHERE parent_id = $1;


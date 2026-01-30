-- name: DropAdminFieldTable :exec
DROP TABLE admin_fields;

-- name: CreateAdminFieldTable :exec
CREATE TABLE IF NOT EXISTS admin_fields (
    admin_field_id TEXT PRIMARY KEY NOT NULL,
    parent_id TEXT
        REFERENCES admin_datatypes
            ON UPDATE CASCADE ON DELETE SET NULL,
    label TEXT DEFAULT 'unlabeled'::TEXT NOT NULL,
    data TEXT DEFAULT ''::TEXT NOT NULL,
    type TEXT DEFAULT 'text'::TEXT NOT NULL,
    author_id TEXT NOT NULL
        REFERENCES users
            ON UPDATE CASCADE ON DELETE SET NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- name: CountAdminField :one
SELECT COUNT(*)
FROM admin_fields;

-- name: GetAdminField :one
SELECT *
FROM admin_fields
WHERE admin_field_id = $1
LIMIT 1;

-- name: ListAdminField :many
SELECT *
FROM admin_fields
ORDER BY admin_field_id;

-- name: ListAdminFieldByParentID :many
SELECT *
FROM admin_fields
WHERE parent_id = $1
ORDER BY admin_field_id;


-- name: CreateAdminField :one
INSERT INTO admin_fields (
    admin_field_id,
    parent_id,
    label,
    data,
    type,
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
    $8
)
RETURNING *;

-- name: UpdateAdminField :exec
UPDATE admin_fields
SET parent_id    = $1,
    label        = $2,
    data         = $3,
    type         = $4,
    author_id    = $5,
    date_created = $6,
    date_modified= $7
WHERE admin_field_id = $8;

-- name: DeleteAdminField :exec
DELETE FROM admin_fields
WHERE admin_field_id = $1;

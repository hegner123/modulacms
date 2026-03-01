-- name: DropAdminFieldTable :exec
DROP TABLE admin_fields;

-- name: CreateAdminFieldTable :exec
CREATE TABLE IF NOT EXISTS admin_fields (
    admin_field_id TEXT PRIMARY KEY NOT NULL,
    parent_id TEXT
        REFERENCES admin_datatypes
            ON UPDATE CASCADE ON DELETE CASCADE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    name TEXT NOT NULL DEFAULT '',
    label TEXT DEFAULT 'unlabeled'::TEXT NOT NULL,
    data TEXT DEFAULT ''::TEXT NOT NULL,
    validation TEXT NOT NULL,
    ui_config TEXT NOT NULL,
    type TEXT DEFAULT 'text'::TEXT NOT NULL,
    translatable BOOLEAN NOT NULL DEFAULT FALSE,
    roles TEXT DEFAULT NULL,
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
ORDER BY sort_order, admin_field_id;

-- name: ListAdminFieldByParentID :many
SELECT *
FROM admin_fields
WHERE parent_id = $1
ORDER BY sort_order, admin_field_id;


-- name: CreateAdminField :one
INSERT INTO admin_fields (
    admin_field_id,
    parent_id,
    sort_order,
    name,
    label,
    data,
    validation,
    ui_config,
    type,
    translatable,
    roles,
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
    $9,
    $10,
    $11,
    $12,
    $13,
    $14
)
RETURNING *;

-- name: UpdateAdminField :exec
UPDATE admin_fields
SET parent_id    = $1,
    sort_order   = $2,
    name         = $3,
    label        = $4,
    data         = $5,
    validation   = $6,
    ui_config    = $7,
    type         = $8,
    translatable = $9,
    roles        = $10,
    author_id    = $11,
    date_created = $12,
    date_modified= $13
WHERE admin_field_id = $14;

-- name: DeleteAdminField :exec
DELETE FROM admin_fields
WHERE admin_field_id = $1;

-- name: ListAdminFieldPaginated :many
SELECT * FROM admin_fields
ORDER BY sort_order, admin_field_id
LIMIT $1 OFFSET $2;

-- name: ListAdminFieldByParentIDPaginated :many
SELECT * FROM admin_fields
WHERE parent_id = $1
ORDER BY sort_order, admin_field_id
LIMIT $2 OFFSET $3;

-- name: UpdateAdminFieldSortOrder :exec
UPDATE admin_fields
SET sort_order = $1
WHERE admin_field_id = $2;

-- name: GetMaxAdminSortOrderByParentID :one
SELECT COALESCE(MAX(sort_order), -1)
FROM admin_fields
WHERE parent_id = $1;

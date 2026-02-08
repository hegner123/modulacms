-- name: DropAdminFieldTable :exec
DROP TABLE admin_fields;

-- name: CreateAdminFieldTable :exec
CREATE TABLE IF NOT EXISTS admin_fields (
    admin_field_id TEXT
        PRIMARY KEY NOT NULL CHECK (length(admin_field_id) = 26),
    parent_id TEXT DEFAULT NULL
        REFERENCES admin_datatypes
            ON DELETE SET NULL,
    label TEXT DEFAULT 'unlabeled' NOT NULL,
    data TEXT DEFAULT '' NOT NULL,
    validation TEXT NOT NULL,
    ui_config TEXT NOT NULL,
    type TEXT DEFAULT 'text' NOT NULL,
    author_id TEXT NOT NULL
        REFERENCES users
            ON DELETE SET NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

-- name: CreateAdminFieldParentIndex :exec
CREATE INDEX admin_fields_parent_id_index
    ON admin_fields (parent_id);

-- name: CountAdminField :one
SELECT COUNT(*)
FROM admin_fields;

-- name: GetAdminField :one
SELECT * FROM admin_fields
WHERE admin_field_id = ? LIMIT 1;

-- name: ListAdminField :many
SELECT * FROM admin_fields
ORDER BY admin_field_id;

-- name: ListAdminFieldByParentID :many
SELECT *
FROM admin_fields
WHERE parent_id = ?
ORDER BY admin_field_id;


-- name: CreateAdminField :one
INSERT INTO admin_fields (
    admin_field_id,
    parent_id,
    label,
    data,
    validation,
    ui_config,
    type,
    author_id,
    date_created,
    date_modified
    ) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
    ) RETURNING *;


-- name: UpdateAdminField :exec
UPDATE admin_fields
SET parent_id = ?,
    label = ?,
    data = ?,
    validation = ?,
    ui_config = ?,
    type = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?
    WHERE admin_field_id = ?
    RETURNING *;

-- name: DeleteAdminField :exec
DELETE FROM admin_fields
WHERE admin_field_id = ?;

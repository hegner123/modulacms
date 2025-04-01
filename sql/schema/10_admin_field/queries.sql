-- name: DropAdminFieldTable :exec
DROP TABLE admin_fields;

-- name: CreateAdminFieldTable :exec
CREATE TABLE admin_fields (
    admin_field_id INTEGER
        PRIMARY KEY,
    parent_id INTEGER DEFAULT NULL
        REFERENCES admin_datatypes
            ON DELETE SET DEFAULT,
    label TEXT DEFAULT 'unlabeled' NOT NULL,
    data TEXT DEFAULT '' NOT NULL,
    type TEXT DEFAULT 'text' NOT NULL,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
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
    parent_id,
    label,
    data,
    type,
    author_id,
    date_created,
    date_modified,
    history
    ) VALUES (
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
    type = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?,
    history = ?
    WHERE admin_field_id = ?
    RETURNING *;

-- name: DeleteAdminField :exec
DELETE FROM admin_fields
WHERE admin_field_id = ?;

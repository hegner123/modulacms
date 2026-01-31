-- name: DropAdminDatatypeTable :exec
DROP TABLE admin_datatypes;

-- name: CreateAdminDatatypeTable :exec
CREATE TABLE IF NOT EXISTS admin_datatypes (
    admin_datatype_id TEXT
        PRIMARY KEY NOT NULL CHECK (length(admin_datatype_id) = 26),
    parent_id TEXT DEFAULT NULL
        REFERENCES admin_datatypes
            ON DELETE SET NULL,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id TEXT NOT NULL
        REFERENCES users
            ON DELETE SET NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

-- name: CreateAdminDatatypeParentIDIndex :exec
CREATE INDEX admin_datatypes_parent_id_index
    ON admin_datatypes (parent_id);

-- name: CountAdminDatatype :one
SELECT COUNT(*)
FROM admin_datatypes;

-- name: GetAdminDatatype :one
SELECT * FROM admin_datatypes
WHERE admin_datatype_id = ? LIMIT 1;

-- name: ListAdminDatatype :many
SELECT * FROM admin_datatypes
ORDER BY admin_datatype_id ;

-- name: ListAdminDatatypeGlobal :many
SELECT * FROM admin_datatypes
WHERE type = 'GLOBAL' LIMIT 1;

-- name: ListAdminDatatypeRoot :many
SELECT * FROM admin_datatypes
WHERE type = 'ROOT' LIMIT 1;

-- name: ListAdminDatatypeChildren :many
SELECT * FROM admin_datatypes
WHERE parent_id = ?;

-- name: CreateAdminDatatype :one
INSERT INTO admin_datatypes (
    admin_datatype_id,
    parent_id,
    label,
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
    ?
    ) RETURNING *;

-- name: UpdateAdminDatatype :exec
UPDATE admin_datatypes
SET parent_id = ?,
    label = ?,
    type = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?
    WHERE admin_datatype_id = ?
    RETURNING *;

-- name: DeleteAdminDatatype :exec
DELETE FROM admin_datatypes
WHERE admin_datatype_id = ?;


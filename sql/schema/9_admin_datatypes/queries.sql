-- name: DropAdminDatatypeTable :exec
DROP TABLE admin_datatypes;

-- name: CreateAdminDatatypeTable :exec
CREATE TABLE admin_datatypes (
    admin_datatype_id INTEGER
        PRIMARY KEY,
    parent_id INTEGER DEFAULT NULL
        REFERENCES admin_datatypes
            ON DELETE SET DEFAULT,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id INTEGER NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
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
    parent_id,
    label,
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
    ?
    ) RETURNING *;

-- name: UpdateAdminDatatype :exec
UPDATE admin_datatypes
SET parent_id = ?,
    label = ?,
    type = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?,
    history = ?
    WHERE admin_datatype_id = ?
    RETURNING *;

-- name: DeleteAdminDatatype :exec
DELETE FROM admin_datatypes
WHERE admin_datatype_id = ?;


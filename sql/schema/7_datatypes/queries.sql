-- name: DropDatatypeTable :exec
DROP TABLE datatypes;

-- name: CreateDatatypeTable :exec
CREATE TABLE IF NOT EXISTS datatypes
(
    datatype_id TEXT PRIMARY KEY NOT NULL CHECK (length(datatype_id) = 26),
    parent_id TEXT DEFAULT NULL
        REFERENCES datatypes
            ON DELETE SET NULL,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id TEXT NOT NULL
        REFERENCES users
            ON DELETE SET NULL,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

-- name: CreateDatatypeParentIDIndex :exec
CREATE INDEX datatypes_parent_id_index
    ON datatypes (parent_id);

-- name: CountDatatype :one
SELECT COUNT(*)
FROM datatypes;

-- name: GetDatatype :one
SELECT * FROM datatypes
WHERE datatype_id = ? LIMIT 1;

-- name: ListDatatype :many
SELECT * FROM datatypes
ORDER BY datatype_id;

-- name: ListDatatypeGlobal :many
SELECT * FROM datatypes
WHERE type = 'GLOBAL'
ORDER BY datatype_id;

-- name: ListDatatypeRoot :many
SELECT * FROM datatypes
WHERE type = 'ROOT'
ORDER BY datatype_id;

-- name: ListDatatypeChildren :many
SELECT * FROM admin_datatypes
WHERE parent_id = ?
ORDER BY datatype_id;

-- name: CreateDatatype :one
INSERT INTO datatypes (
    datatype_id,
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

-- name: UpdateDatatype :exec
UPDATE datatypes
SET parent_id = ?,
    label = ?,
    type = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?
    WHERE datatype_id = ?
    RETURNING *;

-- name: DeleteDatatype :exec
DELETE FROM datatypes
WHERE datatype_id = ?;

-- name: DropDatatypeTable :exec
DROP TABLE datatypes;

-- name: CreateDatatypeTable :exec
CREATE TABLE IF NOT EXISTS datatypes
(
    datatype_id INTEGER
        PRIMARY KEY,
    parent_id INTEGER DEFAULT NULL
        REFERENCES datatypes
            ON DELETE SET DEFAULT,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
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

-- name: UpdateDatatype :exec
UPDATE datatypes
SET parent_id = ?,
    label = ?,
    type = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?,
    history = ?
    WHERE datatype_id = ?
    RETURNING *;

-- name: DeleteDatatype :exec
DELETE FROM datatypes
WHERE datatype_id = ?;

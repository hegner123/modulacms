-- name: DropDatatypeTable :exec
DROP TABLE datatypes;

-- name: CreateDatatypeTable :exec
CREATE TABLE IF NOT EXISTS datatypes (
    datatype_id TEXT PRIMARY KEY NOT NULL,
    parent_id TEXT
        CONSTRAINT fk_datatypes_parent
            REFERENCES datatypes
            ON UPDATE CASCADE ON DELETE SET NULL,
    name TEXT NOT NULL DEFAULT '',
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id TEXT NOT NULL
        CONSTRAINT fk_users_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE SET NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- name: CountDatatype :one
SELECT COUNT(*)
FROM datatypes;

-- name: GetDatatype :one
SELECT * FROM datatypes
WHERE datatype_id = $1 LIMIT 1;

-- name: ListDatatype :many
SELECT * FROM datatypes
ORDER BY datatype_id;

-- name: ListDatatypeGlobal :many
SELECT * FROM datatypes
WHERE type = 'GLOBAL'
ORDER BY datatype_id;

-- name: ListDatatypeRoot :many
SELECT * FROM datatypes
WHERE type = '_root'
ORDER BY datatype_id;

-- name: ListDatatypeChildren :many
SELECT * FROM datatypes
WHERE parent_id = $1
ORDER BY label;

-- name: CreateDatatype :one
INSERT INTO datatypes (
    datatype_id,
    parent_id,
    name,
    label,
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
    ) RETURNING *;

-- name: UpdateDatatype :exec
UPDATE datatypes
SET parent_id = $1,
    name = $2,
    label = $3,
    type = $4,
    author_id = $5,
    date_created = $6,
    date_modified = $7
    WHERE datatype_id = $8
    RETURNING *;

-- name: DeleteDatatype :exec
DELETE FROM datatypes
WHERE datatype_id = $1;

-- name: GetDatatypeByType :one
SELECT * FROM datatypes
WHERE type = $1 LIMIT 1;

-- name: ListDatatypePaginated :many
SELECT * FROM datatypes
ORDER BY datatype_id
LIMIT $1 OFFSET $2;

-- name: ListDatatypeChildrenPaginated :many
SELECT * FROM datatypes
WHERE parent_id = $1
ORDER BY label
LIMIT $2 OFFSET $3;


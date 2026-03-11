-- name: DropDatatypeTable :exec
DROP TABLE datatypes;

-- name: CreateDatatypeTable :exec
CREATE TABLE IF NOT EXISTS datatypes (
    datatype_id TEXT PRIMARY KEY NOT NULL,
    parent_id TEXT
        CONSTRAINT fk_datatypes_parent
            REFERENCES datatypes
            ON UPDATE CASCADE ON DELETE SET NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
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

-- name: CreateDatatypeParentIDIndex :exec
CREATE INDEX datatypes_parent_id_index
    ON datatypes (parent_id);

-- name: CountDatatype :one
SELECT COUNT(*)
FROM datatypes;

-- name: GetDatatype :one
SELECT * FROM datatypes
WHERE datatype_id = $1 LIMIT 1;

-- name: ListDatatype :many
SELECT * FROM datatypes
ORDER BY sort_order, datatype_id;

-- name: ListDatatypeGlobal :many
SELECT * FROM datatypes
WHERE type = '_global'
ORDER BY sort_order, datatype_id;

-- name: ListDatatypeRoot :many
SELECT * FROM datatypes
WHERE type IN ('_root', '_global')
ORDER BY sort_order, datatype_id;

-- name: ListDatatypeChildren :many
SELECT * FROM datatypes
WHERE parent_id = $1
ORDER BY sort_order, label;

-- name: CreateDatatype :one
INSERT INTO datatypes (
    datatype_id,
    parent_id,
    sort_order,
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
    $8,
    $9
    ) RETURNING *;

-- name: UpdateDatatype :exec
UPDATE datatypes
SET parent_id = $1,
    sort_order = $2,
    name = $3,
    label = $4,
    type = $5,
    author_id = $6,
    date_created = $7,
    date_modified = $8
    WHERE datatype_id = $9
    RETURNING *;

-- name: DeleteDatatype :exec
DELETE FROM datatypes
WHERE datatype_id = $1;

-- name: GetDatatypeByType :one
SELECT * FROM datatypes
WHERE type = $1 LIMIT 1;

-- name: GetDatatypeByName :one
SELECT * FROM datatypes
WHERE name = $1 LIMIT 1;

-- name: ListDatatypePaginated :many
SELECT * FROM datatypes
ORDER BY sort_order, datatype_id
LIMIT $1 OFFSET $2;

-- name: ListDatatypeChildrenPaginated :many
SELECT * FROM datatypes
WHERE parent_id = $1
ORDER BY sort_order, label
LIMIT $2 OFFSET $3;

-- name: ReassignDatatypeAuthor :exec
UPDATE datatypes SET author_id = $1 WHERE author_id = $2;

-- name: CountDatatypesByAuthor :one
SELECT COUNT(*) FROM datatypes WHERE author_id = $1;

-- name: UpdateDatatypeSortOrder :exec
UPDATE datatypes SET sort_order = $1 WHERE datatype_id = $2;

-- name: GetMaxDatatypeRootSortOrder :one
SELECT COALESCE(MAX(sort_order), -1) FROM datatypes WHERE parent_id IS NULL;

-- name: GetMaxDatatypeSortOrderByParentID :one
SELECT COALESCE(MAX(sort_order), -1) FROM datatypes WHERE parent_id = $1;

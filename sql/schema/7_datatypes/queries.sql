-- name: DropDatatypeTable :exec
DROP TABLE datatypes;

-- name: CreateDatatypeTable :exec
CREATE TABLE IF NOT EXISTS datatypes
(
    datatype_id TEXT PRIMARY KEY NOT NULL CHECK (length(datatype_id) = 26),
    parent_id TEXT DEFAULT NULL
        REFERENCES datatypes
            ON DELETE SET NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    name TEXT NOT NULL DEFAULT '',
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
WHERE parent_id = ?
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

-- name: UpdateDatatype :exec
UPDATE datatypes
SET parent_id = ?,
    sort_order = ?,
    name = ?,
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

-- name: GetDatatypeByType :one
SELECT * FROM datatypes
WHERE type = ? LIMIT 1;

-- name: GetDatatypeByName :one
SELECT * FROM datatypes
WHERE name = ? LIMIT 1;

-- name: ListDatatypePaginated :many
SELECT * FROM datatypes
ORDER BY sort_order, datatype_id
LIMIT ? OFFSET ?;

-- name: ListDatatypeChildrenPaginated :many
SELECT * FROM datatypes
WHERE parent_id = ?
ORDER BY sort_order, label
LIMIT ? OFFSET ?;

-- name: ReassignDatatypeAuthor :exec
UPDATE datatypes SET author_id = ? WHERE author_id = ?;

-- name: CountDatatypesByAuthor :one
SELECT COUNT(*) FROM datatypes WHERE author_id = ?;

-- name: UpdateDatatypeSortOrder :exec
UPDATE datatypes SET sort_order = ? WHERE datatype_id = ?;

-- name: GetMaxDatatypeRootSortOrder :one
SELECT COALESCE(MAX(sort_order), -1) FROM datatypes WHERE parent_id IS NULL;

-- name: GetMaxDatatypeSortOrderByParentID :one
SELECT COALESCE(MAX(sort_order), -1) FROM datatypes WHERE parent_id = ?;

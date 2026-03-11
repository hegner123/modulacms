-- name: DropDatatypeTable :exec
DROP TABLE datatypes;

-- name: CreateDatatypeTable :exec
CREATE TABLE IF NOT EXISTS datatypes (
    datatype_id VARCHAR(26) PRIMARY KEY NOT NULL,
    parent_id VARCHAR(26) NULL,
    sort_order INT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL DEFAULT '',
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id VARCHAR(26) NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_dt_datatypes_parent
        FOREIGN KEY (parent_id) REFERENCES datatypes (datatype_id)
            ON UPDATE CASCADE,
    CONSTRAINT fk_dt_users_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE
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

-- name: CreateDatatype :exec
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
    );

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
    WHERE datatype_id = ?;

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

-- name: DropDatatypesFieldsTable :exec
DROP TABLE datatypes_fields;

-- name: CreateDatatypesFieldsTable :exec
CREATE TABLE IF NOT EXISTS datatypes_fields (
    id TEXT PRIMARY KEY NOT NULL CHECK (length(id) = 26),
    datatype_id TEXT NOT NULL
        CONSTRAINT fk_df_datatype
            REFERENCES datatypes
            ON DELETE CASCADE,
    field_id TEXT NOT NULL
        CONSTRAINT fk_df_field
            REFERENCES fields
            ON DELETE CASCADE,
    sort_order INTEGER NOT NULL DEFAULT 0
);

-- name: CountDatatypeField :one
SELECT COUNT(*)
FROM datatypes_fields;

-- name: ListDatatypeField :many
SELECT * FROM datatypes_fields
ORDER BY sort_order, id;

-- name: ListDatatypeFieldByDatatypeID :many
SELECT *
FROM datatypes_fields
WHERE datatype_id = ?
ORDER BY sort_order, id;

-- name: ListDatatypeFieldByFieldID :many
SELECT *
FROM datatypes_fields
WHERE field_id = ?
ORDER BY sort_order, id;

-- name: CreateDatatypeField :one
INSERT INTO datatypes_fields (
    id,
    datatype_id,
    field_id,
    sort_order
) VALUES (
    ?,
    ?,
    ?,
    ?
)
RETURNING *;

-- name: UpdateDatatypeField :exec
UPDATE datatypes_fields
SET datatype_id = ?,
    field_id = ?,
    sort_order = ?
WHERE id = ?;

-- name: DeleteDatatypeField :exec
DELETE FROM datatypes_fields
WHERE id = ?;

-- name: UpdateDatatypeFieldSortOrder :exec
UPDATE datatypes_fields
SET sort_order = ?
WHERE id = ?;

-- name: GetMaxSortOrderByDatatypeID :one
SELECT COALESCE(MAX(sort_order), -1)
FROM datatypes_fields
WHERE datatype_id = ?;

-- name: ListDatatypeFieldPaginated :many
SELECT * FROM datatypes_fields
ORDER BY sort_order, id
LIMIT ? OFFSET ?;

-- name: ListDatatypeFieldByDatatypeIDPaginated :many
SELECT * FROM datatypes_fields
WHERE datatype_id = ?
ORDER BY sort_order, id
LIMIT ? OFFSET ?;

-- name: ListDatatypeFieldByFieldIDPaginated :many
SELECT * FROM datatypes_fields
WHERE field_id = ?
ORDER BY sort_order, id
LIMIT ? OFFSET ?;

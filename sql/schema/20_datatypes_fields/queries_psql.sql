-- name: DropDatatypesFieldsTable :exec
DROP TABLE datatypes_fields;

-- name: CreateDatatypesFieldsTable :exec
CREATE TABLE IF NOT EXISTS datatypes_fields (
    id TEXT PRIMARY KEY NOT NULL,
    datatype_id TEXT NOT NULL
        CONSTRAINT fk_df_datatype
            REFERENCES datatypes
            ON UPDATE CASCADE ON DELETE CASCADE,
    field_id TEXT NOT NULL
        CONSTRAINT fk_df_field
            REFERENCES fields
            ON UPDATE CASCADE ON DELETE CASCADE,
    sort_order INTEGER NOT NULL DEFAULT 0
);

-- name: CountDatatypeField :one
SELECT COUNT(*)
FROM datatypes_fields;

-- name: ListDatatypeField :many
SELECT * FROM datatypes_fields
ORDER BY sort_order, id;

-- name: ListDatatypeFieldByDatatypeID :many
SELECT * FROM datatypes_fields
WHERE datatype_id = $1
ORDER BY sort_order, id;

-- name: ListDatatypeFieldByFieldID :many
SELECT * FROM datatypes_fields
WHERE field_id = $1
ORDER BY sort_order, id;

-- name: CreateDatatypeField :one
INSERT INTO datatypes_fields (
    id,
    datatype_id,
    field_id,
    sort_order
) VALUES (
    $1,
    $2,
    $3,
    $4
) RETURNING *;

-- name: UpdateDatatypeField :exec
UPDATE datatypes_fields SET
    datatype_id = $1,
    field_id = $2,
    sort_order = $3
WHERE id = $4;

-- name: DeleteDatatypeField :exec
DELETE FROM datatypes_fields
WHERE id = $1;

-- name: UpdateDatatypeFieldSortOrder :exec
UPDATE datatypes_fields
SET sort_order = $1
WHERE id = $2;

-- name: GetMaxSortOrderByDatatypeID :one
SELECT COALESCE(MAX(sort_order), -1)
FROM datatypes_fields
WHERE datatype_id = $1;

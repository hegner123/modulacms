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
            ON DELETE CASCADE
);

-- name: CountDatatypeField :one
SELECT COUNT(*)
FROM datatypes_fields;

-- name: ListDatatypeField :many
SELECT * FROM datatypes_fields
ORDER BY id;

-- name: ListDatatypeFieldByDatatypeID :many
SELECT * 
FROM datatypes_fields
WHERE datatype_id = ?
ORDER BY id;

-- name: ListDatatypeFieldByFieldID :many
SELECT * 
FROM datatypes_fields
WHERE field_id = ?
ORDER BY id;

-- name: CreateDatatypeField :one
INSERT INTO datatypes_fields (
    id,
    datatype_id,
    field_id
) VALUES (
    ?,
    ?,
    ?
)
RETURNING *;

-- name: UpdateDatatypeField :exec
UPDATE datatypes_fields 
SET datatype_id = ?,
    field_id = ?
WHERE id = ?;

-- name: DeleteDatatypeField :exec
DELETE FROM datatypes_fields
WHERE id = ?;

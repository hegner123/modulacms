-- name: DropDatatypesFieldsTable :exec
DROP TABLE datatypes_fields;

-- name: CreateDatatypesFieldsTable :exec
CREATE TABLE IF NOT EXISTS datatypes_fields (
    id VARCHAR(26) NOT NULL
        PRIMARY KEY,
    datatype_id VARCHAR(26) NOT NULL,
    field_id VARCHAR(26) NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    CONSTRAINT fk_df_datatype
        FOREIGN KEY (datatype_id) REFERENCES datatypes (datatype_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_df_field
        FOREIGN KEY (field_id) REFERENCES fields (field_id)
            ON UPDATE CASCADE ON DELETE CASCADE
);

-- name: CountDatatypeField :one
SELECT COUNT(*)
FROM datatypes_fields;

-- name: GetDatatypeField :one
SELECT * FROM datatypes_fields WHERE id = ? LIMIT 1;

-- name: ListDatatypeField :many
SELECT * FROM datatypes_fields
ORDER BY sort_order, id;

-- name: ListDatatypeFieldByDatatypeID :many
SELECT * FROM datatypes_fields
WHERE datatype_id = ?
ORDER BY sort_order, id;

-- name: ListDatatypeFieldByFieldID :many
SELECT * FROM datatypes_fields
WHERE field_id = ?
ORDER BY sort_order, id;

-- name: CreateDatatypeField :exec
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
);

-- name: UpdateDatatypeField :exec
UPDATE datatypes_fields SET
    datatype_id = ?,
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

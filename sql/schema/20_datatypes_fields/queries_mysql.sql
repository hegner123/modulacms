-- name: DropDatatypesFieldsTable :exec
DROP TABLE datatypes_fields;

-- name: CreateDatatypesFieldsTable :exec
CREATE TABLE IF NOT EXISTS datatypes_fields (
    id VARCHAR(26) NOT NULL
        PRIMARY KEY,
    datatype_id VARCHAR(26) NOT NULL,
    field_id VARCHAR(26) NOT NULL,
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

-- name: ListDatatypeField :many
SELECT * FROM datatypes_fields
ORDER BY id;

-- name: ListDatatypeFieldByDatatypeID :many
SELECT * FROM datatypes_fields
WHERE datatype_id = ?
ORDER BY id;

-- name: ListDatatypeFieldByFieldID :many
SELECT * FROM datatypes_fields
WHERE field_id = ?
ORDER BY id;

-- name: CreateDatatypeField :exec
INSERT INTO datatypes_fields (
    id,
    datatype_id,
    field_id
) VALUES (
    ?,
    ?,
    ?
);

-- name: GetLastDatatypeField :one
SELECT * FROM datatypes_fields WHERE id = LAST_INSERT_ID();

-- name: UpdateDatatypeField :exec
UPDATE datatypes_fields SET
    datatype_id = ?,
    field_id = ?
WHERE id = ?;

-- name: DeleteDatatypeField :exec
DELETE FROM datatypes_fields
WHERE id = ?;

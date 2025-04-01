-- name: DropAdminDatatypesFieldsTable :exec
DROP TABLE admin_datatypes_fields;

-- name: CreateAdminDatatypesFieldsTable :exec
CREATE TABLE IF NOT EXISTS admin_datatypes_fields (
    id INTEGER
        PRIMARY KEY,
    admin_datatype_id INTEGER NOT NULL
        CONSTRAINT fk_df_admin_datatype
            REFERENCES admin_datatypes
            ON DELETE CASCADE,
    admin_field_id INTEGER NOT NULL
        CONSTRAINT fk_df_admin_field
            REFERENCES admin_fields
            ON DELETE CASCADE
);

-- name: CountAdminDatatypeField :one
SELECT COUNT(*)
FROM admin_datatypes_fields;

-- name: ListAdminDatatypeField :many
SELECT * FROM admin_datatypes_fields
ORDER BY id;

-- name: ListAdminDatatypeFieldByDatatypeID :many
SELECT * FROM admin_datatypes_fields
WHERE admin_datatype_id = ?
ORDER BY id;

-- name: ListAdminDatatypeFieldByFieldID :many
SELECT * FROM admin_datatypes_fields
WHERE admin_field_id = ?
ORDER BY id;

-- name: CreateAdminDatatypeField :one
INSERT INTO admin_datatypes_fields (
    admin_datatype_id,
    admin_field_id
) VALUES (
    ?,
    ?
) RETURNING *;

-- name: UpdateAdminDatatypeField :exec
UPDATE admin_datatypes_fields SET
    admin_datatype_id = ?,
    admin_field_id = ?
WHERE id = ?;

-- name: DeleteAdminDatatypeField :exec
DELETE FROM admin_datatypes_fields
WHERE id = ?;

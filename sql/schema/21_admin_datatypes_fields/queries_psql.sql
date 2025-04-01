-- name: DropAdminDatatypesFieldsTable :exec
DROP TABLE admin_datatypes_fields;

-- name: CreateAdminDatatypesFieldsTable :exec
CREATE TABLE IF NOT EXISTS admin_datatypes_fields (
    id SERIAL
        PRIMARY KEY,
    admin_datatype_id INTEGER NOT NULL
        CONSTRAINT fk_df_admin_datatype
            REFERENCES admin_datatypes
            ON UPDATE CASCADE ON DELETE CASCADE,
    admin_field_id INTEGER NOT NULL
        CONSTRAINT fk_df_admin_field
            REFERENCES admin_fields
            ON UPDATE CASCADE ON DELETE CASCADE
);

-- name: CountAdminDatatypeField :one
SELECT COUNT(*)
FROM admin_datatypes_fields;

-- name: ListAdminDatatypeField :many
SELECT * FROM admin_datatypes_fields;

-- name: ListAdminDatatypeFieldByDatatypeID :many
SELECT * FROM admin_datatypes_fields
WHERE admin_datatype_id = $1;

-- name: ListAdminDatatypeFieldByFieldID :many
SELECT * FROM admin_datatypes_fields
WHERE admin_field_id = $1;

-- name: CreateAdminDatatypeField :one
INSERT INTO admin_datatypes_fields (
    admin_datatype_id,
    admin_field_id
) VALUES (
    $1,
    $2
) RETURNING *;

-- name: UpdateAdminDatatypeField :exec
UPDATE admin_datatypes_fields SET
    admin_datatype_id = $1,
    admin_field_id = $2
WHERE id = $3;

-- name: DeleteAdminDatatypeField :exec
DELETE FROM admin_datatypes_fields
WHERE id = $1;

-- name: DropAdminDatatypesFieldsTable :exec
DROP TABLE admin_datatypes_fields;

-- name: CreateAdminDatatypesFieldsTable :exec
CREATE TABLE IF NOT EXISTS admin_datatypes_fields (
    id VARCHAR(26) NOT NULL
        PRIMARY KEY,
    admin_datatype_id VARCHAR(26) NOT NULL,
    admin_field_id VARCHAR(26) NOT NULL,
    CONSTRAINT fk_df_admin_datatype
        FOREIGN KEY (admin_datatype_id) REFERENCES admin_datatypes (admin_datatype_id)
            ON UPDATE CASCADE ON DELETE CASCADE,
    CONSTRAINT fk_df_admin_field
        FOREIGN KEY (admin_field_id) REFERENCES admin_fields (admin_field_id)
            ON UPDATE CASCADE ON DELETE CASCADE
);

-- name: CountAdminDatatypeField :one
SELECT COUNT(*)
FROM admin_datatypes_fields;

-- name: GetAdminDatatypeField :one
SELECT * FROM admin_datatypes_fields WHERE id = ? LIMIT 1;

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

-- name: CreateAdminDatatypeField :exec
INSERT INTO admin_datatypes_fields (
    id,
    admin_datatype_id,
    admin_field_id
) VALUES (
    ?,
    ?,
    ?
);

-- name: UpdateAdminDatatypeField :exec
UPDATE admin_datatypes_fields SET
    admin_datatype_id = ?,
    admin_field_id = ?
WHERE id = ?;

-- name: DeleteAdminDatatypeField :exec
DELETE FROM admin_datatypes_fields
WHERE id = ?;

-- name: ListAdminDatatypeFieldPaginated :many
SELECT * FROM admin_datatypes_fields
ORDER BY id
LIMIT ? OFFSET ?;

-- name: ListAdminDatatypeFieldByDatatypeIDPaginated :many
SELECT * FROM admin_datatypes_fields
WHERE admin_datatype_id = ?
ORDER BY id
LIMIT ? OFFSET ?;

-- name: ListAdminDatatypeFieldByFieldIDPaginated :many
SELECT * FROM admin_datatypes_fields
WHERE admin_field_id = ?
ORDER BY id
LIMIT ? OFFSET ?;

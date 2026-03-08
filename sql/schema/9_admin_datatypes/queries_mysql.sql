-- name: DropAdminDatatypeTable :exec
DROP TABLE admin_datatypes;

-- name: CreateAdminDatatypeTable :exec
CREATE TABLE IF NOT EXISTS admin_datatypes (
    admin_datatype_id VARCHAR(26) PRIMARY KEY NOT NULL,
    parent_id VARCHAR(26) NULL,
    sort_order INT NOT NULL DEFAULT 0,
    name VARCHAR(255) NOT NULL DEFAULT '',
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id VARCHAR(26) NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT fk_admin_datatypes_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE,
    CONSTRAINT fk_admin_datatypes_parent_id
        FOREIGN KEY (parent_id) REFERENCES admin_datatypes (admin_datatype_id)
            ON UPDATE CASCADE
);

-- name: CountAdminDatatype :one
SELECT COUNT(*)
FROM admin_datatypes;

-- name: GetAdminDatatype :one
SELECT * FROM admin_datatypes
WHERE admin_datatype_id = ?
LIMIT 1;

-- name: ListAdminDatatype :many
SELECT * FROM admin_datatypes
ORDER BY sort_order, admin_datatype_id;

-- name: ListAdminDatatypeGlobal :many
SELECT * FROM admin_datatypes
WHERE type = '_global'
ORDER BY sort_order, admin_datatype_id;

-- name: ListAdminDatatypeRoot :many
SELECT * FROM admin_datatypes
WHERE type IN ('_root', '_global')
ORDER BY sort_order, admin_datatype_id;

-- name: ListAdminDatatypeChildren :many
SELECT * FROM admin_datatypes
WHERE parent_id = ?
ORDER BY sort_order, admin_datatype_id;

-- name: CreateAdminDatatype :exec
INSERT INTO admin_datatypes (
    admin_datatype_id,
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

-- name: UpdateAdminDatatype :exec
UPDATE admin_datatypes
SET parent_id = ?,
    sort_order = ?,
    name = ?,
    label = ?,
    type = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?
WHERE admin_datatype_id = ?;

-- name: DeleteAdminDatatype :exec
DELETE FROM admin_datatypes
WHERE admin_datatype_id = ?;

-- name: ListAdminDatatypePaginated :many
SELECT * FROM admin_datatypes
ORDER BY sort_order, admin_datatype_id
LIMIT ? OFFSET ?;

-- name: ListAdminDatatypeChildrenPaginated :many
SELECT * FROM admin_datatypes
WHERE parent_id = ?
ORDER BY sort_order, admin_datatype_id
LIMIT ? OFFSET ?;

-- name: UpdateAdminDatatypeSortOrder :exec
UPDATE admin_datatypes SET sort_order = ? WHERE admin_datatype_id = ?;

-- name: GetMaxAdminDatatypeRootSortOrder :one
SELECT COALESCE(MAX(sort_order), -1) FROM admin_datatypes WHERE parent_id IS NULL;

-- name: GetMaxAdminDatatypeSortOrderByParentID :one
SELECT COALESCE(MAX(sort_order), -1) FROM admin_datatypes WHERE parent_id = ?;

-- name: DropAdminDatatypeTable :exec
DROP TABLE admin_datatypes;

-- name: CreateAdminDatatypeTable :exec
CREATE TABLE IF NOT EXISTS admin_datatypes (
    admin_datatype_id INT AUTO_INCREMENT
        PRIMARY KEY,
    parent_id INT NULL,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id INT DEFAULT 1 NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,

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
ORDER BY admin_datatype_id;

-- name: ListAdminDatatypeGlobal :many
SELECT * FROM admin_datatypes
WHERE type = 'GLOBAL' LIMIT 1;

-- name: ListAdminDatatypeRoot :many
SELECT * FROM admin_datatypes
WHERE type = 'ROOT' LIMIT 1;

-- name: ListAdminDatatypeChildren :many
SELECT * FROM admin_datatypes
WHERE parent_id = ?;

-- name: CreateAdminDatatype :exec
INSERT INTO admin_datatypes (
    parent_id,
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
    ?
);

-- name: GetLastAdminDatatype :one
SELECT * FROM admin_datatypes WHERE admin_datatype_id = LAST_INSERT_ID();

-- name: UpdateAdminDatatype :exec
UPDATE admin_datatypes
SET parent_id = ?,
    label = ?,
    type = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?
WHERE admin_datatype_id = ?;

-- name: DeleteAdminDatatype :exec
DELETE FROM admin_datatypes
WHERE admin_datatype_id = ?;

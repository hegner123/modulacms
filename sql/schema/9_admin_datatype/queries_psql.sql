-- name: DropAdminDatatypeTable :exec
DROP TABLE admin_datatypes;

-- name: CreateAdminDatatypeTable :exec
CREATE TABLE IF NOT EXISTS admin_datatypes (
    admin_datatype_id SERIAL
        PRIMARY KEY,
    parent_id INTEGER
        CONSTRAINT fk_parent_id
            REFERENCES admin_datatypes
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id INTEGER NOT NULL
        CONSTRAINT fk_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);

-- name: CountAdminDatatype :one
SELECT COUNT(*)
FROM admin_datatypes;

-- name: GetAdminDatatype :one
SELECT * FROM admin_datatypes
WHERE admin_datatype_id = $1
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
WHERE parent_id = $1;

-- name: CreateAdminDatatype :one
INSERT INTO admin_datatypes (
    parent_id,
    label,
    type,
    author_id,
    date_created,
    date_modified,
    history
) VALUES ( 
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
)
RETURNING *;

-- name: UpdateAdminDatatype :exec
UPDATE admin_datatypes
SET parent_id = $1,
    label = $2,
    type = $3,
    author_id = $4,
    date_created = $5,
    date_modified = $6,
    history = $7
WHERE admin_datatype_id = $8
RETURNING *;

-- name: DeleteAdminDatatype :exec
DELETE FROM admin_datatypes
WHERE admin_datatype_id = $1;



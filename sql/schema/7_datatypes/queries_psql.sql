-- name: DropDatatypeTable :exec
DROP TABLE datatypes;

-- name: CreateDatatypeTable :exec
CREATE TABLE IF NOT EXISTS datatypes (
    datatype_id SERIAL
        PRIMARY KEY,
    parent_id INTEGER
        CONSTRAINT fk_datatypes_parent
            REFERENCES datatypes
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id INTEGER DEFAULT 1 NOT NULL
        CONSTRAINT fk_users_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- name: CountDatatype :one
SELECT COUNT(*)
FROM datatypes;

-- name: GetDatatype :one
SELECT * FROM datatypes
WHERE datatype_id = $1 LIMIT 1;

-- name: ListDatatype :many
SELECT * FROM datatypes
ORDER BY datatype_id;

-- name: ListDatatypeGlobal :many
SELECT * FROM datatypes
WHERE type = 'GLOBAL'
ORDER BY datatype_id;

-- name: ListDatatypeRoot :many
SELECT * FROM datatypes
WHERE type = 'ROOT'
ORDER BY datatype_id;

-- name: ListDatatypeChildren :many
SELECT * FROM datatypes
WHERE parent_id = $1
ORDER BY datatype_id;

-- name: CreateDatatype :one
INSERT INTO datatypes (    
    parent_id,
    label,
    type,
    author_id,
    date_created,
    date_modified
    ) VALUES (    
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
    ) RETURNING *;

-- name: UpdateDatatype :exec
UPDATE datatypes
SET parent_id = $1,
    label = $2,
    type = $3,
    author_id = $4,
    date_created = $5,
    date_modified = $6
    WHERE datatype_id = $7
    RETURNING *;

-- name: DeleteDatatype :exec
DELETE FROM datatypes
WHERE datatype_id = $1;


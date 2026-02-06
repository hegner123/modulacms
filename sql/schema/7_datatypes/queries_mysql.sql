-- name: DropDatatypeTable :exec
DROP TABLE datatypes;

-- name: CreateDatatypeTable :exec
CREATE TABLE IF NOT EXISTS datatypes (
    datatype_id VARCHAR(26) PRIMARY KEY NOT NULL,
    parent_id VARCHAR(26) NULL,
    label TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id VARCHAR(26) NOT NULL,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,

    CONSTRAINT fk_dt_datatypes_parent
        FOREIGN KEY (parent_id) REFERENCES datatypes (datatype_id)
            ON UPDATE CASCADE,
    CONSTRAINT fk_dt_users_author_id
        FOREIGN KEY (author_id) REFERENCES users (user_id)
            ON UPDATE CASCADE
);


-- name: CountDatatype :one
SELECT COUNT(*)
FROM datatypes;

-- name: GetDatatype :one
SELECT * FROM datatypes
WHERE datatype_id = ? LIMIT 1;

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
WHERE parent_id = ?
ORDER BY label;

-- name: CreateDatatype :exec
INSERT INTO datatypes (
    datatype_id,
    parent_id,
    label,
    type,
    author_id
    ) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?
    );

-- name: GetLastDatatype :one
SELECT * FROM datatypes WHERE datatype_id = LAST_INSERT_ID();

-- name: UpdateDatatype :exec
UPDATE datatypes
set 
    parent_id = ?,
    label = ?,
    type = ?,
    author_id = ?
    WHERE datatype_id = ?;

-- name: DeleteDatatype :exec
DELETE FROM datatypes
WHERE datatype_id = ?;

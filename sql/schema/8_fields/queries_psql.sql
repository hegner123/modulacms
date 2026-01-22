-- name: DropFieldTable :exec
DROP TABLE fields;

-- name: CreateFieldTable :exec
CREATE TABLE IF NOT EXISTS fields (
    field_id SERIAL
        PRIMARY KEY,
    parent_id INTEGER
        CONSTRAINT fk_datatypes
            REFERENCES datatypes
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    label TEXT DEFAULT 'unlabeled'::TEXT NOT NULL,
    data TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id INTEGER DEFAULT 1 NOT NULL
        CONSTRAINT fk_users_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE SET DEFAULT,
    date_created TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    date_modified TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- name: CreateParentIDIndex :exec
CREATE INDEX parent_id
    ON fields (parent_id);

-- name: CountField :one
SELECT COUNT(*)
FROM fields ;

-- name: GetField :one
SELECT * FROM fields 
WHERE field_id = $1 LIMIT 1;

-- name: ListField :many
SELECT * FROM fields 
ORDER BY field_id;

-- name: ListFieldByDatatypeID :many
SELECT * FROM fields 
WHERE parent_id = $1
ORDER BY field_id;

-- name: CreateField :one
INSERT INTO fields  (    
    parent_id,
    label,
    data,
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
    $6,
    $7
    ) RETURNING *;


-- name: UpdateField :exec
UPDATE fields 
SET parent_id = $1,
    label = $2,
    data = $3,
    type = $4,
    author_id = $5,
    date_created = $6,
    date_modified = $7
    WHERE field_id = $8
    RETURNING *;

-- name: DeleteField :exec
DELETE FROM fields 
WHERE field_id = $1;


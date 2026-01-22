-- name: DropFieldTable :exec
DROP TABLE fields;

-- name: CreateFieldTable :exec
CREATE TABLE IF NOT EXISTS fields(
    field_id INTEGER
        PRIMARY KEY,
    parent_id INTEGER DEFAULT NULL
        REFERENCES datatypes
            ON DELETE SET DEFAULT,
    label TEXT DEFAULT 'unlabeled' NOT NULL,
    data TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id INTEGER DEFAULT 1 NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

-- name: CountField :one
SELECT COUNT(*)
FROM fields ;

-- name: GetField :one
SELECT * FROM fields 
WHERE field_id = ? LIMIT 1;

-- name: ListField :many
SELECT * FROM fields 
ORDER BY field_id;

-- name: ListFieldByDatatypeID :many
SELECT * FROM fields 
WHERE parent_id = ?
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
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
    ) RETURNING *;


-- name: UpdateField :exec
UPDATE fields 
SET parent_id = ?,
    label = ?,
    data = ?,
    type = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?
    WHERE field_id = ?
    RETURNING *;

-- name: DeleteField :exec
DELETE FROM fields 
WHERE field_id = ?;


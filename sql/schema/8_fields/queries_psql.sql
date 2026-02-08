-- name: DropFieldTable :exec
DROP TABLE fields;

-- name: CreateFieldTable :exec
CREATE TABLE IF NOT EXISTS fields (
    field_id TEXT PRIMARY KEY NOT NULL,
    parent_id TEXT
        CONSTRAINT fk_datatypes
            REFERENCES datatypes
            ON UPDATE CASCADE ON DELETE SET NULL,
    label TEXT DEFAULT 'unlabeled'::TEXT NOT NULL,
    data TEXT NOT NULL,
    validation TEXT NOT NULL,
    ui_config TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id TEXT NOT NULL
        CONSTRAINT fk_users_author_id
            REFERENCES users
            ON UPDATE CASCADE ON DELETE SET NULL,
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
    field_id,
    parent_id,
    label,
    data,
    validation,
    ui_config,
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
    $7,
    $8,
    $9,
    $10
    ) RETURNING *;


-- name: UpdateField :exec
UPDATE fields
SET parent_id = $1,
    label = $2,
    data = $3,
    validation = $4,
    ui_config = $5,
    type = $6,
    author_id = $7,
    date_created = $8,
    date_modified = $9
    WHERE field_id = $10
    RETURNING *;

-- name: DeleteField :exec
DELETE FROM fields
WHERE field_id = $1;

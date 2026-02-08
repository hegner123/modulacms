-- name: DropFieldTable :exec
DROP TABLE fields;

-- name: CreateFieldTable :exec
CREATE TABLE IF NOT EXISTS fields(
    field_id TEXT PRIMARY KEY NOT NULL CHECK (length(field_id) = 26),
    parent_id TEXT DEFAULT NULL
        REFERENCES datatypes
            ON DELETE SET NULL,
    label TEXT DEFAULT 'unlabeled' NOT NULL,
    data TEXT NOT NULL,
    validation TEXT NOT NULL,
    ui_config TEXT NOT NULL,
    type TEXT NOT NULL,
    author_id TEXT NOT NULL
        REFERENCES users
            ON DELETE SET NULL,
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
    ?,
    ?,
    ?,
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
    validation = ?,
    ui_config = ?,
    type = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?
    WHERE field_id = ?
    RETURNING *;

-- name: DeleteField :exec
DELETE FROM fields
WHERE field_id = ?;

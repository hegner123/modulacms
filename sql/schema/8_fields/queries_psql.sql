-- name: DropFieldTable :exec
DROP TABLE fields;

-- name: CreateFieldTable :exec
CREATE TABLE IF NOT EXISTS fields (
    field_id TEXT PRIMARY KEY NOT NULL,
    parent_id TEXT
        CONSTRAINT fk_datatypes
            REFERENCES datatypes
            ON UPDATE CASCADE ON DELETE CASCADE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    name TEXT NOT NULL DEFAULT '',
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
ORDER BY sort_order, field_id;

-- name: ListFieldByDatatypeID :many
SELECT * FROM fields
WHERE parent_id = $1
ORDER BY sort_order, field_id;

-- name: CreateField :one
INSERT INTO fields  (
    field_id,
    parent_id,
    sort_order,
    name,
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
    $10,
    $11,
    $12
    ) RETURNING *;


-- name: UpdateField :exec
UPDATE fields
SET parent_id = $1,
    sort_order = $2,
    name = $3,
    label = $4,
    data = $5,
    validation = $6,
    ui_config = $7,
    type = $8,
    author_id = $9,
    date_created = $10,
    date_modified = $11
    WHERE field_id = $12
    RETURNING *;

-- name: DeleteField :exec
DELETE FROM fields
WHERE field_id = $1;

-- name: ListFieldPaginated :many
SELECT * FROM fields
ORDER BY sort_order, field_id
LIMIT $1 OFFSET $2;

-- name: UpdateFieldSortOrder :exec
UPDATE fields
SET sort_order = $1
WHERE field_id = $2;

-- name: GetMaxSortOrderByParentID :one
SELECT COALESCE(MAX(sort_order), -1)
FROM fields
WHERE parent_id = $1;

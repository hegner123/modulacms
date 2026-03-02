-- name: DropFieldTable :exec
DROP TABLE fields;

-- name: CreateFieldTable :exec
CREATE TABLE IF NOT EXISTS fields(
    field_id TEXT PRIMARY KEY NOT NULL CHECK (length(field_id) = 26),
    parent_id TEXT DEFAULT NULL
        REFERENCES datatypes
            ON DELETE CASCADE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    name TEXT NOT NULL DEFAULT '',
    label TEXT DEFAULT 'unlabeled' NOT NULL,
    data TEXT NOT NULL,
    validation TEXT NOT NULL,
    ui_config TEXT NOT NULL,
    type TEXT NOT NULL,
    translatable INTEGER NOT NULL DEFAULT 0,
    roles TEXT DEFAULT NULL,
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
ORDER BY sort_order, field_id;

-- name: ListFieldByDatatypeID :many
SELECT * FROM fields
WHERE parent_id = ?
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
    translatable,
    roles,
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
    ?,
    ?,
    ?,
    ?,
    ?
    ) RETURNING *;


-- name: UpdateField :exec
UPDATE fields
SET parent_id = ?,
    sort_order = ?,
    name = ?,
    label = ?,
    data = ?,
    validation = ?,
    ui_config = ?,
    type = ?,
    translatable = ?,
    roles = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?
    WHERE field_id = ?
    RETURNING *;

-- name: DeleteField :exec
DELETE FROM fields
WHERE field_id = ?;

-- name: ListFieldPaginated :many
SELECT * FROM fields
ORDER BY sort_order, field_id
LIMIT ? OFFSET ?;

-- name: UpdateFieldSortOrder :exec
UPDATE fields
SET sort_order = ?
WHERE field_id = ?;

-- name: GetFieldsByIDs :many
SELECT * FROM fields
WHERE field_id IN (sqlc.slice('ids'));

-- name: GetMaxSortOrderByParentID :one
SELECT COALESCE(MAX(sort_order), -1)
FROM fields
WHERE parent_id = ?;

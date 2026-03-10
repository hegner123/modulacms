-- name: DropContentFieldTable :exec
DROP TABLE content_fields;

-- name: CreateContentFieldTable :exec
CREATE TABLE IF NOT EXISTS content_fields (
    content_field_id TEXT PRIMARY KEY NOT NULL CHECK (length(content_field_id) = 26),
    route_id TEXT
        REFERENCES routes
            ON UPDATE CASCADE ON DELETE SET NULL,
    root_id TEXT
        REFERENCES content_data
            ON UPDATE CASCADE ON DELETE SET NULL,
    content_data_id TEXT NOT NULL
        REFERENCES content_data
            ON UPDATE CASCADE ON DELETE CASCADE,
    field_id TEXT NOT NULL
        REFERENCES fields
            ON UPDATE CASCADE ON DELETE CASCADE,
    field_value TEXT NOT NULL,
    locale TEXT NOT NULL DEFAULT '',
    author_id TEXT NOT NULL
        REFERENCES users
            ON DELETE CASCADE,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);

-- name: CountContentField :one
SELECT COUNT(*)
FROM content_fields;

-- name: GetContentField :one
SELECT * FROM content_fields
WHERE content_field_id = ? LIMIT 1;

-- name: ListContentFields :many
SELECT * FROM content_fields
ORDER BY content_field_id;

-- name: ListContentFieldsByRoute :many
SELECT * FROM content_fields
WHERE route_id = ?
ORDER BY content_field_id;

-- name: ListContentFieldsByContentData :many
SELECT * FROM content_fields
WHERE content_data_id = ?
ORDER BY content_field_id;

-- name: CreateContentField :one
INSERT INTO content_fields (
    content_field_id,
    route_id,
    root_id,
    content_data_id,
    field_id,
    field_value,
    locale,
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


-- name: UpdateContentField :exec
UPDATE content_fields
SET route_id = ?,
    root_id = ?,
    content_data_id = ?,
    field_id = ?,
    field_value = ?,
    locale = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?
WHERE content_field_id = ?;

-- name: DeleteContentField :exec
DELETE FROM content_fields
WHERE content_field_id = ?;

-- name: ListContentFieldsPaginated :many
SELECT * FROM content_fields
ORDER BY content_field_id
LIMIT ? OFFSET ?;

-- name: ListContentFieldsByRoutePaginated :many
SELECT * FROM content_fields
WHERE route_id = ?
ORDER BY content_field_id
LIMIT ? OFFSET ?;

-- name: ListContentFieldsByContentDataPaginated :many
SELECT * FROM content_fields
WHERE content_data_id = ?
ORDER BY content_field_id
LIMIT ? OFFSET ?;

-- name: ListContentFieldsByContentDataAndLocale :many
SELECT * FROM content_fields
WHERE content_data_id = ? AND locale IN (?, '')
ORDER BY content_field_id;

-- name: ListContentFieldsByRouteAndLocale :many
SELECT * FROM content_fields
WHERE route_id = ? AND locale IN (?, '')
ORDER BY content_data_id, field_id;

-- name: ListContentFieldsByRootID :many
SELECT * FROM content_fields
WHERE root_id = ?
ORDER BY content_data_id, field_id;

-- name: ListContentFieldsByRootIDAndLocale :many
SELECT * FROM content_fields
WHERE root_id = ? AND locale IN (?, '')
ORDER BY content_data_id, field_id;

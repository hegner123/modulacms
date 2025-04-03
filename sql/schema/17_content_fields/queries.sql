-- name: DropContentFieldTable :exec
DROP TABLE content_fields;

-- name: CreateContentFieldTable :exec
CREATE TABLE content_fields (
    content_field_id INTEGER
        PRIMARY KEY,
    route_id INTEGER
        REFERENCES routes
            ON UPDATE CASCADE ON DELETE SET NULL,
    content_data_id INTEGER NOT NULL
        REFERENCES content_data
            ON UPDATE CASCADE ON DELETE CASCADE,
    field_id INTEGER NOT NULL
        REFERENCES fields
            ON UPDATE CASCADE ON DELETE CASCADE,
    field_value TEXT NOT NULL,
    author_id INTEGER NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT
);

-- name: CountContentField :one
SELECT COUNT(*)
FROM content_fields;

-- name: GetContentField :one
SELECT * FROM content_fields
WHERE content_field_id = ? LIMIT 1;

-- name: ListContentFields :many
SELECT * FROM content_fields
ORDER BY content_fields_id;

-- name: ListContentFieldsByRoute :many
SELECT * FROM content_fields
WHERE route_id = ?
ORDER BY content_fields_id;

-- name: CreateContentField :one
INSERT INTO content_fields (
    route_id,
    content_data_id,
    field_id,
    field_value,
    author_id, 
    date_created,
    date_modified,
    history
) VALUES (
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
    content_data_id = ?,
    field_id = ?,
    field_value = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?,
    history = ?
WHERE content_field_id = ?;

-- name: DeleteContentField :exec
DELETE FROM content_fields
WHERE content_field_id = ?;

-- name: CreateContentFieldTable :exec
CREATE TABLE IF NOT EXISTS content_fields (
    content_field_id INTEGER PRIMARY KEY,
    route_id       INTEGER NOT NULL
    REFERENCES routes(route_id)
    ON UPDATE CASCADE ON DELETE SET NULL,
    content_data_id       INTEGER NOT NULL
        REFERENCES content_data(content_data_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    field_id      INTEGER NOT NULL
        REFERENCES admin_fields(admin_field_id)
        ON UPDATE CASCADE ON DELETE CASCADE,
    field_value         TEXT NOT NULL,
    date_created        TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified       TEXT DEFAULT CURRENT_TIMESTAMP,
    history             TEXT
);

-- name: GetContentField :one
SELECT * FROM content_fields
WHERE content_field_id = ? LIMIT 1;

-- name: CountContentField :one
SELECT COUNT(*)
FROM content_fields;

-- name: ListContentFields :many
SELECT * FROM content_fields
ORDER BY content_fields_id;

-- name: CreateContentField :one
INSERT INTO content_fields (
    content_field_id,
    route_id,
    content_data_id,
    field_id,
    field_value, 
    history,
    date_created, 
    date_modified
    ) VALUES (
  ?,?,?,?,?,?,?,?
    ) RETURNING *;


-- name: UpdateContentField :exec
UPDATE content_fields
set  content_field_id=?,
    route_id=?,
    content_data_id=?,
    field_id=?,
    field_value=?, 
    history=?,
    date_created=?, 
    date_modified=?
    WHERE content_field_id = ?
    RETURNING *;

-- name: DeleteContentField :exec
DELETE FROM content_fields
WHERE content_field_id = ?;

-- name: ListContentFieldsByRoute :many
SELECT * FROM content_fields
WHERE route_id = ?;

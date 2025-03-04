-- name: CreateContentDataTable :exec
CREATE TABLE IF NOT EXISTS content_data (
    content_data_id INTEGER PRIMARY KEY,
    admin_dt_id   INTEGER NOT NULL
        REFERENCES admin_datatypes(admin_dt_id)
        ON UPDATE CASCADE ON DELETE SET NULL,
    history TEXT  DEFAULT NULL,
    date_created  TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP
);
-- name: GetContentData :one
SELECT * FROM content_data
WHERE content_data_id = ? LIMIT 1;

-- name: CountContentData :one
SELECT COUNT(*)
FROM content_data;


-- name: ListContentData :many
SELECT * FROM content_data
ORDER BY content_data_id;


-- name: CreateContentData :one
INSERT INTO content_data (
    admin_dt_id,
    history,
    date_created,
    date_modified
    ) VALUES (
?,?,?,?
    ) RETURNING *;


-- name: UpdateContentData :exec
UPDATE content_data
set admin_dt_id = ?,
    history = ?,
    date_created = ?,
    date_modified = ?
    WHERE content_data_id = ?
    RETURNING *;

-- name: DeleteContentData :exec
DELETE FROM content_data
WHERE content_data_id = ?;

-- name: ListFilteredContentData :many
SELECT * FROM content_data
WHERE content_data_id = ?;

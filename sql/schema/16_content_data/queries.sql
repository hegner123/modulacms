-- name: DropContentDataTable :exec
DROP TABLE content_data;

-- name: CreateContentDataTable :exec
CREATE TABLE content_data (
    content_data_id INTEGER
        PRIMARY KEY,
    parent_id INTEGER
        REFERENCES content_data
            ON DELETE SET NULL,
    route_id INTEGER NOT NULL
        REFERENCES routes
            ON DELETE CASCADE,
    datatype_id INTEGER NOT NULL
        REFERENCES datatypes
            ON DELETE SET NULL,
    author_id INTEGER NOT NULL
        REFERENCES users
            ON DELETE SET DEFAULT,
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,
    history TEXT DEFAULT NULL
);

-- name: CountContentData :one
SELECT COUNT(*)
FROM content_data;

-- name: GetContentData :one
SELECT * FROM content_data
WHERE content_data_id = ? LIMIT 1;

-- name: ListContentData :many
SELECT * FROM content_data
ORDER BY content_data_id;

-- name: ListContentDataByRoute :many
SELECT * FROM content_data
WHERE route_id = ?
ORDER BY content_data_id;

-- name: CreateContentData :one
INSERT INTO content_data (
    route_id,
    parent_id,
    datatype_id,
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
    ?
) RETURNING *;

-- name: UpdateContentData :exec
UPDATE content_data
SET route_id = ?, 
    parent_id = ?,
    datatype_id = ?,
    author_id = ?,
    date_created = ?,
    date_modified = ?,
    history = ?
WHERE content_data_id = ?;

-- name: DeleteContentData :exec
DELETE FROM content_data
WHERE content_data_id = ?;

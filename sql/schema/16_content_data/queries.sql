-- name: DropContentDataTable :exec
DROP TABLE content_data;

-- name: CreateContentDataTable :exec
CREATE TABLE IF NOT EXISTS content_data (
    content_data_id TEXT PRIMARY KEY NOT NULL CHECK (length(content_data_id) = 26),
    parent_id TEXT
        REFERENCES content_data
            ON DELETE SET NULL,
    first_child_id TEXT
        REFERENCES content_data
            ON DELETE SET NULL,
    next_sibling_id TEXT
        REFERENCES content_data
            ON DELETE SET NULL,
    prev_sibling_id TEXT
        REFERENCES content_data
            ON DELETE SET NULL,
    route_id TEXT NOT NULL
        REFERENCES routes
            ON DELETE CASCADE,
    datatype_id TEXT NOT NULL
        REFERENCES datatypes
            ON DELETE SET NULL,
    author_id TEXT NOT NULL
        REFERENCES users
            ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'draft',
    date_created TEXT DEFAULT CURRENT_TIMESTAMP,
    date_modified TEXT DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (parent_id) REFERENCES content_data(content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (first_child_id) REFERENCES content_data(content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (next_sibling_id) REFERENCES content_data(content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (prev_sibling_id) REFERENCES content_data(content_data_id) ON DELETE SET NULL,
    FOREIGN KEY (route_id) REFERENCES routes(route_id) ON DELETE RESTRICT,
    FOREIGN KEY (datatype_id) REFERENCES datatypes(datatype_id) ON DELETE RESTRICT,
    FOREIGN KEY (author_id) REFERENCES users(user_id) ON DELETE SET NULL
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
    content_data_id,
    route_id,
    parent_id,
    first_child_id,
    next_sibling_id,
    prev_sibling_id,
    datatype_id,
    author_id,
    status,
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
    ?
) RETURNING *;

-- name: UpdateContentData :exec
UPDATE content_data
SET route_id = ?,
    parent_id = ?,
    first_child_id = ?,
    next_sibling_id = ?,
    prev_sibling_id = ?,
    datatype_id = ?,
    author_id = ?,
    status = ?,
    date_created = ?,
    date_modified = ?
WHERE content_data_id = ?;

-- name: DeleteContentData :exec
DELETE FROM content_data
WHERE content_data_id = ?;

-- name: ListContentDataPaginated :many
SELECT * FROM content_data
ORDER BY content_data_id
LIMIT ? OFFSET ?;

-- name: ListContentDataByRoutePaginated :many
SELECT * FROM content_data
WHERE route_id = ?
ORDER BY content_data_id
LIMIT ? OFFSET ?;
